package payment

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// PaymentMethod represents payment method type
type PaymentMethod string

const (
	PaymentMethodCreditCard  PaymentMethod = "credit_card"
	PaymentMethodDebitCard   PaymentMethod = "debit_card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodPayPal      PaymentMethod = "paypal"
	PaymentMethodStripe      PaymentMethod = "stripe"
	PaymentMethodCash        PaymentMethod = "cash"
	PaymentMethodCheck       PaymentMethod = "check"
	PaymentMethodWire        PaymentMethod = "wire"
	PaymentMethodACH         PaymentMethod = "ach"
	PaymentMethodOther       PaymentMethod = "other"
)

// PaymentStatus represents payment status
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

// Payment is the aggregate root for payment operations
type Payment struct {
	aggregate.BaseAggregate

	PaymentNo       string
	TransactionID   string
	CustomerID      uuidv7.UUID
	InvoiceID       *uuidv7.UUID
	Amount          valueobject.Money
	Method          PaymentMethod
	Status          PaymentStatus
	CardLast4       string
	CardBrand       string
	BankAccount     string
	PaymentProvider string
	FailureReason   string
	ProcessedAt     time.Time
	RefundedAt      *time.Time
	RefundedAmount  *valueobject.Money
	ProcessedBy     string
	Notes           string
}

// NewPayment creates a new Payment
func NewPayment(customerID uuidv7.UUID, amount valueobject.Money, method PaymentMethod) (*Payment, error) {
	if customerID == uuidv7.Nil {
		return nil, fmt.Errorf("customer ID is required")
	}
	if amount.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if method == "" {
		return nil, fmt.Errorf("payment method is required")
	}

	return &Payment{
		BaseAggregate: aggregate.NewBaseAggregate(),
		CustomerID:    customerID,
		Amount:        amount,
		Method:        method,
		Status:        PaymentStatusPending,
		ProcessedAt:   time.Now().UTC(),
	}, nil
}

// LinkToInvoice links payment to an invoice
func (p *Payment) LinkToInvoice(invoiceID uuidv7.UUID) error {
	if invoiceID == uuidv7.Nil {
		return fmt.Errorf("invoice ID is required")
	}
	p.InvoiceID = &invoiceID
	p.Touch()
	return nil
}

// Process marks payment as processing
func (p *Payment) Process() error {
	if p.Status != PaymentStatusPending {
		return fmt.Errorf("can only process pending payments")
	}
	p.Status = PaymentStatusProcessing
	p.Touch()
	return nil
}

// Complete marks payment as completed
func (p *Payment) Complete(transactionID string) error {
	if p.Status != PaymentStatusProcessing {
		return fmt.Errorf("can only complete processing payments")
	}
	if transactionID == "" {
		return fmt.Errorf("transaction ID is required")
	}
	p.Status = PaymentStatusCompleted
	p.TransactionID = transactionID
	p.Touch()
	return nil
}

// Fail marks payment as failed
func (p *Payment) Fail(reason string) error {
	if p.Status != PaymentStatusProcessing {
		return fmt.Errorf("can only fail processing payments")
	}
	if reason == "" {
		return fmt.Errorf("failure reason is required")
	}
	p.Status = PaymentStatusFailed
	p.FailureReason = reason
	p.Touch()
	return nil
}

// Refund processes a refund
func (p *Payment) Refund(amount valueobject.Money) error {
	if p.Status != PaymentStatusCompleted {
		return fmt.Errorf("can only refund completed payments")
	}
	if amount.Amount <= 0 {
		return fmt.Errorf("refund amount must be greater than zero")
	}
	if amount.Amount > p.Amount.Amount {
		return fmt.Errorf("refund amount cannot exceed payment amount")
	}
	if p.RefundedAmount != nil {
		totalRefunded := p.RefundedAmount.Amount + amount.Amount
		if totalRefunded > p.Amount.Amount {
			return fmt.Errorf("total refunds cannot exceed payment amount")
		}
	}

	p.Status = PaymentStatusRefunded
	now := time.Now().UTC()
	p.RefundedAt = &now
	if p.RefundedAmount == nil {
		p.RefundedAmount = &amount
	} else {
		refundedCents := p.RefundedAmount.Amount + amount.Amount
		newRefundedAmount, _ := valueobject.NewMoney(refundedCents, p.Amount.Currency)
		p.RefundedAmount = &newRefundedAmount
	}
	p.Touch()
	return nil
}

// Cancel cancels a pending payment
func (p *Payment) Cancel() error {
	if p.Status != PaymentStatusPending {
		return fmt.Errorf("can only cancel pending payments")
	}
	p.Status = PaymentStatusCancelled
	p.Touch()
	return nil
}

// SetCardDetails sets credit card details
func (p *Payment) SetCardDetails(last4, brand string) {
	p.CardLast4 = last4
	p.CardBrand = brand
	p.Touch()
}

// SetBankAccount sets bank account details
func (p *Payment) SetBankAccount(account string) {
	p.BankAccount = account
	p.Touch()
}

// SetProvider sets payment provider
func (p *Payment) SetProvider(provider string) {
	p.PaymentProvider = provider
	p.Touch()
}

// SetProcessedBy sets who processed the payment
func (p *Payment) SetProcessedBy(processedBy string) {
	p.ProcessedBy = processedBy
	p.Touch()
}

// AddNote adds a note to the payment
func (p *Payment) AddNote(note string) {
	if p.Notes == "" {
		p.Notes = note
	} else {
		p.Notes = p.Notes + "\n" + note
	}
	p.Touch()
}

// SetPaymentNo sets the payment number
func (p *Payment) SetPaymentNo(paymentNo string) error {
	if paymentNo == "" {
		return fmt.Errorf("payment number cannot be empty")
	}
	p.PaymentNo = paymentNo
	p.Touch()
	return nil
}

// SetTransactionID sets the transaction ID
func (p *Payment) SetTransactionID(transactionID string) {
	p.TransactionID = transactionID
	p.Touch()
}

// IsCompleted returns true if payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted
}

// IsFailed returns true if payment failed
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

// IsRefunded returns true if payment is refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded
}

// IsCancelled returns true if payment is cancelled
func (p *Payment) IsCancelled() bool {
	return p.Status == PaymentStatusCancelled
}

// IsPending returns true if payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

// IsProcessing returns true if payment is processing
func (p *Payment) IsProcessing() bool {
	return p.Status == PaymentStatusProcessing
}

// CanBeModified returns true if payment can be modified
func (p *Payment) CanBeModified() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusCancelled
}

// CanBeDeleted returns true if payment can be deleted
func (p *Payment) CanBeDeleted() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusCancelled || p.Status == PaymentStatusFailed
}

// Validate validates payment data
func (p *Payment) Validate() error {
	if p.CustomerID == uuidv7.Nil {
		return fmt.Errorf("customer ID is required")
	}
	if p.Amount.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if p.Method == "" {
		return fmt.Errorf("payment method is required")
	}
	if p.PaymentNo == "" {
		return fmt.Errorf("payment number is required")
	}
	return nil
}
