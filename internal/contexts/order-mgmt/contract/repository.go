package contract

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IContractRepository defines contract data access operations
type IContractRepository interface {
	// CRUD operations
	Create(ctx context.Context, contract *Contract) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*Contract, error)
	Update(ctx context.Context, contract *Contract) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	List(ctx context.Context, offset, limit int) ([]*Contract, int, error)

	// Business queries
	GetByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*Contract, error)
	GetByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*Contract, error)
	GetActiveContracts(ctx context.Context) ([]*Contract, error)

	// Filtering
	ListByStatus(ctx context.Context, status ContractStatus, offset, limit int) ([]*Contract, int, error)
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, offset, limit int) ([]*Contract, int, error)
	ListExpiringSoon(ctx context.Context, days int) ([]*Contract, error)
}
