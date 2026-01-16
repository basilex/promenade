package receipt

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines receipt data access operations
type IRepository interface {
	// Create creates a new receipt
	Create(ctx context.Context, receipt *Receipt) error

	// GetByID retrieves receipt by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Receipt, error)

	// GetByOrderID retrieves receipt by order ID
	GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*Receipt, error)

	// List retrieves receipts with filters
	List(ctx context.Context, filters *ListFilters) ([]*Receipt, error)

	// Update updates receipt
	Update(ctx context.Context, receipt *Receipt) error

	// Delete soft deletes receipt
	Delete(ctx context.Context, id uuidv7.UUID) error
}

// ListFilters defines filters for listing receipts
type ListFilters struct {
	CashRegisterID *uuidv7.UUID
	OrderID        *uuidv7.UUID
	Status         *ReceiptStatus
}
