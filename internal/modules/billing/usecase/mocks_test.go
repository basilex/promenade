package usecase

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// mockPlanRepository implements IPlanRepository for testing
type mockPlanRepository struct {
	mock.Mock
}

func (m *mockPlanRepository) Create(ctx context.Context, plan *entity.Plan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *mockPlanRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Plan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Plan), args.Error(1)
}

func (m *mockPlanRepository) GetBySlug(ctx context.Context, slug string) (*entity.Plan, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Plan), args.Error(1)
}

func (m *mockPlanRepository) List(ctx context.Context, status *entity.PlanStatus, limit, offset int) ([]*entity.Plan, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Plan), args.Error(1)
}

func (m *mockPlanRepository) Count(ctx context.Context, status *entity.PlanStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *mockPlanRepository) Update(ctx context.Context, plan *entity.Plan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *mockPlanRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mockSubscriptionRepository implements ISubscriptionRepository for testing
type mockSubscriptionRepository struct {
	mock.Mock
}

func (m *mockSubscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *mockSubscriptionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*entity.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepository) GetExpiring(ctx context.Context, days, limit int) ([]*entity.Subscription, error) {
	args := m.Called(ctx, days, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepository) List(ctx context.Context, status *entity.SubscriptionStatus, limit, offset int) ([]*entity.Subscription, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Subscription), args.Error(1)
}

func (m *mockSubscriptionRepository) Count(ctx context.Context, status *entity.SubscriptionStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *mockSubscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *mockSubscriptionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mockInvoiceRepository implements IInvoiceRepository for testing
type mockInvoiceRepository struct {
	mock.Mock
}

func (m *mockInvoiceRepository) Create(ctx context.Context, invoice *entity.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *mockInvoiceRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Invoice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entity.Invoice, error) {
	args := m.Called(ctx, invoiceNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) GetBySubscriptionID(ctx context.Context, subscriptionID uuidv7.UUID) ([]*entity.Invoice, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) GetOverdue(ctx context.Context, limit int) ([]*entity.Invoice, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) List(ctx context.Context, status *entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) Count(ctx context.Context, status *entity.InvoiceStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *mockInvoiceRepository) Update(ctx context.Context, invoice *entity.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *mockInvoiceRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// mockPaymentRepository implements IPaymentRepository for testing
type mockPaymentRepository struct {
	mock.Mock
}

func (m *mockPaymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *mockPaymentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *mockPaymentRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *mockPaymentRepository) GetByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]*entity.Payment, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Payment), args.Error(1)
}

func (m *mockPaymentRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*entity.Payment, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Payment), args.Error(1)
}

func (m *mockPaymentRepository) List(ctx context.Context, status *entity.PaymentStatus, limit, offset int) ([]*entity.Payment, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Payment), args.Error(1)
}

func (m *mockPaymentRepository) Count(ctx context.Context, status *entity.PaymentStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *mockPaymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
