package company

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for company data access
type IRepository interface {
	// Create creates a new company
	Create(ctx context.Context, company *Company) error

	// GetByID retrieves a company by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Company, error)

	// GetByName retrieves a company by name
	GetByName(ctx context.Context, name string) (*Company, error)

	// GetByTaxID retrieves a company by tax ID
	GetByTaxID(ctx context.Context, taxID string) (*Company, error)

	// List retrieves a paginated list of companies
	List(ctx context.Context, page, pageSize int) ([]*Company, int, error)

	// ListByIndustry retrieves companies by industry
	ListByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*Company, int, error)

	// ListBySize retrieves companies by size
	ListBySize(ctx context.Context, size string, page, pageSize int) ([]*Company, int, error)

	// ListSubsidiaries retrieves subsidiaries of a parent company
	ListSubsidiaries(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*Company, error)

	// Update updates a company
	Update(ctx context.Context, company *Company) error

	// Delete soft deletes a company
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Exists checks if a company exists
	Exists(ctx context.Context, id uuidv7.UUID) (bool, error)

	// ExistsByName checks if a company with given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)
}
