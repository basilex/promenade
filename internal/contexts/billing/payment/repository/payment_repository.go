package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/billing/payment/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IPaymentRepository defines the interface for payment data access
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
