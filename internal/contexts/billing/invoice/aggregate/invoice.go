package aggregate

import (
	"time"

	invoiceerrors "github.com/basilex/promenade/internal/contexts/billing/invoice"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// Invoice is an aggregate root for invoice management
type Invoice struct {
	aggregate.BaseAggregate

	// Identity
	InvoiceNo string // INV-2026-000001 (auto-generated)

	// Relations
	CustomerID uuidv7.UUID  // FK to customer_customers
	OrderID    *uuidv7.UUID // Optional FK to order_orders

	// Amounts (Money value objects)
	SubtotalAmount valueobject.Money
	TaxAmount      valueobject.Money
	TotalAmount    valueobject.Money
	Currency       string // USD, EUR, UAH

	// Dates
	IssueDate time.Time
	DueDate   time.Time
	PaidDate  *time.Time

	// Status
	Status InvoiceStatus

	// Line Items (embedded)
	Lines []InvoiceLine
}

// InvoiceLine represents a line item in an invoice
type InvoiceLine struct {
	ID          uuidv7.UUID
	InvoiceID   uuidv7.UUID
	Description string
	Quantity    int
	UnitPrice   valueobject.Money
	Amount      valueobject.Money
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
	InvoiceStatusVoid      InvoiceStatus = "void"
)

// NewInvoice creates a new invoice
func NewInvoice(customerID uuidv7.UUID, dueDate time.Time, currency string) (*Invoice, error) {
	if customerID == uuidv7.Nil {
		return nil, invoiceerrors.ErrInvoiceInvalidCustomerID
	}

	now := time.Now()
	if dueDate.Before(now) {
		return nil, invoiceerrors.ErrInvoiceInvalidDueDate
	}

	if currency == "" {
		currency = "USD"
	}

	subtotal, _ := valueobject.NewMoney(0, currency)
	tax, _ := valueobject.NewMoney(0, currency)
	total, _ := valueobject.NewMoney(0, currency)

	return &Invoice{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		CustomerID:     customerID,
		IssueDate:      now,
		DueDate:        dueDate,
		Status:         InvoiceStatusDraft,
		Currency:       currency,
		SubtotalAmount: subtotal,
		TaxAmount:      tax,
		TotalAmount:    total,
		Lines:          []InvoiceLine{},
	}, nil
}

// AddLine adds a line item to the invoice
func (i *Invoice) AddLine(description string, quantity int, unitPrice valueobject.Money) error {
	if i.Status != InvoiceStatusDraft {
		return invoiceerrors.ErrInvoiceCannotModifyNonDraft
	}

	if description == "" {
		return invoiceerrors.ErrInvoiceLineInvalidDescription
	}

	if quantity <= 0 {
		return invoiceerrors.ErrInvoiceLineInvalidQuantity
	}

	if unitPrice.Amount < 0 {
		return invoiceerrors.ErrInvoiceLineInvalidUnitPrice
	}

	// Calculate line amount
	amount, _ := valueobject.NewMoney(unitPrice.Amount*int64(quantity), unitPrice.Currency)

	line := InvoiceLine{
		ID:          uuidv7.New(),
		InvoiceID:   i.ID,
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Amount:      amount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	i.Lines = append(i.Lines, line)
	i.recalculateTotal()
	i.Touch()

	return nil
}

// RemoveLine removes a line item from the invoice
func (i *Invoice) RemoveLine(lineID uuidv7.UUID) error {
	if i.Status != InvoiceStatusDraft {
		return invoiceerrors.ErrInvoiceCannotModifyNonDraft
	}

	for idx, line := range i.Lines {
		if line.ID == lineID {
			i.Lines = append(i.Lines[:idx], i.Lines[idx+1:]...)
			i.recalculateTotal()
			i.Touch()
			return nil
		}
	}

	return invoiceerrors.ErrInvoiceLineNotFound
}

// MarkAsSent marks the invoice as sent to customer
func (i *Invoice) MarkAsSent() error {
	if i.Status != InvoiceStatusDraft {
		return invoiceerrors.ErrInvoiceInvalidStatusTransition
	}

	if len(i.Lines) == 0 {
		return invoiceerrors.ErrInvoiceNoLines
	}

	i.Status = InvoiceStatusSent
	i.Touch()

	return nil
}

// MarkAsPaid marks the invoice as paid
func (i *Invoice) MarkAsPaid(paidDate time.Time) error {
	if i.Status != InvoiceStatusSent && i.Status != InvoiceStatusOverdue {
		return invoiceerrors.ErrInvoiceInvalidStatusTransition
	}

	i.Status = InvoiceStatusPaid
	i.PaidDate = &paidDate
	i.Touch()

	return nil
}

// MarkAsOverdue marks the invoice as overdue
func (i *Invoice) MarkAsOverdue() error {
	if i.Status != InvoiceStatusSent {
		return invoiceerrors.ErrInvoiceInvalidStatusTransition
	}

	now := time.Now()
	if !now.After(i.DueDate) {
		return invoiceerrors.ErrInvoiceNotYetDue
	}

	i.Status = InvoiceStatusOverdue
	i.Touch()

	return nil
}

// Cancel cancels the invoice
func (i *Invoice) Cancel() error {
	if i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusVoid || i.Status == InvoiceStatusCancelled {
		return invoiceerrors.ErrInvoiceCannotCancelTerminalState
	}

	i.Status = InvoiceStatusCancelled
	i.Touch()

	return nil
}

// Void voids the invoice (for accounting purposes)
func (i *Invoice) Void() error {
	if i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusVoid || i.Status == InvoiceStatusCancelled {
		return invoiceerrors.ErrInvoiceCannotVoidTerminalState
	}

	i.Status = InvoiceStatusVoid
	i.Touch()

	return nil
}

// UpdateTaxAmount sets the tax amount and recalculates total
func (i *Invoice) UpdateTaxAmount(taxAmount valueobject.Money) error {
	if i.Status != InvoiceStatusDraft {
		return invoiceerrors.ErrInvoiceCannotModifyNonDraft
	}

	if taxAmount.Amount < 0 {
		return invoiceerrors.ErrInvoiceInvalidTaxAmount
	}

	i.TaxAmount = taxAmount
	i.recalculateTotal()
	i.Touch()

	return nil
}

// recalculateTotal recalculates subtotal and total amounts
func (i *Invoice) recalculateTotal() {
	// Calculate subtotal from lines
	subtotal := int64(0)
	for _, line := range i.Lines {
		subtotal += line.Amount.Amount
	}
	i.SubtotalAmount, _ = valueobject.NewMoney(subtotal, i.Currency)

	// Calculate total (subtotal + tax)
	total := subtotal + i.TaxAmount.Amount
	i.TotalAmount, _ = valueobject.NewMoney(total, i.Currency)
}

// IsTerminal returns true if invoice is in a terminal state
func (i *Invoice) IsTerminal() bool {
	return i.Status == InvoiceStatusPaid ||
		i.Status == InvoiceStatusCancelled ||
		i.Status == InvoiceStatusVoid
}

// IsOverdue returns true if invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.Status != InvoiceStatusSent {
		return false
	}
	return time.Now().After(i.DueDate)
}

// Validate validates the invoice
func (i *Invoice) Validate() error {
	if i.ID == uuidv7.Nil {
		return invoiceerrors.ErrInvoiceInvalidID
	}

	if i.CustomerID == uuidv7.Nil {
		return invoiceerrors.ErrInvoiceInvalidCustomerID
	}

	if i.DueDate.Before(i.IssueDate) {
		return invoiceerrors.ErrInvoiceInvalidDueDate
	}

	if i.Status == "" {
		return invoiceerrors.ErrInvoiceInvalidStatus
	}

	if i.Currency == "" {
		return invoiceerrors.ErrInvoiceInvalidCurrency
	}

	// Sent/Paid invoices must have lines (cancelled/void are exempt)
	if (i.Status == InvoiceStatusSent || i.Status == InvoiceStatusOverdue || i.Status == InvoiceStatusPaid) && len(i.Lines) == 0 {
		return invoiceerrors.ErrInvoiceNoLines
	}

	return nil
}
