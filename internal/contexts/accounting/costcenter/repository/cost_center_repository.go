package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICostCenterRepository defines the interface for cost center persistence
type ICostCenterRepository interface {
	// Create creates a new cost center
	Create(ctx context.Context, costCenter *aggregate.CostCenter) error

	// GetByID retrieves a cost center by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error)

	// GetByCode retrieves a cost center by organization and code
	GetByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.CostCenter, error)

	// Update updates an existing cost center
	Update(ctx context.Context, costCenter *aggregate.CostCenter) error

	// Delete soft deletes a cost center
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ListByOrganization lists all cost centers for an organization
	ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.CostCenter, error)

	// ListChildren lists child cost centers
	ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error)

	// ListByType lists cost centers by type
	ListByType(ctx context.Context, organizationID uuidv7.UUID, centerType aggregate.CenterType) ([]*aggregate.CostCenter, error)

	// ListActive lists all active cost centers
	ListActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.CostCenter, error)
}
