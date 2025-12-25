package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IPlanRepository defines methods for plan persistence
type IPlanRepository interface {
	// Create creates a new plan
	Create(ctx context.Context, plan *entity.Plan) error

	// GetByID retrieves a plan by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Plan, error)

	// GetBySlug retrieves a plan by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Plan, error)

	// List retrieves plans with optional status filter
	List(ctx context.Context, status *entity.PlanStatus, limit, offset int) ([]*entity.Plan, error)

	// Count counts plans with optional status filter
	Count(ctx context.Context, status *entity.PlanStatus) (int, error)

	// Update updates an existing plan
	Update(ctx context.Context, plan *entity.Plan) error

	// Delete soft-deletes a plan
	Delete(ctx context.Context, id uuidv7.UUID) error
}
