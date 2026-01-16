# Customer Management Context

**Domain:** CRM, customer relationships, sales pipeline, analytics  
**Ubiquitous Language:** Customer, Company, Deal, Interaction, Lead, Opportunity, Pipeline, Analytics  
**Status:**  Production (Customer, Company, Deal, Interaction, Analytics)

---

## Overview

The **Customer Management Context** handles all CRM (Customer Relationship Management) functionality. This is where business relationships are managed, distinct from technical user authentication (Identity context).

### Responsibilities

 **What Customer Management Context DOES:**

- Customer lifecycle management (Lead → Customer → Churned)
- Company management (B2B organizations)
- Deal/opportunity tracking and sales pipeline
- Interaction history (calls, emails, meetings, notes)
- Customer segmentation and tier management

 **What Customer Management Context DOES NOT DO:**

- User authentication → **Identity Context**
- Orders and contracts → **Order Management Context**
- Invoices and payments → **Billing Context**
- Product inventory → **Warehouse Context**

---

## Domain Errors

This context follows the **Gold Standard** domain error pattern with 79 domain error constants across 4 aggregates.

### Customer Errors

**File**: [`customer/errors.go`](customer/errors.go) - 25 error constants

**Repository Errors**:
- `ErrCustomerNotFound` - Customer not found by ID
- `ErrCustomerUnauthorized` - Access denied

**Business Logic Errors**:
- `ErrCustomerEmailExists` - Duplicate email address
- `ErrCustomerPhoneExists` - Duplicate phone number
- `ErrCustomerInvalidTier` - Invalid tier value
- `ErrCustomerInvalidStatus` - Invalid status transition
- `ErrCustomerLeadToChurned` - Cannot go Lead → Churned directly
- `ErrCustomerChurnedToLead` - Cannot revert Churned → Lead

**Technical Errors**:
- `ErrCustomerCreateFailed` - Creation failed
- `ErrCustomerUpdateFailed` - Update failed

### Company Errors

**File**: [`company/errors.go`](company/errors.go) - 15 error constants

**Repository Errors**:
- `ErrCompanyNotFound` - Company not found by ID

**Business Logic Errors**:
- `ErrCompanyNameExists` - Duplicate company name
- `ErrCompanyTaxIDExists` - Duplicate tax identification
- `ErrCompanyInvalidSize` - Invalid size value
- `ErrCompanyParentNotFound` - Invalid parent company reference
- `ErrCompanyCircularRef` - Circular parent relationship

**Technical Errors**:
- `ErrCompanyCreateFailed` - Creation failed

### Deal Errors

**File**: [`deal/errors.go`](deal/errors.go) - 23 error constants

**Repository Errors**:
- `ErrDealNotFound` - Deal not found by ID

**Business Logic Errors**:
- `ErrDealInvalidStage` - Invalid stage transition
- `ErrDealInvalidAmount` - Negative or zero amount
- `ErrDealInvalidProbability` - Invalid probability value
- `ErrDealAlreadyWon` - Cannot modify won deal
- `ErrDealAlreadyLost` - Cannot modify lost deal
- `ErrDealClosedNoReason` - Lost deal requires reason

**Technical Errors**:
- `ErrDealCreateFailed` - Creation failed
- `ErrDealUpdateFailed` - Update failed

### Interaction Errors

**File**: [`interaction/errors.go`](interaction/errors.go) - 16 error constants

**Repository Errors**:
- `ErrInteractionNotFound` - Interaction not found by ID

**Business Logic Errors**:
- `ErrInteractionInvalidType` - Invalid interaction type
- `ErrInteractionInvalidDirection` - Invalid direction
- `ErrInteractionInvalidOutcome` - Invalid outcome
- `ErrInteractionAlreadyEnded` - Cannot end twice
- `ErrInteractionNotEnded` - Cannot get duration if not ended

**Technical Errors**:
- `ErrInteractionCreateFailed` - Creation failed

### Usage Example

```go
// In UseCase Layer
customer, err := uc.repo.GetCustomer(ctx, customerID)
if err != nil {
    return nil, customer.ErrCustomerNotFound  // Domain constant
}

if err := customer.QualifyAsProspect(); err != nil {
    return nil, err  // ErrCustomerInvalidStatus (business rule)
}

// In Test Layer
err := usecase.QualifyAsProspect(ctx, leadID)
assert.True(t, errors.Is(err, customer.ErrCustomerLeadToChurned))  // Type-safe

// In Handler Layer
cust, err := h.usecase.CreateCustomer(ctx, req.Email, req.Name)
if errors.Is(err, customer.ErrCustomerEmailExists) {
    response.BadRequest(c, "Email already registered")  // 400
    return
}
if errors.Is(err, customer.ErrCustomerNotFound) {
    response.NotFound(c, "Customer not found")  // 404
    return
}
```

### HTTP Status Code Mapping

| Domain Error | HTTP Code | User Message |
|--------------|-----------|--------------|
| `ErrCustomerNotFound` | 404 | "Customer not found" |
| `ErrCustomerEmailExists` | 400 | "Email already registered" |
| `ErrCompanyNameExists` | 400 | "Company name already exists" |
| `ErrDealInvalidStage` | 400 | "Invalid deal stage transition" |
| `ErrDealAlreadyWon` | 400 | "Cannot modify won deal" |
| `ErrInteractionInvalidType` | 400 | "Invalid interaction type" |
| System errors (wrapped) | 500 | "Failed to {operation}" |

### Testing Patterns

Customer Management tests demonstrate **PRE-Phase 2** pattern (string comparison) → **Phase 2** pattern (errors.Is()):

**Before (PRE-Phase 2)**:
```go
err := usecase.CreateCustomer(ctx, "existing@email.com", "Test")
assert.Equal(t, "email already exists", err.Error())  //  String comparison
```

**After (Phase 2)**:
```go
err := usecase.CreateCustomer(ctx, "existing@email.com", "Test")
assert.True(t, errors.Is(err, customer.ErrCustomerEmailExists))  //  Type-safe
```

**Smoke Tests** (26 tests across 4 handlers):
```go
resp := smoke.MakeRequest(t, router, "GET", "/customers/"+invalidID, nil)
smoke.AssertErrorResponse(t, resp, 404, "CUSTOMER_NOT_FOUND")
```

### Statistics

- **Total Domain Errors**: 79 across 4 aggregates
- **errors.go Files**: 4/4 (100% coverage)
- **Tests Using errors.Is()**: 9 integration tests, 100+ unit tests
- **Handler Discrimination**: All 48 endpoints use errors.Is()
- **Phase 2 Session**: 12, 13 (customer-mgmt completed)
- **Milestone**:  **FIRST context at 100% Phase 2** (all aggregates refactored)

**See**: [Domain Errors Guide](../../docs/guides/domain-errors.md) for comprehensive patterns, examples, and migration instructions.

---

## Aggregates

### 1. Customer Aggregate

**Aggregate Root:** `Customer`  
**Purpose:** B2C customer or B2B contact person

**Entity Structure:**

```go
type Customer struct {
    aggregate.BaseAggregate

    // Identity
    ID         uuidv7.UUID
    UserID     *uuidv7.UUID // optional link to Identity.User (if they have account)
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

    // Business Data
    contacts   []Contact      // private
    deals      []Deal         // private

    // Lifecycle
    CreatedAt  time.Time
    UpdatedAt  time.Time
    ChurnedAt  *time.Time
}
```

**Business Rules:**

- Email must be unique within company (B2B) or globally (B2C)
- Cannot have both UserID and CompanyID (either B2C user OR B2B company contact)
- Lead → Prospect → Customer (one-way state transitions)
- Churned customers can be reactivated
- Only one assigned sales rep at a time

**Methods:**

```go
func (c *Customer) QualifyAsProspect() error
func (c *Customer) ConvertToCustomer() error
func (c *Customer) Churn(reason string) error
func (c *Customer) Reactivate() error
func (c *Customer) AssignToRep(repID uuidv7.UUID) error
func (c *Customer) UpgradeTier(newTier CustomerTier) error
func (c *Customer) AddContact(contact Contact) error
func (c *Customer) CreateDeal(deal Deal) error
```

---

### 2. Company Aggregate

**Aggregate Root:** `Company`  
**Purpose:** B2B organization management

**Entity Structure:**

```go
type Company struct {
    aggregate.BaseAggregate

    // Identity
    ID            uuidv7.UUID
    Name          string
    LegalName     string
    TaxID         string // VAT, EIN, etc.

    // Contact
    Website       string
    Email         valueobject.Email
    Phone         valueobject.Phone
    Address       valueobject.Address

    // Business Info
    Industry      string
    EmployeeCount int
    AnnualRevenue *valueobject.Money

    // Status
    Status        CompanyStatus // lead, active, inactive

    // Relationship
    PrimaryContact uuidv7.UUID  // references Customer
    AssignedTo     uuidv7.UUID  // Account manager

    // Business Data
    contacts      []Customer    // private, B2B contact persons
    deals         []Deal        // private

    // Lifecycle
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Business Rules:**

- Company name + TaxID must be unique
- Must have at least one contact person (Customer with CompanyID)
- Primary contact must be one of company contacts
- Only active companies can have deals
- Cannot delete company with active deals

**Methods:**

```go
func (c *Company) AddContact(customer Customer) error
func (c *Company) SetPrimaryContact(customerID uuidv7.UUID) error
func (c *Company) Activate() error
func (c *Company) Deactivate() error
func (c *Company) AssignAccountManager(userID uuidv7.UUID) error
func (c *Company) CreateDeal(deal Deal) error
```

---

### 3. Deal Aggregate

**Aggregate Root:** `Deal`  
**Purpose:** Sales opportunity tracking

**Entity Structure:**

```go
type Deal struct {
    aggregate.BaseAggregate

    // Identity
    ID          uuidv7.UUID
    CustomerID  uuidv7.UUID  // references Customer
    CompanyID   *uuidv7.UUID // optional (if B2B)

    // Deal Info
    Name        string       // "Q1 2026 Annual Contract"
    Description string
    Value       valueobject.Money

    // Pipeline
    Stage       DealStage    // discovery, proposal, negotiation, closing, won, lost
    Probability int          // 0-100%

    // Dates
    ExpectedCloseDate time.Time
    ActualCloseDate   *time.Time

    // Ownership
    AssignedTo  uuidv7.UUID  // Sales rep

    // Notes
    notes       []Note       // private

    // Lifecycle
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Business Rules:**

- Deal value must be positive
- Probability must be 0-100%
- Stage transitions follow pipeline order
- Won/Lost are final stages (cannot revert)
- Expected close date must be in future (unless won/lost)

**Methods:**

```go
func (d *Deal) AdvanceStage() error
func (d *Deal) MarkAsWon() error
func (d *Deal) MarkAsLost(reason string) error
func (d *Deal) UpdateValue(newValue valueobject.Money) error
func (d *Deal) UpdateProbability(probability int) error
func (d *Deal) AddNote(note string) error
```

---

### 4. Interaction Aggregate

**Aggregate Root:** `Interaction`  
**Purpose:** Customer interaction history

**Entity Structure:**

```go
type Interaction struct {
    aggregate.BaseAggregate

    // Identity
    ID          uuidv7.UUID
    CustomerID  uuidv7.UUID     // references Customer
    CompanyID   *uuidv7.UUID    // optional (if B2B)
    DealID      *uuidv7.UUID    // optional (if related to deal)

    // Interaction Details
    Type        InteractionType // call, email, meeting, note
    Direction   Direction       // inbound, outbound
    Subject     string
    Content     string

    // Participants
    InitiatedBy uuidv7.UUID     // User who created interaction
    Participants []uuidv7.UUID  // Other participants (meetings)

    // Metadata
    Duration    *time.Duration  // for calls/meetings
    ScheduledAt *time.Time      // for future meetings

    // Lifecycle
    CreatedAt   time.Time
}
```

**Business Rules:**

- Must reference either Customer or Company (or both)
- Duration required for calls and meetings
- Future meetings must have ScheduledAt
- Content required for emails and notes

**Methods:**

```go
func (i *Interaction) AddParticipant(userID uuidv7.UUID) error
func (i *Interaction) Complete() error
func (i *Interaction) Reschedule(newTime time.Time) error
```

---

## Domain Events

```go
// Customer events
type CustomerCreatedEvent struct {
    CustomerID uuidv7.UUID
    Status     CustomerStatus
    Email      string
    Timestamp  time.Time
}

type CustomerConvertedEvent struct {
    CustomerID uuidv7.UUID
    FromStatus CustomerStatus
    ToStatus   CustomerStatus
    Timestamp  time.Time
}

type CustomerChurnedEvent struct {
    CustomerID uuidv7.UUID
    Reason     string
    Timestamp  time.Time
}

// Deal events
type DealCreatedEvent struct {
    DealID     uuidv7.UUID
    CustomerID uuidv7.UUID
    Value      valueobject.Money
    Timestamp  time.Time
}

type DealWonEvent struct {
    DealID     uuidv7.UUID
    CustomerID uuidv7.UUID
    Value      valueobject.Money
    Timestamp  time.Time
}

type DealLostEvent struct {
    DealID     uuidv7.UUID
    Reason     string
    Timestamp  time.Time
}
```

---

## Communication with Other Contexts

### With Identity Context

```go
// CustomerMgmt subscribes to user registration
eventBus.Subscribe("user.registered", func(event UserRegisteredEvent) {
    // Create customer profile for new user
    customerService.CreateCustomerFromUser(event.UserID, event.Email)
})
```

### With Order Management Context

```go
// Publish deal won event
eventBus.Publish("deal.won", DealWonEvent{...})

// OrderMgmt subscribes and creates contract
eventBus.Subscribe("deal.won", func(event DealWonEvent) {
    orderService.CreateContractFromDeal(event.DealID)
})
```

### With Billing Context

```go
// CustomerMgmt publishes customer tier upgrade
eventBus.Publish("customer.tier_upgraded", CustomerTierUpgradedEvent{...})

// Billing subscribes and updates subscription
eventBus.Subscribe("customer.tier_upgraded", func(event CustomerTierUpgradedEvent) {
    billingService.UpgradeSubscription(event.CustomerID, event.NewTier)
})
```

---

## Database Schema (Planned)

```sql
-- Customers
CREATE TABLE crm_customers (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID REFERENCES users(id),
    company_id UUID REFERENCES crm_companies(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    status VARCHAR(50) NOT NULL,
    tier VARCHAR(50) NOT NULL,
    source VARCHAR(100),
    assigned_to UUID REFERENCES users(id),
    tags TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    churned_at TIMESTAMP
);

-- Companies
CREATE TABLE crm_companies (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(255) NOT NULL,
    legal_name VARCHAR(255),
    tax_id VARCHAR(100),
    website VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    address JSONB,
    industry VARCHAR(100),
    employee_count INT,
    annual_revenue JSONB,
    status VARCHAR(50) NOT NULL,
    primary_contact UUID REFERENCES crm_customers(id),
    assigned_to UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Deals
CREATE TABLE crm_deals (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    customer_id UUID NOT NULL REFERENCES crm_customers(id),
    company_id UUID REFERENCES crm_companies(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    value JSONB NOT NULL,
    stage VARCHAR(50) NOT NULL,
    probability INT NOT NULL,
    expected_close_date DATE,
    actual_close_date DATE,
    assigned_to UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Interactions
CREATE TABLE crm_interactions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    customer_id UUID REFERENCES crm_customers(id),
    company_id UUID REFERENCES crm_companies(id),
    deal_id UUID REFERENCES crm_deals(id),
    type VARCHAR(50) NOT NULL,
    direction VARCHAR(50) NOT NULL,
    subject VARCHAR(255),
    content TEXT,
    initiated_by UUID NOT NULL REFERENCES users(id),
    participants UUID[],
    duration INTERVAL,
    scheduled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

---

## API Endpoints (68 Total)

### Customer Management (14 endpoints) 

```
GET    /api/v1/customer-mgmt/customers              # List customers
POST   /api/v1/customer-mgmt/customers              # Create customer
GET    /api/v1/customer-mgmt/customers/:id          # Get customer
PUT    /api/v1/customer-mgmt/customers/:id          # Update customer
DELETE /api/v1/customer-mgmt/customers/:id          # Delete customer
POST   /api/v1/customer-mgmt/customers/:id/qualify  # Lead → Prospect
POST   /api/v1/customer-mgmt/customers/:id/convert  # Prospect → Customer
POST   /api/v1/customer-mgmt/customers/:id/churn    # Mark as churned
POST   /api/v1/customer-mgmt/customers/:id/reactivate # Reactivate churned
POST   /api/v1/customer-mgmt/customers/:id/assign   # Assign to sales rep
POST   /api/v1/customer-mgmt/customers/:id/upgrade  # Upgrade tier
GET    /api/v1/customer-mgmt/customers/status/:status # Filter by status
GET    /api/v1/customer-mgmt/customers/tier/:tier   # Filter by tier
GET    /api/v1/customer-mgmt/customers/rep/:id      # Filter by sales rep
```

### Company Management (14 endpoints) 

```
GET    /api/v1/customer-mgmt/companies              # List companies
POST   /api/v1/customer-mgmt/companies              # Create company
GET    /api/v1/customer-mgmt/companies/:id          # Get company
PUT    /api/v1/customer-mgmt/companies/:id          # Update company
DELETE /api/v1/customer-mgmt/companies/:id          # Delete company
GET    /api/v1/customer-mgmt/companies/:id/customers # Get company customers
POST   /api/v1/customer-mgmt/companies/:id/activate # Activate company
POST   /api/v1/customer-mgmt/companies/:id/deactivate # Deactivate company
PUT    /api/v1/customer-mgmt/companies/:id/parent   # Set parent company
GET    /api/v1/customer-mgmt/companies/:id/subsidiaries # Get subsidiaries
GET    /api/v1/customer-mgmt/companies/type/:type   # Filter by type
GET    /api/v1/customer-mgmt/companies/search       # Search companies
POST   /api/v1/customer-mgmt/companies/:id/merge    # Merge companies
GET    /api/v1/customer-mgmt/companies/:id/hierarchy # Get company hierarchy
```

### Deal Management (12 endpoints) 

```
GET    /api/v1/customer-mgmt/deals                  # List deals
POST   /api/v1/customer-mgmt/deals                  # Create deal
GET    /api/v1/customer-mgmt/deals/:id              # Get deal
PUT    /api/v1/customer-mgmt/deals/:id              # Update deal
DELETE /api/v1/customer-mgmt/deals/:id              # Delete deal
POST   /api/v1/customer-mgmt/deals/:id/move         # Move to stage
POST   /api/v1/customer-mgmt/deals/:id/win          # Mark as won
POST   /api/v1/customer-mgmt/deals/:id/lose         # Mark as lost
GET    /api/v1/customer-mgmt/deals/stage/:stage     # Filter by stage
GET    /api/v1/customer-mgmt/deals/customer/:id     # Filter by customer
GET    /api/v1/customer-mgmt/deals/rep/:id          # Filter by sales rep
GET    /api/v1/customer-mgmt/deals/stats            # Pipeline statistics
```

### Interaction Management (14 endpoints) 

```
GET    /api/v1/customer-mgmt/interactions           # List interactions
POST   /api/v1/customer-mgmt/interactions           # Create interaction
GET    /api/v1/customer-mgmt/interactions/:id       # Get interaction
PUT    /api/v1/customer-mgmt/interactions/:id       # Update interaction
DELETE /api/v1/customer-mgmt/interactions/:id       # Delete interaction
POST   /api/v1/customer-mgmt/interactions/:id/start # Start interaction
POST   /api/v1/customer-mgmt/interactions/:id/end   # End interaction
POST   /api/v1/customer-mgmt/interactions/:id/complete # Complete interaction
GET    /api/v1/customer-mgmt/interactions/type/:type # Filter by type
GET    /api/v1/customer-mgmt/interactions/customer/:id # Filter by customer
GET    /api/v1/customer-mgmt/interactions/company/:id # Filter by company
GET    /api/v1/customer-mgmt/interactions/rep/:id   # Filter by sales rep
GET    /api/v1/customer-mgmt/interactions/follow-up # Pending follow-ups
GET    /api/v1/customer-mgmt/interactions/stats     # Interaction statistics
```

### Analytics (8 endpoints)  **NEW**

**CQRS Read Models** - Optimized analytical queries for business intelligence and reporting.

```
GET    /api/v1/customer-mgmt/analytics/customers/overview    # Customer overview metrics
GET    /api/v1/customer-mgmt/analytics/customers/lifecycle   # Customer lifecycle funnel
GET    /api/v1/customer-mgmt/analytics/customers/segmentation # Customer segmentation
GET    /api/v1/customer-mgmt/analytics/deals/pipeline        # Deal pipeline analysis
GET    /api/v1/customer-mgmt/analytics/deals/conversions     # Deal stage conversions
GET    /api/v1/customer-mgmt/analytics/sales-reps/performance # Sales rep performance
GET    /api/v1/customer-mgmt/analytics/revenue/time-series   # Revenue over time
GET    /api/v1/customer-mgmt/analytics/interactions/insights # Interaction insights
```

**See**: [Analytics Documentation](analytics/README.md) for complete API reference with request/response examples.

**Key Features**:
- **CQRS Pattern**: Separate read models optimized for reporting
- **Denormalized Queries**: Fast analytical queries with joins across aggregates
- **No Repository Pattern**: Direct SQL for read-only analytics
- **Real-time Metrics**: Customer overview, deal pipeline, sales rep performance
- **Time Series**: Revenue trends with configurable granularity (day/week/month)
- **Funnel Analysis**: Customer lifecycle transitions and deal conversions

---

## Implementation Status

###  Completed (All Aggregates in Production)

**Customer Aggregate** (14 endpoints, 100+ tests):
- Customer entity with state machine (Lead → Prospect → Customer → Churned)
- Customer repository (PostgreSQL)
- Customer use cases (lifecycle, tier management, assignment)
- Customer HTTP handlers (CRUD + business operations)
- Integration tests with real database

**Company Aggregate** (14 endpoints, 80+ tests):
- Company entity with hierarchical structure
- Company repository (PostgreSQL)
- Company use cases (hierarchy, customers, activation)
- Company HTTP handlers (CRUD + business operations)
- Integration tests with real database

**Deal Aggregate** (12 endpoints, 90+ tests):
- Deal entity with stage transitions (Lead → ... → Closed Won/Lost)
- Deal repository (PostgreSQL)
- Deal use cases (pipeline, stage movement, win/loss)
- Deal HTTP handlers (CRUD + business operations)
- Integration tests with real database

**Interaction Aggregate** (14 endpoints, 74 tests):
- Interaction entity with lifecycle (Create → In Progress → Completed)
- Interaction repository (PostgreSQL with LEFT JOIN optimization)
- Interaction use cases (create, update, complete, follow-ups)
- Interaction HTTP handlers (CRUD + business operations)
- Integration tests with real database

**Analytics CQRS** (8 endpoints, 15 tests):
- 9 analytical query methods (direct SQL, no repository pattern)
- Customer metrics (overview, lifecycle, segmentation)
- Deal metrics (pipeline, conversions)
- Sales rep performance
- Revenue time series
- Interaction insights
- Unit tests (7) + Integration tests (8)

### Total Statistics

- **Aggregates**: 4 complete (Customer, Company, Deal, Interaction)
- **Endpoints**: 68 (14 + 14 + 12 + 14 + 8 + 6 admin)
- **Tests**: 360+ (unit + integration + benchmarks)
- **Test Coverage**: 85%+ average
- **Database**: 4 tables with proper indexing
- **Performance**: Optimized queries (LEFT JOIN, N+1 prevention)

---

**Status:**  Production-ready (Customer, Company, Deal, Interaction, Analytics)  
**Dependencies:** Identity Context (for User assignments)  
**Next:** Monitoring, caching layer, advanced reporting
