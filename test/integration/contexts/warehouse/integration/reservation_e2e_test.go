package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/integration"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	inventoryUseCase "github.com/basilex/promenade/internal/contexts/warehouse/inventory/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	testutils "github.com/basilex/promenade/test/integration"
)

func TestReservationService_SimpleReserve(t *testing.T) {
	db := testutils.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	// Setup dependencies
	invRepo := inventoryRepo.NewInventoryRepository(db.DB)
	inventoryUC := inventoryUseCase.NewInventoryUseCase(invRepo)
	reservationService := integration.NewReservationService(inventoryUC)

	// Create inventory
	productID := uuidv7.New()
	createdBy := uuidv7.New()
	inv, err := inventoryUC.CreateInventory(ctx, productID, "SKU-SIMPLE", "Simple Test", "WH-SIMPLE", createdBy)
	require.NoError(t, err)

	// Receive stock
	inv, err = inventoryUC.ReceiveStock(ctx, inv.ID, 100, 1000, createdBy)
	require.NoError(t, err)
	assert.Equal(t, 100, inv.QuantityAvailable)

	// Reserve stock
	orderID := uuidv7.New()
	err = reservationService.ReserveForOrder(ctx, orderID, []integration.OrderItem{
		{ProductID: productID, SKU: "SKU-SIMPLE", Quantity: 10},
	}, createdBy)
	require.NoError(t, err)

	// Verify reservation
	inv, err = inventoryUC.GetInventory(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, 10, inv.QuantityReserved)
	assert.Equal(t, 90, inv.QuantityAvailable)
}
