package stockmovement_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestInventory creates a test inventory record (required for FK constraint)
func createTestInventoryForUC(t *testing.T, ctx context.Context, db *integration.TestDB) *inventory.Inventory {
	invRepo := inventoryRepo.NewInventoryRepository(db.DB)
	invUC := inventory.NewUseCase(invRepo)
	inv, err := invUC.CreateInventory(ctx, uuidv7.New(), "SM-UC-"+uuidv7.New().String()[:8], "Test Product", "WH-MAIN", uuidv7.New())
	require.NoError(t, err)
	return inv
}

// TestStockMovementUseCase_RecordReceipt tests recording receipt through UseCase
func TestStockMovementUseCase_RecordReceipt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		createdBy := uuidv7.New()

		// Record receipt
		movement, err := uc.RecordReceipt(ctx, inv.ID, 50, 100, 1500, "USD", createdBy)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, movement.ID)
		assert.Equal(t, stockmovement.MovementTypeReceipt, movement.Type)
		assert.Equal(t, 50, movement.Quantity)
		assert.Equal(t, 100, movement.QuantityBeforeMove)
		assert.Equal(t, 150, movement.QuantityAfterMove)
		assert.NotNil(t, movement.UnitCostCents)
		assert.Equal(t, int64(1500), *movement.UnitCostCents)

		// Verify persistence
		retrieved, err := uc.GetMovement(ctx, movement.ID)
		require.NoError(t, err)
		assert.Equal(t, movement.ID, retrieved.ID)
	})
}

// TestStockMovementUseCase_RecordReservation tests recording reservation through UseCase
func TestStockMovementUseCase_RecordReservation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		orderID := uuidv7.New()
		createdBy := uuidv7.New()

		// Record reservation
		movement, err := uc.RecordReservation(ctx, inv.ID, 30, 150, orderID, createdBy)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, movement.ID)
		assert.Equal(t, stockmovement.MovementTypeReservation, movement.Type)
		assert.Equal(t, -30, movement.Quantity)
		assert.Equal(t, 150, movement.QuantityBeforeMove)
		assert.Equal(t, 120, movement.QuantityAfterMove)
		assert.Equal(t, "order", movement.ReferenceType)
		assert.Equal(t, orderID, *movement.ReferenceID)
	})
}

// TestStockMovementUseCase_RecordCommit tests recording commit through UseCase
func TestStockMovementUseCase_RecordCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		orderID := uuidv7.New()
		createdBy := uuidv7.New()

		// Record commit
		movement, err := uc.RecordCommit(ctx, inv.ID, 25, 120, orderID, createdBy)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, movement.ID)
		assert.Equal(t, stockmovement.MovementTypeCommit, movement.Type)
		assert.Equal(t, -25, movement.Quantity)
		assert.Equal(t, 120, movement.QuantityBeforeMove)
		assert.Equal(t, 95, movement.QuantityAfterMove)
		assert.Equal(t, "order", movement.ReferenceType)
	})
}

// TestStockMovementUseCase_RecordAdjustment tests recording adjustment through UseCase
func TestStockMovementUseCase_RecordAdjustment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		createdBy := uuidv7.New()

		// Record adjustment
		movement, err := uc.RecordAdjustment(ctx, inv.ID, 10, 100, "Stock count correction", createdBy)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, movement.ID)
		assert.Equal(t, stockmovement.MovementTypeAdjustment, movement.Type)
		assert.Equal(t, 10, movement.Quantity)
		assert.Equal(t, "Stock count correction", movement.Reason)
	})
}

// TestStockMovementUseCase_RecordTransfer tests recording transfer through UseCase
func TestStockMovementUseCase_RecordTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		fromWarehouse := uuidv7.New()
		toWarehouse := uuidv7.New()
		createdBy := uuidv7.New()
		fromLocation := "A-01"
		toLocation := "B-02"

		// Record transfer
		movement, err := uc.RecordTransfer(ctx, inv.ID, 20, 100, fromWarehouse, toWarehouse, &fromLocation, &toLocation, createdBy)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, movement.ID)
		assert.Equal(t, stockmovement.MovementTypeTransfer, movement.Type)
		assert.Equal(t, 20, movement.Quantity)
		assert.Equal(t, fromWarehouse, *movement.FromWarehouseID)
		assert.Equal(t, toWarehouse, *movement.ToWarehouseID)
		assert.Equal(t, fromLocation, *movement.FromLocationCode)
		assert.Equal(t, toLocation, *movement.ToLocationCode)
	})
}

// TestStockMovementUseCase_GetMovementsByInventory tests listing movements through UseCase
func TestStockMovementUseCase_GetMovementsByInventory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		createdBy := uuidv7.New()

		// Create multiple movements
		_, err := uc.RecordReceipt(ctx, inv.ID, 50, 0, 1000, "USD", createdBy)
		require.NoError(t, err)

		orderID := uuidv7.New()
		_, err = uc.RecordReservation(ctx, inv.ID, 10, 50, orderID, createdBy)
		require.NoError(t, err)

		_, err = uc.RecordCommit(ctx, inv.ID, 10, 40, orderID, createdBy)
		require.NoError(t, err)

		// Get movements by inventory
		movements, total, err := uc.GetMovementsByInventory(ctx, inv.ID, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, movements, 3)
	})
}

// TestStockMovementUseCase_GetMovementsByType tests filtering by type through UseCase
func TestStockMovementUseCase_GetMovementsByType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		createdBy := uuidv7.New()

		// Create movements of different types
		_, err := uc.RecordReceipt(ctx, inv.ID, 50, 0, 1000, "USD", createdBy)
		require.NoError(t, err)

		_, err = uc.RecordReceipt(ctx, inv.ID, 30, 50, 1000, "USD", createdBy)
		require.NoError(t, err)

		orderID := uuidv7.New()
		_, err = uc.RecordReservation(ctx, inv.ID, 20, 80, orderID, createdBy)
		require.NoError(t, err)

		// Get receipts only
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now().Add(24 * time.Hour)
		movements, total, err := uc.GetMovementsByType(ctx, stockmovement.MovementTypeReceipt, startDate, endDate, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 2)
		assert.GreaterOrEqual(t, len(movements), 2)
		for _, m := range movements {
			assert.Equal(t, stockmovement.MovementTypeReceipt, m.Type)
		}
	})
}

// TestStockMovementUseCase_GetInventorySummary tests summary calculation through UseCase
func TestStockMovementUseCase_GetInventorySummary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewStockMovementRepository(testDB.DB)
		uc := stockmovement.NewUseCase(repo)

		// Create inventory first (FK requirement)
		inv := createTestInventoryForUC(t, ctx, testDB)
		createdBy := uuidv7.New()

		// Create inbound movements
		_, err := uc.RecordReceipt(ctx, inv.ID, 100, 0, 1000, "USD", createdBy)
		require.NoError(t, err)

		// Create outbound movements
		orderID := uuidv7.New()
		_, err = uc.RecordReservation(ctx, inv.ID, 30, 100, orderID, createdBy)
		require.NoError(t, err)

		_, err = uc.RecordCommit(ctx, inv.ID, 20, 70, orderID, createdBy)
		require.NoError(t, err)

		// Get summary
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now().Add(24 * time.Hour)
		totalIn, totalOut, err := uc.GetInventorySummary(ctx, inv.ID, startDate, endDate)
		require.NoError(t, err)
		assert.Equal(t, 100, totalIn)
		assert.Equal(t, 50, totalOut) // 30 + 20
	})
}
