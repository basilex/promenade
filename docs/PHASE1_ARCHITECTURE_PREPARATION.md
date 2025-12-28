# Phase 1: Architecture Preparation

**Status:**  DAY 2 COMPLETE  → DAY 3 READY  
**Start Date:** December 27, 2025  
**Duration:** 1 week (5-7 days)  
**Goal:** Prepare architecture for CRM complexity with Bounded Contexts  
**Progress:** Day 2/7  (Context structure complete: 6 comprehensive READMEs created!)

---

## Overview

### Why This Phase?

Current modular architecture works well for simple modules (posts, profiles), but will hit limitations with CRM complexity:

** Current Problems:**

- Inter-module dependencies (CRM needs contacts from profiles)
- Transaction boundaries (contract + invoice + payment = 3 modules, inconsistent state)
- Event Bus limitations (async-only, no SYNC validation for critical operations)
- Scattered domain logic (Customer spread across profiles, billing, future CRM)

** Solution:**

- **Bounded Contexts** (DDD pattern)
- **Aggregate Roots** (transactional consistency)
- **Saga Pattern** (distributed transactions)
- **Value Objects** (shared types like Money, Address)

---

## Architecture Goals

### 1. Bounded Contexts

```
/internal
  /contexts                    # ← Domain boundaries
    /identity                  # ← User auth, profiles (Core domain)
      /user                    # ← User aggregate
      /profile                 # ← Profile aggregate
      /contact                 # ← Contact aggregate (moved from profiles)

    /customer-mgmt             # ← CRM context (NEW)
      /customer                # ← Customer aggregate
      /company                 # ← Company aggregate
      /deal                    # ← Deal aggregate
      /interaction             # ← Interaction history

    /order-mgmt                # ← Order context (NEW)
      /order                   # ← Order aggregate
      /contract                # ← Contract aggregate
      /fulfillment             # ← Fulfillment process

    /billing                   # ← Billing context (existing, refactored)
      /subscription            # ← Subscription aggregate
      /invoice                 # ← Invoice aggregate
      /payment                 # ← Payment aggregate

    /warehouse                 # ← Warehouse context (NEW)
      /product                 # ← Product aggregate
      /inventory               # ← Inventory aggregate
      /location                # ← Location aggregate

  /shared-kernel               # ← Shared value objects
    /money                     # ← Money type
    /address                   # ← Address type
    /contact-info              # ← Email, phone
    /date-range                # ← Date range

  /modules                     # ← Existing modules (backward compatibility)
    /posts                     # ← Keep as is
    /profiles                  # ← Gradually migrate to contexts/identity
    /analytics                 # ← Keep as is
    /workflows                 # ← Keep as is (production-ready)
    /notifications             # ← Keep as is
    /audit                     # ← Keep as is
```

### 2. Aggregate Roots

Each context has aggregates that enforce business rules:

```go
// Example: Customer Aggregate
type Customer struct {
    // Aggregate Root
    ID          uuidv7.UUID
    CompanyID   *uuidv7.UUID // optional (B2B) or nil (B2C)

    // Value Objects
    Name        string
    Email       Email          // value object with validation
    Phone       Phone          // value object with validation
    Address     Address        // value object (street, city, country)

    // Business State
    Status      CustomerStatus // active, suspended, churned
    Tier        CustomerTier   // free, basic, pro, enterprise

    // Lifecycle
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // Aggregate enforces invariants
    contacts    []Contact      // private, accessed via methods only
    deals       []Deal         // private, accessed via methods only
}

// Methods enforce business rules
func (c *Customer) AddContact(contact Contact) error {
    if len(c.contacts) >= MaxContactsPerCustomer {
        return ErrTooManyContacts
    }
    // Validation logic here
    c.contacts = append(c.contacts, contact)
    return nil
}

func (c *Customer) CreateDeal(deal Deal) error {
    if c.Status != CustomerStatusActive {
        return ErrCustomerNotActive
    }
    // Business rules here
    c.deals = append(c.deals, deal)
    return nil
}
```

### 3. Saga Pattern

For operations spanning multiple aggregates/contexts:

```go
// Example: Order Creation Saga
type CreateOrderSaga struct {
    steps []SagaStep
}

func NewCreateOrderSaga() *CreateOrderSaga {
    return &CreateOrderSaga{
        steps: []SagaStep{
            {Name: "ValidateCustomer", Execute: validateCustomer, Compensate: nil},
            {Name: "ReserveInventory", Execute: reserveInventory, Compensate: releaseInventory},
            {Name: "CreateInvoice", Execute: createInvoice, Compensate: cancelInvoice},
            {Name: "ProcessPayment", Execute: processPayment, Compensate: refundPayment},
            {Name: "CreateOrder", Execute: createOrder, Compensate: cancelOrder},
        },
    }
}

// Saga orchestrates distributed transaction
func (s *CreateOrderSaga) Execute(ctx context.Context, data OrderData) error {
    executed := []SagaStep{}

    for _, step := range s.steps {
        if err := step.Execute(ctx, data); err != nil {
            // Rollback: execute compensations in reverse
            for i := len(executed) - 1; i >= 0; i-- {
                if executed[i].Compensate != nil {
                    executed[i].Compensate(ctx, data)
                }
            }
            return fmt.Errorf("saga failed at %s: %w", step.Name, err)
        }
        executed = append(executed, step)
    }

    return nil
}
```

---

## Implementation Plan

### Day 1: Foundation Packages

#### Task 1.1: Create `pkg/aggregate` package

**File:** `pkg/aggregate/aggregate.go`

```go
package aggregate

import (
    "github.com/basilex/promenade/pkg/uuidv7"
    "time"
)

// Root is the interface that all aggregate roots must implement
type Root interface {
    // GetID returns the aggregate's unique identifier
    GetID() uuidv7.UUID

    // GetVersion returns the aggregate's version (for optimistic locking)
    GetVersion() int

    // IncrementVersion increments the version (called after successful persistence)
    IncrementVersion()
}

// BaseAggregate provides common fields for all aggregates
type BaseAggregate struct {
    ID        uuidv7.UUID
    Version   int
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (a *BaseAggregate) GetID() uuidv7.UUID {
    return a.ID
}

func (a *BaseAggregate) GetVersion() int {
    return a.Version
}

func (a *BaseAggregate) IncrementVersion() {
    a.Version++
    a.UpdatedAt = time.Now()
}

// NewBaseAggregate creates a new base aggregate with generated ID
func NewBaseAggregate() BaseAggregate {
    now := time.Now()
    return BaseAggregate{
        ID:        uuidv7.New(),
        Version:   1,
        CreatedAt: now,
        UpdatedAt: now,
    }
}
```

**Tests:** `pkg/aggregate/aggregate_test.go` (15 tests)

---

#### Task 1.2: Create `pkg/valueobject` package

**Files:**

- `pkg/valueobject/money.go`
- `pkg/valueobject/address.go`
- `pkg/valueobject/contact_info.go`
- `pkg/valueobject/email.go`
- `pkg/valueobject/phone.go`

**Example: Money**

```go
package valueobject

import (
    "fmt"
    "math"
)

// Money represents a monetary value with currency
type Money struct {
    Amount   int64  // stored in smallest unit (cents, kopiykas, etc.)
    Currency string // ISO 4217 code (USD, UAH, EUR)
}

// NewMoney creates a Money value object
func NewMoney(amount int64, currency string) (Money, error) {
    if currency == "" {
        return Money{}, fmt.Errorf("currency is required")
    }
    if len(currency) != 3 {
        return Money{}, fmt.Errorf("currency must be 3-letter ISO code")
    }
    return Money{Amount: amount, Currency: currency}, nil
}

// FromFloat converts float to Money (handles rounding)
func FromFloat(amount float64, currency string) (Money, error) {
    cents := int64(math.Round(amount * 100))
    return NewMoney(cents, currency)
}

// ToFloat converts Money to float (for display)
func (m Money) ToFloat() float64 {
    return float64(m.Amount) / 100.0
}

// Add adds two Money values (same currency)
func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
    }
    return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Multiply multiplies Money by a scalar
func (m Money) Multiply(factor float64) Money {
    return Money{
        Amount:   int64(math.Round(float64(m.Amount) * factor)),
        Currency: m.Currency,
    }
}

// IsZero checks if amount is zero
func (m Money) IsZero() bool {
    return m.Amount == 0
}

// IsPositive checks if amount is positive
func (m Money) IsPositive() bool {
    return m.Amount > 0
}

// IsNegative checks if amount is negative
func (m Money) IsNegative() bool {
    return m.Amount < 0
}

// String returns formatted string (e.g., "$10.50")
func (m Money) String() string {
    return fmt.Sprintf("%s %.2f", m.Currency, m.ToFloat())
}
```

**Tests:** `pkg/valueobject/*_test.go` (60 tests total across all value objects)

---

#### Task 1.3: Create `pkg/saga` package

**File:** `pkg/saga/saga.go`

```go
package saga

import (
    "context"
    "fmt"
    "log/slog"
)

// Step represents a single step in a saga
type Step struct {
    Name       string
    Execute    func(ctx context.Context, data interface{}) error
    Compensate func(ctx context.Context, data interface{}) error
}

// Saga represents a distributed transaction
type Saga struct {
    Name  string
    Steps []Step
    log   *slog.Logger
}

// NewSaga creates a new saga
func NewSaga(name string, steps []Step) *Saga {
    return &Saga{
        Name:  name,
        Steps: steps,
        log:   slog.Default(),
    }
}

// Execute runs the saga steps
func (s *Saga) Execute(ctx context.Context, data interface{}) error {
    executed := []Step{}

    s.log.Info("Starting saga", "saga", s.Name)

    for i, step := range s.Steps {
        s.log.Info("Executing step", "saga", s.Name, "step", step.Name, "index", i+1, "total", len(s.Steps))

        if err := step.Execute(ctx, data); err != nil {
            s.log.Error("Step failed, starting compensation", "saga", s.Name, "step", step.Name, "error", err)

            // Compensate in reverse order
            for j := len(executed) - 1; j >= 0; j-- {
                if executed[j].Compensate != nil {
                    s.log.Info("Compensating step", "saga", s.Name, "step", executed[j].Name)
                    if compErr := executed[j].Compensate(ctx, data); compErr != nil {
                        s.log.Error("Compensation failed", "saga", s.Name, "step", executed[j].Name, "error", compErr)
                    }
                }
            }

            return fmt.Errorf("saga %s failed at step %s: %w", s.Name, step.Name, err)
        }

        executed = append(executed, step)
    }

    s.log.Info("Saga completed successfully", "saga", s.Name)
    return nil
}

// Builder helps construct sagas fluently
type Builder struct {
    name  string
    steps []Step
}

// NewBuilder creates a saga builder
func NewBuilder(name string) *Builder {
    return &Builder{name: name, steps: []Step{}}
}

// AddStep adds a step to the saga
func (b *Builder) AddStep(name string, execute func(ctx context.Context, data interface{}) error, compensate func(ctx context.Context, data interface{}) error) *Builder {
    b.steps = append(b.steps, Step{
        Name:       name,
        Execute:    execute,
        Compensate: compensate,
    })
    return b
}

// Build creates the saga
func (b *Builder) Build() *Saga {
    return NewSaga(b.name, b.steps)
}
```

**Tests:** `pkg/saga/saga_test.go` (30 tests - success, failure, compensation)

---

### Day 2: Context Structure

#### Task 2.1: Create `internal/contexts` directory structure

```bash
mkdir -p internal/contexts/{identity,customer-mgmt,order-mgmt,billing,warehouse}
mkdir -p internal/contexts/identity/{user,profile,contact}
mkdir -p internal/contexts/customer-mgmt/{customer,company,deal,interaction}
mkdir -p internal/contexts/order-mgmt/{order,contract,fulfillment}
mkdir -p internal/contexts/billing/{subscription,invoice,payment}
mkdir -p internal/contexts/warehouse/{product,inventory,location}
```

#### Task 2.2: Create context README files

**File:** `internal/contexts/README.md`

```markdown
# Bounded Contexts

This directory contains the application's bounded contexts following Domain-Driven Design principles.

## Architecture

### What is a Bounded Context?

A bounded context is an explicit boundary within which a domain model is defined and applicable. It represents a specific area of the business with its own ubiquitous language.

### Context Structure

Each context follows this structure:
```

/context-name/
/aggregate-name/
entity.go # Domain entity (aggregate root)
repository.go # Repository interface
usecase.go # Business logic
handler.go # HTTP handlers
dto.go # Data transfer objects
repository_impl.go # Repository implementation

````

## Available Contexts

### 1. Identity Context
**Domain:** User authentication, profiles, contacts
**Aggregates:** User, Profile, Contact
**Responsibilities:**
- User registration/login
- Profile management
- Contact information

### 2. Customer Management Context
**Domain:** CRM, customer relationships
**Aggregates:** Customer, Company, Deal, Interaction
**Responsibilities:**
- Customer lifecycle
- Company management (B2B)
- Deal/opportunity tracking
- Interaction history

### 3. Order Management Context
**Domain:** Order processing, contracts, fulfillment
**Aggregates:** Order, Contract, Fulfillment
**Responsibilities:**
- Order creation and lifecycle
- Contract generation
- Fulfillment tracking

### 4. Billing Context
**Domain:** Subscriptions, invoicing, payments
**Aggregates:** Subscription, Invoice, Payment
**Responsibilities:**
- Subscription management
- Invoice generation
- Payment processing

### 5. Warehouse Context
**Domain:** Inventory management
**Aggregates:** Product, Inventory, Location
**Responsibilities:**
- Product catalog
- Stock management
- Warehouse locations

## Communication Between Contexts

### 1. Domain Events (Preferred)
Contexts communicate via events through the event bus:

```go
// Customer created → trigger welcome email
bus.Publish("customer.created", CustomerCreatedEvent{CustomerID: id})
````

### 2. Anti-Corruption Layer

When one context needs data from another, use an anti-corruption layer:

```go
// OrderContext needs customer data from CustomerContext
type CustomerService interface {
    GetCustomer(ctx context.Context, id uuid.UUID) (CustomerDTO, error)
}
```

### 3. Saga Pattern

For distributed transactions across contexts:

```go
// CreateOrderSaga coordinates: Customer validation + Inventory reservation + Payment
saga := NewCreateOrderSaga()
saga.Execute(ctx, orderData)
```

## Migration Strategy

### Existing Modules

Existing modules (`internal/modules/*`) remain unchanged during migration:

- `posts` → stays as module
- `profiles` → gradually migrate to `identity` context
- `billing` → gradually migrate to `billing` context
- `workflows`, `notifications`, `audit`, `analytics` → stay as modules

### Backward Compatibility

- Old routes continue to work
- New contexts expose new routes
- Gradual migration over 2-3 months

````

---

### Day 3: Identity Context (Reference Implementation)

#### Task 3.1: Create Contact aggregate in Identity context

**File:** `internal/contexts/identity/contact/entity.go`

```go
package contact

import (
    "fmt"
    "github.com/basilex/promenade/pkg/aggregate"
    "github.com/basilex/promenade/pkg/valueobject"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// ContactType represents the type of contact
type ContactType string

const (
    ContactTypeEmail     ContactType = "email"
    ContactTypePhone     ContactType = "phone"
    ContactTypeAddress   ContactType = "address"
    ContactTypeSocial    ContactType = "social"
    ContactTypeMessenger ContactType = "messenger"
)

// Contact is an aggregate root representing a contact method
type Contact struct {
    aggregate.BaseAggregate

    // Identity
    UserID uuidv7.UUID

    // Contact details
    Type       ContactType
    Value      string
    Label      string // "Work", "Home", "Mobile", etc.
    IsPrimary  bool
    IsVerified bool

    // Privacy
    IsPublic bool
}

// NewContact creates a new contact
func NewContact(userID uuidv7.UUID, contactType ContactType, value string, label string) (*Contact, error) {
    if err := validateContactType(contactType); err != nil {
        return nil, err
    }

    if err := validateContactValue(contactType, value); err != nil {
        return nil, err
    }

    return &Contact{
        BaseAggregate: aggregate.NewBaseAggregate(),
        UserID:        userID,
        Type:          contactType,
        Value:         value,
        Label:         label,
        IsPrimary:     false,
        IsVerified:    false,
        IsPublic:      false,
    }, nil
}

// SetAsPrimary marks this contact as primary
func (c *Contact) SetAsPrimary() {
    c.IsPrimary = true
    c.IncrementVersion()
}

// Verify marks the contact as verified
func (c *Contact) Verify() error {
    if c.IsVerified {
        return fmt.Errorf("contact already verified")
    }
    c.IsVerified = true
    c.IncrementVersion()
    return nil
}

// SetPublic changes visibility
func (c *Contact) SetPublic(isPublic bool) {
    c.IsPublic = isPublic
    c.IncrementVersion()
}

// Validation helpers
func validateContactType(t ContactType) error {
    switch t {
    case ContactTypeEmail, ContactTypePhone, ContactTypeAddress, ContactTypeSocial, ContactTypeMessenger:
        return nil
    default:
        return fmt.Errorf("invalid contact type: %s", t)
    }
}

func validateContactValue(t ContactType, value string) error {
    if value == "" {
        return fmt.Errorf("contact value is required")
    }

    switch t {
    case ContactTypeEmail:
        email, err := valueobject.NewEmail(value)
        if err != nil {
            return err
        }
        _ = email
    case ContactTypePhone:
        phone, err := valueobject.NewPhone(value)
        if err != nil {
            return err
        }
        _ = phone
    }

    return nil
}
````

**Tests:** `internal/contexts/identity/contact/entity_test.go` (40 tests)

---

### Day 4-5: Documentation & Examples

#### Task 4.1: Create comprehensive documentation

**Files:**

- `docs/BOUNDED_CONTEXTS_GUIDE.md` - Complete guide
- `docs/AGGREGATE_PATTERN.md` - Aggregate root pattern
- `docs/SAGA_PATTERN.md` - Distributed transactions
- `docs/VALUE_OBJECTS_GUIDE.md` - Value objects reference

#### Task 4.2: Create example implementations

**File:** `examples/crm_context_example/main.go`

Full working example showing:

- Creating customer aggregate
- Using value objects (Money, Address, Email)
- Saga pattern for complex operation
- Event-driven communication

---

### Day 6-7: Testing & Validation

#### Task 6.1: Comprehensive tests for all packages

**Test Coverage Goals:**

- `pkg/aggregate`: 15 tests 
- `pkg/valueobject`: 60 tests 
- `pkg/saga`: 30 tests 
- `internal/contexts/identity/contact`: 40 tests 
- **TOTAL:** 145 tests

#### Task 6.2: Integration tests

**File:** `test/integration/contexts_test.go`

Test inter-context communication:

- Event-driven communication
- Saga pattern with compensation
- Anti-corruption layer

---

## Success Criteria

### Phase 1 Complete When:

-  **DONE:** `pkg/aggregate` package created with tests (29 tests, target: 15)
-  **DONE:** `pkg/valueobject` package created with tests (184 tests, target: 60)
-  **DONE:** `pkg/saga` package created with tests (43 tests, target: 30)
- ⏳ **TODO:** `internal/contexts/` structure created
- ⏳ **TODO:** Identity context with Contact aggregate implemented (40 tests)
- ⏳ **TODO:** Documentation complete (4 guides)
- ⏳ **TODO:** Example implementation working
- ⏳ **TODO:** All tests passing (target: 145 tests, currently: 256 tests)
- ⏳ **TODO:** CI/CD updated

### Deliverables:

1. **Code:**

   - 3 new `pkg/*` packages
   - New `internal/contexts/` structure
   - 1 reference implementation (Contact)
   - 145 tests (100% passing)

2. **Documentation:**

   - BOUNDED_CONTEXTS_GUIDE.md
   - AGGREGATE_PATTERN.md
   - SAGA_PATTERN.md
   - VALUE_OBJECTS_GUIDE.md
   - internal/contexts/README.md

3. **Examples:**
   - examples/crm_context_example/
   - Working demonstration of new patterns

---

## Next Steps After Phase 1

### Phase 2: Customer Management Context (3 weeks)

With foundation ready, implement full CRM:

- Customer aggregate (60 tests)
- Company aggregate (50 tests)
- Deal aggregate (60 tests)
- Interaction aggregate (40 tests)
- Handlers + Integration (50 tests)
- **TOTAL:** 260 tests

---

## Daily Checklist

### Day 1: Foundation  COMPLETE (December 27, 2025)

- [x] Create `pkg/aggregate` package
- [x] Write 15 aggregate tests
- [x] Create `pkg/valueobject` package
- [x] Implement Money, Address, Email, Phone value objects
- [x] Write 184 value object tests (exceeded target of 60!)
- [x] Create `pkg/saga` package
- [x] Write 43 saga tests (exceeded target of 30!)

**Results:**

- **pkg/aggregate:** 84 lines code, 29 tests (100% pass)
- **pkg/valueobject:** 574 lines code, 184 tests (100% pass)
- **pkg/saga:** 164 lines code, 43 tests (100% pass)
- **TOTAL:** 822 lines code, 256 tests (exceeded 90 test target by 166 tests!)
- **Test Coverage:** Comprehensive (Success paths, failure paths, edge cases, real-world scenarios)

### Day 2: Structure  COMPLETE (December 27, 2025)

- [x] Create `internal/contexts/` directories
- [x] Write context READMEs (6 comprehensive READMEs created)
- [ ] Setup CI/CD for new structure

**Results:**

- **internal/contexts/README.md:** ~450 lines (DDD guide, patterns, communication)
- **internal/contexts/identity/README.md:** ~600 lines (User, Profile, Contact aggregates)
- **internal/contexts/customer-mgmt/README.md:** ~500 lines (Customer, Company, Deal, Interaction)
- **internal/contexts/order-mgmt/README.md:** ~450 lines (Order, Contract, Fulfillment + Saga pattern)
- **internal/contexts/billing/README.md:** ~350 lines (Migration plan from modules/billing)
- **internal/contexts/warehouse/README.md:** ~500 lines (Product, Inventory, Location)
- **Total:** ~2,850 lines of comprehensive context documentation!

### Day 3: Reference Implementation

- [ ] Implement Contact aggregate
- [ ] Write 40 Contact tests
- [ ] Verify aggregate pattern works

### Day 4-5: Documentation

- [ ] Write BOUNDED_CONTEXTS_GUIDE.md
- [ ] Write AGGREGATE_PATTERN.md
- [ ] Write SAGA_PATTERN.md
- [ ] Write VALUE_OBJECTS_GUIDE.md
- [ ] Create working example

### Day 6-7: Testing & Polish

- [ ] Integration tests
- [ ] CI/CD validation
- [ ] Code review
- [ ] Phase 1 retrospective

---

**Start Date:** December 27, 2025  
**Target Completion:** January 3, 2026  
**Status:**  READY TO START

**First Task:** Create `pkg/aggregate/aggregate.go` package! 
