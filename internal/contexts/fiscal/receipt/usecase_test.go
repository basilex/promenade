package receipt

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

var testReceiptUserID = uuidv7.New()
var testReceiptOrderID = uuidv7.New()
var testReceiptCashRegisterID = uuidv7.New()

// MockRepository implements IRepository for testing
type MockRepository struct {
	CreateFunc     func(ctx context.Context, rec *Receipt) error
	GetByIDFunc    func(ctx context.Context, id uuidv7.UUID) (*Receipt, error)
	GetByOrderFunc func(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error)
	ListFunc       func(ctx context.Context, filters *ListFilters) ([]*Receipt, error)
	UpdateFunc     func(ctx context.Context, rec *Receipt) error
	DeleteFunc     func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockRepository) Create(ctx context.Context, rec *Receipt) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, rec)
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Receipt, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, ErrReceiptNotFound
}

func (m *MockRepository) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error) {
	if m.GetByOrderFunc != nil {
		return m.GetByOrderFunc(ctx, orderID)
	}
	return nil, ErrReceiptNotFound
}

func (m *MockRepository) List(ctx context.Context, filters *ListFilters) ([]*Receipt, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters)
	}
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, rec *Receipt) error {
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
	uc := NewUseCase(&MockRepository{})
	assert.NotNil(t, uc)
}

func TestCreateReceipt_Success(t *testing.T) {
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error) {
			return nil, ErrReceiptNotFound
		},
		CreateFunc: func(ctx context.Context, rec *Receipt) error {
			assert.Equal(t, testReceiptCashRegisterID, rec.CashRegisterID)
			assert.Equal(t, testReceiptOrderID, rec.OrderID)
			return nil
		},
	}

	uc := NewUseCase(repo)
	rec, err := uc.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		PaymentTypeCash,
		ReceiptTypeSale,
		"UAH",
		[]ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	require.NoError(t, err)
	assert.NotNil(t, rec)
}

func TestCreateReceipt_Duplicate(t *testing.T) {
	existing := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error) {
			return existing, nil
		},
	}

	uc := NewUseCase(repo)
	rec, err := uc.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		PaymentTypeCash,
		ReceiptTypeSale,
		"UAH",
		[]ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	assert.ErrorIs(t, err, ErrReceiptAlreadyExists)
	assert.Nil(t, rec)
}

func TestCreateReceipt_RepoFailure(t *testing.T) {
	repo := &MockRepository{
		GetByOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error) {
			return nil, errors.New("db error")
		},
	}

	uc := NewUseCase(repo)
	rec, err := uc.CreateReceipt(
		context.Background(),
		testReceiptCashRegisterID,
		testReceiptOrderID,
		PaymentTypeCash,
		ReceiptTypeSale,
		"UAH",
		[]ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)

	assert.ErrorIs(t, err, ErrReceiptCreateFailed)
	assert.Nil(t, rec)
}

func TestMarkPrinted_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *Receipt) error {
			return nil
		},
	}

	uc := NewUseCase(repo)
	updated, err := uc.MarkPrinted(context.Background(), rec.GetID(), "FN-001", "url", "qr", testReceiptUserID)

	require.NoError(t, err)
	assert.Equal(t, ReceiptStatusPrinted, updated.Status)
}

func TestCancelReceipt_Success(t *testing.T) {
	rec := buildReceiptEntity(t)
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Receipt, error) {
			return rec, nil
		},
		UpdateFunc: func(ctx context.Context, rec *Receipt) error {
			return nil
		},
	}

	uc := NewUseCase(repo)
	updated, err := uc.CancelReceipt(context.Background(), rec.GetID(), "Customer request", testReceiptUserID)

	require.NoError(t, err)
	assert.Equal(t, ReceiptStatusCancelled, updated.Status)
}

func buildReceiptEntity(t *testing.T) *Receipt {
	rec, err := NewReceipt(
		testReceiptCashRegisterID,
		testReceiptOrderID,
		PaymentTypeCash,
		ReceiptTypeSale,
		"UAH",
		[]ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		testReceiptUserID,
	)
	require.NoError(t, err)
	return rec
}
