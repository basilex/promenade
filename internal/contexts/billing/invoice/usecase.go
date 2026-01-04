package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IUseCase defines business operations for invoice management
type IUseCase interface {
	// Invoice operations
	CreateInvoice(ctx context.Context, customerID uuidv7.UUID, orderID *uuidv7.UUID, dueDate time.Time, currency string) (*Invoice, error)
	GetInvoice(ctx context.Context, id uuidv7.UUID) (*Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNo string) (*Invoice, error)
	UpdateInvoice(ctx context.Context, inv *Invoice) error
	DeleteInvoice(ctx context.Context, id uuidv7.UUID) error

	// Line item operations
	AddLineItem(ctx context.Context, invoiceID uuidv7.UUID, description string, quantity int, unitPrice valueobject.Money) (*Invoice, error)
	RemoveLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID) (*Invoice, error)
	UpdateLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID, quantity int) (*Invoice, error)

	// Status transitions
	SendInvoice(ctx context.Context, id uuidv7.UUID) error
	MarkAsPaid(ctx context.Context, id uuidv7.UUID, paidDate time.Time) error
	MarkAsOverdue(ctx context.Context, id uuidv7.UUID) error
	CancelInvoice(ctx context.Context, id uuidv7.UUID) error
	VoidInvoice(ctx context.Context, id uuidv7.UUID) error

	// Tax operations
	UpdateTaxAmount(ctx context.Context, id uuidv7.UUID, taxAmount valueobject.Money) error

	// Query operations
	ListInvoices(ctx context.Context, page, pageSize int) ([]*Invoice, int, error)
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Invoice, int, error)
	ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*Invoice, error)
	ListByStatus(ctx context.Context, status InvoiceStatus, page, pageSize int) ([]*Invoice, int, error)
	ListOverdue(ctx context.Context, page, pageSize int) ([]*Invoice, int, error)

	// Analytics
	CountByStatus(ctx context.Context, status InvoiceStatus) (int, error)
	GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error)
}

type useCase struct {
	repo IRepository
}

// NewUseCase creates a new invoice use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateInvoice creates a new draft invoice with auto-generated invoice number
func (uc *useCase) CreateInvoice(ctx context.Context, customerID uuidv7.UUID, orderID *uuidv7.UUID, dueDate time.Time, currency string) (*Invoice, error) {
	// Generate invoice number
	invoiceNo, err := uc.repo.GenerateInvoiceNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice entity
	inv, err := NewInvoice(customerID, dueDate, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice entity: %w", err)
	}

	inv.InvoiceNo = invoiceNo
	inv.OrderID = orderID

	// Persist to database
	if err := uc.repo.Create(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	return inv, nil
}

// GetInvoice retrieves an invoice by ID
func (uc *useCase) GetInvoice(ctx context.Context, id uuidv7.UUID) (*Invoice, error) {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return inv, nil
}

// GetInvoiceByNumber retrieves an invoice by invoice number
func (uc *useCase) GetInvoiceByNumber(ctx context.Context, invoiceNo string) (*Invoice, error) {
	inv, err := uc.repo.GetByInvoiceNo(ctx, invoiceNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}

	return inv, nil
}

// UpdateInvoice updates an existing invoice
func (uc *useCase) UpdateInvoice(ctx context.Context, inv *Invoice) error {
	// Validate invoice before update
	if err := inv.Validate(); err != nil {
		return fmt.Errorf("invoice validation failed: %w", err)
	}

	// Update timestamp
	inv.Touch()

	if err := uc.repo.Update(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

// DeleteInvoice soft deletes an invoice
func (uc *useCase) DeleteInvoice(ctx context.Context, id uuidv7.UUID) error {
	// Verify invoice exists
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Only allow deleting draft invoices
	if inv.Status != InvoiceStatusDraft {
		return fmt.Errorf("can only delete draft invoices, current status: %s", inv.Status)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	return nil
}

// AddLineItem adds a line item to a draft invoice
func (uc *useCase) AddLineItem(ctx context.Context, invoiceID uuidv7.UUID, description string, quantity int, unitPrice valueobject.Money) (*Invoice, error) {
	// Get invoice
	inv, err := uc.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Add line item (entity validates status is draft)
	if err := inv.AddLine(description, quantity, unitPrice); err != nil {
		return nil, fmt.Errorf("failed to add line item: %w", err)
	}

	// Update invoice with recalculated totals
	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Create line item in database
	lastLine := inv.Lines[len(inv.Lines)-1]
	if err := uc.repo.CreateLine(ctx, &lastLine); err != nil {
		return nil, fmt.Errorf("failed to save line item: %w", err)
	}

	return inv, nil
}

// RemoveLineItem removes a line item from a draft invoice
func (uc *useCase) RemoveLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID) (*Invoice, error) {
	// Get invoice
	inv, err := uc.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Remove line item (entity validates status is draft)
	if err := inv.RemoveLine(lineID); err != nil {
		return nil, fmt.Errorf("failed to remove line item: %w", err)
	}

	// Delete from database
	if err := uc.repo.DeleteLine(ctx, lineID); err != nil {
		return nil, fmt.Errorf("failed to delete line item: %w", err)
	}

	// Update invoice with recalculated totals
	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	return inv, nil
}

// UpdateLineItem updates the quantity of a line item
func (uc *useCase) UpdateLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID, quantity int) (*Invoice, error) {
	// Get invoice
	inv, err := uc.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check status
	if inv.Status != InvoiceStatusDraft {
		return nil, ErrInvoiceCannotModifyNonDraft
	}

	// Find and update line
	var found bool
	for i := range inv.Lines {
		if inv.Lines[i].ID == lineID {
			if quantity < 1 {
				return nil, ErrInvoiceLineInvalidQuantity
			}

			inv.Lines[i].Quantity = quantity

			// Recalculate line amount
			amount, err := valueobject.NewMoney(
				inv.Lines[i].UnitPrice.Amount*int64(quantity),
				inv.Lines[i].UnitPrice.Currency,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate line amount: %w", err)
			}
			inv.Lines[i].Amount = amount
			inv.Lines[i].UpdatedAt = time.Now()

			// Update line in database
			if err := uc.repo.DeleteLine(ctx, lineID); err != nil {
				return nil, fmt.Errorf("failed to delete old line: %w", err)
			}
			if err := uc.repo.CreateLine(ctx, &inv.Lines[i]); err != nil {
				return nil, fmt.Errorf("failed to create updated line: %w", err)
			}

			found = true
			break
		}
	}

	if !found {
		return nil, ErrInvoiceLineNotFound
	}

	// Recalculate totals
	inv.recalculateTotal()

	// Update invoice
	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	return inv, nil
}

// SendInvoice marks a draft invoice as sent
func (uc *useCase) SendInvoice(ctx context.Context, id uuidv7.UUID) error {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := inv.MarkAsSent(); err != nil {
		return fmt.Errorf("failed to mark invoice as sent: %w", err)
	}

	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

// MarkAsPaid marks an invoice as paid
func (uc *useCase) MarkAsPaid(ctx context.Context, id uuidv7.UUID, paidDate time.Time) error {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := inv.MarkAsPaid(paidDate); err != nil {
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

// MarkAsOverdue marks a sent invoice as overdue
func (uc *useCase) MarkAsOverdue(ctx context.Context, id uuidv7.UUID) error {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := inv.MarkAsOverdue(); err != nil {
		return fmt.Errorf("failed to mark invoice as overdue: %w", err)
	}

	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

// CancelInvoice cancels an invoice
func (uc *useCase) CancelInvoice(ctx context.Context, id uuidv7.UUID) error {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := inv.Cancel(); err != nil {
		return fmt.Errorf("failed to cancel invoice: %w", err)
	}

	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

// VoidInvoice voids an invoice for accounting purposes
func (uc *useCase) VoidInvoice(ctx context.Context, id uuidv7.UUID) error {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := inv.Void(); err != nil {
		return fmt.Errorf("failed to void invoice: %w", err)
	}

	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

// UpdateTaxAmount updates the tax amount on a draft invoice
func (uc *useCase) UpdateTaxAmount(ctx context.Context, id uuidv7.UUID, taxAmount valueobject.Money) error {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if err := inv.UpdateTaxAmount(taxAmount); err != nil {
		return fmt.Errorf("failed to update tax amount: %w", err)
	}

	if err := uc.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

// ListInvoices retrieves all invoices with pagination
func (uc *useCase) ListInvoices(ctx context.Context, page, pageSize int) ([]*Invoice, int, error) {
	invoices, total, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}

	return invoices, total, nil
}

// ListByCustomer retrieves invoices for a specific customer
func (uc *useCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Invoice, int, error) {
	invoices, total, err := uc.repo.ListByCustomer(ctx, customerID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by customer: %w", err)
	}

	return invoices, total, nil
}

// ListByOrder retrieves invoices for a specific order
func (uc *useCase) ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*Invoice, error) {
	invoices, err := uc.repo.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices by order: %w", err)
	}

	return invoices, nil
}

// ListByStatus retrieves invoices by status with pagination
func (uc *useCase) ListByStatus(ctx context.Context, status InvoiceStatus, page, pageSize int) ([]*Invoice, int, error) {
	invoices, total, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by status: %w", err)
	}

	return invoices, total, nil
}

// ListOverdue retrieves overdue invoices with pagination
func (uc *useCase) ListOverdue(ctx context.Context, page, pageSize int) ([]*Invoice, int, error) {
	invoices, total, err := uc.repo.ListOverdue(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list overdue invoices: %w", err)
	}

	return invoices, total, nil
}

// CountByStatus counts invoices by status
func (uc *useCase) CountByStatus(ctx context.Context, status InvoiceStatus) (int, error) {
	count, err := uc.repo.CountByStatus(ctx, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count invoices by status: %w", err)
	}

	return count, nil
}

// GetTotalRevenue calculates total revenue in a date range
func (uc *useCase) GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error) {
	total, err := uc.repo.GetTotalRevenue(ctx, from, to)
	if err != nil {
		return 0, fmt.Errorf("failed to get total revenue: %w", err)
	}

	return total, nil
}
