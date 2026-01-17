package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IDealRepository defines operations for deal persistence
type IDealRepository interface {
	// Create creates a new deal
	Create(ctx context.Context, deal *aggregate.Deal) error

	// GetByID retrieves a deal by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Deal, error)

	// Update updates a deal
	Update(ctx context.Context, deal *aggregate.Deal) error

	// Delete soft deletes a deal
	Delete(ctx context.Context, id uuidv7.UUID) error

	// List returns paginated deals
	List(ctx context.Context, page, pageSize int) ([]*aggregate.Deal, int64, error)

	// ListByStage returns deals in a specific stage
	ListByStage(ctx context.Context, stage aggregate.DealStage, page, pageSize int) ([]*aggregate.Deal, int64, error)

	// ListByCustomer returns deals for a specific customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Deal, int64, error)

	// ListByCompany returns deals for a specific company
	ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*aggregate.Deal, int64, error)

	// ListByAssignedTo returns deals assigned to a sales rep
	ListByAssignedTo(ctx context.Context, userID uuidv7.UUID, page, pageSize int) ([]*aggregate.Deal, int64, error)

	// ListBySource returns deals from a specific source
	ListBySource(ctx context.Context, source aggregate.DealSource, page, pageSize int) ([]*aggregate.Deal, int64, error)

	// GetPipelineStats returns deal counts by stage
	GetPipelineStats(ctx context.Context) (map[aggregate.DealStage]int64, error)

	// GetTotalValue returns total value of all active deals
	GetTotalValue(ctx context.Context) (int64, error)

	// GetWonDeals returns won deals count and value
	GetWonDeals(ctx context.Context) (int64, int64, error)

	// Exists checks if a deal exists
	Exists(ctx context.Context, id uuidv7.UUID) (bool, error)
}
