package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
)

func TestCountryRepository_CRUD(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewCountryRepository(testDB.DB)
	ctx := context.Background()

	country := &entity.Country{
		Name:   "Test Country",
		Code:   "TST",
		ISO2:   "TS",
		ISO3:   "TST",
		Region: "north_america",
	}

	err := repo.Create(ctx, country)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, country.ID)

	retrieved, err := repo.GetByID(ctx, country.ID, false)
	require.NoError(t, err)
	assert.Equal(t, country.Name, retrieved.Name)

	_, err = repo.GetByID(ctx, uuidv7.New(), false)
	assert.ErrorIs(t, err, entity.ErrNotFound)

	err = repo.Delete(ctx, country.ID)
	require.NoError(t, err)
}
