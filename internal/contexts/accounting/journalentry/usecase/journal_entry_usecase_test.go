package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockJournalEntryRepository struct {
	CreateFunc              func(ctx context.Context, je *aggregate.JournalEntry) error
	GetByIDFunc             func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error)
	GetByNumberFunc         func(ctx context.Context, orgID uuidv7.UUID, number string) (*aggregate.JournalEntry, error)
	GetByDocumentNumberFunc func(ctx context.Context, organizationID uuidv7.UUID, documentNumber string) (*aggregate.JournalEntry, error)
	UpdateFunc              func(ctx context.Context, je *aggregate.JournalEntry) error
	DeleteFunc              func(ctx context.Context, id uuidv7.UUID) error
	ListByOrganizationFunc  func(ctx context.Context, orgID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error)
	ListByStatusFunc        func(ctx context.Context, orgID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error)
	ListByPeriodFunc        func(ctx context.Context, organizationID uuidv7.UUID, periodID uuidv7.UUID) ([]*aggregate.JournalEntry, error)
	ListByDateRangeFunc     func(ctx context.Context, orgID uuidv7.UUID, startDate, endDate time.Time) ([]*aggregate.JournalEntry, error)
}

func (m *MockJournalEntryRepository) Create(ctx context.Context, je *aggregate.JournalEntry) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, je)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockJournalEntryRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockJournalEntryRepository) GetByNumber(ctx context.Context, orgID uuidv7.UUID, number string) (*aggregate.JournalEntry, error) {
	if m.GetByNumberFunc != nil {
		return m.GetByNumberFunc(ctx, orgID, number)
	}
	return nil, errors.New("GetByNumberFunc not implemented")
}

func (m *MockJournalEntryRepository) GetByDocumentNumber(ctx context.Context, organizationID uuidv7.UUID, documentNumber string) (*aggregate.JournalEntry, error) {
	if m.GetByDocumentNumberFunc != nil {
		return m.GetByDocumentNumberFunc(ctx, organizationID, documentNumber)
	}
	return nil, errors.New("GetByDocumentNumberFunc not implemented")
}

func (m *MockJournalEntryRepository) Update(ctx context.Context, je *aggregate.JournalEntry) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, je)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockJournalEntryRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockJournalEntryRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, orgID, status)
	}
	return nil, errors.New("ListByOrganizationFunc not implemented")
}

func (m *MockJournalEntryRepository) ListByStatus(ctx context.Context, orgID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, orgID, status)
	}
	return nil, errors.New("ListByStatusFunc not implemented")
}

func (m *MockJournalEntryRepository) ListByPeriod(ctx context.Context, organizationID uuidv7.UUID, periodID uuidv7.UUID) ([]*aggregate.JournalEntry, error) {
	if m.ListByPeriodFunc != nil {
		return m.ListByPeriodFunc(ctx, organizationID, periodID)
	}
	return nil, errors.New("ListByPeriodFunc not implemented")
}

func (m *MockJournalEntryRepository) ListByDateRange(ctx context.Context, orgID uuidv7.UUID, startDate, endDate time.Time) ([]*aggregate.JournalEntry, error) {
	if m.ListByDateRangeFunc != nil {
		return m.ListByDateRangeFunc(ctx, orgID, startDate, endDate)
	}
	return nil, errors.New("ListByDateRangeFunc not implemented")
}

// ============================================================================
// Tests
// ============================================================================

func TestCreateJournalEntry_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	mockRepo := &MockJournalEntryRepository{
		GetByNumberFunc: func(ctx context.Context, orgID uuidv7.UUID, number string) (*aggregate.JournalEntry, error) {
			return nil, journalentry.ErrJournalEntryNotFound
		},
		CreateFunc: func(ctx context.Context, je *aggregate.JournalEntry) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	je, err := uc.CreateJournalEntry(context.Background(), orgID, entryDate, "Opening balance", aggregate.SourceTypeManual, nil, userID)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, je.ID)
	assert.Equal(t, "Opening balance", je.Description)
	assert.Equal(t, entryDate, je.EntryDate)
	assert.Equal(t, aggregate.EntryStatusDraft, je.Status)
}

func TestCreateJournalEntry_DuplicateNumber(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	existing, _ := aggregate.NewJournalEntry(orgID, entryDate, "Existing", aggregate.SourceTypeManual, nil, userID)

	mockRepo := &MockJournalEntryRepository{
		GetByNumberFunc: func(ctx context.Context, orgID uuidv7.UUID, number string) (*aggregate.JournalEntry, error) {
			return existing, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	_, err := uc.CreateJournalEntry(context.Background(), orgID, entryDate, "Duplicate", aggregate.SourceTypeManual, nil, userID)
	assert.Error(t, err)
}

func TestAddLine_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Test Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.JournalEntry) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	updated, err := uc.AddLine(context.Background(), je.ID, accountID, 100000, 0, "USD", "Debit line", userID)
	require.NoError(t, err)
	assert.Len(t, updated.Lines, 1)
	assert.Equal(t, accountID, updated.Lines[0].AccountID)
	assert.Equal(t, int64(100000), updated.Lines[0].DebitCents)
}

func TestAddLine_CannotModifyPosted(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Posted Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()
	_ = je.AddLine(uuidv7.New(), 50000, 0, "USD", "Initial debit")
	_ = je.AddLine(uuidv7.New(), 0, 50000, "USD", "Initial credit")
	_ = je.Post(userID)

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	_, err := uc.AddLine(context.Background(), je.ID, accountID, 10000, 0, "USD", "Cannot add", userID)
	assert.Error(t, err)
	assert.Equal(t, journalentry.ErrJournalEntryAlreadyPosted, err)
}

func TestRemoveLine_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Test Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()
	_ = je.AddLine(uuidv7.New(), 50000, 0, "USD", "Line 1")
	_ = je.AddLine(uuidv7.New(), 0, 50000, "USD", "Line 2")
	lineToRemove := je.Lines[0].ID

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.JournalEntry) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	updated, err := uc.RemoveLine(context.Background(), je.ID, lineToRemove, userID)
	require.NoError(t, err)
	assert.Len(t, updated.Lines, 1)
}

func TestPostJournalEntry_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Test Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()
	_ = je.AddLine(uuidv7.New(), 100000, 0, "USD", "Debit")
	_ = je.AddLine(uuidv7.New(), 0, 100000, "USD", "Credit")

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.JournalEntry) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	posted, err := uc.PostEntry(context.Background(), je.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.EntryStatusPosted, posted.Status)
	assert.NotNil(t, posted.PostedAt)
	assert.Equal(t, userID, *posted.PostedBy)
}

func TestPostJournalEntry_Unbalanced(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Unbalanced Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()
	_ = je.AddLine(uuidv7.New(), 100000, 0, "USD", "Debit")
	_ = je.AddLine(uuidv7.New(), 0, 50000, "USD", "Credit")

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	_, err := uc.PostEntry(context.Background(), je.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, journalentry.ErrJournalEntryCannotPostDraft, err)
}

func TestReverseJournalEntry_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Test Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()
	_ = je.AddLine(uuidv7.New(), 100000, 0, "USD", "Debit")
	_ = je.AddLine(uuidv7.New(), 0, 100000, "USD", "Credit")
	_ = je.Post(userID)

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
		CreateFunc: func(ctx context.Context, reversal *aggregate.JournalEntry) error {
			return nil
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.JournalEntry) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	reversal, err := uc.ReverseEntry(context.Background(), je.ID, userID, "Reversal reason")
	require.NoError(t, err)
	assert.Equal(t, aggregate.EntryStatusPosted, reversal.Status)
	assert.Len(t, reversal.Lines, 2)
	assert.Equal(t, je.ID, *reversal.SourceID)
}

func TestReverseJournalEntry_NotPosted(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Draft Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()
	_ = je.AddLine(uuidv7.New(), 100000, 0, "USD", "Debit")
	_ = je.AddLine(uuidv7.New(), 0, 100000, "USD", "Credit")

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	_, err := uc.ReverseEntry(context.Background(), je.ID, userID, "Cannot reverse")
	assert.Error(t, err)
	assert.Equal(t, journalentry.ErrJournalEntryNotPosted, err)
}

func TestGetJournalEntryByID_Success(t *testing.T) {
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	je, _ := aggregate.NewJournalEntry(uuidv7.New(), entryDate, "Test Entry", aggregate.SourceTypeManual, nil, uuidv7.New())
	je.ID = uuidv7.New()

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	retrieved, err := uc.GetJournalEntryByID(context.Background(), je.ID)
	require.NoError(t, err)
	assert.Equal(t, je.ID, retrieved.ID)
}

func TestGetJournalEntryByID_NotFound(t *testing.T) {
	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			return nil, journalentry.ErrJournalEntryNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	_, err := uc.GetJournalEntryByID(context.Background(), uuidv7.New())
	assert.Error(t, err)
	assert.Equal(t, journalentry.ErrJournalEntryNotFound, err)
}

func TestGetJournalEntryByDocumentNumber_Success(t *testing.T) {
	orgID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Test Entry", aggregate.SourceTypeManual, nil, uuidv7.New())

	mockRepo := &MockJournalEntryRepository{
		GetByDocumentNumberFunc: func(ctx context.Context, orgID uuidv7.UUID, number string) (*aggregate.JournalEntry, error) {
			if number == "JE-2025-001" {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	retrieved, err := uc.GetJournalEntryByDocumentNumber(context.Background(), orgID, "JE-2025-001")
	require.NoError(t, err)
	assert.Equal(t, je.ID, retrieved.ID)
}

func TestDeleteJournalEntry_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Draft Entry", aggregate.SourceTypeManual, nil, userID)
	je.ID = uuidv7.New()

	mockRepo := &MockJournalEntryRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.JournalEntry, error) {
			if id == je.ID {
				return je, nil
			}
			return nil, journalentry.ErrJournalEntryNotFound
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	err := uc.DeleteJournalEntry(context.Background(), je.ID)
	require.NoError(t, err)
}

func TestListJournalEntriesByOrganization_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je1, _ := aggregate.NewJournalEntry(orgID, entryDate, "Entry 1", aggregate.SourceTypeManual, nil, userID)
	je2, _ := aggregate.NewJournalEntry(orgID, entryDate, "Entry 2", aggregate.SourceTypeManual, nil, userID)

	mockRepo := &MockJournalEntryRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error) {
			return []*aggregate.JournalEntry{je1, je2}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	entries, err := uc.ListJournalEntriesByOrganization(context.Background(), orgID, aggregate.EntryStatusDraft)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestListJournalEntriesByStatus_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	postedJE, _ := aggregate.NewJournalEntry(orgID, entryDate, "Posted Entry", aggregate.SourceTypeManual, nil, userID)
	_ = postedJE.AddLine(uuidv7.New(), 100000, 0, "USD", "Debit")
	_ = postedJE.AddLine(uuidv7.New(), 0, 100000, "USD", "Credit")
	_ = postedJE.Post(userID)

	mockRepo := &MockJournalEntryRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, status aggregate.EntryStatus) ([]*aggregate.JournalEntry, error) {
			if status == aggregate.EntryStatusPosted {
				return []*aggregate.JournalEntry{postedJE}, nil
			}
			return []*aggregate.JournalEntry{}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	postedEntries, err := uc.ListJournalEntriesByOrganization(context.Background(), orgID, aggregate.EntryStatusPosted)
	require.NoError(t, err)
	assert.Len(t, postedEntries, 1)
	assert.Equal(t, aggregate.EntryStatusPosted, postedEntries[0].Status)
}

func TestListJournalEntriesByPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	periodID := uuidv7.New()
	entryDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	je, _ := aggregate.NewJournalEntry(orgID, entryDate, "Period Entry", aggregate.SourceTypeManual, nil, userID)

	mockRepo := &MockJournalEntryRepository{
		ListByPeriodFunc: func(ctx context.Context, orgID uuidv7.UUID, fpID uuidv7.UUID) ([]*aggregate.JournalEntry, error) {
			if fpID == periodID {
				return []*aggregate.JournalEntry{je}, nil
			}
			return []*aggregate.JournalEntry{}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	var eventStore *audit.EventStore = nil
	uc := NewJournalEntryUseCase(mockRepo, eventStore, auditLogger)

	entries, err := uc.ListJournalEntriesByPeriod(context.Background(), orgID, periodID)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, orgID, entries[0].OrganizationID)
}
