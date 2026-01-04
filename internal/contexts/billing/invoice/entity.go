package invoice

import (
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// Invoice is an aggregate root for invoice management
type Invoice struct {
	aggregate.BaseAggregate

	// Identity
	ID        uuidv7.UUID
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

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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
		return nil, ErrInvoiceInvalidCustomerID
	}

	now := time.Now()
	if dueDate.Before(now) {
		return nil, ErrInvoiceInvalidDueDate
	}

	if currency == "" {
		currency = "USD"
	}

	subtotal, _ := valueobject.NewMoney(0, currency)
	tax, _ := valueobject.NewMoney(0, currency)
	total, _ := valueobject.NewMoney(0, currency)

	return &Invoice{
		ID:             uuidv7.New(),
		CustomerID:     customerID,
		IssueDate:      now,
		DueDate:        dueDate,
		Status:         InvoiceStatusDraft,
		Currency:       currency,
		SubtotalAmount: subtotal,
		TaxAmount:      tax,
		TotalAmount:    total,
		Lines:          []InvoiceLine{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// AddLine adds a line item to the invoice
func (i *Invoice) AddLine(description string, quantity int, unitPrice valueobject.Money) error {
	if i.Status != InvoiceStatusDraft {
		return ErrInvoiceCannotModifyNonDraft
	}

	if description == "" {
		return ErrInvoiceLineInvalidDescription
	}

	if quantity <= 0 {
		return ErrInvoiceLineInvalidQuantity
	}

	if unitPrice.Amount < 0 {
		return ErrInvoiceLineInvalidUnitPrice
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
		return ErrInvoiceCannotModifyNonDraft
	}

	for idx, line := range i.Lines {
		if line.ID == lineID {
			i.Lines = append(i.Lines[:idx], i.Lines[idx+1:]...)
			i.recalculateTotal()
			i.Touch()
			return nil
		}
	}

	return ErrInvoiceLineNotFound
}

// MarkAsSent marks the invoice as sent to customer
func (i *Invoice) MarkAsSent() error {
	if i.Status != InvoiceStatusDraft {
		return ErrInvoiceInvalidStatusTransition
	}

	if len(i.Lines) == 0 {
		return ErrInvoiceNoLines
	}

	i.Status = InvoiceStatusSent
	i.Touch()

	return nil
}

// MarkAsPaid marks the invoice as paid
func (i *Invoice) MarkAsPaid(paidDate time.Time) error {
	if i.Status != InvoiceStatusSent && i.Status != InvoiceStatusOverdue {
		return ErrInvoiceInvalidStatusTransition
	}

	i.Status = InvoiceStatusPaid
	i.PaidDate = &paidDate
	i.Touch()

	return nil
}

// MarkAsOverdue marks the invoice as overdue
func (i *Invoice) MarkAsOverdue() error {
	if i.Status != InvoiceStatusSent {
		return ErrInvoiceInvalidStatusTransition
	}

	now := time.Now()
	if !now.After(i.DueDate) {
		return ErrInvoiceNotYetDue
	}

	i.Status = InvoiceStatusOverdue
	i.Touch()

	return nil
}

// Cancel cancels the invoice
func (i *Invoice) Cancel() error {
	if i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusVoid || i.Status == InvoiceStatusCancelled {
		return ErrInvoiceCannotCancelTerminalState
	}

	i.Status = InvoiceStatusCancelled
	i.Touch()

	return nil
}

// Void voids the invoice (for accounting purposes)
func (i *Invoice) Void() error {
	if i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusVoid || i.Status == InvoiceStatusCancelled {
		return ErrInvoiceCannotVoidTerminalState
	}

	i.Status = InvoiceStatusVoid
	i.Touch()

	return nil
}

// UpdateTaxAmount sets the tax amount and recalculates total
func (i *Invoice) UpdateTaxAmount(taxAmount valueobject.Money) error {
	if i.Status != InvoiceStatusDraft {
		return ErrInvoiceCannotModifyNonDraft
	}

	if taxAmount.Amount < 0 {
		return ErrInvoiceInvalidTaxAmount
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
		return ErrInvoiceInvalidID
	}

	if i.CustomerID == uuidv7.Nil {
		return ErrInvoiceInvalidCustomerID
	}

	if i.DueDate.Before(i.IssueDate) {
		return ErrInvoiceInvalidDueDate
	}

	if i.Status == "" {
		return ErrInvoiceInvalidStatus
	}

	if i.Currency == "" {
		return ErrInvoiceInvalidCurrency
	}

	// Sent invoices must have lines
	if i.Status != InvoiceStatusDraft && len(i.Lines) == 0 {
		return ErrInvoiceNoLines
	}

	return nil
}

// Touch updates the UpdatedAt timestamp
func (i *Invoice) Touch() {
	i.UpdatedAt = time.Now()
}
