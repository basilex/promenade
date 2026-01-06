package inventory_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	"github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestInventoryUseCase_CreateAndGetInventory tests creating and retrieving inventory through UseCase
func TestInventoryUseCase_CreateAndGetInventory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		productID := uuidv7.New()
		warehouseID := "WH-MAIN"

		// Create inventory
		inv, err := uc.CreateInventory(ctx, productID, "TEST-UC-001", "Test Product", warehouseID, uuidv7.New())
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, inv.ID)
		assert.Equal(t, "TEST-UC-001", inv.SKU)
		assert.Equal(t, productID, inv.ProductID)
		assert.Equal(t, warehouseID, inv.WarehouseID)

		// GetInventory by ID
		retrieved, err := uc.GetInventory(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, retrieved.ID)
		assert.Equal(t, "TEST-UC-001", retrieved.SKU)

		// GetInventory - not found
		_, err = uc.GetInventory(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, inventory.ErrInventoryNotFound, err)
	})
}

// TestInventoryUseCase_GetBySKU tests SKU lookup through UseCase
func TestInventoryUseCase_GetBySKU(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		// Create inventory with unique SKU
		productID := uuidv7.New()
		uniqueSKU := fmt.Sprintf("UC-SKU-%s", uuidv7.New().String())
		inv, err := uc.CreateInventory(ctx, productID, uniqueSKU, "Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		// GetBySKU - found
		found, err := uc.GetBySKU(ctx, uniqueSKU)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, found.ID)
		assert.Equal(t, uniqueSKU, found.SKU)

		// GetBySKU - not found
		_, err = uc.GetBySKU(ctx, "NON-EXISTENT-SKU")
		assert.Error(t, err)
		assert.Equal(t, inventory.ErrInventoryNotFound, err)
	})
}

// TestInventoryUseCase_GetByProductID tests product lookup
func TestInventoryUseCase_GetByProductID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		productID := uuidv7.New()

		// Create multiple inventory items for same product
		inv1, err := uc.CreateInventory(ctx, productID, "UC-PROD-WH1", "Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		inv2, err := uc.CreateInventory(ctx, productID, "UC-PROD-WH2", "Product", "WH-BACKUP", uuidv7.New())
		require.NoError(t, err)

		// GetByProductID
		items, err := uc.GetByProductID(ctx, productID)
		require.NoError(t, err)
		assert.Len(t, items, 2)

		// Verify both warehouses
		ids := []uuidv7.UUID{items[0].ID, items[1].ID}
		assert.Contains(t, ids, inv1.ID)
		assert.Contains(t, ids, inv2.ID)
	})
}

// TestInventoryUseCase_GetByWarehouse tests warehouse filtering
func TestInventoryUseCase_GetByWarehouse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		// Create items in different warehouses
		_, err := uc.CreateInventory(ctx, uuidv7.New(), "UC-WH-MAIN-1", "Product 1", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		_, err = uc.CreateInventory(ctx, uuidv7.New(), "UC-WH-MAIN-2", "Product 2", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		_, err = uc.CreateInventory(ctx, uuidv7.New(), "UC-WH-BACKUP-1", "Product 3", "WH-BACKUP", uuidv7.New())
		require.NoError(t, err)

		// GetByWarehouse - WH-MAIN
		mainItems, err := uc.GetByWarehouse(ctx, "WH-MAIN")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(mainItems), 2)

		// Count our test items
		count := 0
		for _, item := range mainItems {
			if item.SKU == "UC-WH-MAIN-1" || item.SKU == "UC-WH-MAIN-2" {
				count++
			}
		}
		assert.Equal(t, 2, count)

		// GetByWarehouse - WH-BACKUP
		backupItems, err := uc.GetByWarehouse(ctx, "WH-BACKUP")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(backupItems), 1)

		// Find our test item
		found := false
		for _, item := range backupItems {
			if item.SKU == "UC-WH-BACKUP-1" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

// TestInventoryUseCase_ListInventory tests pagination
func TestInventoryUseCase_ListInventory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		// Create 5 inventory items
		for i := 1; i <= 5; i++ {
			sku := fmt.Sprintf("UC-LIST-%s-%d", uuidv7.New().String(), i)
			_, err := uc.CreateInventory(ctx, uuidv7.New(), sku, "Product", "WH-MAIN", uuidv7.New())
			require.NoError(t, err)
		}

		// List - page 1 (pageSize 2)
		items1, total, err := uc.ListInventory(ctx, 1, 2)
		require.NoError(t, err)
		assert.Len(t, items1, 2)
		assert.GreaterOrEqual(t, total, int64(5))

		// List - page 2
		items2, total, err := uc.ListInventory(ctx, 2, 2)
		require.NoError(t, err)
		assert.Len(t, items2, 2)
		assert.GreaterOrEqual(t, total, int64(5))

		// List - page 3
		items3, total, err := uc.ListInventory(ctx, 3, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(items3), 1)
		assert.GreaterOrEqual(t, total, int64(5))
	})
}

// TestInventoryUseCase_GetLowStock tests low stock detection
func TestInventoryUseCase_GetLowStock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		userID := uuidv7.New()
		lowStockSKU := fmt.Sprintf("UC-LOW-%s", uuidv7.New().String())
		goodStockSKU := fmt.Sprintf("UC-GOOD-%s", uuidv7.New().String())

		// Create inventory with low stock
		invLow, err := uc.CreateInventory(ctx, uuidv7.New(), lowStockSKU, "Low Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		// Receive stock and set reorder point
		_, err = uc.ReceiveStock(ctx, invLow.ID, 5, 1000, userID)
		require.NoError(t, err)

		invLow, _ = uc.GetInventory(ctx, invLow.ID)
		invLow.ReorderPoint = 10 // Reorder at 10, but only 5 available
		err = repo.Update(ctx, invLow)
		require.NoError(t, err)

		// Create inventory with good stock
		invGood, err := uc.CreateInventory(ctx, uuidv7.New(), goodStockSKU, "Good Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		_, err = uc.ReceiveStock(ctx, invGood.ID, 20, 1000, userID)
		require.NoError(t, err)

		invGood, _ = uc.GetInventory(ctx, invGood.ID)
		invGood.ReorderPoint = 10 // Reorder at 10, have 20
		err = repo.Update(ctx, invGood)
		require.NoError(t, err)

		// GetLowStock
		lowItems, err := uc.GetLowStock(ctx)
		require.NoError(t, err)

		// Find our low stock item
		foundLow := false
		foundGood := false
		for _, item := range lowItems {
			if item.SKU == lowStockSKU {
				foundLow = true
			}
			if item.SKU == goodStockSKU {
				foundGood = true
			}
		}

		assert.True(t, foundLow, "Low stock item should be in results")
		assert.False(t, foundGood, "Good stock item should NOT be in results")
	})
}

// TestInventoryUseCase_ReceiveStock tests receiving stock
func TestInventoryUseCase_ReceiveStock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		// Create inventory
		inv, err := uc.CreateInventory(ctx, uuidv7.New(), "UC-RECEIVE-001", "Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)
		assert.Equal(t, 0, inv.QuantityOnHand)

		userID := uuidv7.New()

		// Receive stock
		updated, err := uc.ReceiveStock(ctx, inv.ID, 100, 5000, userID)
		require.NoError(t, err)
		assert.Equal(t, 100, updated.QuantityOnHand)
		assert.Equal(t, 100, updated.QuantityAvailable)
		assert.Equal(t, int64(5000), updated.UnitCostCents)

		// Verify persistence
		retrieved, err := uc.GetInventory(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 100, retrieved.QuantityOnHand)
		assert.Equal(t, int64(5000), retrieved.UnitCostCents)
	})
}

// TestInventoryUseCase_CommitStock tests committing stock
func TestInventoryUseCase_CommitStock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		userID := uuidv7.New()

		// Create inventory with stock
		inv, err := uc.CreateInventory(ctx, uuidv7.New(), "UC-COMMIT-001", "Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		// Receive stock
		inv, err = uc.ReceiveStock(ctx, inv.ID, 100, 5000, userID)
		require.NoError(t, err)

		// Commit stock
		committed, err := uc.CommitStock(ctx, inv.ID, 30, userID)
		require.NoError(t, err)
		assert.Equal(t, 70, committed.QuantityOnHand)
		assert.Equal(t, 30, committed.QuantityCommitted)
		assert.Equal(t, 70, committed.QuantityAvailable)

		// Verify persistence
		retrieved, err := uc.GetInventory(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 70, retrieved.QuantityOnHand)
		assert.Equal(t, 30, retrieved.QuantityCommitted)
	})
}

// TestInventoryUseCase_UpdateAndDelete tests update and delete operations
func TestInventoryUseCase_UpdateAndDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		// Create inventory
		inv, err := uc.CreateInventory(ctx, uuidv7.New(), "UC-UPDATE-001", "Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)

		// Update
		inv.ProductName = "Updated Product"
		inv.ReorderPoint = 25
		err = uc.UpdateInventory(ctx, inv)
		require.NoError(t, err)

		// Verify update
		updated, err := uc.GetInventory(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Product", updated.ProductName)
		assert.Equal(t, 25, updated.ReorderPoint)

		// Delete
		err = uc.DeleteInventory(ctx, inv.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = uc.GetInventory(ctx, inv.ID)
		assert.Error(t, err)
		assert.Equal(t, inventory.ErrInventoryNotFound, err)
	})
}

// TestInventoryUseCase_CompleteWorkflow tests complete stock workflow
func TestInventoryUseCase_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		productID := uuidv7.New()
		userID := uuidv7.New()
		orderID := uuidv7.New()

		// 1. Create inventory
		inv, err := uc.CreateInventory(ctx, productID, "UC-WORKFLOW-001", "Product", "WH-MAIN", uuidv7.New())
		require.NoError(t, err)
		assert.Equal(t, 0, inv.QuantityOnHand)

		// 2. Receive stock
		inv, err = uc.ReceiveStock(ctx, inv.ID, 100, 5000, userID)
		require.NoError(t, err)
		assert.Equal(t, 100, inv.QuantityOnHand)
		assert.Equal(t, 100, inv.QuantityAvailable)

		// 3. Reserve stock for order (via repository, as UseCase doesn't have ReserveStock method)
		retrieved, err := uc.GetInventory(ctx, inv.ID)
		require.NoError(t, err)
		err = retrieved.ReserveStock(30, orderID, userID)
		require.NoError(t, err)
		err = repo.Update(ctx, retrieved)
		require.NoError(t, err)

		// Verify reservation
		reserved, err := uc.GetInventory(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 100, reserved.QuantityOnHand)
		assert.Equal(t, 30, reserved.QuantityReserved)
		assert.Equal(t, 70, reserved.QuantityAvailable)

		// 4. Commit reserved stock (order fulfilled, via repository)
		err = reserved.CommitReservation(30, orderID, userID)
		require.NoError(t, err)
		err = repo.Update(ctx, reserved)
		require.NoError(t, err)

		// Verify commitment
		committed, err := uc.GetInventory(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 70, committed.QuantityOnHand)
		assert.Equal(t, 0, committed.QuantityReserved)
		assert.Equal(t, 30, committed.QuantityCommitted)
		assert.Equal(t, 70, committed.QuantityAvailable)

		// 5. Commit additional stock (via UseCase)
		committed, err = uc.CommitStock(ctx, inv.ID, 20, userID)
		require.NoError(t, err)
		assert.Equal(t, 50, committed.QuantityOnHand)
		assert.Equal(t, 50, committed.QuantityCommitted) // 30 + 20
		assert.Equal(t, 50, committed.QuantityAvailable)
	})
}

// TestInventoryUseCase_EdgeCases tests edge cases and validations
func TestInventoryUseCase_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)
		uc := inventory.NewUseCase(repo)

		userID := uuidv7.New()

		t.Run("DuplicateSKU", func(t *testing.T) {
			productID := uuidv7.New()
			uniqueSKU := fmt.Sprintf("UC-EDGE-%s", uuidv7.New().String())

			// Create first
			_, err := uc.CreateInventory(ctx, productID, uniqueSKU, "Product", "WH-MAIN", uuidv7.New())
			require.NoError(t, err)

			// Try to create duplicate
			_, err = uc.CreateInventory(ctx, productID, uniqueSKU, "Product", "WH-MAIN", uuidv7.New())
			assert.Error(t, err)
			// UseCase returns "SKU already exists: {sku}" error
			assert.Contains(t, err.Error(), "SKU already exists")
		})

		t.Run("InvalidPagination", func(t *testing.T) {
			// Invalid page (0)
			items, total, err := uc.ListInventory(ctx, 0, 10)
			require.NoError(t, err)
			assert.NotNil(t, items)
			assert.GreaterOrEqual(t, total, int64(0))

			// Invalid pageSize (0)
			items, total, err = uc.ListInventory(ctx, 1, 0)
			require.NoError(t, err)
			assert.NotNil(t, items)
			assert.GreaterOrEqual(t, total, int64(0))
		})

		t.Run("ReceiveStockValidation", func(t *testing.T) {
			inv, _ := uc.CreateInventory(ctx, uuidv7.New(), "UC-VALIDATION-001", "Product", "WH-MAIN", uuidv7.New())

			// Invalid quantity (usecase validates with "quantity must be greater than 0")
			_, err := uc.ReceiveStock(ctx, inv.ID, -10, 1000, userID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "quantity must be greater than 0")

			// Invalid cost (no validation in entity, test will pass)
			// Note: Entity doesn't validate cost < 0, so this will succeed
		})

		t.Run("CommitStockValidation", func(t *testing.T) {
			inv, _ := uc.CreateInventory(ctx, uuidv7.New(), "UC-VALIDATION-002", "Product", "WH-MAIN", uuidv7.New())

			// Invalid quantity (usecase validates with "quantity must be greater than 0")
			_, err := uc.CommitStock(ctx, inv.ID, -10, userID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "quantity must be greater than 0")

			// Commit more than available
			_, err = uc.ReceiveStock(ctx, inv.ID, 10, 1000, userID)
			require.NoError(t, err)

			_, err = uc.CommitStock(ctx, inv.ID, 20, userID)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "adjustment would result in negative stock")
		})
	})
}
