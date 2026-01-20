package repository

import (
    "context"

    "github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// IAccountRepository defines the repository interface for Account aggregate
type IAccountRepository interface {
    // Single operations
    Create(ctx context.Context, account *aggregate.Account) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error)
    GetByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.Account, error)
    Update(ctx context.Context, account *aggregate.Account) error
    Delete(ctx context.Context, id uuidv7.UUID) error
    
    // Batch operations (for seeding and imports)
    CreateMany(ctx context.Context, accounts []*aggregate.Account) error
    UpdateMany(ctx context.Context, accounts []*aggregate.Account) error
    
    // Query operations
    ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, includeInactive bool) ([]*aggregate.Account, error)
    ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error)
    ListByType(ctx context.Context, organizationID uuidv7.UUID, accountType aggregate.AccountType) ([]*aggregate.Account, error)
    
    // Cache operations
    GetAllActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.Account, error)
}