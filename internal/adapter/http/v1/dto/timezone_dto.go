package dto

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// TimezoneResponse represents timezone response
type TimezoneResponse struct {
	ID           uuidv7.UUID `json:"id"`
	Name         string      `json:"name" example:"Europe/Moscow"`
	Abbreviation string      `json:"abbreviation" example:"MSK"`
	UtcOffset    string      `json:"utc_offset" example:"+03:00"`
	UtcDstOffset string      `json:"utc_dst_offset" example:"+04:00"`
	Description  string      `json:"description" example:"Moscow Standard Time"`
	IsActive     bool        `json:"is_active" example:"true"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// CreateTimezoneRequest represents create timezone request
type CreateTimezoneRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=100" example:"Europe/Moscow"`
	Abbreviation string `json:"abbreviation" binding:"required,min=2,max=10" example:"MSK"`
	UtcOffset    string `json:"utc_offset" binding:"required,timezone_offset" example:"+03:00"`
	UtcDstOffset string `json:"utc_dst_offset" binding:"omitempty,timezone_offset" example:"+04:00"`
	Description  string `json:"description" binding:"omitempty,max=255" example:"Moscow Standard Time"`
	IsActive     bool   `json:"is_active" example:"true"`
}

// UpdateTimezoneRequest represents update timezone request
type UpdateTimezoneRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=100" example:"Europe/Moscow"`
	Abbreviation string `json:"abbreviation" binding:"required,min=2,max=10" example:"MSK"`
	UtcOffset    string `json:"utc_offset" binding:"required,timezone_offset" example:"+03:00"`
	UtcDstOffset string `json:"utc_dst_offset" binding:"omitempty,timezone_offset" example:"+04:00"`
	Description  string `json:"description" binding:"omitempty,max=255" example:"Moscow Standard Time"`
	IsActive     bool   `json:"is_active" example:"true"`
}

// TimezoneListResponse represents paginated timezones response
type TimezoneListResponse struct {
	Timezones []*TimezoneResponse `json:"timezones"`
	Total     int                 `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
}
