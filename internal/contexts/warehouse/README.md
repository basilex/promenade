# Warehouse Context

**Domain:** Inventory management, products, stock tracking  
**Ubiquitous Language:** Product, Inventory, Stock, Location, Movement, Reservation  
**Status:**  In Progress - 45% Complete (Phase 2 Q1 2026)

**Latest Update:** January 6, 2026  
**Completed:** Inventory + StockMovement Aggregates ✅  
**In Progress:** Product & Location Aggregates (Task 2.3)

---

## Overview

The **Warehouse Context** manages physical goods, inventory levels, and stock movements. It's essential for e-commerce and order fulfillment workflows.

### Implementation Status

**Task 2.1: Inventory Aggregate** - ✅ **COMPLETE**
- ✅ Entity (464 lines, 11 business methods)
- ✅ Repository (595 lines, 13 methods)
- ✅ UseCase (267 lines, 11 methods)
- ✅ HTTP Handlers (14 endpoints)
- ✅ Integration Tests (23 tests)
- ✅ Unit Tests (97 tests)
- ✅ Smoke Tests (21 tests)
- ✅ Router & Server Integration
- **Total:** 141 tests, 100% passing

**Task 2.2: StockMovement Aggregate** - ✅ **COMPLETE**
- ✅ Entity (11 entity tests)
- ✅ Repository (550 lines, 11 methods with PostgreSQL placeholders)
- ✅ UseCase (10 usecase tests)
- ✅ Integration Tests (15 tests: 7 repository + 8 usecase)
- ✅ Smoke Tests (9 tests)
- ✅ FK Constraint Fixes (createTestInventory helpers)
- ✅ SQL Syntax Fixes (? → $N placeholders)
- **Total:** 45 tests, 100% passing

**Next Tasks:**
- Task 2.3: Product & Location Aggregates
- Task 2.4: Order Management Integration (stock reservation on order creation)
- Task 2.5: Low Stock Alerts System
- Task 2.6: HTTP API Completion (22+ endpoints)

### Responsibilities

 **What Warehouse Context DOES:**

- Product catalog management
- Inventory tracking (stock levels, reservations)
- Warehouse location management
- Stock movements (in, out, transfers)
- Low stock alerts and reordering
- Serial number / lot tracking

 **What Warehouse Context DOES NOT DO:**

- Pricing → **Billing Context**
- Order processing → **Order Management Context**
- Shipping logistics → **Order Management Context**
- Product sales/marketing → Outside bounded contexts

---

## Aggregates

### 1. Product Aggregate

**Aggregate Root:** `Product`  
**Purpose:** Product catalog and specifications

**Entity Structure:**

```go
type Product struct {
    aggregate.BaseAggregate

    // Identity
    ID          uuidv7.UUID
    SKU         string         // Stock Keeping Unit (unique)
    Name        string
    Description string

    // Classification
    Category    string
    Brand       string
    Tags        []string

    // Physical Properties
    Weight      float64        // kg
    Dimensions  Dimensions     // length, width, height in cm

    // Inventory Settings
    TrackInventory bool         // if false, unlimited stock
    AllowBackorder bool
    ReorderPoint   int          // low stock threshold
    ReorderQuantity int

    // Serial Tracking
    TrackSerialNumbers bool
    TrackLotNumbers    bool

    // Status
    Status      ProductStatus  // active, discontinued, out_of_stock

    // Lifecycle
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Dimensions struct {
    Length float64
    Width  float64
    Height float64
}
```

**Business Rules:**

- SKU must be unique
- Weight and dimensions required for physical products
- Cannot discontinue product with pending orders
- Reorder point must be less than reorder quantity

**Methods:**

```go
func (p *Product) Activate() error
func (p *Product) Discontinue() error
func (p *Product) UpdateInventorySettings(trackInventory, allowBackorder bool) error
func (p *Product) SetReorderPoint(point, quantity int) error
```

---

### 2. Inventory Aggregate ✅ **PRODUCTION READY**

**Aggregate Root:** `Inventory`  
**Purpose:** Stock level tracking and reservations  
**Status:** Fully implemented with 141 tests passing

**API Endpoints:** 14 endpoints at `/api/v1/warehouse/inventory`

**CRUD Operations:**
- `POST /` - Create inventory item
- `GET /:id` - Get by ID
- `PUT /:id` - Update inventory item
- `DELETE /:id` - Soft delete
- `GET /` - List inventory (paginated)

**Query Operations:**
- `GET /sku/:sku` - Get by SKU
- `GET /product/:product_id` - Get by product ID
- `GET /warehouse/:warehouse_id` - Get by warehouse
- `GET /location/:warehouse_id/:location_code` - Get by location
- `GET /low-stock` - Get low stock items

**Stock Operations:**
- `POST /:id/receive` - Receive stock
- `POST /:id/reserve` - Reserve stock for order (Order Management integration)
- `POST /:id/release` - Release reservation (Saga compensation)
- `POST /:id/commit` - Commit stock

**Entity Structure:**

```go
type Inventory struct {
    aggregate.BaseAggregate

    // Identity
    ID           uuidv7.UUID
    ProductID    uuidv7.UUID
    ProductSKU   string
    ProductName  string
    WarehouseID  uuidv7.UUID
    LocationCode string

    // Stock Levels
    QuantityOnHand       int  // Physical stock
    QuantityReserved     int  // Reserved for orders
    QuantityAvailable    int  // OnHand - Reserved
    QuantityCommitted    int  // Committed to fulfilled orders
    QuantityDamaged      int  // Damaged/unusable stock

    // Cost Tracking
    UnitCost        float64      // Weighted average cost
    TotalCost       float64      // Total inventory value

    // Reorder Management
    ReorderPoint    int          // Low stock threshold
    ReorderQuantity int          // Reorder amount
    MinStock        int          // Minimum stock level
    MaxStock        int          // Maximum stock level

    // Status
    Status          InventoryStatus  // active, inactive, discontinued
    IsActive        bool
    LastStockDate   *time.Time       // Last stock movement

    // Metadata
    Notes           string
}
```

**Business Rules:**

- QuantityAvailable = QuantityOnHand - QuantityReserved
- Cannot reserve more than available stock
- Physical stock cannot be negative (except damaged adjustments)
- Weighted average cost calculation on stock receipts
- Optimistic locking via version field
- Soft delete support (deleted_at)

**Business Methods (11 total):**

```go
func (i *Inventory) ReceiveStock(quantity int, unitCost float64) error
func (i *Inventory) ReserveStock(quantity int) error
func (i *Inventory) ReleaseReservation(quantity int) error
func (i *Inventory) CommitReservation(quantity int) error
func (i *Inventory) AdjustStock(adjustment int, reason string) error
func (i *Inventory) SetLocation(warehouseID uuidv7.UUID, locationCode string) error
func (i *Inventory) SetReorderPoint(point, quantity int) error
func (i *Inventory) MarkAsDamaged(quantity int) error
func (i *Inventory) Activate() error
func (i *Inventory) Deactivate() error
func (i *Inventory) IsLowStock() bool
func (i *Inventory) GetStockValue() float64
func (i *Inventory) Validate() error
```

---

### 3. StockMovement Aggregate ✅ **PRODUCTION READY**

**Aggregate Root:** `StockMovement`  
**Purpose:** Audit trail for all stock changes (immutable append-only log)  
**Status:** Fully implemented with 45 tests passing

**Entity Structure:**

```go
type StockMovement struct {
    aggregate.BaseAggregate

    // Identity
    ID                 uuidv7.UUID
    InventoryID        uuidv7.UUID
    Type               MovementType
    Quantity           int  // Positive or negative

    // Snapshot (audit trail)
    QuantityBeforeMove int
    QuantityAfterMove  int

    // Location Tracking (for transfers)
    FromWarehouseID    *uuidv7.UUID
    FromLocationCode   *string
    ToWarehouseID      *uuidv7.UUID
    ToLocationCode     *string

    // Reference Tracking
    ReferenceType      string       // "order", "po", "adjustment"
    ReferenceID        *uuidv7.UUID // Order ID, PO ID, etc.

    // Cost Tracking
    UnitCostCents      int64
    TotalCostCents     int64
    CurrencyCode       string

    // Metadata
    Reason             string       // Required for adjustments
    Notes              string
    CreatedBy          uuidv7.UUID
    MovementDate       time.Time
}

type MovementType string

const (
    MovementTypeReceipt            MovementType = "receipt"             // Incoming stock
    MovementTypeReservation        MovementType = "reservation"         // Reserved for order
    MovementTypeReservationRelease MovementType = "reservation_release" // Cancelled reservation
    MovementTypeCommit             MovementType = "commit"              // Committed to fulfilled order
    MovementTypeAdjustment         MovementType = "adjustment"          // Manual correction
    MovementTypeTransfer           MovementType = "transfer"            // Location transfer
    MovementTypeDamage             MovementType = "damage"              // Damaged goods
    MovementTypeReturn             MovementType = "return"              // Customer return
)
```

**Business Rules:**

- Movements are **immutable** (append-only, no updates/deletes)
- Adjustments require reason
- Transfers require both from and to warehouses
- Quantity cannot be zero
- Cost tracking optional (can be zero)
- Snapshot captures state before and after movement

**Business Methods:**

```go
func NewStockMovement(inventoryID uuidv7.UUID, movType MovementType, quantity int, quantityBefore int, createdBy uuidv7.UUID) (*StockMovement, error)
func (sm *StockMovement) SetReference(refType string, refID uuidv7.UUID) error
func (sm *StockMovement) SetCost(unitCostCents, totalCostCents int64, currencyCode string) error
func (sm *StockMovement) SetReason(reason string) error
func (sm *StockMovement) SetNotes(notes string) error
func (sm *StockMovement) SetWarehouseLocation(fromWarehouseID, toWarehouseID *uuidv7.UUID, fromLocation, toLocation *string) error
func (sm *StockMovement) GetImpact() int  // Returns +/- for stock calculations
func (sm *StockMovement) IsAdjustment() bool
func (sm *StockMovement) Validate() error
```

**Repository Methods (11 total):**

```go
Create(ctx, movement) error
GetByID(ctx, id) (*StockMovement, error)
GetByInventoryID(ctx, inventoryID, page, pageSize) ([]*StockMovement, int, error)
GetByType(ctx, movementType, startDate, endDate, page, pageSize) ([]*StockMovement, int, error)
GetByReference(ctx, refType, refID) ([]*StockMovement, error)
GetByDateRange(ctx, startDate, endDate, page, pageSize) ([]*StockMovement, int, error)
GetSummaryByInventory(ctx, inventoryID, startDate, endDate) (totalIn, totalOut int, err error)
GetRecentMovements(ctx, limit) ([]*StockMovement, error)
CountByType(ctx, startDate, endDate) (map[MovementType]int, error)
```

**UseCase Methods:**

```go
RecordMovement(ctx, inventoryID, movType, quantity, quantityBefore, createdBy) (*StockMovement, error)
RecordReceipt(ctx, inventoryID, quantity, unitCost, createdBy) (*StockMovement, error)
RecordReservation(ctx, inventoryID, quantity, orderID, createdBy) (*StockMovement, error)
RecordCommit(ctx, inventoryID, quantity, orderID, createdBy) (*StockMovement, error)
RecordAdjustment(ctx, inventoryID, adjustment, reason, createdBy) (*StockMovement, error)
RecordTransfer(ctx, inventoryID, quantity, fromWarehouse, toWarehouse, createdBy) (*StockMovement, error)
GetMovementsByInventory(ctx, inventoryID, page, pageSize) ([]*StockMovement, int, error)
GetMovementsByType(ctx, movType, startDate, endDate, page, pageSize) ([]*StockMovement, int, error)
GetInventorySummary(ctx, inventoryID, startDate, endDate) (totalIn, totalOut int, err error)
```

**Test Coverage:**
- Entity Tests: 11 (NewStockMovement, Validate, Set*, Is*)
- UseCase Tests: 10 (all business methods)
- Smoke Tests: 9 (HTTP handler validation)
- Integration Tests: 15 (7 repository + 8 usecase with real DB)
- **Total: 45 tests, 100% passing**

**Key Implementation Details:**

1. **PostgreSQL Placeholders**: All SQL queries use `$1`, `$2`, `$3` (not generic `?`)
2. **FK Constraints**: Integration tests create inventory first via helpers
3. **Date Range Queries**: Tests use real date ranges (not `time.Time{}`)
4. **Immutable Audit Trail**: No update/delete methods, only Create and Get
5. **Cost Tracking**: Stored in cents (int64) to avoid floating-point issues

---

### 4. Location Aggregate

**Aggregate Root:** `Location`  
**Purpose:** Warehouse locations and capacity

**Entity Structure:**

```go
type Location struct {
    aggregate.BaseAggregate

    // Identity
    ID       uuidv7.UUID
    Code     string         // e.g., "WH1-A-01-05" (warehouse-aisle-rack-shelf)
    Name     string
    Type     LocationType   // warehouse, store, dropship

    // Hierarchy
    ParentID *uuidv7.UUID   // for nested locations

    // Address
    Address  valueobject.Address

    // Capacity
    MaxVolume  float64      // cubic meters
    MaxWeight  float64      // kg

    // Status
    Status     LocationStatus // active, inactive, full

    // Lifecycle
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**Business Rules:**

- Location codes must be unique
- Cannot delete location with inventory
- Parent location must exist
- Capacity limits enforced when adding inventory

**Methods:**

```go
func (l *Location) Activate() error
func (l *Location) Deactivate() error
func (l *Location) CanAccommodate(volume, weight float64) bool
```

---

## Domain Events

```go
// Product events
type ProductCreatedEvent struct {
    ProductID uuidv7.UUID
    SKU       string
    Name      string
    Timestamp time.Time
}

type ProductDiscontinuedEvent struct {
    ProductID uuidv7.UUID
    Timestamp time.Time
}

// Inventory events
type StockReservedEvent struct {
    InventoryID   uuidv7.UUID
    ProductID     uuidv7.UUID
    OrderID       uuidv7.UUID
    Quantity      int
    Timestamp     time.Time
}

type StockCommittedEvent struct {
    InventoryID   uuidv7.UUID
    ProductID     uuidv7.UUID
    OrderID       uuidv7.UUID
    Quantity      int
    Timestamp     time.Time
}

type LowStockAlertEvent struct {
    ProductID     uuidv7.UUID
    LocationID    uuidv7.UUID
    Available     int
    ReorderPoint  int
    Timestamp     time.Time
}

// Movement events
type StockMovementEvent struct {
    MovementID    uuidv7.UUID
    ProductID     uuidv7.UUID
    Type          string
    Quantity      int
    FromLocation  *uuidv7.UUID
    ToLocation    *uuidv7.UUID
    Timestamp     time.Time
}
```

---

## Communication with Other Contexts

### With Order Management

```go
// OrderMgmt publishes order created
eventBus.Publish("order.created", OrderCreatedEvent{...})

// Warehouse subscribes and reserves inventory
eventBus.Subscribe("order.created", func(event OrderCreatedEvent) {
    for _, item := range event.Items {
        warehouseService.ReserveInventory(item.ProductID, item.Quantity)
    }
})

// OrderMgmt publishes order cancelled
eventBus.Publish("order.cancelled", OrderCancelledEvent{...})

// Warehouse subscribes and releases reservations
eventBus.Subscribe("order.cancelled", func(event OrderCancelledEvent) {
    warehouseService.ReleaseReservation(event.OrderID)
})
```

### With Billing Context

```go
// Warehouse publishes low stock alert
eventBus.Publish("inventory.low_stock", LowStockAlertEvent{...})

// Billing could subscribe to update pricing (scarcity pricing)
eventBus.Subscribe("inventory.low_stock", func(event LowStockAlertEvent) {
    // Optional: implement dynamic pricing
})
```

---

## Database Schema (Planned)

```sql
-- Products
CREATE TABLE warehouse_products (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    brand VARCHAR(100),
    tags TEXT[],
    weight DECIMAL(10,2),
    dimensions JSONB,
    track_inventory BOOLEAN NOT NULL DEFAULT true,
    allow_backorder BOOLEAN NOT NULL DEFAULT false,
    reorder_point INT,
    reorder_quantity INT,
    track_serial_numbers BOOLEAN NOT NULL DEFAULT false,
    track_lot_numbers BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Inventory
CREATE TABLE warehouse_inventory (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    product_id UUID NOT NULL REFERENCES warehouse_products(id),
    location_id UUID NOT NULL REFERENCES warehouse_locations(id),
    available INT NOT NULL DEFAULT 0,
    reserved INT NOT NULL DEFAULT 0,
    physical INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, location_id)
);

-- Reservations
CREATE TABLE warehouse_reservations (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    inventory_id UUID NOT NULL REFERENCES warehouse_inventory(id),
    order_id UUID NOT NULL,
    quantity INT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Movements
CREATE TABLE warehouse_movements (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    inventory_id UUID NOT NULL REFERENCES warehouse_inventory(id),
    type VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    from_location UUID REFERENCES warehouse_locations(id),
    to_location UUID REFERENCES warehouse_locations(id),
    reference VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Locations
CREATE TABLE warehouse_locations (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    parent_id UUID REFERENCES warehouse_locations(id),
    address JSONB,
    max_volume DECIMAL(10,2),
    max_weight DECIMAL(10,2),
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

---

## API Endpoints (Planned)

```
# Products
GET    /api/v1/warehouse/products
POST   /api/v1/warehouse/products
GET    /api/v1/warehouse/products/:id
PUT    /api/v1/warehouse/products/:id
DELETE /api/v1/warehouse/products/:id

# Inventory
GET    /api/v1/warehouse/inventory
GET    /api/v1/warehouse/inventory/:id
POST   /api/v1/warehouse/inventory/:id/adjust
POST   /api/v1/warehouse/inventory/:id/reserve
POST   /api/v1/warehouse/inventory/:id/transfer

# Locations
GET    /api/v1/warehouse/locations
POST   /api/v1/warehouse/locations
GET    /api/v1/warehouse/locations/:id
PUT    /api/v1/warehouse/locations/:id
DELETE /api/v1/warehouse/locations/:id

# Movements
GET    /api/v1/warehouse/movements
GET    /api/v1/warehouse/movements/:id
```

---

## Implementation Plan

### Phase 4 (4 weeks after Phase 3)

**Week 1: Product Aggregate**

- Product entity (50 tests)
- Product repository
- Product use cases
- Product handlers

**Week 2: Inventory Aggregate**

- Inventory entity (60 tests)
- Reservation logic
- Movement tracking
- Low stock alerts

**Week 3: Location Aggregate**

- Location entity (30 tests)
- Location hierarchy
- Capacity management

**Week 4: Integration**

- Event-driven integration with Order Management
- Saga pattern for stock reservations
- Integration tests (40 tests)

**Total:** 180 tests for Warehouse context

---

**Status:**  Planned for Phase 4  
**Dependencies:** Order Management context  
**Next:** Implement after Order Management and Billing contexts
