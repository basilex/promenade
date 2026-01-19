package stockmovement_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	handler "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/adapter/http"
	stockmovementAggregate "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockStockMovementUseCase mocks IUseCase for smoke tests
type MockStockMovementUseCase struct {
	RecordMovementFunc          func(ctx context.Context, inventoryID uuidv7.UUID, movementType stockmovementAggregate.MovementType, quantity, quantityBefore int, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error)
	RecordReceiptFunc           func(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, unitCostCents int64, currencyCode string, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error)
	RecordReservationFunc       func(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error)
	RecordCommitFunc            func(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error)
	RecordAdjustmentFunc        func(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, reason string, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error)
	RecordTransferFunc          func(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, fromWarehouse, toWarehouse uuidv7.UUID, fromLocation, toLocation *string, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error)
	GetMovementFunc             func(ctx context.Context, id uuidv7.UUID) (*stockmovementAggregate.StockMovement, error)
	GetMovementsByInventoryFunc func(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*stockmovementAggregate.StockMovement, int, error)
	GetMovementsByReferenceFunc func(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*stockmovementAggregate.StockMovement, error)
	GetMovementsByTypeFunc      func(ctx context.Context, movementType stockmovementAggregate.MovementType, startDate, endDate time.Time, page, pageSize int) ([]*stockmovementAggregate.StockMovement, int, error)
	GetRecentMovementsFunc      func(ctx context.Context, limit int) ([]*stockmovementAggregate.StockMovement, error)
	GetInventorySummaryFunc     func(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (int, int, error)
}

func (m *MockStockMovementUseCase) RecordMovement(ctx context.Context, inventoryID uuidv7.UUID, movementType stockmovementAggregate.MovementType, quantity, quantityBefore int, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
	if m.RecordMovementFunc != nil {
		return m.RecordMovementFunc(ctx, inventoryID, movementType, quantity, quantityBefore, createdBy)
	}
	return nil, errors.New("RecordMovementFunc not implemented")
}

func (m *MockStockMovementUseCase) RecordReceipt(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, unitCostCents int64, currencyCode string, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
	if m.RecordReceiptFunc != nil {
		return m.RecordReceiptFunc(ctx, inventoryID, quantity, quantityBefore, unitCostCents, currencyCode, createdBy)
	}
	return nil, errors.New("RecordReceiptFunc not implemented")
}

func (m *MockStockMovementUseCase) RecordReservation(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
	if m.RecordReservationFunc != nil {
		return m.RecordReservationFunc(ctx, inventoryID, quantity, quantityBefore, orderID, createdBy)
	}
	return nil, errors.New("RecordReservationFunc not implemented")
}

func (m *MockStockMovementUseCase) RecordCommit(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
	if m.RecordCommitFunc != nil {
		return m.RecordCommitFunc(ctx, inventoryID, quantity, quantityBefore, orderID, createdBy)
	}
	return nil, errors.New("RecordCommitFunc not implemented")
}

func (m *MockStockMovementUseCase) RecordAdjustment(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, reason string, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
	if m.RecordAdjustmentFunc != nil {
		return m.RecordAdjustmentFunc(ctx, inventoryID, quantity, quantityBefore, reason, createdBy)
	}
	return nil, errors.New("RecordAdjustmentFunc not implemented")
}

func (m *MockStockMovementUseCase) RecordTransfer(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, fromWarehouse, toWarehouse uuidv7.UUID, fromLocation, toLocation *string, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
	if m.RecordTransferFunc != nil {
		return m.RecordTransferFunc(ctx, inventoryID, quantity, quantityBefore, fromWarehouse, toWarehouse, fromLocation, toLocation, createdBy)
	}
	return nil, errors.New("RecordTransferFunc not implemented")
}

func (m *MockStockMovementUseCase) GetMovement(ctx context.Context, id uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
	if m.GetMovementFunc != nil {
		return m.GetMovementFunc(ctx, id)
	}
	return nil, errors.New("GetMovementFunc not implemented")
}

func (m *MockStockMovementUseCase) GetMovementsByInventory(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*stockmovementAggregate.StockMovement, int, error) {
	if m.GetMovementsByInventoryFunc != nil {
		return m.GetMovementsByInventoryFunc(ctx, inventoryID, page, pageSize)
	}
	return nil, 0, errors.New("GetMovementsByInventoryFunc not implemented")
}

func (m *MockStockMovementUseCase) GetMovementsByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*stockmovementAggregate.StockMovement, error) {
	if m.GetMovementsByReferenceFunc != nil {
		return m.GetMovementsByReferenceFunc(ctx, referenceType, referenceID)
	}
	return nil, errors.New("GetMovementsByReferenceFunc not implemented")
}

func (m *MockStockMovementUseCase) GetMovementsByType(ctx context.Context, movementType stockmovementAggregate.MovementType, startDate, endDate time.Time, page, pageSize int) ([]*stockmovementAggregate.StockMovement, int, error) {
	if m.GetMovementsByTypeFunc != nil {
		return m.GetMovementsByTypeFunc(ctx, movementType, startDate, endDate, page, pageSize)
	}
	return nil, 0, errors.New("GetMovementsByTypeFunc not implemented")
}

func (m *MockStockMovementUseCase) GetRecentMovements(ctx context.Context, limit int) ([]*stockmovementAggregate.StockMovement, error) {
	if m.GetRecentMovementsFunc != nil {
		return m.GetRecentMovementsFunc(ctx, limit)
	}
	return nil, errors.New("GetRecentMovementsFunc not implemented")
}

func (m *MockStockMovementUseCase) GetInventorySummary(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (int, int, error) {
	if m.GetInventorySummaryFunc != nil {
		return m.GetInventorySummaryFunc(ctx, inventoryID, startDate, endDate)
	}
	return 0, 0, errors.New("GetInventorySummaryFunc not implemented")
}

// Helper: fake movement
func fakeMovement() *stockmovementAggregate.StockMovement {
	movement, _ := stockmovementAggregate.NewStockMovement(
		uuidv7.New(),
		stockmovementAggregate.MovementTypeReceipt,
		10,
		100,
		uuidv7.New(),
	)
	return movement
}

// TestStockMovementHandler_RecordMovement_Success tests recording movement with success
func TestStockMovementHandler_RecordMovement_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		RecordMovementFunc: func(ctx context.Context, inventoryID uuidv7.UUID, movementType stockmovementAggregate.MovementType, quantity, quantityBefore int, createdBy uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
			return fakeMovement(), nil
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.POST("/stock-movements", h.RecordMovement)

	body := map[string]interface{}{
		"inventory_id":    smoke.FakeUUID(),
		"type":            "receipt",
		"quantity":        10,
		"quantity_before": 100,
		"created_by":      smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, http.MethodPost, "/stock-movements", body)
	smoke.AssertSuccessResponse(t, w, http.StatusCreated)
}

// TestStockMovementHandler_RecordMovement_ValidationError tests validation error
func TestStockMovementHandler_RecordMovement_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{}
	h := handler.NewStockMovementHandler(mockUC)
	router.POST("/stock-movements", h.RecordMovement)

	body := map[string]interface{}{
		"inventory_id": "invalid-uuid",
	}

	w := smoke.MakeRequest(t, router, http.MethodPost, "/stock-movements", body)
	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

// TestStockMovementHandler_GetByID_Success tests retrieving movement by ID
func TestStockMovementHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		GetMovementFunc: func(ctx context.Context, id uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
			return fakeMovement(), nil
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.GET("/stock-movements/:id", h.GetByID)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/stock-movements/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

// TestStockMovementHandler_GetByID_NotFound tests movement not found
func TestStockMovementHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		GetMovementFunc: func(ctx context.Context, id uuidv7.UUID) (*stockmovementAggregate.StockMovement, error) {
			return nil, stockmovement.ErrStockMovementNotFound
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.GET("/stock-movements/:id", h.GetByID)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/stock-movements/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

// TestStockMovementHandler_GetByInventory_Success tests retrieving movements by inventory
func TestStockMovementHandler_GetByInventory_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		GetMovementsByInventoryFunc: func(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*stockmovementAggregate.StockMovement, int, error) {
			return []*stockmovementAggregate.StockMovement{fakeMovement()}, 1, nil
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.GET("/stock-movements/inventory/:inventory_id", h.GetByInventory)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/stock-movements/inventory/"+smoke.FakeUUID()+"?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

// TestStockMovementHandler_GetByReference_Success tests retrieving movements by reference
func TestStockMovementHandler_GetByReference_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		GetMovementsByReferenceFunc: func(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*stockmovementAggregate.StockMovement, error) {
			return []*stockmovementAggregate.StockMovement{fakeMovement()}, nil
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.GET("/stock-movements/reference", h.GetByReference)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/stock-movements/reference?type=order&id="+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

// TestStockMovementHandler_GetByType_Success tests retrieving movements by type
func TestStockMovementHandler_GetByType_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		GetMovementsByTypeFunc: func(ctx context.Context, movementType stockmovementAggregate.MovementType, startDate, endDate time.Time, page, pageSize int) ([]*stockmovementAggregate.StockMovement, int, error) {
			return []*stockmovementAggregate.StockMovement{fakeMovement()}, 1, nil
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.GET("/stock-movements/type/:type", h.GetByType)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/stock-movements/type/receipt?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

// TestStockMovementHandler_GetInventorySummary_Success tests getting inventory summary
func TestStockMovementHandler_GetInventorySummary_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		GetInventorySummaryFunc: func(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (int, int, error) {
			return 100, 50, nil
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.GET("/stock-movements/inventory/:inventory_id/summary", h.GetInventorySummary)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/stock-movements/inventory/"+smoke.FakeUUID()+"/summary", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

// TestStockMovementHandler_GetRecent_Success tests getting recent movements
func TestStockMovementHandler_GetRecent_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockStockMovementUseCase{
		GetRecentMovementsFunc: func(ctx context.Context, limit int) ([]*stockmovementAggregate.StockMovement, error) {
			return []*stockmovementAggregate.StockMovement{fakeMovement()}, nil
		},
	}

	h := handler.NewStockMovementHandler(mockUC)
	router.GET("/stock-movements/recent", h.GetRecent)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/stock-movements/recent?limit=10", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}
