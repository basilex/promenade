package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IReceiptRepository defines receipt data access operations
type IReceiptRepository interface {
	// Create creates a new receipt
	Create(ctx context.Context, receipt *aggregate.Receipt) error

	// GetByID retrieves receipt by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error)

	// GetByOrderID retrieves receipt by order ID
	GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error)

	// List retrieves receipts with filters
	List(ctx context.Context, filters *ListFilters) ([]*aggregate.Receipt, error)

	// Update updates receipt
	Update(ctx context.Context, receipt *aggregate.Receipt) error

	// Delete soft deletes receipt
	Delete(ctx context.Context, id uuidv7.UUID) error
}

// ListFilters defines filters for listing receipts
type ListFilters struct {
	CashRegisterID *uuidv7.UUID
	OrderID        *uuidv7.UUID
	Status         *aggregate.ReceiptStatus
}
