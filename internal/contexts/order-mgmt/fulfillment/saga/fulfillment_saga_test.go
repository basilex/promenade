package saga

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFulfillmentSaga(t *testing.T) {
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	saga := NewFulfillmentSaga(orderID, customerID)

	require.NotNil(t, saga)
	assert.NotEqual(t, uuidv7.UUID{}, saga.ID)
	assert.Equal(t, orderID, saga.OrderID)
	assert.Equal(t, customerID, saga.CustomerID)
	assert.Equal(t, FulfillmentSagaStatePending, saga.State)
	assert.Equal(t, 0, saga.CurrentStep)
	assert.Empty(t, saga.GetCompletedSteps())
	assert.Empty(t, saga.GetReservedItems())
	assert.Nil(t, saga.PaymentID)
	assert.Nil(t, saga.ShipmentID)
	assert.Nil(t, saga.TrackingNumber)
	assert.Nil(t, saga.FailedStep)
	assert.Nil(t, saga.FailureReason)
	assert.Nil(t, saga.CompletedAt)
	assert.Nil(t, saga.CancelledAt)
	assert.False(t, saga.StartedAt.IsZero())
	assert.False(t, saga.CreatedAt.IsZero())
	assert.False(t, saga.UpdatedAt.IsZero())
}

func TestFulfillmentSaga_StartPaymentProcessing(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	initialUpdatedAt := saga.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure time difference

	saga.StartPaymentProcessing()

	assert.Equal(t, FulfillmentSagaStatePaymentProcessing, saga.State)
	assert.Equal(t, 1, saga.CurrentStep)
	assert.True(t, saga.UpdatedAt.After(initialUpdatedAt))
}

func TestFulfillmentSaga_CompletePayment(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.StartPaymentProcessing()
	paymentID := uuidv7.New()

	saga.CompletePayment(paymentID)

	assert.Equal(t, FulfillmentSagaStateInventoryProcessing, saga.State)
	assert.Equal(t, 2, saga.CurrentStep)
	assert.NotNil(t, saga.PaymentID)
	assert.Equal(t, paymentID, *saga.PaymentID)
	assert.Contains(t, saga.GetCompletedSteps(), "payment")
}

func TestFulfillmentSaga_CompleteInventory(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.StartPaymentProcessing()
	saga.CompletePayment(uuidv7.New())
	items := []ReservedItem{
		{
			ProductID:     uuidv7.New(),
			Quantity:      5,
			ReservationID: uuidv7.New(),
		},
		{
			ProductID:     uuidv7.New(),
			Quantity:      3,
			ReservationID: uuidv7.New(),
		},
	}

	saga.CompleteInventory(items)

	assert.Equal(t, FulfillmentSagaStateShippingProcessing, saga.State)
	assert.Equal(t, 3, saga.CurrentStep)
	assert.Len(t, saga.GetReservedItems(), 2)
	assert.Equal(t, items[0].ProductID, saga.GetReservedItems()[0].ProductID)
	assert.Equal(t, items[1].Quantity, saga.GetReservedItems()[1].Quantity)
	assert.Contains(t, saga.GetCompletedSteps(), "payment")
	assert.Contains(t, saga.GetCompletedSteps(), "inventory")
}

func TestFulfillmentSaga_CompleteShipping(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.StartPaymentProcessing()
	saga.CompletePayment(uuidv7.New())
	saga.CompleteInventory([]ReservedItem{})
	shipmentID := uuidv7.New()
	tracking := "TRACK-12345678"

	saga.CompleteShipping(shipmentID, tracking)

	assert.Equal(t, FulfillmentSagaStateCompleted, saga.State)
	assert.NotNil(t, saga.ShipmentID)
	assert.Equal(t, shipmentID, *saga.ShipmentID)
	assert.NotNil(t, saga.TrackingNumber)
	assert.Equal(t, tracking, *saga.TrackingNumber)
	assert.NotNil(t, saga.CompletedAt)
	assert.Contains(t, saga.GetCompletedSteps(), "payment")
	assert.Contains(t, saga.GetCompletedSteps(), "inventory")
	assert.Contains(t, saga.GetCompletedSteps(), "shipping")
}

func TestFulfillmentSaga_StartCompensation(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.StartPaymentProcessing()
	saga.CompletePayment(uuidv7.New())
	failedStep := "inventory"
	reason := "Out of stock"

	saga.StartCompensation(failedStep, reason)

	assert.Equal(t, FulfillmentSagaStateCompensating, saga.State)
	assert.NotNil(t, saga.FailedStep)
	assert.Equal(t, failedStep, *saga.FailedStep)
	assert.NotNil(t, saga.FailureReason)
	assert.Equal(t, reason, *saga.FailureReason)
}

func TestFulfillmentSaga_Cancel(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.StartPaymentProcessing()
	saga.StartCompensation("payment", "Card declined")
	reason := "Payment failed after retries"

	saga.Cancel(reason)

	assert.Equal(t, FulfillmentSagaStateCancelled, saga.State)
	assert.NotNil(t, saga.FailureReason)
	assert.Equal(t, reason, *saga.FailureReason)
	assert.NotNil(t, saga.CancelledAt)
	assert.False(t, saga.CancelledAt.IsZero())
}

func TestFulfillmentSaga_IsInProgress(t *testing.T) {
	tests := []struct {
		name     string
		state    FulfillmentSagaState
		expected bool
	}{
		{"pending is not in progress", FulfillmentSagaStatePending, false},
		{"payment processing is in progress", FulfillmentSagaStatePaymentProcessing, true},
		{"inventory processing is in progress", FulfillmentSagaStateInventoryProcessing, true},
		{"shipping processing is in progress", FulfillmentSagaStateShippingProcessing, true},
		{"compensating is in progress", FulfillmentSagaStateCompensating, true},
		{"completed is not in progress", FulfillmentSagaStateCompleted, false},
		{"cancelled is not in progress", FulfillmentSagaStateCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
			saga.State = tt.state

			assert.Equal(t, tt.expected, saga.IsInProgress())
		})
	}
}

func TestFulfillmentSaga_GetCompletedSteps(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	assert.Empty(t, saga.GetCompletedSteps())

	saga.StartPaymentProcessing()
	saga.CompletePayment(uuidv7.New())
	steps := saga.GetCompletedSteps()
	assert.Len(t, steps, 1)
	assert.Equal(t, "payment", steps[0])

	saga.CompleteInventory([]ReservedItem{})
	steps = saga.GetCompletedSteps()
	assert.Len(t, steps, 2)
	assert.Equal(t, "payment", steps[0])
	assert.Equal(t, "inventory", steps[1])
}

func TestFulfillmentSaga_GetReservedItems(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	assert.Empty(t, saga.GetReservedItems())

	items := []ReservedItem{
		{ProductID: uuidv7.New(), Quantity: 10, ReservationID: uuidv7.New()},
	}
	saga.StartPaymentProcessing()
	saga.CompletePayment(uuidv7.New())
	saga.CompleteInventory(items)

	reservedItems := saga.GetReservedItems()
	assert.Len(t, reservedItems, 1)
	assert.Equal(t, items[0].ProductID, reservedItems[0].ProductID)
	assert.Equal(t, items[0].Quantity, reservedItems[0].Quantity)
}

func TestFulfillmentSaga_FullWorkflow_Success(t *testing.T) {
	// Create saga
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	saga := NewFulfillmentSaga(orderID, customerID)
	assert.Equal(t, FulfillmentSagaStatePending, saga.State)

	// Step 1: Payment
	saga.StartPaymentProcessing()
	paymentID := uuidv7.New()
	saga.CompletePayment(paymentID)
	assert.Equal(t, FulfillmentSagaStateInventoryProcessing, saga.State)

	// Step 2: Inventory
	items := []ReservedItem{
		{ProductID: uuidv7.New(), Quantity: 5, ReservationID: uuidv7.New()},
	}
	saga.CompleteInventory(items)
	assert.Equal(t, FulfillmentSagaStateShippingProcessing, saga.State)

	// Step 3: Shipping
	shipmentID := uuidv7.New()
	tracking := "TRACK-ABC123"
	saga.CompleteShipping(shipmentID, tracking)

	// Verify final state
	assert.Equal(t, FulfillmentSagaStateCompleted, saga.State)
	assert.Len(t, saga.GetCompletedSteps(), 3)
	assert.NotNil(t, saga.CompletedAt)
	assert.False(t, saga.IsInProgress())
}

func TestFulfillmentSaga_FullWorkflow_FailureWithCompensation(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	// Step 1: Payment succeeds
	saga.StartPaymentProcessing()
	saga.CompletePayment(uuidv7.New())

	// Step 2: Inventory fails
	failedStep := "inventory"
	reason := "Product out of stock"
	saga.StartCompensation(failedStep, reason)

	// Compensation completes
	saga.Cancel("Saga compensation completed")

	// Verify final state
	assert.Equal(t, FulfillmentSagaStateCancelled, saga.State)
	assert.Equal(t, failedStep, *saga.FailedStep)
	assert.NotNil(t, saga.CancelledAt)
	assert.False(t, saga.IsInProgress())
	assert.Len(t, saga.GetCompletedSteps(), 1) // Only payment completed
}

func TestFulfillmentSaga_MultipleInventoryItems(t *testing.T) {
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.StartPaymentProcessing()
	saga.CompletePayment(uuidv7.New())

	items := []ReservedItem{
		{ProductID: uuidv7.New(), Quantity: 5, ReservationID: uuidv7.New()},
		{ProductID: uuidv7.New(), Quantity: 10, ReservationID: uuidv7.New()},
		{ProductID: uuidv7.New(), Quantity: 2, ReservationID: uuidv7.New()},
	}
	saga.CompleteInventory(items)

	reservedItems := saga.GetReservedItems()
	assert.Len(t, reservedItems, 3)

	// Verify all items are preserved
	for i, item := range items {
		assert.Equal(t, item.ProductID, reservedItems[i].ProductID)
		assert.Equal(t, item.Quantity, reservedItems[i].Quantity)
		assert.Equal(t, item.ReservationID, reservedItems[i].ReservationID)
	}
}
