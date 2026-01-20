package repository

import (
    "context"

    "github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// ITaxCodeRepository defines the interface for tax code persistence
type ITaxCodeRepository interface {
    // Create creates a new tax code
    Create(ctx context.Context, taxCode *aggregate.TaxCode) error

    // GetByID retrieves a tax code by ID
    GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error)

    // GetByCode retrieves a tax code by organization and code
    GetByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.TaxCode, error)

    // Update updates an existing tax code
    Update(ctx context.Context, taxCode *aggregate.TaxCode) error

    // Delete soft deletes a tax code
    Delete(ctx context.Context, id uuidv7.UUID) error

    // ListByOrganization lists all tax codes for an organization
    ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.TaxCode, error)

    // ListByType lists tax codes by type
    ListByType(ctx context.Context, organizationID uuidv7.UUID, taxType aggregate.TaxType) ([]*aggregate.TaxCode, error)

    // ListActive lists all active tax codes
    ListActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.TaxCode, error)
}