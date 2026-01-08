# Warehouse Management

**Complete inventory management system** with stock tracking, product catalog, warehouse locations, and **automated order-inventory integration**.

## Overview

Warehouse Management is a **Bounded Context** that handles inventory tracking, stock movements, product catalog, and warehouse locations. The most powerful feature is **Warehouse Integration** - automated synchronization with Order Management via Event Bus.

**Status**: Production Ready (100% complete - January 2026)  
**Aggregates**: Product, Inventory, StockMovement, Location  
**Endpoints**: 58 HTTP routes  
**Tests**: 433 passing (139 Product + 141 Inventory + 45 StockMovement + 74 Location + 34 Integration)  
**Database**: 7 tables with audit trails

---

## Key Features

### Product Catalog
Manage products with SKUs, categories, brands, and physical properties.

**Capabilities**:
- SKU-based product identification
- Categorization and search
- Brand management
- Physical properties (weight, dimensions)
- Lifecycle management (active, discontinued)

**16 API Endpoints**: Complete CRUD + business operations

### Inventory Tracking
Real-time stock level tracking with multiple quantity states.

**Stock States**:
- **QuantityOnHand**: Physical stock in warehouse
- **QuantityReserved**: Reserved for orders (pending fulfillment)
- **QuantityAvailable**: Available for new orders (OnHand - Reserved)
- **QuantityCommitted**: Shipped/delivered stock

**Reorder Management**:
- ReorderPoint threshold
- MinStock and MaxStock levels
- Automatic reorder suggestions

**14 API Endpoints**: CRUD + stock operations (receive, reserve, release, commit)

### Stock Movements
Immutable audit trail for all inventory changes.

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
- Before/After snapshots
- Cost tracking per movement
- Reference to orders/POs
- Audit compliance

**7 API Endpoints**: Create and query movements

### Warehouse Locations
Hierarchical location management for warehouses, zones, and bins.

**Location Types**:
- **Warehouse**: Main facility
- **Retail**: Store location
- **Dropship**: Supplier warehouse
- **Virtual**: Digital products

**Features**:
- Full address management
- Capacity tracking
- Active/inactive status
- Country integration

**14 API Endpoints**: Complete CRUD + location operations

---

## Warehouse Integration

**The most powerful feature** - automated order-inventory synchronization via Event Bus.

### Architecture

```
Order Management Context          Event Bus          Warehouse Context
                      
 Order.Confirm()          ReservationService 
                                                .ReserveForOrder() 
                      

                      
 Order.Cancel()           ReservationService 
                                                .ReleaseForOrder() 
                      

                      
 Order.Fulfill()          ReservationService 
                                                .CommitForOrder()  
                      
```

### Components

**1. ReservationService** (230 lines)

Business logic for stock operations:

```go
func (s *ReservationService) ReserveForOrder(ctx context.Context, orderID uuid.UUID, items []OrderItem) error
func (s *ReservationService) ReleaseForOrder(ctx context.Context, orderID uuid.UUID) error
func (s *ReservationService) CommitForOrder(ctx context.Context, orderID uuid.UUID) error
```

**2. OrderEventHandler** (230 lines)

Event handlers for order lifecycle:

```go
func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event bus.Event) error
func (h *OrderEventHandler) HandleOrderCancelled(ctx context.Context, event bus.Event) error
func (h *OrderEventHandler) HandleOrderFulfilled(ctx context.Context, event bus.Event) error
```

**3. Bootstrap Integration**

Initialized in `cmd/api/bootstrap.go`:

```go
func initWarehouseIntegration(db *sqlx.DB, eventBus bus.IBus) (*integration.OrderEventHandler, error) {
    // Initialize Inventory Use Case
    invRepo := inventoryRepo.NewInventoryRepository(db)
    inventoryUC := inventory.NewUseCase(invRepo)

    // Initialize Reservation Service
    reservationService := integration.NewReservationService(inventoryUC)

    // Initialize Order Event Handler
    orderEventHandler := integration.NewOrderEventHandler(reservationService)

    // Register event handlers
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
5. Warehouse: For each line item:
   - Validate inventory availability
   - Call `Inventory.ReserveStock(quantity, orderID)`
   - Create StockMovement record (type: Reservation)
6. Result: Stock reserved, order can proceed to fulfillment

**Order Cancellation** (order.cancelled):
1. Order Management: `Order.Cancel()` publishes `order.cancelled` event
2. Event Bus: Routes event to Warehouse Context
3. Warehouse: `OrderEventHandler.HandleOrderCancelled()`
4. Warehouse: `ReservationService.ReleaseForOrder()`
5. Warehouse: For each line item:
   - Call `Inventory.ReleaseReservation(quantity, orderID)`
   - Create StockMovement record (type: ReservationRelease)
6. Result: Stock returned to available pool

**Order Fulfillment** (order.fulfilled):
1. Order Management: `Order.MarkFulfilled()` publishes `order.fulfilled` event
2. Event Bus: Routes event to Warehouse Context
3. Warehouse: `OrderEventHandler.HandleOrderFulfilled()`
4. Warehouse: `ReservationService.CommitForOrder()`
5. Warehouse: For each line item:
   - Call `Inventory.CommitStock(quantity, orderID)`
   - Create StockMovement record (type: Commit)
6. Result: Stock permanently removed from inventory

---

## Benefits

### Decoupling
- **No Direct Dependencies**: Order Management doesn't know about Warehouse
- **Context Isolation**: Each context evolves independently
- **API Stability**: Changes in one context don't break the other
- **Testability**: Can test contexts in isolation with mocked events

### Event-Driven Architecture
- **Asynchronous Processing**: Orders don't wait for inventory updates
- **Retry Mechanism**: Failed stock operations retry automatically
- **Graceful Degradation**: System continues if Event Bus temporarily unavailable
- **Audit Trail**: All events logged for compliance

### Auditability
- **Complete Audit Trail**: Every stock change recorded in StockMovement
- **Before/After Snapshots**: QuantityBeforeMove and QuantityAfterMove
- **Reference Tracking**: Links to orders, POs, adjustments
- **Immutable Records**: StockMovements never modified, only created
- **Compliance Ready**: Full audit history for inventory regulations

### Scalability
- **Independent Scaling**: Scale Order and Warehouse contexts separately
- **Event Bus Buffering**: Handles traffic spikes
- **Horizontal Scaling**: Multiple Warehouse instances process events
- **Distributed Systems**: Redis Event Bus supports multi-instance deployments

### Business Rules Enforcement
- **Inventory Constraints**: Prevent overselling (available < quantity)
- **Automatic Validation**: Stock checks before reservation
- **Consistent State**: Events ensure eventual consistency
- **Error Handling**: Failed reservations bubble up to Order Management

---

## Testing

### Test Coverage

| Component          | Tests | Type        | Description                                    |
| ------------------ | ----- | ----------- | ---------------------------------------------- |
| **Product**        | 139   | All         | Entity (25) + UseCase (83) + Smoke (10) + Integration (21) |
| **Inventory**      | 141   | All         | Entity (97) + UseCase (?) + Integration (23) + Smoke (21) |
| **StockMovement**  | 45    | All         | Entity (11) + UseCase (10) + Smoke (9) + Integration (15) |
| **Location**       | 74    | All         | Entity (48) + Integration (17) + Smoke (9)    |
| **Integration**    | 34    | E2E         | Order-Warehouse synchronization tests         |
| **Total**          | **433** | **All**   | **100% passing** (~4.2s execution)            |

### Integration Tests

**End-to-End Testing** (34 tests):

1. **Order Confirmation Flow**:
   - Create order with line items
   - Confirm order (triggers event)
   - Verify stock reserved in Inventory
   - Verify StockMovement records created
   - Verify QuantityReserved increased

2. **Order Cancellation Flow**:
   - Confirm order (reserves stock)
   - Cancel order (triggers event)
   - Verify reservations released
   - Verify QuantityAvailable restored
   - Verify StockMovement records for release

3. **Order Fulfillment Flow**:
   - Confirm order (reserves stock)
   - Fulfill order (triggers event)
   - Verify stock committed
   - Verify QuantityOnHand decreased
   - Verify QuantityCommitted increased

4. **Edge Cases**:
   - Insufficient inventory (order confirmation fails)
   - Multiple concurrent reservations
   - Partial fulfillment scenarios
   - Event retry on transient failures

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
  retry_max_delay: 30s       # Maximum retry delay
  retry_multiplier: 2.0      # Exponential backoff
```

**No Additional Configuration**: Warehouse Integration uses existing database and Event Bus connections.

---

## API Endpoints

### Product Endpoints (16)

```
POST   /api/v1/warehouse/products           # Create product
GET    /api/v1/warehouse/products/:id       # Get by ID
GET    /api/v1/warehouse/products           # List products
PUT    /api/v1/warehouse/products/:id       # Update product
DELETE /api/v1/warehouse/products/:id       # Soft delete
GET    /api/v1/warehouse/products/:id/inventory  # Get inventory for product
```

### Inventory Endpoints (14)

```
POST   /api/v1/warehouse/inventory          # Create inventory
GET    /api/v1/warehouse/inventory/:id      # Get by ID
PUT    /api/v1/warehouse/inventory/:id      # Update inventory
DELETE /api/v1/warehouse/inventory/:id      # Soft delete
POST   /api/v1/warehouse/inventory/:id/receive    # Receive stock
POST   /api/v1/warehouse/inventory/:id/reserve    # Reserve stock
POST   /api/v1/warehouse/inventory/:id/release    # Release reservation
POST   /api/v1/warehouse/inventory/:id/commit     # Commit stock
```

### StockMovement Endpoints (7)

```
POST   /api/v1/warehouse/stock-movements    # Create movement
GET    /api/v1/warehouse/stock-movements/:id  # Get by ID
GET    /api/v1/warehouse/stock-movements    # List movements
GET    /api/v1/warehouse/stock-movements/inventory/:id  # By inventory
GET    /api/v1/warehouse/stock-movements/recent  # Recent movements
```

### Location Endpoints (14)

```
POST   /api/v1/warehouse/locations          # Create location
GET    /api/v1/warehouse/locations/:id      # Get by ID
GET    /api/v1/warehouse/locations          # List locations
PUT    /api/v1/warehouse/locations/:id      # Update location
DELETE /api/v1/warehouse/locations/:id      # Soft delete
```

**Total**: 58 endpoints

---

## Best Practices

### DO

- **Use Event Bus**: Always communicate between contexts via events
- **Check Inventory**: Validate stock availability before order confirmation
- **Record Movements**: Create StockMovement for every inventory change
- **Handle Failures**: Implement retry logic and error handling
- **Test E2E**: Integration tests for order-warehouse flows
- **Monitor Events**: Track event processing latency and failures

### DON'T

- **Skip Reservation**: Never ship orders without reserving stock first
- **Modify Directly**: Don't update Inventory without creating StockMovement
- **Delete Movements**: StockMovements are immutable, never delete
- **Ignore Events**: Always handle order lifecycle events
- **Hardcode Logic**: Use configuration for thresholds and limits
- **Bypass Event Bus**: Don't call Warehouse methods directly from Order context

---

## Future Enhancements

- [ ] **Low Stock Alerts**: Notify when inventory below reorder point
- [ ] **Multi-warehouse Transfers**: Automated transfers between warehouses
- [ ] **Batch Operations**: Bulk stock movements for efficiency
- [ ] **Reporting Dashboard**: Real-time inventory analytics
- [ ] **Serial/Lot Tracking**: Track individual units (e.g., electronics, pharma)

---

## Related Documentation

**Technical Details**:
- [Warehouse Context README](../contexts/warehouse/README.md) - Implementation details (1087 lines)
- [ReservationService Source](../../../internal/contexts/warehouse/integration/reservation_service.go) - Business logic
- [OrderEventHandler Source](../../../internal/contexts/warehouse/integration/order_event_handler.go) - Event handlers
- [Integration Tests](../../../test/integration/contexts/warehouse/integration/) - E2E tests

**Related Concepts**:
- [Order Management](./order-management.md) - Order lifecycle and state machine
- [Event-Driven Architecture](./event-driven.md) - Event Bus patterns
- [Clean Architecture](./clean-architecture.md) - Bounded Contexts and DDD

**Implementation Guides**:
- [Testing Patterns](../guide/testing-patterns.md) - Four-tier testing strategy
- [Bounded Contexts](./bounded-contexts.md) - Context isolation patterns

---

**Status**: Production Ready (100% complete)  
**Last Updated**: January 7, 2026  
**Test Coverage**: 433 tests, 100% passing  
**Main Feature**: Automated order-inventory synchronization via Event Bus
