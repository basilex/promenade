package reconciliation_test

import (
    "context"
    "fmt"
    "net/http"
    "testing"
    "time"

    "github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
    reconciliationHTTP "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/adapter/http"
    reconciliationAggregate "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/test/smoke"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

type MockReconciliationUseCase struct {
    CreateReconciliationFunc            func(ctx context.Context, organizationID, bankAccountID, accountID uuidv7.UUID, reconciliationDate, statementDate time.Time, bankStatementBalanceCents, bookBalanceCents int64, currencyCode string, createdBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error)
    AddReconciliationItemFunc           func(ctx context.Context, reconciliationID uuidv7.UUID, transactionType reconciliation.TransactionType, transactionID *uuidv7.UUID, transactionDate time.Time, description string, amountCents int64, notes string, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error)
    RemoveReconciliationItemFunc        func(ctx context.Context, reconciliationID, itemID uuidv7.UUID, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error)
    MarkItemMatchedFunc                 func(ctx context.Context, reconciliationID, itemID uuidv7.UUID, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error)
    CompleteReconciliationFunc          func(ctx context.Context, reconciliationID, reconciledBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error)
    ReopenReconciliationFunc            func(ctx context.Context, reconciliationID, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error)
    GetReconciliationByIDFunc           func(ctx context.Context, id uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error)
    DeleteReconciliationFunc            func(ctx context.Context, id uuidv7.UUID) error
    ListReconciliationsByBankAccountFunc func(ctx context.Context, bankAccountID uuidv7.UUID) ([]*reconciliationAggregate.Reconciliation, error)
    ListReconciliationsByOrganizationFunc func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*reconciliationAggregate.Reconciliation, error)
    ListReconciliationsByStatusFunc     func(ctx context.Context, organizationID uuidv7.UUID, status reconciliation.Status) ([]*reconciliationAggregate.Reconciliation, error)
}

func (m *MockReconciliationUseCase) CreateReconciliation(ctx context.Context, organizationID, bankAccountID, accountID uuidv7.UUID, reconciliationDate, statementDate time.Time, bankStatementBalanceCents, bookBalanceCents int64, currencyCode string, createdBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
    if m.CreateReconciliationFunc != nil {
        return m.CreateReconciliationFunc(ctx, organizationID, bankAccountID, accountID, reconciliationDate, statementDate, bankStatementBalanceCents, bookBalanceCents, currencyCode, createdBy)
    }
    return nil, fmt.Errorf("CreateReconciliationFunc not implemented")
}

func (m *MockReconciliationUseCase) AddReconciliationItem(ctx context.Context, reconciliationID uuidv7.UUID, transactionType reconciliation.TransactionType, transactionID *uuidv7.UUID, transactionDate time.Time, description string, amountCents int64, notes string, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
    if m.AddReconciliationItemFunc != nil {
        return m.AddReconciliationItemFunc(ctx, reconciliationID, transactionType, transactionID, transactionDate, description, amountCents, notes, updatedBy)
    }
    return nil, fmt.Errorf("AddReconciliationItemFunc not implemented")
}

func (m *MockReconciliationUseCase) RemoveReconciliationItem(ctx context.Context, reconciliationID, itemID uuidv7.UUID, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
    if m.RemoveReconciliationItemFunc != nil {
        return m.RemoveReconciliationItemFunc(ctx, reconciliationID, itemID, updatedBy)
    }
    return nil, fmt.Errorf("RemoveReconciliationItemFunc not implemented")
}

func (m *MockReconciliationUseCase) MarkItemMatched(ctx context.Context, reconciliationID, itemID uuidv7.UUID, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
    if m.MarkItemMatchedFunc != nil {
        return m.MarkItemMatchedFunc(ctx, reconciliationID, itemID, updatedBy)
    }
    return nil, fmt.Errorf("MarkItemMatchedFunc not implemented")
}

func (m *MockReconciliationUseCase) CompleteReconciliation(ctx context.Context, reconciliationID, reconciledBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
    if m.CompleteReconciliationFunc != nil {
        return m.CompleteReconciliationFunc(ctx, reconciliationID, reconciledBy)
    }
    return nil, fmt.Errorf("CompleteReconciliationFunc not implemented")
}

func (m *MockReconciliationUseCase) ReopenReconciliation(ctx context.Context, reconciliationID, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
    if m.ReopenReconciliationFunc != nil {
        return m.ReopenReconciliationFunc(ctx, reconciliationID, updatedBy)
    }
    return nil, fmt.Errorf("ReopenReconciliationFunc not implemented")
}

func (m *MockReconciliationUseCase) GetReconciliationByID(ctx context.Context, id uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
    if m.GetReconciliationByIDFunc != nil {
        return m.GetReconciliationByIDFunc(ctx, id)
    }
    return nil, fmt.Errorf("GetReconciliationByIDFunc not implemented")
}

func (m *MockReconciliationUseCase) DeleteReconciliation(ctx context.Context, id uuidv7.UUID) error {
    if m.DeleteReconciliationFunc != nil {
        return m.DeleteReconciliationFunc(ctx, id)
    }
    return fmt.Errorf("DeleteReconciliationFunc not implemented")
}

func (m *MockReconciliationUseCase) ListReconciliationsByBankAccount(ctx context.Context, bankAccountID uuidv7.UUID) ([]*reconciliationAggregate.Reconciliation, error) {
    if m.ListReconciliationsByBankAccountFunc != nil {
        return m.ListReconciliationsByBankAccountFunc(ctx, bankAccountID)
    }
    return nil, fmt.Errorf("ListReconciliationsByBankAccountFunc not implemented")
}

func (m *MockReconciliationUseCase) ListReconciliationsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*reconciliationAggregate.Reconciliation, error) {
    if m.ListReconciliationsByOrganizationFunc != nil {
        return m.ListReconciliationsByOrganizationFunc(ctx, organizationID, limit, offset)
    }
    return nil, fmt.Errorf("ListReconciliationsByOrganizationFunc not implemented")
}

func (m *MockReconciliationUseCase) ListReconciliationsByStatus(ctx context.Context, organizationID uuidv7.UUID, status reconciliation.Status) ([]*reconciliationAggregate.Reconciliation, error) {
    if m.ListReconciliationsByStatusFunc != nil {
        return m.ListReconciliationsByStatusFunc(ctx, organizationID, status)
    }
    return nil, fmt.Errorf("ListReconciliationsByStatusFunc not implemented")
}

func setupRouter(mockUC *MockReconciliationUseCase) *gin.Engine {
    router := smoke.SetupRouter()
    handler := reconciliationHTTP.NewReconciliationHandler(mockUC)

    router.Use(func(c *gin.Context) {
        c.Set("organization_id", uuidv7.New().String())
        c.Set("user_id", uuidv7.New().String())
        c.Next()
    })

    handler.RegisterRoutes(router.Group("/api/v1/accounting"))
    return router
}

func TestCreateReconciliation_Success(t *testing.T) {
    reconciliationID := uuidv7.New()
    bankAccountID := uuidv7.New()
    accountID := uuidv7.New()
    reconciliationDate := time.Now()
    statementDate := time.Now()

    mockUC := &MockReconciliationUseCase{
        CreateReconciliationFunc: func(ctx context.Context, organizationID, bankAccID, accID uuidv7.UUID, reconDate, stmtDate time.Time, bankStatementBalanceCents, bookBalanceCents int64, currencyCode string, createdBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
            r, _ := reconciliationAggregate.NewReconciliation(organizationID, bankAccID, accID, reconDate, stmtDate, bankStatementBalanceCents, bookBalanceCents, currencyCode, createdBy)
            r.ID = reconciliationID
            return r, nil
        },
    }

    router := setupRouter(mockUC)

    body := map[string]any{
        "bank_account_id":               bankAccountID.String(),
        "account_id":                    accountID.String(),
        "reconciliation_date":           reconciliationDate.Format(time.RFC3339),
        "statement_date":                statementDate.Format(time.RFC3339),
        "bank_statement_balance_cents":  100000,
        "book_balance_cents":            98000,
        "currency_code":                 "UAH",
    }

    w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/bank-reconciliations", body)
    if w.Code != http.StatusCreated {
        t.Logf("Response body: %s", w.Body.String())
    }
    assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateReconciliation_ValidationError(t *testing.T) {
    mockUC := &MockReconciliationUseCase{
        CreateReconciliationFunc: func(ctx context.Context, organizationID, bankAccID, accID uuidv7.UUID, reconDate, stmtDate time.Time, bankStatementBalanceCents, bookBalanceCents int64, currencyCode string, createdBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
			return nil, reconciliation.ErrDescriptionRequired
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"bank_account_id":               uuidv7.New().String(),
		"account_id":                    uuidv7.New().String(),
		"reconciliation_date":           time.Now().Format(time.RFC3339),
		"statement_date":                time.Now().Format(time.RFC3339),
		"bank_statement_balance_cents":  100000,
		"book_balance_cents":            98000,
		"currency_code":                 "",
	}

    w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/bank-reconciliations", body)
    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetReconciliation_Success(t *testing.T) {
    bankAccountID := uuidv7.New()
    accountID := uuidv7.New()
    reconciliationID := uuidv7.New()
    reconciliationDate := time.Now()
    statementDate := time.Now()

    r, _ := reconciliationAggregate.NewReconciliation(uuidv7.New(), bankAccountID, accountID, reconciliationDate, statementDate, 100000, 98000, "UAH", uuidv7.New())
    r.ID = reconciliationID

    mockUC := &MockReconciliationUseCase{
        GetReconciliationByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
            return r, nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/bank-reconciliations/"+reconciliationID.String(), nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetReconciliation_NotFound(t *testing.T) {
    reconciliationID := uuidv7.New()

    mockUC := &MockReconciliationUseCase{
        GetReconciliationByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
            return nil, reconciliation.ErrReconciliationNotFound
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/bank-reconciliations/"+reconciliationID.String(), nil)
    assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddReconciliationItem_Success(t *testing.T) {
    bankAccountID := uuidv7.New()
    accountID := uuidv7.New()
    reconciliationID := uuidv7.New()
    reconciliationDate := time.Now()
    statementDate := time.Now()

    r, _ := reconciliationAggregate.NewReconciliation(uuidv7.New(), bankAccountID, accountID, reconciliationDate, statementDate, 100000, 98000, "UAH", uuidv7.New())
    r.ID = reconciliationID

    mockUC := &MockReconciliationUseCase{
        AddReconciliationItemFunc: func(ctx context.Context, reconID uuidv7.UUID, transactionType reconciliation.TransactionType, transactionID *uuidv7.UUID, transactionDate time.Time, description string, amountCents int64, notes string, updatedBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
            return r, nil
        },
    }

    router := setupRouter(mockUC)

    body := map[string]any{
        "transaction_type": "bank_transaction",
        "transaction_date": time.Now().Format(time.RFC3339),
        "description":      "Bank deposit",
        "amount_cents":     50000,
        "notes":            "Monthly transfer",
    }

    w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/bank-reconciliations/"+reconciliationID.String()+"/items", body)
    if w.Code != http.StatusOK {
        t.Logf("Response body: %s", w.Body.String())
    }
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestCompleteReconciliation_Success(t *testing.T) {
    bankAccountID := uuidv7.New()
    accountID := uuidv7.New()
    reconciliationID := uuidv7.New()
    reconciliationDate := time.Now()
    statementDate := time.Now()

    r, _ := reconciliationAggregate.NewReconciliation(uuidv7.New(), bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", uuidv7.New())
    r.ID = reconciliationID

    mockUC := &MockReconciliationUseCase{
        CompleteReconciliationFunc: func(ctx context.Context, reconID, reconciledBy uuidv7.UUID) (*reconciliationAggregate.Reconciliation, error) {
            return r, nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/bank-reconciliations/"+reconciliationID.String()+"/complete", nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestListReconciliations_Success(t *testing.T) {
    bankAccountID := uuidv7.New()
    accountID := uuidv7.New()
    reconciliationDate := time.Now()
    statementDate := time.Now()

    reconciliations := []*reconciliationAggregate.Reconciliation{
        func() *reconciliationAggregate.Reconciliation {
            r, _ := reconciliationAggregate.NewReconciliation(uuidv7.New(), bankAccountID, accountID, reconciliationDate, statementDate, 100000, 98000, "UAH", uuidv7.New())
            return r
        }(),
        func() *reconciliationAggregate.Reconciliation {
            r, _ := reconciliationAggregate.NewReconciliation(uuidv7.New(), bankAccountID, accountID, reconciliationDate, statementDate, 98000, 98000, "UAH", uuidv7.New())
            return r
        }(),
    }

    mockUC := &MockReconciliationUseCase{
        ListReconciliationsByOrganizationFunc: func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*reconciliationAggregate.Reconciliation, error) {
            return reconciliations, nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/bank-reconciliations", nil)
    assert.Equal(t, http.StatusOK, w.Code)
}