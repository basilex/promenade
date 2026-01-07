package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockReservationService implements IReservationService for testing
type MockReservationService struct {
	ReserveForOrderFunc func(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, reservedBy uuidv7.UUID) error
	ReleaseForOrderFunc func(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, releasedBy uuidv7.UUID) error
	CommitForOrderFunc  func(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, committedBy uuidv7.UUID) error
}

func (m *MockReservationService) ReserveForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, reservedBy uuidv7.UUID) error {
	if m.ReserveForOrderFunc != nil {
		return m.ReserveForOrderFunc(ctx, orderID, items, reservedBy)
	}
	return nil
}

func (m *MockReservationService) ReleaseForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, releasedBy uuidv7.UUID) error {
	if m.ReleaseForOrderFunc != nil {
		return m.ReleaseForOrderFunc(ctx, orderID, items, releasedBy)
	}
	return nil
}

func (m *MockReservationService) CommitForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, committedBy uuidv7.UUID) error {
	if m.CommitForOrderFunc != nil {
		return m.CommitForOrderFunc(ctx, orderID, items, committedBy)
	}
	return nil
}

// createOrderConfirmedEvent creates a test order.confirmed event
func createOrderConfirmedEvent(orderID, customerID, confirmedBy uuidv7.UUID, items []OrderItem) bus.Event {
	payload := OrderConfirmedEvent{
		OrderID:     orderID,
		CustomerID:  customerID,
		Items:       items,
		ConfirmedBy: confirmedBy,
	}
	payloadJSON, _ := json.Marshal(payload)
	
	metadata := map[string]string{
		"payload": string(payloadJSON),
	}
	
	return &testEvent{
		eventType:   "order.confirmed",
		aggregateID: orderID,
		occurredAt:  time.Now(),
		metadata:    metadata,
	}
}

// createOrderCancelledEvent creates a test order.cancelled event
func createOrderCancelledEvent(orderID, cancelledBy uuidv7.UUID, items []OrderItem, reason string) bus.Event {
	payload := OrderCancelledEvent{
		OrderID:     orderID,
		Items:       items,
		Reason:      reason,
		CancelledBy: cancelledBy,
	}
	payloadJSON, _ := json.Marshal(payload)
	
	metadata := map[string]string{
		"payload": string(payloadJSON),
	}
	
	return &testEvent{
		eventType:   "order.cancelled",
		aggregateID: orderID,
		occurredAt:  time.Now(),
		metadata:    metadata,
	}
}

// createOrderFulfilledEvent creates a test order.fulfilled event
func createOrderFulfilledEvent(orderID, fulfilledBy uuidv7.UUID, items []OrderItem) bus.Event {
	payload := OrderFulfilledEvent{
		OrderID:     orderID,
		Items:       items,
		FulfilledBy: fulfilledBy,
	}
	payloadJSON, _ := json.Marshal(payload)
	
	metadata := map[string]string{
		"payload": string(payloadJSON),
	}
	
	return &testEvent{
		eventType:   "order.fulfilled",
		aggregateID: orderID,
		occurredAt:  time.Now(),
		metadata:    metadata,
	}
}

// testEvent implements bus.Event interface for testing
type testEvent struct {
	eventType   string
	aggregateID uuidv7.UUID
	occurredAt  time.Time
	metadata    map[string]string
}

func (e *testEvent) Type() string {
	return e.eventType
}

func (e *testEvent) AggregateID() uuidv7.UUID {
	return e.aggregateID
}

func (e *testEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *testEvent) Metadata() map[string]string {
	return e.metadata
}

// Test HandleOrderConfirmed - Success
func TestOrderEventHandler_HandleOrderConfirmed_Success(t *testing.T) {
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}}

	reserveCalled := false
	mockService := &MockReservationService{
		ReserveForOrderFunc: func(ctx context.Context, oid uuidv7.UUID, itms []OrderItem, reservedBy uuidv7.UUID) error {
			reserveCalled = true
			assert.Equal(t, orderID, oid)
			assert.Equal(t, confirmedBy, reservedBy)
			assert.Len(t, itms, 1)
			return nil
		},
	}

	handler := NewOrderEventHandler(mockService)
	event := createOrderConfirmedEvent(orderID, customerID, confirmedBy, items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.NoError(t, err)
	assert.True(t, reserveCalled)
}

// Test HandleOrderConfirmed - Missing OrderID
func TestOrderEventHandler_HandleOrderConfirmed_MissingOrderID(t *testing.T) {
	mockService := &MockReservationService{}
	handler := NewOrderEventHandler(mockService)

	// Create event with empty order_id to trigger validation error
	payload := OrderConfirmedEvent{
		OrderID:     uuidv7.UUID{}, // Empty UUID
		CustomerID:  uuidv7.New(),
		ConfirmedBy: uuidv7.New(),
		Items:       []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}},
	}
	payloadJSON, _ := json.Marshal(payload)
	metadata := map[string]string{"payload": string(payloadJSON)}
	event := &testEvent{
		eventType:   "order.confirmed",
		aggregateID: uuidv7.New(),
		occurredAt:  time.Now(),
		metadata:    metadata,
	}

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing order_id")
}

// Test HandleOrderConfirmed - Empty Items
func TestOrderEventHandler_HandleOrderConfirmed_EmptyItems(t *testing.T) {
	mockService := &MockReservationService{}
	handler := NewOrderEventHandler(mockService)

	// Create event with empty items array to trigger validation error
	payload := OrderConfirmedEvent{
		OrderID:     uuidv7.New(),
		CustomerID:  uuidv7.New(),
		ConfirmedBy: uuidv7.New(),
		Items:       []OrderItem{}, // Empty items
	}
	payloadJSON, _ := json.Marshal(payload)
	metadata := map[string]string{"payload": string(payloadJSON)}
	event := &testEvent{
		eventType:   "order.confirmed",
		aggregateID: uuidv7.New(),
		occurredAt:  time.Now(),
		metadata:    metadata,
	}

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no items")
}

// Test HandleOrderConfirmed - Reservation Service Error
func TestOrderEventHandler_HandleOrderConfirmed_ServiceError(t *testing.T) {
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}}

	mockService := &MockReservationService{
		ReserveForOrderFunc: func(ctx context.Context, oid uuidv7.UUID, itms []OrderItem, reservedBy uuidv7.UUID) error {
			return assert.AnError
		},
	}

	handler := NewOrderEventHandler(mockService)
	event := createOrderConfirmedEvent(orderID, customerID, confirmedBy, items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reserve stock")
}

// Test HandleOrderCancelled - Success
func TestOrderEventHandler_HandleOrderCancelled_Success(t *testing.T) {
	orderID := uuidv7.New()
	cancelledBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}}

	releaseCalled := false
	mockService := &MockReservationService{
		ReleaseForOrderFunc: func(ctx context.Context, oid uuidv7.UUID, itms []OrderItem, releasedBy uuidv7.UUID) error {
			releaseCalled = true
			assert.Equal(t, orderID, oid)
			assert.Equal(t, cancelledBy, releasedBy)
			return nil
		},
	}

	handler := NewOrderEventHandler(mockService)
	event := createOrderCancelledEvent(orderID, cancelledBy, items, "Customer request")

	err := handler.HandleOrderCancelled(context.Background(), event)
	require.NoError(t, err)
	assert.True(t, releaseCalled)
}

// Test HandleOrderCancelled - Service Error (should not fail)
func TestOrderEventHandler_HandleOrderCancelled_ServiceError(t *testing.T) {
	orderID := uuidv7.New()
	cancelledBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}}

	mockService := &MockReservationService{
		ReleaseForOrderFunc: func(ctx context.Context, oid uuidv7.UUID, itms []OrderItem, releasedBy uuidv7.UUID) error {
			return assert.AnError
		},
	}

	handler := NewOrderEventHandler(mockService)
	event := createOrderCancelledEvent(orderID, cancelledBy, items, "Out of stock")

	err := handler.HandleOrderCancelled(context.Background(), event)
	require.NoError(t, err) // Should not fail - release is best effort
}

// Test HandleOrderCancelled - Missing OrderID
func TestOrderEventHandler_HandleOrderCancelled_MissingOrderID(t *testing.T) {
	mockService := &MockReservationService{}
	handler := NewOrderEventHandler(mockService)

	// Create event with empty order_id to trigger validation error
	payload := OrderCancelledEvent{
		OrderID:     uuidv7.UUID{}, // Empty UUID
		CancelledBy: uuidv7.New(),
		Reason:      "Test",
		Items:       []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}},
	}
	payloadJSON, _ := json.Marshal(payload)
	metadata := map[string]string{"payload": string(payloadJSON)}
	event := &testEvent{
		eventType:   "order.cancelled",
		aggregateID: uuidv7.New(),
		occurredAt:  time.Now(),
		metadata:    metadata,
	}

	err := handler.HandleOrderCancelled(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing order_id")
}

// Test HandleOrderFulfilled - Success
func TestOrderEventHandler_HandleOrderFulfilled_Success(t *testing.T) {
	orderID := uuidv7.New()
	fulfilledBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}}

	commitCalled := false
	mockService := &MockReservationService{
		CommitForOrderFunc: func(ctx context.Context, oid uuidv7.UUID, itms []OrderItem, committedBy uuidv7.UUID) error {
			commitCalled = true
			assert.Equal(t, orderID, oid)
			assert.Equal(t, fulfilledBy, committedBy)
			return nil
		},
	}

	handler := NewOrderEventHandler(mockService)
	event := createOrderFulfilledEvent(orderID, fulfilledBy, items)

	err := handler.HandleOrderFulfilled(context.Background(), event)
	require.NoError(t, err)
	assert.True(t, commitCalled)
}

// Test HandleOrderFulfilled - Service Error
func TestOrderEventHandler_HandleOrderFulfilled_ServiceError(t *testing.T) {
	orderID := uuidv7.New()
	fulfilledBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}}

	mockService := &MockReservationService{
		CommitForOrderFunc: func(ctx context.Context, oid uuidv7.UUID, itms []OrderItem, committedBy uuidv7.UUID) error {
			return assert.AnError
		},
	}

	handler := NewOrderEventHandler(mockService)
	event := createOrderFulfilledEvent(orderID, fulfilledBy, items)

	err := handler.HandleOrderFulfilled(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit stock")
}

// Test HandleOrderFulfilled - Missing OrderID
func TestOrderEventHandler_HandleOrderFulfilled_MissingOrderID(t *testing.T) {
	mockService := &MockReservationService{}
	handler := NewOrderEventHandler(mockService)

	// Create event with empty order_id to trigger validation error
	payload := OrderFulfilledEvent{
		OrderID:     uuidv7.UUID{}, // Empty UUID
		FulfilledBy: uuidv7.New(),
		Items:       []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 10}},
	}
	payloadJSON, _ := json.Marshal(payload)
	metadata := map[string]string{"payload": string(payloadJSON)}
	event := &testEvent{
		eventType:   "order.fulfilled",
		aggregateID: uuidv7.New(),
		occurredAt:  time.Now(),
		metadata:    metadata,
	}

	err := handler.HandleOrderFulfilled(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing order_id")
}

// Test HandleOrderFulfilled - Empty Items
func TestOrderEventHandler_HandleOrderFulfilled_EmptyItems(t *testing.T) {
	mockService := &MockReservationService{}
	handler := NewOrderEventHandler(mockService)

	// Create event with empty items array to trigger validation error
	payload := OrderFulfilledEvent{
		OrderID:     uuidv7.New(),
		FulfilledBy: uuidv7.New(),
		Items:       []OrderItem{}, // Empty items
	}
	payloadJSON, _ := json.Marshal(payload)
	metadata := map[string]string{"payload": string(payloadJSON)}
	event := &testEvent{
		eventType:   "order.fulfilled",
		aggregateID: uuidv7.New(),
		occurredAt:  time.Now(),
		metadata:    metadata,
	}

	err := handler.HandleOrderFulfilled(context.Background(), event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no items")
}

// Test parseEventMetadata - Nil Metadata
func TestOrderEventHandler_parseEventMetadata_NilMetadata(t *testing.T) {
	mockService := &MockReservationService{}
	handler := NewOrderEventHandler(mockService)

	event := &testEvent{
		eventType:   "test.event",
		aggregateID: uuidv7.New(),
		occurredAt:  time.Now(),
		metadata:    nil,
	}

	var target OrderConfirmedEvent
	err := handler.parseEventMetadata(event, &target)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "metadata is nil")
}

// Test RegisterHandlers - Success
func TestOrderEventHandler_RegisterHandlers(t *testing.T) {
	mockService := &MockReservationService{}
	handler := NewOrderEventHandler(mockService)

	// Create mock event bus
	subscribeCalls := make(map[string]int)
	mockBus := &mockEventBus{
		SubscribeFunc: func(topic string, h bus.Handler) error {
			subscribeCalls[topic]++
			return nil
		},
	}

	err := handler.RegisterHandlers(mockBus)
	require.NoError(t, err)
	assert.Equal(t, 1, subscribeCalls["order.confirmed"])
	assert.Equal(t, 1, subscribeCalls["order.cancelled"])
	assert.Equal(t, 1, subscribeCalls["order.fulfilled"])
}

// mockEventBus implements bus.IBus for testing
type mockEventBus struct {
	SubscribeFunc func(topic string, handler bus.Handler) error
}

func (m *mockEventBus) Publish(ctx context.Context, topic string, event bus.Event) error {
	return nil
}

func (m *mockEventBus) Subscribe(topic string, handler bus.Handler) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(topic, handler)
	}
	return nil
}

func (m *mockEventBus) Unsubscribe(topic string, handler bus.Handler) error {
	return nil
}

func (m *mockEventBus) Close(ctx context.Context) error {
	return nil
}

func (m *mockEventBus) Health(ctx context.Context) error {
	return nil
}
