package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockReconciliationRepository struct {
	CreateFunc             func(ctx context.Context, rec *aggregate.Reconciliation) error
	GetByIDFunc            func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error)
	UpdateFunc             func(ctx context.Context, rec *aggregate.Reconciliation) error
	DeleteFunc             func(ctx context.Context, id uuidv7.UUID) error
	ListByOrganizationFunc func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.Reconciliation, error)
	ListByBankAccountFunc  func(ctx context.Context, bankAccountID uuidv7.UUID) ([]*aggregate.Reconciliation, error)
	ListByAccountFunc      func(ctx context.Context, accountID uuidv7.UUID) ([]*aggregate.Reconciliation, error)
	ListByStatusFunc       func(ctx context.Context, orgID uuidv7.UUID, status reconciliation.Status) ([]*aggregate.Reconciliation, error)
	ListByDateRangeFunc    func(ctx context.Context, orgID uuidv7.UUID, startDate, endDate string) ([]*aggregate.Reconciliation, error)
}

func (m *MockReconciliationRepository) Create(ctx context.Context, rec *aggregate.Reconciliation) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, rec)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockReconciliationRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockReconciliationRepository) Update(ctx context.Context, rec *aggregate.Reconciliation) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, rec)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockReconciliationRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockReconciliationRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.Reconciliation, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, orgID, limit, offset)
	}
	return nil, errors.New("ListByOrganizationFunc not implemented")
}

func (m *MockReconciliationRepository) ListByBankAccount(ctx context.Context, bankAccountID uuidv7.UUID) ([]*aggregate.Reconciliation, error) {
	if m.ListByBankAccountFunc != nil {
		return m.ListByBankAccountFunc(ctx, bankAccountID)
	}
	return nil, errors.New("ListByBankAccountFunc not implemented")
}

func (m *MockReconciliationRepository) ListByAccount(ctx context.Context, accountID uuidv7.UUID) ([]*aggregate.Reconciliation, error) {
	if m.ListByAccountFunc != nil {
		return m.ListByAccountFunc(ctx, accountID)
	}
	return nil, errors.New("ListByAccountFunc not implemented")
}

func (m *MockReconciliationRepository) ListByStatus(ctx context.Context, orgID uuidv7.UUID, status reconciliation.Status) ([]*aggregate.Reconciliation, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, orgID, status)
	}
	return nil, errors.New("ListByStatusFunc not implemented")
}

func (m *MockReconciliationRepository) ListByDateRange(ctx context.Context, orgID uuidv7.UUID, startDate, endDate string) ([]*aggregate.Reconciliation, error) {
	if m.ListByDateRangeFunc != nil {
		return m.ListByDateRangeFunc(ctx, orgID, startDate, endDate)
	}
	return nil, errors.New("ListByDateRangeFunc not implemented")
}

// ============================================================================
// Tests
// ============================================================================

func TestCreateReconciliation_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	reconciliationDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	mockRepo := &MockReconciliationRepository{
		CreateFunc: func(ctx context.Context, rec *aggregate.Reconciliation) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	rec, err := uc.CreateReconciliation(context.Background(), orgID, bankAccountID, accountID, reconciliationDate, statementDate, 500000, 450000, "UAH", userID)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, rec.ID)
	assert.Equal(t, accountID, rec.AccountID)
	assert.Equal(t, statementDate, rec.StatementDate)
	assert.Equal(t, int64(500000), rec.BankStatementBalanceCents)
	assert.Equal(t, int64(450000), rec.BookBalanceCents)
	assert.Equal(t, reconciliation.StatusInProgress, rec.Status)
}

func TestAddMatchRecord_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", userID)
	rec.ID = uuidv7.New()

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Reconciliation) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	transactionID := uuidv7.New()
	transactionDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	updated, err := uc.AddReconciliationItem(context.Background(), rec.ID, reconciliation.TransactionTypeBankTransaction, &transactionID, transactionDate, "Payment match", 25000, "", userID)
	require.NoError(t, err)
	assert.Len(t, updated.Items, 1)
	assert.Equal(t, &transactionID, updated.Items[0].TransactionID)
	assert.Equal(t, int64(25000), updated.Items[0].AmountCents)
}

func TestAddMatchRecord_CannotModifyCompleted(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 500000, "UAH", userID)
	rec.ID = uuidv7.New()
	_ = rec.Complete(userID)

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	transactionID := uuidv7.New()
	transactionDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	_, err := uc.AddReconciliationItem(context.Background(), rec.ID, reconciliation.TransactionTypeBankTransaction, &transactionID, transactionDate, "Cannot add", 25000, "", userID)
	assert.Error(t, err)
	assert.Equal(t, reconciliation.ErrCannotModifyCompleted, err)
}

func TestRemoveMatchRecord_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", userID)
	rec.ID = uuidv7.New()
	tx1ID := uuidv7.New()
	tx2ID := uuidv7.New()
	_ = rec.AddItem(reconciliation.TransactionTypeBankTransaction, &tx1ID, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "Match 1", 25000, "")
	_ = rec.AddItem(reconciliation.TransactionTypeBankTransaction, &tx2ID, time.Date(2025, 1, 16, 0, 0, 0, 0, time.UTC), "Match 2", 30000, "")
	recordToRemove := rec.Items[0].ID

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Reconciliation) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	updated, err := uc.RemoveReconciliationItem(context.Background(), rec.ID, recordToRemove, userID)
	require.NoError(t, err)
	assert.Len(t, updated.Items, 1)
}

func TestCompleteReconciliation_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 500000, "UAH", userID)
	rec.ID = uuidv7.New()

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Reconciliation) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	completed, err := uc.CompleteReconciliation(context.Background(), rec.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, reconciliation.StatusCompleted, completed.Status)
	assert.NotNil(t, completed.ReconciledAt)
	assert.Equal(t, userID, *completed.ReconciledBy)
}

func TestCompleteReconciliation_Unbalanced(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", userID)
	rec.ID = uuidv7.New()
	// Add unmatched item to trigger ErrHasUnmatchedItems
	txID := uuidv7.New()
	_ = rec.AddItem(reconciliation.TransactionTypeBankTransaction, &txID, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "Unmatched", 10000, "")

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Reconciliation) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	_, err := uc.CompleteReconciliation(context.Background(), rec.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, reconciliation.ErrHasUnmatchedItems, err)
}

func TestReopenReconciliation_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 500000, "UAH", userID)
	rec.ID = uuidv7.New()
	_ = rec.Complete(userID)

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Reconciliation) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	reopened, err := uc.ReopenReconciliation(context.Background(), rec.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, reconciliation.StatusInProgress, reopened.Status)
}

func TestReopenReconciliation_NotCompleted(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", userID)
	rec.ID = uuidv7.New()

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	_, err := uc.ReopenReconciliation(context.Background(), rec.ID, userID)
	assert.Error(t, err)
	assert.Equal(t, reconciliation.ErrAlreadyInProgress, err)
}

func TestGetReconciliationByID_Success(t *testing.T) {
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	rec, _ := aggregate.NewReconciliation(uuidv7.New(), uuidv7.New(), uuidv7.New(), time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", uuidv7.New())
	rec.ID = uuidv7.New()

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	retrieved, err := uc.GetReconciliationByID(context.Background(), rec.ID)
	require.NoError(t, err)
	assert.Equal(t, rec.ID, retrieved.ID)
}

func TestGetReconciliationByID_NotFound(t *testing.T) {
	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			return nil, reconciliation.ErrReconciliationNotFound
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	_, err := uc.GetReconciliationByID(context.Background(), uuidv7.New())
	assert.Error(t, err)
	assert.Equal(t, reconciliation.ErrReconciliationNotFound, err)
}

func TestDeleteReconciliation_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", userID)
	rec.ID = uuidv7.New()

	mockRepo := &MockReconciliationRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
			if id == rec.ID {
				return rec, nil
			}
			return nil, reconciliation.ErrReconciliationNotFound
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	err := uc.DeleteReconciliation(context.Background(), rec.ID)
	require.NoError(t, err)
}

func TestListReconciliationsByOrganization_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec1, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", userID)
	rec2, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 600000, 550000, "UAH", userID)

	mockRepo := &MockReconciliationRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.Reconciliation, error) {
			return []*aggregate.Reconciliation{rec1, rec2}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	reconciliations, err := uc.ListReconciliationsByOrganization(context.Background(), orgID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, reconciliations, 2)
}

func TestListReconciliationsByAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	rec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 450000, "UAH", userID)

	mockRepo := &MockReconciliationRepository{
		ListByBankAccountFunc: func(ctx context.Context, accID uuidv7.UUID) ([]*aggregate.Reconciliation, error) {
			if accID == accountID {
				return []*aggregate.Reconciliation{rec}, nil
			}
			return []*aggregate.Reconciliation{}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	reconciliations, err := uc.ListReconciliationsByBankAccount(context.Background(), accountID)
	require.NoError(t, err)
	assert.Len(t, reconciliations, 1)
	assert.Equal(t, accountID, reconciliations[0].AccountID)
}

func TestListReconciliationsByStatus_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()
	statementDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	completedRec, _ := aggregate.NewReconciliation(orgID, uuidv7.New(), accountID, time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC), statementDate, 500000, 500000, "UAH", userID)
	_ = completedRec.Complete(userID)

	mockRepo := &MockReconciliationRepository{
		ListByStatusFunc: func(ctx context.Context, orgID uuidv7.UUID, status reconciliation.Status) ([]*aggregate.Reconciliation, error) {
			if status == reconciliation.StatusCompleted {
				return []*aggregate.Reconciliation{completedRec}, nil
			}
			return []*aggregate.Reconciliation{}, nil
		},
	}

	var auditLogger *audit.AuditLogger = nil
	uc := NewReconciliationUseCase(mockRepo, auditLogger)

	completedRecs, err := uc.ListReconciliationsByStatus(context.Background(), orgID, reconciliation.StatusCompleted)
	require.NoError(t, err)
	assert.Len(t, completedRecs, 1)
	assert.Equal(t, reconciliation.StatusCompleted, completedRecs[0].Status)
}
