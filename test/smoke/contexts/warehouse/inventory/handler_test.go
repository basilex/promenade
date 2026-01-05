package inventory_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	inventoryHTTP "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockInventoryUseCase implements IUseCase interface for smoke testing
type MockInventoryUseCase struct {
	CreateInventoryFunc  func(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*inventory.Inventory, error)
	GetInventoryFunc     func(ctx context.Context, id uuidv7.UUID) (*inventory.Inventory, error)
	GetBySKUFunc         func(ctx context.Context, sku string) (*inventory.Inventory, error)
	GetByProductIDFunc   func(ctx context.Context, productID uuidv7.UUID) ([]*inventory.Inventory, error)
	GetByWarehouseFunc   func(ctx context.Context, warehouseID string) ([]*inventory.Inventory, error)
	GetLowStockFunc      func(ctx context.Context) ([]*inventory.Inventory, error)
	ListInventoryFunc    func(ctx context.Context, page, pageSize int) ([]*inventory.Inventory, int64, error)
	UpdateInventoryFunc  func(ctx context.Context, inv *inventory.Inventory) error
	DeleteInventoryFunc  func(ctx context.Context, id uuidv7.UUID) error
	ReceiveStockFunc     func(ctx context.Context, id uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*inventory.Inventory, error)
	CommitStockFunc      func(ctx context.Context, id uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*inventory.Inventory, error)
}

// Implement IUseCase methods with nil checks

func (m *MockInventoryUseCase) CreateInventory(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*inventory.Inventory, error) {
	if m.CreateInventoryFunc != nil {
		return m.CreateInventoryFunc(ctx, productID, sku, productName, warehouseID, createdBy)
	}
	return nil, fmt.Errorf("CreateInventoryFunc not implemented")
}

func (m *MockInventoryUseCase) GetInventory(ctx context.Context, id uuidv7.UUID) (*inventory.Inventory, error) {
	if m.GetInventoryFunc != nil {
		return m.GetInventoryFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetInventoryFunc not implemented")
}

func (m *MockInventoryUseCase) GetBySKU(ctx context.Context, sku string) (*inventory.Inventory, error) {
	if m.GetBySKUFunc != nil {
		return m.GetBySKUFunc(ctx, sku)
	}
	return nil, fmt.Errorf("GetBySKUFunc not implemented")
}

func (m *MockInventoryUseCase) GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*inventory.Inventory, error) {
	if m.GetByProductIDFunc != nil {
		return m.GetByProductIDFunc(ctx, productID)
	}
	return nil, fmt.Errorf("GetByProductIDFunc not implemented")
}

func (m *MockInventoryUseCase) GetByWarehouse(ctx context.Context, warehouseID string) ([]*inventory.Inventory, error) {
	if m.GetByWarehouseFunc != nil {
		return m.GetByWarehouseFunc(ctx, warehouseID)
	}
	return nil, fmt.Errorf("GetByWarehouseFunc not implemented")
}

func (m *MockInventoryUseCase) GetLowStock(ctx context.Context) ([]*inventory.Inventory, error) {
	if m.GetLowStockFunc != nil {
		return m.GetLowStockFunc(ctx)
	}
	return nil, fmt.Errorf("GetLowStockFunc not implemented")
}

func (m *MockInventoryUseCase) ListInventory(ctx context.Context, page, pageSize int) ([]*inventory.Inventory, int64, error) {
	if m.ListInventoryFunc != nil {
		return m.ListInventoryFunc(ctx, page, pageSize)
	}
	return nil, 0, fmt.Errorf("ListInventoryFunc not implemented")
}

func (m *MockInventoryUseCase) UpdateInventory(ctx context.Context, inv *inventory.Inventory) error {
	if m.UpdateInventoryFunc != nil {
		return m.UpdateInventoryFunc(ctx, inv)
	}
	return fmt.Errorf("UpdateInventoryFunc not implemented")
}

func (m *MockInventoryUseCase) DeleteInventory(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteInventoryFunc != nil {
		return m.DeleteInventoryFunc(ctx, id)
	}
	return fmt.Errorf("DeleteInventoryFunc not implemented")
}

func (m *MockInventoryUseCase) ReceiveStock(ctx context.Context, id uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*inventory.Inventory, error) {
	if m.ReceiveStockFunc != nil {
		return m.ReceiveStockFunc(ctx, id, quantity, unitCostCents, receivedBy)
	}
	return nil, fmt.Errorf("ReceiveStockFunc not implemented")
}

func (m *MockInventoryUseCase) CommitStock(ctx context.Context, id uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*inventory.Inventory, error) {
	if m.CommitStockFunc != nil {
		return m.CommitStockFunc(ctx, id, quantity, committedBy)
	}
	return nil, fmt.Errorf("CommitStockFunc not implemented")
}

// fakeInventory creates a fake inventory item for testing
func fakeInventory() *inventory.Inventory {
	productID := uuidv7.New()
	inv, _ := inventory.NewInventory(productID, "TEST-SKU-001", "Test Product", "WH-MAIN", uuidv7.New())
	userID := uuidv7.New()
	_ = inv.ReceiveStock(100, 1000, userID)
	inv.LocationCode = "SHELF-A1"
	inv.LocationZone = "Zone-A"
	return inv
}

// ============================================================================
// Smoke Tests
// ============================================================================

func TestInventoryHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		CreateInventoryFunc: func(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*inventory.Inventory, error) {
			return fakeInventory(), nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.POST("/inventory", handler.Create)

	body := map[string]any{
		"product_id":   smoke.FakeUUID(),
		"sku":          "TEST-SKU-001",
		"product_name": "Test Product",
		"warehouse_id": "WH-MAIN",
		"created_by":   smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/inventory", body)

	smoke.AssertSuccessResponse(t, w, http.StatusCreated)
}

func TestInventoryHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{}
	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.POST("/inventory", handler.Create)

	body := map[string]any{
		"sku": "TEST-SKU-001",
		// Missing required fields
	}
	w := smoke.MakeRequest(t, router, "POST", "/inventory", body)

	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestInventoryHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		GetInventoryFunc: func(ctx context.Context, id uuidv7.UUID) (*inventory.Inventory, error) {
			return fakeInventory(), nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.GET("/inventory/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/inventory/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		GetInventoryFunc: func(ctx context.Context, id uuidv7.UUID) (*inventory.Inventory, error) {
			return nil, inventory.ErrInventoryNotFound
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.GET("/inventory/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/inventory/"+smoke.FakeUUID(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInventoryHandler_GetBySKU_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*inventory.Inventory, error) {
			return fakeInventory(), nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.GET("/inventory/sku/:sku", handler.GetBySKU)

	w := smoke.MakeRequest(t, router, "GET", "/inventory/sku/TEST-SKU-001", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_GetBySKU_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		GetBySKUFunc: func(ctx context.Context, sku string) (*inventory.Inventory, error) {
			return nil, inventory.ErrInventoryNotFound
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.GET("/inventory/sku/:sku", handler.GetBySKU)

	w := smoke.MakeRequest(t, router, "GET", "/inventory/sku/NONEXISTENT", nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInventoryHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		ListInventoryFunc: func(ctx context.Context, page, pageSize int) ([]*inventory.Inventory, int64, error) {
			return []*inventory.Inventory{fakeInventory()}, 1, nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.GET("/inventory", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/inventory?page=1&page_size=20", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		ListInventoryFunc: func(ctx context.Context, page, pageSize int) ([]*inventory.Inventory, int64, error) {
			return []*inventory.Inventory{}, 0, nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.GET("/inventory", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/inventory", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		GetInventoryFunc: func(ctx context.Context, id uuidv7.UUID) (*inventory.Inventory, error) {
			return fakeInventory(), nil
		},
		UpdateInventoryFunc: func(ctx context.Context, inv *inventory.Inventory) error {
			return nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.PUT("/inventory/:id", handler.Update)

	body := map[string]any{
		"reorder_point":    30,
		"reorder_quantity": 100,
	}
	w := smoke.MakeRequest(t, router, "PUT", "/inventory/"+smoke.FakeUUID(), body)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_Update_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		GetInventoryFunc: func(ctx context.Context, id uuidv7.UUID) (*inventory.Inventory, error) {
			return nil, inventory.ErrInventoryNotFound
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.PUT("/inventory/:id", handler.Update)

	body := map[string]any{
		"reorder_point": 30,
	}
	w := smoke.MakeRequest(t, router, "PUT", "/inventory/"+smoke.FakeUUID(), body)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInventoryHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		DeleteInventoryFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.DELETE("/inventory/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/inventory/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		DeleteInventoryFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return inventory.ErrInventoryNotFound
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.DELETE("/inventory/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/inventory/"+smoke.FakeUUID(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestInventoryHandler_GetLowStock_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		GetLowStockFunc: func(ctx context.Context) ([]*inventory.Inventory, error) {
			return []*inventory.Inventory{fakeInventory()}, nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.GET("/inventory/low-stock", handler.GetLowStock)

	w := smoke.MakeRequest(t, router, "GET", "/inventory/low-stock", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_ReceiveStock_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		ReceiveStockFunc: func(ctx context.Context, id uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*inventory.Inventory, error) {
			return fakeInventory(), nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.POST("/inventory/:id/receive", handler.ReceiveStock)

	body := map[string]any{
		"quantity":        50,
		"unit_cost_cents": 1000,
		"received_by":     smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/inventory/"+smoke.FakeUUID()+"/receive", body)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestInventoryHandler_CommitStock_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInventoryUseCase{
		CommitStockFunc: func(ctx context.Context, id uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*inventory.Inventory, error) {
			return fakeInventory(), nil
		},
	}

	handler := inventoryHTTP.NewInventoryHandler(mockUC)
	router.POST("/inventory/:id/commit", handler.CommitStock)

	body := map[string]any{
		"quantity":     25,
		"committed_by": smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/inventory/"+smoke.FakeUUID()+"/commit", body)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

