package bankaccount_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestBankAccountRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()
		acc, err := aggregate.NewBankAccount(orgID, "Test Account", "Test Bank", "UAH", uuidv7.New())
		require.NoError(t, err)

		err = repo.Create(ctx, acc)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, acc.GetID())
		require.NoError(t, err)
		assert.Equal(t, acc.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "Test Account", found.Name)
	})
}

func TestBankAccountRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.Error(t, err)
	})
}

func TestBankAccountRepository_Update(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()
		acc, _ := aggregate.NewBankAccount(orgID, "Old Name", "Old Bank", "UAH", uuidv7.New())
		require.NoError(t, repo.Create(ctx, acc))

		acc.Name = "New Name"
		acc.BalanceCents = 100000

		err := repo.Update(ctx, acc)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, acc.GetID())
		require.NoError(t, err)
		assert.Equal(t, "New Name", found.Name)
		assert.Equal(t, int64(100000), found.BalanceCents)
	})
}

func TestBankAccountRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()
		acc, _ := aggregate.NewBankAccount(orgID, "To Delete", "Delete Bank", "UAH", uuidv7.New())
		require.NoError(t, repo.Create(ctx, acc))

		err := repo.Delete(ctx, acc.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, acc.GetID())
		assert.Error(t, err, "Soft deleted account should not be found")
	})
}

func TestBankAccountRepository_List(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()

		for i := 0; i < 5; i++ {
			acc, _ := aggregate.NewBankAccount(orgID, "Account", "Test Bank", "UAH", uuidv7.New())
			require.NoError(t, repo.Create(ctx, acc))
		}

		accounts, err := repo.List(ctx, orgID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, accounts, 2)

		accounts, err = repo.List(ctx, orgID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, accounts, 5)
	})
}

func TestBankAccountRepository_GetByProviderAccountID(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()
		acc, _ := aggregate.NewBankAccount(orgID, "Provider Acc", "Provider Bank", "UAH", uuidv7.New())
		err := acc.ConnectProvider(aggregate.ProviderMonobank, "mono_12345", "UA1234", "1234")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, acc))

		found, err := repo.GetByProviderAccountID(ctx, aggregate.ProviderMonobank, "mono_12345")
		require.NoError(t, err)
		assert.Equal(t, acc.GetID(), found.GetID())
		assert.Equal(t, "mono_12345", found.ProviderAccountID)
	})
}

func TestBankAccountRepository_CountByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()

		for i := 0; i < 3; i++ {
			acc, _ := aggregate.NewBankAccount(orgID, "Account", "Test Bank", "UAH", uuidv7.New())
			require.NoError(t, repo.Create(ctx, acc))
		}

		count, err := repo.CountByOrganization(ctx, orgID)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

func TestBankAccountRepository_BalancePrecision(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()
		acc, _ := aggregate.NewBankAccount(orgID, "Precision Test", "Precision Bank", "UAH", uuidv7.New())

		acc.BalanceCents = 123456
		require.NoError(t, repo.Create(ctx, acc))

		found, err := repo.GetByID(ctx, acc.GetID())
		require.NoError(t, err)
		assert.Equal(t, int64(123456), found.BalanceCents,
			"Balance precision must be exact to the kopiyok! 1234.56 UAH = 123456 kopiyky")
	})
}

func TestBankAccountRepository_ProviderAccounts(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()

		mono, _ := aggregate.NewBankAccount(orgID, "Monobank", "Monobank", "UAH", uuidv7.New())
		_ = mono.ConnectProvider(aggregate.ProviderMonobank, "mono_123", "UA111", "111")
		require.NoError(t, repo.Create(ctx, mono))

		privat, _ := aggregate.NewBankAccount(orgID, "Privat24", "Privat24", "UAH", uuidv7.New())
		_ = privat.ConnectProvider(aggregate.ProviderPrivat24, "priv_456", "UA222", "222")
		require.NoError(t, repo.Create(ctx, privat))

		pumb, _ := aggregate.NewBankAccount(orgID, "PUMB", "PUMB", "UAH", uuidv7.New())
		_ = pumb.ConnectProvider(aggregate.ProviderPUMB, "pumb_789", "UA333", "333")
		require.NoError(t, repo.Create(ctx, pumb))

		foundMono, _ := repo.GetByProviderAccountID(ctx, aggregate.ProviderMonobank, "mono_123")
		assert.Equal(t, aggregate.ProviderMonobank, foundMono.Provider)

		foundPrivat, _ := repo.GetByProviderAccountID(ctx, aggregate.ProviderPrivat24, "priv_456")
		assert.Equal(t, aggregate.ProviderPrivat24, foundPrivat.Provider)

		foundPumb, _ := repo.GetByProviderAccountID(ctx, aggregate.ProviderPUMB, "pumb_789")
		assert.Equal(t, aggregate.ProviderPUMB, foundPumb.Provider)
	})
}

func TestBankAccountRepository_StatusTransitions(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankAccountRepository(db.DB)

		orgID := uuidv7.New()
		acc, _ := aggregate.NewBankAccount(orgID, "Status Test", "Status Bank", "UAH", uuidv7.New())
		require.NoError(t, repo.Create(ctx, acc))

		_ = acc.Deactivate()
		require.NoError(t, repo.Update(ctx, acc))
		found, err := repo.GetByID(ctx, acc.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.BankAccountStatusInactive, found.Status)

		_ = acc.Activate()
		require.NoError(t, repo.Update(ctx, acc))
		found, _ = repo.GetByID(ctx, acc.GetID())
		assert.Equal(t, aggregate.BankAccountStatusActive, found.Status)

		_ = acc.Archive()
		require.NoError(t, repo.Update(ctx, acc))
		found, _ = repo.GetByID(ctx, acc.GetID())
		assert.Equal(t, aggregate.BankAccountStatusArchived, found.Status)
	})
}
