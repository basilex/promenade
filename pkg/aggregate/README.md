# Aggregate Package

**Purpose:** Base Domain-Driven Design aggregate pattern  
**Status:** Production-ready  
**Tests:** 5 tests, 100% coverage

---

## Overview

The `aggregate` package provides **base types and interfaces** for implementing Domain-Driven Design aggregates. Aggregates are clusters of domain objects treated as a single unit with transactional consistency boundaries.

## Features

- **Root Interface:** Standard interface for all aggregate roots
- **BaseAggregate:** Common fields (ID, Version, timestamps)
- **Optimistic Locking:** Version-based concurrency control
- **Immutable IDs:** UUID v7 time-ordered identifiers

---

## Installation

```go
import "github.com/basilex/promenade/pkg/aggregate"
```

---

## Quick Start

```go
package user

import (
    "github.com/basilex/promenade/pkg/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type User struct {
    aggregate.BaseAggregate
    Email    string
    Name     string
    Status   UserStatus
}

func NewUser(email, name string) *User {
    return &User{
        BaseAggregate: aggregate.NewBase(),
        Email:         email,
        Name:          name,
        Status:        UserStatusPending,
    }
}
```

---

## Core Concepts

### What is an Aggregate?

An **aggregate** is a cluster of domain objects that can be treated as a single unit:

- **Aggregate Root:** Entry point with global identity (e.g., User, Order, Customer)
- **Entities:** Objects with identity within aggregate
- **Value Objects:** Immutable objects defined by attributes
- **Invariants:** Business rules enforced within aggregate boundary
- **Transactional Boundary:** Changes committed atomically

### Aggregate Root Pattern

```
┌─────────────────────────────────────┐
│         User (Root)                 │
│  - ID: UUID                         │
│  - Email: string                    │
│  - Version: int                     │
│                                     │
│  ┌─────────────────────────────┐   │
│  │   Contacts (Entities)       │   │
│  │  - Email contacts           │   │
│  │  - Phone contacts           │   │
│  │  - Address contacts         │   │
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │   Profile (Value Object)    │   │
│  │  - DisplayName              │   │
│  │  - Bio                      │   │
│  │  - AvatarURL                │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

**Rules:**
- External objects can only reference aggregate root (User)
- Internal entities (Contacts) accessed through root
- All changes go through root to maintain invariants
- Root coordinates transactional consistency

---

## Usage Examples

### 1. Define Aggregate Root

```go
package customer

import (
    "github.com/basilex/promenade/pkg/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/pkg/valueobject"
)

type Customer struct {
    aggregate.BaseAggregate
    
    // Identity
    Email valueobject.Email
    
    // Personal Info
    FirstName string
    LastName  string
    
    // Business State
    Status    CustomerStatus
    Segment   CustomerSegment
    
    // Relationships (owned entities)
    Contacts  []*Contact
    Addresses []*Address
}

func NewCustomer(email, firstName, lastName string) (*Customer, error) {
    emailVO, err := valueobject.NewEmail(email)
    if err != nil {
        return nil, err
    }
    
    return &Customer{
        BaseAggregate: aggregate.NewBase(),
        Email:         emailVO,
        FirstName:     firstName,
        LastName:      lastName,
        Status:        CustomerStatusProspect,
        Segment:       CustomerSegmentStandard,
        Contacts:      make([]*Contact, 0),
        Addresses:     make([]*Address, 0),
    }, nil
}
```

### 2. Implement Business Methods

```go
// Activate transitions customer from Prospect to Active
func (c *Customer) Activate() error {
    if c.Status != CustomerStatusProspect {
        return fmt.Errorf("only prospects can be activated")
    }
    
    c.Status = CustomerStatusActive
    c.IncrementVersion() // Optimistic locking
    c.UpdateTimestamp()
    
    return nil
}

// AddContact adds a new contact to customer
func (c *Customer) AddContact(contact *Contact) error {
    // Validate business rule: max 10 contacts
    if len(c.Contacts) >= 10 {
        return fmt.Errorf("customer can have maximum 10 contacts")
    }
    
    c.Contacts = append(c.Contacts, contact)
    c.IncrementVersion()
    c.UpdateTimestamp()
    
    return nil
}

// UpgradeSegment changes customer segment
func (c *Customer) UpgradeSegment(segment CustomerSegment) error {
    if segment <= c.Segment {
        return fmt.Errorf("can only upgrade to higher segment")
    }
    
    c.Segment = segment
    c.IncrementVersion()
    c.UpdateTimestamp()
    
    return nil
}
```

### 3. Optimistic Locking

```go
// Repository Update with version check
func (r *customerRepository) Update(ctx context.Context, customer *Customer) error {
    query := `
        UPDATE customers 
        SET 
            first_name = $1,
            last_name = $2,
            status = $3,
            segment = $4,
            version = version + 1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $5 AND version = $6
    `
    
    result, err := r.db.ExecContext(ctx, query,
        customer.FirstName,
        customer.LastName,
        customer.Status,
        customer.Segment,
        customer.ID,
        customer.Version, // Check current version
    )
    if err != nil {
        return err
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("version conflict: customer was modified")
    }
    
    customer.IncrementVersion() // Update in-memory version
    return nil
}
```

### 4. Use in Use Case

```go
package customer

import "context"

type useCase struct {
    repo IRepository
}

func (uc *useCase) ActivateCustomer(ctx context.Context, customerID uuidv7.UUID) error {
    // Load aggregate
    customer, err := uc.repo.GetByID(ctx, customerID)
    if err != nil {
        return err
    }
    
    // Execute business logic (through aggregate method)
    if err := customer.Activate(); err != nil {
        return err
    }
    
    // Save aggregate (optimistic locking)
    if err := uc.repo.Update(ctx, customer); err != nil {
        return err
    }
    
    return nil
}
```

---

## API Reference

### Root Interface

```go
// Root defines the interface that all aggregate roots must implement
type Root interface {
    GetID() uuidv7.UUID
    GetVersion() int
    IncrementVersion()
}
```

### BaseAggregate

```go
// BaseAggregate provides common fields for all aggregates
type BaseAggregate struct {
    ID        uuidv7.UUID
    Version   int
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NewBase creates a new BaseAggregate with UUID v7 and timestamps
func NewBase() BaseAggregate

// GetID returns the aggregate ID
func (b BaseAggregate) GetID() uuidv7.UUID

// GetVersion returns the current version
func (b BaseAggregate) GetVersion() int

// IncrementVersion increments the version (for optimistic locking)
func (b *BaseAggregate) IncrementVersion()

// UpdateTimestamp updates the UpdatedAt field
func (b *BaseAggregate) UpdateTimestamp()
```

---

## Best Practices

### DO

- **Embed BaseAggregate in all roots** - Standard ID, Version, timestamps
  ```go
  type Customer struct {
      aggregate.BaseAggregate  // Embed base
      Email    string
      Status   CustomerStatus
  }
  ```

- **Use factory methods** - Ensure valid initial state
  ```go
  func NewCustomer(email string) (*Customer, error) {
      return &Customer{
          BaseAggregate: aggregate.NewBase(),
          Email:         email,
          Status:        CustomerStatusProspect,
      }, nil
  }
  ```

- **Encapsulate business logic** - All changes through aggregate methods
  ```go
  // Good: Business logic in aggregate
  func (c *Customer) Activate() error {
      if c.Status != CustomerStatusProspect {
          return fmt.Errorf("invalid state transition")
      }
      c.Status = CustomerStatusActive
      c.IncrementVersion()
      return nil
  }
  
  // Bad: Direct state modification
  customer.Status = CustomerStatusActive // Bypasses business rules
  ```

- **Increment version on changes** - Enable optimistic locking
  ```go
  func (c *Customer) UpdateEmail(email string) {
      c.Email = email
      c.IncrementVersion()    // Always increment
      c.UpdateTimestamp()     // Update timestamp
  }
  ```

### DON'T

- **Don't expose internal entities** - Access through aggregate root
  ```go
  // Bad: Direct access to internal entity
  customer.Contacts[0].Email = "new@email.com"
  
  // Good: Method on aggregate root
  customer.UpdateContactEmail(contactID, "new@email.com")
  ```

- **Don't reference other aggregates directly** - Use IDs
  ```go
  // Bad: Direct aggregate reference
  type Order struct {
      Customer *Customer // Don't hold reference
  }
  
  // Good: Reference by ID
  type Order struct {
      CustomerID uuidv7.UUID // Just the ID
  }
  ```

- **Don't forget version checks** - Prevent lost updates
  ```go
  // Bad: No version check
  UPDATE customers SET status = $1 WHERE id = $2
  
  // Good: Optimistic locking
  UPDATE customers SET status = $1, version = version + 1 
  WHERE id = $2 AND version = $3
  ```

---

## Aggregate Design Guidelines

### 1. Keep Aggregates Small

```go
// Good: Focused aggregate
type Customer struct {
    aggregate.BaseAggregate
    Email     string
    FirstName string
    LastName  string
    Status    CustomerStatus
}

// Bad: Too many responsibilities
type Customer struct {
    aggregate.BaseAggregate
    // ... customer fields ...
    Orders    []*Order      // Should be separate aggregate
    Invoices  []*Invoice    // Should be separate aggregate
    Payments  []*Payment    // Should be separate aggregate
}
```

### 2. Enforce Invariants

```go
func (c *Customer) SetStatus(status CustomerStatus) error {
    // Invariant: Valid state transitions
    validTransitions := map[CustomerStatus][]CustomerStatus{
        CustomerStatusProspect: {CustomerStatusActive},
        CustomerStatusActive:   {CustomerStatusSuspended, CustomerStatusChurned},
        CustomerStatusSuspended: {CustomerStatusActive, CustomerStatusChurned},
    }
    
    allowed := validTransitions[c.Status]
    for _, valid := range allowed {
        if status == valid {
            c.Status = status
            c.IncrementVersion()
            return nil
        }
    }
    
    return fmt.Errorf("invalid state transition: %s -> %s", c.Status, status)
}
```

### 3. Transaction Boundaries

One aggregate = One transaction:

```go
// Good: Single aggregate per transaction
func (uc *useCase) ActivateCustomer(ctx context.Context, id uuidv7.UUID) error {
    customer, _ := uc.repo.GetByID(ctx, id)
    customer.Activate()
    return uc.repo.Update(ctx, customer)
}

// Bad: Multiple aggregates in one transaction
func (uc *useCase) CreateOrderAndUpdateCustomer(ctx context.Context) error {
    tx, _ := uc.db.BeginTx(ctx)
    
    customer.UpdateStatus() // Customer aggregate
    order.Create()          // Order aggregate
    
    tx.Commit() // Multiple aggregates in one transaction - avoid this
}
```

---

## Complete Example

```go
package customer

import (
    "fmt"
    "github.com/basilex/promenade/pkg/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type CustomerStatus string

const (
    CustomerStatusProspect  CustomerStatus = "prospect"
    CustomerStatusActive    CustomerStatus = "active"
    CustomerStatusSuspended CustomerStatus = "suspended"
    CustomerStatusChurned   CustomerStatus = "churned"
)

type Customer struct {
    aggregate.BaseAggregate
    
    Email     string
    FirstName string
    LastName  string
    Status    CustomerStatus
}

func NewCustomer(email, firstName, lastName string) *Customer {
    return &Customer{
        BaseAggregate: aggregate.NewBase(),
        Email:         email,
        FirstName:     firstName,
        LastName:      lastName,
        Status:        CustomerStatusProspect,
    }
}

func (c *Customer) Activate() error {
    if c.Status != CustomerStatusProspect {
        return fmt.Errorf("can only activate prospects")
    }
    c.Status = CustomerStatusActive
    c.IncrementVersion()
    c.UpdateTimestamp()
    return nil
}

func (c *Customer) Suspend(reason string) error {
    if c.Status != CustomerStatusActive {
        return fmt.Errorf("can only suspend active customers")
    }
    c.Status = CustomerStatusSuspended
    c.IncrementVersion()
    c.UpdateTimestamp()
    return nil
}

func (c *Customer) Churn() error {
    if c.Status == CustomerStatusChurned {
        return fmt.Errorf("customer already churned")
    }
    c.Status = CustomerStatusChurned
    c.IncrementVersion()
    c.UpdateTimestamp()
    return nil
}
```

---

## Testing

```go
package customer_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCustomer_Activate(t *testing.T) {
    customer := NewCustomer("test@example.com", "John", "Doe")
    
    // Initial state
    assert.Equal(t, CustomerStatusProspect, customer.Status)
    assert.Equal(t, 1, customer.GetVersion())
    
    // Activate
    err := customer.Activate()
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, CustomerStatusActive, customer.Status)
    assert.Equal(t, 2, customer.GetVersion()) // Version incremented
}

func TestCustomer_Activate_InvalidState(t *testing.T) {
    customer := NewCustomer("test@example.com", "John", "Doe")
    customer.Status = CustomerStatusActive // Already active
    
    // Try to activate again
    err := customer.Activate()
    
    // Should fail
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "can only activate prospects")
}
```

---

## Related Packages

- `pkg/uuidv7` - Time-ordered UUID generation for aggregate IDs
- `pkg/valueobject` - Value objects used within aggregates
- `internal/contexts/*/` - Bounded contexts with aggregate implementations

---

**Last Updated:** 2025-12-28  
**Status:** Production-ready  
**Maintainer:** Promenade Team