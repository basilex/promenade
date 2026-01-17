package receipt

import (
	"context"
	"errors"
	"log/slog"

	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines business operations for fiscal receipts
type IUseCase interface {
	// CreateReceipt creates a new receipt
	CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType PaymentType, receiptType ReceiptType, currency string, lines []ReceiptLine, createdBy uuidv7.UUID) (*Receipt, error)

	// GetReceipt retrieves receipt by ID
	GetReceipt(ctx context.Context, id uuidv7.UUID) (*Receipt, error)

	// GetByOrderID retrieves receipt by order ID
	GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error)

	// ListReceipts retrieves receipts with optional filters
	ListReceipts(ctx context.Context, filters *ListFilters) ([]*Receipt, error)

	// MarkPrinted sets fiscal data and marks receipt as printed
	MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*Receipt, error)

	// PrintReceipt prints receipt via provider and stores fiscal data
	PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*Receipt, error)

	// CancelReceipt cancels a receipt
	CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*Receipt, error)

	// DeleteReceipt soft-deletes receipt
	DeleteReceipt(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
	repo    IRepository
	printer IPrinter
}

// NewUseCase creates a new receipt use case
func NewUseCase(repo IRepository, printer IPrinter) IUseCase {
	return &useCase{repo: repo, printer: printer}
}

// CreateReceipt creates a new fiscal receipt
func (uc *useCase) CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType PaymentType, receiptType ReceiptType, currency string, lines []ReceiptLine, createdBy uuidv7.UUID) (*Receipt, error) {
	// Prevent duplicate receipts for the same order
	if existing, err := uc.repo.GetByOrderID(ctx, orderID); err == nil && existing != nil {
		return nil, ErrReceiptAlreadyExists
	} else if err != nil && !errors.Is(err, ErrReceiptNotFound) {
		return nil, ErrReceiptCreateFailed
	}

	newReceipt, err := NewReceipt(cashRegisterID, orderID, paymentType, receiptType, currency, lines, createdBy)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, newReceipt); err != nil {
		return nil, ErrReceiptCreateFailed
	}

	return newReceipt, nil
}

// GetReceipt retrieves receipt by ID
func (uc *useCase) GetReceipt(ctx context.Context, id uuidv7.UUID) (*Receipt, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetByOrderID retrieves receipt by order ID
func (uc *useCase) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}

// ListReceipts retrieves receipts with optional filters
func (uc *useCase) ListReceipts(ctx context.Context, filters *ListFilters) ([]*Receipt, error) {
	receipts, err := uc.repo.List(ctx, filters)
	if err != nil {
		return nil, ErrReceiptListFailed
	}
	return receipts, nil
}

// MarkPrinted sets fiscal data and marks receipt as printed
func (uc *useCase) MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*Receipt, error) {
	rec, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := rec.MarkPrinted(fiscalNumber, fiscalURL, qrCode, printedBy); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, rec); err != nil {
		return nil, ErrReceiptUpdateFailed
	}

	return rec, nil
}

// PrintReceipt prints receipt via provider and stores fiscal data
func (uc *useCase) PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*Receipt, error) {
	rec, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if uc.printer == nil {
		logger.FromContext(ctx).Error("Receipt printer is not configured")
		return nil, ErrReceiptPrintFailed
	}

	result, err := uc.printer.Print(ctx, rec)
	if err != nil {
		logger.FromContext(ctx).Error("Failed to print receipt",
			slog.String("receipt_id", rec.GetID().String()),
			slog.Any("error", err),
		)
		return nil, ErrReceiptPrintFailed
	}

	if result != nil && result.ProviderReceiptID != "" {
		rec.ProviderReceiptID = result.ProviderReceiptID
	}

	if err := rec.MarkPrinted(result.FiscalNumber, result.FiscalURL, result.QRCode, printedBy); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, rec); err != nil {
		return nil, ErrReceiptUpdateFailed
	}

	return rec, nil
}

// CancelReceipt cancels a receipt
func (uc *useCase) CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*Receipt, error) {
	rec, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rec.Status == ReceiptStatusPrinted {
		if canceler, ok := uc.printer.(ICancelPrinter); ok && rec.ProviderReceiptID != "" {
			if err := canceler.Cancel(ctx, rec, reason); err != nil {
				logger.FromContext(ctx).Error("Failed to cancel receipt at provider",
					slog.String("receipt_id", rec.GetID().String()),
					slog.Any("error", err),
				)
				return nil, ErrReceiptCancelFailed
			}
		}
	}

	if err := rec.Cancel(reason, cancelledBy); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, rec); err != nil {
		return nil, ErrReceiptUpdateFailed
	}

	return rec, nil
}

// DeleteReceipt soft-deletes receipt
func (uc *useCase) DeleteReceipt(ctx context.Context, id uuidv7.UUID) error {
	if _, err := uc.repo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return ErrReceiptDeleteFailed
	}

	return nil
}
