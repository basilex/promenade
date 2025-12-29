package language_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	"github.com/basilex/promenade/internal/contexts/shared/language/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestLanguageRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		l := &language.Language{
			ID:         uuidv7.New(),
			Code:       "ts",
			Code3:      "tst",
			Name:       "Test",
			NativeName: "Local",
			IsActive:   true,
		}
		require.NoError(t, repo.Create(ctx, l))
		found, err := repo.GetByID(ctx, l.ID)
		require.NoError(t, err)
		assert.Equal(t, "ts", found.Code)
		found.Name = "Updated"
		require.NoError(t, repo.Update(ctx, found))
		require.NoError(t, repo.Delete(ctx, l.ID))
		_, err = repo.GetByID(ctx, l.ID)
		assert.Error(t, err)
	})
}

func TestLanguageRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		l := &language.Language{
			ID:         uuidv7.New(),
			Code:       "ts",
			Code3:      "tst",
			Name:       "Test",
			NativeName: "Local",
			IsActive:   true,
		}
		require.NoError(t, repo.Create(ctx, l))
		found, err := repo.GetByCode(ctx, "ts")
		require.NoError(t, err)
		assert.Equal(t, "Test", found.Name)
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Greater(t, len(all), 0)
	})
}
