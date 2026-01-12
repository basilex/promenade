package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IUseCase defines the interface for payment business logic
type IUseCase interface {
	// CreatePayment creates a new payment
	CreatePayment(ctx context.Context, customerID uuidv7.UUID, amount valueobject.Money, method PaymentMethod) (*Payment, error)

	// GetPayment retrieves a payment by ID
	GetPayment(ctx context.Context, id uuidv7.UUID) (*Payment, error)

	// GetPaymentByNumber retrieves a payment by payment number
	GetPaymentByNumber(ctx context.Context, paymentNo string) (*Payment, error)

	// GetPaymentByTransactionID retrieves a payment by transaction ID
	GetPaymentByTransactionID(ctx context.Context, transactionID string) (*Payment, error)

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
	ListPayments(ctx context.Context, page, pageSize int) ([]*Payment, error)

	// ListPaymentsByCustomer lists payments for a specific customer
	ListPaymentsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Payment, error)

	// ListPaymentsByInvoice lists payments for a specific invoice
	ListPaymentsByInvoice(ctx context.Context, invoiceID uuidv7.UUID) ([]*Payment, error)

	// ListPaymentsByStatus lists payments by status
	ListPaymentsByStatus(ctx context.Context, status PaymentStatus, page, pageSize int) ([]*Payment, error)

	// GetTotalByCustomer gets total payment amount for a customer
	GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (valueobject.Money, error)

	// GetTotalByInvoice gets total payment amount for an invoice
	GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (valueobject.Money, error)

	// DeletePayment deletes a payment (soft delete)
	DeletePayment(ctx context.Context, paymentID uuidv7.UUID) error
}

// useCase implements the payment use case
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new payment use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreatePayment creates a new payment
func (uc *useCase) CreatePayment(ctx context.Context, customerID uuidv7.UUID, amount valueobject.Money, method PaymentMethod) (*Payment, error) {
	payment, err := NewPayment(customerID, amount, method)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentCreateFailed, err)
	}

	// Generate payment number (PAY-YYYY-NNNNNN format)
	now := time.Now()
	paymentNo := fmt.Sprintf("PAY-%d-%06d", now.Year(), now.UnixNano()%1000000)
	if err := payment.SetPaymentNo(paymentNo); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentNumberGenerationFailed, err)
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentCreateFailed, err)
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (uc *useCase) GetPayment(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
	payment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}
	return payment, nil
}

// GetPaymentByNumber retrieves a payment by payment number
func (uc *useCase) GetPaymentByNumber(ctx context.Context, paymentNo string) (*Payment, error) {
	payment, err := uc.repo.GetByPaymentNo(ctx, paymentNo)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}
	return payment, nil
}

// GetPaymentByTransactionID retrieves a payment by transaction ID
func (uc *useCase) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*Payment, error) {
	payment, err := uc.repo.GetByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}
	return payment, nil
}

// LinkToInvoice links a payment to an invoice
func (uc *useCase) LinkToInvoice(ctx context.Context, paymentID, invoiceID uuidv7.UUID) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	if err := payment.LinkToInvoice(invoiceID); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentLinkFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	return nil
}

// ProcessPayment processes a pending payment
func (uc *useCase) ProcessPayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	if err := payment.Process(); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentProcessingFailed, err)
	}

	if transactionID != "" {
		payment.SetTransactionID(transactionID)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	return nil
}

// CompletePayment marks a payment as completed
func (uc *useCase) CompletePayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	if err := payment.Complete(transactionID); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentCompletionFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	// TODO: Publish payment.completed event via Event Bus
	// This will trigger invoice.mark_paid in Invoice context

	return nil
}

// FailPayment marks a payment as failed
func (uc *useCase) FailPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	if err := payment.Fail(reason); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentFailureFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	// TODO: Publish payment.failed event via Event Bus

	return nil
}

// RefundPayment refunds a payment (full or partial)
func (uc *useCase) RefundPayment(ctx context.Context, paymentID uuidv7.UUID, refundAmount valueobject.Money) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	if err := payment.Refund(refundAmount); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentRefundOperationFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	// TODO: Publish payment.refunded event via Event Bus

	return nil
}

// CancelPayment cancels a pending payment
func (uc *useCase) CancelPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	if err := payment.Cancel(); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentCancellationFailed, err)
	}

	// Add cancellation reason as note
	if reason != "" {
		payment.AddNote(fmt.Sprintf("Cancelled: %s", reason))
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	return nil
}

// SetCardDetails sets card payment details
func (uc *useCase) SetCardDetails(ctx context.Context, paymentID uuidv7.UUID, last4, brand string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	payment.SetCardDetails(last4, brand)

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	return nil
}

// SetProvider sets the payment provider
func (uc *useCase) SetProvider(ctx context.Context, paymentID uuidv7.UUID, provider string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	payment.SetProvider(provider)

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	return nil
}

// AddNote adds a note to the payment
func (uc *useCase) AddNote(ctx context.Context, paymentID uuidv7.UUID, note string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	payment.AddNote(note)

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentUpdateFailed, err)
	}

	return nil
}

// ListPayments lists all payments with pagination
func (uc *useCase) ListPayments(ctx context.Context, page, pageSize int) ([]*Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	payments, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentListFailed, err)
	}

	return payments, nil
}

// ListPaymentsByCustomer lists payments for a specific customer
func (uc *useCase) ListPaymentsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	payments, err := uc.repo.ListByCustomerID(ctx, customerID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentListFailed, err)
	}

	return payments, nil
}

// ListPaymentsByInvoice lists payments for a specific invoice
func (uc *useCase) ListPaymentsByInvoice(ctx context.Context, invoiceID uuidv7.UUID) ([]*Payment, error) {
	payments, err := uc.repo.ListByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentListFailed, err)
	}

	return payments, nil
}

// ListPaymentsByStatus lists payments by status
func (uc *useCase) ListPaymentsByStatus(ctx context.Context, status PaymentStatus, page, pageSize int) ([]*Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	payments, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPaymentListFailed, err)
	}

	return payments, nil
}

// GetTotalByCustomer gets total payment amount for a customer
func (uc *useCase) GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (valueobject.Money, error) {
	total, err := uc.repo.GetTotalByCustomer(ctx, customerID)
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", ErrPaymentTotalCalculationFailed, err)
	}

	// Assume USD currency (in real system, we'd need to specify or query)
	money, err := valueobject.NewMoney(total, "USD")
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", ErrMoneyCreationFailed, err)
	}
	return money, nil
}

// GetTotalByInvoice gets total payment amount for an invoice
func (uc *useCase) GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (valueobject.Money, error) {
	total, err := uc.repo.GetTotalByInvoice(ctx, invoiceID)
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", ErrPaymentTotalCalculationFailed, err)
	}

	// Assume USD currency
	money, err := valueobject.NewMoney(total, "USD")
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", ErrMoneyCreationFailed, err)
	}
	return money, nil
}

// DeletePayment deletes a payment (soft delete)
func (uc *useCase) DeletePayment(ctx context.Context, paymentID uuidv7.UUID) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentGetFailed, err)
	}

	// Only allow deleting payments in certain statuses
	if payment.Status != PaymentStatusPending && payment.Status != PaymentStatusCancelled {
		return ErrPaymentInvalidStatusForDeletion
	}

	if err := uc.repo.Delete(ctx, paymentID); err != nil {
		return fmt.Errorf("%w: %w", ErrPaymentDeleteFailed, err)
	}

	return nil
}
