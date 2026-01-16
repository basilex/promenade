# Order Management Context

**Domain:** Order processing, contracts, fulfillment  
**Ubiquitous Language:** Order, Contract, Fulfillment, OrderLine, Shipment  
**Status:**  Production (Order, OrderLine) | Contract, Fulfillment planned Q1-Q2 2026

---

## Overview

The **Order Management Context** handles the complete order lifecycle from creation to fulfillment. It coordinates between Customer Management, Billing, and Warehouse contexts.

### Responsibilities

 **What Order Management Context DOES:**

- Order creation and lifecycle management
- Contract generation and management
- Fulfillment tracking and shipment coordination
- Order status updates and notifications
- Order history and audit trail

 **What Order Management Context DOES NOT DO:**

- Payment processing → **Billing Context**
- Inventory management → **Warehouse Context**
- Customer relationships → **Customer Management Context**
- Pricing and subscriptions → **Billing Context**

---

## Domain Errors

This context follows the **Gold Standard** domain error pattern with 18 error constants across 2 aggregates.

### Order Errors

**File**: [`order/errors.go`](order/errors.go) - 10 error constants

**Repository Errors**:
- `ErrOrderNotFound` - Order not found by ID
- `ErrOrderUnauthorized` - Access denied

**Business Logic Errors**:
- `ErrOrderInvalidStatus` - Invalid status transition
- `ErrOrderAlreadyConfirmed` - Cannot modify confirmed order
- `ErrOrderAlreadyCancelled` - Cannot modify cancelled order
- `ErrOrderNoLineItems` - Cannot confirm empty order
- `ErrOrderLineNotFound` - Line item not found

**Technical Errors**:
- `ErrOrderCreateFailed` - Creation failed
- `ErrOrderUpdateFailed` - Update failed

### Contract Errors

**File**: [`contract/errors.go`](contract/errors.go) - 8 error constants

**Repository Errors**:
- `ErrContractNotFound` - Contract not found by ID
- `ErrContractUnauthorized` - Access denied

**Business Logic Errors**:
- `ErrContractInvalidStatus` - Invalid status transition
- `ErrContractExpired` - Contract expired
- `ErrContractAlreadySigned` - Cannot modify signed contract
- `ErrContractInvalidTerms` - Invalid contract terms

**Technical Errors**:
- `ErrContractCreateFailed` - Creation failed

### Usage Example

```go
// In UseCase Layer
order, err := uc.repo.GetOrder(ctx, orderID)
if err != nil {
    return nil, order.ErrOrderNotFound  // Domain constant
}

if err := order.Confirm(); err != nil {
    return nil, err  // ErrOrderAlreadyConfirmed (business rule)
}

// In Test Layer
err := usecase.ConfirmOrder(ctx, orderID)
assert.True(t, errors.Is(err, order.ErrOrderNoLineItems))  // Type-safe

// In Handler Layer
order, err := h.usecase.CreateOrder(ctx, req.CustomerID, req.Lines)
if errors.Is(err, order.ErrOrderNotFound) {
    response.NotFound(c, "Order not found")  // 404
    return
}
if errors.Is(err, order.ErrOrderAlreadyConfirmed) {
    response.BadRequest(c, "Order already confirmed")  // 400
    return
}
```

### HTTP Status Code Mapping

| Domain Error | HTTP Code | User Message |
|--------------|-----------|--------------|
| `ErrOrderNotFound` | 404 | "Order not found" |
| `ErrOrderAlreadyConfirmed` | 400 | "Order already confirmed" |
| `ErrOrderNoLineItems` | 400 | "Cannot confirm empty order" |
| `ErrContractExpired` | 400 | "Contract expired" |
| `ErrContractAlreadySigned` | 400 | "Cannot modify signed contract" |
| System errors (wrapped) | 500 | "Failed to {operation}" |

### Testing Patterns

Baseline budget and risk tiers follow [docs/guides/testing-patterns.md](../../docs/guides/testing-patterns.md): Unit 20–30, Smoke 6–9, Integration 6–10, with high-risk aggregates +30–50%.

**Entity Tests** (active development):
```go
// Phase 2 pattern
order := NewOrder(customerID, "USD")
err := order.Confirm()
assert.True(t, errors.Is(err, order.ErrOrderNoLineItems))  //  Type-safe
```

**Integration Tests** (6 tests):
```go
// Real database with domain errors
err := repo.Create(ctx, order)
if err != nil {
    assert.False(t, errors.Is(err, order.ErrOrderNotFound))  // Wrong error
}
```

**Smoke Tests** (10 tests):
```go
// HTTP validation
resp := smoke.MakeRequest(t, router, "POST", "/orders/"+invalidID+"/confirm", nil)
smoke.AssertErrorResponse(t, resp, 404, "ORDER_NOT_FOUND")
```

### Statistics

- **Total Domain Errors**: 18 across 2 aggregates (Order, Contract)
- **errors.go Files**: 2/2 (100% coverage)
- **Tests Using errors.Is()**: Entity tests in active development
- **Handler Discrimination**: All 14 Order endpoints + 12 Contract endpoints use errors.Is()
- **Fulfillment Saga**: 57 tests (100% passing) with separate error handling
- **Phase 2 Session**: 9 (order-mgmt completed)

**See**: [Domain Errors Guide](../../docs/guides/domain-errors.md) for comprehensive patterns and migration instructions.

---

## Aggregates

### 1. Order Aggregate

**Aggregate Root:** `Order`  
**Purpose:** Order lifecycle management

**Entity Structure:**

```go
type Order struct {
    aggregate.BaseAggregate

    // Identity
    ID         uuidv7.UUID
    OrderNumber string        // human-readable: ORD-2026-001
    CustomerID uuidv7.UUID    // references CustomerMgmt.Customer
    CompanyID  *uuidv7.UUID   // optional (B2B)

    // Order Details
    lines      []OrderLine    // private
    Total      valueobject.Money
    Currency   string

    // Status
    Status     OrderStatus    // pending, confirmed, processing, fulfilled, cancelled

    // Dates
    OrderDate     time.Time
    ConfirmedAt   *time.Time
    FulfilledAt   *time.Time
    CancelledAt   *time.Time

    // References
    ContractID    *uuidv7.UUID
    InvoiceID     *uuidv7.UUID // references Billing.Invoice

    // Lifecycle
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type OrderLine struct {
    ID         uuidv7.UUID
    ProductID  uuidv7.UUID    // references Warehouse.Product
    Quantity   int
    UnitPrice  valueobject.Money
    Total      valueobject.Money
}
```

**Business Rules:**

- Order must have at least one line item
- Total is computed from line items
- Cannot modify confirmed orders (must cancel and recreate)
- Cancelled orders cannot be reactivated
- Order status follows lifecycle: pending → confirmed → processing → fulfilled

**Methods:**

```go
func (o *Order) AddLine(productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) error
func (o *Order) RemoveLine(lineID uuidv7.UUID) error
func (o *Order) Confirm() error
func (o *Order) StartProcessing() error
func (o *Order) MarkFulfilled() error
func (o *Order) Cancel(reason string) error
func (o *Order) CalculateTotal() valueobject.Money
```

---

### 2. Contract Aggregate

**Aggregate Root:** `Contract`  
**Purpose:** Legal agreements and terms

**Entity Structure:**

```go
type Contract struct {
    aggregate.BaseAggregate

    // Identity
    ID            uuidv7.UUID
    ContractNumber string
    CustomerID    uuidv7.UUID
    CompanyID     *uuidv7.UUID
    DealID        *uuidv7.UUID // references CustomerMgmt.Deal

    // Contract Details
    Title         string
    Description   string
    Terms         string        // Legal terms and conditions
    Value         valueobject.Money

    // Dates
    StartDate     time.Time
    EndDate       time.Time
    SignedDate    *time.Time

    // Status
    Status        ContractStatus // draft, pending_signature, active, completed, terminated

    // Documents
    DocumentURL   string
    SignatureURL  *string

    // Lifecycle
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Business Rules:**

- Start date must be before end date
- Cannot modify active contracts (must create amendment)
- Signed contracts require signature URL
- Contract value must match related deal value

**Methods:**

```go
func (c *Contract) Sign(signatureURL string) error
func (c *Contract) Activate() error
func (c *Contract) Complete() error
func (c *Contract) Terminate(reason string) error
func (c *Contract) CreateAmendment() (*Contract, error)
```

---

### 3. Fulfillment Aggregate

**Aggregate Root:** `Fulfillment`  
**Purpose:** Order fulfillment tracking

**Entity Structure:**

```go
type Fulfillment struct {
    aggregate.BaseAggregate

    // Identity
    ID             uuidv7.UUID
    OrderID        uuidv7.UUID

    // Shipment
    shipments      []Shipment     // private

    // Status
    Status         FulfillmentStatus // pending, picking, packing, shipped, delivered

    // Dates
    ScheduledDate  time.Time
    ShippedDate    *time.Time
    DeliveredDate  *time.Time

    // Lifecycle
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type Shipment struct {
    ID             uuidv7.UUID
    TrackingNumber string
    Carrier        string
    Weight         float64
    ShippedDate    time.Time
    EstimatedDelivery time.Time
    items          []ShipmentItem
}
```

**Business Rules:**

- One fulfillment per order
- Cannot ship before inventory reserved
- Multiple shipments allowed (partial fulfillment)
- All items must be shipped before marking as delivered

**Methods:**

```go
func (f *Fulfillment) StartPicking() error
func (f *Fulfillment) StartPacking() error
func (f *Fulfillment) AddShipment(shipment Shipment) error
func (f *Fulfillment) MarkShipped() error
func (f *Fulfillment) MarkDelivered() error
```

---

## Saga Pattern: Create Order

```go
type CreateOrderSaga struct {
    customerMgmt CustomerMgmtService
    warehouse    WarehouseService
    billing      BillingService
    orderMgmt    OrderService
}

func (s *CreateOrderSaga) Execute(ctx context.Context, req CreateOrderRequest) error {
    saga := saga.NewBuilder("create-order").
        // Step 1: Validate customer
        Step("validate-customer",
            func(ctx context.Context) error {
                return s.customerMgmt.ValidateCustomer(ctx, req.CustomerID)
            },
            nil).

        // Step 2: Reserve inventory
        Step("reserve-inventory",
            func(ctx context.Context) error {
                return s.warehouse.ReserveInventory(ctx, req.Items)
            },
            func(ctx context.Context) error {
                return s.warehouse.ReleaseInventory(ctx, req.Items)
            }).

        // Step 3: Create invoice
        Step("create-invoice",
            func(ctx context.Context) error {
                return s.billing.CreateInvoice(ctx, req.CustomerID, req.Total)
            },
            func(ctx context.Context) error {
                return s.billing.CancelInvoice(ctx, invoiceID)
            }).

        // Step 4: Process payment
        Step("process-payment",
            func(ctx context.Context) error {
                return s.billing.ProcessPayment(ctx, invoiceID)
            },
            func(ctx context.Context) error {
                return s.billing.RefundPayment(ctx, paymentID)
            }).

        // Step 5: Create order
        Step("create-order",
            func(ctx context.Context) error {
                return s.orderMgmt.CreateOrder(ctx, req)
            },
            func(ctx context.Context) error {
                return s.orderMgmt.CancelOrder(ctx, orderID)
            }).

        Build()

    return saga.Execute(ctx)
}
```

---

## Domain Events

```go
// Order events
type OrderCreatedEvent struct {
    OrderID    uuidv7.UUID
    CustomerID uuidv7.UUID
    Total      valueobject.Money
    Timestamp  time.Time
}

type OrderConfirmedEvent struct {
    OrderID   uuidv7.UUID
    Timestamp time.Time
}

type OrderFulfilledEvent struct {
    OrderID   uuidv7.UUID
    Timestamp time.Time
}

// Contract events
type ContractSignedEvent struct {
    ContractID uuidv7.UUID
    CustomerID uuidv7.UUID
    Value      valueobject.Money
    Timestamp  time.Time
}

// Fulfillment events
type ShipmentCreatedEvent struct {
    FulfillmentID  uuidv7.UUID
    TrackingNumber string
    Timestamp      time.Time
}
```

---

## API Endpoints (Planned)

```
# Orders
GET    /api/v1/orders              # List orders
POST   /api/v1/orders              # Create order
GET    /api/v1/orders/:id          # Get order
PUT    /api/v1/orders/:id          # Update order
DELETE /api/v1/orders/:id          # Cancel order
POST   /api/v1/orders/:id/confirm  # Confirm order

# Contracts
GET    /api/v1/contracts           # List contracts
POST   /api/v1/contracts           # Create contract
GET    /api/v1/contracts/:id       # Get contract
POST   /api/v1/contracts/:id/sign  # Sign contract

# Fulfillment
GET    /api/v1/fulfillments        # List fulfillments
GET    /api/v1/fulfillments/:id    # Get fulfillment
POST   /api/v1/fulfillments/:id/ship  # Add shipment
```

---

**Status:**  Planned for Phase 3  
**Dependencies:** Customer Management, Billing, Warehouse contexts  
**Next:** Implement after Customer Management context
