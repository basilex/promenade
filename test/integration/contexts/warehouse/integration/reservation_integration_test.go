package integration_test

import (
    "context"
    "encoding/json"
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/basilex/promenade/internal/contexts/warehouse/integration"
    "github.com/basilex/promenade/internal/contexts/warehouse/inventory"
    inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
    "github.com/basilex/promenade/pkg/bus"
    "github.com/basilex/promenade/pkg/bus/memory"
    "github.com/basilex/promenade/pkg/uuidv7"
    testutils "github.com/basilex/promenade/test/integration"
)

// Test event implementation for Event Bus
type testEvent struct {
    eventType   string
    aggregateID uuidv7.UUID
    occurredAt  time.Time
    metadata    map[string]string
}

func (e *testEvent) Type() string                 { return e.eventType }
func (e *testEvent) AggregateID() uuidv7.UUID     { return e.aggregateID }
func (e *testEvent) OccurredAt() time.Time        { return e.occurredAt }
func (e *testEvent) Metadata() map[string]string  { return e.metadata }

// Helper: Create order.confirmed event
func createOrderConfirmedEvent(orderID uuidv7.UUID, items []integration.OrderItem, confirmedBy uuidv7.UUID) bus.Event {
	payload := map[string]interface{}{
		"order_id":     orderID.String(),
		"items":        items,
		"confirmed_by": confirmedBy.String(),
	}

	payloadJSON, _ := json.Marshal(payload)

	return &testEvent{
		eventType:   "order.confirmed",
		aggregateID: orderID,
		occurredAt:  time.Now(),
		metadata: map[string]string{
			"payload": string(payloadJSON),
		},
    }
}

// Helper: Create order.cancelled event
func createOrderCancelledEvent(orderID uuidv7.UUID, items []integration.OrderItem, cancelledBy uuidv7.UUID) bus.Event {
	payload := map[string]interface{}{
		"order_id":      orderID.String(),
		"items":         items,
		"reason":        "test cancellation",
		"cancelled_by":  cancelledBy.String(),
	}

	payloadJSON, _ := json.Marshal(payload)

	return &testEvent{
		eventType:   "order.cancelled",
		aggregateID: orderID,
		occurredAt:  time.Now(),
		metadata: map[string]string{
			"payload": string(payloadJSON),
		
		},
	}
}

// Helper: Create order.fulfilled event
func createOrderFulfilledEvent(orderID uuidv7.UUID, items []integration.OrderItem, fulfilledBy uuidv7.UUID) bus.Event {
	payload := map[string]interface{}{
		"order_id":      orderID.String(),
		"items":         items,
		"fulfilled_by":  fulfilledBy.String(),
	}

	payloadJSON, _ := json.Marshal(payload)

	return &testEvent{
		eventType:   "order.fulfilled",
		aggregateID: orderID,
		occurredAt:  time.Now(),
		metadata: map[string]string{
			"payload": string(payloadJSON),
		},
    }
}

func TestReservationIntegration_FullOrderLifecycle(t *testing.T) {
    db := testutils.SetupTestDBWithCleanTables(t)
    ctx := context.Background()

    // Setup
    invRepo := inventoryRepo.NewInventoryRepository(db.DB)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)
    eventBus := memory.NewMemoryBus(bus.NewConfig(10, 100, 3, 100*time.Millisecond, 1*time.Second, 2.0))
    defer func() {
        if err := eventBus.Close(ctx); err != nil {
            t.Logf("Failed to close event bus: %v", err)
        }
    }()

    handler := integration.NewOrderEventHandler(reservationService)
    require.NoError(t, handler.RegisterHandlers(eventBus))

    // Create inventory
    productID := uuidv7.New()
    createdBy := uuidv7.New()
    inv, err := inventoryUC.CreateInventory(ctx, productID, "SKU-001", "Product 1", "WH-001", createdBy)
    require.NoError(t, err)
    inv, err = inventoryUC.ReceiveStock(ctx, inv.ID, 100, 1000, createdBy)
    require.NoError(t, err)

    // Publish order.confirmed event
    orderID := uuidv7.New()
    confirmedBy := uuidv7.New()
    items := []integration.OrderItem{{ProductID: productID, SKU: "SKU-001", Quantity: 10}}
    event := createOrderConfirmedEvent(orderID, items, confirmedBy)
    require.NoError(t, eventBus.Publish(ctx, "order.confirmed", event))

    // Wait for processing
    time.Sleep(200 * time.Millisecond)

    // Verify reservation
    inv, err = inventoryUC.GetInventory(ctx, inv.ID)
    require.NoError(t, err)
    assert.Equal(t, 10, inv.QuantityReserved)
    assert.Equal(t, 90, inv.QuantityAvailable)

    // Publish order.fulfilled event
    fulfilledBy := uuidv7.New()
    fulfilledEvent := createOrderFulfilledEvent(orderID, items, fulfilledBy)
    require.NoError(t, eventBus.Publish(ctx, "order.fulfilled", fulfilledEvent))

    // Wait for processing
    time.Sleep(200 * time.Millisecond)

    // Verify commit
    inv, err = inventoryUC.GetInventory(ctx, inv.ID)
    require.NoError(t, err)
    assert.Equal(t, 0, inv.QuantityReserved)
    assert.Equal(t, 90, inv.QuantityAvailable)
    assert.Equal(t, 10, inv.QuantityCommitted)
}

func TestReservationIntegration_OrderCancellation(t *testing.T) {
    db := testutils.SetupTestDBWithCleanTables(t)
    ctx := context.Background()

    // Setup
    invRepo := inventoryRepo.NewInventoryRepository(db.DB)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)
    eventBus := memory.NewMemoryBus(bus.NewConfig(10, 100, 3, 100*time.Millisecond, 1*time.Second, 2.0))
    defer func() {
        if err := eventBus.Close(ctx); err != nil {
            t.Logf("Failed to close event bus: %v", err)
        }
    }()

    handler := integration.NewOrderEventHandler(reservationService)
    require.NoError(t, handler.RegisterHandlers(eventBus))

    // Create inventory
    productID := uuidv7.New()
    createdBy := uuidv7.New()
    inv, err := inventoryUC.CreateInventory(ctx, productID, "SKU-002", "Product 2", "WH-002", createdBy)
    require.NoError(t, err)
    inv, err = inventoryUC.ReceiveStock(ctx, inv.ID, 100, 1000, createdBy)
    require.NoError(t, err)

    // Reserve stock
    orderID := uuidv7.New()
    confirmedBy := uuidv7.New()
    items := []integration.OrderItem{{ProductID: productID, SKU: "SKU-002", Quantity: 20}}
    event := createOrderConfirmedEvent(orderID, items, confirmedBy)
    require.NoError(t, eventBus.Publish(ctx, "order.confirmed", event))
    time.Sleep(200 * time.Millisecond)

    // Verify reservation
    inv, err = inventoryUC.GetInventory(ctx, inv.ID)
    require.NoError(t, err)
    assert.Equal(t, 20, inv.QuantityReserved)

    // Cancel order
    cancelledBy := uuidv7.New()
    cancelEvent := createOrderCancelledEvent(orderID, items, cancelledBy)
    require.NoError(t, eventBus.Publish(ctx, "order.cancelled", cancelEvent))
    time.Sleep(200 * time.Millisecond)

    // Verify release
    inv, err = inventoryUC.GetInventory(ctx, inv.ID)
    require.NoError(t, err)
    assert.Equal(t, 0, inv.QuantityReserved)
    assert.Equal(t, 100, inv.QuantityAvailable)
}

func TestReservationIntegration_MultipleItems(t *testing.T) {
    db := testutils.SetupTestDBWithCleanTables(t)
    ctx := context.Background()

    // Setup
    invRepo := inventoryRepo.NewInventoryRepository(db.DB)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)
    eventBus := memory.NewMemoryBus(bus.NewConfig(10, 100, 3, 100*time.Millisecond, 1*time.Second, 2.0))
    defer func() {
        if err := eventBus.Close(ctx); err != nil {
            t.Logf("Failed to close event bus: %v", err)
        }
    }()

    handler := integration.NewOrderEventHandler(reservationService)
    require.NoError(t, handler.RegisterHandlers(eventBus))

    // Create multiple inventories
    createdBy := uuidv7.New()
    product1 := uuidv7.New()
    product2 := uuidv7.New()
    product3 := uuidv7.New()

    inv1, err := inventoryUC.CreateInventory(ctx, product1, "SKU-A", "Product A", "WH-MULTI", createdBy)
    require.NoError(t, err)
    inv1, err = inventoryUC.ReceiveStock(ctx, inv1.ID, 100, 1000, createdBy)
    require.NoError(t, err)

    inv2, err := inventoryUC.CreateInventory(ctx, product2, "SKU-B", "Product B", "WH-MULTI", createdBy)
    require.NoError(t, err)
    inv2, err = inventoryUC.ReceiveStock(ctx, inv2.ID, 50, 2000, createdBy)
    require.NoError(t, err)

    inv3, err := inventoryUC.CreateInventory(ctx, product3, "SKU-C", "Product C", "WH-MULTI", createdBy)
    require.NoError(t, err)
    inv3, err = inventoryUC.ReceiveStock(ctx, inv3.ID, 75, 1500, createdBy)
    require.NoError(t, err)

    // Reserve all three products
    orderID := uuidv7.New()
    confirmedBy := uuidv7.New()
    items := []integration.OrderItem{
        {ProductID: product1, SKU: "SKU-A", Quantity: 10},
        {ProductID: product2, SKU: "SKU-B", Quantity: 5},
        {ProductID: product3, SKU: "SKU-C", Quantity: 15},
    }
    event := createOrderConfirmedEvent(orderID, items, confirmedBy)
    require.NoError(t, eventBus.Publish(ctx, "order.confirmed", event))
    time.Sleep(300 * time.Millisecond)

    // Verify all reservations
    inv1, err = inventoryUC.GetInventory(ctx, inv1.ID)
    require.NoError(t, err)
    assert.Equal(t, 10, inv1.QuantityReserved)
    assert.Equal(t, 90, inv1.QuantityAvailable)

    inv2, err = inventoryUC.GetInventory(ctx, inv2.ID)
    require.NoError(t, err)
    assert.Equal(t, 5, inv2.QuantityReserved)
    assert.Equal(t, 45, inv2.QuantityAvailable)

    inv3, err = inventoryUC.GetInventory(ctx, inv3.ID)
    require.NoError(t, err)
    assert.Equal(t, 15, inv3.QuantityReserved)
    assert.Equal(t, 60, inv3.QuantityAvailable)
}

func TestReservationIntegration_InsufficientStock(t *testing.T) {
    db := testutils.SetupTestDBWithCleanTables(t)
    ctx := context.Background()

    // Setup
    invRepo := inventoryRepo.NewInventoryRepository(db.DB)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)

    // Create inventory with low stock
    productID := uuidv7.New()
    createdBy := uuidv7.New()
    inv, err := inventoryUC.CreateInventory(ctx, productID, "SKU-LOW", "Low Stock", "WH-LOW", createdBy)
    require.NoError(t, err)
    _, err = inventoryUC.ReceiveStock(ctx, inv.ID, 5, 1000, createdBy)
    require.NoError(t, err)

    // Attempt to reserve more than available
    orderID := uuidv7.New()
    err = reservationService.ReserveForOrder(ctx, orderID, []integration.OrderItem{
        {ProductID: productID, SKU: "SKU-LOW", Quantity: 10},
    }, createdBy)

    // Should fail
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "insufficient stock")
}

func TestReservationIntegration_IdempotentRelease(t *testing.T) {
    db := testutils.SetupTestDBWithCleanTables(t)
    ctx := context.Background()

    // Setup
    invRepo := inventoryRepo.NewInventoryRepository(db.DB)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)

    // Create inventory
    productID := uuidv7.New()
    createdBy := uuidv7.New()
    inv, err := inventoryUC.CreateInventory(ctx, productID, "SKU-IDEM", "Idempotent Test", "WH-IDEM", createdBy)
    require.NoError(t, err)
    inv, err = inventoryUC.ReceiveStock(ctx, inv.ID, 100, 1000, createdBy)
    require.NoError(t, err)

    // Reserve stock
    orderID := uuidv7.New()
    items := []integration.OrderItem{
        {ProductID: productID, SKU: "SKU-IDEM", Quantity: 10},
    }
    err = reservationService.ReserveForOrder(ctx, orderID, items, createdBy)
    require.NoError(t, err)

    // Release once
    err = reservationService.ReleaseForOrder(ctx, orderID, items, createdBy)
    require.NoError(t, err)

    inv, err = inventoryUC.GetInventory(ctx, inv.ID)
    require.NoError(t, err)
    assert.Equal(t, 0, inv.QuantityReserved)
    assert.Equal(t, 100, inv.QuantityAvailable)

    // Release again (idempotent - should not error)
    err = reservationService.ReleaseForOrder(ctx, orderID, items, createdBy)
    assert.NoError(t, err)

    // Verify still released
    inv, err = inventoryUC.GetInventory(ctx, inv.ID)
    require.NoError(t, err)
    assert.Equal(t, 0, inv.QuantityReserved)
    assert.Equal(t, 100, inv.QuantityAvailable)
}

func TestReservationIntegration_ConcurrentReservations(t *testing.T) {
    db := testutils.SetupTestDBWithCleanTables(t)
    ctx := context.Background()

    // Setup
    invRepo := inventoryRepo.NewInventoryRepository(db.DB)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)

    // Create inventory
    productID := uuidv7.New()
    createdBy := uuidv7.New()
    inv, err := inventoryUC.CreateInventory(ctx, productID, "SKU-CONC", "Concurrent Test", "WH-CONC", createdBy)
    require.NoError(t, err)
    inv, err = inventoryUC.ReceiveStock(ctx, inv.ID, 100, 1000, createdBy)
    require.NoError(t, err)

    // Concurrent reservations
    var wg sync.WaitGroup
    successCount := 0
    var mu sync.Mutex

    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            orderID := uuidv7.New()
            err := reservationService.ReserveForOrder(ctx, orderID, []integration.OrderItem{
                {ProductID: productID, SKU: "SKU-CONC", Quantity: 10},
            }, createdBy)
            if err == nil {
                mu.Lock()
                successCount++
                mu.Unlock()
            }
        }()
    }

    wg.Wait()

    // In concurrent scenario without retry logic, expect 1-10 successful reservations
    // Version conflicts cause some goroutines to fail - this is expected behavior
    assert.GreaterOrEqual(t, successCount, 1, "At least one reservation should succeed")
    assert.LessOrEqual(t, successCount, 10, "At most 10 reservations should succeed (100 / 10)")

    // Verify final state matches successful reservations
    inv, err = inventoryUC.GetInventory(ctx, inv.ID)
    require.NoError(t, err)
    assert.Equal(t, successCount*10, inv.QuantityReserved, "Reserved quantity should match successful count")
    assert.Equal(t, 100-(successCount*10), inv.QuantityAvailable, "Available quantity should match reserved")
}

func TestReservationIntegration_ProductNotFound(t *testing.T) {
    db := testutils.SetupTestDBWithCleanTables(t)
    ctx := context.Background()

    // Setup
    invRepo := inventoryRepo.NewInventoryRepository(db.DB)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)

    // Attempt to reserve non-existent product
    orderID := uuidv7.New()
    nonExistentProduct := uuidv7.New()
    createdBy := uuidv7.New()
    
    err := reservationService.ReserveForOrder(ctx, orderID, []integration.OrderItem{
        {ProductID: nonExistentProduct, SKU: "SKU-NONE", Quantity: 10},
    }, createdBy)

    // Should fail
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}