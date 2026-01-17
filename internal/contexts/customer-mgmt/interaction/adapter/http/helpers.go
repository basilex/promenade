package http

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// parseInteractionUUID parses a string into a UUID
func parseInteractionUUID(s string) (uuidv7.UUID, error) {
	return uuidv7.Parse(s)
}

// parseInteractionTime parses a string into a time.Time
func parseInteractionTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// parseOptionalTime parses an optional time string (returns nil if empty)
func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
