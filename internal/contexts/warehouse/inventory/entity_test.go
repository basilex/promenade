package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Test fixtures
func createTestInventory(t *testing.T) *Inventory {
	productID := uuidv7.New()
	inv, err := NewInventory(productID, "SKU-001", "Test Product", "WH-001", uuidv7.New())
	require.NoError(t, err)
	return inv
}

func createTestUser() uuidv7.UUID {
	return uuidv7.New()
}

// ==============================================================================
// Factory Method Tests (NewInventory)
// ==============================================================================

func TestNewInventory_Success(t *testing.T) {
	productID := uuidv7.New()
	sku := "SKU-001"
	productName := "Test Product"
	warehouseID := "WH-001"
	createdBy := uuidv7.New()

	inv, err := NewInventory(productID, sku, productName, warehouseID, createdBy)

	assert.NoError(t, err)
	assert.NotNil(t, inv)
	assert.NotEqual(t, uuidv7.Nil, inv.GetID())
	assert.Equal(t, productID, inv.ProductID)
	assert.Equal(t, sku, inv.SKU)
	assert.Equal(t, productName, inv.ProductName)
	assert.Equal(t, warehouseID, inv.WarehouseID)
	assert.Equal(t, 0, inv.QuantityOnHand)
	assert.Equal(t, 0, inv.QuantityReserved)
	assert.Equal(t, 0, inv.QuantityCommitted)
	assert.Equal(t, 0, inv.QuantityAvailable)
	assert.Equal(t, 10, inv.ReorderPoint)
	assert.Equal(t, 50, inv.ReorderQuantity)
	assert.Equal(t, InventoryStatusAvailable, inv.Status)
	assert.True(t, inv.IsActive)
	assert.Equal(t, "USD", inv.CurrencyCode)
}

func TestNewInventory_MissingProductID(t *testing.T) {
	inv, err := NewInventory(uuidv7.Nil, "SKU-001", "Test Product", "WH-001", uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.Contains(t, err.Error(), "product ID is required")
}

func TestNewInventory_MissingSKU(t *testing.T) {
	productID := uuidv7.New()
	inv, err := NewInventory(productID, "", "Test Product", "WH-001", uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.Contains(t, err.Error(), "SKU is required")
}

func TestNewInventory_MissingProductName(t *testing.T) {
	productID := uuidv7.New()
	inv, err := NewInventory(productID, "SKU-001", "", "WH-001", uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.Contains(t, err.Error(), "product name is required")
}

func TestNewInventory_MissingWarehouseID(t *testing.T) {
	productID := uuidv7.New()
	inv, err := NewInventory(productID, "SKU-001", "Test Product", "", uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, inv)
	assert.Contains(t, err.Error(), "warehouse ID is required")
}

// ==============================================================================
// ReceiveStock Tests
// ==============================================================================

func TestReceiveStock_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	initialTime := inv.LastRestocked

	err := inv.ReceiveStock(100, 1000, userID)

	assert.NoError(t, err)
	assert.Equal(t, 100, inv.QuantityOnHand)
	assert.Equal(t, 100, inv.QuantityAvailable)
	assert.Equal(t, int64(1000), inv.UnitCostCents)
	assert.Equal(t, userID, inv.LastUpdatedBy)
	assert.True(t, inv.LastRestocked.After(initialTime))
}

func TestReceiveStock_WeightedAverageCost(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	// First receipt: 100 units at $10.00 each
	err := inv.ReceiveStock(100, 1000, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), inv.UnitCostCents)

	// Second receipt: 50 units at $12.00 each
	// Expected: ((100 * 1000) + (50 * 1200)) / 150 = 160000 / 150 = 1066.66 cents
	err = inv.ReceiveStock(50, 1200, userID)
	require.NoError(t, err)
	assert.Equal(t, 150, inv.QuantityOnHand)
	assert.Equal(t, int64(1066), inv.UnitCostCents)
}

func TestReceiveStock_ZeroQuantity(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.ReceiveStock(0, 1000, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestReceiveStock_NegativeQuantity(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.ReceiveStock(-10, 1000, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestReceiveStock_InactiveInventory(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	inv.IsActive = false

	err := inv.ReceiveStock(100, 1000, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot receive stock for inactive inventory")
}

// ==============================================================================
// ReserveStock Tests
// ==============================================================================

func TestReserveStock_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)

	err := inv.ReserveStock(30, orderID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 30, inv.QuantityReserved)
	assert.Equal(t, 70, inv.QuantityAvailable)
	assert.Equal(t, 100, inv.QuantityOnHand)
}

func TestReserveStock_MultipleReservations(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.ReceiveStock(100, 1000, userID)

	_ = inv.ReserveStock(30, uuidv7.New(), userID)
	assert.Equal(t, 30, inv.QuantityReserved)
	assert.Equal(t, 70, inv.QuantityAvailable)

	_ = inv.ReserveStock(20, uuidv7.New(), userID)
	assert.Equal(t, 50, inv.QuantityReserved)
	assert.Equal(t, 50, inv.QuantityAvailable)
}

func TestReserveStock_InsufficientStock(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(50, 1000, userID)

	err := inv.ReserveStock(100, orderID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")
	assert.Equal(t, 0, inv.QuantityReserved)
}

func TestReserveStock_ZeroQuantity(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	err := inv.ReserveStock(0, orderID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestReserveStock_InactiveInventory(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	inv.IsActive = false

	err := inv.ReserveStock(30, orderID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reserve stock for inactive inventory")
}

// ==============================================================================
// ReleaseReservation Tests
// ==============================================================================

func TestReleaseReservation_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(30, orderID, userID)

	err := inv.ReleaseReservation(30, orderID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 0, inv.QuantityReserved)
	assert.Equal(t, 100, inv.QuantityAvailable)
}

func TestReleaseReservation_PartialRelease(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(50, orderID, userID)

	err := inv.ReleaseReservation(20, orderID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 30, inv.QuantityReserved)
	assert.Equal(t, 70, inv.QuantityAvailable)
}

func TestReleaseReservation_InsufficientReserved(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(30, orderID, userID)

	err := inv.ReleaseReservation(50, orderID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient reserved stock")
	assert.Equal(t, 30, inv.QuantityReserved)
}

func TestReleaseReservation_ZeroQuantity(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	err := inv.ReleaseReservation(0, orderID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

// ==============================================================================
// CommitReservation Tests
// ==============================================================================

func TestCommitReservation_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(30, orderID, userID)

	err := inv.CommitReservation(30, orderID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 0, inv.QuantityReserved)
	assert.Equal(t, 30, inv.QuantityCommitted)
	assert.Equal(t, 70, inv.QuantityOnHand)
	assert.Equal(t, 70, inv.QuantityAvailable)
}

func TestCommitReservation_PartialCommit(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(50, orderID, userID)

	err := inv.CommitReservation(20, orderID, userID)

	assert.NoError(t, err)
	assert.Equal(t, 30, inv.QuantityReserved)
	assert.Equal(t, 20, inv.QuantityCommitted)
	assert.Equal(t, 80, inv.QuantityOnHand)
	assert.Equal(t, 50, inv.QuantityAvailable)
}

func TestCommitReservation_InsufficientReserved(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(30, orderID, userID)

	err := inv.CommitReservation(50, orderID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient reserved stock")
}

func TestCommitReservation_ZeroQuantity(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	err := inv.CommitReservation(0, orderID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

// ==============================================================================
// AdjustStock Tests
// ==============================================================================

func TestAdjustStock_PositiveAdjustment(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.ReceiveStock(100, 1000, userID)

	err := inv.AdjustStock(20, "Found extra units in audit", userID)

	assert.NoError(t, err)
	assert.Equal(t, 120, inv.QuantityOnHand)
	assert.Equal(t, 120, inv.QuantityAvailable)
}

func TestAdjustStock_NegativeAdjustment(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.ReceiveStock(100, 1000, userID)

	err := inv.AdjustStock(-15, "Damaged units removed", userID)

	assert.NoError(t, err)
	assert.Equal(t, 85, inv.QuantityOnHand)
	assert.Equal(t, 85, inv.QuantityAvailable)
}

func TestAdjustStock_ResultsInNegative(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.ReceiveStock(50, 1000, userID)

	err := inv.AdjustStock(-100, "Theft", userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "adjustment would result in negative stock")
	assert.Equal(t, 50, inv.QuantityOnHand)
}

func TestAdjustStock_MissingReason(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.AdjustStock(10, "", userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "adjustment reason is required")
}

// ==============================================================================
// SetLocation Tests
// ==============================================================================

func TestSetLocation_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.SetLocation("A-12-03", "Zone-A", userID)

	assert.NoError(t, err)
	assert.Equal(t, "A-12-03", inv.LocationCode)
	assert.Equal(t, "Zone-A", inv.LocationZone)
	assert.Equal(t, userID, inv.LastUpdatedBy)
}

func TestSetLocation_MissingLocationCode(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.SetLocation("", "Zone-A", userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "location code is required")
}

func TestSetLocation_UpdateExisting(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.SetLocation("A-12-03", "Zone-A", userID)
	err := inv.SetLocation("B-05-10", "Zone-B", userID)

	assert.NoError(t, err)
	assert.Equal(t, "B-05-10", inv.LocationCode)
	assert.Equal(t, "Zone-B", inv.LocationZone)
}

// ==============================================================================
// SetReorderPoint Tests
// ==============================================================================

func TestSetReorderPoint_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.SetReorderPoint(25, 100, userID)

	assert.NoError(t, err)
	assert.Equal(t, 25, inv.ReorderPoint)
	assert.Equal(t, 100, inv.ReorderQuantity)
}

func TestSetReorderPoint_NegativePoint(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.SetReorderPoint(-5, 100, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reorder point cannot be negative")
}

func TestSetReorderPoint_ZeroQuantity(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.SetReorderPoint(10, 0, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reorder quantity must be positive")
}

// ==============================================================================
// MarkAsDamaged Tests
// ==============================================================================

func TestMarkAsDamaged_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.ReceiveStock(100, 1000, userID)

	err := inv.MarkAsDamaged(15, "Water damage", userID)

	assert.NoError(t, err)
	assert.Equal(t, 85, inv.QuantityOnHand)
	assert.Equal(t, 85, inv.QuantityAvailable)
	assert.Equal(t, InventoryStatusDamaged, inv.Status)
	assert.Contains(t, inv.Notes, "Water damage")
}

func TestMarkAsDamaged_InsufficientStock(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.ReceiveStock(50, 1000, userID)

	err := inv.MarkAsDamaged(100, "Water damage", userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient available stock")
}

func TestMarkAsDamaged_ZeroQuantity(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.MarkAsDamaged(0, "Damage", userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

// ==============================================================================
// Activate/Deactivate Tests
// ==============================================================================

func TestActivate_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	inv.IsActive = false
	inv.Status = InventoryStatusDamaged

	err := inv.Activate(userID)

	assert.NoError(t, err)
	assert.True(t, inv.IsActive)
	assert.Equal(t, InventoryStatusAvailable, inv.Status)
}

func TestActivate_AlreadyActive(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.Activate(userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inventory is already active")
}

func TestDeactivate_Success(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	err := inv.Deactivate(userID)

	assert.NoError(t, err)
	assert.False(t, inv.IsActive)
}

func TestDeactivate_WithReservedStock(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(30, orderID, userID)

	err := inv.Deactivate(userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot deactivate inventory with reserved stock")
	assert.True(t, inv.IsActive)
}

func TestDeactivate_AlreadyInactive(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	inv.IsActive = false

	err := inv.Deactivate(userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inventory is already inactive")
}

// ==============================================================================
// IsLowStock Tests
// ==============================================================================

func TestIsLowStock_BelowReorderPoint(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.SetReorderPoint(20, 50, userID)
	_ = inv.ReceiveStock(15, 1000, userID)

	assert.True(t, inv.IsLowStock())
}

func TestIsLowStock_AtReorderPoint(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.SetReorderPoint(20, 50, userID)
	_ = inv.ReceiveStock(20, 1000, userID)

	assert.True(t, inv.IsLowStock())
}

func TestIsLowStock_AboveReorderPoint(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.SetReorderPoint(20, 50, userID)
	_ = inv.ReceiveStock(50, 1000, userID)

	assert.False(t, inv.IsLowStock())
}

func TestIsLowStock_WithReservations(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.SetReorderPoint(20, 50, userID)
	_ = inv.ReceiveStock(50, 1000, userID)
	_ = inv.ReserveStock(35, orderID, userID)

	assert.True(t, inv.IsLowStock())
}

// ==============================================================================
// GetStockValue Tests
// ==============================================================================

func TestGetStockValue_EmptyStock(t *testing.T) {
	inv := createTestInventory(t)

	value := inv.GetStockValue()

	assert.Equal(t, int64(0), value)
}

func TestGetStockValue_WithStock(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()

	_ = inv.ReceiveStock(100, 1250, userID)

	value := inv.GetStockValue()

	assert.Equal(t, int64(125000), value)
}

func TestGetStockValue_AfterCommit(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1000, userID)
	_ = inv.ReserveStock(30, orderID, userID)
	_ = inv.CommitReservation(30, orderID, userID)

	value := inv.GetStockValue()

	assert.Equal(t, int64(70000), value)
}

// ==============================================================================
// Validate Tests
// ==============================================================================

func TestValidate_ValidInventory(t *testing.T) {
	inv := createTestInventory(t)

	err := inv.Validate()

	assert.NoError(t, err)
}

func TestValidate_MissingProductID(t *testing.T) {
	inv := createTestInventory(t)
	inv.ProductID = uuidv7.Nil

	err := inv.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product ID is required")
}

func TestValidate_MissingSKU(t *testing.T) {
	inv := createTestInventory(t)
	inv.SKU = ""

	err := inv.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SKU is required")
}

func TestValidate_NegativeQuantityOnHand(t *testing.T) {
	inv := createTestInventory(t)
	inv.QuantityOnHand = -10

	err := inv.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity on hand cannot be negative")
}

func TestValidate_NegativeReserved(t *testing.T) {
	inv := createTestInventory(t)
	inv.QuantityReserved = -5

	err := inv.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "quantity reserved cannot be negative")
}

func TestValidate_NotesTooLong(t *testing.T) {
	inv := createTestInventory(t)
	inv.Notes = string(make([]byte, 501))

	err := inv.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notes exceed 500 characters")
}

func TestValidate_InvalidReorderQuantity(t *testing.T) {
	inv := createTestInventory(t)
	inv.ReorderQuantity = 0

	err := inv.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reorder quantity must be positive")
}

// ==============================================================================
// Complex Workflow Tests
// ==============================================================================

func TestComplexWorkflow_FullOrderFulfillment(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(200, 1000, userID)
	assert.Equal(t, 200, inv.QuantityAvailable)

	_ = inv.ReserveStock(50, orderID, userID)
	assert.Equal(t, 150, inv.QuantityAvailable)
	assert.Equal(t, 50, inv.QuantityReserved)

	_ = inv.CommitReservation(50, orderID, userID)
	assert.Equal(t, 150, inv.QuantityOnHand)
	assert.Equal(t, 150, inv.QuantityAvailable)
	assert.Equal(t, 0, inv.QuantityReserved)
	assert.Equal(t, 50, inv.QuantityCommitted)

	expectedValue := int64(150 * 1000)
	assert.Equal(t, expectedValue, inv.GetStockValue())
}

func TestComplexWorkflow_OrderCancellation(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.ReceiveStock(100, 1500, userID)

	_ = inv.ReserveStock(40, orderID, userID)
	assert.Equal(t, 60, inv.QuantityAvailable)
	assert.Equal(t, 40, inv.QuantityReserved)

	err := inv.ReleaseReservation(40, orderID, userID)
	require.NoError(t, err)
	assert.Equal(t, 100, inv.QuantityAvailable)
	assert.Equal(t, 0, inv.QuantityReserved)
}

func TestComplexWorkflow_MultipleOrders(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	order1 := uuidv7.New()
	order2 := uuidv7.New()
	order3 := uuidv7.New()

	_ = inv.ReceiveStock(200, 1000, userID)

	_ = inv.ReserveStock(50, order1, userID)
	_ = inv.ReserveStock(30, order2, userID)
	_ = inv.ReserveStock(20, order3, userID)

	assert.Equal(t, 100, inv.QuantityReserved)
	assert.Equal(t, 100, inv.QuantityAvailable)

	_ = inv.CommitReservation(50, order1, userID)
	assert.Equal(t, 50, inv.QuantityReserved)
	assert.Equal(t, 100, inv.QuantityAvailable)

	_ = inv.ReleaseReservation(30, order2, userID)
	assert.Equal(t, 20, inv.QuantityReserved)
	assert.Equal(t, 130, inv.QuantityAvailable)

	_ = inv.CommitReservation(20, order3, userID)
	assert.Equal(t, 0, inv.QuantityReserved)
	assert.Equal(t, 130, inv.QuantityAvailable)
	assert.Equal(t, 70, inv.QuantityCommitted)
}

func TestComplexWorkflow_LowStockAlert(t *testing.T) {
	inv := createTestInventory(t)
	userID := createTestUser()
	orderID := uuidv7.New()

	_ = inv.SetReorderPoint(30, 100, userID)
	_ = inv.ReceiveStock(100, 1000, userID)

	assert.False(t, inv.IsLowStock())

	_ = inv.ReserveStock(75, orderID, userID)
	assert.True(t, inv.IsLowStock())
}
