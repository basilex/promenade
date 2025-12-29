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

func TestTimezoneRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			tz := &timezone.Timezone{
				ID:           uuidv7.New(),
				Name:         "Test/Timezone",
				Abbreviation: "TST",
				UTCOffset:    18000, // +5 hours in seconds
				IsActive:     true,
			}

			err := repo.Create(ctx, tz)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, tz.ID)
			require.NoError(t, err)
			assert.Equal(t, tz.ID, found.ID)
			assert.Equal(t, "Test/Timezone", found.Name)
			assert.Equal(t, 18000, found.UTCOffset)
		})
	})
}

func TestTimezoneRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			nonExistentID := uuidv7.New()
			found, err := repo.GetByID(ctx, nonExistentID)
			assert.Error(t, err)
			assert.Equal(t, timezone.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}

func TestTimezoneRepository_GetByName(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			tz := &timezone.Timezone{
				ID:           uuidv7.New(),
				Name:         "Europe/TestCity",
				Abbreviation: "TTC",
				UTCOffset:    7200, // +2 hours
				IsActive:     true,
			}

			err := repo.Create(ctx, tz)
			require.NoError(t, err)

			found, err := repo.GetByName(ctx, "Europe/TestCity")
			require.NoError(t, err)
			assert.Equal(t, "Europe/TestCity", found.Name)
		})
	})
}

func TestTimezoneRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("returns_active_only", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			tz1 := &timezone.Timezone{
				ID:           uuidv7.New(),
				Name:         "Test/Alpha",
				Abbreviation: "TA",
				UTCOffset:    3600, // +1 hour
				IsActive:     true,
			}

			err := repo.Create(ctx, tz1)
			require.NoError(t, err)

			found, err := repo.List(ctx)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(found), 1)

			for _, tz := range found {
				assert.True(t, tz.IsActive)
			}
		})
	})
}

func TestTimezoneRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			tz := &timezone.Timezone{
				ID:           uuidv7.New(),
				Name:         "Update/Test",
				Abbreviation: "UT",
				UTCOffset:    10800, // +3 hours
				IsActive:     true,
			}

			err := repo.Create(ctx, tz)
			require.NoError(t, err)

			tz.Abbreviation = "UPD"
			tz.UTCOffset = 14400 // +4 hours
			err = repo.Update(ctx, tz)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, tz.ID)
			require.NoError(t, err)
			assert.Equal(t, "UPD", found.Abbreviation)
			assert.Equal(t, 14400, found.UTCOffset)
		})
	})
}

func TestTimezoneRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("soft_delete", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			tz := &timezone.Timezone{
				ID:           uuidv7.New(),
				Name:         "Delete/Test",
				Abbreviation: "DEL",
				UTCOffset:    39600, // +11 hours
				IsActive:     true,
			}

			err := repo.Create(ctx, tz)
			require.NoError(t, err)

			err = repo.Delete(ctx, tz.ID)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, tz.ID)
			assert.Error(t, err)
			assert.Equal(t, timezone.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}
