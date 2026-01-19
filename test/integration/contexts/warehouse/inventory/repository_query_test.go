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
	inventoryAggregate "github.com/basilex/promenade/internal/contexts/warehouse/inventory/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// Helper: Create test inventory for specific product/warehouse/location
func createTestInventoryForProduct(productID uuidv7.UUID, warehouseID, locationCode string, quantity int) *inventoryAggregate.Inventory {
	// Generate unique SKU with full UUID to guarantee uniqueness
	// Even if same product/warehouse/location, each inventory record is unique
	uniqueID := uuidv7.New()
	sku := fmt.Sprintf("SKU-%s", uniqueID.String())

	userID := uuidv7.New()
	inv, _ := inventoryAggregate.NewInventory(
		productID,
		sku,
		"Test Product",
		warehouseID,
		userID, // Created by
	)
	_ = inv.ReceiveStock(quantity, 1000, userID)
	inv.LocationCode = locationCode
	inv.LocationZone = "Zone-A"
	inv.ReorderPoint = 20
	inv.ReorderQuantity = 50
	return inv
}

// Test 1: GetBySKU
func TestInventoryRepository_GetBySKU(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create test inventory
		inv := createTestInventory()
		err := repo.Create(ctx, inv)
		require.NoError(t, err)

		// Test: Found
		found, err := repo.GetBySKU(ctx, inv.SKU)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, found.ID)
		assert.Equal(t, inv.SKU, found.SKU)

		// Test: Not found
		notFound, err := repo.GetBySKU(ctx, "NON-EXISTENT-SKU")
		assert.Error(t, err)
		assert.Nil(t, notFound)
		assert.ErrorIs(t, err, inventory.ErrInventoryNotFound)
	})
}

// Test 2: GetByProductID (multi-warehouse scenario)
func TestInventoryRepository_GetByProductID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create 3 inventory items for same product in different warehouses
		productID := uuidv7.New()
		inv1 := createTestInventoryForProduct(productID, "WH-NY", "A1", 100)
		inv2 := createTestInventoryForProduct(productID, "WH-LA", "B2", 200)
		inv3 := createTestInventoryForProduct(productID, "WH-CHI", "C3", 150)

		require.NoError(t, repo.Create(ctx, inv1))
		require.NoError(t, repo.Create(ctx, inv2))
		require.NoError(t, repo.Create(ctx, inv3))

		// Test: Get all items for product
		items, err := repo.GetByProductID(ctx, productID)
		require.NoError(t, err)
		assert.Len(t, items, 3)

		// Verify warehouses
		warehouses := make([]string, 3)
		for i, item := range items {
			warehouses[i] = item.WarehouseID
		}
		assert.Contains(t, warehouses, "WH-NY")
		assert.Contains(t, warehouses, "WH-LA")
		assert.Contains(t, warehouses, "WH-CHI")
	})
}

// Test 3: GetByWarehouse
func TestInventoryRepository_GetByWarehouse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create items in different warehouses
		prod1 := uuidv7.New()
		prod2 := uuidv7.New()
		inv1 := createTestInventoryForProduct(prod1, "WH-MAIN", "A1", 100)
		inv2 := createTestInventoryForProduct(prod2, "WH-MAIN", "A2", 200)
		inv3 := createTestInventoryForProduct(prod1, "WH-BACKUP", "B1", 50)

		require.NoError(t, repo.Create(ctx, inv1))
		require.NoError(t, repo.Create(ctx, inv2))
		require.NoError(t, repo.Create(ctx, inv3))

		// Test: Get items from WH-MAIN only
		items, err := repo.GetByWarehouse(ctx, "WH-MAIN")
		require.NoError(t, err)
		assert.Len(t, items, 2)
		for _, item := range items {
			assert.Equal(t, "WH-MAIN", item.WarehouseID)
		}
	})
}

// Test 4: GetByLocation (warehouse + location composite)
func TestInventoryRepository_GetByLocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create items in same warehouse, different locations
		prod1 := uuidv7.New()
		prod2 := uuidv7.New()
		inv1 := createTestInventoryForProduct(prod1, "WH-MAIN", "SHELF-A1", 100)
		inv2 := createTestInventoryForProduct(prod2, "WH-MAIN", "SHELF-A1", 200)
		inv3 := createTestInventoryForProduct(prod1, "WH-MAIN", "SHELF-B2", 50)

		require.NoError(t, repo.Create(ctx, inv1))
		require.NoError(t, repo.Create(ctx, inv2))
		require.NoError(t, repo.Create(ctx, inv3))

		// Test: Get items from specific location
		items, err := repo.GetByLocation(ctx, "WH-MAIN", "SHELF-A1")
		require.NoError(t, err)
		assert.Len(t, items, 2)
		for _, item := range items {
			assert.Equal(t, "WH-MAIN", item.WarehouseID)
			assert.Equal(t, "SHELF-A1", item.LocationCode)
		}
	})
}

// Test 5: List with pagination
func TestInventoryRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create 5 inventory items
		for i := 0; i < 5; i++ {
			inv := createTestInventory()
			require.NoError(t, repo.Create(ctx, inv))
		}

		// Test: First page (limit=2, offset=0)
		items, total, err := repo.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.GreaterOrEqual(t, total, int64(5))

		// Test: Second page (limit=2, offset=2)
		items2, total2, err := repo.List(ctx, 2, 2)
		require.NoError(t, err)
		assert.Len(t, items2, 2)
		assert.Equal(t, total, total2)

		// Verify different items
		assert.NotEqual(t, items[0].ID, items2[0].ID)
	})
}

// Test 6: GetLowStock (CQRS read model for reorder alerts)
func TestInventoryRepository_GetLowStock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create low stock items (available < reorder_point)
		userID := uuidv7.New()

		// Low stock: 10 available, reorder_point=20
		inv1 := createTestInventoryForProduct(uuidv7.New(), "WH-MAIN", "A1", 10)
		inv1.ReorderPoint = 20
		require.NoError(t, repo.Create(ctx, inv1))

		// Normal stock: 50 available, reorder_point=20
		inv2 := createTestInventoryForProduct(uuidv7.New(), "WH-MAIN", "A2", 50)
		inv2.ReorderPoint = 20
		require.NoError(t, repo.Create(ctx, inv2))

		// Low stock but inactive: should be excluded
		inv3 := createTestInventoryForProduct(uuidv7.New(), "WH-MAIN", "A3", 5)
		inv3.ReorderPoint = 20
		inv3.IsActive = false
		inv3.LastUpdatedBy = userID
		require.NoError(t, repo.Create(ctx, inv3))

		// Test: Get low stock items (active only)
		lowStockItems, err := repo.GetLowStock(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(lowStockItems), 1)

		// Verify low stock item is in results
		found := false
		for _, item := range lowStockItems {
			if item.ID == inv1.ID {
				found = true
				assert.True(t, item.IsActive)
				assert.Less(t, item.QuantityAvailable, item.ReorderPoint)
			}
		}
		assert.True(t, found, "Low stock item should be in results")
	})
}

// Test 7: GetByStatus (status filter)
func TestInventoryRepository_GetByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create items with different statuses
		userID := uuidv7.New()

		inv1 := createTestInventory()
		inv1.SetStatus(inventoryAggregate.InventoryStatusAvailable)
		inv1.LastUpdatedBy = userID
		require.NoError(t, repo.Create(ctx, inv1))

		inv2 := createTestInventory()
		inv2.SetStatus(inventoryAggregate.InventoryStatusDamaged)
		inv2.LastUpdatedBy = userID
		require.NoError(t, repo.Create(ctx, inv2))

		inv3 := createTestInventory()
		inv3.SetStatus(inventoryAggregate.InventoryStatusQuarantined)
		inv3.LastUpdatedBy = userID
		require.NoError(t, repo.Create(ctx, inv3))

		// Test: Get damaged items only
		damagedItems, err := repo.GetByStatus(ctx, inventoryAggregate.InventoryStatusDamaged)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(damagedItems), 1)
		for _, item := range damagedItems {
			assert.Equal(t, inventoryAggregate.InventoryStatusDamaged, item.Status)
		}
	})
}

// Test 8: BulkUpdate (batch operations in transaction)
func TestInventoryRepository_BulkUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Create 3 inventory items
		userID := uuidv7.New()
		items := make([]*inventoryAggregate.Inventory, 3)
		for i := 0; i < 3; i++ {
			items[i] = createTestInventory()
			require.NoError(t, repo.Create(ctx, items[i]))
		}

		// Update all items: increase quantity by 100
		for i := range items {
			err := items[i].ReceiveStock(100, 1500, userID)
			require.NoError(t, err)
		}

		// Test: Bulk update
		err := repo.BulkUpdate(ctx, items)
		require.NoError(t, err)

		// Verify all updated
		for _, item := range items {
			reloaded, err := repo.GetByID(ctx, item.ID)
			require.NoError(t, err)
			assert.Equal(t, item.QuantityOnHand, reloaded.QuantityOnHand)
			assert.GreaterOrEqual(t, reloaded.QuantityOnHand, 100)
		}
	})
}

// Test 9: Edge case - Empty warehouse
func TestInventoryRepository_GetByWarehouse_Empty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		// Test: Get items from non-existent warehouse
		items, err := repo.GetByWarehouse(ctx, "WH-NONEXISTENT")
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}
