package aggregate

import (
    "testing"
    "time"

    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewFiscalPeriod(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    tests := []struct {
        name        string
        code        string
        periodName  string
        periodType  PeriodType
        startDate   time.Time
        endDate     time.Time
        expectError error
    }{
        {
            name:        "valid monthly period",
            code:        "2024-01",
            periodName:  "January 2024",
            periodType:  PeriodTypeMonth,
            startDate:   startDate,
            endDate:     endDate,
            expectError: nil,
        },
        {
            name:        "valid quarterly period",
            code:        "2024-Q1",
            periodName:  "Q1 2024",
            periodType:  PeriodTypeQuarter,
            startDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
            endDate:     time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
            expectError: nil,
        },
        {
            name:        "valid yearly period",
            code:        "2024",
            periodName:  "Fiscal Year 2024",
            periodType:  PeriodTypeYear,
            startDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
            endDate:     time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
            expectError: nil,
        },
        {
            name:        "empty code",
            code:        "",
            periodName:  "Test Period",
            periodType:  PeriodTypeMonth,
            startDate:   startDate,
            endDate:     endDate,
            expectError: fiscalperiod.ErrPeriodCodeEmpty,
        },
        {
            name:        "empty name",
            code:        "2024-01",
            periodName:  "",
            periodType:  PeriodTypeMonth,
            startDate:   startDate,
            endDate:     endDate,
            expectError: fiscalperiod.ErrPeriodNameEmpty,
        },
        {
            name:        "invalid period type",
            code:        "2024-01",
            periodName:  "Test Period",
            periodType:  PeriodType("invalid"),
            startDate:   startDate,
            endDate:     endDate,
            expectError: fiscalperiod.ErrInvalidPeriodType,
        },
        {
            name:        "end date before start date",
            code:        "2024-01",
            periodName:  "Invalid Period",
            periodType:  PeriodTypeMonth,
            startDate:   endDate,
            endDate:     startDate,
            expectError: fiscalperiod.ErrInvalidDateRange,
        },
        {
            name:        "start date equals end date",
            code:        "2024-01",
            periodName:  "Same Date Period",
            periodType:  PeriodTypeMonth,
            startDate:   startDate,
            endDate:     startDate,
            expectError: fiscalperiod.ErrInvalidDateRange,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p, err := NewFiscalPeriod(orgID, tt.code, tt.periodName, tt.periodType, tt.startDate, tt.endDate, userID)

            if tt.expectError != nil {
                assert.ErrorIs(t, err, tt.expectError)
                assert.Nil(t, p)
            } else {
                require.NoError(t, err)
                require.NotNil(t, p)
                assert.Equal(t, orgID, p.OrganizationID)
                assert.Equal(t, tt.code, p.Code)
                assert.Equal(t, tt.periodName, p.Name)
                assert.Equal(t, tt.periodType, p.PeriodType)
                assert.Equal(t, tt.startDate, p.StartDate)
                assert.Equal(t, tt.endDate, p.EndDate)
                assert.Equal(t, PeriodStatusOpen, p.Status)
                assert.Nil(t, p.ClosedBy)
                assert.Nil(t, p.ClosedAt)
                assert.Nil(t, p.LockDate)
            }
        })
    }
}

func TestFiscalPeriod_Close(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    t.Run("close open period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)

        err = p.Close(userID)
        require.NoError(t, err)
        assert.Equal(t, PeriodStatusClosed, p.Status)
        assert.NotNil(t, p.ClosedBy)
        assert.Equal(t, userID, *p.ClosedBy)
        assert.NotNil(t, p.ClosedAt)
    })

    t.Run("close already closed period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)

        err = p.Close(userID)
        assert.ErrorIs(t, err, fiscalperiod.ErrPeriodAlreadyClosed)
    })

    t.Run("cannot close locked period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)
        err = p.Lock(userID)
        require.NoError(t, err)

        err = p.Close(userID)
        assert.ErrorIs(t, err, fiscalperiod.ErrPeriodAlreadyLocked)
    })
}

func TestFiscalPeriod_Reopen(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    t.Run("reopen closed period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)

        err = p.Reopen(userID)
        require.NoError(t, err)
        assert.Equal(t, PeriodStatusOpen, p.Status)
        assert.NotNil(t, p.ReopenedBy)
        assert.Equal(t, userID, *p.ReopenedBy)
        assert.NotNil(t, p.ReopenedAt)
    })

    t.Run("cannot reopen open period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)

        err = p.Reopen(userID)
        assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotClosed)
    })

    t.Run("cannot reopen locked period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)
        err = p.Lock(userID)
        require.NoError(t, err)

        err = p.Reopen(userID)
        assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotClosed)
    })
}

func TestFiscalPeriod_Lock(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    t.Run("lock closed period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)

        err = p.Lock(userID)
        require.NoError(t, err)
        assert.Equal(t, PeriodStatusLocked, p.Status)
        assert.NotNil(t, p.LockedBy)
        assert.Equal(t, userID, *p.LockedBy)
        assert.NotNil(t, p.LockedAt)
    })

    t.Run("cannot lock open period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)

        err = p.Lock(userID)
        assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotClosed)
    })

    t.Run("cannot lock already locked period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)
        err = p.Lock(userID)
        require.NoError(t, err)

        err = p.Lock(userID)
        assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotClosed)
    })
}

func TestFiscalPeriod_SetLockDate(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
    lockDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

    t.Run("set lock date on open period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)

        err = p.SetLockDate(lockDate)
        require.NoError(t, err)
        assert.NotNil(t, p.LockDate)
        assert.Equal(t, lockDate, *p.LockDate)
    })

    t.Run("cannot set lock date on locked period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)
        err = p.Lock(userID)
        require.NoError(t, err)

        err = p.SetLockDate(lockDate)
        assert.ErrorIs(t, err, fiscalperiod.ErrPeriodAlreadyLocked)
    })
}

func TestFiscalPeriod_CanPostTransaction(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
    lockDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
    transactionDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

    t.Run("can post to open period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)

        assert.True(t, p.CanPostTransaction(transactionDate))
    })

    t.Run("cannot post to locked period", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.Close(userID)
        require.NoError(t, err)
        err = p.Lock(userID)
        require.NoError(t, err)

        assert.False(t, p.CanPostTransaction(transactionDate))
    })

    t.Run("cannot post transaction before lock date", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.SetLockDate(lockDate)
        require.NoError(t, err)

        beforeLockDate := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
        assert.False(t, p.CanPostTransaction(beforeLockDate))
    })

    t.Run("can post transaction after lock date", func(t *testing.T) {
        p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
        require.NoError(t, err)
        err = p.SetLockDate(lockDate)
        require.NoError(t, err)

        afterLockDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)
        assert.True(t, p.CanPostTransaction(afterLockDate))
    })
}

func TestFiscalPeriod_Contains(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
    require.NoError(t, err)

    tests := []struct {
        name     string
        date     time.Time
        expected bool
    }{
        {
            name:     "date equals start date",
            date:     startDate,
            expected: true,
        },
        {
            name:     "date equals end date",
            date:     endDate,
            expected: true,
        },
        {
            name:     "date within period",
            date:     time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
            expected: true,
        },
        {
            name:     "date before period",
            date:     time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
            expected: false,
        },
        {
            name:     "date after period",
            date:     time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := p.Contains(tt.date)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestFiscalPeriod_PeriodLifecycle(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

    p, err := NewFiscalPeriod(orgID, "2024-01", "January 2024", PeriodTypeMonth, startDate, endDate, userID)
    require.NoError(t, err)

    // Initial state: Open
    assert.Equal(t, PeriodStatusOpen, p.Status)

    // Close period
    err = p.Close(userID)
    require.NoError(t, err)
    assert.Equal(t, PeriodStatusClosed, p.Status)

    // Reopen period
    err = p.Reopen(userID)
    require.NoError(t, err)
    assert.Equal(t, PeriodStatusOpen, p.Status)

    // Close again
    err = p.Close(userID)
    require.NoError(t, err)
    assert.Equal(t, PeriodStatusClosed, p.Status)

    // Lock period
    err = p.Lock(userID)
    require.NoError(t, err)
    assert.Equal(t, PeriodStatusLocked, p.Status)

    // Cannot reopen locked period
    err = p.Reopen(userID)
    assert.ErrorIs(t, err, fiscalperiod.ErrPeriodNotClosed)
}