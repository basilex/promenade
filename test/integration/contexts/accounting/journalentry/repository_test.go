package journalentry_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestJournalEntryRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		entry, err := aggregate.NewJournalEntry(
			orgID,
			time.Now(),
			"Test Entry",
			aggregate.SourceTypeManual,
			nil,
			userID,
		)
		require.NoError(t, err)

		// Add balanced lines
		accountID1 := uuidv7.New()
		accountID2 := uuidv7.New()
		err = entry.AddLine(accountID1, 10000, 0, "USD", "Debit line")
		require.NoError(t, err)
		err = entry.AddLine(accountID2, 0, 10000, "USD", "Credit line")
		require.NoError(t, err)

		err = repo.Create(ctx, entry)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, entry.GetID())
		require.NoError(t, err)
		assert.Equal(t, entry.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "Test Entry", found.Description)
		assert.Equal(t, aggregate.EntryStatusDraft, found.Status)
		assert.Len(t, found.Lines, 2)
	})
}

func TestJournalEntryRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.Error(t, err)
	})
}

func TestJournalEntryRepository_Update(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		entry, _ := aggregate.NewJournalEntry(
			orgID,
			time.Now(),
			"Old Description",
			aggregate.SourceTypeManual,
			nil,
			userID,
		)

		accountID1 := uuidv7.New()
		accountID2 := uuidv7.New()
		_ = entry.AddLine(accountID1, 5000, 0, "USD", "Debit")
		_ = entry.AddLine(accountID2, 0, 5000, "USD", "Credit")

		// Create the entry first
		require.NoError(t, repo.Create(ctx, entry))

		// Update description
		entry.Description = "New Description"

		err := repo.Update(ctx, entry)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, entry.GetID())
		require.NoError(t, err)
		assert.Equal(t, "New Description", found.Description)
	})
}

func TestJournalEntryRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()
		entry, _ := aggregate.NewJournalEntry(
			orgID,
			time.Now(),
			"To Delete",
			aggregate.SourceTypeManual,
			nil,
			userID,
		)

		accountID1 := uuidv7.New()
		accountID2 := uuidv7.New()
		_ = entry.AddLine(accountID1, 1000, 0, "USD", "Debit")
		_ = entry.AddLine(accountID2, 0, 1000, "USD", "Credit")

		// Create the entry first
		require.NoError(t, repo.Create(ctx, entry))

		err := repo.Delete(ctx, entry.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, entry.GetID())
		assert.Error(t, err, "Soft deleted entry should not be found")
	})
}

func TestJournalEntryRepository_ListByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		accountID1 := uuidv7.New()
		accountID2 := uuidv7.New()

		// Create 2 draft entries
		for i := 0; i < 2; i++ {
			entry, _ := aggregate.NewJournalEntry(
				orgID,
				time.Now(),
				"Draft Entry",
				aggregate.SourceTypeManual,
				nil,
				userID,
			)
			_ = entry.AddLine(accountID1, 1000, 0, "USD", "Debit")
			_ = entry.AddLine(accountID2, 0, 1000, "USD", "Credit")
			require.NoError(t, repo.Create(ctx, entry))
		}

		// Create 1 posted entry
		postedEntry, _ := aggregate.NewJournalEntry(
			orgID,
			time.Now(),
			"Posted Entry",
			aggregate.SourceTypeManual,
			nil,
			userID,
		)
		_ = postedEntry.AddLine(accountID1, 1000, 0, "USD", "Debit")
		_ = postedEntry.AddLine(accountID2, 0, 1000, "USD", "Credit")
		_ = postedEntry.Post(userID)
		require.NoError(t, repo.Create(ctx, postedEntry))

		// List draft entries
		drafts, err := repo.ListByOrganization(ctx, orgID, aggregate.EntryStatusDraft)
		require.NoError(t, err)
		assert.Len(t, drafts, 2)

		// List posted entries
		posted, err := repo.ListByOrganization(ctx, orgID, aggregate.EntryStatusPosted)
		require.NoError(t, err)
		assert.Len(t, posted, 1)
	})
}

func TestJournalEntryRepository_EntryWithMultipleLines(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		entry, _ := aggregate.NewJournalEntry(
			orgID,
			time.Now(),
			"Complex Entry",
			aggregate.SourceTypeManual,
			nil,
			userID,
		)

		// Multiple debit/credit lines (balanced)
		accounts := []uuidv7.UUID{uuidv7.New(), uuidv7.New(), uuidv7.New(), uuidv7.New()}
		_ = entry.AddLine(accounts[0], 5000, 0, "USD", "Debit 1")
		_ = entry.AddLine(accounts[1], 3000, 0, "USD", "Debit 2")
		_ = entry.AddLine(accounts[2], 0, 4000, "USD", "Credit 1")
		_ = entry.AddLine(accounts[3], 0, 4000, "USD", "Credit 2")

		require.NoError(t, repo.Create(ctx, entry))

		found, err := repo.GetByID(ctx, entry.GetID())
		require.NoError(t, err)
		assert.Len(t, found.Lines, 4)

		// Verify totals
		totalDebit := int64(0)
		totalCredit := int64(0)
		for _, line := range found.Lines {
			totalDebit += line.DebitCents
			totalCredit += line.CreditCents
		}
		assert.Equal(t, int64(8000), totalDebit)
		assert.Equal(t, int64(8000), totalCredit)
	})
}

func TestJournalEntryRepository_PostEntry(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		entry, _ := aggregate.NewJournalEntry(
			orgID,
			time.Now(),
			"Entry to Post",
			aggregate.SourceTypeManual,
			nil,
			userID,
		)

		accountID1 := uuidv7.New()
		accountID2 := uuidv7.New()
		_ = entry.AddLine(accountID1, 10000, 0, "USD", "Debit")
		_ = entry.AddLine(accountID2, 0, 10000, "USD", "Credit")

		// Create the entry first
		require.NoError(t, repo.Create(ctx, entry))

		// Post entry
		err := entry.Post(userID)
		require.NoError(t, err)

		err = repo.Update(ctx, entry)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, entry.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.EntryStatusPosted, found.Status)
		assert.NotNil(t, found.PostedBy)
		assert.NotNil(t, found.PostedAt)
		assert.Equal(t, userID, *found.PostedBy)
	})
}

func TestJournalEntryRepository_ReverseEntry(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewJournalEntryRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create and post entry
		entry, _ := aggregate.NewJournalEntry(
			orgID,
			time.Now(),
			"Entry to Reverse",
			aggregate.SourceTypeManual,
			nil,
			userID,
		)

		accountID1 := uuidv7.New()
		accountID2 := uuidv7.New()
		_ = entry.AddLine(accountID1, 5000, 0, "USD", "Debit")
		_ = entry.AddLine(accountID2, 0, 5000, "USD", "Credit")
		_ = entry.Post(userID)
		require.NoError(t, repo.Create(ctx, entry))

		// Reverse entry
		err := entry.Reverse(userID)
		require.NoError(t, err)

		err = repo.Update(ctx, entry)
		require.NoError(t, err)

		// Verify original entry is marked as reversed
		found, err := repo.GetByID(ctx, entry.GetID())
		require.NoError(t, err)
		assert.Equal(t, aggregate.EntryStatusReversed, found.Status)
		assert.NotNil(t, found.ReversedBy)
		assert.NotNil(t, found.ReversedAt)
	})
}
