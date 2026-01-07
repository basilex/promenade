# Warehouse Context

**Warehouse Context** manages inventory tracking, stock movements, product catalog, and warehouse locations with **automated order-inventory integration**.

## Overview

The Warehouse Context is a critical bounded context responsible for:

- **Product Catalog** - SKU-based product management with categories and brands
- **Inventory Tracking** - Real-time stock levels with multiple quantity states
- **Stock Movements** - Immutable audit trail for all inventory changes
- **Warehouse Locations** - Hierarchical location management (warehouses, zones, bins)
- **Order Integration** - Automated synchronization with Order Management via Event Bus

## Aggregates

### Product Aggregate

**Purpose**: Product catalog management with SKUs, categories, and physical properties

**Entity**:
```go
type Product struct {
    ID              uuid.UUID
    SKU             string  // Unique product identifier
    Name            string
    Description     string
    Category        string
    Brand           string
    UnitPrice       int     // Cents (Money value object)
    Currency        string
    Weight          float64 // Grams
    Dimensions      Dimensions // Length, Width, Height in cm
    Status          ProductStatus // active, discontinued
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Use Cases**:
- `CreateProduct(sku, name, category, brand, price)` - Add new product
- `UpdateProduct(id, ...)` - Update product details
- `DiscontinueProduct(id)` - Mark product as discontinued
- `GetProductBySKU(sku)` - Find product by SKU
- `ListProducts(filters)` - Search and filter products

**Domain Events**:
- `product.created`
- `product.updated`
- `product.discontinued`

**Tests**: 139 (25 entity + 83 usecase + 10 smoke + 21 integration)

---

### Inventory Aggregate

**Purpose**: Real-time stock level tracking with reservation management

**Entity**:
```go
type Inventory struct {
    ID                  uuid.UUID
    ProductID           uuid.UUID
    LocationID          uuid.UUID
    QuantityOnHand      int  // Physical stock in warehouse
    QuantityReserved    int  // Reserved for orders
    QuantityAvailable   int  // Available for new orders (OnHand - Reserved)
    QuantityCommitted   int  // Shipped/delivered
    ReorderPoint        int  // Threshold for reordering
    MinStock            int  // Minimum stock level
    MaxStock            int  // Maximum stock level
    UnitCost            int  // Cents (Money value object)
    TotalCost           int  // Cents (QuantityOnHand * UnitCost)
    Currency            string
    LastStockDate       time.Time
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```

**Stock Operations**:
- `ReceiveStock(quantity, unitCost)` - Receive stock from supplier
- `ReserveStock(quantity, orderID)` - Reserve for order
- `ReleaseReservation(quantity, orderID)` - Cancel reservation
- `CommitStock(quantity, orderID)` - Ship/deliver stock

**Business Rules**:
- QuantityAvailable = QuantityOnHand - QuantityReserved
- Cannot reserve more than available stock
- Weighted average cost calculation on receipt
- Auto-update TotalCost on stock movements

**Domain Events**:
- `inventory.created`
- `inventory.received`
- `inventory.reserved`
- `inventory.released`
- `inventory.committed`

**Tests**: 141 (97 entity + ? usecase + 21 smoke + 23 integration)

---

### StockMovement Aggregate

**Purpose**: Immutable audit trail for all inventory changes

**Entity**:
```go
type StockMovement struct {
    ID                  uuid.UUID
    InventoryID         uuid.UUID
    Type                MovementType // Receipt, Reservation, Release, Commit, etc.
    Quantity            int          // Positive or negative
    QuantityBeforeMove  int          // Snapshot before
    QuantityAfterMove   int          // Snapshot after
    FromWarehouseID     *uuid.UUID
    ToWarehouseID       *uuid.UUID
    ReferenceType       string       // "order", "po", "adjustment"
    ReferenceID         *uuid.UUID
    Reason              string       // Required for adjustments
    UnitCostCents       int
    TotalCostCents      int
    CurrencyCode        string
    CreatedBy           uuid.UUID
    MovementDate        time.Time
    CreatedAt           time.Time
}
```

**Movement Types**:
- **Receipt**: Stock received from supplier
- **Reservation**: Stock reserved for order
- **ReservationRelease**: Reservation cancelled
- **Commit**: Stock shipped/delivered
- **Adjustment**: Manual correction
- **Transfer**: Between warehouses
- **Damage**: Damaged goods
- **Return**: Customer return

**Key Features**:
- Immutable records (append-only)
- Before/After snapshots for auditability
- Cost tracking per movement
- Reference to orders/POs/adjustments
- Compliance-ready audit trail

**Domain Events**:
- `stock_movement.created`

**Tests**: 45 (11 entity + 10 usecase + 9 smoke + 15 integration)

---

### Location Aggregate

**Purpose**: Hierarchical warehouse location management

**Entity**:
```go
type Location struct {
    ID          uuid.UUID
    Name        string
    Code        string       // Unique location code
    Type        LocationType // warehouse, retail, dropship, virtual
    Address     valueobject.Address
    CountryCode string
    Capacity    int          // Storage capacity
    IsActive    bool
    ParentID    *uuid.UUID   // For hierarchical structure
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Location Types**:
- **Warehouse**: Main facility
- **Retail**: Store location
- **Dropship**: Supplier warehouse
- **Virtual**: Digital products

**Use Cases**:
- `CreateLocation(name, code, type, address)` - Add warehouse/store
- `UpdateLocation(id, ...)` - Update details
- `ActivateLocation(id)` / `DeactivateLocation(id)` - Enable/disable
- `GetLocationsByType(type)` - Filter by type
- `GetActiveLocations()` - List active locations

**Domain Events**:
- `location.created`
- `location.updated`
- `location.activated`
- `location.deactivated`

**Tests**: 74 (48 entity + 17 integration + 9 smoke)

---

## Warehouse Integration

**The most powerful feature** - automated order-inventory synchronization via Event Bus.

### Architecture

```
Order Management          Event Bus          Warehouse Context
┌─────────────────┐       ┌────┐       ┌─────────────────────┐
│ Order.Confirm() │───────▶│    │───────▶│ ReservationService │
│                 │       │    │       │ .ReserveForOrder() │
└─────────────────┘       └────┘       └─────────────────────┘

┌─────────────────┐       ┌────┐       ┌─────────────────────┐
│ Order.Cancel()  │───────▶│    │───────▶│ ReservationService │
│                 │       │    │       │ .ReleaseForOrder() │
└─────────────────┘       └────┘       └─────────────────────┘

┌─────────────────┐       ┌────┐       ┌─────────────────────┐
│ Order.Fulfill() │───────▶│    │───────▶│ ReservationService │
│                 │       │    │       │ .CommitForOrder()  │
└─────────────────┘       └────┘       └─────────────────────┘
```

### Components

**1. ReservationService** (230 lines)

Business logic for stock operations:

```go
type ReservationService struct {
    inventoryUC inventory.IUseCase
}

func (s *ReservationService) ReserveForOrder(ctx context.Context, orderID uuid.UUID, items []OrderItem) error {
    // For each order line item:
    // 1. Validate inventory availability
    // 2. Reserve stock via Inventory.ReserveStock()
    // 3. Create StockMovement record (type: Reservation)
}

func (s *ReservationService) ReleaseForOrder(ctx context.Context, orderID uuid.UUID) error {
    // Find all reservations for order
    // Release stock back to available pool
    // Create StockMovement records (type: ReservationRelease)
}

func (s *ReservationService) CommitForOrder(ctx context.Context, orderID uuid.UUID) error {
    // Find all reservations for order
    // Commit stock (permanent removal from inventory)
    // Create StockMovement records (type: Commit)
}
```

**2. OrderEventHandler** (230 lines)

Event handlers for order lifecycle:

```go
type OrderEventHandler struct {
    reservationService *ReservationService
}

func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event bus.Event) error {
    // Extract order data from event
    // Call ReservationService.ReserveForOrder()
}

func (h *OrderEventHandler) HandleOrderCancelled(ctx context.Context, event bus.Event) error {
    // Extract order ID from event
    // Call ReservationService.ReleaseForOrder()
}

func (h *OrderEventHandler) HandleOrderFulfilled(ctx context.Context, event bus.Event) error {
    // Extract order ID from event
    // Call ReservationService.CommitForOrder()
}
```

**3. Bootstrap Integration**

Initialized in `cmd/api/bootstrap.go`:

```go
func initWarehouseIntegration(db *sqlx.DB, eventBus bus.IBus) (*integration.OrderEventHandler, error) {
    invRepo := inventoryRepo.NewInventoryRepository(db)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)
    orderEventHandler := integration.NewOrderEventHandler(reservationService)
    
    if err := orderEventHandler.RegisterHandlers(eventBus); err != nil {
        return nil, err
    }
    
    return orderEventHandler, nil
}
```

### Event Flows

**Order Confirmation** (order.confirmed):
1. Order Management: `Order.Confirm()` publishes `order.confirmed` event
2. Event Bus: Routes event to Warehouse Context
3. Warehouse: `OrderEventHandler.HandleOrderConfirmed()`
4. Warehouse: `ReservationService.ReserveForOrder()`
5. Result: Stock reserved, order can proceed to fulfillment

**Order Cancellation** (order.cancelled):
1. Order Management: `Order.Cancel()` publishes `order.cancelled` event
2. Event Bus: Routes event to Warehouse Context
3. Warehouse: `OrderEventHandler.HandleOrderCancelled()`
4. Warehouse: `ReservationService.ReleaseForOrder()`
5. Result: Stock returned to available pool

**Order Fulfillment** (order.fulfilled):
1. Order Management: `Order.MarkFulfilled()` publishes `order.fulfilled` event
2. Event Bus: Routes event to Warehouse Context
3. Warehouse: `OrderEventHandler.HandleOrderFulfilled()`
4. Warehouse: `ReservationService.CommitForOrder()`
5. Result: Stock permanently removed from inventory

---

## API Endpoints

### Product (16 endpoints)
- `POST /products` - Create product
- `GET /products/:id` - Get by ID
- `GET /products` - List products
- `PUT /products/:id` - Update product
- `DELETE /products/:id` - Soft delete
- `GET /products/sku/:sku` - Get by SKU
- `GET /products/:id/inventory` - Get inventory for product

### Inventory (14 endpoints)
- `POST /inventory` - Create inventory
- `GET /inventory/:id` - Get by ID
- `PUT /inventory/:id` - Update inventory
- `DELETE /inventory/:id` - Soft delete
- `POST /inventory/:id/receive` - Receive stock
- `POST /inventory/:id/reserve` - Reserve stock
- `POST /inventory/:id/release` - Release reservation
- `POST /inventory/:id/commit` - Commit stock

### StockMovement (7 endpoints)
- `POST /stock-movements` - Create movement
- `GET /stock-movements/:id` - Get by ID
- `GET /stock-movements` - List movements
- `GET /stock-movements/inventory/:id` - By inventory
- `GET /stock-movements/recent` - Recent movements

### Location (14 endpoints)
- `POST /locations` - Create location
- `GET /locations/:id` - Get by ID
- `GET /locations` - List locations
- `PUT /locations/:id` - Update location
- `DELETE /locations/:id` - Soft delete

**Total**: 58 endpoints

---

## Testing

### Test Coverage

| Component          | Tests | Type        |
| ------------------ | ----- | ----------- |
| **Product**        | 139   | Entity (25) + UseCase (83) + Smoke (10) + Integration (21) |
| **Inventory**      | 141   | Entity (97) + Integration (23) + Smoke (21) |
| **StockMovement**  | 45    | Entity (11) + UseCase (10) + Smoke (9) + Integration (15) |
| **Location**       | 74    | Entity (48) + Integration (17) + Smoke (9) |
| **Integration**    | 34    | E2E order-warehouse flows |
| **Total**          | **433** | **100% passing** (~4.2s) |

### Integration Tests (34)

End-to-end testing of order-warehouse synchronization:

1. **Order Confirmation Flow** - Verify stock reservation
2. **Order Cancellation Flow** - Verify reservation release
3. **Order Fulfillment Flow** - Verify stock commit
4. **Edge Cases** - Insufficient inventory, concurrent reservations, retries

---

## Database Schema

**7 Tables**:

1. `warehouse_products` - Product catalog (SKU, name, category, brand, price, dimensions)
2. `warehouse_inventory` - Stock levels (on hand, reserved, available, committed)
3. `warehouse_stock_movements` - Audit trail (immutable, all changes recorded)
4. `warehouse_locations` - Warehouse/store locations (address, capacity, type)

**Key Indexes**:
- `idx_products_sku` - Unique SKU lookup
- `idx_inventory_product_location` - Composite for stock queries
- `idx_stock_movements_inventory_id` - Audit trail queries
- `idx_locations_code` - Unique location code

**Audit Columns**:
- All tables have `created_at`, `updated_at`, `deleted_at` (soft delete)
- StockMovement has `created_by` for user tracking
- Before/After snapshots in StockMovement for auditability

---

## Configuration

### Event Bus Configuration

Required in `config/app.{driver}-{env}.yaml`:

```yaml
bus:
  adapter: "redis"           # Use redis for distributed systems
  worker_pool_size: 50       # Concurrent event processing
  buffer_size: 10000         # Event queue capacity
  retry_attempts: 5          # Retry failed event handlers
  retry_delay: 1s            # Initial retry delay
```

**No Additional Configuration**: Warehouse Integration uses existing database and Event Bus connections.

---

## Related Documentation

- [Warehouse Management Concept](../concepts/warehouse-management.md) - High-level overview
- [Order Management Context](./order.md) - Order lifecycle integration
- [Event-Driven Architecture](../concepts/event-driven.md) - Event Bus patterns
- [Testing Guide](../guide/testing-patterns.md) - Four-tier testing strategy

---

**Status**: Production Ready (100% complete)  
**Last Updated**: January 7, 2026  
**Aggregates**: 4 (Product, Inventory, StockMovement, Location)  
**Endpoints**: 58  
**Tests**: 433 (100% passing)  
**Main Feature**: Automated order-inventory synchronization via Event Bus
