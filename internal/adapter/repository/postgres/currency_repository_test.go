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

func TestCurrencyRepository_CRUD(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewCurrencyRepository(testDB.DB)
	ctx := context.Background()

	currency := &entity.Currency{
		Name:   "Test Currency",
		Code:   "TST",
		Symbol: "T$",
	}

	err := repo.Create(ctx, currency)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, currency.ID)

	retrieved, err := repo.GetByID(ctx, currency.ID, false)
	require.NoError(t, err)
	assert.Equal(t, currency.Name, retrieved.Name)

	_, err = repo.GetByID(ctx, uuidv7.New(), false)
	assert.ErrorIs(t, err, entity.ErrNotFound)

	err = repo.Delete(ctx, currency.ID)
	require.NoError(t, err)
}
