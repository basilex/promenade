package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IPaymentRepository defines methods for payment persistence
type IPaymentRepository interface {
	// Create creates a new payment
	Create(ctx context.Context, payment *entity.Payment) error

	// GetByID retrieves a payment by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Payment, error)

	// GetByTransactionID retrieves a payment by transaction ID
	GetByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error)

	// GetByInvoiceID retrieves payments for an invoice
	GetByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]*entity.Payment, error)

	// GetByUserID retrieves payments for a user
	GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*entity.Payment, error)

	// List retrieves payments with optional status filter
	List(ctx context.Context, status *entity.PaymentStatus, limit, offset int) ([]*entity.Payment, error)

	// Count counts payments with optional status filter
	Count(ctx context.Context, status *entity.PaymentStatus) (int, error)

	// Update updates an existing payment
	Update(ctx context.Context, payment *entity.Payment) error

	// Delete soft-deletes a payment
	Delete(ctx context.Context, id uuidv7.UUID) error
}
