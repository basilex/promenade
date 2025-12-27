package postgres

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/country"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestCountryRepository_Create(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := NewRepository(tx)

			c := &country.Country{
				ID:          uuidv7.New(),
				Code:        "TS",
				Code3:       "TST",
				NumericCode: "999",
				Name:        "Test Country",
				NameLocal:   "Test Country Local",
				PhoneCode:   "+999",
				IsActive:    true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, c.ID)
			require.NoError(t, err)
			assert.Equal(t, c.ID, found.ID)
			assert.Equal(t, "TS", found.Code)
			assert.Equal(t, "Test Country", found.Name)
		})
	})
}

func TestCountryRepository_GetByID(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := NewRepository(tx)

			nonExistentID := uuidv7.New()
			found, err := repo.GetByID(ctx, nonExistentID)
			assert.Error(t, err)
			assert.Equal(t, country.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}

func TestCountryRepository_GetByCode(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := NewRepository(tx)

			c := &country.Country{
				ID:          uuidv7.New(),
				Code:        "GC",
				Code3:       "GCT",
				NumericCode: "994",
				Name:        "GetByCode Test",
				NameLocal:   "GetByCode Test Local",
				PhoneCode:   "+994",
				IsActive:    true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			found, err := repo.GetByCode(ctx, "GC")
			require.NoError(t, err)
			assert.Equal(t, "GC", found.Code)
		})
	})
}

func TestCountryRepository_List(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("returns_active_only", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := NewRepository(tx)

			c1 := &country.Country{
				ID:          uuidv7.New(),
				Code:        "L1",
				Code3:       "LS1",
				NumericCode: "989",
				Name:        "List Test Alpha",
				NameLocal:   "List Alpha Local",
				PhoneCode:   "+989",
				IsActive:    true,
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

func TestCountryRepository_Update(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := NewRepository(tx)

			c := &country.Country{
				ID:          uuidv7.New(),
				Code:        "UP",
				Code3:       "UPD",
				NumericCode: "992",
				Name:        "Update Test",
				NameLocal:   "Update Test Local",
				PhoneCode:   "+992",
				IsActive:    true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			c.Name = "Updated Name"
			err = repo.Update(ctx, c)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, c.ID)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", found.Name)
		})
	})
}

func TestCountryRepository_Delete(t *testing.T) {
	testDB := integration.SetupTestDB(t)

	t.Run("soft_delete", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := NewRepository(tx)

			c := &country.Country{
				ID:          uuidv7.New(),
				Code:        "DL",
				Code3:       "DEL",
				NumericCode: "990",
				Name:        "Delete Test Country",
				NameLocal:   "Delete Test Local",
				PhoneCode:   "+990",
				IsActive:    true,
			}

			err := repo.Create(ctx, c)
			require.NoError(t, err)

			err = repo.Delete(ctx, c.ID)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, c.ID)
			assert.Error(t, err)
			assert.Equal(t, country.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}
