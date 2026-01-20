package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IBankTransactionRepository defines the interface for bank transaction persistence
type IBankTransactionRepository interface {
	Create(ctx context.Context, transaction *aggregate.BankTransaction) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.BankTransaction, error)
	Update(ctx context.Context, transaction *aggregate.BankTransaction) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	ListByAccount(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error)
	GetByExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string) (*aggregate.BankTransaction, error)
	ListUnmatched(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error)
	CountByAccount(ctx context.Context, accountID uuidv7.UUID) (int, error)
}
