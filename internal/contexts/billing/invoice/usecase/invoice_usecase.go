package usecase

import (
	"context"
	"fmt"
	"time"

	invoiceerrors "github.com/basilex/promenade/internal/contexts/billing/invoice"
	"github.com/basilex/promenade/internal/contexts/billing/invoice/aggregate"
	"github.com/basilex/promenade/internal/contexts/billing/invoice/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IInvoiceUseCase defines business operations for invoice management
type IInvoiceUseCase interface {
	// Invoice operations
	CreateInvoice(ctx context.Context, customerID uuidv7.UUID, orderID *uuidv7.UUID, dueDate time.Time, currency string) (*aggregate.Invoice, error)
	GetInvoice(ctx context.Context, id uuidv7.UUID) (*aggregate.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNo string) (*aggregate.Invoice, error)
	UpdateInvoice(ctx context.Context, inv *aggregate.Invoice) error
	DeleteInvoice(ctx context.Context, id uuidv7.UUID) error

	// Line item operations
	AddLineItem(ctx context.Context, invoiceID uuidv7.UUID, description string, quantity int, unitPrice valueobject.Money) (*aggregate.Invoice, error)
	RemoveLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID) (*aggregate.Invoice, error)
	UpdateLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID, quantity int) (*aggregate.Invoice, error)

	// Status transitions
	SendInvoice(ctx context.Context, id uuidv7.UUID) error
	MarkAsPaid(ctx context.Context, id uuidv7.UUID, paidDate time.Time) error
	MarkAsOverdue(ctx context.Context, id uuidv7.UUID) error
	CancelInvoice(ctx context.Context, id uuidv7.UUID) error
	VoidInvoice(ctx context.Context, id uuidv7.UUID) error

	// Tax operations
	UpdateTaxAmount(ctx context.Context, id uuidv7.UUID, taxAmount valueobject.Money) error

	// Query operations
	ListInvoices(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error)
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Invoice, int, error)
	ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*aggregate.Invoice, error)
	ListByStatus(ctx context.Context, status aggregate.InvoiceStatus, page, pageSize int) ([]*aggregate.Invoice, int, error)
	ListOverdue(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error)

	// Analytics
	CountByStatus(ctx context.Context, status aggregate.InvoiceStatus) (int, error)
	GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error)
}

type InvoiceUseCase struct {
	repo repository.IInvoiceRepository
}

// NewInvoiceUseCase creates a new invoice use case
func NewInvoiceUseCase(repo repository.IInvoiceRepository) IInvoiceUseCase {
	return &InvoiceUseCase{
		repo: repo,
	}
}

// CreateInvoice creates a new draft invoice with auto-generated invoice number
func (u *InvoiceUseCase) CreateInvoice(ctx context.Context, customerID uuidv7.UUID, orderID *uuidv7.UUID, dueDate time.Time, currency string) (*aggregate.Invoice, error) {
	// Generate invoice number
	invoiceNo, err := u.repo.GenerateInvoiceNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGenerateNumberFailed, err)
	}

	// Create invoice entity
	inv, err := aggregate.NewInvoice(customerID, dueDate, currency)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	inv.InvoiceNo = invoiceNo
	inv.OrderID = orderID

	// Persist to database
	if err := u.repo.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceCreateFailed, err)
	}

	return inv, nil
}

// GetInvoice retrieves an invoice by ID
func (u *InvoiceUseCase) GetInvoice(ctx context.Context, id uuidv7.UUID) (*aggregate.Invoice, error) {
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	return inv, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (u *InvoiceUseCase) GetInvoiceByNumber(ctx context.Context, invoiceNo string) (*aggregate.Invoice, error) {
	inv, err := u.repo.GetByInvoiceNo(ctx, invoiceNo)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	return inv, nil
}

// UpdateInvoice updates an existing invoice
func (u *InvoiceUseCase) UpdateInvoice(ctx context.Context, inv *aggregate.Invoice) error {
	// Validate invoice before update
	if err := inv.Validate(); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	// Update timestamp
	inv.Touch()

	if err := u.repo.Update(ctx, inv); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return nil
}

// DeleteInvoice soft deletes an invoice
func (u *InvoiceUseCase) DeleteInvoice(ctx context.Context, id uuidv7.UUID) error {
	// Verify invoice exists
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	// Only allow deleting draft invoices
	if inv.Status != aggregate.InvoiceStatusDraft {
		return invoiceerrors.ErrInvoiceCannotModifyNonDraft
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceDeleteFailed, err)
	}

	return nil
}

// AddLineItem adds a line item to a draft invoice
func (u *InvoiceUseCase) AddLineItem(ctx context.Context, invoiceID uuidv7.UUID, description string, quantity int, unitPrice valueobject.Money) (*aggregate.Invoice, error) {
	// Get invoice
	inv, err := u.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	// Add line item (entity validates status is draft)
	if err := inv.AddLine(description, quantity, unitPrice); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	// Update invoice with recalculated totals
	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	// Create line item in database
	lastLine := inv.Lines[len(inv.Lines)-1]
	if err := u.repo.CreateLine(ctx, &lastLine); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceLineCreateFailed, err)
	}

	return inv, nil
}

// RemoveLineItem removes a line item from a draft invoice
func (u *InvoiceUseCase) RemoveLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID) (*aggregate.Invoice, error) {
	// Get invoice
	inv, err := u.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	// Remove line item (entity validates status is draft)
	if err := inv.RemoveLine(lineID); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	// Delete from database
	if err := u.repo.DeleteLine(ctx, lineID); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceLineDeleteFailed, err)
	}

	// Update invoice with recalculated totals
	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return inv, nil
}

// UpdateLineItem updates the quantity of a line item
func (u *InvoiceUseCase) UpdateLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID, quantity int) (*aggregate.Invoice, error) {
	// Get invoice
	inv, err := u.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	// Check status
	if inv.Status != aggregate.InvoiceStatusDraft {
		return nil, invoiceerrors.ErrInvoiceCannotModifyNonDraft
	}

	// Find and update line
	var found bool
	for i := range inv.Lines {
		if inv.Lines[i].ID == lineID {
			if quantity < 1 {
				return nil, invoiceerrors.ErrInvoiceLineInvalidQuantity
			}

			inv.Lines[i].Quantity = quantity

			// Recalculate line amount
			amount, err := valueobject.NewMoney(
				inv.Lines[i].UnitPrice.Amount*int64(quantity),
				inv.Lines[i].UnitPrice.Currency,
			)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceCalculationFailed, err)
			}
			inv.Lines[i].Amount = amount
			inv.Lines[i].UpdatedAt = time.Now()

			// Update line in database
			if err := u.repo.DeleteLine(ctx, lineID); err != nil {
				return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceLineDeleteFailed, err)
			}
			if err := u.repo.CreateLine(ctx, &inv.Lines[i]); err != nil {
				return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceLineCreateFailed, err)
			}

			found = true
			break
		}
	}

	if !found {
		return nil, invoiceerrors.ErrInvoiceLineNotFound
	}

	// Update invoice (recalculation is done by RemoveLine)
	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return inv, nil
}

// SendInvoice marks a draft invoice as sent
func (u *InvoiceUseCase) SendInvoice(ctx context.Context, id uuidv7.UUID) error {
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	if err := inv.MarkAsSent(); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return nil
}

// MarkAsPaid marks an invoice as paid
func (u *InvoiceUseCase) MarkAsPaid(ctx context.Context, id uuidv7.UUID, paidDate time.Time) error {
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	if err := inv.MarkAsPaid(paidDate); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return nil
}

// MarkAsOverdue marks a sent invoice as overdue
func (u *InvoiceUseCase) MarkAsOverdue(ctx context.Context, id uuidv7.UUID) error {
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	if err := inv.MarkAsOverdue(); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return nil
}

// CancelInvoice cancels an invoice
func (u *InvoiceUseCase) CancelInvoice(ctx context.Context, id uuidv7.UUID) error {
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	if err := inv.Cancel(); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return nil
}

// VoidInvoice voids an invoice for accounting purposes
func (u *InvoiceUseCase) VoidInvoice(ctx context.Context, id uuidv7.UUID) error {
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	if err := inv.Void(); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return nil
}

// UpdateTaxAmount updates the tax amount on a draft invoice
func (u *InvoiceUseCase) UpdateTaxAmount(ctx context.Context, id uuidv7.UUID, taxAmount valueobject.Money) error {
	inv, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceGetFailed, err)
	}

	if err := inv.UpdateTaxAmount(taxAmount); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceValidationFailed, err)
	}

	if err := u.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceUpdateFailed, err)
	}

	return nil
}

// ListInvoices retrieves all invoices with pagination
func (u *InvoiceUseCase) ListInvoices(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	invoices, total, err := u.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceListFailed, err)
	}

	return invoices, total, nil
}

// ListByCustomer retrieves invoices for a specific customer
func (u *InvoiceUseCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	invoices, total, err := u.repo.ListByCustomer(ctx, customerID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceListFailed, err)
	}

	return invoices, total, nil
}

// ListByOrder retrieves invoices for a specific order
func (u *InvoiceUseCase) ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*aggregate.Invoice, error) {
	invoices, err := u.repo.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceListFailed, err)
	}

	return invoices, nil
}

// ListByStatus retrieves invoices by status with pagination
func (u *InvoiceUseCase) ListByStatus(ctx context.Context, status aggregate.InvoiceStatus, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	invoices, total, err := u.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceListFailed, err)
	}

	return invoices, total, nil
}

// ListOverdue retrieves overdue invoices with pagination
func (u *InvoiceUseCase) ListOverdue(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	invoices, total, err := u.repo.ListOverdue(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceListFailed, err)
	}

	return invoices, total, nil
}

// CountByStatus counts invoices by status
func (u *InvoiceUseCase) CountByStatus(ctx context.Context, status aggregate.InvoiceStatus) (int, error) {
	count, err := u.repo.CountByStatus(ctx, status)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceListFailed, err)
	}

	return count, nil
}

// GetTotalRevenue calculates total revenue in a date range
func (u *InvoiceUseCase) GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error) {
	total, err := u.repo.GetTotalRevenue(ctx, from, to)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", invoiceerrors.ErrInvoiceRevenueCalculationFailed, err)
	}

	return total, nil
}
