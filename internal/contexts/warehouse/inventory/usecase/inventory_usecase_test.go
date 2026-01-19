package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inventoryerrors "github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	"github.com/basilex/promenade/internal/contexts/warehouse/inventory/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Test constants
var testUserID = uuidv7.New()

// MockRepository implements IRepository for testing
type MockRepository struct {
	CreateFunc         func(ctx context.Context, inv *aggregate.Inventory) error
	GetByIDFunc        func(ctx context.Context, id uuidv7.UUID) (*aggregate.Inventory, error)
	GetBySKUFunc       func(ctx context.Context, sku string) (*aggregate.Inventory, error)
	GetByProductIDFunc func(ctx context.Context, productID uuidv7.UUID) ([]*aggregate.Inventory, error)
	GetByWarehouseFunc func(ctx context.Context, warehouseID string) ([]*aggregate.Inventory, error)
	GetLowStockFunc    func(ctx context.Context) ([]*aggregate.Inventory, error)
	ListInventoryFunc  func(ctx context.Context, limit, offset int) ([]*aggregate.Inventory, int64, error)
	UpdateFunc         func(ctx context.Context, inventory *aggregate.Inventory) error
	DeleteFunc         func(ctx context.Context, id uuidv7.UUID) error
	GetByLocationFunc  func(ctx context.Context, warehouseID, locationCode string) ([]*aggregate.Inventory, error)
	GetByStatusFunc    func(ctx context.Context, status aggregate.InventoryStatus) ([]*aggregate.Inventory, error)
	BulkUpdateFunc     func(ctx context.Context, items []*aggregate.Inventory) error
}

func (m *MockRepository) Create(ctx context.Context, inv *aggregate.Inventory) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, inv)
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Inventory, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRepository) GetBySKU(ctx context.Context, sku string) (*aggregate.Inventory, error) {
	if m.GetBySKUFunc != nil {
		return m.GetBySKUFunc(ctx, sku)
	}
	return nil, nil
}

func (m *MockRepository) GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*aggregate.Inventory, error) {
	if m.GetByProductIDFunc != nil {
		return m.GetByProductIDFunc(ctx, productID)
	}
	return nil, nil
}

func (m *MockRepository) GetByWarehouse(ctx context.Context, warehouseID string) ([]*aggregate.Inventory, error) {
	if m.GetByWarehouseFunc != nil {
		return m.GetByWarehouseFunc(ctx, warehouseID)
	}
	return nil, nil
}

func (m *MockRepository) GetLowStock(ctx context.Context) ([]*aggregate.Inventory, error) {
	if m.GetLowStockFunc != nil {
		return m.GetLowStockFunc(ctx)
	}
	return nil, nil
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*aggregate.Inventory, int64, error) {
	if m.ListInventoryFunc != nil {
		return m.ListInventoryFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockRepository) Update(ctx context.Context, inv *aggregate.Inventory) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, inv)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRepository) BulkUpdate(ctx context.Context, items []*aggregate.Inventory) error {
	if m.BulkUpdateFunc != nil {
		return m.BulkUpdateFunc(ctx, items)
	}
	return nil
}

func (m *MockRepository) GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*aggregate.Inventory, error) {
	if m.GetByLocationFunc != nil {
		return m.GetByLocationFunc(ctx, warehouseID, locationCode)
	}
	return nil, nil
}

func (m *MockRepository) GetByStatus(ctx context.Context, status aggregate.InventoryStatus) ([]*aggregate.Inventory, error) {
	if m.GetByStatusFunc != nil {
		return m.GetByStatusFunc(ctx, status)
	}
	return nil, nil
}

// TestNewUseCase tests the constructor
func TestNewUseCase(t *testing.T) {
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	assert.NotNil(t, uc)
}

// TestCreateInventory_Success tests successful inventory creation
func TestCreateInventory_Success(t *testing.T) {
	productID := uuidv7.New()
	repo := &MockRepository{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return nil, errors.New("not found") // SKU doesn't exist
		},
		CreateFunc: func(ctx context.Context, inv *aggregate.Inventory) error {
			assert.Equal(t, "TEST-SKU-001", inv.SKU)
			assert.Equal(t, "Test Product", inv.ProductName)
			return nil
		},
	}

	uc := NewInventoryUseCase(repo)
	inv, err := uc.CreateInventory(context.Background(), productID, "TEST-SKU-001", "Test Product", "WH-MAIN", testUserID)

	require.NoError(t, err)
	assert.NotNil(t, inv)
	assert.Equal(t, "TEST-SKU-001", inv.SKU)
	assert.Equal(t, "Test Product", inv.ProductName)
	assert.Equal(t, "WH-MAIN", inv.WarehouseID)
}

// TestCreateInventory_DuplicateSKU tests duplicate SKU validation
func TestCreateInventory_DuplicateSKU(t *testing.T) {
	productID := uuidv7.New()
	existingInv, _ := aggregate.NewInventory(productID, "TEST-SKU-001", "Existing", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return existingInv, nil // SKU exists
		},
	}

	uc := NewInventoryUseCase(repo)
	inv, err := uc.CreateInventory(context.Background(), productID, "TEST-SKU-001", "New Product", "WH-MAIN", testUserID)

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventorySKUExists))
}

// TestCreateInventory_EmptySKU tests SKU validation
func TestCreateInventory_EmptySKU(t *testing.T) {
	productID := uuidv7.New()
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	inv, err := uc.CreateInventory(context.Background(), productID, "", "Product", "WH-MAIN", testUserID)

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventorySKURequired))
}

// TestCreateInventory_EmptyProductName tests product name validation
func TestCreateInventory_EmptyProductName(t *testing.T) {
	productID := uuidv7.New()
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	inv, err := uc.CreateInventory(context.Background(), productID, "TEST-SKU", "", "WH-MAIN", testUserID)

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventoryProductNameRequired))
}

// TestCreateInventory_EmptyWarehouseID tests warehouse ID validation
func TestCreateInventory_EmptyWarehouseID(t *testing.T) {
	productID := uuidv7.New()
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	inv, err := uc.CreateInventory(context.Background(), productID, "TEST-SKU", "Product", "", testUserID)

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventoryWarehouseRequired))
}

// TestGetInventory_Success tests successful retrieval
func TestGetInventory_Success(t *testing.T) {
	id := uuidv7.New()
	productID := uuidv7.New()
	expected, _ := aggregate.NewInventory(productID, "TEST-SKU", "Product", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			assert.Equal(t, id, invID)
			return expected, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	inv, err := uc.GetInventory(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, expected, inv)
}

// TestGetInventory_NotFound tests not found error
func TestGetInventory_NotFound(t *testing.T) {
	id := uuidv7.New()
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			return nil, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	inv, err := uc.GetInventory(context.Background(), id)

	assert.ErrorIs(t, err, inventoryerrors.ErrInventoryNotFound)
	assert.Nil(t, inv)
}

// TestGetBySKU_Success tests successful retrieval by SKU
func TestGetBySKU_Success(t *testing.T) {
	productID := uuidv7.New()
	expected, _ := aggregate.NewInventory(productID, "TEST-SKU", "Product", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			assert.Equal(t, "TEST-SKU", sku)
			return expected, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	inv, err := uc.GetBySKU(context.Background(), "TEST-SKU")

	require.NoError(t, err)
	assert.Equal(t, expected, inv)
}

// TestGetBySKU_EmptySKU tests empty SKU validation
func TestGetBySKU_EmptySKU(t *testing.T) {
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	inv, err := uc.GetBySKU(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventorySKURequired))
}

// TestGetBySKU_NotFound tests not found error
func TestGetBySKU_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Inventory, error) {
			return nil, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	inv, err := uc.GetBySKU(context.Background(), "NONEXISTENT")

	assert.ErrorIs(t, err, inventoryerrors.ErrInventoryNotFound)
	assert.Nil(t, inv)
}

// TestGetByProductID tests retrieval by product ID
func TestGetByProductID(t *testing.T) {
	productID := uuidv7.New()
	inv1, _ := aggregate.NewInventory(productID, "SKU-1", "Product", "WH-1", testUserID)
	inv2, _ := aggregate.NewInventory(productID, "SKU-2", "Product", "WH-2", testUserID)

	repo := &MockRepository{
		GetByProductIDFunc: func(ctx context.Context, pid uuidv7.UUID) ([]*aggregate.Inventory, error) {
			assert.Equal(t, productID, pid)
			return []*aggregate.Inventory{inv1, inv2}, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	items, err := uc.GetByProductID(context.Background(), productID)

	require.NoError(t, err)
	assert.Len(t, items, 2)
}

// TestGetByWarehouse tests retrieval by warehouse
func TestGetByWarehouse(t *testing.T) {
	productID := uuidv7.New()
	inv1, _ := aggregate.NewInventory(productID, "SKU-1", "Product 1", "WH-MAIN", testUserID)
	inv2, _ := aggregate.NewInventory(productID, "SKU-2", "Product 2", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetByWarehouseFunc: func(ctx context.Context, warehouseID string) ([]*aggregate.Inventory, error) {
			assert.Equal(t, "WH-MAIN", warehouseID)
			return []*aggregate.Inventory{inv1, inv2}, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	items, err := uc.GetByWarehouse(context.Background(), "WH-MAIN")

	require.NoError(t, err)
	assert.Len(t, items, 2)
}

// TestGetByWarehouse_EmptyWarehouseID tests warehouse ID validation
func TestGetByWarehouse_EmptyWarehouseID(t *testing.T) {
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	items, err := uc.GetByWarehouse(context.Background(), "")

	assert.Error(t, err)
	assert.Nil(t, items)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventoryWarehouseRequired))
}

// TestGetLowStock tests retrieval of low stock items
func TestGetLowStock(t *testing.T) {
	productID := uuidv7.New()
	inv1, _ := aggregate.NewInventory(productID, "SKU-1", "Product 1", "WH-MAIN", testUserID)
	inv1.ReorderPoint = 100
	inv1.QuantityOnHand = 50 // Below reorder point

	repo := &MockRepository{
		GetLowStockFunc: func(ctx context.Context) ([]*aggregate.Inventory, error) {
			return []*aggregate.Inventory{inv1}, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	items, err := uc.GetLowStock(context.Background())

	require.NoError(t, err)
	assert.Len(t, items, 1)
}

// TestListInventory tests paginated listing
func TestListInventory(t *testing.T) {
	productID := uuidv7.New()
	inv1, _ := aggregate.NewInventory(productID, "SKU-1", "Product 1", "WH-MAIN", testUserID)
	inv2, _ := aggregate.NewInventory(productID, "SKU-2", "Product 2", "WH-MAIN", testUserID)

	repo := &MockRepository{
		ListInventoryFunc: func(ctx context.Context, limit, offset int) ([]*aggregate.Inventory, int64, error) {
			assert.Equal(t, 20, limit) // pageSize becomes limit
			assert.Equal(t, 0, offset) // (page-1)*pageSize = (1-1)*20 = 0
			return []*aggregate.Inventory{inv1, inv2}, 2, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	items, total, err := uc.ListInventory(context.Background(), 1, 20)

	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(2), total)
}

// TestListInventory_InvalidPage tests page normalization
func TestListInventory_InvalidPage(t *testing.T) {
	repo := &MockRepository{
		ListInventoryFunc: func(ctx context.Context, limit, offset int) ([]*aggregate.Inventory, int64, error) {
			assert.Equal(t, 20, limit) // Default pageSize
			assert.Equal(t, 0, offset) // Normalized to page 1
			return nil, 0, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	_, _, err := uc.ListInventory(context.Background(), 0, 20)

	require.NoError(t, err)
}

// TestListInventory_InvalidPageSize tests page size normalization
func TestListInventory_InvalidPageSize(t *testing.T) {
	repo := &MockRepository{
		ListInventoryFunc: func(ctx context.Context, limit, offset int) ([]*aggregate.Inventory, int64, error) {
			assert.Equal(t, 20, limit) // Should normalize to 20
			assert.Equal(t, 0, offset) // (1-1)*20 = 0
			return nil, 0, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	_, _, err := uc.ListInventory(context.Background(), 1, 200)

	require.NoError(t, err)
}

// TestUpdateInventory_Success tests successful update
func TestUpdateInventory_Success(t *testing.T) {
	productID := uuidv7.New()
	inv, _ := aggregate.NewInventory(productID, "TEST-SKU", "Product", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Inventory, error) {
			return inv, nil
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Inventory) error {
			assert.Equal(t, inv.GetID(), updated.GetID())
			return nil
		},
	}

	uc := NewInventoryUseCase(repo)
	err := uc.UpdateInventory(context.Background(), inv)

	require.NoError(t, err)
}

// TestUpdateInventory_NilInventory tests nil inventory validation
func TestUpdateInventory_NilInventory(t *testing.T) {
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	err := uc.UpdateInventory(context.Background(), nil)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventoryNil))
}

// TestUpdateInventory_NotFound tests not found error
func TestUpdateInventory_NotFound(t *testing.T) {
	productID := uuidv7.New()
	inv, _ := aggregate.NewInventory(productID, "TEST-SKU", "Product", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Inventory, error) {
			return nil, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	err := uc.UpdateInventory(context.Background(), inv)

	assert.ErrorIs(t, err, inventoryerrors.ErrInventoryNotFound)
}

// TestDeleteInventory_Success tests successful deletion
func TestDeleteInventory_Success(t *testing.T) {
	id := uuidv7.New()
	productID := uuidv7.New()
	inv, _ := aggregate.NewInventory(productID, "TEST-SKU", "Product", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			assert.Equal(t, id, invID)
			return inv, nil
		},
		DeleteFunc: func(ctx context.Context, invID uuidv7.UUID) error {
			assert.Equal(t, id, invID)
			return nil
		},
	}

	uc := NewInventoryUseCase(repo)
	err := uc.DeleteInventory(context.Background(), id)

	require.NoError(t, err)
}

// TestDeleteInventory_NotFound tests not found error
func TestDeleteInventory_NotFound(t *testing.T) {
	id := uuidv7.New()
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			return nil, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	err := uc.DeleteInventory(context.Background(), id)

	assert.ErrorIs(t, err, inventoryerrors.ErrInventoryNotFound)
}

// TestUseCase_ReceiveStock_Success tests successful stock receipt
func TestUseCase_ReceiveStock_Success(t *testing.T) {
	id := uuidv7.New()
	productID := uuidv7.New()
	userID := uuidv7.New()
	inv, _ := aggregate.NewInventory(productID, "TEST-SKU", "Product", "WH-MAIN", testUserID)

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			return inv, nil
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Inventory) error {
			assert.Equal(t, 100, updated.QuantityOnHand)
			return nil
		},
	}

	uc := NewInventoryUseCase(repo)
	result, err := uc.ReceiveStock(context.Background(), id, 100, 1000, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100, result.QuantityOnHand)
}

// TestUseCase_ReceiveStock_InvalidQuantity tests quantity validation
func TestUseCase_ReceiveStock_InvalidQuantity(t *testing.T) {
	id := uuidv7.New()
	userID := uuidv7.New()
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	result, err := uc.ReceiveStock(context.Background(), id, 0, 1000, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventoryQuantityInvalid))
}

// TestUseCase_ReceiveStock_NegativeUnitCost tests unit cost validation
func TestUseCase_ReceiveStock_NegativeUnitCost(t *testing.T) {
	id := uuidv7.New()
	userID := uuidv7.New()
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	result, err := uc.ReceiveStock(context.Background(), id, 100, -100, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventoryUnitCostNegative))
}

// TestUseCase_ReceiveStock_NotFound tests not found error
func TestUseCase_ReceiveStock_NotFound(t *testing.T) {
	id := uuidv7.New()
	userID := uuidv7.New()
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			return nil, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	result, err := uc.ReceiveStock(context.Background(), id, 100, 1000, userID)

	assert.ErrorIs(t, err, inventoryerrors.ErrInventoryNotFound)
	assert.Nil(t, result)
}

// TestUseCase_CommitStock_Success tests successful stock commitment
func TestUseCase_CommitStock_Success(t *testing.T) {
	id := uuidv7.New()
	productID := uuidv7.New()
	userID := uuidv7.New()
	receivedBy := uuidv7.New()
	inv, _ := aggregate.NewInventory(productID, "TEST-SKU", "Product", "WH-MAIN", testUserID)
	_ = inv.ReceiveStock(100, 1000, receivedBy) // Add stock first

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			return inv, nil
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Inventory) error {
			assert.Equal(t, 75, updated.QuantityAvailable)
			assert.Equal(t, 25, updated.QuantityCommitted)
			return nil
		},
	}

	uc := NewInventoryUseCase(repo)
	result, err := uc.CommitStock(context.Background(), id, 25, userID)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 75, result.QuantityAvailable)
	assert.Equal(t, 25, result.QuantityCommitted)
}

// TestUseCase_CommitStock_InvalidQuantity tests quantity validation
func TestUseCase_CommitStock_InvalidQuantity(t *testing.T) {
	id := uuidv7.New()
	userID := uuidv7.New()
	repo := &MockRepository{}
	uc := NewInventoryUseCase(repo)

	result, err := uc.CommitStock(context.Background(), id, -5, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, inventoryerrors.ErrInventoryQuantityInvalid))
}

// TestUseCase_CommitStock_NotFound tests not found error
func TestUseCase_CommitStock_NotFound(t *testing.T) {
	id := uuidv7.New()
	userID := uuidv7.New()
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, invID uuidv7.UUID) (*aggregate.Inventory, error) {
			return nil, nil
		},
	}

	uc := NewInventoryUseCase(repo)
	result, err := uc.CommitStock(context.Background(), id, 25, userID)

	assert.ErrorIs(t, err, inventoryerrors.ErrInventoryNotFound)
	assert.Nil(t, result)
}
