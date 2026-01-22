package budget_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/budget"
	"github.com/basilex/promenade/internal/contexts/accounting/budget/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestBudgetRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		b, err := aggregate.NewBudget(orgID, "Annual Budget 2026", 2026, userID)
		require.NoError(t, err)

		err = repo.Create(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Equal(t, b.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "Annual Budget 2026", found.Name)
		assert.Equal(t, 2026, found.FiscalYear)
		assert.Equal(t, aggregate.BudgetStatusDraft, found.Status)
		assert.Equal(t, int64(0), found.TotalBudget)
		assert.Equal(t, int64(0), found.TotalActual)
	})
}

func TestBudgetRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.ErrorIs(t, err, budget.ErrBudgetNotFound)
	})
}

func TestBudgetRepository_GetByName(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		b, _ := aggregate.NewBudget(orgID, "Q1 Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		found, err := repo.GetByName(ctx, orgID, "Q1 Budget")
		require.NoError(t, err)
		assert.Equal(t, b.GetID(), found.GetID())
		assert.Equal(t, "Q1 Budget", found.Name)
	})
}

func TestBudget_AddLine(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create GL accounts
		accountID1, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Salaries")
		require.NoError(t, err)
		accountID2, err := createTestGLAccount(ctx, tx, orgID, userID, "6120", "Office supplies")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Add budget lines
		err = b.AddLine(accountID1, 100000000, "Salaries") // 1,000,000.00
		require.NoError(t, err)

		err = b.AddLine(accountID2, 50000000, "Office supplies") // 500,000.00
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Len(t, found.Lines, 2)
		assert.Equal(t, int64(150000000), found.TotalBudget) // 1,500,000.00

		// Check that both descriptions are present (order may vary)
		descriptions := []string{found.Lines[0].Description, found.Lines[1].Description}
		assert.Contains(t, descriptions, "Salaries")
		assert.Contains(t, descriptions, "Office supplies")
	})
}

func TestBudget_UpdateLine(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Salaries")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Add line
		require.NoError(t, b.AddLine(accountID, 100000000, "Original amount"))
		lineID := b.Lines[0].ID
		require.NoError(t, repo.Update(ctx, b))

		// Update line
		err = b.UpdateLine(lineID, 150000000)
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Equal(t, int64(150000000), found.Lines[0].BudgetAmount)
		assert.Equal(t, int64(150000000), found.TotalBudget)
	})
}

func TestBudget_RemoveLine(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID1, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Account 1")
		require.NoError(t, err)
		accountID2, err := createTestGLAccount(ctx, tx, orgID, userID, "6120", "Account 2")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Add two lines
		require.NoError(t, b.AddLine(accountID1, 100000000, "Line 1"))
		require.NoError(t, b.AddLine(accountID2, 50000000, "Line 2"))
		lineIDToRemove := b.Lines[0].ID
		require.NoError(t, repo.Update(ctx, b))

		// Remove first line
		err = b.RemoveLine(lineIDToRemove)
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Len(t, found.Lines, 1)
		assert.Equal(t, "Line 2", found.Lines[0].Description)
		assert.Equal(t, int64(50000000), found.TotalBudget)
	})
}

func TestBudget_Approve(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		approverID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Budget account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Add line (required for approval)
		require.NoError(t, b.AddLine(accountID, 100000000, "Budget item"))
		require.NoError(t, repo.Update(ctx, b))

		// Approve
		err = b.Approve(approverID)
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.BudgetStatusApproved, found.Status)
		assert.NotNil(t, found.ApprovedBy)
		assert.Equal(t, approverID, *found.ApprovedBy)
	})
}

func TestBudget_Activate(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Budget account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Add line and approve
		require.NoError(t, b.AddLine(accountID, 100000000, "Budget item"))
		require.NoError(t, b.Approve(userID))
		require.NoError(t, repo.Update(ctx, b))

		// Activate
		err = b.Activate()
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.BudgetStatusActive, found.Status)
	})
}

func TestBudget_Close(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Budget account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Add line, approve, and activate
		require.NoError(t, b.AddLine(accountID, 100000000, "Budget item"))
		require.NoError(t, b.Approve(userID))
		require.NoError(t, b.Activate())
		require.NoError(t, repo.Update(ctx, b))

		// Close
		err = b.Close()
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.BudgetStatusClosed, found.Status)
	})
}

func TestBudget_UpdateActuals(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Budget account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Add line, approve, and activate
		require.NoError(t, b.AddLine(accountID, 100000000, "Budget item"))
		lineID := b.Lines[0].ID
		require.NoError(t, b.Approve(userID))
		require.NoError(t, b.Activate())
		require.NoError(t, repo.Update(ctx, b))

		// Update actuals
		err = b.UpdateActuals(lineID, 80000000) // 80% of budget
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Equal(t, int64(80000000), found.Lines[0].ActualAmount)
		assert.Equal(t, int64(-20000000), found.Lines[0].VarianceAmount) // Under budget
		assert.Equal(t, int64(80000000), found.TotalActual)
	})
}

func TestBudget_VarianceCalculation(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Budget account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		// Budget: 1,000,000.00
		require.NoError(t, b.AddLine(accountID, 100000000, "Budget item"))
		lineID := b.Lines[0].ID
		require.NoError(t, b.Approve(userID))
		require.NoError(t, b.Activate())
		require.NoError(t, repo.Update(ctx, b))

		// Actual: 1,200,000.00 (20% over budget)
		err = b.UpdateActuals(lineID, 120000000)
		require.NoError(t, err)

		err = repo.Update(ctx, b)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, b.GetID())
		require.NoError(t, err)
		assert.Equal(t, int64(20000000), found.Lines[0].VarianceAmount) // Over budget
		assert.Equal(t, 2000, found.Lines[0].VariancePercent)           // 20% (in basis points)
	})
}

func TestBudgetRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		b, _ := aggregate.NewBudget(orgID, "Temporary Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, b))

		err := repo.Delete(ctx, b.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, b.GetID())
		assert.ErrorIs(t, err, budget.ErrBudgetNotFound)
	})
}

func TestBudgetRepository_ListByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create multiple budgets
		budgets := []string{"Budget 2024", "Budget 2025", "Budget 2026"}
		for i, name := range budgets {
			b, _ := aggregate.NewBudget(orgID, name, 2024+i, userID)
			require.NoError(t, repo.Create(ctx, b))
		}

		found, err := repo.ListByOrganization(ctx, orgID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, found, 3)
	})
}

func TestBudgetRepository_ListByStatus(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Test account")
		require.NoError(t, err)

		// Draft budget
		draft, _ := aggregate.NewBudget(orgID, "Draft Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, draft))

		// Approved budget
		approved, _ := aggregate.NewBudget(orgID, "Approved Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, approved))
		require.NoError(t, approved.AddLine(accountID, 100000000, "Item"))
		require.NoError(t, approved.Approve(userID))
		require.NoError(t, repo.Update(ctx, approved))

		// Active budget
		active, _ := aggregate.NewBudget(orgID, "Active Budget", 2026, userID)
		require.NoError(t, repo.Create(ctx, active))
		require.NoError(t, active.AddLine(accountID, 100000000, "Item"))
		require.NoError(t, active.Approve(userID))
		require.NoError(t, active.Activate())
		require.NoError(t, repo.Update(ctx, active))

		// List approved budgets
		found, err := repo.ListByStatus(ctx, orgID, aggregate.BudgetStatusApproved)
		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, aggregate.BudgetStatusApproved, found[0].Status)
	})
}

func TestBudgetRepository_ListActive(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBudgetRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Test account")
		require.NoError(t, err)

		// Draft budget
		draft, _ := aggregate.NewBudget(orgID, "Draft", 2026, userID)
		require.NoError(t, repo.Create(ctx, draft))

		// Active budget
		active, _ := aggregate.NewBudget(orgID, "Active", 2026, userID)
		require.NoError(t, repo.Create(ctx, active))
		require.NoError(t, active.AddLine(accountID, 100000000, "Item"))
		require.NoError(t, active.Approve(userID))
		require.NoError(t, active.Activate())
		require.NoError(t, repo.Update(ctx, active))

		found, err := repo.ListActive(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, aggregate.BudgetStatusActive, found[0].Status)
	})
}

func TestBudget_BusinessRules_EmptyName(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		_, err := aggregate.NewBudget(orgID, "", 2026, userID)
		assert.ErrorIs(t, err, budget.ErrBudgetNameEmpty)
	})
}

func TestBudget_BusinessRules_InvalidFiscalYear(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		_, err := aggregate.NewBudget(orgID, "Budget", 1999, userID)
		assert.ErrorIs(t, err, budget.ErrInvalidFiscalYear)

		_, err = aggregate.NewBudget(orgID, "Budget", 2101, userID)
		assert.ErrorIs(t, err, budget.ErrInvalidFiscalYear)
	})
}

func TestBudget_BusinessRules_CannotModifyApproved(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Test account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, b.AddLine(accountID, 100000000, "Item"))
		require.NoError(t, b.Approve(userID))

		// Try to add line after approval
		err = b.AddLine(uuidv7.New(), 50000000, "New item")
		assert.ErrorIs(t, err, budget.ErrCannotModifyApprovedBudget)
	})
}

func TestBudget_BusinessRules_CannotApproveWithoutLines(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		b, _ := aggregate.NewBudget(orgID, "Empty Budget", 2026, userID)

		err := b.Approve(userID)
		assert.ErrorIs(t, err, budget.ErrBudgetHasNoLines)
	})
}

func TestBudget_BusinessRules_CannotActivateWithoutApproval(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Test account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, b.AddLine(accountID, 100000000, "Item"))

		// Try to activate without approval
		err = b.Activate()
		assert.ErrorIs(t, err, budget.ErrBudgetNotApproved)
	})
}

func TestBudget_BusinessRules_CannotUpdateActualsWhenNotActive(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Test account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, b.AddLine(accountID, 100000000, "Item"))
		lineID := b.Lines[0].ID

		// Try to update actuals when budget is draft
		err = b.UpdateActuals(lineID, 50000000)
		assert.ErrorIs(t, err, budget.ErrBudgetNotActive)
	})
}

func TestBudget_BusinessRules_DuplicateAccount(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID, err := createTestGLAccount(ctx, tx, orgID, userID, "6110", "Test account")
		require.NoError(t, err)

		b, _ := aggregate.NewBudget(orgID, "Test Budget", 2026, userID)
		require.NoError(t, b.AddLine(accountID, 100000000, "First item"))

		// Try to add same account again
		err = b.AddLine(accountID, 50000000, "Duplicate item")
		assert.ErrorIs(t, err, budget.ErrAccountAlreadyExists)
	})
}
