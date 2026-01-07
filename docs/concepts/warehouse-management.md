# Warehouse Management

**Complete inventory management system** with stock tracking, product catalog, warehouse locations, and **automated order-inventory integration**.

---

## Overview

The **Warehouse Context** manages physical goods, inventory levels, stock movements, and warehouse locations. It provides the foundation for e-commerce operations and order fulfillment workflows with **automated integration** between orders and inventory.

### Key Capabilities

- **Product Catalog**: Manage product specifications, SKUs, categories, and brands
- **Inventory Tracking**: Real-time stock levels with reservations and availability
- **Stock Movements**: Audit trail for all inventory changes (receipts, transfers, adjustments)
- **Warehouse Locations**: Hierarchical location management (warehouses, zones, bins)
- **Order Integration**: Automated stock reservation when orders are confirmed

---

## Core Aggregates

### 1. Product Aggregate

**Purpose**: Product catalog and specifications

**Key Features**:
- SKU-based product identification
- Category and brand organization
- Physical properties (weight, dimensions, packaging)
- Inventory settings (tracking enabled, reorder points)
- Product lifecycle (active, inactive, discontinued)

**Business Operations**:
- Create/update products with validation
- Activate/deactivate/discontinue products
- Track inventory for each product
- Set reorder thresholds

**Status**: Production ready (139 tests passing)

### 2. Inventory Aggregate

**Purpose**: Real-time stock tracking and availability

**Key Features**:
- Quantity tracking (on hand, reserved, available, committed)
- Warehouse and location assignment
- Cost tracking (weighted average)
- Reorder management (min/max thresholds)
- Stock operations (receive, reserve, release, commit)

**Business Operations**:
- Receive stock from suppliers
- Reserve stock for orders
- Release cancelled reservations
- Commit fulfilled stock
- Track low stock items

**Status**: Production ready (141 tests passing)

### 3. StockMovement Aggregate

**Purpose**: Immutable audit trail for all inventory changes

**Key Features**:
- Movement types (receipt, reservation, release, commit, adjustment, transfer, damage, return)
- Quantity snapshots (before/after each movement)
- Cost tracking per movement
- Reference tracking (links to orders, POs, adjustments)
- Location tracking (from/to warehouses for transfers)

**Business Operations**:
- Record all stock movements automatically
- Audit inventory history
- Track inventory by reference (order ID, PO number)
- Generate movement summaries

**Status**: Production ready (45 tests passing)

### 4. Location Aggregate

**Purpose**: Warehouse location management and organization

**Key Features**:
- Hierarchical structure (warehouse → zone → bin)
- Location types (warehouse, retail, dropship, virtual)
- Capacity management
- Physical dimensions
- Active/inactive status
- Maintenance mode support

**Business Operations**:
- Create location hierarchies
- Manage warehouse capacity
- Track location utilization
- Enable/disable locations

**Status**: Production ready (74 tests passing)

---

## Warehouse Integration

**The most powerful feature** - automated synchronization between Order Management and Warehouse contexts via Event Bus.

### Architecture

```
Order Management Context          Event Bus          Warehouse Context
       ↓                              ↓                     ↓
   Order.Confirm()  ────→  order.confirmed event  ────→  ReservationService
                                                          Reserve inventory
       
   Order.Cancel()   ────→  order.cancelled event  ────→  ReservationService
                                                          Release reservation
       
   Order.Fulfill()  ────→  order.fulfilled event  ────→  ReservationService
                                                          Commit stock
```

### Components

#### ReservationService

**Purpose**: Business logic for order-inventory reservations

**Key Operations**:
```go
// Reserve inventory for order (triggered by order.confirmed event)
ReserveForOrder(ctx, orderID, orderLines) error

// Release reservation (triggered by order.cancelled event)
ReleaseForOrder(ctx, orderID) error

// Commit reserved stock (triggered by order.fulfilled event)
CommitForOrder(ctx, orderID) error
```

**Business Rules**:
- Checks product availability before reservation
- Returns error if insufficient stock (order stays pending)
- Handles partial reservations (all-or-nothing)
- Idempotent operations (safe to retry)
- Tracks reservations by order ID

**Implementation**: `internal/contexts/warehouse/integration/reservation_service.go` (230 lines)

#### OrderEventHandler

**Purpose**: Event Bus integration for order lifecycle events

**Event Handlers**:
- `order.confirmed` → calls `ReserveForOrder()`
- `order.cancelled` → calls `ReleaseForOrder()`
- `order.fulfilled` → calls `CommitForOrder()`

**Error Handling**:
- Logs errors but doesn't fail event processing
- Supports retry via Event Bus retry policy
- Graceful degradation if warehouse unavailable

**Implementation**: `internal/contexts/warehouse/integration/order_event_handler.go` (230 lines)

### Bootstrap Integration

Warehouse Integration initializes automatically during application startup:

```go
// cmd/api/bootstrap.go
func initWarehouseIntegration(db *sqlx.DB, eventBus bus.IBus) (*integration.OrderEventHandler, error) {
    // Initialize Inventory Use Case
    invRepo := inventoryRepo.NewInventoryRepository(db)
    inventoryUC := inventory.NewUseCase(invRepo)
    
    // Initialize Reservation Service
    reservationService := integration.NewReservationService(inventoryUC)
    
    // Initialize Order Event Handler
    orderEventHandler := integration.NewOrderEventHandler(reservationService)
    
    // Register event handlers with Event Bus
    if err := orderEventHandler.RegisterHandlers(eventBus); err != nil {
        return nil, err
    }
    
    logger.Info("Warehouse Integration initialized",
        slog.String("component", "ReservationService + OrderEventHandler"),
        slog.Int("event_handlers", 3),
    )
    
    return orderEventHandler, nil
}
```

**Initialization logs**:
```
level=INFO msg="Event Bus initialized" adapter=memory worker_pool_size=10
level=INFO msg="Warehouse Integration initialized" component="ReservationService + OrderEventHandler" event_handlers=3
```

### End-to-End Flow

**Order Confirmation Flow**:

1. **Order Management**: Customer confirms order
   ```go
   order.Confirm() // Changes status to "confirmed"
   eventBus.Publish(ctx, "order.confirmed", orderEvent)
   ```

2. **Event Bus**: Delivers event to registered handlers
   ```
   order.confirmed → OrderEventHandler.HandleOrderConfirmed()
   ```

3. **Warehouse Integration**: Reserves inventory
   ```go
   reservationService.ReserveForOrder(ctx, orderID, orderLines)
   // For each line item:
   // - Find inventory by SKU
   // - Check availability
   // - Reserve quantity
   // - Record stock movement
   ```

4. **Result**: 
   - Success: Stock reserved, order proceeds to processing
   - Failure: Error logged, order stays pending (manual intervention)

**Order Cancellation Flow**:

1. **Order Management**: Order cancelled
   ```go
   order.Cancel(reason)
   eventBus.Publish(ctx, "order.cancelled", orderEvent)
   ```

2. **Warehouse Integration**: Releases reservation
   ```go
   reservationService.ReleaseForOrder(ctx, orderID)
   // - Find reservations by order ID
   // - Release reserved quantity
   // - Record release movement
   ```

**Order Fulfillment Flow**:

1. **Order Management**: Order fulfilled
   ```go
   order.MarkFulfilled()
   eventBus.Publish(ctx, "order.fulfilled", orderEvent)
   ```

2. **Warehouse Integration**: Commits stock
   ```go
   reservationService.CommitForOrder(ctx, orderID)
   // - Find reservations by order ID
   // - Commit quantity (decrease on-hand stock)
   // - Record commit movement
   ```

---

## Benefits

### 1. Decoupling

**Order Management** and **Warehouse** contexts are completely decoupled:
- No direct imports between contexts
- No compile-time dependencies
- Independent deployment and scaling
- Can swap implementations without breaking contracts

### 2. Event-Driven Architecture

**Asynchronous processing** improves performance:
- Order confirmation doesn't wait for inventory reservation
- Event Bus handles retries automatically
- Failure in one context doesn't crash the other
- Graceful degradation if warehouse unavailable

### 3. Auditability

**Complete audit trail** for all inventory changes:
- Every reservation recorded in StockMovement
- Track which order caused which stock change
- Immutable append-only log
- Reference tracking (order ID → stock movements)

### 4. Scalability

**Independent scaling** of contexts:
- Order Management can scale separately from Warehouse
- Event Bus supports distributed systems (Redis adapter)
- Horizontal scaling without coordination

### 5. Business Rules Enforcement

**Inventory constraints** enforced at reservation:
- Cannot reserve more than available
- Automatic low stock detection
- Prevents overselling
- All-or-nothing reservation (transactional)

---

## Testing

### Test Coverage

| Component                 | Tests | Status | Type        |
|---------------------------|-------|--------|-------------|
| Product Aggregate         | 139   | PASS   | Unit        |
| Inventory Aggregate       | 141   | PASS   | Unit        |
| StockMovement Aggregate   | 45    | PASS   | Unit        |
| Location Aggregate        | 74    | PASS   | Unit        |
| ReservationService (unit) | 25    | PASS   | Unit        |
| Warehouse Integration     | 9     | PASS   | Integration |
| **Total**                 | **433** | **PASS** | **Mixed** |

### Integration Tests

**E2E scenarios with real database and Event Bus**:

1. **Full Order Lifecycle** - Order created → confirmed → fulfilled → inventory committed
2. **Order Cancellation** - Order confirmed → cancelled → inventory released
3. **Multiple Items** - Reserve inventory for order with multiple line items
4. **Insufficient Stock** - Order cannot be reserved if stock unavailable
5. **Idempotent Release** - Releasing same order multiple times is safe
6. **Concurrent Reservations** - Multiple orders reserving same product
7. **Product Not Found** - Graceful handling of invalid SKUs

**Test duration**: ~4.2 seconds (9 tests)

---

## Configuration

Warehouse Integration requires Event Bus to be initialized first:

```yaml
# config/app.postgres-dev.yaml
bus:
  adapter: "memory"  # or "redis" for production
  worker_pool_size: 10
  buffer_size: 1000
  retry_attempts: 3
  retry_delay: 100ms
```

**No additional configuration needed** - Warehouse Integration uses existing Event Bus and database connections.

---

## API Endpoints

**Total**: 58 endpoints across 4 aggregates

### Product Endpoints (16)

- `POST /api/v1/warehouse/products` - Create product
- `GET /api/v1/warehouse/products/:id` - Get by ID
- `GET /api/v1/warehouse/products/sku/:sku` - Get by SKU
- `GET /api/v1/warehouse/products` - List all
- `GET /api/v1/warehouse/products/category/:category` - Filter by category
- `POST /api/v1/warehouse/products/:id/activate` - Activate
- `POST /api/v1/warehouse/products/:id/deactivate` - Deactivate
- And more...

### Inventory Endpoints (14)

- `POST /api/v1/warehouse/inventory` - Create inventory
- `GET /api/v1/warehouse/inventory/:id` - Get by ID
- `GET /api/v1/warehouse/inventory/sku/:sku` - Get by SKU
- `POST /api/v1/warehouse/inventory/:id/receive` - Receive stock
- `POST /api/v1/warehouse/inventory/:id/reserve` - Reserve stock
- `POST /api/v1/warehouse/inventory/:id/release` - Release reservation
- `POST /api/v1/warehouse/inventory/:id/commit` - Commit stock
- And more...

### StockMovement Endpoints (7)

- `POST /api/v1/warehouse/stock-movements` - Record movement
- `GET /api/v1/warehouse/stock-movements/:id` - Get by ID
- `GET /api/v1/warehouse/stock-movements/inventory/:inventory_id` - Get by inventory
- `GET /api/v1/warehouse/stock-movements/reference` - Get by reference (order ID, PO number)
- And more...

### Location Endpoints (14)

- `POST /api/v1/warehouse/locations` - Create location
- `GET /api/v1/warehouse/locations/:id` - Get by ID
- `GET /api/v1/warehouse/locations/code/:code` - Get by code
- `GET /api/v1/warehouse/locations/:id/hierarchy` - Get hierarchy
- And more...

---

## Best Practices

### DO

- **Use Event Bus for cross-context communication** - Never call Warehouse from Order Management directly
- **Check inventory before order confirmation** - Prevent overselling
- **Record all stock movements** - Maintain complete audit trail
- **Use SKU for product identification** - Unique and immutable
- **Set reorder points** - Automate low stock alerts
- **Test with real Event Bus** - Integration tests verify actual behavior

### DON'T

- **Don't skip reservation step** - Always reserve before fulfilling
- **Don't directly modify inventory** - Use business methods (receive, reserve, commit)
- **Don't delete stock movements** - They're immutable audit records
- **Don't ignore insufficient stock errors** - Handle gracefully in Order Management
- **Don't bypass Event Bus** - Maintain architectural boundaries

---

## Future Enhancements

- **Low Stock Alerts System** - Automatic notifications when stock below reorder point
- **Multi-warehouse Transfer Workflows** - Move inventory between locations
- **Batch Operations** - Bulk receive/reserve/commit for performance
- **Inventory Reporting** - Analytics and dashboards
- **Serial Number Tracking** - Track individual items (electronics, high-value goods)
- **Lot/Batch Tracking** - Track groups of items (expiration dates, manufacturing dates)

---

## Related Documentation

- [Warehouse Context README](../../internal/contexts/warehouse/README.md) - Complete technical reference (1087 lines)
- [ReservationService Source](../../internal/contexts/warehouse/integration/reservation_service.go) - Implementation (230 lines)
- [OrderEventHandler Source](../../internal/contexts/warehouse/integration/order_event_handler.go) - Event handlers (230 lines)
- [Integration Tests](../../test/integration/contexts/warehouse/integration/reservation_integration_test.go) - E2E tests (417 lines)
- [Order Management Guide](order-management.md) - Order lifecycle and events
- [Event Bus Guide](../../pkg/bus/README.md) - Event-driven architecture
- [Clean Architecture](clean-architecture.md) - Bounded Contexts and DDD

---

**Status**: Production Ready  
**Version**: 1.0.0  
**Last Updated**: January 7, 2026  
**Maintainer**: Promenade Team
