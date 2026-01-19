package repository

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/contexts/billing/invoice/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IInvoiceRepository defines the interface for invoice persistence operations
type IInvoiceRepository interface {
	// CRUD operations
	Create(ctx context.Context, invoice *aggregate.Invoice) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Invoice, error)
	GetByInvoiceNo(ctx context.Context, invoiceNo string) (*aggregate.Invoice, error)
	Update(ctx context.Context, invoice *aggregate.Invoice) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Query operations
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Invoice, int, error)
	ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*aggregate.Invoice, error)
	ListByStatus(ctx context.Context, status aggregate.InvoiceStatus, page, pageSize int) ([]*aggregate.Invoice, int, error)
	ListOverdue(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error)
	List(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error)

	// Aggregation operations
	CountByStatus(ctx context.Context, status aggregate.InvoiceStatus) (int, error)
	GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error)

	// Line item operations
	CreateLine(ctx context.Context, line *aggregate.InvoiceLine) error
	DeleteLine(ctx context.Context, lineID uuidv7.UUID) error
	GetLinesByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]aggregate.InvoiceLine, error)

	// Utility
	ExistsByInvoiceNo(ctx context.Context, invoiceNo string) (bool, error)
	GenerateInvoiceNumber(ctx context.Context) (string, error)
}
