package reconciliation_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestReconciliationRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		// Create GL account for reconciliation
		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(
			orgID,
			bankAccountID,
			accountID,
			reconciliationDate,
			statementDate,
			50000000, // 500,000.00 UAH
			49500000, // 495,000.00 UAH
			"UAH",
			userID,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.Equal(t, recon.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, bankAccountID, found.BankAccountID)
		assert.Equal(t, accountID, found.AccountID)
		assert.Equal(t, int64(50000000), found.BankStatementBalanceCents)
		assert.Equal(t, int64(49500000), found.BookBalanceCents)
		assert.Equal(t, reconciliation.StatusInProgress, found.Status)
	})
}

func TestReconciliationRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.ErrorIs(t, err, reconciliation.ErrReconciliationNotFound)
	})
}

func TestReconciliationRepository_Update_CompleteReconciliation(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		// Add and match an item
		transactionDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Payment received", 500000, "Test payment")
		require.NoError(t, err)

		err = recon.MarkItemMatched(recon.Items[0].ID)
		require.NoError(t, err)

		// Complete reconciliation
		err = recon.Complete(userID)
		require.NoError(t, err)

		err = repo.Update(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.Equal(t, reconciliation.StatusCompleted, found.Status)
		assert.NotNil(t, found.ReconciledBy)
		assert.NotNil(t, found.ReconciledAt)
	})
}

func TestReconciliationRepository_Update_ApproveReconciliation(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		approverID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		// Complete and then approve
		require.NoError(t, recon.Complete(userID))
		require.NoError(t, repo.Update(ctx, recon))

		err = recon.Approve(approverID)
		require.NoError(t, err)

		err = repo.Update(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.Equal(t, reconciliation.StatusApproved, found.Status)
		assert.NotNil(t, found.ApprovedBy)
		assert.NotNil(t, found.ApprovedAt)
	})
}

func TestReconciliationRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		err = repo.Delete(ctx, recon.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, recon.GetID())
		assert.ErrorIs(t, err, reconciliation.ErrReconciliationNotFound)
	})
}

func TestReconciliationRepository_ListByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		// Create multiple reconciliations
		for i := 1; i <= 3; i++ {
			bankAccountID := uuidv7.New()
			recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
			require.NoError(t, err)
			require.NoError(t, repo.Create(ctx, recon))
		}

		found, err := repo.ListByOrganization(ctx, orgID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, found, 3)
	})
}

func TestReconciliationRepository_ListByBankAccount(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		// Create reconciliations for different bank accounts
		bankAccount1 := uuidv7.New()
		recon1, err := aggregate.NewReconciliation(orgID, bankAccount1, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon1))

		bankAccount2 := uuidv7.New()
		recon2, err := aggregate.NewReconciliation(orgID, bankAccount2, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon2))

		bankAccount3 := uuidv7.New()
		recon3, err := aggregate.NewReconciliation(orgID, bankAccount3, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon3))

		found, err := repo.ListByBankAccount(ctx, bankAccount1)
		require.NoError(t, err)
		assert.Len(t, found, 1)
	})
}

func TestReconciliationRepository_ListByStatus(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		// In progress
		bankAccount1 := uuidv7.New()
		recon1, err := aggregate.NewReconciliation(orgID, bankAccount1, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon1))

		// Completed
		bankAccount2 := uuidv7.New()
		recon2, err := aggregate.NewReconciliation(orgID, bankAccount2, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon2))
		require.NoError(t, recon2.Complete(userID))
		require.NoError(t, repo.Update(ctx, recon2))

		found, err := repo.ListByStatus(ctx, orgID, reconciliation.StatusInProgress)
		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, reconciliation.StatusInProgress, found[0].Status)
	})
}

func TestReconciliation_AddItem(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		// Add item
		transactionDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Bank fee", -25000, "Monthly fee")
		require.NoError(t, err)

		err = repo.Update(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.Len(t, found.Items, 1)
		assert.Equal(t, "Bank fee", found.Items[0].Description)
		assert.Equal(t, int64(-25000), found.Items[0].AmountCents)
		assert.False(t, found.Items[0].IsMatched)
	})
}

func TestReconciliation_MarkItemMatched(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		// Add and match item
		transactionDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		require.NoError(t, recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Payment", 500000, ""))
		itemID := recon.Items[0].ID

		err = recon.MarkItemMatched(itemID)
		require.NoError(t, err)

		err = repo.Update(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.True(t, found.Items[0].IsMatched)
		assert.NotNil(t, found.Items[0].MatchedAt)
	})
}

func TestReconciliation_RemoveItem(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		// Add two items
		transactionDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		require.NoError(t, recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item 1", 100000, ""))
		require.NoError(t, recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item 2", 200000, ""))
		itemIDToRemove := recon.Items[0].ID

		// Remove first item
		err = recon.RemoveItem(itemIDToRemove)
		require.NoError(t, err)

		err = repo.Update(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.Len(t, found.Items, 1)
		assert.Equal(t, "Item 2", found.Items[0].Description)
	})
}

func TestReconciliation_BusinessRules_CannotCompleteWithUnmatchedItems(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)

		// Add unmatched item
		transactionDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		require.NoError(t, recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Unmatched", 100000, ""))

		// Try to complete
		err = recon.Complete(userID)
		assert.ErrorIs(t, err, reconciliation.ErrHasUnmatchedItems)
	})
}

func TestReconciliation_BusinessRules_CannotModifyCompleted(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)

		// Complete without items
		require.NoError(t, recon.Complete(userID))

		// Try to add item after completion
		transactionDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "New item", 100000, "")
		assert.ErrorIs(t, err, reconciliation.ErrCannotModifyCompleted)
	})
}

func TestReconciliation_BusinessRules_ReopenCompleted(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		// Complete
		require.NoError(t, recon.Complete(userID))
		require.NoError(t, repo.Update(ctx, recon))

		// Reopen
		err = recon.Reopen()
		require.NoError(t, err)

		err = repo.Update(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.Equal(t, reconciliation.StatusInProgress, found.Status)
		assert.Nil(t, found.ReconciledBy)
		assert.Nil(t, found.ReconciledAt)
	})
}

func TestReconciliation_BusinessRules_CannotReopenApproved(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		approverID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)

		// Complete and approve
		require.NoError(t, recon.Complete(userID))
		require.NoError(t, recon.Approve(approverID))

		// Try to reopen
		err = recon.Reopen()
		assert.ErrorIs(t, err, reconciliation.ErrCannotReopenApproved)
	})
}

func TestReconciliation_CalculateDifference(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)

		difference := recon.CalculateDifference()
		assert.Equal(t, int64(500000), difference)

		isReconciled := recon.IsReconciled()
		assert.False(t, isReconciled)
	})
}

func TestReconciliation_UpdateBalances(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewReconciliationRepository(db.DB)

		orgID := uuidv7.New()
		bankAccountID := uuidv7.New()
		userID := uuidv7.New()
		reconciliationDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		statementDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID)
		require.NoError(t, err)

		recon, err := aggregate.NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 50000000, 49500000, "UAH", userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, recon))

		// Update balances
		err = recon.UpdateBalances(10500000, 10500000, userID)
		require.NoError(t, err)

		err = repo.Update(ctx, recon)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, recon.GetID())
		require.NoError(t, err)
		assert.Equal(t, int64(10500000), found.BankStatementBalanceCents)
		assert.Equal(t, int64(10500000), found.BookBalanceCents)
		assert.True(t, found.IsReconciled())
	})
}
