package aggregate

import (
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReconciliation(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	t.Run("valid reconciliation with currency", func(t *testing.T) {
		recon, err := NewReconciliation(
			orgID,
			bankAccountID,
			accountID,
			reconciliationDate,
			statementDate,
			100000,
			95000,
			"UAH",
			userID,
		)

		require.NoError(t, err)
		require.NotNil(t, recon)
		assert.Equal(t, orgID, recon.OrganizationID)
		assert.Equal(t, bankAccountID, recon.BankAccountID)
		assert.Equal(t, accountID, recon.AccountID)
		assert.Equal(t, int64(100000), recon.BankStatementBalanceCents)
		assert.Equal(t, int64(95000), recon.BookBalanceCents)
		assert.Equal(t, "UAH", recon.CurrencyCode)
		assert.Equal(t, reconciliation.StatusInProgress, recon.Status)
		assert.Empty(t, recon.Items)
	})

	t.Run("default currency", func(t *testing.T) {
		recon, err := NewReconciliation(
			orgID,
			bankAccountID,
			accountID,
			reconciliationDate,
			statementDate,
			100000,
			95000,
			"",
			userID,
		)

		require.NoError(t, err)
		assert.Equal(t, "UAH", recon.CurrencyCode)
	})
}

func TestReconciliation_AddItem(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()
	transactionID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
	require.NoError(t, err)

	t.Run("add item with transaction ID", func(t *testing.T) {
		err := recon.AddItem(
			reconciliation.TransactionTypeBankTransaction,
			&transactionID,
			transactionDate,
			"Deposit from customer",
			50000,
			"Test note",
		)

		require.NoError(t, err)
		assert.Len(t, recon.Items, 1)
		assert.Equal(t, reconciliation.TransactionTypeBankTransaction, recon.Items[0].TransactionType)
		assert.Equal(t, &transactionID, recon.Items[0].TransactionID)
		assert.Equal(t, "Deposit from customer", recon.Items[0].Description)
		assert.Equal(t, int64(50000), recon.Items[0].AmountCents)
		assert.False(t, recon.Items[0].IsMatched)
		assert.Nil(t, recon.Items[0].MatchedAt)
	})

	t.Run("add item without transaction ID", func(t *testing.T) {
		err := recon.AddItem(
			reconciliation.TransactionTypeOutstanding,
			nil,
			transactionDate,
			"Check payment",
			30000,
			"",
		)

		require.NoError(t, err)
		assert.Len(t, recon.Items, 2)
		assert.Nil(t, recon.Items[1].TransactionID)
	})

	t.Run("empty description", func(t *testing.T) {
		err := recon.AddItem(
			reconciliation.TransactionTypeBankTransaction,
			nil,
			transactionDate,
			"",
			10000,
			"",
		)

		assert.ErrorIs(t, err, reconciliation.ErrDescriptionRequired)
	})

	t.Run("cannot add to completed reconciliation", func(t *testing.T) {
		completedRecon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = completedRecon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		err = completedRecon.MarkItemMatched(completedRecon.Items[0].ID)
		require.NoError(t, err)
		err = completedRecon.Complete(userID)
		require.NoError(t, err)

		err = completedRecon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "New item", 5000, "")
		assert.ErrorIs(t, err, reconciliation.ErrCannotModifyCompleted)
	})
}

func TestReconciliation_RemoveItem(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
	require.NoError(t, err)
	err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item 1", 10000, "")
	require.NoError(t, err)
	err = recon.AddItem(reconciliation.TransactionTypeOutstanding, nil, transactionDate, "Item 2", 20000, "")
	require.NoError(t, err)
	itemID := recon.Items[0].ID

	t.Run("remove item", func(t *testing.T) {
		err := recon.RemoveItem(itemID)
		require.NoError(t, err)
		assert.Len(t, recon.Items, 1)
	})

	t.Run("remove non-existent item", func(t *testing.T) {
		err := recon.RemoveItem(uuidv7.New())
		assert.ErrorIs(t, err, reconciliation.ErrItemNotFound)
	})

	t.Run("cannot remove from completed reconciliation", func(t *testing.T) {
		completedRecon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = completedRecon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		removeItemID := completedRecon.Items[0].ID
		err = completedRecon.MarkItemMatched(removeItemID)
		require.NoError(t, err)
		err = completedRecon.Complete(userID)
		require.NoError(t, err)

		err = completedRecon.RemoveItem(removeItemID)
		assert.ErrorIs(t, err, reconciliation.ErrCannotModifyCompleted)
	})
}

func TestReconciliation_MarkItemMatched(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
	require.NoError(t, err)
	err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
	require.NoError(t, err)
	itemID := recon.Items[0].ID

	t.Run("mark item matched", func(t *testing.T) {
		err := recon.MarkItemMatched(itemID)
		require.NoError(t, err)
		assert.True(t, recon.Items[0].IsMatched)
		assert.NotNil(t, recon.Items[0].MatchedAt)
	})

	t.Run("mark non-existent item", func(t *testing.T) {
		err := recon.MarkItemMatched(uuidv7.New())
		assert.ErrorIs(t, err, reconciliation.ErrItemNotFound)
	})

	t.Run("cannot mark in completed reconciliation", func(t *testing.T) {
		completedRecon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = completedRecon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		matchItemID := completedRecon.Items[0].ID
		err = completedRecon.MarkItemMatched(matchItemID)
		require.NoError(t, err)
		err = completedRecon.Complete(userID)
		require.NoError(t, err)

		err = completedRecon.MarkItemMatched(matchItemID)
		assert.ErrorIs(t, err, reconciliation.ErrCannotModifyCompleted)
	})
}

func TestReconciliation_Complete(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("complete with all matched items", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item 1", 10000, "")
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeOutstanding, nil, transactionDate, "Item 2", 20000, "")
		require.NoError(t, err)
		err = recon.MarkItemMatched(recon.Items[0].ID)
		require.NoError(t, err)
		err = recon.MarkItemMatched(recon.Items[1].ID)
		require.NoError(t, err)

		err = recon.Complete(userID)
		require.NoError(t, err)
		assert.Equal(t, reconciliation.StatusCompleted, recon.Status)
		assert.NotNil(t, recon.ReconciledBy)
		assert.Equal(t, userID, *recon.ReconciledBy)
		assert.NotNil(t, recon.ReconciledAt)
	})

	t.Run("cannot complete with unmatched items", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item 1", 10000, "")
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeOutstanding, nil, transactionDate, "Item 2", 20000, "")
		require.NoError(t, err)
		err = recon.MarkItemMatched(recon.Items[0].ID)
		require.NoError(t, err)

		err = recon.Complete(userID)
		assert.ErrorIs(t, err, reconciliation.ErrHasUnmatchedItems)
	})

	t.Run("cannot complete already completed", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		err = recon.MarkItemMatched(recon.Items[0].ID)
		require.NoError(t, err)
		err = recon.Complete(userID)
		require.NoError(t, err)

		err = recon.Complete(userID)
		assert.ErrorIs(t, err, reconciliation.ErrAlreadyCompleted)
	})
}

func TestReconciliation_Approve(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("approve completed reconciliation", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		err = recon.MarkItemMatched(recon.Items[0].ID)
		require.NoError(t, err)
		err = recon.Complete(userID)
		require.NoError(t, err)

		err = recon.Approve(userID)
		require.NoError(t, err)
		assert.Equal(t, reconciliation.StatusApproved, recon.Status)
		assert.NotNil(t, recon.ApprovedBy)
		assert.Equal(t, userID, *recon.ApprovedBy)
		assert.NotNil(t, recon.ApprovedAt)
	})

	t.Run("cannot approve in progress reconciliation", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
		require.NoError(t, err)

		err = recon.Approve(userID)
		assert.ErrorIs(t, err, reconciliation.ErrNotCompleted)
	})
}

func TestReconciliation_Reopen(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("reopen completed reconciliation", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		err = recon.MarkItemMatched(recon.Items[0].ID)
		require.NoError(t, err)
		err = recon.Complete(userID)
		require.NoError(t, err)

		err = recon.Reopen()
		require.NoError(t, err)
		assert.Equal(t, reconciliation.StatusInProgress, recon.Status)
		assert.Nil(t, recon.ReconciledBy)
		assert.Nil(t, recon.ReconciledAt)
	})

	t.Run("cannot reopen approved reconciliation", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = recon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		err = recon.MarkItemMatched(recon.Items[0].ID)
		require.NoError(t, err)
		err = recon.Complete(userID)
		require.NoError(t, err)
		err = recon.Approve(userID)
		require.NoError(t, err)

		err = recon.Reopen()
		assert.ErrorIs(t, err, reconciliation.ErrCannotReopenApproved)
	})

	t.Run("cannot reopen in progress reconciliation", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
		require.NoError(t, err)

		err = recon.Reopen()
		assert.ErrorIs(t, err, reconciliation.ErrAlreadyInProgress)
	})
}

func TestReconciliation_UpdateBalances(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	transactionDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
	require.NoError(t, err)

	t.Run("update balances", func(t *testing.T) {
		err := recon.UpdateBalances(110000, 105000, userID)
		require.NoError(t, err)
		assert.Equal(t, int64(110000), recon.BankStatementBalanceCents)
		assert.Equal(t, int64(105000), recon.BookBalanceCents)
	})

	t.Run("cannot update completed reconciliation", func(t *testing.T) {
		completedRecon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)
		err = completedRecon.AddItem(reconciliation.TransactionTypeBankTransaction, nil, transactionDate, "Item", 10000, "")
		require.NoError(t, err)
		err = completedRecon.MarkItemMatched(completedRecon.Items[0].ID)
		require.NoError(t, err)
		err = completedRecon.Complete(userID)
		require.NoError(t, err)

		err = completedRecon.UpdateBalances(110000, 105000, userID)
		assert.ErrorIs(t, err, reconciliation.ErrCannotModifyCompleted)
	})
}

func TestReconciliation_CalculateDifference(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	t.Run("positive difference", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
		require.NoError(t, err)

		diff := recon.CalculateDifference()
		assert.Equal(t, int64(5000), diff)
	})

	t.Run("negative difference", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 95000, 100000, "UAH", userID)
		require.NoError(t, err)

		diff := recon.CalculateDifference()
		assert.Equal(t, int64(-5000), diff)
	})

	t.Run("zero difference", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)

		diff := recon.CalculateDifference()
		assert.Equal(t, int64(0), diff)
	})
}

func TestReconciliation_IsReconciled(t *testing.T) {
	orgID := uuidv7.New()
	bankAccountID := uuidv7.New()
	accountID := uuidv7.New()
	userID := uuidv7.New()

	reconciliationDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statementDate := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	t.Run("balanced", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 100000, "UAH", userID)
		require.NoError(t, err)

		assert.True(t, recon.IsReconciled())
	})

	t.Run("unbalanced", func(t *testing.T) {
		recon, err := NewReconciliation(orgID, bankAccountID, accountID, reconciliationDate, statementDate, 100000, 95000, "UAH", userID)
		require.NoError(t, err)

		assert.False(t, recon.IsReconciled())
	})
}
