package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusSucceeded  PaymentStatus = "succeeded"
	PaymentStatusCompleted  PaymentStatus = "completed" // Alias for succeeded
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCanceled   PaymentStatus = "canceled"
)

// IsValid checks if status is valid
func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentStatusPending, PaymentStatusProcessing, PaymentStatusSucceeded, 
		PaymentStatusCompleted, PaymentStatusFailed, PaymentStatusRefunded, PaymentStatusCanceled:
		return true
	}
	return false
}

// PaymentMethod represents payment method
type PaymentMethod string

const (
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodPayPal       PaymentMethod = "paypal"
	PaymentMethodStripe       PaymentMethod = "stripe"
	PaymentMethodLiqPay       PaymentMethod = "liqpay"
	PaymentMethodCrypto       PaymentMethod = "crypto"
	PaymentMethodOther        PaymentMethod = "other"
)

// IsValid checks if method is valid
func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCard, PaymentMethodBankTransfer, PaymentMethodPayPal, 
		PaymentMethodStripe, PaymentMethodLiqPay, PaymentMethodCrypto, PaymentMethodOther:
		return true
	}
	return false
}

// Payment represents a payment transaction
type Payment struct {
	ID             uuidv7.UUID       `db:"id" json:"id"`
	InvoiceID      uuidv7.UUID       `db:"invoice_id" json:"invoice_id"`
	UserID         uuidv7.UUID       `db:"user_id" json:"user_id" validate:"required"`
	TransactionID  string            `db:"transaction_id" json:"transaction_id" validate:"required"`
	Amount         int64             `db:"amount" json:"amount" validate:"required,min=1"`
	Currency       string            `db:"currency" json:"currency" validate:"required,len=3"`
	Status         PaymentStatus     `db:"status" json:"status" validate:"required"`
	Method         PaymentMethod     `db:"method" json:"method" validate:"required"`
	FailureReason  string            `db:"failure_reason" json:"failure_reason,omitempty"`
	RefundReason   string            `db:"refund_reason" json:"refund_reason,omitempty"`
	CompletedAt    *time.Time        `db:"completed_at" json:"completed_at"`
	FailedAt       *time.Time        `db:"failed_at" json:"failed_at"`
	RefundedAt     *time.Time        `db:"refunded_at" json:"refunded_at"`
	CreatedAt      time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time         `db:"updated_at" json:"updated_at"`
	DeletedAt      *time.Time        `db:"deleted_at" json:"deleted_at,omitempty"`
}

// NewPayment creates a new payment
func NewPayment(userID, invoiceID uuidv7.UUID, transactionID string, amount int64, currency string, method PaymentMethod) (*Payment, error) {
	now := time.Now()
	
	payment := &Payment{
		ID:            uuidv7.New(),
		UserID:        userID,
		InvoiceID:     invoiceID,
		TransactionID: transactionID,
		Amount:        amount,
		Currency:      currency,
		Status:        PaymentStatusPending,
		Method:        method,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	
	if err := payment.Validate(); err != nil {
		return nil, err
	}
	
	return payment, nil
}

// Validate validates payment data
func (p *Payment) Validate() error {
	if p.UserID == uuidv7.Nil {
		return errors.New("user ID is required")
	}
	
	if p.TransactionID == "" {
		return errors.New("transaction ID is required")
	}
	
	if p.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	
	if p.Currency == "" {
		return errors.New("currency is required")
	}
	
	if !p.Method.IsValid() {
		return errors.New("invalid payment method")
	}
	
	if !p.Status.IsValid() {
		return errors.New("invalid status")
	}
	
	return nil
}

// MarkProcessing marks payment as processing
func (p *Payment) MarkProcessing() {
	p.Status = PaymentStatusProcessing
	p.UpdatedAt = time.Now()
}

// MarkSucceeded marks payment as succeeded
func (p *Payment) MarkSucceeded() {
	p.Status = PaymentStatusSucceeded
	p.UpdatedAt = time.Now()
}

// MarkCompleted marks payment as completed (same as succeeded)
func (p *Payment) MarkCompleted() {
	p.Status = PaymentStatusCompleted
	now := time.Now()
	p.CompletedAt = &now
	p.UpdatedAt = now
}

// MarkFailed marks payment as failed
func (p *Payment) MarkFailed(reason string) {
	p.Status = PaymentStatusFailed
	now := time.Now()
	p.FailedAt = &now
	p.FailureReason = reason
	p.UpdatedAt = now
}

// MarkRefunded marks payment as refunded
func (p *Payment) MarkRefunded(reason string) {
	p.Status = PaymentStatusRefunded
	now := time.Now()
	p.RefundedAt = &now
	p.RefundReason = reason
	p.UpdatedAt = now
}

// IsCompleted checks if payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted || p.Status == PaymentStatusSucceeded
}

// IsPending checks if payment is pending or processing
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusProcessing
}

// IsFailed checks if payment has failed
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

// IsRefunded checks if payment is refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded
}

// String returns string representation
func (p *Payment) String() string {
	return fmt.Sprintf("Payment{ID: %s, Amount: %d %s, Status: %s, Method: %s}",
		p.ID, p.Amount, p.Currency, p.Status, p.Method)
}
