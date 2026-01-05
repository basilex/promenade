# Billing Context - Implementation Plan

**Created:** January 4, 2026  
**Status:**  In Progress  
**Target:** Q2 2026 (Accelerated to Q1 2026)  
**Priority:** High

---

## Executive Summary

Implementing **Billing Context** as a new bounded context following all established patterns:
- Domain-Driven Design with Aggregates
- Multi-database support (PostgreSQL/SQLite)
- Event-Driven Architecture
- Complete CRUD + business logic
- Comprehensive testing (4-tier strategy)

**3 Core Aggregates**: Invoice, Payment, Subscription

---

## Architecture Overview

### Bounded Context Design

**Billing Context Responsibilities**:
-  Invoice generation and management
-  Payment processing and tracking
-  Subscription lifecycle management
-  Revenue recognition
-  Payment method management

**What Billing Does NOT Do**:
-  Customer relationships → Customer Management Context
-  Order fulfillment → Order Management Context
-  Product catalog → Warehouse Context

**Communication**:
- Listens: `order.confirmed` → Generate invoice
- Publishes: `invoice.generated`, `payment.received`, `subscription.created`

---

## Phase 1: Invoice Aggregate (Days 1-3)

### 1.1 Domain Model

**Entity**: `internal/contexts/billing/invoice/entity.go`

```go
type Invoice struct {
    aggregate.BaseAggregate
    
    // Identity
    ID          uuidv7.UUID
    InvoiceNo   string        // INV-2026-000001 (auto-generated)
    
    // Relations
    CustomerID  uuidv7.UUID   // FK to customer_customers
    OrderID     *uuidv7.UUID  // Optional FK to order_orders
    
    // Amounts (in cents, Money value object)
    SubtotalAmount Money      // Before tax
    TaxAmount      Money      // Tax
    TotalAmount    Money      // Final amount
    Currency       string      // USD, EUR, UAH
    
    // Dates
    IssueDate      time.Time
    DueDate        time.Time
    PaidDate       *time.Time
    
    // Status
    Status         InvoiceStatus // draft, sent, paid, overdue, cancelled, void
    
    // Line Items (embedded)
    Lines          []InvoiceLine
    
    // Audit
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time
}

type InvoiceLine struct {
    ID          uuidv7.UUID
    InvoiceID   uuidv7.UUID
    Description string
    Quantity    int
    UnitPrice   Money
    Amount      Money
}

type InvoiceStatus string

const (
    InvoiceStatusDraft     InvoiceStatus = "draft"
    InvoiceStatusSent      InvoiceStatus = "sent"
    InvoiceStatusPaid      InvoiceStatus = "paid"
    InvoiceStatusOverdue   InvoiceStatus = "overdue"
    InvoiceStatusCancelled InvoiceStatus = "cancelled"
    InvoiceStatusVoid      InvoiceStatus = "void"
)
```

**Factory Methods**:
```go
func NewInvoice(customerID uuidv7.UUID, dueDate time.Time) (*Invoice, error)
func (i *Invoice) AddLine(description string, qty int, unitPrice Money) error
func (i *Invoice) RemoveLine(lineID uuidv7.UUID) error
func (i *Invoice) MarkAsSent() error
func (i *Invoice) MarkAsPaid(paidDate time.Time) error
func (i *Invoice) MarkAsOverdue() error
func (i *Invoice) Cancel() error
func (i *Invoice) Void() error
func (i *Invoice) CalculateTotal() Money
```

**Business Rules**:
- Invoice number auto-generated: `INV-YYYY-NNNNNN`
- Cannot modify sent/paid invoices (must void and create new)
- Due date must be after issue date
- Total = subtotal + tax
- Overdue status auto-set if unpaid after due date
- Terminal states: paid, cancelled, void (immutable)

### 1.2 Database Schema

**Migration**: `migrations/billing/000001_invoices.up.sql`

```sql
-- Invoice table with multi-database support
CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY,
    invoice_no VARCHAR(20) NOT NULL UNIQUE,
    
    -- Relations
    customer_id UUID NOT NULL,
    order_id UUID,
    
    -- Amounts (cents)
    subtotal_amount BIGINT NOT NULL,
    tax_amount BIGINT NOT NULL,
    total_amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Dates
    issue_date TIMESTAMP NOT NULL,
    due_date TIMESTAMP NOT NULL,
    paid_date TIMESTAMP,
    
    -- Status
    status VARCHAR(20) NOT NULL,
    
    -- Audit
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_billing_invoices_status 
        CHECK (status IN ('draft', 'sent', 'paid', 'overdue', 'cancelled', 'void')),
    CONSTRAINT chk_billing_invoices_due_date 
        CHECK (due_date >= issue_date)
);

-- Invoice lines table
CREATE TABLE billing_invoice_lines (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- FK
    CONSTRAINT fk_billing_invoice_lines_invoice 
        FOREIGN KEY (invoice_id) REFERENCES billing_invoices(id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT chk_billing_invoice_lines_quantity 
        CHECK (quantity > 0)
);

-- Indexes
CREATE INDEX idx_billing_invoices_customer ON billing_invoices(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_invoices_order ON billing_invoices(order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_invoices_status ON billing_invoices(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_invoices_due_date ON billing_invoices(due_date) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_billing_invoices_invoice_no ON billing_invoices(invoice_no) WHERE deleted_at IS NULL;

CREATE INDEX idx_billing_invoice_lines_invoice ON billing_invoice_lines(invoice_id);
```

**SQLite Compatibility**:
- UUID stored as TEXT
- BIGINT → INTEGER
- JSONB → TEXT (if needed for metadata)

### 1.3 Repository Layer

**Interface**: `internal/contexts/billing/invoice/repository.go`

```go
type IRepository interface {
    Create(ctx context.Context, invoice *Invoice) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*Invoice, error)
    GetByInvoiceNo(ctx context.Context, invoiceNo string) (*Invoice, error)
    Update(ctx context.Context, invoice *Invoice) error
    Delete(ctx context.Context, id uuidv7.UUID) error
    
    // Queries
    ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Invoice, int, error)
    ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*Invoice, error)
    ListByStatus(ctx context.Context, status InvoiceStatus, page, pageSize int) ([]*Invoice, int, error)
    ListOverdue(ctx context.Context, page, pageSize int) ([]*Invoice, int, error)
    
    // Aggregations
    CountByStatus(ctx context.Context, status InvoiceStatus) (int, error)
    GetTotalRevenue(ctx context.Context, from, to time.Time) (Money, error)
}
```

**Implementation**: `internal/contexts/billing/invoice/adapter/repository/postgres/invoice_repository.go`

- Embed `BaseRepository`
- Use `jsonstore.Field[T]` if needed for metadata
- Soft delete with `deleted_at IS NULL`
- Multi-database SQL (? placeholders)

### 1.4 Use Case Layer

**Interface**: `internal/contexts/billing/invoice/usecase.go`

```go
type IUseCase interface {
    // CRUD
    CreateInvoice(ctx context.Context, customerID uuidv7.UUID, dueDate time.Time) (*Invoice, error)
    GetInvoice(ctx context.Context, invoiceID uuidv7.UUID) (*Invoice, error)
    GetInvoiceByNo(ctx context.Context, invoiceNo string) (*Invoice, error)
    DeleteInvoice(ctx context.Context, invoiceID uuidv7.UUID) error
    
    // Line Items
    AddLineItem(ctx context.Context, invoiceID uuidv7.UUID, desc string, qty int, unitPrice Money) error
    RemoveLineItem(ctx context.Context, invoiceID, lineID uuidv7.UUID) error
    
    // Status Changes
    SendInvoice(ctx context.Context, invoiceID uuidv7.UUID) error
    MarkAsPaid(ctx context.Context, invoiceID uuidv7.UUID, paidDate time.Time) error
    CancelInvoice(ctx context.Context, invoiceID uuidv7.UUID) error
    VoidInvoice(ctx context.Context, invoiceID uuidv7.UUID) error
    
    // Queries
    ListCustomerInvoices(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Invoice, int, error)
    ListOverdueInvoices(ctx context.Context, page, pageSize int) ([]*Invoice, int, error)
    
    // Business Logic
    GenerateInvoiceFromOrder(ctx context.Context, orderID uuidv7.UUID) (*Invoice, error)
    ProcessOverdueInvoices(ctx context.Context) (int, error) // Auto-mark overdue
}
```

### 1.5 HTTP Layer

**Handler**: `internal/contexts/billing/invoice/adapter/http/handler/invoice_handler.go`

**Endpoints** (14 endpoints):
- POST `/api/v1/billing/invoices` - Create invoice
- GET `/api/v1/billing/invoices/:id` - Get by ID
- GET `/api/v1/billing/invoices/no/:invoice_no` - Get by invoice number
- PUT `/api/v1/billing/invoices/:id` - Update (draft only)
- DELETE `/api/v1/billing/invoices/:id` - Delete (soft)
- POST `/api/v1/billing/invoices/:id/lines` - Add line item
- DELETE `/api/v1/billing/invoices/:id/lines/:line_id` - Remove line
- POST `/api/v1/billing/invoices/:id/send` - Mark as sent
- POST `/api/v1/billing/invoices/:id/pay` - Mark as paid
- POST `/api/v1/billing/invoices/:id/cancel` - Cancel
- POST `/api/v1/billing/invoices/:id/void` - Void
- GET `/api/v1/billing/invoices` - List all (paginated)
- GET `/api/v1/billing/invoices/customer/:customer_id` - List by customer
- GET `/api/v1/billing/invoices/overdue` - List overdue

**DTOs**: `internal/contexts/billing/invoice/adapter/http/handler/dto/invoice_dto.go`

### 1.6 Testing

**Unit Tests**:
- `entity_test.go` - Factory methods, business rules
- `usecase_test.go` - Business logic with mocks

**Smoke Tests**:
- `test/smoke/contexts/billing/invoice/handler_test.go` - HTTP handlers with mocks

**Integration Tests**:
- `test/integration/contexts/billing/invoice/repository_test.go` - Real DB

**Benchmark Tests**:
- `test/benchmark/contexts/billing/invoice/repository_bench_test.go` - Performance

**Target**: 40+ tests, 85%+ coverage

---

## Phase 2: Payment Aggregate (Days 4-6)

### 2.1 Domain Model

**Entity**: `internal/contexts/billing/payment/entity.go`

```go
type Payment struct {
    aggregate.BaseAggregate
    
    // Identity
    ID            uuidv7.UUID
    PaymentNo     string        // PAY-2026-000001
    
    // Relations
    CustomerID    uuidv7.UUID
    InvoiceID     *uuidv7.UUID  // Optional (can be standalone)
    
    // Payment Details
    Amount        Money
    Currency      string
    PaymentMethod PaymentMethod // credit_card, bank_transfer, paypal, stripe, cash
    
    // Card Info (optional)
    CardLast4     *string
    CardBrand     *string       // visa, mastercard, amex
    
    // External References
    TransactionID *string       // From payment gateway
    GatewayName   *string       // stripe, paypal, etc.
    
    // Status
    Status        PaymentStatus // pending, processing, completed, failed, refunded
    
    // Dates
    PaymentDate   time.Time
    ProcessedAt   *time.Time
    
    // Audit
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

type PaymentMethod string
type PaymentStatus string
```

**Business Rules**:
- Payment number auto-generated: `PAY-YYYY-NNNNNN`
- Completed payments are immutable
- Refunds create new payment with negative amount
- Failed payments can be retried

### 2.2 Database Schema

**Migration**: `migrations/billing/000002_payments.up.sql`

**Table**: `billing_payments` with indexes on customer_id, invoice_id, status

### 2.3 API Endpoints (12 endpoints)

- Create, Get, Update, Delete
- Process payment
- Refund payment
- List by customer, invoice, status

---

## Phase 3: Subscription Aggregate (Days 7-10)

### 3.1 Domain Model

**Entity**: `internal/contexts/billing/subscription/entity.go`

```go
type Subscription struct {
    aggregate.BaseAggregate
    
    // Identity
    ID             uuidv7.UUID
    SubscriptionNo string // SUB-2026-000001
    
    // Relations
    CustomerID     uuidv7.UUID
    PlanID         uuidv7.UUID // FK to billing_plans
    
    // Subscription Details
    Status         SubscriptionStatus // trial, active, paused, cancelled, expired
    BillingCycle   BillingCycle       // monthly, quarterly, yearly
    Amount         Money
    Currency       string
    
    // Dates
    StartDate      time.Time
    EndDate        *time.Time
    TrialEndDate   *time.Time
    NextBillingDate time.Time
    CancelledAt    *time.Time
    
    // Audit
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time
}

type SubscriptionStatus string
type BillingCycle string
```

**Business Rules**:
- One active subscription per customer per plan
- Trial converts to active after trial_end_date
- Cancelled subscriptions have grace period until end_date
- Auto-renew on next_billing_date

### 3.2 Database Schema

**Migration**: `migrations/billing/000003_subscriptions.up.sql`

**Tables**:
- `billing_plans` - Subscription plans (name, price, features)
- `billing_subscriptions` - Customer subscriptions

### 3.3 API Endpoints (14 endpoints)

- Create, Get, Update, Delete
- Activate, Pause, Resume, Cancel
- List by customer, status
- Process renewals (batch)

---

## Phase 4: Integration & Testing (Days 11-14)

### 4.1 Router Setup

**File**: `internal/contexts/billing/router.go`

```go
type Router struct {
    invoiceHandler      *invoiceHTTP.InvoiceHandler
    paymentHandler      *paymentHTTP.PaymentHandler
    subscriptionHandler *subscriptionHTTP.SubscriptionHandler
}

func NewRouter(db *sqlx.DB, eventBus bus.EventBus, jwtManager *jwt.Manager) *Router {
    // Initialize Invoice aggregate
    invoiceRepo := invoiceRepo.NewInvoiceRepository(db)
    invoiceUC := invoice.NewUseCase(invoiceRepo, eventBus)
    invoiceHandler := invoiceHTTP.NewInvoiceHandler(invoiceUC)
    
    // Initialize Payment aggregate
    // Initialize Subscription aggregate
    
    return &Router{
        invoiceHandler:      invoiceHandler,
        paymentHandler:      paymentHandler,
        subscriptionHandler: subscriptionHandler,
    }
}

func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
    billing := api.Group("/billing")
    billing.Use(jwt.AuthMiddleware(r.jwtManager)) // Protected routes
    {
        // Invoice routes
        invoices := billing.Group("/invoices")
        {
            invoices.POST("", r.invoiceHandler.Create)
            invoices.GET("/:id", r.invoiceHandler.GetByID)
            // ... 12 more endpoints
        }
        
        // Payment routes
        payments := billing.Group("/payments")
        {
            payments.POST("", r.paymentHandler.Create)
            // ... 10 more endpoints
        }
        
        // Subscription routes
        subscriptions := billing.Group("/subscriptions")
        {
            subscriptions.POST("", r.subscriptionHandler.Create)
            // ... 12 more endpoints
        }
    }
}
```

### 4.2 Event Bus Integration

**Published Events**:
- `billing.invoice.generated` - Invoice created
- `billing.invoice.sent` - Invoice sent to customer
- `billing.invoice.paid` - Invoice paid
- `billing.invoice.overdue` - Invoice overdue
- `billing.payment.received` - Payment completed
- `billing.payment.failed` - Payment failed
- `billing.subscription.created` - New subscription
- `billing.subscription.activated` - Trial → Active
- `billing.subscription.cancelled` - Subscription cancelled

**Event Handlers** (subscribe to):
- `order.confirmed` → Generate invoice
- `customer.created` → Setup billing profile

### 4.3 Main App Integration

**File**: `cmd/api/main.go`

```go
// Initialize Billing Context
billingRouter := billing.NewRouter(db, eventBus, jwtManager)
billingRouter.RegisterRoutes(api)
```

### 4.4 Comprehensive Testing

**Test Matrix**:
- Unit tests: 120+ (40 per aggregate)
- Smoke tests: 36+ (12 per aggregate)
- Integration tests: 30+ (10 per aggregate)
- Benchmark tests: 12+ (4 per aggregate)

**Total**: 200+ tests, 85%+ coverage

---

## Implementation Checklist

### Week 1: Invoice Aggregate
- [ ] Create domain model (`entity.go`)
- [ ] Create repository interface
- [ ] Create PostgreSQL repository with BaseRepository
- [ ] Create SQLite-compatible migration
- [ ] Create use case with business logic
- [ ] Create HTTP handler with 14 endpoints
- [ ] Create DTOs
- [ ] Write unit tests (40+)
- [ ] Write smoke tests (12+)
- [ ] Write integration tests (10+)
- [ ] Write benchmark tests (4+)

### Week 2: Payment Aggregate
- [ ] Domain model
- [ ] Repository (PostgreSQL + SQLite)
- [ ] Migration
- [ ] Use case
- [ ] Handler (12 endpoints)
- [ ] All tests (4 tiers)

### Week 3: Subscription Aggregate
- [ ] Domain model
- [ ] Repository
- [ ] Migration
- [ ] Use case
- [ ] Handler (14 endpoints)
- [ ] All tests

### Week 4: Integration & Polish
- [ ] Router setup
- [ ] Event bus integration
- [ ] Main app registration
- [ ] End-to-end tests
- [ ] Performance optimization
- [ ] Documentation
- [ ] API documentation (Swagger)

---

## Database Conventions

**Table Names**: `billing_{aggregate}` (lowercase, singular)
- `billing_invoices`
- `billing_invoice_lines`
- `billing_payments`
- `billing_subscriptions`
- `billing_plans`

**Column Names**: `snake_case`
- Timestamps: `created_at`, `updated_at`, `deleted_at`
- Foreign keys: `customer_id`, `invoice_id`, `order_id`
- Money: `{field}_amount` (BIGINT in cents), `currency` (VARCHAR(3))

**Indexes**: `idx_{table}_{column}`
- FKs: `idx_billing_invoices_customer`
- Status: `idx_billing_invoices_status`
- Dates: `idx_billing_invoices_due_date`

**Constraints**:
- Primary key: UUID
- Foreign keys: CASCADE on delete for child entities
- Check constraints: Status enums, positive amounts

---

## Patterns & Best Practices

### Multi-Database Support
- Use `?` placeholders in SQL (auto-converted)
- Store JSON as TEXT with `jsonstore.Field[T]`
- UUID generation in Go (`uuidv7.New()`)
- Timestamp management via `.Touch()` method

### Repository Pattern
- Embed `BaseRepository` for common operations
- Interface in aggregate package (`IRepository`)
- Implementation in `adapter/repository/postgres/`

### Use Case Pattern
- Interface `IUseCase` with business operations
- Implementation as private `useCase` struct
- Constructor: `NewUseCase(repo IRepository) IUseCase`

### Error Handling
- Define errors in `errors.go` per aggregate
- Standard naming: `Err{Aggregate}{Condition}`
- Wrap errors with context

### Testing Strategy
1. **Unit tests** - Entity methods, use case logic (mocks)
2. **Smoke tests** - HTTP handlers (mock use case)
3. **Integration tests** - Repository with real DB
4. **Benchmark tests** - Query performance

---

## Documentation

**Guides to Create**:
- `docs/concepts/billing-management.md` - Complete Billing Context guide
- Update `docs/INDEX.md` with Billing section
- Update `README.md` Bounded Contexts table
- API documentation with Swagger

---

## Success Metrics

**Code Quality**:
- 200+ tests (all 4 tiers)
- 85%+ test coverage
- 0 linting errors
- All patterns followed

**Performance**:
- Invoice creation: <100ms
- Payment processing: <200ms
- List queries: <50ms (20 items)

**Documentation**:
- Complete README per aggregate
- API documentation (Swagger)
- Migration guides
- Testing guides

---

## Next Steps

1. **Start with Invoice Aggregate** (highest priority)
2. Create entity.go with factory methods
3. Create PostgreSQL migration
4. Implement repository with BaseRepository
5. Implement use case with business logic
6. Create HTTP handler with 14 endpoints
7. Write all 4 tiers of tests

**Target**: Complete Invoice aggregate in 3 days

---

## Questions & Decisions

### Q1: Money Storage Format
**Decision**: Store as BIGINT in cents (like Order Management)
- Avoids floating-point precision issues
- Cross-database compatible
- Use Money value object for business logic

### Q2: Invoice Numbering
**Decision**: `INV-YYYY-NNNNNN` format (auto-increment per year)
- Similar to Order Management pattern
- Human-readable
- Sortable

### Q3: Payment Gateway Integration
**Decision**: Phase 2 (after basic functionality)
- Start with manual payment recording
- Add Stripe/PayPal integration later
- Focus on domain model first

### Q4: Subscription Billing Logic
**Decision**: Manual trigger initially, auto-renew later
- Add cron job for auto-renewals in Phase 3
- Focus on subscription lifecycle first

---

**Status**: Ready to start!   
**Next**: Create Invoice aggregate domain model

