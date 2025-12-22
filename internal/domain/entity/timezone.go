package entity

import (
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Timezone represents a timezone entity
type Timezone struct {
	ID           uuidv7.UUID `db:"id" json:"id"`
	Name         string      `db:"name" json:"name" validate:"required,min=3,max=100"`                    // e.g., "Europe/Moscow"
	Abbreviation string      `db:"abbreviation" json:"abbreviation" validate:"required,min=2,max=10"`     // e.g., "MSK"
	UtcOffset    string      `db:"utc_offset" json:"utc_offset" validate:"required,timezone_offset"`     // e.g., "+03:00"
	UtcDstOffset string      `db:"utc_dst_offset" json:"utc_dst_offset" validate:"omitempty,timezone_offset"` // e.g., "+04:00"
	Description  string      `db:"description" json:"description" validate:"omitempty,max=255"`
	IsActive     bool        `db:"is_active" json:"is_active"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `db:"updated_at" json:"updated_at"`
}

// Validate validates timezone fields
func (t *Timezone) Validate() error {
	if t.Name == "" {
		return errors.New("timezone name is required")
	}
	if t.Abbreviation == "" {
		return errors.New("timezone abbreviation is required")
	}
	if t.UtcOffset == "" {
		return errors.New("UTC offset is required")
	}
	return nil
}

// NewTimezone creates a new timezone with generated UUID v7
func NewTimezone(name, abbreviation, utcOffset, utcDstOffset, description string) *Timezone {
	now := time.Now()
	return &Timezone{
		ID:           uuidv7.New(),
		Name:         name,
		Abbreviation: abbreviation,
		UtcOffset:    utcOffset,
		UtcDstOffset: utcDstOffset,
		Description:  description,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// GetUTCOffsetMinutes returns the UTC offset in minutes
func (t *Timezone) GetUTCOffsetMinutes() int {
	// Parse "+03:00" or "-05:00" format
	var hours, minutes int
	_, err := time.Parse("-07:00", t.UtcOffset)
	if err == nil {
		duration, _ := time.ParseDuration(t.UtcOffset + "h")
		return int(duration.Minutes())
	}
	return hours*60 + minutes
}
