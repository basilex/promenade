package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IReconciliationRepository defines the interface for bank reconciliation persistence
type IReconciliationRepository interface {
	// Create creates a new bank reconciliation
	Create(ctx context.Context, reconciliation *aggregate.Reconciliation) error

	// GetByID retrieves a bank reconciliation by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error)

	// Update updates an existing bank reconciliation
	Update(ctx context.Context, reconciliation *aggregate.Reconciliation) error

	// Delete soft deletes a bank reconciliation
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ListByOrganization lists all reconciliations for an organization
	ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Reconciliation, error)

	// ListByBankAccount lists reconciliations for a specific bank account
	ListByBankAccount(ctx context.Context, bankAccountID uuidv7.UUID) ([]*aggregate.Reconciliation, error)

	// ListByStatus lists reconciliations by status
	ListByStatus(ctx context.Context, organizationID uuidv7.UUID, status reconciliation.Status) ([]*aggregate.Reconciliation, error)

	// ListByDateRange lists reconciliations within a date range
	ListByDateRange(ctx context.Context, organizationID uuidv7.UUID, startDate, endDate string) ([]*aggregate.Reconciliation, error)
}
