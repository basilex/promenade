package stockmovement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// mockRepository implements IRepository interface with function fields for testing
type mockRepository struct {
	CreateFunc                 func(ctx context.Context, movement *StockMovement) error
	GetByIDFunc                func(ctx context.Context, id uuidv7.UUID) (*StockMovement, error)
	GetByInventoryIDFunc       func(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*StockMovement, int, error)
	GetByReferenceFunc         func(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*StockMovement, error)
	GetByTypeFunc              func(ctx context.Context, movementType MovementType, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error)
	GetByDateRangeFunc         func(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error)
	GetSummaryByInventoryFunc  func(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (int, int, error)
	GetRecentMovementsFunc     func(ctx context.Context, limit int) ([]*StockMovement, error)
	CountByTypeFunc            func(ctx context.Context, startDate, endDate time.Time) (map[MovementType]int, error)
}

func (m *mockRepository) Create(ctx context.Context, movement *StockMovement) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, movement)
	}
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*StockMovement, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, ErrStockMovementNotFound
}

func (m *mockRepository) GetByInventoryID(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*StockMovement, int, error) {
	if m.GetByInventoryIDFunc != nil {
		return m.GetByInventoryIDFunc(ctx, inventoryID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockRepository) GetByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*StockMovement, error) {
	if m.GetByReferenceFunc != nil {
		return m.GetByReferenceFunc(ctx, referenceType, referenceID)
	}
	return nil, nil
}

func (m *mockRepository) GetByType(ctx context.Context, movementType MovementType, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error) {
	if m.GetByTypeFunc != nil {
		return m.GetByTypeFunc(ctx, movementType, startDate, endDate, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error) {
	if m.GetByDateRangeFunc != nil {
		return m.GetByDateRangeFunc(ctx, startDate, endDate, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockRepository) GetSummaryByInventory(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (int, int, error) {
	if m.GetSummaryByInventoryFunc != nil {
		return m.GetSummaryByInventoryFunc(ctx, inventoryID, startDate, endDate)
	}
	return 0, 0, nil
}

func (m *mockRepository) GetRecentMovements(ctx context.Context, limit int) ([]*StockMovement, error) {
	if m.GetRecentMovementsFunc != nil {
		return m.GetRecentMovementsFunc(ctx, limit)
	}
	return nil, nil
}

func (m *mockRepository) CountByType(ctx context.Context, startDate, endDate time.Time) (map[MovementType]int, error) {
	if m.CountByTypeFunc != nil {
		return m.CountByTypeFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

// TestNewUseCase tests the constructor
func TestNewUseCase(t *testing.T) {
	repo := &mockRepository{}
	uc := NewUseCase(repo)
	assert.NotNil(t, uc)
}

// TestUseCase_RecordMovement tests recording a generic movement
func TestUseCase_RecordMovement(t *testing.T) {
	repo := &mockRepository{
		CreateFunc: func(ctx context.Context, movement *StockMovement) error {
			return nil
		},
	}
	uc := NewUseCase(repo)
	ctx := context.Background()

	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()

	movement, err := uc.RecordMovement(ctx, inventoryID, MovementTypeReceipt, 10, 100, createdBy)
	require.NoError(t, err)
	assert.NotNil(t, movement)
	assert.Equal(t, inventoryID, movement.InventoryID)
	assert.Equal(t, MovementTypeReceipt, movement.Type)
	assert.Equal(t, 10, movement.Quantity)
	assert.Equal(t, 100, movement.QuantityBeforeMove)
	assert.Equal(t, 110, movement.QuantityAfterMove)
}

// TestUseCase_RecordReceipt tests recording a receipt movement
func TestUseCase_RecordReceipt(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			CreateFunc: func(ctx context.Context, movement *StockMovement) error {
				return nil
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()

		movement, err := uc.RecordReceipt(ctx, inventoryID, 50, 100, 1000, "USD", createdBy)
		require.NoError(t, err)
		assert.NotNil(t, movement)
		assert.Equal(t, 50, movement.Quantity)
		assert.Equal(t, 100, movement.QuantityBeforeMove)
		assert.Equal(t, 150, movement.QuantityAfterMove)
		assert.NotNil(t, movement.UnitCostCents)
		assert.Equal(t, int64(1000), *movement.UnitCostCents)
	})

	t.Run("zero quantity", func(t *testing.T) {
		repo := &mockRepository{}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()

		_, err := uc.RecordReceipt(ctx, inventoryID, 0, 100, 1000, "USD", createdBy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity cannot be zero")
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &mockRepository{
			CreateFunc: func(ctx context.Context, movement *StockMovement) error {
				return errors.New("database error")
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()

		_, err := uc.RecordReceipt(ctx, inventoryID, 50, 100, 1000, "USD", createdBy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
	})
}

// TestUseCase_RecordReservation tests recording a reservation movement
func TestUseCase_RecordReservation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			CreateFunc: func(ctx context.Context, movement *StockMovement) error {
				return nil
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()
		orderID := uuidv7.New()

		movement, err := uc.RecordReservation(ctx, inventoryID, 20, 100, orderID, createdBy)
		require.NoError(t, err)
		assert.NotNil(t, movement)
		assert.Equal(t, -20, movement.Quantity)
		assert.Equal(t, MovementTypeReservation, movement.Type)
		assert.Equal(t, "order", movement.ReferenceType)
		assert.Equal(t, orderID, *movement.ReferenceID)
	})

	t.Run("zero quantity", func(t *testing.T) {
		repo := &mockRepository{}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()
		orderID := uuidv7.New()

		_, err := uc.RecordReservation(ctx, inventoryID, 0, 100, orderID, createdBy)
		assert.Error(t, err)
	})
}

// TestUseCase_RecordCommit tests recording a commit movement
func TestUseCase_RecordCommit(t *testing.T) {
	repo := &mockRepository{
		CreateFunc: func(ctx context.Context, movement *StockMovement) error {
			return nil
		},
	}
	uc := NewUseCase(repo)
	ctx := context.Background()

	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()
	orderID := uuidv7.New()

	movement, err := uc.RecordCommit(ctx, inventoryID, 15, 100, orderID, createdBy)
	require.NoError(t, err)
	assert.NotNil(t, movement)
	assert.Equal(t, -15, movement.Quantity)
	assert.Equal(t, MovementTypeCommit, movement.Type)
	assert.Equal(t, "order", movement.ReferenceType)
}

// TestUseCase_RecordAdjustment tests recording an adjustment movement
func TestUseCase_RecordAdjustment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			CreateFunc: func(ctx context.Context, movement *StockMovement) error {
				return nil
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()

		movement, err := uc.RecordAdjustment(ctx, inventoryID, 5, 100, "Inventory count correction", createdBy)
		require.NoError(t, err)
		assert.NotNil(t, movement)
		assert.Equal(t, MovementTypeAdjustment, movement.Type)
		assert.Equal(t, "Inventory count correction", movement.Reason)
	})

	t.Run("missing reason", func(t *testing.T) {
		repo := &mockRepository{}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()

		_, err := uc.RecordAdjustment(ctx, inventoryID, 5, 100, "", createdBy)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reason cannot be empty")
	})
}

// TestUseCase_RecordTransfer tests recording a transfer movement
func TestUseCase_RecordTransfer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			CreateFunc: func(ctx context.Context, movement *StockMovement) error {
				return nil
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		inventoryID := uuidv7.New()
		createdBy := uuidv7.New()
		fromWarehouse := uuidv7.New()
		toWarehouse := uuidv7.New()
		fromLocation := "A-01"
		toLocation := "B-02"

		movement, err := uc.RecordTransfer(ctx, inventoryID, 25, 100, fromWarehouse, toWarehouse, &fromLocation, &toLocation, createdBy)
		require.NoError(t, err)
		assert.NotNil(t, movement)
		assert.Equal(t, MovementTypeTransfer, movement.Type)
		assert.Equal(t, fromWarehouse, *movement.FromWarehouseID)
		assert.Equal(t, toWarehouse, *movement.ToWarehouseID)
	})
}

// TestUseCase_GetMovement tests retrieving a movement by ID
func TestUseCase_GetMovement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		movementID := uuidv7.New()
		repo := &mockRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*StockMovement, error) {
				movement, _ := NewStockMovement(uuidv7.New(), MovementTypeReceipt, 10, 100, uuidv7.New())
				return movement, nil
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		movement, err := uc.GetMovement(ctx, movementID)
		require.NoError(t, err)
		assert.NotNil(t, movement)
	})

	t.Run("not found", func(t *testing.T) {
		movementID := uuidv7.New()
		repo := &mockRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*StockMovement, error) {
				return nil, ErrStockMovementNotFound
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		_, err := uc.GetMovement(ctx, movementID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrStockMovementNotFound)
	})
}

// TestUseCase_GetMovementsByInventory tests retrieving movements by inventory ID
func TestUseCase_GetMovementsByInventory(t *testing.T) {
	inventoryID := uuidv7.New()
	repo := &mockRepository{
		GetByInventoryIDFunc: func(ctx context.Context, invID uuidv7.UUID, page, pageSize int) ([]*StockMovement, int, error) {
			m1, _ := NewStockMovement(invID, MovementTypeReceipt, 10, 100, uuidv7.New())
			m2, _ := NewStockMovement(invID, MovementTypeReservation, -5, 110, uuidv7.New())
			movements := []*StockMovement{m1, m2}
			return movements, 2, nil
		},
	}
	uc := NewUseCase(repo)
	ctx := context.Background()

	movements, total, err := uc.GetMovementsByInventory(ctx, inventoryID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, movements, 2)
	assert.Equal(t, 2, total)
}

// TestUseCase_GetMovementsByType tests retrieving movements by type
func TestUseCase_GetMovementsByType(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			GetByTypeFunc: func(ctx context.Context, movementType MovementType, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error) {
				movement, _ := NewStockMovement(uuidv7.New(), movementType, 10, 100, uuidv7.New())
				movements := []*StockMovement{movement}
				return movements, 1, nil
			},
		}
		uc := NewUseCase(repo)
		ctx := context.Background()

		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()

		movements, total, err := uc.GetMovementsByType(ctx, MovementTypeReceipt, startDate, endDate, 1, 10)
		require.NoError(t, err)
		assert.Len(t, movements, 1)
		assert.Equal(t, 1, total)
	})

	t.Run("invalid date range", func(t *testing.T) {
		repo := &mockRepository{}
		uc := NewUseCase(repo)
		ctx := context.Background()

		startDate := time.Now()
		endDate := time.Now().Add(-24 * time.Hour)

		_, _, err := uc.GetMovementsByType(ctx, MovementTypeReceipt, startDate, endDate, 1, 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start date must be before end date")
	})
}

// TestUseCase_GetInventorySummary tests getting inventory summary
func TestUseCase_GetInventorySummary(t *testing.T) {
	inventoryID := uuidv7.New()
	repo := &mockRepository{
		GetSummaryByInventoryFunc: func(ctx context.Context, invID uuidv7.UUID, startDate, endDate time.Time) (int, int, error) {
			return 100, 50, nil
		},
	}
	uc := NewUseCase(repo)
	ctx := context.Background()

	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now()

	totalIn, totalOut, err := uc.GetInventorySummary(ctx, inventoryID, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 100, totalIn)
	assert.Equal(t, 50, totalOut)
}
