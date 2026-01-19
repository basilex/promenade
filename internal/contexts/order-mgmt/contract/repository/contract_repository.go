package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IContractRepository defines contract data access operations
type IContractRepository interface {
	// CRUD operations
	Create(ctx context.Context, contract *aggregate.Contract) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Contract, error)
	Update(ctx context.Context, contract *aggregate.Contract) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	List(ctx context.Context, offset, limit int) ([]*aggregate.Contract, int, error)

	// Business queries
	GetByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*aggregate.Contract, error)
	GetByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*aggregate.Contract, error)
	GetActiveContracts(ctx context.Context) ([]*aggregate.Contract, error)

	// Filtering
	ListByStatus(ctx context.Context, status aggregate.ContractStatus, offset, limit int) ([]*aggregate.Contract, int, error)
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, offset, limit int) ([]*aggregate.Contract, int, error)
	ListExpiringSoon(ctx context.Context, days int) ([]*aggregate.Contract, error)
}
