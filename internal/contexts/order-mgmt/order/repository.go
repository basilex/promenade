package order

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for order data access
type IRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *Order) error

	// GetByID retrieves an order by its ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Order, error)

	// GetByOrderNumber retrieves an order by its order number
	GetByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *Order) error

	// Delete soft-deletes an order
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ListByCustomerID retrieves all orders for a customer
	ListByCustomerID(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Order, int64, error)

	// ListByStatus retrieves all orders with a specific status
	ListByStatus(ctx context.Context, status OrderStatus, page, pageSize int) ([]*Order, int64, error)

	// List retrieves all orders with pagination
	List(ctx context.Context, page, pageSize int) ([]*Order, int64, error)

	// GetLines retrieves all line items for an order
	GetLines(ctx context.Context, orderID uuidv7.UUID) ([]OrderLine, error)

	// CreateLine creates a new order line
	CreateLine(ctx context.Context, line *OrderLine) error

	// UpdateLine updates an order line
	UpdateLine(ctx context.Context, line *OrderLine) error

	// DeleteLine deletes an order line
	DeleteLine(ctx context.Context, lineID uuidv7.UUID) error
}
