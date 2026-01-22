package aggregate

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type PeriodType string

const (
	PeriodTypeMonth   PeriodType = "month"
	PeriodTypeQuarter PeriodType = "quarter"
	PeriodTypeYear    PeriodType = "year"
)

type PeriodStatus string

const (
	PeriodStatusOpen   PeriodStatus = "open"
	PeriodStatusClosed PeriodStatus = "closed"
	PeriodStatusLocked PeriodStatus = "locked"
)

type FiscalPeriod struct {
	aggregate.BaseAggregate

	OrganizationID uuidv7.UUID
	Code           string
	Name           string
	PeriodType     PeriodType
	StartDate      time.Time
	EndDate        time.Time
	Status         PeriodStatus
	LockDate       *time.Time

	ClosedBy      *uuidv7.UUID
	ClosedAt      *time.Time
	ReopenedBy    *uuidv7.UUID
	ReopenedAt    *time.Time
	LockedBy      *uuidv7.UUID
	LockedAt      *time.Time
	LastUpdatedBy uuidv7.UUID
}

func NewFiscalPeriod(organizationID uuidv7.UUID, code, name string, periodType PeriodType, startDate, endDate time.Time, createdBy uuidv7.UUID) (*FiscalPeriod, error) {
	if code == "" {
		return nil, fiscalperiod.ErrPeriodCodeEmpty
	}
	if name == "" {
		return nil, fiscalperiod.ErrPeriodNameEmpty
	}
	if !isValidPeriodType(periodType) {
		return nil, fiscalperiod.ErrInvalidPeriodType
	}
	if startDate.After(endDate) || startDate.Equal(endDate) {
		return nil, fiscalperiod.ErrInvalidDateRange
	}

	return &FiscalPeriod{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		OrganizationID: organizationID,
		Code:           code,
		Name:           name,
		PeriodType:     periodType,
		StartDate:      startDate,
		EndDate:        endDate,
		Status:         PeriodStatusOpen,
		LastUpdatedBy:  createdBy,
	}, nil
}

func (p *FiscalPeriod) Close(userID uuidv7.UUID) error {
	if p.Status == PeriodStatusClosed {
		return fiscalperiod.ErrPeriodAlreadyClosed
	}
	if p.Status == PeriodStatusLocked {
		return fiscalperiod.ErrPeriodAlreadyLocked
	}

	now := time.Now()
	p.Status = PeriodStatusClosed
	p.ClosedBy = &userID
	p.ClosedAt = &now
	p.Touch()
	return nil
}

func (p *FiscalPeriod) Reopen(userID uuidv7.UUID) error {
	if p.Status != PeriodStatusClosed {
		return fiscalperiod.ErrPeriodNotClosed
	}

	now := time.Now()
	p.Status = PeriodStatusOpen
	p.ReopenedBy = &userID
	p.ReopenedAt = &now
	p.Touch()
	return nil
}

func (p *FiscalPeriod) Lock(userID uuidv7.UUID) error {
	if p.Status != PeriodStatusClosed {
		return fiscalperiod.ErrPeriodNotClosed
	}
	if p.Status == PeriodStatusLocked {
		return fiscalperiod.ErrPeriodAlreadyLocked
	}

	now := time.Now()
	p.Status = PeriodStatusLocked
	p.LockedBy = &userID
	p.LockedAt = &now
	p.Touch()
	return nil
}

func (p *FiscalPeriod) SetLockDate(lockDate time.Time) error {
	if p.Status == PeriodStatusLocked {
		return fiscalperiod.ErrPeriodAlreadyLocked
	}

	p.LockDate = &lockDate
	p.Touch()
	return nil
}

func (p *FiscalPeriod) CanPostTransaction(transactionDate time.Time) bool {
	if p.Status == PeriodStatusLocked {
		return false
	}
	if p.LockDate != nil && transactionDate.Before(*p.LockDate) {
		return false
	}
	return true
}

func (p *FiscalPeriod) Contains(date time.Time) bool {
	return (date.Equal(p.StartDate) || date.After(p.StartDate)) &&
		(date.Equal(p.EndDate) || date.Before(p.EndDate))
}

func isValidPeriodType(pt PeriodType) bool {
	return pt == PeriodTypeMonth || pt == PeriodTypeQuarter || pt == PeriodTypeYear
}
