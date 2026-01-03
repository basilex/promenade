package currency_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/internal/contexts/shared/currency/adapter/repository/postgres"
	"github.com/basilex/promenade/test/integration"
)

func TestCurrencyRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		c := func() *currency.Currency {
			c, _ := currency.NewCurrency("TST", "Test Currency", "T$", 2)
			c.NumericCode = "999"
			return c
		}()
		require.NoError(t, repo.Create(ctx, c))
		found, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "TST", found.Code)
		found.Name = "Updated"
		require.NoError(t, repo.Update(ctx, found))
		require.NoError(t, repo.Delete(ctx, c.ID))
		_, err = repo.GetByID(ctx, c.ID)
		assert.Error(t, err)
	})
}

func TestCurrencyRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		c := func() *currency.Currency {
			c, _ := currency.NewCurrency("TST", "Test", "$", 2)
			c.NumericCode = "999"
			return c
		}()
		require.NoError(t, repo.Create(ctx, c))
		found, err := repo.GetByCode(ctx, "TST")
		require.NoError(t, err)
		assert.Equal(t, "Test", found.Name)
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Greater(t, len(all), 0)
	})
}
