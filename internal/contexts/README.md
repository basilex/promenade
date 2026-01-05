**Navigation**: [Home](../../README.md) > [Internal](../README.md) > Bounded Contexts

---

# Bounded Contexts

This directory contains the application's **bounded contexts** following Domain-Driven Design (DDD) principles.

**See Also**: [Bounded Contexts Strategy](../../docs/concepts/bounded-contexts.md) - Complete architectural guide

## What is a Bounded Context?

A **bounded context** is an explicit boundary within which a domain model is defined and applicable. It represents a specific area of the business with its own:

- **Ubiquitous Language** - Common terminology used by developers and domain experts
- **Domain Model** - Entities, value objects, aggregates specific to this context
- **Business Rules** - Logic that applies within this context
- **Consistency Boundaries** - Transactional boundaries (aggregates)

**Key Principle**: Contexts communicate ONLY via Event Bus (no direct dependencies)

## Architecture

### Context vs Module

**Modules** (`internal/modules/*`) are technical features that can be enabled/disabled:

- Posts, Profiles, Analytics, Workflows, Billing
- Plugin-based architecture
- Licensable, independent
- HTTP-focused (handlers, routes, DTOs)

**Contexts** (`internal/contexts/*`) are business domains:

- Identity, Customer Management, Order Management, Billing, Warehouse
- DDD-based architecture
- Always enabled (if implemented)
- Domain-focused (aggregates, entities, value objects)

### Why Both?

- **Modules** = Backward compatibility + existing features
- **Contexts** = New CRM features with complex business logic
- They can coexist: Billing module → Billing context migration path

## Context Structure

Each context follows this structure:

```
/context-name/
   README.md                    # Context documentation
   /aggregate-name/             # Each aggregate root
       entity.go                # Domain entity (aggregate root)
       value_objects.go         # Context-specific value objects
       repository.go            # Repository interface
       repository_impl.go       # Repository implementation (Postgres)
       usecase.go               # Business logic (use cases)
       handler.go               # HTTP handlers
       dto.go                   # Data transfer objects
```

**Example:**

```
/customer-mgmt/
   README.md
   /customer/
      customer.go              # Customer aggregate root
      repository.go
      repository_impl.go
      usecase.go
      handler.go
   /company/
      company.go               # Company aggregate root
      ...
   /deal/
       deal.go                  # Deal aggregate root
       ...
```

## Available Contexts

### 1. Identity Context

**Domain:** User authentication, profiles, contacts  
**Directory:** `internal/contexts/identity/`  
**Aggregates:**

- **User** - Authentication, credentials, sessions
- **Profile** - User profile information (personal, business)
- **Contact** - Contact information (email, phone, address)

**Responsibilities:**

- User registration and login
- Profile management
- Contact information management
- Session management

**Boundaries:**

- Does NOT handle billing or subscriptions
- Does NOT handle customer relationships (CRM)
- Only manages authenticated user data

---

### 2. Customer Management Context

**Domain:** CRM, customer relationships  
**Directory:** `internal/contexts/customer-mgmt/`  
**Aggregates:**

- **Customer** - Customer lifecycle (B2C or B2B contact person)
- **Company** - Company management (B2B organizations)
- **Deal** - Sales opportunities and pipeline
- **Interaction** - Customer interaction history

**Responsibilities:**

- Customer lifecycle management (lead → customer → churned)
- Company management for B2B scenarios
- Deal/opportunity tracking
- Interaction history (calls, emails, meetings)

**Boundaries:**

- Does NOT handle orders (that's Order Management)
- Does NOT handle invoices (that's Billing)
- Focuses on relationships, not transactions

---

### 3. Order Management Context

**Domain:** Order processing, contracts, fulfillment  
**Directory:** `internal/contexts/order-mgmt/`  
**Aggregates:**

- **Order** - Order lifecycle (created → paid → fulfilled)
- **Contract** - Legal agreements
- **Fulfillment** - Order fulfillment process

**Responsibilities:**

- Order creation and lifecycle
- Contract generation and management
- Fulfillment tracking
- Order status updates

**Boundaries:**

- Does NOT handle payment processing (that's Billing)
- Does NOT handle inventory (that's Warehouse)
- Coordinates between Customer, Billing, and Warehouse contexts

---

### 4. Billing Context

**Domain:** Subscriptions, invoicing, payments  
**Directory:** `internal/contexts/billing/`  
**Aggregates:**

- **Subscription** ✅ - Subscription lifecycle (trial, active, paused, cancelled, expired)
- **Invoice** ✅ - Invoice generation and management
- **Payment** ✅ - Payment processing and reconciliation

**Responsibilities:**

- Subscription management
- Invoice generation
- Payment processing
- Revenue recognition

**Boundaries:**

- Does NOT handle customer relationships (that's Customer Management)
- Does NOT handle order fulfillment (that's Order Management)
- Focuses on financial transactions

**Note:** Currently implemented as `internal/modules/billing/` - will be gradually migrated to this context structure.

---

### 5. Warehouse Context

**Domain:** Inventory management, products  
**Directory:** `internal/contexts/warehouse/`  
**Aggregates:**

- **Product** - Product catalog
- **Inventory** - Stock levels and movements
- **Location** - Warehouse locations

**Responsibilities:**

- Product catalog management
- Inventory tracking (stock levels, reservations)
- Warehouse location management
- Stock movements (in, out, transfers)

**Boundaries:**

- Does NOT handle pricing (that's Billing)
- Does NOT handle orders (that's Order Management)
- Focuses on physical goods and availability

---

## Communication Between Contexts

### Anti-Corruption Layer (ACL)

Each context is isolated and uses an **Anti-Corruption Layer** to communicate:

```go
// Example: Order context needs customer data
//  WRONG: Direct dependency
order := Order{
    CustomerID: customer.ID, // Direct coupling
    Customer: customer,       // Leaking domain model
}

//  CORRECT: Through ACL
type CustomerReference struct {
    ID    uuidv7.UUID
    Name  string
    Email string
}

order := Order{
    CustomerRef: acl.GetCustomerReference(customerID), // ACL translates
}
```

### Event-Driven Communication

Contexts communicate asynchronously via domain events:

```go
// Customer context publishes
eventBus.Publish("customer.created", CustomerCreatedEvent{
    CustomerID: customer.ID,
    Email:      customer.Email,
})

// Billing context subscribes
eventBus.Subscribe("customer.created", func(event CustomerCreatedEvent) {
    // Create default subscription for new customer
    billingService.CreateDefaultSubscription(event.CustomerID)
})
```

### Saga Pattern for Distributed Transactions

For operations spanning multiple contexts, use **Saga pattern**:

```go
saga := saga.NewBuilder("create-order").
    Step("validate-customer", customerMgmt.ValidateCustomer, nil).
    Step("reserve-inventory", warehouse.ReserveInventory, warehouse.ReleaseInventory).
    Step("create-invoice", billing.CreateInvoice, billing.CancelInvoice).
    Step("process-payment", billing.ProcessPayment, billing.RefundPayment).
    Step("create-order", orderMgmt.CreateOrder, orderMgmt.CancelOrder).
    Build()

result := saga.ExecuteWithResult(ctx)
```

## Shared Kernel

Value objects shared across all contexts live in `pkg/valueobject/`:

- **Money** - Monetary amounts with currency
- **Address** - Physical addresses
- **Email** - Email addresses with validation
- **Phone** - Phone numbers in E.164 format

## Development Guidelines

### 1. Aggregate Rules

- **One aggregate per transaction** - Don't modify multiple aggregates in one transaction
- **Enforce invariants** - Aggregates validate business rules internally
- **Reference by ID** - Aggregates reference other aggregates by ID, not object reference

### 2. Repository Pattern

```go
// Interface in domain layer
type CustomerRepository interface {
    GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error)
    Save(ctx context.Context, customer *Customer) error
    Delete(ctx context.Context, id uuidv7.UUID) error
}

// Implementation in infrastructure layer
type PostgresCustomerRepository struct {
    db *sqlx.DB
}
```

### 3. Use Case Pattern

```go
type CreateCustomerUseCase struct {
    customerRepo CustomerRepository
    eventBus     bus.IBus
}

func (uc *CreateCustomerUseCase) Execute(ctx context.Context, req CreateCustomerRequest) (*Customer, error) {
    // 1. Validate business rules
    // 2. Create aggregate
    customer := NewCustomer(req.Name, req.Email)

    // 3. Persist
    if err := uc.customerRepo.Save(ctx, customer); err != nil {
        return nil, err
    }

    // 4. Publish domain event
    uc.eventBus.Publish(ctx, "customer.created", CustomerCreatedEvent{
        CustomerID: customer.ID,
    })

    return customer, nil
}
```

## Testing Strategy

### Aggregate Tests

Test domain logic in isolation:

```go
func TestCustomer_AddContact(t *testing.T) {
    customer := NewCustomer("Test Customer", "test@example.com")
    contact := NewContact("Primary", "test@example.com", "+1234567890")

    err := customer.AddContact(contact)
    assert.NoError(t, err)
    assert.Len(t, customer.Contacts(), 1)
}
```

### Repository Tests

Integration tests with real database:

```go
func TestPostgresCustomerRepository_Save(t *testing.T) {
    db := setupTestDB(t)
    repo := NewPostgresCustomerRepository(db)

    customer := NewCustomer("Test", "test@example.com")
    err := repo.Save(context.Background(), customer)
    assert.NoError(t, err)
}
```

### Use Case Tests

Test business logic with mocks:

```go
func TestCreateCustomerUseCase_Execute(t *testing.T) {
    mockRepo := new(MockCustomerRepository)
    mockBus := new(MockEventBus)
    uc := NewCreateCustomerUseCase(mockRepo, mockBus)

    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
    mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

    customer, err := uc.Execute(context.Background(), CreateCustomerRequest{
        Name: "Test", Email: "test@example.com",
    })

    assert.NoError(t, err)
    assert.NotNil(t, customer)
}
```

## Migration Strategy

### From Modules to Contexts

Existing modules will gradually migrate to contexts:

1. **Keep existing module working** - No breaking changes
2. **Implement context in parallel** - New features use context structure
3. **Migrate routes gradually** - Route `/api/v1/billing/...` → context handler
4. **Deprecate module code** - Once all routes migrated
5. **Remove module** - After deprecation period

**Example: Billing migration**

```
Phase 1: internal/modules/billing/ (existing)
Phase 2: internal/contexts/billing/ (new structure)
Phase 3: Routes use context handlers
Phase 4: Remove module code
```

## Related Documentation

- [Main README](../../README.md) - Project overview
- [Documentation Index](../../docs/INDEX.md) - Complete documentation catalog
- [Bounded Contexts Strategy](../../docs/concepts/bounded-contexts.md) - Complete architectural guide
- [Clean Architecture with DDD](../../docs/concepts/clean-architecture.md) - Architecture principles
- [Event-Driven Architecture](../../docs/concepts/event-driven.md) - Event Bus and communication

**Context Documentation**:
- [Identity Context](identity/README.md) - User, Contact, Profile, RBAC
- [Shared Context](shared/README.md) - Reference data (Country, Currency, Language, Timezone)
- [Customer Management](customer-mgmt/README.md) - Customer aggregate

**Packages**:
- [Package Overview](../../pkg/README.md) - All shared packages
- [Event Bus](../../pkg/bus/README.md) - Event Bus implementation
- [Aggregate Pattern](../../pkg/aggregate/README.md) - Base aggregate
- [Value Objects](../../pkg/valueobject/README.md) - Shared value objects
- [Saga Pattern](../../pkg/saga/README.md) - Distributed transactions

---

**Status:**  Ready for implementation  
**Next Steps:** Implement Contact aggregate in Identity context  
**Target:** Full CRM system with Customer Management context
