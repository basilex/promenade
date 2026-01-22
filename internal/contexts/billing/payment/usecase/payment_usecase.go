package usecase

import (
	"context"
	"fmt"
	"time"

	paymenterrors "github.com/basilex/promenade/internal/contexts/billing/payment"
	"github.com/basilex/promenade/internal/contexts/billing/payment/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IPaymentUseCase defines the interface for payment business logic
type IPaymentUseCase interface {
	// CreatePayment creates a new payment
	CreatePayment(ctx context.Context, customerID uuidv7.UUID, amount valueobject.Money, method aggregate.PaymentMethod) (*aggregate.Payment, error)

	// GetPayment retrieves a payment by ID
	GetPayment(ctx context.Context, id uuidv7.UUID) (*aggregate.Payment, error)

	// GetPaymentByNumber retrieves a payment by payment number
	GetPaymentByNumber(ctx context.Context, paymentNo string) (*aggregate.Payment, error)

	// GetPaymentByTransactionID retrieves a payment by transaction ID
	GetPaymentByTransactionID(ctx context.Context, transactionID string) (*aggregate.Payment, error)

	// LinkToInvoice links a payment to an invoice
	LinkToInvoice(ctx context.Context, paymentID, invoiceID uuidv7.UUID) error

	// ProcessPayment processes a pending payment
	ProcessPayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error

	// CompletePayment marks a payment as completed
	CompletePayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error

	// FailPayment marks a payment as failed
	FailPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error

	// RefundPayment refunds a payment (full or partial)
	RefundPayment(ctx context.Context, paymentID uuidv7.UUID, refundAmount valueobject.Money) error

	// CancelPayment cancels a pending payment
	CancelPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error

	// SetCardDetails sets card payment details
	SetCardDetails(ctx context.Context, paymentID uuidv7.UUID, last4, brand string) error

	// SetProvider sets the payment provider
	SetProvider(ctx context.Context, paymentID uuidv7.UUID, provider string) error

	// AddNote adds a note to the payment
	AddNote(ctx context.Context, paymentID uuidv7.UUID, note string) error

	// ListPayments lists all payments with pagination
	ListPayments(ctx context.Context, page, pageSize int) ([]*aggregate.Payment, error)

	// ListPaymentsByCustomer lists payments for a specific customer
	ListPaymentsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Payment, error)

	// ListPaymentsByInvoice lists payments for a specific invoice
	ListPaymentsByInvoice(ctx context.Context, invoiceID uuidv7.UUID) ([]*aggregate.Payment, error)

	// ListPaymentsByStatus lists payments by status
	ListPaymentsByStatus(ctx context.Context, status aggregate.PaymentStatus, page, pageSize int) ([]*aggregate.Payment, error)

	// GetTotalByCustomer gets total payment amount for a customer
	GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (valueobject.Money, error)

	// GetTotalByInvoice gets total payment amount for an invoice
	GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (valueobject.Money, error)

	// DeletePayment deletes a payment (soft delete)
	DeletePayment(ctx context.Context, paymentID uuidv7.UUID) error
}

// IPaymentRepository defines the interface for payment data access (dependency inversion)
type IPaymentRepository interface {
	// Create creates a new payment
	Create(ctx context.Context, payment *aggregate.Payment) error

	// GetByID retrieves a payment by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Payment, error)

	// GetByPaymentNo retrieves a payment by payment number
	GetByPaymentNo(ctx context.Context, paymentNo string) (*aggregate.Payment, error)

	// GetByTransactionID retrieves a payment by transaction ID
	GetByTransactionID(ctx context.Context, transactionID string) (*aggregate.Payment, error)

	// Update updates an existing payment
	Update(ctx context.Context, payment *aggregate.Payment) error

	// Delete soft deletes a payment
	Delete(ctx context.Context, id uuidv7.UUID) error

	// List retrieves payments with pagination
	List(ctx context.Context, page, pageSize int) ([]*aggregate.Payment, error)

	// ListByCustomerID retrieves payments for a specific customer
	ListByCustomerID(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Payment, error)

	// ListByInvoiceID retrieves payments for a specific invoice
	ListByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]*aggregate.Payment, error)

	// ListByStatus retrieves payments by status
	ListByStatus(ctx context.Context, status aggregate.PaymentStatus, page, pageSize int) ([]*aggregate.Payment, error)

	// GetTotalByCustomer retrieves total payment amount for a customer
	GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (int64, error)

	// GetTotalByInvoice retrieves total payment amount for an invoice
	GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (int64, error)
}

// PaymentUseCase implements the payment use case
type PaymentUseCase struct {
	repo IPaymentRepository
}

// NewPaymentUseCase creates a new payment use case
func NewPaymentUseCase(repo IPaymentRepository) IPaymentUseCase {
	return &PaymentUseCase{
		repo: repo,
	}
}

// CreatePayment creates a new payment
func (uc *PaymentUseCase) CreatePayment(ctx context.Context, customerID uuidv7.UUID, amount valueobject.Money, method aggregate.PaymentMethod) (*aggregate.Payment, error) {
	payment, err := aggregate.NewPayment(customerID, amount, method)
	if err != nil {
		return nil, err
	}

	// Generate payment number (PAY-YYYY-NNNNNN format)
	now := time.Now()
	paymentNo := fmt.Sprintf("PAY-%d-%06d", now.Year(), now.UnixNano()%1000000)
	if err := payment.SetPaymentNo(paymentNo); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (uc *PaymentUseCase) GetPayment(ctx context.Context, id uuidv7.UUID) (*aggregate.Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetPaymentByNumber retrieves a payment by payment number
func (uc *PaymentUseCase) GetPaymentByNumber(ctx context.Context, paymentNo string) (*aggregate.Payment, error) {
	return uc.repo.GetByPaymentNo(ctx, paymentNo)
}

// GetPaymentByTransactionID retrieves a payment by transaction ID
func (uc *PaymentUseCase) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*aggregate.Payment, error) {
	return uc.repo.GetByTransactionID(ctx, transactionID)
}

// LinkToInvoice links a payment to an invoice
func (uc *PaymentUseCase) LinkToInvoice(ctx context.Context, paymentID, invoiceID uuidv7.UUID) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if err := payment.LinkToInvoice(invoiceID); err != nil {
		return err
	}

	return uc.repo.Update(ctx, payment)
}

// ProcessPayment processes a pending payment
func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if err := payment.Process(); err != nil {
		return err
	}

	if transactionID != "" {
		payment.SetTransactionID(transactionID)
	}

	return uc.repo.Update(ctx, payment)
}

// CompletePayment marks a payment as completed
func (uc *PaymentUseCase) CompletePayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if err := payment.Complete(transactionID); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return err
	}

	// TODO: Publish payment.completed event via Event Bus
	// This will trigger invoice.mark_paid in Invoice context

	return nil
}

// FailPayment marks a payment as failed
func (uc *PaymentUseCase) FailPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if err := payment.Fail(reason); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return err
	}

	// TODO: Publish payment.failed event via Event Bus

	return nil
}

// RefundPayment refunds a payment (full or partial)
func (uc *PaymentUseCase) RefundPayment(ctx context.Context, paymentID uuidv7.UUID, refundAmount valueobject.Money) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if err := payment.Refund(refundAmount); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return err
	}

	// TODO: Publish payment.refunded event via Event Bus

	return nil
}

// CancelPayment cancels a pending payment
func (uc *PaymentUseCase) CancelPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if err := payment.Cancel(); err != nil {
		return err
	}

	// Add cancellation reason as note
	if reason != "" {
		payment.AddNote(fmt.Sprintf("Cancelled: %s", reason))
	}

	return uc.repo.Update(ctx, payment)
}

// SetCardDetails sets card payment details
func (uc *PaymentUseCase) SetCardDetails(ctx context.Context, paymentID uuidv7.UUID, last4, brand string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	payment.SetCardDetails(last4, brand)

	return uc.repo.Update(ctx, payment)
}

// SetProvider sets the payment provider
func (uc *PaymentUseCase) SetProvider(ctx context.Context, paymentID uuidv7.UUID, provider string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	payment.SetProvider(provider)

	return uc.repo.Update(ctx, payment)
}

// AddNote adds a note to the payment
func (uc *PaymentUseCase) AddNote(ctx context.Context, paymentID uuidv7.UUID, note string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	payment.AddNote(note)

	return uc.repo.Update(ctx, payment)
}

// ListPayments lists all payments with pagination
func (uc *PaymentUseCase) ListPayments(ctx context.Context, page, pageSize int) ([]*aggregate.Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return uc.repo.List(ctx, page, pageSize)
}

// ListPaymentsByCustomer lists payments for a specific customer
func (uc *PaymentUseCase) ListPaymentsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return uc.repo.ListByCustomerID(ctx, customerID, page, pageSize)
}

// ListPaymentsByInvoice lists payments for a specific invoice
func (uc *PaymentUseCase) ListPaymentsByInvoice(ctx context.Context, invoiceID uuidv7.UUID) ([]*aggregate.Payment, error) {
	return uc.repo.ListByInvoiceID(ctx, invoiceID)
}

// ListPaymentsByStatus lists payments by status
func (uc *PaymentUseCase) ListPaymentsByStatus(ctx context.Context, status aggregate.PaymentStatus, page, pageSize int) ([]*aggregate.Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return uc.repo.ListByStatus(ctx, status, page, pageSize)
}

// GetTotalByCustomer gets total payment amount for a customer
func (uc *PaymentUseCase) GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (valueobject.Money, error) {
	total, err := uc.repo.GetTotalByCustomer(ctx, customerID)
	if err != nil {
		return valueobject.Money{}, err
	}

	// Assume USD currency (in real system, we'd need to specify or query)
	return valueobject.NewMoney(total, "USD")
}

// GetTotalByInvoice gets total payment amount for an invoice
func (uc *PaymentUseCase) GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (valueobject.Money, error) {
	total, err := uc.repo.GetTotalByInvoice(ctx, invoiceID)
	if err != nil {
		return valueobject.Money{}, err
	}

	// Assume USD currency
	return valueobject.NewMoney(total, "USD")
}

// DeletePayment deletes a payment (soft delete)
func (uc *PaymentUseCase) DeletePayment(ctx context.Context, paymentID uuidv7.UUID) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	// Only allow deleting payments in certain statuses
	if payment.Status != aggregate.PaymentStatusPending && payment.Status != aggregate.PaymentStatusCancelled {
		return paymenterrors.ErrPaymentInvalidStatusForDeletion
	}

	return uc.repo.Delete(ctx, paymentID)
}
