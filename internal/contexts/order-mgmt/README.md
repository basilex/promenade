# Order Management Context

**Domain:** Order processing, contracts, fulfillment  
**Ubiquitous Language:** Order, Contract, Fulfillment, OrderLine, Shipment  
**Status:** 📋 Planned (Phase 3 implementation)

---

## Overview

The **Order Management Context** handles the complete order lifecycle from creation to fulfillment. It coordinates between Customer Management, Billing, and Warehouse contexts.

### Responsibilities

✅ **What Order Management Context DOES:**

- Order creation and lifecycle management
- Contract generation and management
- Fulfillment tracking and shipment coordination
- Order status updates and notifications
- Order history and audit trail

❌ **What Order Management Context DOES NOT DO:**

- Payment processing → **Billing Context**
- Inventory management → **Warehouse Context**
- Customer relationships → **Customer Management Context**
- Pricing and subscriptions → **Billing Context**

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

**Status:** 📋 Planned for Phase 3  
**Dependencies:** Customer Management, Billing, Warehouse contexts  
**Next:** Implement after Customer Management context
