package fiscalperiod_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestFiscalPeriodRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

		period, err := aggregate.NewFiscalPeriod(
			orgID,
			"2026-01",
			"January 2026",
			aggregate.PeriodTypeMonth,
			startDate,
			endDate,
			userID,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, period)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, period.GetID())
		require.NoError(t, err)
		assert.Equal(t, period.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "2026-01", found.Code)
		assert.Equal(t, "January 2026", found.Name)
		assert.Equal(t, aggregate.PeriodTypeMonth, found.PeriodType)
		assert.Equal(t, startDate.Format("2006-01-02"), found.StartDate.Format("2006-01-02"))
		assert.Equal(t, endDate.Format("2006-01-02"), found.EndDate.Format("2006-01-02"))
		assert.Equal(t, aggregate.PeriodStatusOpen, found.Status)
	})
}

func TestFiscalPeriodRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotFound)
	})
}

func TestFiscalPeriodRepository_Update_ClosePeriod(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-02", "February 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)
		require.NoError(t, repo.Create(ctx, period))

		// Close the period
		err := period.Close(userID)
		require.NoError(t, err)

		err = repo.Update(ctx, period)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, period.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.PeriodStatusClosed, found.Status)
		assert.NotNil(t, found.ClosedBy)
		assert.NotNil(t, found.ClosedAt)
	})
}

func TestFiscalPeriodRepository_Update_ReopenPeriod(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-03", "March 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)
		require.NoError(t, repo.Create(ctx, period))

		// Close and then reopen
		require.NoError(t, period.Close(userID))
		require.NoError(t, repo.Update(ctx, period))

		err := period.Reopen(userID)
		require.NoError(t, err)

		err = repo.Update(ctx, period)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, period.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.PeriodStatusOpen, found.Status)
		assert.NotNil(t, found.ReopenedBy)
		assert.NotNil(t, found.ReopenedAt)
	})
}

func TestFiscalPeriodRepository_Update_LockPeriod(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-04", "April 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)
		require.NoError(t, repo.Create(ctx, period))

		// Close first, then lock
		require.NoError(t, period.Close(userID))
		require.NoError(t, repo.Update(ctx, period))

		err := period.Lock(userID)
		require.NoError(t, err)

		err = repo.Update(ctx, period)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, period.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.PeriodStatusLocked, found.Status)
		assert.NotNil(t, found.LockedBy)
		assert.NotNil(t, found.LockedAt)
	})
}

func TestFiscalPeriodRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-05", "May 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)
		require.NoError(t, repo.Create(ctx, period))

		err := repo.Delete(ctx, period.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, period.GetID())
		assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotFound)
	})
}

func TestFiscalPeriodRepository_ListByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create multiple periods
		periods := []struct {
			code      string
			name      string
			startDate time.Time
			endDate   time.Time
		}{
			{"2026-01", "January 2026", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)},
			{"2026-02", "February 2026", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)},
			{"2026-03", "March 2026", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)},
		}

		for _, p := range periods {
			period, _ := aggregate.NewFiscalPeriod(orgID, p.code, p.name, aggregate.PeriodTypeMonth, p.startDate, p.endDate, userID)
			require.NoError(t, repo.Create(ctx, period))
		}

		found, err := repo.ListByOrganization(ctx, orgID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, found, 3)
	})
}

func TestFiscalPeriodRepository_ListByYear(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create periods for 2026 and 2027
		period2026, _ := aggregate.NewFiscalPeriod(orgID, "2026-06", "June 2026", aggregate.PeriodTypeMonth,
			time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), userID)
		require.NoError(t, repo.Create(ctx, period2026))

		period2027, _ := aggregate.NewFiscalPeriod(orgID, "2027-01", "January 2027", aggregate.PeriodTypeMonth,
			time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC), userID)
		require.NoError(t, repo.Create(ctx, period2027))

		found, err := repo.ListByYear(ctx, orgID, 2026)
		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, "2026-06", found[0].Code)
	})
}

func TestFiscalPeriodRepository_ListOpen(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create open and closed periods
		openPeriod, _ := aggregate.NewFiscalPeriod(orgID, "2026-07", "July 2026", aggregate.PeriodTypeMonth,
			time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC), userID)
		require.NoError(t, repo.Create(ctx, openPeriod))

		closedPeriod, _ := aggregate.NewFiscalPeriod(orgID, "2026-08", "August 2026", aggregate.PeriodTypeMonth,
			time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), userID)
		require.NoError(t, repo.Create(ctx, closedPeriod))

		// Close one period
		require.NoError(t, closedPeriod.Close(userID))
		require.NoError(t, repo.Update(ctx, closedPeriod))

		found, err := repo.ListOpen(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, aggregate.PeriodStatusOpen, found[0].Status)
	})
}

func TestFiscalPeriodRepository_GetByOrganizationAndDate(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewFiscalPeriodRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-09", "September 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)
		require.NoError(t, repo.Create(ctx, period))

		// Find period containing a date
		found, err := repo.GetByOrganizationAndDate(ctx, orgID, "2026-09-15")
		require.NoError(t, err)
		assert.Equal(t, period.GetID(), found.GetID())
		assert.Equal(t, "2026-09", found.Code)
	})
}

func TestFiscalPeriod_BusinessRules_CannotLockOpenPeriod(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-10", "October 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)

		// Try to lock without closing
		err := period.Lock(userID)
		assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotClosed)
	})
}

func TestFiscalPeriod_BusinessRules_CannotReopenLockedPeriod(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-11", "November 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)

		// Close and lock
		require.NoError(t, period.Close(userID))
		require.NoError(t, period.Lock(userID))

		// Try to reopen
		err := period.Reopen(userID)
		assert.Error(t, err)
	})
}

func TestFiscalPeriod_BusinessRules_CanPostTransaction(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()
		startDate := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

		period, _ := aggregate.NewFiscalPeriod(orgID, "2026-12", "December 2026", aggregate.PeriodTypeMonth, startDate, endDate, userID)

		// Open period - can post
		transactionDate := time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)
		assert.True(t, period.CanPostTransaction(transactionDate))

		// Set lock date
		lockDate := time.Date(2026, 12, 10, 0, 0, 0, 0, time.UTC)
		require.NoError(t, period.SetLockDate(lockDate))

		// Transaction before lock date - cannot post
		beforeLock := time.Date(2026, 12, 5, 0, 0, 0, 0, time.UTC)
		assert.False(t, period.CanPostTransaction(beforeLock))

		// Transaction after lock date - can post
		afterLock := time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)
		assert.True(t, period.CanPostTransaction(afterLock))

		// Locked period - cannot post
		require.NoError(t, period.Close(userID))
		require.NoError(t, period.Lock(userID))
		assert.False(t, period.CanPostTransaction(transactionDate))
	})
}
