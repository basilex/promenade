package repository

import (
"context"

"github.com/basilex/promenade/internal/contexts/banking/bankaccount/aggregate"
"github.com/basilex/promenade/pkg/uuidv7"
)

// IBankAccountRepository defines the interface for bank account persistence
type IBankAccountRepository interface {
	Create(ctx context.Context, account *aggregate.BankAccount) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.BankAccount, error)
	Update(ctx context.Context, account *aggregate.BankAccount) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	List(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.BankAccount, error)
	GetByProviderAccountID(ctx context.Context, provider aggregate.BankProvider, providerAccountID string) (*aggregate.BankAccount, error)
	CountByOrganization(ctx context.Context, organizationID uuidv7.UUID) (int, error)
}
