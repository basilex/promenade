package timezone_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/internal/contexts/shared/timezone/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestTimezoneRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		tz := &timezone.Timezone{
			ID:           uuidv7.New(),
			Name:         "Test/Zone",
			Abbreviation: "TST",
			UTCOffset:    0,
			IsActive:     true,
		}
		require.NoError(t, repo.Create(ctx, tz))
		found, err := repo.GetByID(ctx, tz.ID)
		require.NoError(t, err)
		assert.Equal(t, "Test/Zone", found.Name)
		found.UTCOffset = 3600
		require.NoError(t, repo.Update(ctx, found))
		require.NoError(t, repo.Delete(ctx, tz.ID))
		_, err = repo.GetByID(ctx, tz.ID)
		assert.Error(t, err)
	})
}

func TestTimezoneRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		tz := &timezone.Timezone{
			ID:           uuidv7.New(),
			Name:         "Test/Zone",
			Abbreviation: "TST",
			UTCOffset:    0,
			IsActive:     true,
		}
		require.NoError(t, repo.Create(ctx, tz))
		found, err := repo.GetByName(ctx, "Test/Zone")
		require.NoError(t, err)
		assert.Equal(t, 0, found.UTCOffset)
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Greater(t, len(all), 0)
	})
}
