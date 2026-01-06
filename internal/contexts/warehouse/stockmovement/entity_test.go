package stockmovement

import (
"testing"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"github.com/basilex/promenade/pkg/uuidv7"
)

func createTestStockMovement(t *testing.T) *StockMovement {
	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()
	sm, err := NewStockMovement(inventoryID, MovementTypeReceipt, 100, 50, createdBy)
	require.NoError(t, err)
	return sm
}

func TestNewStockMovement_Success(t *testing.T) {
	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()
	sm, err := NewStockMovement(inventoryID, MovementTypeReceipt, 100, 50, createdBy)
	assert.NoError(t, err)
	assert.NotNil(t, sm)
	assert.Equal(t, inventoryID, sm.InventoryID)
}

func TestNewStockMovement_MissingInventoryID(t *testing.T) {
	createdBy := uuidv7.New()
	sm, err := NewStockMovement(uuidv7.Nil, MovementTypeReceipt, 100, 50, createdBy)
	assert.Error(t, err)
	assert.Nil(t, sm)
}

func TestValidate_Success(t *testing.T) {
	sm := createTestStockMovement(t)
	err := sm.Validate()
	assert.NoError(t, err)
}

func TestNewStockMovement_ZeroQuantity(t *testing.T) {
	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()
	sm, err := NewStockMovement(inventoryID, MovementTypeReceipt, 0, 50, createdBy)
	assert.Error(t, err)
	assert.Nil(t, sm)
	assert.Contains(t, err.Error(), "quantity cannot be zero")
}

func TestNewStockMovement_InvalidMovementType(t *testing.T) {
	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()
	sm, err := NewStockMovement(inventoryID, MovementType("invalid"), 100, 50, createdBy)
	assert.Error(t, err)
	assert.Nil(t, sm)
	assert.Contains(t, err.Error(), "invalid movement type")
}

func TestSetReference_Success(t *testing.T) {
	sm := createTestStockMovement(t)
	orderID := uuidv7.New()
	sm.SetReference("order", &orderID)
	assert.Equal(t, "order", sm.ReferenceType)
	assert.Equal(t, &orderID, sm.ReferenceID)
}

func TestSetCost_Success(t *testing.T) {
	sm := createTestStockMovement(t)
	unitCost := int64(1500)
	err := sm.SetCost(unitCost, "USD")
	assert.NoError(t, err)
	assert.Equal(t, &unitCost, sm.UnitCostCents)
}

func TestSetReason_Success(t *testing.T) {
	sm := createTestStockMovement(t)
	reason := "Stock damaged"
	err := sm.SetReason(reason)
	assert.NoError(t, err)
	assert.Equal(t, reason, sm.Reason)
}

func TestIsAdjustment(t *testing.T) {
	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()
	sm, _ := NewStockMovement(inventoryID, MovementTypeAdjustment, 10, 50, createdBy)
	assert.True(t, sm.IsAdjustment())
}

func TestValidate_AdjustmentWithoutReason(t *testing.T) {
	inventoryID := uuidv7.New()
	createdBy := uuidv7.New()
	sm, _ := NewStockMovement(inventoryID, MovementTypeAdjustment, 10, 50, createdBy)
	err := sm.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "adjustment movements require a reason")
}
