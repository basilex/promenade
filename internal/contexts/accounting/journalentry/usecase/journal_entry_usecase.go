package usecase

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IJournalEntryUseCase defines the interface for journal entry business logic
type IJournalEntryUseCase interface {
	CreateJournalEntry(ctx context.Context, organizationID uuidv7.UUID, entryDate time.Time, description string, sourceType aggregate.SourceType, sourceID *uuidv7.UUID, createdBy uuidv7.UUID) (*aggregate.JournalEntry, error)
	AddLine(ctx context.Context, entryID, accountID uuidv7.UUID, debitCents, creditCents int64, currencyCode, description string, updatedBy uuidv7.UUID) (*aggregate.JournalEntry, error)
	RemoveLine(ctx context.Context, entryID, lineID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.JournalEntry, error)
	PostEntry(ctx context.Context, entryID, postedBy uuidv7.UUID) (*aggregate.JournalEntry, error)
	ReverseEntry(ctx context.Context, entryID, reversedBy uuidv7.UUID, reverseDescription string) (*aggregate.JournalEntry, error)
	GetJournalEntryByID(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error)
	GetJournalEntryByDocumentNumber(ctx context.Context, organizationID uuidv7.UUID, documentNumber string) (*aggregate.JournalEntry, error)
	UpdateDescription(ctx context.Context, entryID uuidv7.UUID, description string, updatedBy uuidv7.UUID) (*aggregate.JournalEntry, error)
	DeleteJournalEntry(ctx context.Context, id uuidv7.UUID) error
	ListJournalEntriesByOrganization(ctx context.Context, organizationID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error)
	ListJournalEntriesByPeriod(ctx context.Context, organizationID uuidv7.UUID, periodID uuidv7.UUID) ([]*aggregate.JournalEntry, error)
}

type journalEntryUseCase struct {
	journalEntryRepo repository.IJournalEntryRepository
	eventStore       *audit.EventStore
	auditLogger      *audit.AuditLogger
}

// NewJournalEntryUseCase creates a new journal entry use case
func NewJournalEntryUseCase(
	journalEntryRepo repository.IJournalEntryRepository,
	eventStore *audit.EventStore,
	auditLogger *audit.AuditLogger,
) IJournalEntryUseCase {
	return &journalEntryUseCase{
		journalEntryRepo: journalEntryRepo,
		eventStore:       eventStore,
		auditLogger:      auditLogger,
	}
}

func (uc *journalEntryUseCase) CreateJournalEntry(
	ctx context.Context,
	organizationID uuidv7.UUID,
	entryDate time.Time,
	description string,
	sourceType aggregate.SourceType,
	sourceID *uuidv7.UUID,
	createdBy uuidv7.UUID,
) (*aggregate.JournalEntry, error) {
	// Create new journal entry
	entry, err := aggregate.NewJournalEntry(organizationID, entryDate, description, sourceType, sourceID, createdBy)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.journalEntryRepo.Create(ctx, entry); err != nil {
		return nil, err
	}

	// Append event to event store
	if uc.eventStore != nil {
		_ = uc.eventStore.AppendEvent(ctx, audit.JournalEntryEvent{
			JournalEntryID: entry.ID,
			EventType:      audit.EventEntryCreated,
			EventData: map[string]interface{}{
				"entry_date":  entryDate,
				"description": description,
				"source_type": sourceType,
				"source_id":   sourceID,
			},
			OrganizationID: organizationID,
			UserID:         createdBy,
		})
	}

	// Log audit
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogCreate(ctx, audit.AuditRecord{
			EntityType:     "journal_entry",
			EntityID:       entry.ID,
			Action:         "create",
			OrganizationID: organizationID,
			UserID:         createdBy,
			Details:        map[string]interface{}{"description": description, "entry_date": entryDate},
		})
	}

	return entry, nil
}

func (uc *journalEntryUseCase) AddLine(
	ctx context.Context,
	entryID, accountID uuidv7.UUID,
	debitCents, creditCents int64,
	currencyCode, description string,
	updatedBy uuidv7.UUID,
) (*aggregate.JournalEntry, error) {
	// Get journal entry
	entry, err := uc.journalEntryRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Verify entry is in draft status
	if entry.Status != aggregate.EntryStatusDraft {
		return nil, journalentry.ErrJournalEntryAlreadyPosted
	}

	// Add line
	if err := entry.AddLine(accountID, debitCents, creditCents, currencyCode, description); err != nil {
		return nil, err
	}

	entry.LastUpdatedBy = updatedBy
	entry.Touch()

	// Save
	if err := uc.journalEntryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}

	// Append event
	if uc.eventStore != nil {
		_ = uc.eventStore.AppendEvent(ctx, audit.JournalEntryEvent{
			JournalEntryID: entry.ID,
			EventType:      audit.EventLineAdded,
			EventData: map[string]interface{}{
				"account_id":   accountID,
				"debit_cents":  debitCents,
				"credit_cents": creditCents,
				"currency":     currencyCode,
				"description":  description,
			},
			OrganizationID: entry.OrganizationID,
			UserID:         updatedBy,
		})
	}

	return entry, nil
}

func (uc *journalEntryUseCase) RemoveLine(
	ctx context.Context,
	entryID, lineID uuidv7.UUID,
	updatedBy uuidv7.UUID,
) (*aggregate.JournalEntry, error) {
	// Get journal entry
	entry, err := uc.journalEntryRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Verify entry is in draft status
	if entry.Status != aggregate.EntryStatusDraft {
		return nil, journalentry.ErrJournalEntryAlreadyPosted
	}

	// Remove line
	if err := entry.RemoveLine(lineID); err != nil {
		return nil, err
	}

	entry.LastUpdatedBy = updatedBy
	entry.Touch()

	// Save
	if err := uc.journalEntryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}

	// Append event
	if uc.eventStore != nil {
		_ = uc.eventStore.AppendEvent(ctx, audit.JournalEntryEvent{
			JournalEntryID: entry.ID,
			EventType:      audit.EventLineRemoved,
			EventData: map[string]interface{}{
				"line_id": lineID,
			},
			OrganizationID: entry.OrganizationID,
			UserID:         updatedBy,
		})
	}

	return entry, nil
}

func (uc *journalEntryUseCase) PostEntry(
	ctx context.Context,
	entryID, postedBy uuidv7.UUID,
) (*aggregate.JournalEntry, error) {
	// Get journal entry
	entry, err := uc.journalEntryRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Post entry
	if err := entry.Post(postedBy); err != nil {
		return nil, err
	}

	// Save
	if err := uc.journalEntryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}

	// Append event
	if uc.eventStore != nil {
		_ = uc.eventStore.AppendEvent(ctx, audit.JournalEntryEvent{
			JournalEntryID: entry.ID,
			EventType:      audit.EventEntryPosted,
			EventData: map[string]interface{}{
				"posted_at": entry.PostedAt,
			},
			OrganizationID: entry.OrganizationID,
			UserID:         postedBy,
		})
	}

	// Log audit
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogPost(ctx, audit.AuditRecord{
			EntityType:     "journal_entry",
			EntityID:       entry.ID,
			Action:         "post",
			OrganizationID: entry.OrganizationID,
			UserID:         postedBy,
			Details:        map[string]interface{}{"posted_at": entry.PostedAt},
		})
	}

	return entry, nil
}

func (uc *journalEntryUseCase) ReverseEntry(
	ctx context.Context,
	entryID, reversedBy uuidv7.UUID,
	reverseDescription string,
) (*aggregate.JournalEntry, error) {
	// Get journal entry
	entry, err := uc.journalEntryRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Reverse entry
	if err := entry.Reverse(reversedBy); err != nil {
		return nil, err
	}

	// Create reverse entry
	reverseEntry, err := aggregate.NewJournalEntry(
		entry.OrganizationID,
		time.Now(),
		reverseDescription,
		aggregate.SourceTypeManual,
		&entry.ID,
		reversedBy,
	)
	if err != nil {
		return nil, err
	}

	// Copy lines with inverted amounts
	for _, line := range entry.Lines {
		// Swap debit/credit
		if err := reverseEntry.AddLine(
			line.AccountID,
			line.CreditCents, // Swap: credit becomes debit
			line.DebitCents,  // Swap: debit becomes credit
			line.CurrencyCode,
			line.Description,
		); err != nil {
			return nil, err
		}
	}

	// Post reverse entry immediately
	if err := reverseEntry.Post(reversedBy); err != nil {
		return nil, err
	}

	// Save both entries
	if err := uc.journalEntryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}
	if err := uc.journalEntryRepo.Create(ctx, reverseEntry); err != nil {
		return nil, err
	}

	// Append event
	if uc.eventStore != nil {
		_ = uc.eventStore.AppendEvent(ctx, audit.JournalEntryEvent{
			JournalEntryID: entry.ID,
			EventType:      audit.EventEntryReversed,
			EventData: map[string]interface{}{
				"reverse_entry_id": reverseEntry.ID,
				"reversed_at":      entry.ReversedAt,
			},
			OrganizationID: entry.OrganizationID,
			UserID:         reversedBy,
		})
	}

	// Log audit
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogReverse(ctx, audit.AuditRecord{
			EntityType:     "journal_entry",
			EntityID:       entry.ID,
			Action:         "reverse",
			OrganizationID: entry.OrganizationID,
			UserID:         reversedBy,
			Details:        map[string]interface{}{"reverse_entry_id": reverseEntry.ID},
		})
	}

	return reverseEntry, nil
}

func (uc *journalEntryUseCase) GetJournalEntryByID(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
	return uc.journalEntryRepo.GetByID(ctx, id)
}

func (uc *journalEntryUseCase) GetJournalEntryByDocumentNumber(ctx context.Context, organizationID uuidv7.UUID, documentNumber string) (*aggregate.JournalEntry, error) {
	return uc.journalEntryRepo.GetByDocumentNumber(ctx, organizationID, documentNumber)
}

func (uc *journalEntryUseCase) UpdateDescription(ctx context.Context, entryID uuidv7.UUID, description string, updatedBy uuidv7.UUID) (*aggregate.JournalEntry, error) {
	// Get journal entry
	entry, err := uc.journalEntryRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Verify entry is in draft status
	if entry.Status != aggregate.EntryStatusDraft {
		return nil, journalentry.ErrJournalEntryAlreadyPosted
	}

	// Update description
	if err := entry.UpdateDescription(description); err != nil {
		return nil, err
	}

	entry.LastUpdatedBy = updatedBy
	entry.Touch()

	// Save
	if err := uc.journalEntryRepo.Update(ctx, entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (uc *journalEntryUseCase) DeleteJournalEntry(ctx context.Context, id uuidv7.UUID) error {
	// Get entry to verify it's in draft status
	entry, err := uc.journalEntryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Can only delete draft entries
	if entry.Status != aggregate.EntryStatusDraft {
		return journalentry.ErrJournalEntryAlreadyPosted
	}

	return uc.journalEntryRepo.Delete(ctx, id)
}

func (uc *journalEntryUseCase) ListJournalEntriesByOrganization(ctx context.Context, organizationID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error) {
	return uc.journalEntryRepo.ListByOrganization(ctx, organizationID, status)
}

func (uc *journalEntryUseCase) ListJournalEntriesByPeriod(ctx context.Context, organizationID uuidv7.UUID, periodID uuidv7.UUID) ([]*aggregate.JournalEntry, error) {
	return uc.journalEntryRepo.ListByPeriod(ctx, organizationID, periodID)
}
