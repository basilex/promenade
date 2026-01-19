package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	crRepo "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/repository/postgres"
	cashregisterAggregate "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/integration"
	receiptPrinter "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/printer"
	receiptRepo "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/repository/postgres"
	receiptAggregate "github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	receiptUseCase "github.com/basilex/promenade/internal/contexts/fiscal/receipt/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/bus/memory"
	"github.com/basilex/promenade/pkg/uuidv7"
	testutils "github.com/basilex/promenade/test/integration"
)

type testEvent struct {
	eventType   string
	aggregateID uuidv7.UUID
	occurredAt  time.Time
	metadata    map[string]string
}

func (e *testEvent) Type() string                { return e.eventType }
func (e *testEvent) AggregateID() uuidv7.UUID    { return e.aggregateID }
func (e *testEvent) OccurredAt() time.Time       { return e.occurredAt }
func (e *testEvent) Metadata() map[string]string { return e.metadata }

func createOrderConfirmedEvent(orderID, customerID, confirmedBy uuidv7.UUID, currency string, items []integration.OrderItem) bus.Event {
	payload := integration.OrderConfirmedEvent{
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

func TestFiscalIntegration_OrderConfirmed_CreatesReceipt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutils.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	crRepository := crRepo.NewCashRegisterRepository(db.DB)
	recRepository := receiptRepo.NewReceiptRepository(db.DB)
	receiptUC := receiptUseCase.NewReceiptUseCase(recRepository, nil)

	createdBy := uuidv7.New()
	cr, err := cashregisterAggregate.NewCashRegister(uuidv7.New(), "FN-001", "Model-X", createdBy)
	require.NoError(t, err)
	require.NoError(t, cr.Activate("LIC-001", createdBy))
	require.NoError(t, crRepository.Create(ctx, cr))

	eventBus := memory.NewMemoryBus(bus.NewConfig(10, 100, 3, 100*time.Millisecond, 1*time.Second, 2.0))
	defer func() {
		_ = eventBus.Close(ctx)
	}()

	handler := integration.NewOrderEventHandler(receiptUC, crRepository, false)
	require.NoError(t, handler.RegisterHandlers(eventBus))

	orderID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []integration.OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-001", Quantity: 2, UnitPrice: 1500}}
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	require.NoError(t, eventBus.Publish(ctx, bus.TopicOrderConfirmed, event))
	time.Sleep(200 * time.Millisecond)

	created, err := recRepository.GetByOrderID(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, receiptAggregate.ReceiptStatusPending, created.Status)
	require.Equal(t, "UAH", created.Currency)
	require.Len(t, created.Lines.Get(), 1)
}

func TestFiscalIntegration_OrderConfirmed_PrintsReceipt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutils.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	crRepository := crRepo.NewCashRegisterRepository(db.DB)
	recRepository := receiptRepo.NewReceiptRepository(db.DB)
	outputDir := t.TempDir()
	printer := receiptPrinter.NewPDFPrinter(outputDir)
	receiptUC := receiptUseCase.NewReceiptUseCase(recRepository, printer)

	createdBy := uuidv7.New()
	cr, err := cashregisterAggregate.NewCashRegister(uuidv7.New(), "FN-002", "Model-Y", createdBy)
	require.NoError(t, err)
	require.NoError(t, cr.Activate("LIC-002", createdBy))
	require.NoError(t, crRepository.Create(ctx, cr))

	eventBus := memory.NewMemoryBus(bus.NewConfig(10, 100, 3, 100*time.Millisecond, 1*time.Second, 2.0))
	defer func() {
		_ = eventBus.Close(ctx)
	}()

	handler := integration.NewOrderEventHandler(receiptUC, crRepository, true)
	require.NoError(t, handler.RegisterHandlers(eventBus))

	orderID := uuidv7.New()
	confirmedBy := uuidv7.New()
	items := []integration.OrderItem{{ProductID: uuidv7.New(), SKU: "SKU-002", Quantity: 1, UnitPrice: 1200}}
	event := createOrderConfirmedEvent(orderID, uuidv7.New(), confirmedBy, "UAH", items)

	require.NoError(t, eventBus.Publish(ctx, bus.TopicOrderConfirmed, event))
	time.Sleep(300 * time.Millisecond)

	created, err := recRepository.GetByOrderID(ctx, orderID)
	require.NoError(t, err)
	require.Equal(t, receiptAggregate.ReceiptStatusPrinted, created.Status)
	require.NotEmpty(t, created.FiscalURL)

	info, err := os.Stat(created.FiscalURL)
	require.NoError(t, err)
	require.False(t, info.IsDir())
}
