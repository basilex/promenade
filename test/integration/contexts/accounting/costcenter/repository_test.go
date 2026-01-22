package costcenter_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/costcenter"
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestCostCenterRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		cc, err := aggregate.NewCostCenter(
			orgID,
			"CC-100",
			"Sales Department",
			aggregate.CenterTypeCost,
			userID,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, cc)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, cc.GetID())
		require.NoError(t, err)
		assert.Equal(t, cc.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "CC-100", found.Code)
		assert.Equal(t, "Sales Department", found.Name)
		assert.Equal(t, aggregate.CenterTypeCost, found.CenterType)
		assert.Equal(t, 1, found.Level)
		assert.True(t, found.IsActive)
	})
}

func TestCostCenterRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.ErrorIs(t, err, costcenter.ErrCostCenterNotFound)
	})
}

func TestCostCenterRepository_GetByCode(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		cc, _ := aggregate.NewCostCenter(orgID, "CC-200", "IT Department", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, cc))

		found, err := repo.GetByCode(ctx, orgID, "CC-200")
		require.NoError(t, err)
		assert.Equal(t, cc.GetID(), found.GetID())
		assert.Equal(t, "CC-200", found.Code)
	})
}

func TestCostCenterRepository_Update_Details(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		cc, _ := aggregate.NewCostCenter(orgID, "CC-300", "Old Name", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, cc))

		// Update details
		err := cc.UpdateDetails("New Name", "Updated description", userID)
		require.NoError(t, err)

		err = repo.Update(ctx, cc)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, cc.GetID())
		require.NoError(t, err)
		assert.Equal(t, "New Name", found.Name)
		assert.Equal(t, "Updated description", found.Description)
	})
}

func TestCostCenterRepository_Update_ActivateDeactivate(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		cc, _ := aggregate.NewCostCenter(orgID, "CC-400", "Test Center", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, cc))

		// Deactivate
		cc.Deactivate()
		err := repo.Update(ctx, cc)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, cc.GetID())
		require.NoError(t, err)
		assert.False(t, found.IsActive)

		// Reactivate
		cc.Activate()
		err = repo.Update(ctx, cc)
		require.NoError(t, err)

		found, err = repo.GetByID(ctx, cc.GetID())
		require.NoError(t, err)
		assert.True(t, found.IsActive)
	})
}

func TestCostCenterRepository_Update_SetManager(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		managerID := uuidv7.New()

		cc, _ := aggregate.NewCostCenter(orgID, "CC-500", "Managed Center", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, cc))

		// Set manager
		cc.SetManager(managerID)

		err := repo.Update(ctx, cc)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, cc.GetID())
		require.NoError(t, err)
		assert.NotNil(t, found.ManagerID)
		assert.Equal(t, managerID, *found.ManagerID)
	})
}

func TestCostCenterRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		cc, _ := aggregate.NewCostCenter(orgID, "CC-600", "Temporary Center", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, cc))

		err := repo.Delete(ctx, cc.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, cc.GetID())
		assert.ErrorIs(t, err, costcenter.ErrCostCenterNotFound)
	})
}

func TestCostCenterRepository_HierarchicalStructure(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create parent
		parent, _ := aggregate.NewCostCenter(orgID, "CC-700", "Company", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, parent))

		// Create children
		child1, err := aggregate.NewChildCostCenter(orgID, "CC-701", "Department 1", aggregate.CenterTypeCost, parent.GetID(), parent.Level, userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, child1))

		child2, err := aggregate.NewChildCostCenter(orgID, "CC-702", "Department 2", aggregate.CenterTypeCost, parent.GetID(), parent.Level, userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, child2))

		// Create grandchild
		grandchild, err := aggregate.NewChildCostCenter(orgID, "CC-7011", "Team 1", aggregate.CenterTypeCost, child1.GetID(), child1.Level, userID)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, grandchild))

		// Verify hierarchy
		assert.Equal(t, 1, parent.Level)
		assert.Equal(t, 2, child1.Level)
		assert.Equal(t, 2, child2.Level)
		assert.Equal(t, 3, grandchild.Level)

		// List children of parent
		children, err := repo.ListChildren(ctx, parent.GetID())
		require.NoError(t, err)
		assert.Len(t, children, 2)

		// List children of child1
		grandchildren, err := repo.ListChildren(ctx, child1.GetID())
		require.NoError(t, err)
		assert.Len(t, grandchildren, 1)
	})
}

func TestCostCenterRepository_ListByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create multiple cost centers
		centers := []struct {
			code string
			name string
		}{
			{"CC-800", "Center 1"},
			{"CC-801", "Center 2"},
			{"CC-802", "Center 3"},
		}

		for _, c := range centers {
			cc, _ := aggregate.NewCostCenter(orgID, c.code, c.name, aggregate.CenterTypeCost, userID)
			require.NoError(t, repo.Create(ctx, cc))
		}

		found, err := repo.ListByOrganization(ctx, orgID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, found, 3)
	})
}

func TestCostCenterRepository_ListByType(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create different types
		costCenter, _ := aggregate.NewCostCenter(orgID, "CC-900", "Cost Center", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, costCenter))

		profitCenter, _ := aggregate.NewCostCenter(orgID, "PC-900", "Profit Center", aggregate.CenterTypeProfit, userID)
		require.NoError(t, repo.Create(ctx, profitCenter))

		investCenter, _ := aggregate.NewCostCenter(orgID, "IC-900", "Investment Center", aggregate.CenterTypeInvestment, userID)
		require.NoError(t, repo.Create(ctx, investCenter))

		// List cost centers only
		found, err := repo.ListByType(ctx, orgID, aggregate.CenterTypeCost)
		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, aggregate.CenterTypeCost, found[0].CenterType)
	})
}

func TestCostCenterRepository_ListActive(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create active and inactive centers
		active1, _ := aggregate.NewCostCenter(orgID, "CC-1000", "Active 1", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, active1))

		inactive, _ := aggregate.NewCostCenter(orgID, "CC-1001", "Inactive", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, inactive))
		inactive.Deactivate()
		require.NoError(t, repo.Update(ctx, inactive))

		active2, _ := aggregate.NewCostCenter(orgID, "CC-1002", "Active 2", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, active2))

		// List only active
		found, err := repo.ListActive(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, found, 2)
		for _, cc := range found {
			assert.True(t, cc.IsActive)
		}
	})
}

func TestCostCenter_BusinessRules_EmptyCodeOrName(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Empty code
		_, err := aggregate.NewCostCenter(orgID, "", "Test Center", aggregate.CenterTypeCost, userID)
		assert.ErrorIs(t, err, costcenter.ErrCodeRequired)

		// Empty name
		_, err = aggregate.NewCostCenter(orgID, "CC-TEST", "", aggregate.CenterTypeCost, userID)
		assert.ErrorIs(t, err, costcenter.ErrNameRequired)
	})
}

func TestCostCenter_BusinessRules_InvalidCenterType(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Invalid type
		_, err := aggregate.NewCostCenter(orgID, "CC-TEST", "Test Center", "invalid_type", userID)
		assert.ErrorIs(t, err, costcenter.ErrInvalidCenterType)
	})
}

func TestCostCenter_BusinessRules_CannotBeOwnParent(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		cc, _ := aggregate.NewCostCenter(orgID, "CC-TEST", "Test Center", aggregate.CenterTypeCost, userID)

		// Try to set itself as parent
		err := cc.SetParent(cc.GetID(), 1)
		assert.ErrorIs(t, err, costcenter.ErrCannotBeOwnParent)
	})
}

func TestCostCenter_SetParent(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create parent
		parent, _ := aggregate.NewCostCenter(orgID, "CC-PARENT", "Parent", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, parent))

		// Create child without parent
		child, _ := aggregate.NewCostCenter(orgID, "CC-CHILD", "Child", aggregate.CenterTypeCost, userID)
		require.NoError(t, repo.Create(ctx, child))

		// Set parent later
		err := child.SetParent(parent.GetID(), parent.Level)
		require.NoError(t, err)

		err = repo.Update(ctx, child)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, child.GetID())
		require.NoError(t, err)
		assert.NotNil(t, found.ParentID)
		assert.Equal(t, parent.GetID(), *found.ParentID)
		assert.Equal(t, 2, found.Level)
	})
}

func TestCostCenter_AllTypes(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCostCenterRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		types := []aggregate.CenterType{
			aggregate.CenterTypeCost,
			aggregate.CenterTypeProfit,
			aggregate.CenterTypeInvestment,
		}

		for _, centerType := range types {
			cc, err := aggregate.NewCostCenter(
				orgID,
				string(centerType),
				string(centerType)+" Center",
				centerType,
				userID,
			)
			require.NoError(t, err)
			require.NoError(t, repo.Create(ctx, cc))

			found, err := repo.GetByCode(ctx, orgID, string(centerType))
			require.NoError(t, err)
			assert.Equal(t, centerType, found.CenterType)
		}
	})
}
