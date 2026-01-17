package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICompanyRepository defines the interface for company data access
type ICompanyRepository interface {
	// Create creates a new company
	Create(ctx context.Context, company *aggregate.Company) error

	// GetByID retrieves a company by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Company, error)

	// GetByName retrieves a company by name
	GetByName(ctx context.Context, name string) (*aggregate.Company, error)

	// GetByTaxID retrieves a company by tax ID
	GetByTaxID(ctx context.Context, taxID string) (*aggregate.Company, error)

	// List retrieves a paginated list of companies
	List(ctx context.Context, page, pageSize int) ([]*aggregate.Company, int, error)

	// ListByIndustry retrieves companies by industry
	ListByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*aggregate.Company, int, error)

	// ListBySize retrieves companies by size
	ListBySize(ctx context.Context, size string, page, pageSize int) ([]*aggregate.Company, int, error)

	// ListSubsidiaries retrieves subsidiaries of a parent company
	ListSubsidiaries(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*aggregate.Company, error)

	// Update updates a company
	Update(ctx context.Context, company *aggregate.Company) error

	// Delete soft deletes a company
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Exists checks if a company exists
	Exists(ctx context.Context, id uuidv7.UUID) (bool, error)

	// ExistsByName checks if a company with given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)
}
