package integration

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cashregisteraggregate "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/repository"
	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	receiptaggregate "github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	receiptrepo "github.com/basilex/promenade/internal/contexts/fiscal/receipt/repository"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type testEvent struct {
	eventType   string
	aggregateID uuidv7.UUID
	occurredAt  time.Time
	metadata    map[string]string
}

func (e *testEvent) Type() string                { return e.eventType }
func (e *testEvent) OccurredAt() time.Time       { return e.occurredAt }
func (e *testEvent) AggregateID() uuidv7.UUID    { return e.aggregateID }
func (e *testEvent) Metadata() map[string]string { return e.metadata }

func createOrderConfirmedEvent(orderID, customerID, confirmedBy uuidv7.UUID, currency string, items []OrderItem) bus.Event {
	payload := OrderConfirmedEvent{
		OrderID:     orderID,
		CustomerID:  customerID,
		Items:       items,
		Currency:    currency,
		ConfirmedBy: confirmedBy,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &testEvent{
		eventType:   bus.TopicOrderConfirmed,
		aggregateID: orderID,
		occurredAt:  time.Now(),
		metadata: map[string]string{
			"payload": string(payloadJSON),
		},
	}
}

type MockReceiptUseCase struct {
	CreateReceiptFunc func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType receiptaggregate.PaymentType, receiptType receiptaggregate.ReceiptType, currency string, lines []receiptaggregate.ReceiptLine, createdBy uuidv7.UUID) (*receiptaggregate.Receipt, error)
	PrintReceiptFunc  func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receiptaggregate.Receipt, error)
}

func (m *MockReceiptUseCase) CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType receiptaggregate.PaymentType, receiptType receiptaggregate.ReceiptType, currency string, lines []receiptaggregate.ReceiptLine, createdBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
	if m.CreateReceiptFunc != nil {
		return m.CreateReceiptFunc(ctx, cashRegisterID, orderID, paymentType, receiptType, currency, lines, createdBy)
	}
	return nil, nil
}

func (m *MockReceiptUseCase) PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
	if m.PrintReceiptFunc != nil {
		return m.PrintReceiptFunc(ctx, id, printedBy)
	}
	return nil, nil
}

func (m *MockReceiptUseCase) GetReceipt(ctx context.Context, id uuidv7.UUID) (*receiptaggregate.Receipt, error) {
	return nil, nil
}

func (m *MockReceiptUseCase) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*receiptaggregate.Receipt, error) {
	return nil, nil
}

func (m *MockReceiptUseCase) ListReceipts(ctx context.Context, filters *receiptrepo.ListFilters) ([]*receiptaggregate.Receipt, error) {
	return nil, nil
}

func (m *MockReceiptUseCase) MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
	return nil, nil
}

func (m *MockReceiptUseCase) CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
	return nil, nil
}

func (m *MockReceiptUseCase) DeleteReceipt(ctx context.Context, id uuidv7.UUID) error {
	return nil
}

type MockCashRegisterRepo struct {
	ListActiveFunc func(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error)
}

func (m *MockCashRegisterRepo) Create(ctx context.Context, cr *cashregisteraggregate.CashRegister) error {
	return nil
}

func (m *MockCashRegisterRepo) GetByID(ctx context.Context, id uuidv7.UUID) (*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *MockCashRegisterRepo) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *MockCashRegisterRepo) GetByLocation(ctx context.Context, locationID uuidv7.UUID) ([]*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *MockCashRegisterRepo) ListActive(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
	if m.ListActiveFunc != nil {
		return m.ListActiveFunc(ctx)
	}
	return nil, nil
}

func (m *MockCashRegisterRepo) List(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *MockCashRegisterRepo) Update(ctx context.Context, cr *cashregisteraggregate.CashRegister) error {
	return nil
}

func (m *MockCashRegisterRepo) Delete(ctx context.Context, id uuidv7.UUID) error {
	return nil
}

type mockEventBus struct {
	SubscribeFunc func(topic string, handler bus.Handler) error
}

func (m *mockEventBus) Publish(ctx context.Context, topic string, event bus.Event) error { return nil }
func (m *mockEventBus) Subscribe(topic string, handler bus.Handler) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(topic, handler)
	}
	return nil
}
func (m *mockEventBus) Unsubscribe(topic string, handler bus.Handler) error { return nil }
func (m *mockEventBus) Close(ctx context.Context) error                     { return nil }
func (m *mockEventBus) Health(ctx context.Context) error                    { return nil }

func TestOrderEventHandler_HandleOrderConfirmed_Success_PrinterDisabled(t *testing.T) {
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	confirmedBy := uuidv7.New()
	currency := "UAH"
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 2, UnitPrice: 1500}}

	crID := uuidv7.New()
	activeCR := &cashregisteraggregate.CashRegister{
		BaseAggregate: aggregate.NewBaseAggregateWithID(crID),
		Status:        cashregisteraggregate.StatusActive,
	}

	createCalled := false
	printCalled := false

	mockReceiptUC := &MockReceiptUseCase{
		CreateReceiptFunc: func(ctx context.Context, cashRegisterID, oid uuidv7.UUID, paymentType receiptaggregate.PaymentType, receiptType receiptaggregate.ReceiptType, cur string, lines []receiptaggregate.ReceiptLine, createdBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			createCalled = true
			assert.Equal(t, crID, cashRegisterID)
			assert.Equal(t, orderID, oid)
			assert.Equal(t, receiptaggregate.PaymentTypeCard, paymentType)
			assert.Equal(t, receiptaggregate.ReceiptTypeSale, receiptType)
			assert.Equal(t, currency, cur)
			assert.Equal(t, confirmedBy, createdBy)
			require.Len(t, lines, 1)
			assert.Equal(t, "SKU-001", lines[0].Name)
			assert.Equal(t, 2, lines[0].Quantity)
			assert.Equal(t, int64(1500), lines[0].PriceCents)
			return &receiptaggregate.Receipt{BaseAggregate: aggregate.NewBaseAggregate()}, nil
		},
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			printCalled = true
			return &receiptaggregate.Receipt{BaseAggregate: aggregate.NewBaseAggregateWithID(id)}, nil
		},
	}

	mockCRRepo := &MockCashRegisterRepo{
		ListActiveFunc: func(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{activeCR}, nil
		},
	}

	handler := NewOrderEventHandler(mockReceiptUC, mockCRRepo, false)
	event := createOrderConfirmedEvent(orderID, customerID, confirmedBy, currency, items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.NoError(t, err)
	assert.True(t, createCalled)
	assert.False(t, printCalled)
}

func TestOrderEventHandler_HandleOrderConfirmed_Success_PrinterEnabled(t *testing.T) {
	orderID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "", Quantity: 1, UnitPrice: 1000}}

	crID := uuidv7.New()
	activeCR := &cashregisteraggregate.CashRegister{BaseAggregate: aggregate.NewBaseAggregateWithID(crID)}

	createCalled := false
	printCalled := false

	mockReceiptUC := &MockReceiptUseCase{
		CreateReceiptFunc: func(ctx context.Context, cashRegisterID, oid uuidv7.UUID, paymentType receiptaggregate.PaymentType, receiptType receiptaggregate.ReceiptType, cur string, lines []receiptaggregate.ReceiptLine, createdBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			createCalled = true
			require.Len(t, lines, 1)
			assert.Contains(t, lines[0].Name, items[0].ProductID.String())
			return &receiptaggregate.Receipt{BaseAggregate: aggregate.NewBaseAggregate()}, nil
		},
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			printCalled = true
			assert.Equal(t, confirmedBy, printedBy)
			return &receiptaggregate.Receipt{BaseAggregate: aggregate.NewBaseAggregateWithID(id)}, nil
		},
	}

	mockCRRepo := &MockCashRegisterRepo{
		ListActiveFunc: func(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{activeCR}, nil
		},
	}

	handler := NewOrderEventHandler(mockReceiptUC, mockCRRepo, true)
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.NoError(t, err)
	assert.True(t, createCalled)
	assert.True(t, printCalled)
}

func TestOrderEventHandler_HandleOrderConfirmed_ReceiptAlreadyExists(t *testing.T) {
	orderID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 1, UnitPrice: 1000}}

	mockReceiptUC := &MockReceiptUseCase{
		CreateReceiptFunc: func(ctx context.Context, cashRegisterID, oid uuidv7.UUID, paymentType receiptaggregate.PaymentType, receiptType receiptaggregate.ReceiptType, cur string, lines []receiptaggregate.ReceiptLine, createdBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptAlreadyExists
		},
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			return nil, errors.New("should not print")
		},
	}

	mockCRRepo := &MockCashRegisterRepo{
		ListActiveFunc: func(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{{BaseAggregate: aggregate.NewBaseAggregate()}}, nil
		},
	}

	handler := NewOrderEventHandler(mockReceiptUC, mockCRRepo, true)
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.NoError(t, err)
}

func TestOrderEventHandler_HandleOrderConfirmed_PrintReceiptError(t *testing.T) {
	orderID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 1, UnitPrice: 1000}}

	mockReceiptUC := &MockReceiptUseCase{
		CreateReceiptFunc: func(ctx context.Context, cashRegisterID, oid uuidv7.UUID, paymentType receiptaggregate.PaymentType, receiptType receiptaggregate.ReceiptType, cur string, lines []receiptaggregate.ReceiptLine, createdBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			return &receiptaggregate.Receipt{BaseAggregate: aggregate.NewBaseAggregate()}, nil
		},
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receiptaggregate.Receipt, error) {
			return nil, errors.New("print failed")
		},
	}

	mockCRRepo := &MockCashRegisterRepo{
		ListActiveFunc: func(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{{BaseAggregate: aggregate.NewBaseAggregate()}}, nil
		},
	}

	handler := NewOrderEventHandler(mockReceiptUC, mockCRRepo, true)
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.Error(t, err)
}

func TestOrderEventHandler_HandleOrderConfirmed_NoCashRegisters(t *testing.T) {
	orderID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 1, UnitPrice: 1000}}

	mockReceiptUC := &MockReceiptUseCase{}
	mockCRRepo := &MockCashRegisterRepo{
		ListActiveFunc: func(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{}, nil
		},
	}

	handler := NewOrderEventHandler(mockReceiptUC, mockCRRepo, true)
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.Error(t, err)
}

func TestOrderEventHandler_HandleOrderConfirmed_ListActiveError(t *testing.T) {
	orderID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 1, UnitPrice: 1000}}

	mockReceiptUC := &MockReceiptUseCase{}
	mockCRRepo := &MockCashRegisterRepo{
		ListActiveFunc: func(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
			return nil, errors.New("db error")
		},
	}

	handler := NewOrderEventHandler(mockReceiptUC, mockCRRepo, true)
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.Error(t, err)
}

func TestOrderEventHandler_HandleOrderConfirmed_InvalidPayload(t *testing.T) {
	orderID := uuidv7.Nil
	confirmedBy := uuidv7.New()
	items := []OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 1, UnitPrice: 1000}}

	handler := NewOrderEventHandler(&MockReceiptUseCase{}, &MockCashRegisterRepo{}, false)
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	err := handler.HandleOrderConfirmed(context.Background(), event)
	require.Error(t, err)
}

func TestOrderEventHandler_parseEventMetadata_NilMetadata(t *testing.T) {
	handler := NewOrderEventHandler(&MockReceiptUseCase{}, &MockCashRegisterRepo{}, false)
	var target OrderConfirmedEvent

	err := handler.parseEventMetadata(&testEvent{metadata: nil}, &target)
	require.Error(t, err)
}

func TestOrderEventHandler_parseEventMetadata_MissingPayload(t *testing.T) {
	handler := NewOrderEventHandler(&MockReceiptUseCase{}, &MockCashRegisterRepo{}, false)
	var target OrderConfirmedEvent

	err := handler.parseEventMetadata(&testEvent{metadata: map[string]string{}}, &target)
	require.Error(t, err)
}

func TestOrderEventHandler_parseEventMetadata_InvalidJSON(t *testing.T) {
	handler := NewOrderEventHandler(&MockReceiptUseCase{}, &MockCashRegisterRepo{}, false)
	var target OrderConfirmedEvent

	err := handler.parseEventMetadata(&testEvent{metadata: map[string]string{"payload": "{"}}, &target)
	require.Error(t, err)
}

func TestOrderEventHandler_RegisterHandlers(t *testing.T) {
	handler := NewOrderEventHandler(&MockReceiptUseCase{}, &MockCashRegisterRepo{}, false)

	subscribeCalls := make(map[string]int)
	mockBus := &mockEventBus{
		SubscribeFunc: func(topic string, h bus.Handler) error {
			subscribeCalls[topic]++
			return nil
		},
	}

	err := handler.RegisterHandlers(mockBus)
	require.NoError(t, err)
	assert.Equal(t, 1, subscribeCalls[bus.TopicOrderConfirmed])
}
