package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type IInvoiceUseCase interface {
	CreateInvoice(ctx context.Context, subscriptionID uuidv7.UUID, subtotal, tax int64, currency string, dueDays int) (*entity.Invoice, error)
	GetInvoice(ctx context.Context, id uuidv7.UUID) (*entity.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*entity.Invoice, error)
	GetSubscriptionInvoices(ctx context.Context, subscriptionID uuidv7.UUID) ([]*entity.Invoice, error)
	FinalizeInvoice(ctx context.Context, id uuidv7.UUID) error
	VoidInvoice(ctx context.Context, id uuidv7.UUID) error
	ListInvoices(ctx context.Context, status *entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error)
	CountInvoices(ctx context.Context, status *entity.InvoiceStatus) (int, error)
}

type invoiceUseCase struct {
	invoiceRepo      repository.IInvoiceRepository
	paymentRepo      repository.IPaymentRepository
	subscriptionRepo repository.ISubscriptionRepository
	eventBus         bus.IBus
}

func NewInvoiceUseCase(invoiceRepo repository.IInvoiceRepository, paymentRepo repository.IPaymentRepository, subscriptionRepo repository.ISubscriptionRepository, eventBus bus.IBus) IInvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepo:      invoiceRepo,
		paymentRepo:      paymentRepo,
		subscriptionRepo: subscriptionRepo,
		eventBus:         eventBus,
	}
}

func (uc *invoiceUseCase) CreateInvoice(ctx context.Context, subscriptionID uuidv7.UUID, subtotal, tax int64, currency string, dueDays int) (*entity.Invoice, error) {
	invoice, err := entity.NewInvoice(subscriptionID, subtotal, tax, currency, dueDays)
	if err != nil {
		return nil, err
	}
	if err := uc.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (uc *invoiceUseCase) GetInvoice(ctx context.Context, id uuidv7.UUID) (*entity.Invoice, error) {
	return uc.invoiceRepo.GetByID(ctx, id)
}

func (uc *invoiceUseCase) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*entity.Invoice, error) {
	return uc.invoiceRepo.GetByInvoiceNumber(ctx, invoiceNumber)
}

func (uc *invoiceUseCase) GetSubscriptionInvoices(ctx context.Context, subscriptionID uuidv7.UUID) ([]*entity.Invoice, error) {
	return uc.invoiceRepo.GetBySubscriptionID(ctx, subscriptionID)
}

func (uc *invoiceUseCase) FinalizeInvoice(ctx context.Context, id uuidv7.UUID) error {
	invoice, err := uc.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	invoice.Finalize()
	return uc.invoiceRepo.Update(ctx, invoice)
}

func (uc *invoiceUseCase) VoidInvoice(ctx context.Context, id uuidv7.UUID) error {
	invoice, err := uc.invoiceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	invoice.Void()
	return uc.invoiceRepo.Update(ctx, invoice)
}

func (uc *invoiceUseCase) ListInvoices(ctx context.Context, status *entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error) {
	return uc.invoiceRepo.List(ctx, status, limit, offset)
}

func (uc *invoiceUseCase) CountInvoices(ctx context.Context, status *entity.InvoiceStatus) (int, error) {
	return uc.invoiceRepo.Count(ctx, status)
}
