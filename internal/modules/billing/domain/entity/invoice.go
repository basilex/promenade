package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"
	InvoiceStatusOpen          InvoiceStatus = "open"
	InvoiceStatusPaid          InvoiceStatus = "paid"
	InvoiceStatusVoid          InvoiceStatus = "void"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
)

// IsValid checks if status is valid
func (s InvoiceStatus) IsValid() bool {
	switch s {
	case InvoiceStatusDraft, InvoiceStatusOpen, InvoiceStatusPaid, InvoiceStatusVoid, InvoiceStatusUncollectible:
		return true
	}
	return false
}

// Invoice represents a billing invoice
type Invoice struct {
	ID             uuidv7.UUID   `db:"id" json:"id"`
	SubscriptionID uuidv7.UUID   `db:"subscription_id" json:"subscription_id" validate:"required"`
	InvoiceNumber  string        `db:"invoice_number" json:"invoice_number" validate:"required"`
	Status         InvoiceStatus `db:"status" json:"status" validate:"required"`
	SubtotalAmount int64         `db:"subtotal_amount" json:"subtotal_amount" validate:"min=0"`
	TaxAmount      int64         `db:"tax_amount" json:"tax_amount" validate:"min=0"`
	TotalAmount    int64         `db:"total_amount" json:"total_amount" validate:"min=0"`
	AmountPaid     int64         `db:"amount_paid" json:"amount_paid" validate:"min=0"`
	AmountDue      int64         `db:"amount_due" json:"amount_due" validate:"min=0"`
	Currency       string        `db:"currency" json:"currency" validate:"required,len=3"`
	DueDate        *time.Time    `db:"due_date" json:"due_date"`
	PaidAt         *time.Time    `db:"paid_at" json:"paid_at"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time    `db:"deleted_at" json:"deleted_at,omitempty"`
}

// NewInvoice creates a new invoice
func NewInvoice(subscriptionID uuidv7.UUID, subtotalAmount, taxAmount int64, currency string, dueDays int) (*Invoice, error) {
	now := time.Now()
	dueDate := now.AddDate(0, 0, dueDays)

	totalAmount := subtotalAmount + taxAmount

	invoice := &Invoice{
		ID:             uuidv7.New(),
		SubscriptionID: subscriptionID,
		InvoiceNumber:  generateInvoiceNumber(now),
		Status:         InvoiceStatusDraft,
		SubtotalAmount: subtotalAmount,
		TaxAmount:      taxAmount,
		TotalAmount:    totalAmount,
		AmountPaid:     0,
		AmountDue:      totalAmount,
		Currency:       currency,
		DueDate:        &dueDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := invoice.Validate(); err != nil {
		return nil, err
	}

	return invoice, nil
}

// generateInvoiceNumber generates unique invoice number (INV-YYYYMMDD-XXXXXX)
func generateInvoiceNumber(t time.Time) string {
	return fmt.Sprintf("INV-%s-%06d", t.Format("20060102"), t.Unix()%1000000)
}

// Validate validates invoice data
func (i *Invoice) Validate() error {
	if i.SubscriptionID == uuidv7.Nil {
		return errors.New("subscription ID is required")
	}

	if i.InvoiceNumber == "" {
		return errors.New("invoice number is required")
	}

	if i.SubtotalAmount < 0 {
		return errors.New("subtotal amount cannot be negative")
	}

	if i.TaxAmount < 0 {
		return errors.New("tax amount cannot be negative")
	}

	if i.TotalAmount != i.SubtotalAmount+i.TaxAmount {
		return errors.New("total amount must equal subtotal + tax")
	}

	if i.AmountDue != i.TotalAmount-i.AmountPaid {
		return errors.New("amount due must equal total - paid")
	}

	if i.AmountPaid > i.TotalAmount {
		return errors.New("amount paid cannot exceed total amount")
	}

	if i.Currency == "" {
		return errors.New("currency is required")
	}

	// Validate currency format (3-letter code)
	if len(i.Currency) != 3 {
		return errors.New("invalid currency code")
	}

	return nil
}

// Finalize finalizes the invoice (draft -> open)
func (i *Invoice) Finalize() {
	i.Status = InvoiceStatusOpen
	i.UpdatedAt = time.Now()
}

// FinalizeWithValidation finalizes with validation
func (i *Invoice) FinalizeWithValidation() error {
	if i.Status != InvoiceStatusDraft {
		return errors.New("only draft invoices can be finalized")
	}

	i.Status = InvoiceStatusOpen
	i.UpdatedAt = time.Now()
	return nil
}

// MarkPaid marks invoice as fully paid (for tests and simple cases)
func (i *Invoice) MarkPaid() {
	i.Status = InvoiceStatusPaid
	i.AmountPaid = i.TotalAmount
	i.AmountDue = 0
	now := time.Now()
	i.PaidAt = &now
	i.UpdatedAt = now
}

// RecordPayment records a payment against the invoice
func (i *Invoice) RecordPayment(amount int64) error {
	if i.Status != InvoiceStatusOpen {
		return errors.New("only open invoices can receive payments")
	}

	if amount <= 0 {
		return errors.New("payment amount must be positive")
	}

	i.AmountPaid += amount
	i.AmountDue = i.TotalAmount - i.AmountPaid

	if i.AmountDue <= 0 {
		i.Status = InvoiceStatusPaid
		now := time.Now()
		i.PaidAt = &now
	}

	i.UpdatedAt = time.Now()
	return nil
}

// Void voids the invoice
func (i *Invoice) Void() {
	i.Status = InvoiceStatusVoid
	i.UpdatedAt = time.Now()
}

// VoidWithValidation voids with validation
func (i *Invoice) VoidWithValidation() error {
	if i.Status == InvoiceStatusPaid {
		return errors.New("cannot void paid invoice")
	}

	i.Status = InvoiceStatusVoid
	i.UpdatedAt = time.Now()
	return nil
}

// MarkUncollectible marks invoice as uncollectible
func (i *Invoice) MarkUncollectible() error {
	if i.Status == InvoiceStatusPaid {
		return errors.New("cannot mark paid invoice as uncollectible")
	}

	i.Status = InvoiceStatusUncollectible
	i.UpdatedAt = time.Now()
	return nil
}

// IsOverdue checks if invoice is overdue
func (i *Invoice) IsOverdue() bool {
	if i.Status != InvoiceStatusOpen {
		return false
	}

	if i.DueDate == nil {
		return false
	}

	return time.Now().After(*i.DueDate)
}

// IsOpen checks if invoice is open
func (i *Invoice) IsOpen() bool {
	return i.Status == InvoiceStatusOpen
}

// IsDraft checks if invoice is draft
func (i *Invoice) IsDraft() bool {
	return i.Status == InvoiceStatusDraft
}

// IsVoid checks if invoice is void
func (i *Invoice) IsVoid() bool {
	return i.Status == InvoiceStatusVoid
}

// IsPaid checks if invoice is fully paid
func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}

// String returns string representation
func (i *Invoice) String() string {
	return fmt.Sprintf("Invoice{Number: %s, Total: %d %s, Status: %s}",
		i.InvoiceNumber, i.TotalAmount, i.Currency, i.Status)
}
