# Customer Management Context

**Domain:** CRM, customer relationships, sales pipeline  
**Ubiquitous Language:** Customer, Company, Deal, Interaction, Lead, Opportunity  
**Status:** ✅ Production (Customer, Company, Deal) | Interaction planned Q1 2026

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

## API Endpoints (Planned)

### Customer Management

```
GET    /api/v1/crm/customers              # List customers
POST   /api/v1/crm/customers              # Create customer
GET    /api/v1/crm/customers/:id          # Get customer
PUT    /api/v1/crm/customers/:id          # Update customer
DELETE /api/v1/crm/customers/:id          # Delete customer
POST   /api/v1/crm/customers/:id/qualify  # Lead → Prospect
POST   /api/v1/crm/customers/:id/convert  # Prospect → Customer
POST   /api/v1/crm/customers/:id/churn    # Mark as churned
```

### Company Management

```
GET    /api/v1/crm/companies              # List companies
POST   /api/v1/crm/companies              # Create company
GET    /api/v1/crm/companies/:id          # Get company
PUT    /api/v1/crm/companies/:id          # Update company
DELETE /api/v1/crm/companies/:id          # Delete company
POST   /api/v1/crm/companies/:id/contacts # Add contact person
```

### Deal Management

```
GET    /api/v1/crm/deals                  # List deals
POST   /api/v1/crm/deals                  # Create deal
GET    /api/v1/crm/deals/:id              # Get deal
PUT    /api/v1/crm/deals/:id              # Update deal
DELETE /api/v1/crm/deals/:id              # Delete deal
POST   /api/v1/crm/deals/:id/advance      # Advance stage
POST   /api/v1/crm/deals/:id/win          # Mark as won
POST   /api/v1/crm/deals/:id/lose         # Mark as lost
```

### Interaction Management

```
GET    /api/v1/crm/interactions           # List interactions
POST   /api/v1/crm/interactions           # Create interaction
GET    /api/v1/crm/interactions/:id       # Get interaction
PUT    /api/v1/crm/interactions/:id       # Update interaction
DELETE /api/v1/crm/interactions/:id       # Delete interaction
```

---

## Implementation Plan

### Phase 2 (3 weeks after Phase 1)

**Week 1: Customer Aggregate**

- Customer entity (60 tests)
- Customer repository
- Customer use cases
- Customer handlers

**Week 2: Company & Deal Aggregates**

- Company aggregate (50 tests)
- Deal aggregate (60 tests)
- Repositories and use cases

**Week 3: Interactions & Integration**

- Interaction aggregate (40 tests)
- Event-driven integration with other contexts
- Saga pattern for complex workflows
- Integration tests (50 tests)

**Total:** 260 tests for Customer Management context

---

**Status:**  Planned for Phase 2  
**Dependencies:** Phase 1 complete (Foundation packages + Identity context)  
**Next:** Wait for Phase 1 completion, then implement Customer aggregate
