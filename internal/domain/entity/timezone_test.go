package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestTimezone_Validate(t *testing.T) {
	tests := []struct {
		name     string
		timezone *Timezone
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid timezone",
			timezone: &Timezone{
				ID:           uuidv7.New(),
				Name:         "Europe/Moscow",
				Abbreviation: "MSK",
				UtcOffset:    "+03:00",
				UtcDstOffset: "+04:00",
				Description:  "Moscow Standard Time",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			timezone: &Timezone{
				ID:           uuidv7.New(),
				Abbreviation: "MSK",
				UtcOffset:    "+03:00",
			},
			wantErr: true,
			errMsg:  "timezone name is required",
		},
		{
			name: "missing abbreviation",
			timezone: &Timezone{
				ID:        uuidv7.New(),
				Name:      "Europe/Moscow",
				UtcOffset: "+03:00",
			},
			wantErr: true,
			errMsg:  "timezone abbreviation is required",
		},
		{
			name: "missing UTC offset",
			timezone: &Timezone{
				ID:           uuidv7.New(),
				Name:         "Europe/Moscow",
				Abbreviation: "MSK",
			},
			wantErr: true,
			errMsg:  "UTC offset is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.timezone.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewTimezone(t *testing.T) {
	name := "Europe/Moscow"
	abbreviation := "MSK"
	utcOffset := "+03:00"
	utcDstOffset := "+04:00"
	description := "Moscow Standard Time"

	timezone := NewTimezone(name, abbreviation, utcOffset, utcDstOffset, description)

	assert.NotEmpty(t, timezone.ID)
	assert.Equal(t, name, timezone.Name)
	assert.Equal(t, abbreviation, timezone.Abbreviation)
	assert.Equal(t, utcOffset, timezone.UtcOffset)
	assert.Equal(t, utcDstOffset, timezone.UtcDstOffset)
	assert.Equal(t, description, timezone.Description)
	assert.True(t, timezone.IsActive)
	assert.NotZero(t, timezone.CreatedAt)
	assert.NotZero(t, timezone.UpdatedAt)
}

func TestTimezone_CommonTimezones(t *testing.T) {
	tests := []struct {
		name         string
		timezone     *Timezone
		wantOffset   string
		wantDSTValid bool
	}{
		{
			name:         "UTC",
			timezone:     NewTimezone("UTC", "UTC", "+00:00", "", "Coordinated Universal Time"),
			wantOffset:   "+00:00",
			wantDSTValid: false,
		},
		{
			name:         "New York with DST",
			timezone:     NewTimezone("America/New_York", "EST", "-05:00", "-04:00", "Eastern Standard Time"),
			wantOffset:   "-05:00",
			wantDSTValid: true,
		},
		{
			name:         "Tokyo (no DST)",
			timezone:     NewTimezone("Asia/Tokyo", "JST", "+09:00", "", "Japan Standard Time"),
			wantOffset:   "+09:00",
			wantDSTValid: false,
		},
		{
			name:         "London with DST",
			timezone:     NewTimezone("Europe/London", "GMT", "+00:00", "+01:00", "Greenwich Mean Time"),
			wantOffset:   "+00:00",
			wantDSTValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantOffset, tt.timezone.UtcOffset)
			if tt.wantDSTValid {
				assert.NotEmpty(t, tt.timezone.UtcDstOffset)
			} else {
				assert.Empty(t, tt.timezone.UtcDstOffset)
			}
			assert.NoError(t, tt.timezone.Validate())
		})
	}
}
