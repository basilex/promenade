package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/account/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestAccountRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		account, err := aggregate.NewAccount(orgID, "1000", "Cash", aggregate.AccountTypeAsset, "USD", userID)
		require.NoError(t, err)

		err = repo.Create(ctx, account)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, account.GetID())
		require.NoError(t, err)
		assert.Equal(t, account.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "1000", found.Code)
		assert.Equal(t, "Cash", found.Name)
		assert.Equal(t, aggregate.AccountTypeAsset, found.Type)
		assert.Equal(t, "USD", found.CurrencyCode)
	})
}

func TestAccountRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.Error(t, err)
	})
}

func TestAccountRepository_GetByCode(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		account, _ := aggregate.NewAccount(orgID, "2000", "Accounts Receivable", aggregate.AccountTypeAsset, "USD", userID)
		require.NoError(t, repo.Create(ctx, account))

		found, err := repo.GetByCode(ctx, orgID, "2000")
		require.NoError(t, err)
		assert.Equal(t, account.GetID(), found.GetID())
		assert.Equal(t, "2000", found.Code)
	})
}

func TestAccountRepository_GetByCode_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		_, err := repo.GetByCode(ctx, orgID, "9999")
		assert.Error(t, err)
	})
}

func TestAccountRepository_Update(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		account, _ := aggregate.NewAccount(orgID, "3000", "Old Name", aggregate.AccountTypeAsset, "USD", userID)
		require.NoError(t, repo.Create(ctx, account))

		err := account.UpdateName("New Name")
		require.NoError(t, err)

		err = repo.Update(ctx, account)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, account.GetID())
		require.NoError(t, err)
		assert.Equal(t, "New Name", found.Name)
	})
}

func TestAccountRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		account, _ := aggregate.NewAccount(orgID, "4000", "To Delete", aggregate.AccountTypeAsset, "USD", userID)
		require.NoError(t, repo.Create(ctx, account))

		err := repo.Delete(ctx, account.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, account.GetID())
		assert.Error(t, err, "Soft deleted account should not be found")
	})
}

func TestAccountRepository_CreateMany(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		accounts := make([]*aggregate.Account, 3)
		for i := 0; i < 3; i++ {
			code := fmt.Sprintf("%d", 1000+i*100)
			acc, _ := aggregate.NewAccount(orgID, code, "Account "+code, aggregate.AccountTypeAsset, "USD", userID)
			accounts[i] = acc
		}

		err := repo.CreateMany(ctx, accounts)
		require.NoError(t, err)

		for _, acc := range accounts {
			found, err := repo.GetByID(ctx, acc.GetID())
			require.NoError(t, err)
			assert.Equal(t, acc.Code, found.Code)
		}
	})
}

func TestAccountRepository_ListByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create 3 active accounts
		for i := 0; i < 3; i++ {
			code := fmt.Sprintf("%d", 1000+i*100)
			acc, _ := aggregate.NewAccount(orgID, code, "Account", aggregate.AccountTypeAsset, "USD", userID)
			require.NoError(t, repo.Create(ctx, acc))
		}

		// Create 1 inactive account
		inactiveAcc, _ := aggregate.NewAccount(orgID, "9000", "Inactive", aggregate.AccountTypeAsset, "USD", userID)
		_ = inactiveAcc.Deactivate()
		require.NoError(t, repo.Create(ctx, inactiveAcc))

		// List only active
		activeAccounts, err := repo.ListByOrganization(ctx, orgID, false)
		require.NoError(t, err)
		assert.Len(t, activeAccounts, 3)

		// List all (including inactive)
		allAccounts, err := repo.ListByOrganization(ctx, orgID, true)
		require.NoError(t, err)
		assert.Len(t, allAccounts, 4)
	})
}

func TestAccountRepository_ListByType(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create assets
		for i := 0; i < 2; i++ {
			code := fmt.Sprintf("%d", 1000+i*100)
			acc, _ := aggregate.NewAccount(orgID, code, "Asset", aggregate.AccountTypeAsset, "USD", userID)
			require.NoError(t, repo.Create(ctx, acc))
		}

		// Create liabilities
		for i := 0; i < 3; i++ {
			code := fmt.Sprintf("%d", 2000+i*100)
			acc, _ := aggregate.NewAccount(orgID, code, "Liability", aggregate.AccountTypeLiability, "USD", userID)
			require.NoError(t, repo.Create(ctx, acc))
		}

		assets, err := repo.ListByType(ctx, orgID, aggregate.AccountTypeAsset)
		require.NoError(t, err)
		assert.Len(t, assets, 2)

		liabilities, err := repo.ListByType(ctx, orgID, aggregate.AccountTypeLiability)
		require.NoError(t, err)
		assert.Len(t, liabilities, 3)
	})
}

func TestAccountRepository_ListChildren(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create parent account
		parent, _ := aggregate.NewAccount(orgID, "1000", "Parent", aggregate.AccountTypeAsset, "USD", userID)
		require.NoError(t, repo.Create(ctx, parent))

		// Create child accounts
		for i := 0; i < 3; i++ {
			code := fmt.Sprintf("%d", 1010+i)
			child, _ := aggregate.NewChildAccount(orgID, code, "Child", aggregate.AccountTypeAsset, "USD", parent.GetID(), parent.Level, userID)
			require.NoError(t, repo.Create(ctx, child))
		}

		children, err := repo.ListChildren(ctx, parent.GetID())
		require.NoError(t, err)
		assert.Len(t, children, 3)

		for _, child := range children {
			assert.Equal(t, parent.GetID(), *child.ParentID)
		}
	})
}

func TestAccountRepository_GetAllActive(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewAccountRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create active accounts
		for i := 0; i < 5; i++ {
			code := fmt.Sprintf("%d", 1000+i*100)
			acc, _ := aggregate.NewAccount(orgID, code, "Active", aggregate.AccountTypeAsset, "USD", userID)
			require.NoError(t, repo.Create(ctx, acc))
		}

		// Create inactive account
		inactiveAcc, _ := aggregate.NewAccount(orgID, "9000", "Inactive", aggregate.AccountTypeAsset, "USD", userID)
		_ = inactiveAcc.Deactivate()
		active, err := repo.GetAllActive(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, active, 5)

		for _, acc := range active {
			assert.True(t, acc.IsActive)
		}
	})
}
