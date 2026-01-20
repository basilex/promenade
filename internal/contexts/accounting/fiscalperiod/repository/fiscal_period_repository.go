package repository

import (
    "context"

    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// IFiscalPeriodRepository defines the interface for fiscal period persistence
type IFiscalPeriodRepository interface {
    // Create creates a new fiscal period
    Create(ctx context.Context, period *aggregate.FiscalPeriod) error

    // GetByID retrieves a fiscal period by ID
    GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error)

    // GetByOrganizationAndDate finds a fiscal period containing a specific date
    GetByOrganizationAndDate(ctx context.Context, organizationID uuidv7.UUID, date string) (*aggregate.FiscalPeriod, error)

    // Update updates an existing fiscal period
    Update(ctx context.Context, period *aggregate.FiscalPeriod) error

    // Delete soft deletes a fiscal period
    Delete(ctx context.Context, id uuidv7.UUID) error

    // ListByOrganization lists all fiscal periods for an organization
    ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error)

    // ListByYear lists all fiscal periods for a specific fiscal year
    ListByYear(ctx context.Context, organizationID uuidv7.UUID, fiscalYear int) ([]*aggregate.FiscalPeriod, error)

    // ListOpen lists all open fiscal periods
    ListOpen(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error)
}