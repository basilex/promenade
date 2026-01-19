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
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentCreateFailed, err)
	}

	// Generate payment number (PAY-YYYY-NNNNNN format)
	now := time.Now()
	paymentNo := fmt.Sprintf("PAY-%d-%06d", now.Year(), now.UnixNano()%1000000)
	if err := payment.SetPaymentNo(paymentNo); err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentNumberGenerationFailed, err)
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentCreateFailed, err)
	}

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (uc *PaymentUseCase) GetPayment(ctx context.Context, id uuidv7.UUID) (*aggregate.Payment, error) {
	payment, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}
	return payment, nil
}

// GetPaymentByNumber retrieves a payment by payment number
func (uc *PaymentUseCase) GetPaymentByNumber(ctx context.Context, paymentNo string) (*aggregate.Payment, error) {
	payment, err := uc.repo.GetByPaymentNo(ctx, paymentNo)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}
	return payment, nil
}

// GetPaymentByTransactionID retrieves a payment by transaction ID
func (uc *PaymentUseCase) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*aggregate.Payment, error) {
	payment, err := uc.repo.GetByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}
	return payment, nil
}

// LinkToInvoice links a payment to an invoice
func (uc *PaymentUseCase) LinkToInvoice(ctx context.Context, paymentID, invoiceID uuidv7.UUID) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	if err := payment.LinkToInvoice(invoiceID); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentLinkFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	return nil
}

// ProcessPayment processes a pending payment
func (uc *PaymentUseCase) ProcessPayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	if err := payment.Process(); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentProcessingFailed, err)
	}

	if transactionID != "" {
		payment.SetTransactionID(transactionID)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	return nil
}

// CompletePayment marks a payment as completed
func (uc *PaymentUseCase) CompletePayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	if err := payment.Complete(transactionID); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentCompletionFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	// TODO: Publish payment.completed event via Event Bus
	// This will trigger invoice.mark_paid in Invoice context

	return nil
}

// FailPayment marks a payment as failed
func (uc *PaymentUseCase) FailPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	if err := payment.Fail(reason); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentFailureFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	// TODO: Publish payment.failed event via Event Bus

	return nil
}

// RefundPayment refunds a payment (full or partial)
func (uc *PaymentUseCase) RefundPayment(ctx context.Context, paymentID uuidv7.UUID, refundAmount valueobject.Money) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	if err := payment.Refund(refundAmount); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentRefundOperationFailed, err)
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	// TODO: Publish payment.refunded event via Event Bus

	return nil
}

// CancelPayment cancels a pending payment
func (uc *PaymentUseCase) CancelPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	if err := payment.Cancel(); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentCancellationFailed, err)
	}

	// Add cancellation reason as note
	if reason != "" {
		payment.AddNote(fmt.Sprintf("Cancelled: %s", reason))
	}

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	return nil
}

// SetCardDetails sets card payment details
func (uc *PaymentUseCase) SetCardDetails(ctx context.Context, paymentID uuidv7.UUID, last4, brand string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	payment.SetCardDetails(last4, brand)

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	return nil
}

// SetProvider sets the payment provider
func (uc *PaymentUseCase) SetProvider(ctx context.Context, paymentID uuidv7.UUID, provider string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	payment.SetProvider(provider)

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	return nil
}

// AddNote adds a note to the payment
func (uc *PaymentUseCase) AddNote(ctx context.Context, paymentID uuidv7.UUID, note string) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	payment.AddNote(note)

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentUpdateFailed, err)
	}

	return nil
}

// ListPayments lists all payments with pagination
func (uc *PaymentUseCase) ListPayments(ctx context.Context, page, pageSize int) ([]*aggregate.Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	payments, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentListFailed, err)
	}

	return payments, nil
}

// ListPaymentsByCustomer lists payments for a specific customer
func (uc *PaymentUseCase) ListPaymentsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	payments, err := uc.repo.ListByCustomerID(ctx, customerID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentListFailed, err)
	}

	return payments, nil
}

// ListPaymentsByInvoice lists payments for a specific invoice
func (uc *PaymentUseCase) ListPaymentsByInvoice(ctx context.Context, invoiceID uuidv7.UUID) ([]*aggregate.Payment, error) {
	payments, err := uc.repo.ListByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentListFailed, err)
	}

	return payments, nil
}

// ListPaymentsByStatus lists payments by status
func (uc *PaymentUseCase) ListPaymentsByStatus(ctx context.Context, status aggregate.PaymentStatus, page, pageSize int) ([]*aggregate.Payment, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	payments, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentListFailed, err)
	}

	return payments, nil
}

// GetTotalByCustomer gets total payment amount for a customer
func (uc *PaymentUseCase) GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (valueobject.Money, error) {
	total, err := uc.repo.GetTotalByCustomer(ctx, customerID)
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentTotalCalculationFailed, err)
	}

	// Assume USD currency (in real system, we'd need to specify or query)
	money, err := valueobject.NewMoney(total, "USD")
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", paymenterrors.ErrMoneyCreationFailed, err)
	}
	return money, nil
}

// GetTotalByInvoice gets total payment amount for an invoice
func (uc *PaymentUseCase) GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (valueobject.Money, error) {
	total, err := uc.repo.GetTotalByInvoice(ctx, invoiceID)
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", paymenterrors.ErrPaymentTotalCalculationFailed, err)
	}

	// Assume USD currency
	money, err := valueobject.NewMoney(total, "USD")
	if err != nil {
		return valueobject.Money{}, fmt.Errorf("%w: %w", paymenterrors.ErrMoneyCreationFailed, err)
	}
	return money, nil
}

// DeletePayment deletes a payment (soft delete)
func (uc *PaymentUseCase) DeletePayment(ctx context.Context, paymentID uuidv7.UUID) error {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentGetFailed, err)
	}

	// Only allow deleting payments in certain statuses
	if payment.Status != aggregate.PaymentStatusPending && payment.Status != aggregate.PaymentStatusCancelled {
		return paymenterrors.ErrPaymentInvalidStatusForDeletion
	}

	if err := uc.repo.Delete(ctx, paymentID); err != nil {
		return fmt.Errorf("%w: %w", paymenterrors.ErrPaymentDeleteFailed, err)
	}

	return nil
}
