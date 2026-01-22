package journalentry_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/journalentry"
	journalentryHTTP "github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/http"
	journalentryAggregate "github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockJournalEntryUseCase struct {
	CreateJournalEntryFunc               func(ctx context.Context, organizationID uuidv7.UUID, entryDate time.Time, description string, sourceType journalentryAggregate.SourceType, sourceID *uuidv7.UUID, createdBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error)
	AddLineFunc                          func(ctx context.Context, entryID, accountID uuidv7.UUID, debitCents, creditCents int64, currencyCode, description string, updatedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error)
	RemoveLineFunc                       func(ctx context.Context, entryID, lineID uuidv7.UUID, updatedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error)
	PostEntryFunc                        func(ctx context.Context, entryID, postedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error)
	ReverseEntryFunc                     func(ctx context.Context, entryID, reversedBy uuidv7.UUID, reverseDescription string) (*journalentryAggregate.JournalEntry, error)
	GetJournalEntryByIDFunc              func(ctx context.Context, id uuidv7.UUID) (*journalentryAggregate.JournalEntry, error)
	GetJournalEntryByDocumentNumberFunc  func(ctx context.Context, organizationID uuidv7.UUID, documentNumber string) (*journalentryAggregate.JournalEntry, error)
	UpdateDescriptionFunc                func(ctx context.Context, entryID uuidv7.UUID, description string, updatedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error)
	DeleteJournalEntryFunc               func(ctx context.Context, id uuidv7.UUID) error
	ListJournalEntriesByOrganizationFunc func(ctx context.Context, organizationID uuidv7.UUID, status journalentryAggregate.EntryStatus) ([]*journalentryAggregate.JournalEntry, error)
	ListJournalEntriesByPeriodFunc       func(ctx context.Context, organizationID uuidv7.UUID, periodID uuidv7.UUID) ([]*journalentryAggregate.JournalEntry, error)
}

func (m *MockJournalEntryUseCase) CreateJournalEntry(ctx context.Context, organizationID uuidv7.UUID, entryDate time.Time, description string, sourceType journalentryAggregate.SourceType, sourceID *uuidv7.UUID, createdBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
	if m.CreateJournalEntryFunc != nil {
		return m.CreateJournalEntryFunc(ctx, organizationID, entryDate, description, sourceType, sourceID, createdBy)
	}
	return nil, fmt.Errorf("CreateJournalEntryFunc not implemented")
}

func (m *MockJournalEntryUseCase) AddLine(ctx context.Context, entryID, accountID uuidv7.UUID, debitCents, creditCents int64, currencyCode, description string, updatedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
	if m.AddLineFunc != nil {
		return m.AddLineFunc(ctx, entryID, accountID, debitCents, creditCents, currencyCode, description, updatedBy)
	}
	return nil, fmt.Errorf("AddLineFunc not implemented")
}

func (m *MockJournalEntryUseCase) RemoveLine(ctx context.Context, entryID, lineID uuidv7.UUID, updatedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
	if m.RemoveLineFunc != nil {
		return m.RemoveLineFunc(ctx, entryID, lineID, updatedBy)
	}
	return nil, fmt.Errorf("RemoveLineFunc not implemented")
}

func (m *MockJournalEntryUseCase) PostEntry(ctx context.Context, entryID, postedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
	if m.PostEntryFunc != nil {
		return m.PostEntryFunc(ctx, entryID, postedBy)
	}
	return nil, fmt.Errorf("PostEntryFunc not implemented")
}

func (m *MockJournalEntryUseCase) ReverseEntry(ctx context.Context, entryID, reversedBy uuidv7.UUID, reverseDescription string) (*journalentryAggregate.JournalEntry, error) {
	if m.ReverseEntryFunc != nil {
		return m.ReverseEntryFunc(ctx, entryID, reversedBy, reverseDescription)
	}
	return nil, fmt.Errorf("ReverseEntryFunc not implemented")
}

func (m *MockJournalEntryUseCase) GetJournalEntryByID(ctx context.Context, id uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
	if m.GetJournalEntryByIDFunc != nil {
		return m.GetJournalEntryByIDFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetJournalEntryByIDFunc not implemented")
}

func (m *MockJournalEntryUseCase) GetJournalEntryByDocumentNumber(ctx context.Context, organizationID uuidv7.UUID, documentNumber string) (*journalentryAggregate.JournalEntry, error) {
	if m.GetJournalEntryByDocumentNumberFunc != nil {
		return m.GetJournalEntryByDocumentNumberFunc(ctx, organizationID, documentNumber)
	}
	return nil, fmt.Errorf("GetJournalEntryByDocumentNumberFunc not implemented")
}

func (m *MockJournalEntryUseCase) UpdateDescription(ctx context.Context, entryID uuidv7.UUID, description string, updatedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
	if m.UpdateDescriptionFunc != nil {
		return m.UpdateDescriptionFunc(ctx, entryID, description, updatedBy)
	}
	return nil, fmt.Errorf("UpdateDescriptionFunc not implemented")
}

func (m *MockJournalEntryUseCase) DeleteJournalEntry(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteJournalEntryFunc != nil {
		return m.DeleteJournalEntryFunc(ctx, id)
	}
	return fmt.Errorf("DeleteJournalEntryFunc not implemented")
}

func (m *MockJournalEntryUseCase) ListJournalEntriesByOrganization(ctx context.Context, organizationID uuidv7.UUID, status journalentryAggregate.EntryStatus) ([]*journalentryAggregate.JournalEntry, error) {
	if m.ListJournalEntriesByOrganizationFunc != nil {
		return m.ListJournalEntriesByOrganizationFunc(ctx, organizationID, status)
	}
	return nil, fmt.Errorf("ListJournalEntriesByOrganizationFunc not implemented")
}

func (m *MockJournalEntryUseCase) ListJournalEntriesByPeriod(ctx context.Context, organizationID uuidv7.UUID, periodID uuidv7.UUID) ([]*journalentryAggregate.JournalEntry, error) {
	if m.ListJournalEntriesByPeriodFunc != nil {
		return m.ListJournalEntriesByPeriodFunc(ctx, organizationID, periodID)
	}
	return nil, fmt.Errorf("ListJournalEntriesByPeriodFunc not implemented")
}

func setupRouter(mockUC *MockJournalEntryUseCase) *gin.Engine {
	router := smoke.SetupRouter()
	handler := journalentryHTTP.NewJournalEntryHandler(mockUC)

	router.Use(func(c *gin.Context) {
		c.Set("organization_id", uuidv7.New().String())
		c.Set("user_id", uuidv7.New().String())
		c.Next()
	})

	handler.RegisterRoutes(router.Group("/api/v1/accounting"))
	return router
}

func TestCreateJournalEntry_Success(t *testing.T) {
	entryID := uuidv7.New()
	entryDate := time.Now()

	mockUC := &MockJournalEntryUseCase{
		CreateJournalEntryFunc: func(ctx context.Context, organizationID uuidv7.UUID, ed time.Time, description string, sourceType journalentryAggregate.SourceType, sourceID *uuidv7.UUID, createdBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
			je, _ := journalentryAggregate.NewJournalEntry(organizationID, ed, description, sourceType, sourceID, createdBy)
			je.ID = entryID
			return je, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"entry_date":  entryDate.Format("2006-01-02"),
		"description": "Test entry",
		"source_type": "manual",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/journal-entries", body)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateJournalEntry_ValidationError(t *testing.T) {
	mockUC := &MockJournalEntryUseCase{
		CreateJournalEntryFunc: func(ctx context.Context, organizationID uuidv7.UUID, ed time.Time, description string, sourceType journalentryAggregate.SourceType, sourceID *uuidv7.UUID, createdBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
			return nil, journalentry.ErrJournalEntryDescriptionEmpty
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"entry_date":  time.Now().Format("2006-01-02"),
		"description": "",
		"source_type": "manual",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/journal-entries", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetJournalEntry_Success(t *testing.T) {
	entryID := uuidv7.New()
	entryDate := time.Now()

	je, _ := journalentryAggregate.NewJournalEntry(uuidv7.New(), entryDate, "Test entry", journalentryAggregate.SourceTypeManual, nil, uuidv7.New())
	je.ID = entryID

	mockUC := &MockJournalEntryUseCase{
		GetJournalEntryByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
			return je, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/journal-entries/"+entryID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetJournalEntry_NotFound(t *testing.T) {
	entryID := uuidv7.New()

	mockUC := &MockJournalEntryUseCase{
		GetJournalEntryByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
			return nil, journalentry.ErrJournalEntryNotFound
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/journal-entries/"+entryID.String(), nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddJournalEntryLine_Success(t *testing.T) {
	entryID := uuidv7.New()
	accountID := uuidv7.New()
	entryDate := time.Now()

	je, _ := journalentryAggregate.NewJournalEntry(uuidv7.New(), entryDate, "Test entry", journalentryAggregate.SourceTypeManual, nil, uuidv7.New())
	je.ID = entryID

	mockUC := &MockJournalEntryUseCase{
		AddLineFunc: func(ctx context.Context, eID, accID uuidv7.UUID, debitCents, creditCents int64, currencyCode, description string, updatedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
			return je, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"account_id":    accountID.String(),
		"debit_cents":   10000,
		"credit_cents":  0,
		"currency_code": "UAH",
		"description":   "Debit line",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/journal-entries/"+entryID.String()+"/lines", body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostJournalEntry_Success(t *testing.T) {
	entryID := uuidv7.New()
	entryDate := time.Now()

	je, _ := journalentryAggregate.NewJournalEntry(uuidv7.New(), entryDate, "Test entry", journalentryAggregate.SourceTypeManual, nil, uuidv7.New())
	je.ID = entryID

	mockUC := &MockJournalEntryUseCase{
		PostEntryFunc: func(ctx context.Context, eID, postedBy uuidv7.UUID) (*journalentryAggregate.JournalEntry, error) {
			return je, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/journal-entries/"+entryID.String()+"/post", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReverseJournalEntry_Success(t *testing.T) {
	entryID := uuidv7.New()
	entryDate := time.Now()

	je, _ := journalentryAggregate.NewJournalEntry(uuidv7.New(), entryDate, "Test entry", journalentryAggregate.SourceTypeManual, nil, uuidv7.New())
	je.ID = entryID

	mockUC := &MockJournalEntryUseCase{
		ReverseEntryFunc: func(ctx context.Context, eID, reversedBy uuidv7.UUID, reverseDescription string) (*journalentryAggregate.JournalEntry, error) {
			return je, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"reverse_description": "Reversal entry",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/journal-entries/"+entryID.String()+"/reverse", body)
	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListJournalEntries_Success(t *testing.T) {
	entryDate := time.Now()

	entries := []*journalentryAggregate.JournalEntry{
		func() *journalentryAggregate.JournalEntry {
			je, _ := journalentryAggregate.NewJournalEntry(uuidv7.New(), entryDate, "Entry 1", journalentryAggregate.SourceTypeManual, nil, uuidv7.New())
			return je
		}(),
		func() *journalentryAggregate.JournalEntry {
			je, _ := journalentryAggregate.NewJournalEntry(uuidv7.New(), entryDate, "Entry 2", journalentryAggregate.SourceTypeManual, nil, uuidv7.New())
			return je
		}(),
	}

	mockUC := &MockJournalEntryUseCase{
		ListJournalEntriesByOrganizationFunc: func(ctx context.Context, organizationID uuidv7.UUID, status journalentryAggregate.EntryStatus) ([]*journalentryAggregate.JournalEntry, error) {
			return entries, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/journal-entries", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}
