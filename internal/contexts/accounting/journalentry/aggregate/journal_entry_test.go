package aggregate

import (
    "testing"
    "time"

    "github.com/basilex/promenade/internal/contexts/accounting/journalentry"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewJournalEntry(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    sourceID := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    tests := []struct {
        name        string
        description string
        entryDate   time.Time
        sourceType  SourceType
        sourceID    *uuidv7.UUID
        expectError error
    }{
        {
            name:        "valid manual entry",
            description: "Manual journal entry",
            entryDate:   entryDate,
            sourceType:  SourceTypeManual,
            sourceID:    nil,
            expectError: nil,
        },
        {
            name:        "valid invoice entry",
            description: "Invoice #123",
            entryDate:   entryDate,
            sourceType:  SourceTypeInvoice,
            sourceID:    &sourceID,
            expectError: nil,
        },
        {
            name:        "valid payment entry",
            description: "Payment received",
            entryDate:   entryDate,
            sourceType:  SourceTypePayment,
            sourceID:    &sourceID,
            expectError: nil,
        },
        {
            name:        "empty description",
            description: "",
            entryDate:   entryDate,
            sourceType:  SourceTypeManual,
            sourceID:    nil,
            expectError: journalentry.ErrJournalEntryDescriptionEmpty,
        },
        {
            name:        "zero entry date",
            description: "Test entry",
            entryDate:   time.Time{},
            sourceType:  SourceTypeManual,
            sourceID:    nil,
            expectError: journalentry.ErrJournalEntryInvalidDate,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            entry, err := NewJournalEntry(orgID, tt.entryDate, tt.description, tt.sourceType, tt.sourceID, userID)

            if tt.expectError != nil {
                assert.ErrorIs(t, err, tt.expectError)
                assert.Nil(t, entry)
            } else {
                require.NoError(t, err)
                require.NotNil(t, entry)
                assert.Equal(t, orgID, entry.OrganizationID)
                assert.Equal(t, tt.description, entry.Description)
                assert.Equal(t, tt.entryDate, entry.EntryDate)
                assert.Equal(t, tt.sourceType, entry.SourceType)
                assert.Equal(t, tt.sourceID, entry.SourceID)
                assert.Equal(t, EntryStatusDraft, entry.Status)
                assert.Empty(t, entry.Lines)
                assert.Nil(t, entry.PostedBy)
                assert.Nil(t, entry.PostedAt)
            }
        })
    }
}

func TestJournalEntry_AddLine(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID1 := uuidv7.New()
    accountID2 := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    entry, err := NewJournalEntry(orgID, entryDate, "Test Entry", SourceTypeManual, nil, userID)
    require.NoError(t, err)

    t.Run("add debit line", func(t *testing.T) {
        err := entry.AddLine(accountID1, 100000, 0, "UAH", "Debit line")
        require.NoError(t, err)
        assert.Len(t, entry.Lines, 1)
        assert.Equal(t, accountID1, entry.Lines[0].AccountID)
        assert.Equal(t, int64(100000), entry.Lines[0].DebitCents)
        assert.Equal(t, int64(0), entry.Lines[0].CreditCents)
        assert.Equal(t, "Debit line", entry.Lines[0].Description)
        assert.Equal(t, 1, entry.Lines[0].LineOrder)
    })

    t.Run("add credit line", func(t *testing.T) {
        err := entry.AddLine(accountID2, 0, 100000, "UAH", "Credit line")
        require.NoError(t, err)
        assert.Len(t, entry.Lines, 2)
        assert.Equal(t, int64(0), entry.Lines[1].DebitCents)
        assert.Equal(t, int64(100000), entry.Lines[1].CreditCents)
        assert.Equal(t, 2, entry.Lines[1].LineOrder)
    })

    t.Run("nil account", func(t *testing.T) {
        err := entry.AddLine(uuidv7.Nil, 10000, 0, "UAH", "Invalid")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryLineInvalidAccount)
    })

    t.Run("negative debit", func(t *testing.T) {
        err := entry.AddLine(accountID1, -10000, 0, "UAH", "Negative")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryLineInvalidAmount)
    })

    t.Run("negative credit", func(t *testing.T) {
        err := entry.AddLine(accountID1, 0, -10000, "UAH", "Negative")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryLineInvalidAmount)
    })

    t.Run("both debit and credit", func(t *testing.T) {
        err := entry.AddLine(accountID1, 10000, 10000, "UAH", "Both sides")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryLineBothSides)
    })

    t.Run("zero amounts", func(t *testing.T) {
        err := entry.AddLine(accountID1, 0, 0, "UAH", "Zero")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryLineInvalidAmount)
    })

    t.Run("default currency", func(t *testing.T) {
        entry2, err := NewJournalEntry(orgID, entryDate, "Test", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry2.AddLine(accountID1, 10000, 0, "", "No currency")
        require.NoError(t, err)
        assert.Equal(t, "UAH", entry2.Lines[0].CurrencyCode)
    })

    t.Run("cannot add line to posted entry", func(t *testing.T) {
        postedEntry, err := NewJournalEntry(orgID, entryDate, "Posted", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = postedEntry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = postedEntry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)
        err = postedEntry.Post(userID)
        require.NoError(t, err)

        err = postedEntry.AddLine(accountID1, 5000, 0, "UAH", "New line")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryAlreadyPosted)
    })
}

func TestJournalEntry_RemoveLine(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID1 := uuidv7.New()
    accountID2 := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    entry, err := NewJournalEntry(orgID, entryDate, "Test Entry", SourceTypeManual, nil, userID)
    require.NoError(t, err)
    err = entry.AddLine(accountID1, 10000, 0, "UAH", "Line 1")
    require.NoError(t, err)
    err = entry.AddLine(accountID2, 0, 10000, "UAH", "Line 2")
    require.NoError(t, err)
    lineID := entry.Lines[0].ID

    t.Run("remove line", func(t *testing.T) {
        err := entry.RemoveLine(lineID)
        require.NoError(t, err)
        assert.Len(t, entry.Lines, 1)
        assert.Equal(t, 1, entry.Lines[0].LineOrder)
    })

    t.Run("remove non-existent line", func(t *testing.T) {
        err := entry.RemoveLine(uuidv7.New())
        require.NoError(t, err) // No error for non-existent line
    })

    t.Run("cannot remove from posted entry", func(t *testing.T) {
        postedEntry, err := NewJournalEntry(orgID, entryDate, "Posted", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = postedEntry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = postedEntry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)
        removeLineID := postedEntry.Lines[0].ID
        err = postedEntry.Post(userID)
        require.NoError(t, err)

        err = postedEntry.RemoveLine(removeLineID)
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryAlreadyPosted)
    })
}

func TestJournalEntry_UpdateDescription(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    entry, err := NewJournalEntry(orgID, entryDate, "Original Description", SourceTypeManual, nil, userID)
    require.NoError(t, err)

    t.Run("update description", func(t *testing.T) {
        err := entry.UpdateDescription("New Description")
        require.NoError(t, err)
        assert.Equal(t, "New Description", entry.Description)
    })

    t.Run("empty description", func(t *testing.T) {
        err := entry.UpdateDescription("")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryDescriptionEmpty)
    })

    t.Run("cannot update posted entry", func(t *testing.T) {
        accountID1 := uuidv7.New()
        accountID2 := uuidv7.New()
        postedEntry, err := NewJournalEntry(orgID, entryDate, "Posted", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = postedEntry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = postedEntry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)
        err = postedEntry.Post(userID)
        require.NoError(t, err)

        err = postedEntry.UpdateDescription("Cannot change")
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryAlreadyPosted)
    })
}

func TestJournalEntry_Validate(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID1 := uuidv7.New()
    accountID2 := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    t.Run("no lines", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Empty Entry", SourceTypeManual, nil, userID)
        require.NoError(t, err)

        err = entry.Validate()
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryNoLines)
    })

    t.Run("unbalanced entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Unbalanced", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 5000, "UAH", "Credit")
        require.NoError(t, err)

        err = entry.Validate()
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryUnbalanced)
    })

    t.Run("balanced entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Balanced", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)

        err = entry.Validate()
        require.NoError(t, err)
    })

    t.Run("balanced with multiple lines", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Multiple Lines", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit 1")
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 5000, 0, "UAH", "Debit 2")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 7000, "UAH", "Credit 1")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 8000, "UAH", "Credit 2")
        require.NoError(t, err)

        err = entry.Validate()
        require.NoError(t, err)
        assert.Equal(t, int64(15000), entry.GetTotalDebit())
        assert.Equal(t, int64(15000), entry.GetTotalCredit())
    })
}

func TestJournalEntry_Post(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID1 := uuidv7.New()
    accountID2 := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    t.Run("post valid entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Valid Entry", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)

        err = entry.Post(userID)
        require.NoError(t, err)
        assert.Equal(t, EntryStatusPosted, entry.Status)
        assert.NotNil(t, entry.PostedBy)
        assert.Equal(t, userID, *entry.PostedBy)
        assert.NotNil(t, entry.PostedAt)
    })

    t.Run("cannot post invalid entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Invalid", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 5000, "UAH", "Credit")
        require.NoError(t, err)

        err = entry.Post(userID)
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryCannotPostDraft)
    })

    t.Run("cannot post already posted", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Posted", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)
        err = entry.Post(userID)
        require.NoError(t, err)

        err = entry.Post(userID)
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryAlreadyPosted)
    })
}

func TestJournalEntry_Reverse(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID1 := uuidv7.New()
    accountID2 := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    t.Run("reverse posted entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "To Reverse", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)
        err = entry.Post(userID)
        require.NoError(t, err)

        err = entry.Reverse(userID)
        require.NoError(t, err)
        assert.Equal(t, EntryStatusReversed, entry.Status)
        assert.NotNil(t, entry.ReversedBy)
        assert.Equal(t, userID, *entry.ReversedBy)
        assert.NotNil(t, entry.ReversedAt)
    })

    t.Run("cannot reverse draft entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Draft", SourceTypeManual, nil, userID)
        require.NoError(t, err)

        err = entry.Reverse(userID)
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryNotPosted)
    })

    t.Run("cannot reverse already reversed", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Reversed", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)
        err = entry.Post(userID)
        require.NoError(t, err)
        err = entry.Reverse(userID)
        require.NoError(t, err)

        err = entry.Reverse(userID)
        assert.ErrorIs(t, err, journalentry.ErrJournalEntryNotPosted)
    })
}

func TestJournalEntry_IsBalanced(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID1 := uuidv7.New()
    accountID2 := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    t.Run("balanced entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Balanced", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 10000, "UAH", "Credit")
        require.NoError(t, err)

        assert.True(t, entry.IsBalanced())
    })

    t.Run("unbalanced entry", func(t *testing.T) {
        entry, err := NewJournalEntry(orgID, entryDate, "Unbalanced", SourceTypeManual, nil, userID)
        require.NoError(t, err)
        err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit")
        require.NoError(t, err)
        err = entry.AddLine(accountID2, 0, 5000, "UAH", "Credit")
        require.NoError(t, err)

        assert.False(t, entry.IsBalanced())
    })
}

func TestJournalEntry_GetTotals(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID1 := uuidv7.New()
    accountID2 := uuidv7.New()
    accountID3 := uuidv7.New()
    entryDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    entry, err := NewJournalEntry(orgID, entryDate, "Test Totals", SourceTypeManual, nil, userID)
    require.NoError(t, err)
    err = entry.AddLine(accountID1, 10000, 0, "UAH", "Debit 1")
    require.NoError(t, err)
    err = entry.AddLine(accountID2, 5000, 0, "UAH", "Debit 2")
    require.NoError(t, err)
    err = entry.AddLine(accountID3, 0, 15000, "UAH", "Credit")
    require.NoError(t, err)

    assert.Equal(t, int64(15000), entry.GetTotalDebit())
    assert.Equal(t, int64(15000), entry.GetTotalCredit())
    assert.True(t, entry.IsBalanced())
}