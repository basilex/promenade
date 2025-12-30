package timezone

import "errors"

// Domain errors for timezone aggregate
var (
	// ErrTimezoneNotFound is returned when a timezone cannot be found
	ErrTimezoneNotFound = errors.New("timezone not found")

	// ErrInvalidTimezoneName is returned when timezone name validation fails
	ErrInvalidTimezoneName = errors.New("invalid timezone name")

	// ErrInvalidUTCOffset is returned when UTC offset validation fails
	ErrInvalidUTCOffset = errors.New("invalid UTC offset")
)
