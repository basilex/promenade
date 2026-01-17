package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/scheduler"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type mockReceiptUseCase struct {
	ListReceiptsFunc func(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error)
	PrintReceiptFunc func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error)
}

func (m *mockReceiptUseCase) CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType receipt.PaymentType, receiptType receipt.ReceiptType, currency string, lines []receipt.ReceiptLine, createdBy uuidv7.UUID) (*receipt.Receipt, error) {
	return nil, nil
}

func (m *mockReceiptUseCase) GetReceipt(ctx context.Context, id uuidv7.UUID) (*receipt.Receipt, error) {
	return nil, nil
}

func (m *mockReceiptUseCase) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*receipt.Receipt, error) {
	return nil, nil
}

func (m *mockReceiptUseCase) ListReceipts(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error) {
	if m.ListReceiptsFunc != nil {
		return m.ListReceiptsFunc(ctx, filters)
	}
	return nil, nil
}

func (m *mockReceiptUseCase) MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
	return nil, nil
}

func (m *mockReceiptUseCase) PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
	if m.PrintReceiptFunc != nil {
		return m.PrintReceiptFunc(ctx, id, printedBy)
	}
	return nil, nil
}

func (m *mockReceiptUseCase) CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receipt.Receipt, error) {
	return nil, nil
}

func (m *mockReceiptUseCase) DeleteReceipt(ctx context.Context, id uuidv7.UUID) error {
	return nil
}

func TestReceiptRetryExecutor_Execute_NoPendingReceipts(t *testing.T) {
	listCalled := false
	printCalled := false

	uc := &mockReceiptUseCase{
		ListReceiptsFunc: func(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error) {
			listCalled = true
			return []*receipt.Receipt{}, nil
		},
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
			printCalled = true
			return nil, nil
		},
	}

	executor := NewReceiptRetryExecutor(uc)
	err := executor.Execute(context.Background(), &scheduler.Job{Name: receiptRetryJobName})
	require.NoError(t, err)
	assert.True(t, listCalled)
	assert.False(t, printCalled)
}

func TestReceiptRetryExecutor_Execute_PrintsPendingReceipts(t *testing.T) {
	printCount := 0

	rec1 := &receipt.Receipt{BaseAggregate: aggregate.NewBaseAggregateWithID(uuidv7.New()), LastUpdatedBy: uuidv7.New()}
	rec2 := &receipt.Receipt{BaseAggregate: aggregate.NewBaseAggregateWithID(uuidv7.New()), LastUpdatedBy: uuidv7.New()}

	uc := &mockReceiptUseCase{
		ListReceiptsFunc: func(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error) {
			return []*receipt.Receipt{rec1, rec2}, nil
		},
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
			printCount++
			return &receipt.Receipt{BaseAggregate: aggregate.NewBaseAggregateWithID(id)}, nil
		},
	}

	executor := NewReceiptRetryExecutor(uc)
	err := executor.Execute(context.Background(), &scheduler.Job{Name: receiptRetryJobName})
	require.NoError(t, err)
	assert.Equal(t, 2, printCount)
}

func TestReceiptRetryExecutor_Execute_ListError(t *testing.T) {
	expectedErr := errors.New("list error")
	uc := &mockReceiptUseCase{
		ListReceiptsFunc: func(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error) {
			return nil, expectedErr
		},
	}

	executor := NewReceiptRetryExecutor(uc)
	err := executor.Execute(context.Background(), &scheduler.Job{Name: receiptRetryJobName})
	require.ErrorIs(t, err, expectedErr)
}

func TestRegisterReceiptRetryJob_DefaultCron(t *testing.T) {
	cfg := scheduler.DefaultConfig()
	engine, err := scheduler.NewEngine(cfg)
	require.NoError(t, err)
	require.NoError(t, engine.Start())
	defer func() {
		_ = engine.Stop()
	}()

	uc := &mockReceiptUseCase{}
	require.NoError(t, RegisterReceiptRetryJob(engine, uc, ""))

	jobs := engine.ListJobs()
	require.Len(t, jobs, 1)
	assert.Equal(t, receiptRetryJobName, jobs[0].Name)
	assert.Equal(t, defaultReceiptRetryCron, jobs[0].CronExpression)
}
