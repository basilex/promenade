package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var testReceiptUserID = uuidv7.New()
var testReceiptOrderID = uuidv7.New()
var testReceiptCashRegisterID = uuidv7.New()

// MockRepository implements IReceiptRepository for testing
type MockRepository struct {
	CreateFunc     func(ctx context.Context, rec *aggregate.Receipt) error
	GetByIDFunc    func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error)
	GetByOrderFunc func(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error)
	ListFunc       func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error)
	UpdateFunc     func(ctx context.Context, rec *aggregate.Receipt) error
	DeleteFunc     func(ctx context.Context, id uuidv7.UUID) error
}

// MockPrinter implements IPrinter for testing
type MockPrinter struct {
	PrintFunc  func(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error)
	CancelFunc func(ctx context.Context, rec *aggregate.Receipt, reason string) error
}

func (m *MockPrinter) Print(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error) {
	if m.PrintFunc != nil {
		return m.PrintFunc(ctx, rec)
	}
	return &PrintResult{FiscalNumber: "FN-TEST", FiscalURL: "url", QRCode: "qr"}, nil
}

func (m *MockPrinter) Cancel(ctx context.Context, rec *aggregate.Receipt, reason string) error {
	if m.CancelFunc != nil {
		return m.CancelFunc(ctx, rec, reason)
	}
	return nil
}

func (m *MockRepository) Create(ctx context.Context, rec *aggregate.Receipt) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, rec)
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, receipterrors.ErrReceiptNotFound
}

func (m *MockRepository) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
	if m.GetByOrderFunc != nil {
		return m.GetByOrderFunc(ctx, orderID)
	}
	return nil, receipterrors.ErrReceiptNotFound
}

func (m *MockRepository) List(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters)
	}
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, rec *aggregate.Receipt) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, rec)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func TestNewUseCase(t *testing.T) {
	u := NewReceiptUseCase(&MockRepository{}, nil)
	assert.NotNil(t, u)
}

func TestCreateReceipt_Success(t *testing.T) {
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
		CreateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			assert.Equal(t, testReceiptCashRegisterID, rec.CashRegisterID)
			assert.Equal(t, testReceiptOrderID, rec.OrderID)
			return nil
		},
	}

	u := NewReceiptUseCase(repo, nil)
	rec, err := u.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	require.NoError(t, err)
	assert.NotNil(t, rec)
}

func TestCreateReceipt_Duplicate(t *testing.T) {
	existing := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
			return existing, nil
		},
	}

	u := NewReceiptUseCase(repo, nil)
	rec, err := u.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptAlreadyExists)
	assert.Nil(t, rec)
}

func TestCreateReceipt_RepoFailure(t *testing.T) {
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, errors.New("db error")
		},
	}

	u := NewReceiptUseCase(repo, nil)
	rec, err := u.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	assert.Nil(t, rec)
}

func TestCreateReceipt_CreateFailed(t *testing.T) {
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
		CreateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return errors.New("create error")
		},
	}

	u := NewReceiptUseCase(repo, nil)

	created, err := u.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	assert.Nil(t, created)
}

func TestCreateReceipt_GetByOrderError(t *testing.T) {
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, errors.New("db error")
		},
	}

	u := NewReceiptUseCase(repo, nil)

	created, err := u.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	assert.Nil(t, created)
}

func TestListReceipts_Error(t *testing.T) {
	repo := &MockRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
			return nil, errors.New("list error")
		},
	}

	u := NewReceiptUseCase(repo, nil)

	list, err := u.ListReceipts(context.Background(), &repository.ListFilters{})

	assert.ErrorIs(t, err, receipterrors.ErrReceiptListFailed)
	assert.Nil(t, list)
}

func TestListReceipts_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
			return []*aggregate.Receipt{rec}, nil
		},
	}

	u := NewReceiptUseCase(repo, nil)

	list, err := u.ListReceipts(context.Background(), &repository.ListFilters{})

	assert.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestMarkPrinted_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return nil
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.MarkPrinted(context.Background(), rec.GetID(), "FN-001", "url", "qr", testReceiptUserID)

	require.NoError(t, err)
	assert.Equal(t, aggregate.ReceiptStatusPrinted, updated.Status)
}

func TestMarkPrinted_UpdateFailure(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return errors.New("update failed")
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.MarkPrinted(context.Background(), rec.GetID(), "FN-001", "url", "qr", testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptUpdateFailed)
	assert.Nil(t, updated)
}

func TestMarkPrinted_InvalidFiscalNumber(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.MarkPrinted(context.Background(), rec.GetID(), "", "", "", testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrFiscalNumberRequired)
	assert.Nil(t, updated)
}

func TestMarkPrinted_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.MarkPrinted(context.Background(), uuidv7.New(), "FN", "url", "qr", testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	assert.Nil(t, updated)
}

func TestCancelReceipt_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return nil
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.CancelReceipt(context.Background(), rec.GetID(), "Customer request", testReceiptUserID)

	require.NoError(t, err)
	assert.Equal(t, aggregate.ReceiptStatusCancelled, updated.Status)
}

func TestCancelReceipt_Printed_ProviderCancelSuccess(t *testing.T) {
	rec := buildReceiptEntity(t)
	rec.Status = aggregate.ReceiptStatusPrinted
	rec.ProviderReceiptID = "provider-1"

	printer := &MockPrinter{
		CancelFunc: func(ctx context.Context, rec *aggregate.Receipt, reason string) error {
			return nil
		},
	}
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return nil
		},
	}

	u := NewReceiptUseCase(repo, printer)

	updated, err := u.CancelReceipt(context.Background(), rec.GetID(), "Customer request", testReceiptUserID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.ReceiptStatusCancelled, updated.Status)
}

func TestCancelReceipt_Printed_ProviderCancelError(t *testing.T) {
	rec := buildReceiptEntity(t)
	rec.Status = aggregate.ReceiptStatusPrinted
	rec.ProviderReceiptID = "provider-1"

	printer := &MockPrinter{
		CancelFunc: func(ctx context.Context, rec *aggregate.Receipt, reason string) error {
			return errors.New("cancel failed")
		},
	}
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
	}

	u := NewReceiptUseCase(repo, printer)

	updated, err := u.CancelReceipt(context.Background(), rec.GetID(), "Customer request", testReceiptUserID)
	assert.Nil(t, updated)
	assert.ErrorIs(t, err, receipterrors.ErrReceiptCancelFailed)
}

func TestCancelReceipt_UpdateFailure(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return errors.New("update failed")
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.CancelReceipt(context.Background(), rec.GetID(), "Customer request", testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptUpdateFailed)
	assert.Nil(t, updated)
}

func TestPrintReceipt_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return nil
		},
	}
	printer := &MockPrinter{
		PrintFunc: func(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error) {
			return &PrintResult{FiscalNumber: "FN-PRINT", FiscalURL: "url", QRCode: "qr"}, nil
		},
	}

	u := NewReceiptUseCase(repo, printer)
	updated, err := u.PrintReceipt(context.Background(), rec.GetID(), testReceiptUserID)

	require.NoError(t, err)
	assert.Equal(t, aggregate.ReceiptStatusPrinted, updated.Status)
	assert.Equal(t, "FN-PRINT", updated.FiscalNumber)
}

func TestPrintReceipt_PrinterError(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
	}
	printer := &MockPrinter{
		PrintFunc: func(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error) {
			return nil, errors.New("print failed")
		},
	}

	u := NewReceiptUseCase(repo, printer)
	updated, err := u.PrintReceipt(context.Background(), rec.GetID(), testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptPrintFailed)
	assert.Nil(t, updated)
}

func TestPrintReceipt_UpdateFailure(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *aggregate.Receipt) error {
			return errors.New("update failed")
		},
	}
	printer := &MockPrinter{
		PrintFunc: func(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error) {
			return &PrintResult{FiscalNumber: "FN-PRINT", FiscalURL: "url", QRCode: "qr"}, nil
		},
	}

	u := NewReceiptUseCase(repo, printer)
	updated, err := u.PrintReceipt(context.Background(), rec.GetID(), testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptUpdateFailed)
	assert.Nil(t, updated)
}

func TestPrintReceipt_NoPrinter(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.PrintReceipt(context.Background(), rec.GetID(), testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptPrintFailed)
	assert.Nil(t, updated)
}

func TestPrintReceipt_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	u := NewReceiptUseCase(repo, nil)
	updated, err := u.PrintReceipt(context.Background(), uuidv7.New(), testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	assert.Nil(t, updated)
}

func TestPrintReceipt_MarkPrintedError(t *testing.T) {
	rec := buildReceiptEntity(t)
	rec.Status = aggregate.ReceiptStatusCancelled

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
	}
	printer := &MockPrinter{
		PrintFunc: func(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error) {
			return &PrintResult{FiscalNumber: "FN-PRINT", FiscalURL: "url", QRCode: "qr"}, nil
		},
	}

	u := NewReceiptUseCase(repo, printer)
	updated, err := u.PrintReceipt(context.Background(), rec.GetID(), testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptAlreadyCancelled)
	assert.Nil(t, updated)
}

func TestDeleteReceipt_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	u := NewReceiptUseCase(repo, nil)

	err := u.DeleteReceipt(context.Background(), uuidv7.New())

	assert.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
}

func TestGetReceipt_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
	}

	u := NewReceiptUseCase(repo, nil)

	found, err := u.GetReceipt(context.Background(), rec.GetID())

	assert.NoError(t, err)
	assert.Equal(t, rec, found)
}

func TestGetReceipt_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	u := NewReceiptUseCase(repo, nil)

	found, err := u.GetReceipt(context.Background(), uuidv7.New())

	assert.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	assert.Nil(t, found)
}

func TestCancelReceipt_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	u := NewReceiptUseCase(repo, nil)

	updated, err := u.CancelReceipt(context.Background(), uuidv7.New(), "reason", testReceiptUserID)

	assert.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	assert.Nil(t, updated)
}

func TestDeleteReceipt_DeleteFailed(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return errors.New("delete failed")
		},
	}

	u := NewReceiptUseCase(repo, nil)

	err := u.DeleteReceipt(context.Background(), rec.GetID())

	assert.ErrorIs(t, err, receipterrors.ErrReceiptDeleteFailed)
}

func TestGetByOrderID_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
			return rec, nil
		},
	}

	u := NewReceiptUseCase(repo, nil)

	found, err := u.GetByOrderID(context.Background(), rec.OrderID)

	assert.NoError(t, err)
	assert.Equal(t, rec, found)
}

func buildReceiptEntity(t *testing.T) *aggregate.Receipt {
	rec, err := aggregate.NewReceipt(
		testReceiptCashRegisterID,
		testReceiptOrderID,
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)
	require.NoError(t, err)
	return rec
}
