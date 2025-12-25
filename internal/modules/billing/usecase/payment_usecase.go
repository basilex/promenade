package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type IPaymentUseCase interface {
	CreatePayment(ctx context.Context, userID uuidv7.UUID, invoiceID uuidv7.UUID, amount int64, currency string, method entity.PaymentMethod, transactionID string) (*entity.Payment, error)
	GetPayment(ctx context.Context, id uuidv7.UUID) (*entity.Payment, error)
	GetPaymentByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error)
	GetInvoicePayments(ctx context.Context, invoiceID uuidv7.UUID) ([]*entity.Payment, error)
	GetUserPayments(ctx context.Context, userID uuidv7.UUID) ([]*entity.Payment, error)
	CompletePayment(ctx context.Context, id uuidv7.UUID, transactionID string) error
	RefundPayment(ctx context.Context, id uuidv7.UUID, reason string) error
	ListPayments(ctx context.Context, status *entity.PaymentStatus, limit, offset int) ([]*entity.Payment, error)
	CountPayments(ctx context.Context, status *entity.PaymentStatus) (int, error)
}

type paymentUseCase struct {
	paymentRepo repository.IPaymentRepository
	invoiceRepo repository.IInvoiceRepository
	eventBus    bus.IBus
}

func NewPaymentUseCase(paymentRepo repository.IPaymentRepository, invoiceRepo repository.IInvoiceRepository, eventBus bus.IBus) IPaymentUseCase {
	return &paymentUseCase{
		paymentRepo: paymentRepo,
		invoiceRepo: invoiceRepo,
		eventBus:    eventBus,
	}
}

func (uc *paymentUseCase) CreatePayment(ctx context.Context, userID uuidv7.UUID, invoiceID uuidv7.UUID, amount int64, currency string, method entity.PaymentMethod, transactionID string) (*entity.Payment, error) {
	payment, err := entity.NewPayment(userID, invoiceID, transactionID, amount, currency, method)
	if err != nil {
		return nil, err
	}
	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}
	return payment, nil
}

func (uc *paymentUseCase) GetPayment(ctx context.Context, id uuidv7.UUID) (*entity.Payment, error) {
	return uc.paymentRepo.GetByID(ctx, id)
}

func (uc *paymentUseCase) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error) {
	return uc.paymentRepo.GetByTransactionID(ctx, transactionID)
}

func (uc *paymentUseCase) GetInvoicePayments(ctx context.Context, invoiceID uuidv7.UUID) ([]*entity.Payment, error) {
	return uc.paymentRepo.GetByInvoiceID(ctx, invoiceID)
}

func (uc *paymentUseCase) GetUserPayments(ctx context.Context, userID uuidv7.UUID) ([]*entity.Payment, error) {
	return uc.paymentRepo.GetByUserID(ctx, userID)
}

func (uc *paymentUseCase) CompletePayment(ctx context.Context, id uuidv7.UUID, transactionID string) error {
	payment, err := uc.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	payment.MarkCompleted()
	payment.TransactionID = transactionID
	return uc.paymentRepo.Update(ctx, payment)
}

func (uc *paymentUseCase) RefundPayment(ctx context.Context, id uuidv7.UUID, reason string) error {
	payment, err := uc.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	payment.MarkRefunded(reason)
	return uc.paymentRepo.Update(ctx, payment)
}

func (uc *paymentUseCase) ListPayments(ctx context.Context, status *entity.PaymentStatus, limit, offset int) ([]*entity.Payment, error) {
	return uc.paymentRepo.List(ctx, status, limit, offset)
}

func (uc *paymentUseCase) CountPayments(ctx context.Context, status *entity.PaymentStatus) (int, error) {
	return uc.paymentRepo.Count(ctx, status)
}
