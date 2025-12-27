package dto

import "github.com/basilex/promenade/pkg/reference"

// TimezoneResponse represents a timezone in API responses
type TimezoneResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	UTCOffset    int    `json:"utc_offset"`
	UTCFormatted string `json:"utc_formatted"` // e.g., "UTC+02:00"
	IsActive     bool   `json:"is_active"`
}

// ToTimezoneResponse converts reference.Timezone to TimezoneResponse
func ToTimezoneResponse(timezone *reference.Timezone) TimezoneResponse {
	return TimezoneResponse{
		ID:           timezone.ID.String(),
		Name:         timezone.Name,
		Abbreviation: timezone.Abbreviation,
		UTCOffset:    timezone.UTCOffset,
		UTCFormatted: timezone.FormatOffset(),
		IsActive:     timezone.IsActive,
	}
}

// ToTimezoneResponses converts slice of timezones to response DTOs
func ToTimezoneResponses(timezones []reference.Timezone) []TimezoneResponse {
	responses := make([]TimezoneResponse, len(timezones))
	for i, timezone := range timezones {
		responses[i] = ToTimezoneResponse(&timezone)
	}
	return responses
}
