package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/cache"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockFiscalPeriodRepository struct {
	CreateFunc                   func(ctx context.Context, fp *aggregate.FiscalPeriod) error
	GetByIDFunc                  func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error)
	GetByNameFunc                func(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.FiscalPeriod, error)
	GetByOrganizationAndDateFunc func(ctx context.Context, organizationID uuidv7.UUID, date string) (*aggregate.FiscalPeriod, error)
	UpdateFunc                   func(ctx context.Context, fp *aggregate.FiscalPeriod) error
	DeleteFunc                   func(ctx context.Context, id uuidv7.UUID) error
	ListByOrganizationFunc       func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error)
	ListByYearFunc               func(ctx context.Context, orgID uuidv7.UUID, fiscalYear int) ([]*aggregate.FiscalPeriod, error)
	ListByStatusFunc             func(ctx context.Context, orgID uuidv7.UUID, status aggregate.PeriodStatus) ([]*aggregate.FiscalPeriod, error)
	ListOpenPeriodsFunc          func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error)
	ListOpenFunc                 func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error)
	GetCurrentPeriodFunc         func(ctx context.Context, orgID uuidv7.UUID) (*aggregate.FiscalPeriod, error)
}

func (m *MockFiscalPeriodRepository) Create(ctx context.Context, fp *aggregate.FiscalPeriod) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, fp)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockFiscalPeriodRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockFiscalPeriodRepository) GetByName(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.FiscalPeriod, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, orgID, name)
	}
	return nil, errors.New("GetByNameFunc not implemented")
}

func (m *MockFiscalPeriodRepository) Update(ctx context.Context, fp *aggregate.FiscalPeriod) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, fp)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockFiscalPeriodRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockFiscalPeriodRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, orgID, limit, offset)
	}
	return nil, errors.New("ListByOrganizationFunc not implemented")
}

func (m *MockFiscalPeriodRepository) ListByStatus(ctx context.Context, orgID uuidv7.UUID, status aggregate.PeriodStatus) ([]*aggregate.FiscalPeriod, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, orgID, status)
	}
	return nil, errors.New("ListByStatusFunc not implemented")
}

func (m *MockFiscalPeriodRepository) ListOpenPeriods(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error) {
	if m.ListOpenPeriodsFunc != nil {
		return m.ListOpenPeriodsFunc(ctx, orgID)
	}
	return nil, errors.New("ListOpenPeriodsFunc not implemented")
}

func (m *MockFiscalPeriodRepository) GetCurrentPeriod(ctx context.Context, orgID uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	if m.GetCurrentPeriodFunc != nil {
		return m.GetCurrentPeriodFunc(ctx, orgID)
	}
	return nil, errors.New("GetCurrentPeriodFunc not implemented")
}

func (m *MockFiscalPeriodRepository) GetByOrganizationAndDate(ctx context.Context, organizationID uuidv7.UUID, date string) (*aggregate.FiscalPeriod, error) {
	if m.GetByOrganizationAndDateFunc != nil {
		return m.GetByOrganizationAndDateFunc(ctx, organizationID, date)
	}
	return nil, errors.New("GetByOrganizationAndDateFunc not implemented")
}

func (m *MockFiscalPeriodRepository) ListByYear(ctx context.Context, orgID uuidv7.UUID, fiscalYear int) ([]*aggregate.FiscalPeriod, error) {
	if m.ListByYearFunc != nil {
		return m.ListByYearFunc(ctx, orgID, fiscalYear)
	}
	return nil, errors.New("ListByYearFunc not implemented")
}

func (m *MockFiscalPeriodRepository) ListOpen(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error) {
	if m.ListOpenFunc != nil {
		return m.ListOpenFunc(ctx, orgID)
	}
	return nil, errors.New("ListOpenFunc not implemented")
}

// ============================================================================
// Tests
// ============================================================================

func TestCreateFiscalPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	mockRepo := &MockFiscalPeriodRepository{
		GetByNameFunc: func(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.FiscalPeriod, error) {
			return nil, fiscalperiod.ErrPeriodNotFound
		},
		CreateFunc: func(ctx context.Context, fp *aggregate.FiscalPeriod) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	fp, err := uc.CreateFiscalPeriod(context.Background(), orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, fp.ID)
	assert.Equal(t, "Fiscal Year 2025", fp.Name)
	assert.Equal(t, startDate, fp.StartDate)
	assert.Equal(t, endDate, fp.EndDate)
	assert.Equal(t, aggregate.PeriodStatusOpen, fp.Status)
}

func TestCreateFiscalPeriod_DuplicateName(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	existing, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)

	mockRepo := &MockFiscalPeriodRepository{
		GetByNameFunc: func(ctx context.Context, orgID uuidv7.UUID, name string) (*aggregate.FiscalPeriod, error) {
			return existing, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	_, err := uc.CreateFiscalPeriod(context.Background(), orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	assert.Error(t, err)
	// This test expects repository-level duplicate constraint error
}

func TestGetFiscalPeriodByID_Success(t *testing.T) {
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	fp, _ := aggregate.NewFiscalPeriod(uuidv7.New(), "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, uuidv7.New())
	fp.ID = uuidv7.New()

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	retrieved, err := uc.GetFiscalPeriodByID(context.Background(), fp.ID)
	require.NoError(t, err)
	assert.Equal(t, fp.ID, retrieved.ID)
	assert.Equal(t, "Fiscal Year 2025", retrieved.Name)
}

func TestGetFiscalPeriodByID_NotFound(t *testing.T) {
	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			return nil, fiscalperiod.ErrPeriodNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	_, err := uc.GetFiscalPeriodByID(context.Background(), uuidv7.New())
	assert.Error(t, err)
	assert.Equal(t, fiscalperiod.ErrPeriodNotFound, err)
}

func TestGetFiscalPeriodByName_Success(t *testing.T) {
	orgID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, uuidv7.New())

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	retrieved, err := uc.GetFiscalPeriodByID(context.Background(), fp.ID)
	require.NoError(t, err)
	assert.Equal(t, "Fiscal Year 2025", retrieved.Name)
}

func TestOpenFiscalPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp.ID = uuidv7.New()
	_ = fp.Close(userID) // Close it first before reopening

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.FiscalPeriod) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	opened, err := uc.ReopenPeriod(context.Background(), fp.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.PeriodStatusOpen, opened.Status)
	assert.NotNil(t, opened.ReopenedAt)
}

func TestCloseFiscalPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp.ID = uuidv7.New()
	_ = fp.Reopen(userID)

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.FiscalPeriod) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	closed, err := uc.ClosePeriod(context.Background(), fp.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.PeriodStatusClosed, closed.Status)
	assert.NotNil(t, closed.ClosedAt)
}

func TestCloseFiscalPeriod_NotOpen(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp.ID = uuidv7.New()
	_ = fp.Close(userID) // Close it first

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	_, err := uc.ClosePeriod(context.Background(), fp.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, fiscalperiod.ErrPeriodAlreadyClosed, err)
}

func TestLockFiscalPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp.ID = uuidv7.New()
	_ = fp.Reopen(userID)
	_ = fp.Close(userID)

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.FiscalPeriod) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	locked, err := uc.LockPeriod(context.Background(), fp.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.PeriodStatusLocked, locked.Status)
	assert.NotNil(t, locked.LockedAt)
}

func TestLockFiscalPeriod_NotClosed(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp.ID = uuidv7.New()
	_ = fp.Reopen(userID)

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	_, err := uc.LockPeriod(context.Background(), fp.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, fiscalperiod.ErrPeriodNotClosed, err)
}

func TestReopenFiscalPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp.ID = uuidv7.New()
	_ = fp.Reopen(userID)
	_ = fp.Close(userID)

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.FiscalPeriod) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	reopened, err := uc.ReopenPeriod(context.Background(), fp.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, aggregate.PeriodStatusOpen, reopened.Status)
}

func TestDeleteFiscalPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp.ID = uuidv7.New()

	mockRepo := &MockFiscalPeriodRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
			if id == fp.ID {
				return fp, nil
			}
			return nil, fiscalperiod.ErrPeriodNotFound
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	err := uc.DeleteFiscalPeriod(context.Background(), fp.ID)
	require.NoError(t, err)
}

func TestListFiscalPeriodsByOrganization_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp1, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	fp2, _ := aggregate.NewFiscalPeriod(orgID, "FY2026", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)

	mockRepo := &MockFiscalPeriodRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error) {
			return []*aggregate.FiscalPeriod{fp1, fp2}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	periods, err := uc.ListFiscalPeriodsByOrganization(context.Background(), orgID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, periods, 2)
}

func TestListFiscalPeriodsByStatus_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	openFP, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	_ = openFP.Reopen(userID)

	mockRepo := &MockFiscalPeriodRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error) {
			return []*aggregate.FiscalPeriod{openFP}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	openPeriods, err := uc.ListFiscalPeriodsByOrganization(context.Background(), orgID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, openPeriods, 1)
	assert.Equal(t, aggregate.PeriodStatusOpen, openPeriods[0].Status)
}

func TestListOpenFiscalPeriods_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	_ = fp.Reopen(userID)

	mockRepo := &MockFiscalPeriodRepository{
		ListOpenFunc: func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error) {
			return []*aggregate.FiscalPeriod{fp}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	openPeriods, err := uc.ListOpenPeriods(context.Background(), orgID)
	require.NoError(t, err)
	assert.Len(t, openPeriods, 1)
	assert.Equal(t, aggregate.PeriodStatusOpen, openPeriods[0].Status)
}

func TestGetCurrentFiscalPeriod_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	fp, _ := aggregate.NewFiscalPeriod(orgID, "FY2025", "Fiscal Year 2025", aggregate.PeriodTypeYear, startDate, endDate, userID)
	_ = fp.Reopen(userID)

	mockRepo := &MockFiscalPeriodRepository{
		GetByOrganizationAndDateFunc: func(ctx context.Context, organizationID uuidv7.UUID, date string) (*aggregate.FiscalPeriod, error) {
			return fp, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewFiscalPeriodUseCase(mockRepo, cache.NewFiscalPeriodCache(mockRepo), auditLogger)

	current, err := uc.GetCurrentPeriod(context.Background(), orgID, "2025-06-15")
	require.NoError(t, err)
	assert.Equal(t, "Fiscal Year 2025", current.Name)
	assert.Equal(t, aggregate.PeriodStatusOpen, current.Status)
}
