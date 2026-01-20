package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IJournalEntryRepository defines the repository interface for JournalEntry aggregate
type IJournalEntryRepository interface {
	Create(ctx context.Context, entry *aggregate.JournalEntry) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error)
	GetByDocumentNumber(ctx context.Context, organizationID uuidv7.UUID, documentNumber string) (*aggregate.JournalEntry, error)
	Update(ctx context.Context, entry *aggregate.JournalEntry) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error)
	ListByPeriod(ctx context.Context, organizationID uuidv7.UUID, periodID uuidv7.UUID) ([]*aggregate.JournalEntry, error)
}
