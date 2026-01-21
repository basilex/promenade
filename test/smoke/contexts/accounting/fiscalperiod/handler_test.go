package fiscalperiod_test

import (
    "context"
    "fmt"
    "net/http"
    "testing"
    "time"

    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
    fiscalperiodHTTP "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/adapter/http"
    fiscalperiodAggregate "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/test/smoke"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

type MockFiscalPeriodUseCase struct {
    CreateFiscalPeriodFunc              func(ctx context.Context, organizationID uuidv7.UUID, code, name string, periodType fiscalperiodAggregate.PeriodType, startDate, endDate time.Time, createdBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error)
    GetFiscalPeriodByIDFunc             func(ctx context.Context, id uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error)
    GetCurrentPeriodFunc                func(ctx context.Context, organizationID uuidv7.UUID, date string) (*fiscalperiodAggregate.FiscalPeriod, error)
    ClosePeriodFunc                     func(ctx context.Context, id, closedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error)
    ReopenPeriodFunc                    func(ctx context.Context, id, reopenedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error)
    LockPeriodFunc                      func(ctx context.Context, id, lockedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error)
    DeleteFiscalPeriodFunc              func(ctx context.Context, id uuidv7.UUID) error
    ListFiscalPeriodsByOrganizationFunc func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*fiscalperiodAggregate.FiscalPeriod, error)
    ListFiscalPeriodsByYearFunc         func(ctx context.Context, organizationID uuidv7.UUID, fiscalYear int) ([]*fiscalperiodAggregate.FiscalPeriod, error)
    ListOpenPeriodsFunc                 func(ctx context.Context, organizationID uuidv7.UUID) ([]*fiscalperiodAggregate.FiscalPeriod, error)
}

func (m *MockFiscalPeriodUseCase) CreateFiscalPeriod(ctx context.Context, organizationID uuidv7.UUID, code, name string, periodType fiscalperiodAggregate.PeriodType, startDate, endDate time.Time, createdBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.CreateFiscalPeriodFunc != nil {
        return m.CreateFiscalPeriodFunc(ctx, organizationID, code, name, periodType, startDate, endDate, createdBy)
    }
    return nil, fmt.Errorf("CreateFiscalPeriodFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) GetFiscalPeriodByID(ctx context.Context, id uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.GetFiscalPeriodByIDFunc != nil {
        return m.GetFiscalPeriodByIDFunc(ctx, id)
    }
    return nil, fmt.Errorf("GetFiscalPeriodByIDFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) GetCurrentPeriod(ctx context.Context, organizationID uuidv7.UUID, date string) (*fiscalperiodAggregate.FiscalPeriod, error) {
	if m.GetCurrentPeriodFunc != nil {
		return m.GetCurrentPeriodFunc(ctx, organizationID, date)
	}
	return nil, fmt.Errorf("GetCurrentPeriodFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) ClosePeriod(ctx context.Context, id, closedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.ClosePeriodFunc != nil {
        return m.ClosePeriodFunc(ctx, id, closedBy)
    }
    return nil, fmt.Errorf("ClosePeriodFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) ReopenPeriod(ctx context.Context, id, reopenedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.ReopenPeriodFunc != nil {
        return m.ReopenPeriodFunc(ctx, id, reopenedBy)
    }
    return nil, fmt.Errorf("ReopenPeriodFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) LockPeriod(ctx context.Context, id, lockedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.LockPeriodFunc != nil {
        return m.LockPeriodFunc(ctx, id, lockedBy)
    }
    return nil, fmt.Errorf("LockPeriodFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) DeleteFiscalPeriod(ctx context.Context, id uuidv7.UUID) error {
    if m.DeleteFiscalPeriodFunc != nil {
        return m.DeleteFiscalPeriodFunc(ctx, id)
    }
    return fmt.Errorf("DeleteFiscalPeriodFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) ListFiscalPeriodsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.ListFiscalPeriodsByOrganizationFunc != nil {
        return m.ListFiscalPeriodsByOrganizationFunc(ctx, organizationID, limit, offset)
    }
    return nil, fmt.Errorf("ListFiscalPeriodsByOrganizationFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) ListFiscalPeriodsByYear(ctx context.Context, organizationID uuidv7.UUID, fiscalYear int) ([]*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.ListFiscalPeriodsByYearFunc != nil {
        return m.ListFiscalPeriodsByYearFunc(ctx, organizationID, fiscalYear)
    }
    return nil, fmt.Errorf("ListFiscalPeriodsByYearFunc not implemented")
}

func (m *MockFiscalPeriodUseCase) ListOpenPeriods(ctx context.Context, organizationID uuidv7.UUID) ([]*fiscalperiodAggregate.FiscalPeriod, error) {
    if m.ListOpenPeriodsFunc != nil {
        return m.ListOpenPeriodsFunc(ctx, organizationID)
    }
    return nil, fmt.Errorf("ListOpenPeriodsFunc not implemented")
}

func setupRouter(mockUC *MockFiscalPeriodUseCase) *gin.Engine {
    router := smoke.SetupRouter()
    handler := fiscalperiodHTTP.NewFiscalPeriodHandler(mockUC)

    router.Use(func(c *gin.Context) {
        c.Set("organization_id", uuidv7.New().String())
        c.Set("user_id", uuidv7.New().String())
        c.Next()
    })

    handler.RegisterRoutes(router.Group("/api/v1/accounting"))
    return router
}

func TestCreateFiscalPeriod_Success(t *testing.T) {
	periodID := uuidv7.New()

	mockUC := &MockFiscalPeriodUseCase{
		CreateFiscalPeriodFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, periodType fiscalperiodAggregate.PeriodType, startDate, endDate time.Time, createdBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
			fp, _ := fiscalperiodAggregate.NewFiscalPeriod(organizationID, code, name, periodType, startDate, endDate, createdBy)
			fp.ID = periodID
			return fp, nil
		},
	}

	router := setupRouter(mockUC)

    body := map[string]any{
        "code":        "2024-01",
        "name":        "January 2024",
        "period_type": "month",
        "start_date":  "2024-01-01",
        "end_date":    "2024-01-31",
    }

    w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/fiscal-periods", body)
    assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateFiscalPeriod_ValidationError(t *testing.T) {
	mockUC := &MockFiscalPeriodUseCase{
		CreateFiscalPeriodFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, periodType fiscalperiodAggregate.PeriodType, startDate, endDate time.Time, createdBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
			return nil, fiscalperiod.ErrPeriodCodeEmpty
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"code":        "",
		"name":        "January 2024",
		"period_type": "month",
		"start_date":  "2024-01-01",
		"end_date":    "2024-01-31",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/fiscal-periods", body)
    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetFiscalPeriod_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    periodID := uuidv7.New()
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    fp, _ := fiscalperiodAggregate.NewFiscalPeriod(orgID, "2024-01", "January 2024", fiscalperiodAggregate.PeriodTypeMonth, startDate, endDate, userID)
    fp.ID = periodID

    mockUC := &MockFiscalPeriodUseCase{
        GetFiscalPeriodByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
            return fp, nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/fiscal-periods/"+periodID.String(), nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetFiscalPeriod_NotFound(t *testing.T) {
    periodID := uuidv7.New()

    mockUC := &MockFiscalPeriodUseCase{
        GetFiscalPeriodByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
            return nil, fiscalperiod.ErrPeriodNotFound
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/fiscal-periods/"+periodID.String(), nil)
    assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestClosePeriod_Success(t *testing.T) {
    periodID := uuidv7.New()
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    fp, _ := fiscalperiodAggregate.NewFiscalPeriod(uuidv7.New(), "2024-01", "January 2024", fiscalperiodAggregate.PeriodTypeMonth, startDate, endDate, uuidv7.New())
    fp.ID = periodID

    mockUC := &MockFiscalPeriodUseCase{
        ClosePeriodFunc: func(ctx context.Context, id, closedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
            return fp, nil
        },
    }

    router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/fiscal-periods/"+periodID.String()+"/close", nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestReopenPeriod_Success(t *testing.T) {
    periodID := uuidv7.New()
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    fp, _ := fiscalperiodAggregate.NewFiscalPeriod(uuidv7.New(), "2024-01", "January 2024", fiscalperiodAggregate.PeriodTypeMonth, startDate, endDate, uuidv7.New())
    fp.ID = periodID

    mockUC := &MockFiscalPeriodUseCase{
        ReopenPeriodFunc: func(ctx context.Context, id, reopenedBy uuidv7.UUID) (*fiscalperiodAggregate.FiscalPeriod, error) {
            return fp, nil
        },
    }

    router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/fiscal-periods/"+periodID.String()+"/reopen", nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestListFiscalPeriods_Success(t *testing.T) {
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    periods := []*fiscalperiodAggregate.FiscalPeriod{
        func() *fiscalperiodAggregate.FiscalPeriod {
			fp, _ := fiscalperiodAggregate.NewFiscalPeriod(uuidv7.New(), "2024-01", "January 2024", fiscalperiodAggregate.PeriodTypeMonth, startDate, endDate, uuidv7.New())
			return fp
		}(),
		func() *fiscalperiodAggregate.FiscalPeriod {
			fp, _ := fiscalperiodAggregate.NewFiscalPeriod(uuidv7.New(), "2024-Q1", "Q1 2024", fiscalperiodAggregate.PeriodTypeQuarter, startDate, endDate, uuidv7.New())
			return fp
		}(),
    }

    mockUC := &MockFiscalPeriodUseCase{
        ListFiscalPeriodsByOrganizationFunc: func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*fiscalperiodAggregate.FiscalPeriod, error) {
            return periods, nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/fiscal-periods", nil)
    assert.Equal(t, http.StatusOK, w.Code)
}