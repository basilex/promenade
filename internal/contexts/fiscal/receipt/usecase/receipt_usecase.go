package usecase

import (
	"context"
	"errors"
	"log/slog"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/repository"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IReceiptUseCase defines business operations for fiscal receipts
type IReceiptUseCase interface {
	// CreateReceipt creates a new receipt
	CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error)

	// GetReceipt retrieves receipt by ID
	GetReceipt(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error)

	// GetByOrderID retrieves receipt by order ID
	GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error)

	// ListReceipts retrieves receipts with optional filters
	ListReceipts(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error)

	// MarkPrinted sets fiscal data and marks receipt as printed
	MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*aggregate.Receipt, error)

	// PrintReceipt prints receipt via provider and stores fiscal data
	PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error)

	// CancelReceipt cancels a receipt
	CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error)

	// DeleteReceipt soft-deletes receipt
	DeleteReceipt(ctx context.Context, id uuidv7.UUID) error
}

type ReceiptUseCase struct {
	repo    repository.IReceiptRepository
	printer IPrinter
}

// NewReceiptUseCase creates a new receipt use case
func NewReceiptUseCase(repo repository.IReceiptRepository, printer IPrinter) IReceiptUseCase {
	return &ReceiptUseCase{repo: repo, printer: printer}
}

// CreateReceipt creates a new fiscal receipt
func (u *ReceiptUseCase) CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error) {
	// Prevent duplicate receipts for the same order
	if existing, err := u.repo.GetByOrderID(ctx, orderID); err == nil && existing != nil {
		return nil, receipterrors.ErrReceiptAlreadyExists
	} else if err != nil && !errors.Is(err, receipterrors.ErrReceiptNotFound) {
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	newReceipt, err := aggregate.NewReceipt(cashRegisterID, orderID, paymentType, receiptType, currency, lines, createdBy)
	if err != nil {
		return nil, err
	}

	if err := u.repo.Create(ctx, newReceipt); err != nil {
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	return newReceipt, nil
}

// GetReceipt retrieves receipt by ID
func (u *ReceiptUseCase) GetReceipt(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
	return u.repo.GetByID(ctx, id)
}

// GetByOrderID retrieves receipt by order ID
func (u *ReceiptUseCase) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
	return u.repo.GetByOrderID(ctx, orderID)
}

// ListReceipts retrieves receipts with optional filters
func (u *ReceiptUseCase) ListReceipts(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
	receipts, err := u.repo.List(ctx, filters)
	if err != nil {
		return nil, receipterrors.ErrReceiptListFailed
	}
	return receipts, nil
}

// MarkPrinted sets fiscal data and marks receipt as printed
func (u *ReceiptUseCase) MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
	rec, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := rec.MarkPrinted(fiscalNumber, fiscalURL, qrCode, printedBy); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, rec); err != nil {
		return nil, receipterrors.ErrReceiptUpdateFailed
	}

	return rec, nil
}

// PrintReceipt prints receipt via provider and stores fiscal data
func (u *ReceiptUseCase) PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
	rec, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if u.printer == nil {
		logger.FromContext(ctx).Error("Receipt printer is not configured")
		return nil, receipterrors.ErrReceiptPrintFailed
	}

	result, err := u.printer.Print(ctx, rec)
	if err != nil {
		logger.FromContext(ctx).Error("Failed to print receipt",
			slog.String("receipt_id", rec.GetID().String()),
			slog.Any("error", err),
		)
		return nil, receipterrors.ErrReceiptPrintFailed
	}

	if result != nil && result.ProviderReceiptID != "" {
		rec.ProviderReceiptID = result.ProviderReceiptID
	}

	if err := rec.MarkPrinted(result.FiscalNumber, result.FiscalURL, result.QRCode, printedBy); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, rec); err != nil {
		return nil, receipterrors.ErrReceiptUpdateFailed
	}

	return rec, nil
}

// CancelReceipt cancels a receipt
func (u *ReceiptUseCase) CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error) {
	rec, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rec.Status == aggregate.ReceiptStatusPrinted {
		if canceler, ok := u.printer.(ICancelPrinter); ok && rec.ProviderReceiptID != "" {
			if err := canceler.Cancel(ctx, rec, reason); err != nil {
				logger.FromContext(ctx).Error("Failed to cancel receipt at provider",
					slog.String("receipt_id", rec.GetID().String()),
					slog.Any("error", err),
				)
				return nil, receipterrors.ErrReceiptCancelFailed
			}
		}
	}

	if err := rec.Cancel(reason, cancelledBy); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, rec); err != nil {
		return nil, receipterrors.ErrReceiptUpdateFailed
	}

	return rec, nil
}

// DeleteReceipt soft-deletes receipt
func (u *ReceiptUseCase) DeleteReceipt(ctx context.Context, id uuidv7.UUID) error {
	if _, err := u.repo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return receipterrors.ErrReceiptDeleteFailed
	}

	return nil
}
