package country_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/country"
	"github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository/postgres"
	countryAggregate "github.com/basilex/promenade/internal/contexts/shared/country/aggregate"
	"github.com/basilex/promenade/test/integration"
)

// TestCountryRepository_CRUD tests Create, Read, Update, Delete operations
func TestCountryRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)

		// Create
		c := func() *countryAggregate.Country {
			c, _ := countryAggregate.NewCountry("TS", "Test Country", "+999")
			c.Code3 = "TST"
			c.NumericCode = "999"
			c.NameLocal = "Test Country Local"
			return c
		}()

		err := repo.Create(ctx, c)
		require.NoError(t, err, "Create should succeed")

		// Read
		found, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err, "GetByID should succeed")
		assert.Equal(t, c.ID, found.ID)
		assert.Equal(t, "TS", found.Code)
		assert.Equal(t, "Test Country", found.Name)

		// Update
		found.Name = "Updated Country"
		err = repo.Update(ctx, found)
		require.NoError(t, err, "Update should succeed")

		updated, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Country", updated.Name)

		// Delete
		err = repo.Delete(ctx, c.ID)
		require.NoError(t, err, "Delete should succeed")

		deleted, err := repo.GetByID(ctx, c.ID)
		assert.Error(t, err, "GetByID should fail after delete")
		assert.Equal(t, country.ErrCountryNotFound, err)
		assert.Nil(t, deleted)
	})
}

// TestCountryRepository_Queries tests query and list operations
func TestCountryRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)

		// Create test data
		countries := []*countryAggregate.Country{
			func() *countryAggregate.Country {
				c, _ := countryAggregate.NewCountry("T1", "Test Country 1", "+1")
				c.Code3 = "TS1"
				c.NumericCode = "001"
				c.NameLocal = "Local 1"
				return c
			}(),
			func() *countryAggregate.Country {
				c, _ := countryAggregate.NewCountry("T2", "Test Country 2", "+2")
				c.Code3 = "TS2"
				c.NumericCode = "002"
				c.NameLocal = "Local 2"
				return c
			}(),
		}

		for _, c := range countries {
			err := repo.Create(ctx, c)
			require.NoError(t, err)
		}

		// GetByCode - found
		found, err := repo.GetByCode(ctx, "T1")
		require.NoError(t, err)
		assert.Equal(t, "Test Country 1", found.Name)

		// GetByCode - not found
		notFound, err := repo.GetByCode(ctx, "XX")
		assert.Error(t, err)
		assert.Equal(t, country.ErrCountryNotFound, err)
		assert.Nil(t, notFound)

		// List
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 2, "Should have at least 2 countries")
	})
}
