package repository

import (
    "context"

    "github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// IBudgetRepository defines the interface for budget persistence
type IBudgetRepository interface {
    // Create creates a new budget
    Create(ctx context.Context, budget *aggregate.Budget) error

    // GetByID retrieves a budget by ID
    GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error)

    // GetByName retrieves a budget by organization and name
    GetByName(ctx context.Context, organizationID uuidv7.UUID, name string) (*aggregate.Budget, error)

    // Update updates an existing budget
    Update(ctx context.Context, budget *aggregate.Budget) error

    // Delete soft deletes a budget
    Delete(ctx context.Context, id uuidv7.UUID) error

    // ListByOrganization lists all budgets for an organization
    ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Budget, error)

    // ListByStatus lists budgets by status
    ListByStatus(ctx context.Context, organizationID uuidv7.UUID, status aggregate.BudgetStatus) ([]*aggregate.Budget, error)

    // ListActive lists all active budgets
    ListActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.Budget, error)
}