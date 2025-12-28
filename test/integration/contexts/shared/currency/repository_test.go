package currency_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/internal/contexts/shared/currency/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestCurrencyRepository_Create(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			uniqueID := uuidv7.New().String()[:4]
			c := &currency.Currency{
				ID:            uuidv7.New(),
				Code:          fmt.Sprintf("T%s", uniqueID[:2]),
				NumericCode:   fmt.Sprintf("9%s", uniqueID[2:4]),
				Name:          "Test Currency",
				Symbol:        "Ŧ",
				DecimalPlaces: 2,
				IsActive:      true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, c.ID)
			require.NoError(t, err)
			assert.Equal(t, c.ID, found.ID)
			assert.NotEmpty(t, found.Code)
			assert.Equal(t, 2, found.DecimalPlaces)
		})
	})
}

func TestCurrencyRepository_GetByID(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			nonExistentID := uuidv7.New()
			found, err := repo.GetByID(ctx, nonExistentID)
			assert.Error(t, err)
			assert.Equal(t, currency.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}

func TestCurrencyRepository_GetByCode(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			c := &currency.Currency{
				ID:            uuidv7.New(),
				Code:          "GBC",
				NumericCode:   "994",
				Name:          "GetByCode Test",
				Symbol:        "Ģ",
				DecimalPlaces: 2,
				IsActive:      true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			found, err := repo.GetByCode(ctx, "GBC")
			require.NoError(t, err)
			assert.Equal(t, "GBC", found.Code)
		})
	})
}

func TestCurrencyRepository_List(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("returns_active_only", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			c1 := &currency.Currency{
				ID:            uuidv7.New(),
				Code:          "LS1",
				NumericCode:   "989",
				Name:          "List Test",
				Symbol:        "Ł",
				DecimalPlaces: 2,
				IsActive:      true,
			}

			err := repo.Create(ctx, c1)
			require.NoError(t, err)

			found, err := repo.List(ctx)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(found), 1)

			for _, c := range found {
				assert.True(t, c.IsActive)
			}
		})
	})
}

func TestCurrencyRepository_Update(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			c := &currency.Currency{
				ID:            uuidv7.New(),
				Code:          "UPD",
				NumericCode:   "992",
				Name:          "Update Test",
				Symbol:        "Ü",
				DecimalPlaces: 2,
				IsActive:      true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			c.Name = "Updated Name"
			c.DecimalPlaces = 3
			err = repo.Update(ctx, c)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, c.ID)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", found.Name)
			assert.Equal(t, 3, found.DecimalPlaces)
		})
	})
}

func TestCurrencyRepository_Delete(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("soft_delete", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			c := &currency.Currency{
				ID:            uuidv7.New(),
				Code:          "DEL",
				NumericCode:   "990",
				Name:          "Delete Test",
				Symbol:        "Ð",
				DecimalPlaces: 2,
				IsActive:      true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			err = repo.Delete(ctx, c.ID)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, c.ID)
			assert.Error(t, err)
			assert.Equal(t, currency.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}
