package aggregate

import (
	"testing"

	"github.com/basilex/promenade/internal/contexts/accounting/budget"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBudget(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tests := []struct {
		name        string
		budgetName  string
		fiscalYear  int
		expectError error
	}{
		{
			name:        "valid budget",
			budgetName:  "Operational Budget 2024",
			fiscalYear:  2024,
			expectError: nil,
		},
		{
			name:        "valid budget 2025",
			budgetName:  "Capital Budget 2025",
			fiscalYear:  2025,
			expectError: nil,
		},
		{
			name:        "empty name",
			budgetName:  "",
			fiscalYear:  2024,
			expectError: budget.ErrBudgetNameEmpty,
		},
		{
			name:        "invalid fiscal year too low",
			budgetName:  "Test Budget",
			fiscalYear:  1999,
			expectError: budget.ErrInvalidFiscalYear,
		},
		{
			name:        "invalid fiscal year too high",
			budgetName:  "Test Budget",
			fiscalYear:  2101,
			expectError: budget.ErrInvalidFiscalYear,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBudget(orgID, tt.budgetName, tt.fiscalYear, userID)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, b)
			} else {
				require.NoError(t, err)
				require.NotNil(t, b)
				assert.Equal(t, orgID, b.OrganizationID)
				assert.Equal(t, tt.budgetName, b.Name)
				assert.Equal(t, tt.fiscalYear, b.FiscalYear)
				assert.Equal(t, BudgetStatusDraft, b.Status)
				assert.Empty(t, b.Lines)
				assert.Equal(t, int64(0), b.TotalBudget)
				assert.Equal(t, int64(0), b.TotalActual)
				assert.Nil(t, b.ApprovedBy)
				assert.Nil(t, b.ApprovedAt)
			}
		})
	}
}

func TestBudget_AddLine(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID1 := uuidv7.New()
	accountID2 := uuidv7.New()

	b, err := NewBudget(orgID, "Test Budget", 2024, userID)
	require.NoError(t, err)

	t.Run("add valid line", func(t *testing.T) {
		err := b.AddLine(accountID1, 100000, "Marketing expenses")
		require.NoError(t, err)
		assert.Len(t, b.Lines, 1)
		assert.Equal(t, accountID1, b.Lines[0].AccountID)
		assert.Equal(t, int64(100000), b.Lines[0].BudgetAmount)
		assert.Equal(t, "Marketing expenses", b.Lines[0].Description)
		assert.Equal(t, int64(100000), b.TotalBudget)
	})

	t.Run("add second line", func(t *testing.T) {
		err := b.AddLine(accountID2, 50000, "R&D expenses")
		require.NoError(t, err)
		assert.Len(t, b.Lines, 2)
		assert.Equal(t, int64(150000), b.TotalBudget)
	})

	t.Run("add line with nil account", func(t *testing.T) {
		err := b.AddLine(uuidv7.Nil, 10000, "Invalid")
		assert.ErrorIs(t, err, budget.ErrAccountRequired)
	})

	t.Run("add line with negative amount", func(t *testing.T) {
		err := b.AddLine(uuidv7.New(), -1000, "Negative")
		assert.ErrorIs(t, err, budget.ErrInvalidBudgetAmount)
	})

	t.Run("add duplicate account", func(t *testing.T) {
		err := b.AddLine(accountID1, 20000, "Duplicate")
		assert.ErrorIs(t, err, budget.ErrAccountAlreadyExists)
	})

	t.Run("cannot add line to approved budget", func(t *testing.T) {
		approvedBudget, err := NewBudget(orgID, "Approved Budget", 2024, userID)
		require.NoError(t, err)
		err = approvedBudget.AddLine(accountID1, 10000, "Line")
		require.NoError(t, err)
		err = approvedBudget.Approve(userID)
		require.NoError(t, err)

		err = approvedBudget.AddLine(accountID2, 20000, "New line")
		assert.ErrorIs(t, err, budget.ErrCannotModifyApprovedBudget)
	})
}

func TestBudget_UpdateLine(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	b, err := NewBudget(orgID, "Test Budget", 2024, userID)
	require.NoError(t, err)
	err = b.AddLine(accountID, 100000, "Test line")
	require.NoError(t, err)
	lineID := b.Lines[0].ID

	t.Run("update line amount", func(t *testing.T) {
		err := b.UpdateLine(lineID, 150000)
		require.NoError(t, err)
		assert.Equal(t, int64(150000), b.Lines[0].BudgetAmount)
		assert.Equal(t, int64(150000), b.TotalBudget)
	})

	t.Run("update with negative amount", func(t *testing.T) {
		err := b.UpdateLine(lineID, -1000)
		assert.ErrorIs(t, err, budget.ErrInvalidBudgetAmount)
	})

	t.Run("update non-existent line", func(t *testing.T) {
		err := b.UpdateLine(uuidv7.New(), 20000)
		assert.ErrorIs(t, err, budget.ErrBudgetLineNotFound)
	})

	t.Run("cannot update active budget", func(t *testing.T) {
		activeBudget, err := NewBudget(orgID, "Active Budget", 2024, userID)
		require.NoError(t, err)
		err = activeBudget.AddLine(accountID, 10000, "Line")
		require.NoError(t, err)
		activeLine := activeBudget.Lines[0].ID
		err = activeBudget.Approve(userID)
		require.NoError(t, err)
		err = activeBudget.Activate()
		require.NoError(t, err)

		err = activeBudget.UpdateLine(activeLine, 20000)
		assert.ErrorIs(t, err, budget.ErrCannotModifyActiveBudget)
	})
}

func TestBudget_RemoveLine(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID1 := uuidv7.New()
	accountID2 := uuidv7.New()

	b, err := NewBudget(orgID, "Test Budget", 2024, userID)
	require.NoError(t, err)
	err = b.AddLine(accountID1, 100000, "Line 1")
	require.NoError(t, err)
	err = b.AddLine(accountID2, 50000, "Line 2")
	require.NoError(t, err)
	lineID := b.Lines[0].ID

	t.Run("remove line", func(t *testing.T) {
		err := b.RemoveLine(lineID)
		require.NoError(t, err)
		assert.Len(t, b.Lines, 1)
		assert.Equal(t, int64(50000), b.TotalBudget)
	})

	t.Run("remove non-existent line", func(t *testing.T) {
		err := b.RemoveLine(uuidv7.New())
		assert.ErrorIs(t, err, budget.ErrBudgetLineNotFound)
	})

	t.Run("cannot remove from approved budget", func(t *testing.T) {
		approvedBudget, err := NewBudget(orgID, "Approved", 2024, userID)
		require.NoError(t, err)
		err = approvedBudget.AddLine(accountID1, 10000, "Line")
		require.NoError(t, err)
		removeLineID := approvedBudget.Lines[0].ID
		err = approvedBudget.Approve(userID)
		require.NoError(t, err)

		err = approvedBudget.RemoveLine(removeLineID)
		assert.ErrorIs(t, err, budget.ErrCannotModifyApprovedBudget)
	})
}

func TestBudget_UpdateActuals(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	b, err := NewBudget(orgID, "Test Budget", 2024, userID)
	require.NoError(t, err)
	err = b.AddLine(accountID, 100000, "Test line")
	require.NoError(t, err)
	lineID := b.Lines[0].ID
	err = b.Approve(userID)
	require.NoError(t, err)
	err = b.Activate()
	require.NoError(t, err)

	t.Run("update actuals", func(t *testing.T) {
		err := b.UpdateActuals(lineID, 75000)
		require.NoError(t, err)
		assert.Equal(t, int64(75000), b.Lines[0].ActualAmount)
		assert.Equal(t, int64(-25000), b.Lines[0].VarianceAmount)
		assert.Equal(t, int64(75000), b.TotalActual)
	})

	t.Run("update actuals over budget", func(t *testing.T) {
		err := b.UpdateActuals(lineID, 120000)
		require.NoError(t, err)
		assert.Equal(t, int64(120000), b.Lines[0].ActualAmount)
		assert.Equal(t, int64(20000), b.Lines[0].VarianceAmount)
	})

	t.Run("update non-existent line", func(t *testing.T) {
		err := b.UpdateActuals(uuidv7.New(), 50000)
		assert.ErrorIs(t, err, budget.ErrBudgetLineNotFound)
	})

	t.Run("cannot update actuals on non-active budget", func(t *testing.T) {
		draftBudget, err := NewBudget(orgID, "Draft", 2024, userID)
		require.NoError(t, err)
		err = draftBudget.AddLine(accountID, 10000, "Line")
		require.NoError(t, err)
		draftLineID := draftBudget.Lines[0].ID

		err = draftBudget.UpdateActuals(draftLineID, 5000)
		assert.ErrorIs(t, err, budget.ErrBudgetNotActive)
	})
}

func TestBudget_Approve(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	t.Run("approve budget with lines", func(t *testing.T) {
		b, err := NewBudget(orgID, "Test Budget", 2024, userID)
		require.NoError(t, err)
		err = b.AddLine(accountID, 100000, "Line")
		require.NoError(t, err)

		err = b.Approve(userID)
		require.NoError(t, err)
		assert.Equal(t, BudgetStatusApproved, b.Status)
		assert.NotNil(t, b.ApprovedBy)
		assert.Equal(t, userID, *b.ApprovedBy)
		assert.NotNil(t, b.ApprovedAt)
	})

	t.Run("cannot approve empty budget", func(t *testing.T) {
		b, err := NewBudget(orgID, "Empty Budget", 2024, userID)
		require.NoError(t, err)

		err = b.Approve(userID)
		assert.ErrorIs(t, err, budget.ErrBudgetHasNoLines)
	})

	t.Run("cannot approve already approved", func(t *testing.T) {
		b, err := NewBudget(orgID, "Test Budget", 2024, userID)
		require.NoError(t, err)
		err = b.AddLine(accountID, 10000, "Line")
		require.NoError(t, err)
		err = b.Approve(userID)
		require.NoError(t, err)

		err = b.Approve(userID)
		assert.ErrorIs(t, err, budget.ErrBudgetAlreadyApproved)
	})
}

func TestBudget_Activate(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	t.Run("activate approved budget", func(t *testing.T) {
		b, err := NewBudget(orgID, "Test Budget", 2024, userID)
		require.NoError(t, err)
		err = b.AddLine(accountID, 100000, "Line")
		require.NoError(t, err)
		err = b.Approve(userID)
		require.NoError(t, err)

		err = b.Activate()
		require.NoError(t, err)
		assert.Equal(t, BudgetStatusActive, b.Status)
	})

	t.Run("cannot activate draft budget", func(t *testing.T) {
		b, err := NewBudget(orgID, "Draft Budget", 2024, userID)
		require.NoError(t, err)
		err = b.AddLine(accountID, 10000, "Line")
		require.NoError(t, err)

		err = b.Activate()
		assert.ErrorIs(t, err, budget.ErrBudgetNotApproved)
	})
}

func TestBudget_Close(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	t.Run("close active budget", func(t *testing.T) {
		b, err := NewBudget(orgID, "Test Budget", 2024, userID)
		require.NoError(t, err)
		err = b.AddLine(accountID, 100000, "Line")
		require.NoError(t, err)
		err = b.Approve(userID)
		require.NoError(t, err)
		err = b.Activate()
		require.NoError(t, err)

		err = b.Close()
		require.NoError(t, err)
		assert.Equal(t, BudgetStatusClosed, b.Status)
	})

	t.Run("cannot close non-active budget", func(t *testing.T) {
		b, err := NewBudget(orgID, "Draft Budget", 2024, userID)
		require.NoError(t, err)
		err = b.AddLine(accountID, 10000, "Line")
		require.NoError(t, err)

		err = b.Close()
		assert.ErrorIs(t, err, budget.ErrBudgetNotActive)
	})
}