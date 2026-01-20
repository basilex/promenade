package fiscalperiod

import "errors"

var (
// Validation errors
ErrPeriodCodeEmpty   = errors.New("period code cannot be empty")
ErrPeriodNameEmpty   = errors.New("period name cannot be empty")
ErrInvalidPeriodType = errors.New("invalid period type")
ErrInvalidDateRange  = errors.New("end date must be after start date")

// Business logic errors
ErrPeriodAlreadyClosed = errors.New("period is already closed")
ErrPeriodAlreadyLocked = errors.New("period is already locked")
ErrPeriodNotClosed     = errors.New("period must be closed before locking")
ErrCannotReopenLocked  = errors.New("cannot reopen locked period")
ErrPeriodNotFound      = errors.New("fiscal period not found")
ErrPeriodOverlap       = errors.New("period dates overlap with existing period")
)
