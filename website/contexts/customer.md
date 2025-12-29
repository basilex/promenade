# Customer Management Context

**CRM functionality for customer relationships and sales pipeline**

---

## Overview

The **Customer Management Context** handles all CRM (Customer Relationship Management) functionality. This is where business relationships are managed, distinct from technical user authentication (Identity context).

### Key Concepts

- **Customer** - B2C customer or B2B contact person
- **Company** - B2B organization management (planned)
- **Deal** - Sales opportunity tracking (planned)
- **Interaction** - Customer interaction history (planned)

### Status

| Aggregate     | Status     | Description                    |
| ------------- | ---------- | ------------------------------ |
| **Customer**  | Production | Core customer lifecycle        |
| **Company**   | Planned    | B2B organizations              |
| **Deal**      | Planned    | Sales pipeline                 |
| **Interaction**| Planned   | Calls, emails, meetings, notes |

---

## Customer Aggregate (Production)

### Overview

The **Customer** aggregate manages the complete customer lifecycle from lead to churned customer, with support for both B2C and B2B scenarios.

### Entity Structure

```go
type Customer struct {
    aggregate.BaseAggregate

    // Identity
    ID         uuidv7.UUID
    UserID     *uuidv7.UUID // optional link to Identity.User
    CompanyID  *uuidv7.UUID // optional (B2B) or nil (B2C)

    // Basic Info
    Name       string
    Email      valueobject.Email
    Phone      *valueobject.Phone

    // Status
    Status     CustomerStatus // lead, prospect, customer, churned
    Tier       CustomerTier   // free, basic, pro, enterprise
    Source     string         // website, referral, cold-call, event

    // Relationship
    AssignedTo uuidv7.UUID    // Sales rep (Identity.User)
    Tags       []string       // marketing, vip, high-value

    // Lifecycle
    CreatedAt  time.Time
    UpdatedAt  time.Time
    ChurnedAt  *time.Time
}
```

### Customer Status (Lifecycle)

```
┌──────┐   QualifyAsProspect   ┌──────────┐   ConvertToCustomer   ┌──────────┐
│ Lead │ ───────────────────> │ Prospect │ ───────────────────> │ Customer │
└──────┘                       └──────────┘                       └──────────┘
                                                                         │
                                                                         │ Churn
                                                                         v
                                                                   ┌─────────┐
                                                                   │ Churned │
                                                                   └─────────┘
                                                                         │
                                                                         │ Reactivate
                                                                         v
                                                                   ┌──────────┐
                                                                   │ Customer │
                                                                   └──────────┘
```

### Customer Tiers

| Tier         | Description              | Typical Use Case            |
| ------------ | ------------------------ | --------------------------- |
| **free**     | Free tier customer       | Trial users, freemium       |
| **basic**    | Basic paid plan          | Small businesses, startups  |
| **pro**      | Professional plan        | Growing businesses          |
| **enterprise**| Enterprise plan         | Large organizations         |

### Business Rules

1. **Email uniqueness**:
   - B2C: Email must be globally unique
   - B2B: Email must be unique within company

2. **Mutual exclusivity**: Customer cannot have both UserID and CompanyID
   - UserID → B2C customer (individual user)
   - CompanyID → B2B contact person

3. **Status transitions** (one-way):
   - Lead → Prospect → Customer → Churned
   - Churned → Customer (reactivation allowed)
   - Cannot go backwards (e.g., Customer → Prospect)

4. **Assignment**: Only one assigned sales rep at a time

5. **Tier upgrades**: Can only upgrade tier (not downgrade)

### Key Methods

```go
// Status transitions
func (c *Customer) QualifyAsProspect() error
func (c *Customer) ConvertToCustomer() error
func (c *Customer) Churn(reason string) error
func (c *Customer) Reactivate() error

// Management
func (c *Customer) AssignToRep(repID uuidv7.UUID) error
func (c *Customer) UpgradeTier(newTier CustomerTier) error
func (c *Customer) AddTag(tag string) error
func (c *Customer) RemoveTag(tag string) error
```

---

## Database Schema

### Customers Table

```sql
CREATE TABLE customer_mgmt_customers (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID REFERENCES identity_users(id),
    company_id UUID REFERENCES customer_mgmt_companies(id),
    
    -- Basic Info
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    
    -- Status
    status VARCHAR(50) NOT NULL,
    tier VARCHAR(50) NOT NULL,
    source VARCHAR(100),
    
    -- Relationship
    assigned_to UUID REFERENCES identity_users(id),
    tags TEXT[],
    
    -- Lifecycle
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    churned_at TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT unique_email_per_company UNIQUE (email, company_id)
);

-- Indexes
CREATE INDEX idx_customers_status ON customer_mgmt_customers(status);
CREATE INDEX idx_customers_tier ON customer_mgmt_customers(tier);
CREATE INDEX idx_customers_assigned_to ON customer_mgmt_customers(assigned_to);
CREATE INDEX idx_customers_user_id ON customer_mgmt_customers(user_id);
CREATE INDEX idx_customers_company_id ON customer_mgmt_customers(company_id);
```

---

## API Endpoints

### Customer Management

**Base URL**: `/api/v1/customer-mgmt`

#### Create Customer

```http
POST /customers
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "phone": "+441234567890",
  "status": "lead",
  "tier": "free",
  "source": "website",
  "tags": ["marketing", "newsletter"]
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "phone": "+441234567890",
    "status": "lead",
    "tier": "free",
    "source": "website",
    "tags": ["marketing", "newsletter"],
    "assigned_to": null,
    "created_at": "2025-12-29T12:00:00Z",
    "updated_at": "2025-12-29T12:00:00Z"
  }
}
```

#### List Customers

```http
GET /customers?status=lead&tier=free&page=1&page_size=20
Authorization: Bearer <token>
```

**Query Parameters**:
- `status` - Filter by customer status (lead, prospect, customer, churned)
- `tier` - Filter by customer tier (free, basic, pro, enterprise)
- `assigned_to` - Filter by assigned sales rep ID
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGABC1234567890DEFGHIJK",
      "name": "John Doe",
      "email": "john.doe@example.com",
      "status": "lead",
      "tier": "free"
    }
  ],
  "pagination": {
    "total": 150,
    "page": 1,
    "page_size": 20,
    "total_pages": 8
  }
}
```

#### Get Customer by ID

```http
GET /customers/:id
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "phone": "+441234567890",
    "status": "lead",
    "tier": "free",
    "source": "website",
    "tags": ["marketing", "newsletter"],
    "assigned_to": null,
    "created_at": "2025-12-29T12:00:00Z",
    "updated_at": "2025-12-29T12:00:00Z"
  }
}
```

#### Update Customer

```http
PUT /customers/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "John Doe Updated",
  "phone": "+441234567890",
  "tags": ["vip", "high-value"]
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "name": "John Doe Updated",
    "phone": "+441234567890",
    "tags": ["vip", "high-value"],
    "updated_at": "2025-12-29T12:05:00Z"
  }
}
```

#### Qualify as Prospect

```http
POST /customers/:id/qualify
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "status": "prospect",
    "updated_at": "2025-12-29T12:10:00Z"
  }
}
```

#### Convert to Customer

```http
POST /customers/:id/convert
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "status": "customer",
    "updated_at": "2025-12-29T12:15:00Z"
  }
}
```

#### Churn Customer

```http
POST /customers/:id/churn
Authorization: Bearer <token>
Content-Type: application/json

{
  "reason": "Price too high"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "status": "churned",
    "churned_at": "2025-12-29T12:20:00Z"
  }
}
```

#### Assign Sales Rep

```http
POST /customers/:id/assign
Authorization: Bearer <token>
Content-Type: application/json

{
  "assigned_to": "01JGXYZ9876543210ABCDEFGH"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "assigned_to": "01JGXYZ9876543210ABCDEFGH",
    "updated_at": "2025-12-29T12:25:00Z"
  }
}
```

---

## Usage Examples

### Create and Qualify Lead

```bash
# 1. Create lead
CUSTOMER_ID=$(curl -s -X POST http://localhost:8081/api/v1/customer-mgmt/customers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@startup.com",
    "phone": "+441234567890",
    "status": "lead",
    "tier": "free",
    "source": "website",
    "tags": ["marketing"]
  }' | jq -r '.data.id')

# 2. Qualify as prospect
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/qualify \
  -H "Authorization: Bearer $TOKEN"

# 3. Assign to sales rep
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/assign \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"assigned_to":"01JGREP9876543210ABCDEFGH"}'

# 4. Convert to customer
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/convert \
  -H "Authorization: Bearer $TOKEN"
```

### Filter and Search

```bash
# Get all leads assigned to me
curl "http://localhost:8081/api/v1/customer-mgmt/customers?status=lead&assigned_to=$MY_USER_ID" \
  -H "Authorization: Bearer $TOKEN"

# Get all enterprise customers
curl "http://localhost:8081/api/v1/customer-mgmt/customers?tier=enterprise" \
  -H "Authorization: Bearer $TOKEN"

# Get VIP customers
curl "http://localhost:8081/api/v1/customer-mgmt/customers?tags=vip" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Domain Events

Customer aggregate publishes events for asynchronous communication:

```go
// Customer created
type CustomerCreatedEvent struct {
    CustomerID uuidv7.UUID
    Status     CustomerStatus
    Email      string
    Timestamp  time.Time
}

// Status changed
type CustomerStatusChangedEvent struct {
    CustomerID uuidv7.UUID
    FromStatus CustomerStatus
    ToStatus   CustomerStatus
    Timestamp  time.Time
}

// Customer converted
type CustomerConvertedEvent struct {
    CustomerID uuidv7.UUID
    Timestamp  time.Time
}

// Customer churned
type CustomerChurnedEvent struct {
    CustomerID uuidv7.UUID
    Reason     string
    Timestamp  time.Time
}

// Tier upgraded
type CustomerTierUpgradedEvent struct {
    CustomerID uuidv7.UUID
    FromTier   CustomerTier
    ToTier     CustomerTier
    Timestamp  time.Time
}
```

---

## Testing

### Test Statistics

| Type        | Tests | Coverage | Description                  |
| ----------- | ----- | -------- | ---------------------------- |
| Unit        | 25    | 90%      | Customer entity logic        |
| Smoke       | 18    | -        | HTTP handlers (mock-based)   |
| Integration | 14    | -        | Repository with real DB      |
| **Total**   | **57**| **85%**  | Comprehensive test coverage  |

### Run Tests

```bash
# All customer tests
go test ./internal/contexts/customer-mgmt/... -v

# Smoke tests
go test ./test/smoke/contexts/customer-mgmt/... -v

# Integration tests
go test ./test/integration/contexts/customer-mgmt/... -v
```

---

## Communication with Other Contexts

### With Identity Context

```go
// Subscribe to user registration
eventBus.Subscribe("identity.user.registered", func(event UserRegisteredEvent) {
    // Create customer profile for new user
    customerService.CreateCustomerFromUser(event.UserID, event.Email)
})
```

### With Order Management (Planned)

```go
// Publish customer converted event
eventBus.Publish("customer.converted", CustomerConvertedEvent{...})

// Order Management creates initial contract
eventBus.Subscribe("customer.converted", func(event CustomerConvertedEvent) {
    orderService.CreateWelcomeContract(event.CustomerID)
})
```

### With Billing (Planned)

```go
// Publish tier upgrade event
eventBus.Publish("customer.tier_upgraded", CustomerTierUpgradedEvent{...})

// Billing updates subscription
eventBus.Subscribe("customer.tier_upgraded", func(event CustomerTierUpgradedEvent) {
    billingService.UpgradeSubscription(event.CustomerID, event.ToTier)
})
```

---

## Planned Features

### Company Aggregate (Q1 2026)

- B2B organization management
- Multiple contacts per company
- Account manager assignment
- Company-level deals and interactions

### Deal Aggregate (Q1 2026)

- Sales opportunity tracking
- Pipeline stages (discovery, proposal, negotiation, closing)
- Deal value and probability
- Won/lost tracking

### Interaction Aggregate (Q2 2026)

- Call logging
- Email tracking
- Meeting notes
- Customer interaction history

---

## Best Practices

### Customer Lifecycle Management

1. **Lead qualification**: Verify email and phone before qualifying
2. **Proper assignment**: Assign sales rep based on territory or workload
3. **Status tracking**: Always use proper state transitions (no shortcuts)
4. **Churn handling**: Record churn reason for analytics
5. **Reactivation**: Verify customer details before reactivation

### Performance Optimization

1. **Indexes**: Ensure indexes on `status`, `tier`, `assigned_to` columns
2. **Pagination**: Always use pagination for list endpoints (max 100 items)
3. **Caching**: Cache customer counts by status/tier (refresh every 5 minutes)
4. **Search**: Use database full-text search for customer name/email lookups

### Security

1. **Authorization**: Verify user has permission to manage customers
2. **Soft delete**: Never hard-delete customers (use `deleted_at`)
3. **Audit trail**: Log all status changes and assignments
4. **Data privacy**: Implement GDPR-compliant data export/deletion

---

## Next Steps

- [Identity Context](/contexts/identity) - User authentication and profiles
- [Shared Context](/contexts/shared) - Reference data (Countries, Timezones, etc.)
- [Architecture Guide](/guide/architecture) - Domain-Driven Design principles
- [API Reference](/guide/api-reference) - Complete endpoint documentation
