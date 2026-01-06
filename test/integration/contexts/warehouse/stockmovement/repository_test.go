package stockmovement_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	stockMovementRepo "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestInventory creates an inventory for testing stock movements
func createTestInventory(t *testing.T, ctx context.Context, db *integration.TestDB) *inventory.Inventory {
	invRepo := inventoryRepo.NewInventoryRepository(db.DB)
	invUC := inventory.NewUseCase(invRepo)
	inv, err := invUC.CreateInventory(ctx, uuidv7.New(), "SM-TEST-"+uuidv7.New().String()[:8], "Test Product", "WH-MAIN", uuidv7.New())
	require.NoError(t, err)
	return inv
}

// TestStockMovementRepository_Create tests creating a stock movement record.
func TestStockMovementRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := stockMovementRepo.NewStockMovementRepository(db.DB)
	ctx := context.Background()

	// Create inventory first (foreign key requirement)
	inv := createTestInventory(t, ctx, db)

	// Create stock movement
	createdBy := uuidv7.New()

	movement, err := stockmovement.NewStockMovement(
		inv.ID,
		stockmovement.MovementTypeReceipt,
		10,
		100,
		createdBy,
	)
	require.NoError(t, err)

	// Create movement
	err = repo.Create(ctx, movement)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, movement.ID)

	// Verify creation
	retrieved, err := repo.GetByID(ctx, movement.ID)
	require.NoError(t, err)
	assert.Equal(t, movement.ID, retrieved.ID)
	assert.Equal(t, movement.InventoryID, retrieved.InventoryID)
	assert.Equal(t, movement.Type, retrieved.Type)
	assert.Equal(t, movement.Quantity, retrieved.Quantity)
	assert.Equal(t, movement.QuantityBeforeMove, retrieved.QuantityBeforeMove)
	assert.Equal(t, movement.QuantityAfterMove, retrieved.QuantityAfterMove)
}

// TestStockMovementRepository_GetByInventoryID tests querying movements by inventory ID.
func TestStockMovementRepository_GetByInventoryID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := stockMovementRepo.NewStockMovementRepository(db.DB)
	ctx := context.Background()

	// Create inventory first
	inv := createTestInventory(t, ctx, db)
	createdBy := uuidv7.New()

	// Create multiple movements
	for i := 0; i < 5; i++ {
		movement, _ := stockmovement.NewStockMovement(
			inv.ID,
			stockmovement.MovementTypeReceipt,
			10,
			100+i*10,
			createdBy,
		)
		err := repo.Create(ctx, movement)
		require.NoError(t, err)
	}

	// Query movements
	movements, total, err := repo.GetByInventoryID(ctx, inv.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, movements, 5)

	// Verify ordering (newest first)
	for i := 0; i < len(movements)-1; i++ {
		assert.True(t, movements[i].MovementDate.After(movements[i+1].MovementDate) ||
			movements[i].MovementDate.Equal(movements[i+1].MovementDate))
	}
}

// TestStockMovementRepository_GetByType tests querying movements by type.
func TestStockMovementRepository_GetByType(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := stockMovementRepo.NewStockMovementRepository(db.DB)
	ctx := context.Background()

	// Create inventory first
	inv := createTestInventory(t, ctx, db)
	createdBy := uuidv7.New()

	// Create movements of different types
	types := []stockmovement.MovementType{
		stockmovement.MovementTypeReceipt,
		stockmovement.MovementTypeReceipt,
		stockmovement.MovementTypeReservation,
		stockmovement.MovementTypeCommit,
	}
	for _, movType := range types {
		movement, _ := stockmovement.NewStockMovement(
			inv.ID,
			movType,
			10,
			100,
			createdBy,
		)
		err := repo.Create(ctx, movement)
		require.NoError(t, err)
	}

	// Query receipts only
	// Use wide date range to include all movements
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)
	movements, total, err := repo.GetByType(ctx, stockmovement.MovementTypeReceipt, startDate, endDate, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, movements, 2)
	for _, m := range movements {
		assert.Equal(t, stockmovement.MovementTypeReceipt, m.Type)
	}
}

// TestStockMovementRepository_GetByReference tests querying movements by reference.
func TestStockMovementRepository_GetByReference(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := stockMovementRepo.NewStockMovementRepository(db.DB)
	ctx := context.Background()

	// Create inventory first
	inv := createTestInventory(t, ctx, db)
	createdBy := uuidv7.New()
	orderID := uuidv7.New()

	// Create movements with order reference
	for i := 0; i < 3; i++ {
		movement, _ := stockmovement.NewStockMovement(
			inv.ID,
			stockmovement.MovementTypeReservation,
			10,
			100-i*10,
			createdBy,
		)
		movement.SetReference("order", &orderID)
		err := repo.Create(ctx, movement)
		require.NoError(t, err)
	}

	// Query by order reference
	movements, err := repo.GetByReference(ctx, "order", orderID)
	require.NoError(t, err)
	assert.Len(t, movements, 3)
	for _, m := range movements {
		assert.Equal(t, "order", m.ReferenceType)
		assert.Equal(t, orderID, *m.ReferenceID)
	}
}

// TestStockMovementRepository_GetSummaryByInventory tests getting movement summary.
func TestStockMovementRepository_GetSummaryByInventory(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := stockMovementRepo.NewStockMovementRepository(db.DB)
	ctx := context.Background()

	// Create inventory first
	inv := createTestInventory(t, ctx, db)
	createdBy := uuidv7.New()

	// Create inbound movements (receipt: +50, return: +20)
	receipt, _ := stockmovement.NewStockMovement(inv.ID, stockmovement.MovementTypeReceipt, 50, 100, createdBy)
	err := repo.Create(ctx, receipt)
	require.NoError(t, err)

	returnMov, _ := stockmovement.NewStockMovement(inv.ID, stockmovement.MovementTypeReturn, 20, 150, createdBy)
	err = repo.Create(ctx, returnMov)
	require.NoError(t, err)

	// Create outbound movements (reservation: -30, commit: -10, damage: -5)
	reservation, _ := stockmovement.NewStockMovement(inv.ID, stockmovement.MovementTypeReservation, -30, 170, createdBy)
	err = repo.Create(ctx, reservation)
	require.NoError(t, err)

	commit, _ := stockmovement.NewStockMovement(inv.ID, stockmovement.MovementTypeCommit, -10, 140, createdBy)
	err = repo.Create(ctx, commit)
	require.NoError(t, err)

	damage, _ := stockmovement.NewStockMovement(inv.ID, stockmovement.MovementTypeDamage, -5, 130, createdBy)
	err = repo.Create(ctx, damage)
	require.NoError(t, err)

	// Get summary
	// Use wide date range to include all movements
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)
	totalIn, totalOut, err := repo.GetSummaryByInventory(ctx, inv.ID, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 70, totalIn)   // 50 + 20
	assert.Equal(t, 45, totalOut)  // 30 + 10 + 5
}

// TestStockMovementRepository_CountByType tests counting movements by type.
func TestStockMovementRepository_CountByType(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := stockMovementRepo.NewStockMovementRepository(db.DB)
	ctx := context.Background()

	// Create inventory first
	inv := createTestInventory(t, ctx, db)
	createdBy := uuidv7.New()

	// Create movements
	movements := []struct {
		movType stockmovement.MovementType
		count   int
	}{
		{stockmovement.MovementTypeReceipt, 3},
		{stockmovement.MovementTypeReservation, 5},
		{stockmovement.MovementTypeCommit, 2},
	}
	for _, m := range movements {
		for i := 0; i < m.count; i++ {
			movement, _ := stockmovement.NewStockMovement(inv.ID, m.movType, 10, 100, createdBy)
			err := repo.Create(ctx, movement)
			require.NoError(t, err)
		}
	}

	// Count by type
	// Use wide date range to include all movements
	startDate := time.Now().Add(-24 * time.Hour)
	endDate := time.Now().Add(24 * time.Hour)
	counts, err := repo.CountByType(ctx, startDate, endDate)
	require.NoError(t, err)
	assert.Equal(t, 3, counts[stockmovement.MovementTypeReceipt])
	assert.Equal(t, 5, counts[stockmovement.MovementTypeReservation])
	assert.Equal(t, 2, counts[stockmovement.MovementTypeCommit])
}

// TestStockMovementRepository_GetByDateRange tests querying movements by date range.
func TestStockMovementRepository_GetByDateRange(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := stockMovementRepo.NewStockMovementRepository(db.DB)
	ctx := context.Background()

	// Create inventory first
	inv := createTestInventory(t, ctx, db)
	createdBy := uuidv7.New()

	// Create movements across different dates
	baseDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ {
		movement, _ := stockmovement.NewStockMovement(
			inv.ID,
			stockmovement.MovementTypeReceipt,
			10,
			100,
			createdBy,
		)
		// Manually set movement date for testing (hack for test)
		movement.MovementDate = baseDate.Add(time.Duration(i) * 24 * time.Hour)
		err := repo.Create(ctx, movement)
		require.NoError(t, err)
	}

	// Query date range (days 2-5)
	startDate := baseDate.Add(2 * 24 * time.Hour)
	endDate := baseDate.Add(5 * 24 * time.Hour)
	movements, total, err := repo.GetByDateRange(ctx, startDate, endDate, 1, 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 4) // At least days 2, 3, 4, 5
	assert.GreaterOrEqual(t, len(movements), 4)
	for _, m := range movements {
		assert.True(t, m.MovementDate.After(startDate) || m.MovementDate.Equal(startDate))
		assert.True(t, m.MovementDate.Before(endDate) || m.MovementDate.Equal(endDate))
	}
}
