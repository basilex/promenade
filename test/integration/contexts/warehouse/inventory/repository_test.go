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

func createTestInventory() *inventoryAggregate.Inventory {
	productID := uuidv7.New()
	// Use full UUID for guaranteed uniqueness
	sku := fmt.Sprintf("TEST-SKU-%s", uuidv7.New().String())
	userID := uuidv7.New()
	inv, _ := inventoryAggregate.NewInventory(
		productID,
		sku,
		"Test Product",
		"WH-MAIN",
		userID, // Created by
	)
	// Set initial quantities
	_ = inv.ReceiveStock(100, 1000, userID)
	inv.LocationCode = "A1"
	inv.ReorderPoint = 20
	inv.ReorderQuantity = 50
	return inv
}

func TestInventoryRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		inv := createTestInventory()
		err := repo.Create(ctx, inv)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, found.ID)
		assert.Equal(t, inv.SKU, found.SKU)
	})
}

func TestInventoryRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		inv := createTestInventory()
		require.NoError(t, repo.Create(ctx, inv))

		found, err := repo.GetByID(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, found.ID)

		notFound, err := repo.GetByID(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, inventory.ErrInventoryNotFound, err)
		assert.Nil(t, notFound)
	})
}

func TestInventoryRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		inv := createTestInventory()
		require.NoError(t, repo.Create(ctx, inv))

		userID := uuidv7.New()
		err := inv.ReceiveStock(50, 1500, userID)
		require.NoError(t, err)

		err = repo.Update(ctx, inv)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, 150, found.QuantityOnHand)
	})
}

func TestInventoryRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInventoryRepository(testDB.DB)

		inv := createTestInventory()
		require.NoError(t, repo.Create(ctx, inv))

		err := repo.Delete(ctx, inv.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, inv.ID)
		assert.Error(t, err)
		assert.Equal(t, inventory.ErrInventoryNotFound, err)
	})
}
