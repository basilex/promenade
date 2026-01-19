package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockInventoryUseCase implements inventory.IUseCase for testing
type MockInventoryUseCase struct {
	GetBySKUFunc        func(ctx context.Context, sku string) (*aggregate.Inventory, error)
	GetByProductIDFunc  func(ctx context.Context, productID uuidv7.UUID) ([]*aggregate.Inventory, error)
	UpdateInventoryFunc func(ctx context.Context, inv *aggregate.Inventory) error
}

func (m *MockInventoryUseCase) GetBySKU(ctx context.Context, sku string) (*aggregate.Inventory, error) {
	if m.GetBySKUFunc != nil {
		return m.GetBySKUFunc(ctx, sku)
	}
	return nil, fmt.Errorf("GetBySKU not implemented")
}

func (m *MockInventoryUseCase) GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*aggregate.Inventory, error) {
	if m.GetByProductIDFunc != nil {
		return m.GetByProductIDFunc(ctx, productID)
	}
	return nil, fmt.Errorf("GetByProductID not implemented")
}

func (m *MockInventoryUseCase) UpdateInventory(ctx context.Context, inv *aggregate.Inventory) error {
	if m.UpdateInventoryFunc != nil {
		return m.UpdateInventoryFunc(ctx, inv)
	}
	return nil
}

// Stub implementations for remaining IUseCase methods (not used in reservation service)
func (m *MockInventoryUseCase) CreateInventory(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*aggregate.Inventory, error) {
	return nil, fmt.Errorf("CreateInventory not implemented")
}

func (m *MockInventoryUseCase) GetInventory(ctx context.Context, id uuidv7.UUID) (*aggregate.Inventory, error) {
	return nil, fmt.Errorf("GetInventory not implemented")
}

func (m *MockInventoryUseCase) GetByWarehouse(ctx context.Context, warehouseID string) ([]*aggregate.Inventory, error) {
	return nil, fmt.Errorf("GetByWarehouse not implemented")
}

func (m *MockInventoryUseCase) GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*aggregate.Inventory, error) {
	return nil, fmt.Errorf("GetByLocation not implemented")
}

func (m *MockInventoryUseCase) GetLowStock(ctx context.Context) ([]*aggregate.Inventory, error) {
	return nil, fmt.Errorf("GetLowStock not implemented")
}

func (m *MockInventoryUseCase) ListInventory(ctx context.Context, page, pageSize int) ([]*aggregate.Inventory, int64, error) {
	return nil, 0, fmt.Errorf("ListInventory not implemented")
}

func (m *MockInventoryUseCase) DeleteInventory(ctx context.Context, id uuidv7.UUID) error {
	return fmt.Errorf("DeleteInventory not implemented")
}

func (m *MockInventoryUseCase) ReceiveStock(ctx context.Context, inventoryID uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*aggregate.Inventory, error) {
	return nil, fmt.Errorf("ReceiveStock not implemented")
}

func (m *MockInventoryUseCase) CommitStock(ctx context.Context, inventoryID uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*aggregate.Inventory, error) {
	return nil, fmt.Errorf("CommitStock not implemented")
}

// createTestInventory creates an inventory entity for testing
func createTestInventory(productID uuidv7.UUID, sku string, onHand, reserved int) *aggregate.Inventory {
	inv, _ := aggregate.NewInventory(
		productID,
		sku,
		"Test Product", // productName
		"WH-001",       // warehouseID
		uuidv7.New(),   // createdBy
	)
	// Manually set quantities
	if onHand > 0 {
		_ = inv.ReceiveStock(onHand, 1000, uuidv7.New()) // Use cents
	}
	if reserved > 0 {
		_ = inv.ReserveStock(reserved, uuidv7.New(), uuidv7.New())
	}
	return inv
}

// Test ReserveForOrder - Success with single item
func TestReservationService_ReserveForOrder_Success_SingleItem(t *testing.T) {
	productID := uuidv7.New()
	orderID := uuidv7.New()
	userID := uuidv7.New()

	inv := createTestInventory(productID, "SKU-001", 100, 0)

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			if sku == "SKU-001" {
				return inv, nil
			}
			return nil, fmt.Errorf("inventory not found")
		},
		UpdateInventoryFunc: func(ctx context.Context, i *aggregate.Inventory) error {
			return nil
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{{ProductID: productID, SKU: "SKU-001", Quantity: 10}}

	err := service.ReserveForOrder(context.Background(), orderID, items, userID)
	require.NoError(t, err)
	assert.Equal(t, 10, inv.GetQuantityReserved())
	assert.Equal(t, 90, inv.GetQuantityAvailable())
}

// Test ReserveForOrder - Success with multiple items
func TestReservationService_ReserveForOrder_Success_MultipleItems(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()

	inv1 := createTestInventory(uuidv7.New(), "SKU-001", 100, 0)
	inv2 := createTestInventory(uuidv7.New(), "SKU-002", 50, 0)

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			if sku == "SKU-001" {
				return inv1, nil
			}
			if sku == "SKU-002" {
				return inv2, nil
			}
			return nil, fmt.Errorf("inventory not found")
		},
		UpdateInventoryFunc: func(ctx context.Context, i *aggregate.Inventory) error {
			return nil
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{
		{SKU: "SKU-001", Quantity: 10},
		{SKU: "SKU-002", Quantity: 5},
	}

	err := service.ReserveForOrder(context.Background(), orderID, items, userID)
	require.NoError(t, err)
	assert.Equal(t, 10, inv1.GetQuantityReserved())
	assert.Equal(t, 5, inv2.GetQuantityReserved())
}

// Test ReserveForOrder - Insufficient stock error
func TestReservationService_ReserveForOrder_InsufficientStock(t *testing.T) {
	productID := uuidv7.New()
	orderID := uuidv7.New()
	userID := uuidv7.New()

	inv := createTestInventory(productID, "SKU-001", 5, 0)

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return inv, nil
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{{ProductID: productID, SKU: "SKU-001", Quantity: 10}}

	err := service.ReserveForOrder(context.Background(), orderID, items, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")
}

// Test ReserveForOrder - Inventory not found
func TestReservationService_ReserveForOrder_InventoryNotFound(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return nil, fmt.Errorf("inventory not found")
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{{SKU: "SKU-MISSING", Quantity: 10}}

	err := service.ReserveForOrder(context.Background(), orderID, items, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inventory not found")
}

// Test ReserveForOrder - Missing orderID
func TestReservationService_ReserveForOrder_MissingOrderID(t *testing.T) {
	userID := uuidv7.New()
	mockUC := &MockInventoryUseCase{}
	service := NewReservationService(mockUC)
	items := []OrderItem{{SKU: "SKU-001", Quantity: 10}}

	err := service.ReserveForOrder(context.Background(), uuidv7.Nil, items, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "order ID is required")
}

// Test ReserveForOrder - Empty items
func TestReservationService_ReserveForOrder_EmptyItems(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()
	mockUC := &MockInventoryUseCase{}
	service := NewReservationService(mockUC)

	err := service.ReserveForOrder(context.Background(), orderID, []OrderItem{}, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no items to reserve")
}

// Test ReserveForOrder - Rollback on partial failure
func TestReservationService_ReserveForOrder_RollbackOnError(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()

	inv1 := createTestInventory(uuidv7.New(), "SKU-001", 100, 0)
	inv2 := createTestInventory(uuidv7.New(), "SKU-002", 5, 0)

	updateCallCount := 0

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			if sku == "SKU-001" {
				return inv1, nil
			}
			if sku == "SKU-002" {
				return inv2, nil
			}
			return nil, fmt.Errorf("inventory not found")
		},
		UpdateInventoryFunc: func(ctx context.Context, i *aggregate.Inventory) error {
			updateCallCount++
			return nil
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{
		{SKU: "SKU-001", Quantity: 10},
		{SKU: "SKU-002", Quantity: 10}, // Insufficient stock - should trigger rollback
	}

	err := service.ReserveForOrder(context.Background(), orderID, items, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")
	// Verify rollback was called (1 for SKU-001 reserve + 1 for rollback)
	assert.Equal(t, 2, updateCallCount)
}

// Test ReleaseForOrder - Success
func TestReservationService_ReleaseForOrder_Success(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()

	inv := createTestInventory(uuidv7.New(), "SKU-001", 100, 10)

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return inv, nil
		},
		UpdateInventoryFunc: func(ctx context.Context, i *aggregate.Inventory) error {
			return nil
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{{SKU: "SKU-001", Quantity: 10}}

	err := service.ReleaseForOrder(context.Background(), orderID, items, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, inv.GetQuantityReserved())
	assert.Equal(t, 100, inv.GetQuantityAvailable())
}

// Test ReleaseForOrder - Idempotent (no error if inventory not found)
func TestReservationService_ReleaseForOrder_Idempotent(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return nil, fmt.Errorf("inventory not found")
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{{SKU: "SKU-MISSING", Quantity: 10}}

	err := service.ReleaseForOrder(context.Background(), orderID, items, userID)
	require.NoError(t, err) // Should not fail even if inventory not found
}

// Test CommitForOrder - Success
func TestReservationService_CommitForOrder_Success(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()

	inv := createTestInventory(uuidv7.New(), "SKU-001", 100, 10)

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return inv, nil
		},
		UpdateInventoryFunc: func(ctx context.Context, i *aggregate.Inventory) error {
			return nil
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{{SKU: "SKU-001", Quantity: 10}}

	err := service.CommitForOrder(context.Background(), orderID, items, userID)
	require.NoError(t, err)
	assert.Equal(t, 0, inv.GetQuantityReserved())
	assert.Equal(t, 90, inv.GetQuantityOnHand())
}

// Test CommitForOrder - Error when no reservation exists
func TestReservationService_CommitForOrder_NoReservation(t *testing.T) {
	orderID := uuidv7.New()
	userID := uuidv7.New()

	inv := createTestInventory(uuidv7.New(), "SKU-001", 100, 0) // No reservation

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return inv, nil
		},
	}

	service := NewReservationService(mockUC)
	items := []OrderItem{{SKU: "SKU-001", Quantity: 10}}

	err := service.CommitForOrder(context.Background(), orderID, items, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient reserved stock")
}

// Test CommitForOrder - Missing orderID
func TestReservationService_CommitForOrder_MissingOrderID(t *testing.T) {
	userID := uuidv7.New()
	mockUC := &MockInventoryUseCase{}
	service := NewReservationService(mockUC)
	items := []OrderItem{{SKU: "SKU-001", Quantity: 10}}

	err := service.CommitForOrder(context.Background(), uuidv7.Nil, items, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "order ID is required")
}
