package invoice

import (
	"context"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for invoice persistence operations
type IRepository interface {
	// CRUD operations
	Create(ctx context.Context, invoice *Invoice) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*Invoice, error)
	GetByInvoiceNo(ctx context.Context, invoiceNo string) (*Invoice, error)
	Update(ctx context.Context, invoice *Invoice) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Query operations
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Invoice, int, error)
	ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*Invoice, error)
	ListByStatus(ctx context.Context, status InvoiceStatus, page, pageSize int) ([]*Invoice, int, error)
	ListOverdue(ctx context.Context, page, pageSize int) ([]*Invoice, int, error)
	List(ctx context.Context, page, pageSize int) ([]*Invoice, int, error)

	// Aggregation operations
	CountByStatus(ctx context.Context, status InvoiceStatus) (int, error)
	GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error)

	// Line item operations
	CreateLine(ctx context.Context, line *InvoiceLine) error
	DeleteLine(ctx context.Context, lineID uuidv7.UUID) error
	GetLinesByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]InvoiceLine, error)

	// Utility
	ExistsByInvoiceNo(ctx context.Context, invoiceNo string) (bool, error)
	GenerateInvoiceNumber(ctx context.Context) (string, error)
}
