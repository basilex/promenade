package http

import (
	"fmt"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
)

// TimezoneResponse represents a timezone in API responses
type TimezoneResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	UTCOffset    string `json:"utc_offset"`
	IsActive     bool   `json:"is_active"`
}

// ToTimezoneResponse converts timezone entity to TimezoneResponse
func ToTimezoneResponse(t *timezone.Timezone) TimezoneResponse {
	// Convert UTC offset from seconds to hours:minutes format
	hours := t.UTCOffset / 3600
	minutes := (t.UTCOffset % 3600) / 60
	offsetStr := fmt.Sprintf("%+03d:%02d", hours, minutes)
	
	return TimezoneResponse{
		ID:           t.ID.String(),
		Name:         t.Name,
		Abbreviation: t.Abbreviation,
		UTCOffset:    offsetStr,
		IsActive:     t.IsActive,
	}
}

// ToTimezoneResponses converts slice of timezones to response DTOs
func ToTimezoneResponses(timezones []*timezone.Timezone) []TimezoneResponse {
	responses := make([]TimezoneResponse, len(timezones))
	for i, t := range timezones {
		responses[i] = ToTimezoneResponse(t)
	}
	return responses
}

// CreateTimezoneRequest
type CreateTimezoneRequest struct {
	Name string `json:"name" binding:"required"`
	Abbreviation string `json:"abbreviation" binding:"required"`
	UTCOffset string `json:"utc_offset" binding:"required"`
}

// UpdateTimezoneRequest
type UpdateTimezoneRequest struct {
	Abbreviation string `json:"abbreviation" binding:"required"`
	UTCOffset string `json:"utc_offset" binding:"required"`
	IsActive bool `json:"is_active"`
}
