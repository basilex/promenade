package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IInvoiceRepository defines methods for invoice persistence
type IInvoiceRepository interface {
	// Create creates a new invoice
	Create(ctx context.Context, invoice *entity.Invoice) error

	// GetByID retrieves an invoice by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Invoice, error)

	// GetByInvoiceNumber retrieves an invoice by invoice number
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entity.Invoice, error)

	// GetBySubscriptionID retrieves invoices for a subscription
	GetBySubscriptionID(ctx context.Context, subscriptionID uuidv7.UUID) ([]*entity.Invoice, error)

	// GetOverdue retrieves overdue invoices
	GetOverdue(ctx context.Context, limit int) ([]*entity.Invoice, error)

	// List retrieves invoices with optional status filter
	List(ctx context.Context, status *entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error)

	// Count counts invoices with optional status filter
	Count(ctx context.Context, status *entity.InvoiceStatus) (int, error)

	// Update updates an existing invoice
	Update(ctx context.Context, invoice *entity.Invoice) error

	// Delete soft-deletes an invoice
	Delete(ctx context.Context, id uuidv7.UUID) error
}
