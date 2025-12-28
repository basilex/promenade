# Warehouse Context

**Domain:** Inventory management, products, stock tracking  
**Ubiquitous Language:** Product, Inventory, Stock, Location, Movement, Reservation  
**Status:**  Planned (Phase 4 implementation)

---

## Overview

The **Warehouse Context** manages physical goods, inventory levels, and stock movements. It's essential for e-commerce and order fulfillment workflows.

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

### 2. Inventory Aggregate

**Aggregate Root:** `Inventory`  
**Purpose:** Stock level tracking and reservations

**Entity Structure:**

```go
type Inventory struct {
    aggregate.BaseAggregate

    // Identity
    ID         uuidv7.UUID
    ProductID  uuidv7.UUID
    LocationID uuidv7.UUID

    // Stock Levels
    Available  int            // physical stock - reserved
    Reserved   int            // temporarily allocated to orders
    Physical   int            // actual physical stock

    // Tracking
    reservations []Reservation  // private
    movements    []Movement     // private

    // Lifecycle
    UpdatedAt  time.Time
}

type Reservation struct {
    ID        uuidv7.UUID
    OrderID   uuidv7.UUID
    Quantity  int
    ExpiresAt time.Time
    CreatedAt time.Time
}

type Movement struct {
    ID          uuidv7.UUID
    Type        MovementType  // in, out, transfer, adjustment
    Quantity    int
    FromLocation *uuidv7.UUID
    ToLocation   *uuidv7.UUID
    Reference   string        // order ID, PO number, etc.
    CreatedAt   time.Time
}
```

**Business Rules:**

- Available = Physical - Reserved
- Cannot reserve more than available stock
- Reservations expire after N hours (configurable)
- Negative adjustments require reason
- Physical stock cannot be negative

**Methods:**

```go
func (i *Inventory) Reserve(orderID uuidv7.UUID, quantity int, expiresAt time.Time) error
func (i *Inventory) ReleaseReservation(reservationID uuidv7.UUID) error
func (i *Inventory) CommitReservation(reservationID uuidv7.UUID) error
func (i *Inventory) AdjustStock(quantity int, reason string) error
func (i *Inventory) TransferTo(targetLocationID uuidv7.UUID, quantity int) error
func (i *Inventory) IsLowStock(reorderPoint int) bool
```

---

### 3. Location Aggregate

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
