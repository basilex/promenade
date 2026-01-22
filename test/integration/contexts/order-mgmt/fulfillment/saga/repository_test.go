package saga_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/fulfillment/saga"
	"github.com/basilex/promenade/internal/contexts/order-mgmt/fulfillment/saga/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// setupTestDB creates a test database and saga table using standard test infrastructure
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	// Use standard test infrastructure (CI-aware: port 5432 in CI, 5433 locally)
	testDB := integration.SetupTestDB(t)
	db := testDB.DB

	// Create saga table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS order_fulfillment_sagas (
			id UUID PRIMARY KEY,
			order_id UUID NOT NULL,
			customer_id UUID NOT NULL,
			state VARCHAR(50) NOT NULL,
			current_step INT NOT NULL,
			completed_steps JSONB NOT NULL,
			failed_step VARCHAR(100),
			payment_id UUID,
			reserved_items JSONB NOT NULL,
			shipment_id UUID,
			tracking_number VARCHAR(200),
			started_at TIMESTAMP NOT NULL,
			completed_at TIMESTAMP,
			cancelled_at TIMESTAMP,
			failure_reason TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	require.NoError(t, err, "Failed to create saga table")

	// Clean up existing data
	_, err = db.Exec("TRUNCATE order_fulfillment_sagas")
	require.NoError(t, err, "Failed to truncate saga table")

	return db
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE order_fulfillment_sagas")
	require.NoError(t, err)
	_ = db.Close() // Ignore error on close
}

// ============================================================================
// CRUD Tests
// ============================================================================

func TestSagaRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Save saga
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Verify saved
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM order_fulfillment_sagas WHERE id = $1", s.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify field values
	var dbSaga struct {
		ID             uuidv7.UUID `db:"id"`
		OrderID        uuidv7.UUID `db:"order_id"`
		CustomerID     uuidv7.UUID `db:"customer_id"`
		State          string      `db:"state"`
		CurrentStep    int         `db:"current_step"`
		CompletedSteps string      `db:"completed_steps"`
		ReservedItems  string      `db:"reserved_items"`
	}
	err = db.Get(&dbSaga, "SELECT id, order_id, customer_id, state, current_step, completed_steps, reserved_items FROM order_fulfillment_sagas WHERE id = $1", s.ID)
	require.NoError(t, err)

	assert.Equal(t, s.ID, dbSaga.ID)
	assert.Equal(t, orderID, dbSaga.OrderID)
	assert.Equal(t, customerID, dbSaga.CustomerID)
	assert.Equal(t, "pending", dbSaga.State)
	assert.Equal(t, 0, dbSaga.CurrentStep)
	assert.Equal(t, "[]", dbSaga.CompletedSteps) // Empty array
	assert.Equal(t, "[]", dbSaga.ReservedItems)  // Empty array
}

func TestSagaRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Save initial
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Update saga state
	s.StartPaymentProcessing()
	paymentID := uuidv7.New()
	s.CompletePayment(paymentID)

	// Update in DB
	err = repo.Update(ctx, s)
	require.NoError(t, err)

	// Verify updated
	var dbState string
	var dbCurrentStep int
	var dbCompletedSteps string
	var dbPaymentID *uuidv7.UUID
	err = db.QueryRow("SELECT state, current_step, completed_steps, payment_id FROM order_fulfillment_sagas WHERE id = $1", s.ID).
		Scan(&dbState, &dbCurrentStep, &dbCompletedSteps, &dbPaymentID)
	require.NoError(t, err)

	assert.Equal(t, "inventory_processing", dbState)
	assert.Equal(t, 2, dbCurrentStep) // After CompletePayment, CurrentStep is 2 (inventory step)
	assert.Contains(t, dbCompletedSteps, "payment")
	require.NotNil(t, dbPaymentID)
	assert.Equal(t, paymentID, *dbPaymentID)
}

func TestSagaRepository_FindByID_Success(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)
	s.StartPaymentProcessing()

	// Save
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Find
	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, s.ID, found.ID)
	assert.Equal(t, orderID, found.OrderID)
	assert.Equal(t, customerID, found.CustomerID)
	assert.Equal(t, saga.FulfillmentSagaStatePaymentProcessing, found.State)
	assert.Equal(t, 1, found.CurrentStep) // StartPaymentProcessing sets CurrentStep to 1
}

func TestSagaRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	nonExistentID := uuidv7.New()
	found, err := repo.FindByID(ctx, nonExistentID)

	assert.NoError(t, err) // Not found is not an error
	assert.Nil(t, found)
}

// ============================================================================
// Query Tests
// ============================================================================

func TestSagaRepository_FindByOrderID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Save
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Find by order ID
	found, err := repo.FindByOrderID(ctx, orderID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, s.ID, found.ID)
	assert.Equal(t, orderID, found.OrderID)
}

func TestSagaRepository_FindByOrderID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	nonExistentOrderID := uuidv7.New()
	found, err := repo.FindByOrderID(ctx, nonExistentOrderID)

	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestSagaRepository_FindInProgressSagas(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	// Create sagas in different states
	orderID1 := uuidv7.New()
	customerID1 := uuidv7.New()
	saga1 := saga.NewFulfillmentSaga(orderID1, customerID1)
	saga1.StartPaymentProcessing() // In progress
	err := repo.Save(ctx, saga1)
	require.NoError(t, err)

	orderID2 := uuidv7.New()
	customerID2 := uuidv7.New()
	saga2 := saga.NewFulfillmentSaga(orderID2, customerID2)
	saga2.StartPaymentProcessing()
	paymentID := uuidv7.New()
	saga2.CompletePayment(paymentID)
	saga2.CompleteInventory([]saga.ReservedItem{}) // In progress
	err = repo.Save(ctx, saga2)
	require.NoError(t, err)

	orderID3 := uuidv7.New()
	customerID3 := uuidv7.New()
	saga3 := saga.NewFulfillmentSaga(orderID3, customerID3) // Pending (not in progress)
	err = repo.Save(ctx, saga3)
	require.NoError(t, err)

	orderID4 := uuidv7.New()
	customerID4 := uuidv7.New()
	saga4 := saga.NewFulfillmentSaga(orderID4, customerID4)
	saga4.StartPaymentProcessing()
	saga4.StartCompensation("payment", "test failure")
	err = repo.Save(ctx, saga4)
	require.NoError(t, err)

	// Find in-progress sagas
	inProgress, err := repo.FindInProgressSagas(ctx)
	require.NoError(t, err)

	// Should find 3 sagas (payment_processing, inventory_processing, compensating)
	assert.Len(t, inProgress, 3)

	// Verify states
	states := make(map[string]bool)
	for _, s := range inProgress {
		states[string(s.State)] = true
	}
	assert.True(t, states["payment_processing"])
	assert.True(t, states["shipping_processing"]) // saga2 is in shipping_processing after CompleteInventory
	assert.True(t, states["compensating"])
}

// ============================================================================
// JSON Serialization Tests
// ============================================================================

func TestSagaRepository_CompletedSteps_RoundTrip(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Complete multiple steps
	s.StartPaymentProcessing()
	paymentID := uuidv7.New()
	s.CompletePayment(paymentID)

	// Complete inventory (only once, with all items)
	s.CompleteInventory([]saga.ReservedItem{
		{ProductID: uuidv7.New(), Quantity: 2},
		{ProductID: uuidv7.New(), Quantity: 5},
	})

	// Save
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Retrieve
	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify completed steps
	completedSteps := found.GetCompletedSteps()
	assert.Len(t, completedSteps, 2)
	assert.Contains(t, completedSteps, "payment")
	assert.Contains(t, completedSteps, "inventory")
}

func TestSagaRepository_ReservedItems_RoundTrip(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Add reserved items
	product1 := uuidv7.New()
	product2 := uuidv7.New()
	items := []saga.ReservedItem{
		{ProductID: product1, Quantity: 3},
		{ProductID: product2, Quantity: 7},
	}

	s.StartPaymentProcessing()
	paymentID := uuidv7.New()
	s.CompletePayment(paymentID)
	s.CompleteInventory([]saga.ReservedItem{})
	s.CompleteInventory(items)

	// Save
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Retrieve
	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify reserved items
	reservedItems := found.GetReservedItems()
	assert.Len(t, reservedItems, 2)
	assert.Equal(t, product1, reservedItems[0].ProductID)
	assert.Equal(t, 3, reservedItems[0].Quantity)
	assert.Equal(t, product2, reservedItems[1].ProductID)
	assert.Equal(t, 7, reservedItems[1].Quantity)
}

func TestSagaRepository_EmptyArrays_RoundTrip(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Save with empty arrays
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Retrieve
	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify empty arrays (not nil)
	completedSteps := found.GetCompletedSteps()
	reservedItems := found.GetReservedItems()
	assert.NotNil(t, completedSteps)
	assert.Len(t, completedSteps, 0)
	assert.NotNil(t, reservedItems)
	assert.Len(t, reservedItems, 0)
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestSagaRepository_NullableFields(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Save with nullable fields unset
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Retrieve
	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify nullable fields are nil
	assert.Nil(t, found.FailedStep)
	assert.Nil(t, found.PaymentID)
	assert.Nil(t, found.ShipmentID)
	assert.Nil(t, found.TrackingNumber)
	assert.Nil(t, found.CompletedAt)
	assert.Nil(t, found.CancelledAt)
	assert.Nil(t, found.FailureReason)
}

func TestSagaRepository_TimestampPrecision(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	beforeSave := time.Now().UTC()

	// Save
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	afterSave := time.Now().UTC()

	// Retrieve
	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify timestamp is within range (with 1-second tolerance)
	// StartedAt should be >= beforeSave (with 1s tolerance) and <= afterSave (with 1s tolerance)
	assert.True(t, found.StartedAt.After(beforeSave.Add(-time.Second)) || found.StartedAt.Equal(beforeSave.Add(-time.Second)),
		"StartedAt should be >= beforeSave-1s, got %v vs %v", found.StartedAt, beforeSave)
	assert.True(t, found.StartedAt.Before(afterSave.Add(time.Second)) || found.StartedAt.Equal(afterSave.Add(time.Second)),
		"StartedAt should be <= afterSave+1s, got %v vs %v", found.StartedAt, afterSave)
	assert.True(t, found.CreatedAt.After(beforeSave.Add(-time.Second)) || found.CreatedAt.Equal(beforeSave.Add(-time.Second)),
		"CreatedAt should be >= beforeSave-1s, got %v vs %v", found.CreatedAt, beforeSave)
	assert.True(t, found.CreatedAt.Before(afterSave.Add(time.Second)) || found.CreatedAt.Equal(afterSave.Add(time.Second)),
		"CreatedAt should be <= afterSave+1s, got %v vs %v", found.CreatedAt, afterSave)
}

func TestSagaRepository_ConcurrentUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Save initial
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Simulate concurrent updates (last write wins)
	s1 := *s
	s1.StartPaymentProcessing()

	s2 := *s
	s2.StartCompensation("payment", "test failure")

	// Both updates should succeed (last write wins)
	err = repo.Update(ctx, &s1)
	require.NoError(t, err)

	err = repo.Update(ctx, &s2)
	require.NoError(t, err)

	// Retrieve final state (should be s2)
	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, saga.FulfillmentSagaStateCompensating, found.State)
	assert.Equal(t, "payment", *found.FailedStep)
}

// ============================================================================
// Error Tests
// ============================================================================

func TestSagaRepository_SaveDuplicateID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Save first time
	err := repo.Save(ctx, s)
	require.NoError(t, err)

	// Try to save again with same ID
	err = repo.Save(ctx, s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key")
}

func TestSagaRepository_UpdateNonExistent(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := postgres.NewSagaRepository(db)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()
	s := saga.NewFulfillmentSaga(orderID, customerID)

	// Try to update without saving first
	err := repo.Update(ctx, s)

	// Should not error (UPDATE with 0 rows affected is valid)
	assert.NoError(t, err)
}
