package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestNewUserPreference(t *testing.T) {
	userID := uuidv7.New()
	pref := NewUserPreference(userID)

	assert.NotEqual(t, uuidv7.Nil, pref.ID)
	assert.Equal(t, userID, pref.UserID)
	assert.True(t, pref.EmailEnabled)
	assert.False(t, pref.SMSEnabled)
	assert.True(t, pref.PushEnabled)
	assert.True(t, pref.InAppEnabled)
	assert.True(t, pref.SystemEnabled)
	assert.True(t, pref.SecurityEnabled)
	assert.False(t, pref.MarketingEnabled)
	assert.True(t, pref.ProductEnabled)
	assert.True(t, pref.SocialEnabled)
	assert.Equal(t, "UTC", pref.Timezone)
	assert.NotZero(t, pref.CreatedAt)
	assert.NotZero(t, pref.UpdatedAt)
}

func TestUserPreference_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pref    *UserPreference
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid preference",
			pref: &UserPreference{
				UserID:   uuidv7.New(),
				Timezone: "UTC",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			pref: &UserPreference{
				UserID:   uuidv7.Nil,
				Timezone: "UTC",
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "missing timezone",
			pref: &UserPreference{
				UserID: uuidv7.New(),
			},
			wantErr: true,
			errMsg:  "timezone is required",
		},
		{
			name: "quiet_hours_start without end",
			pref: func() *UserPreference {
				start, _ := ParseTimeOfDay("22:00")
				return &UserPreference{
					UserID:          uuidv7.New(),
					Timezone:        "UTC",
					QuietHoursStart: start,
				}
			}(),
			wantErr: true,
			errMsg:  "quiet_hours_end is required",
		},
		{
			name: "quiet_hours_end without start",
			pref: func() *UserPreference {
				end, _ := ParseTimeOfDay("08:00")
				return &UserPreference{
					UserID:        uuidv7.New(),
					Timezone:      "UTC",
					QuietHoursEnd: end,
				}
			}(),
			wantErr: true,
			errMsg:  "quiet_hours_start is required",
		},
		{
			name: "valid quiet hours",
			pref: func() *UserPreference {
				start, _ := ParseTimeOfDay("22:00")
				end, _ := ParseTimeOfDay("08:00")
				return &UserPreference{
					UserID:          uuidv7.New(),
					Timezone:        "UTC",
					QuietHoursStart: start,
					QuietHoursEnd:   end,
				}
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pref.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserPreference_IsChannelEnabled(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())

	tests := []struct {
		channel NotificationChannel
		enabled bool
	}{
		{NotificationChannelEmail, true},
		{NotificationChannelSMS, false},
		{NotificationChannelPush, true},
		{NotificationChannelInApp, true},
		{NotificationChannel("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			result := pref.IsChannelEnabled(tt.channel)
			assert.Equal(t, tt.enabled, result)
		})
	}
}

func TestUserPreference_IsTypeEnabled(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())

	tests := []struct {
		notifType NotificationType
		enabled   bool
	}{
		{NotificationTypeSystem, true},
		{NotificationTypeSecurity, true},
		{NotificationTypeMarketing, false},
		{NotificationTypeProduct, true},
		{NotificationTypeSocial, true},
		{NotificationType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.notifType), func(t *testing.T) {
			result := pref.IsTypeEnabled(tt.notifType)
			assert.Equal(t, tt.enabled, result)
		})
	}
}

func TestUserPreference_DisableChannel(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())

	// Email is enabled by default
	assert.True(t, pref.EmailEnabled)
	assert.True(t, pref.IsChannelEnabled(NotificationChannelEmail))

	// Disable email
	pref.EmailEnabled = false
	assert.False(t, pref.IsChannelEnabled(NotificationChannelEmail))

	// Other channels unchanged
	assert.True(t, pref.IsChannelEnabled(NotificationChannelPush))
	assert.True(t, pref.IsChannelEnabled(NotificationChannelInApp))
}

func TestUserPreference_DisableType(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())

	// Product notifications enabled by default
	assert.True(t, pref.ProductEnabled)
	assert.True(t, pref.IsTypeEnabled(NotificationTypeProduct))

	// Disable product notifications
	pref.ProductEnabled = false
	assert.False(t, pref.IsTypeEnabled(NotificationTypeProduct))

	// System/Security always enabled
	assert.True(t, pref.IsTypeEnabled(NotificationTypeSystem))
	assert.True(t, pref.IsTypeEnabled(NotificationTypeSecurity))
}

func TestUserPreference_IsInQuietHours(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())

	// No quiet hours set - always false
	assert.False(t, pref.IsInQuietHours(time.Now()))

	// Set quiet hours: 22:00 - 08:00
	start, _ := ParseTimeOfDay("22:00")
	end, _ := ParseTimeOfDay("08:00")
	pref.QuietHoursStart = start
	pref.QuietHoursEnd = end

	// Test time at 23:00 (in quiet hours)
	testTime, _ := time.Parse("15:04", "23:00")
	assert.True(t, pref.IsInQuietHours(testTime))

	// Test time at 12:00 (not in quiet hours)
	testTime2, _ := time.Parse("15:04", "12:00")
	assert.False(t, pref.IsInQuietHours(testTime2))

	// Test time at 02:00 (in quiet hours - spans midnight)
	testTime3, _ := time.Parse("15:04", "02:00")
	assert.True(t, pref.IsInQuietHours(testTime3))
}

func TestUserPreference_EnableAllChannels(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())

	// Disable some channels
	pref.SMSEnabled = false
	pref.PushEnabled = false

	// Enable all
	pref.EmailEnabled = true
	pref.SMSEnabled = true
	pref.PushEnabled = true
	pref.InAppEnabled = true

	assert.True(t, pref.IsChannelEnabled(NotificationChannelEmail))
	assert.True(t, pref.IsChannelEnabled(NotificationChannelSMS))
	assert.True(t, pref.IsChannelEnabled(NotificationChannelPush))
	assert.True(t, pref.IsChannelEnabled(NotificationChannelInApp))
}

func TestUserPreference_EnableMarketingNotifications(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())

	// Marketing disabled by default
	assert.False(t, pref.MarketingEnabled)
	assert.False(t, pref.IsTypeEnabled(NotificationTypeMarketing))

	// Enable marketing
	pref.MarketingEnabled = true
	assert.True(t, pref.IsTypeEnabled(NotificationTypeMarketing))
}

func TestUserPreference_TimezoneSettings(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
	}{
		{"UTC", "UTC"},
		{"New York", "America/New_York"},
		{"Tokyo", "Asia/Tokyo"},
		{"London", "Europe/London"},
		{"Kyiv", "Europe/Kiev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pref := NewUserPreference(uuidv7.New())
			pref.Timezone = tt.timezone

			assert.Equal(t, tt.timezone, pref.Timezone)
			assert.NoError(t, pref.Validate())
		})
	}
}

func TestUserPreference_Update(t *testing.T) {
	pref := NewUserPreference(uuidv7.New())
	originalUpdatedAt := pref.UpdatedAt

	time.Sleep(1 * time.Millisecond)

	// Update some fields
	pref.EmailEnabled = false
	pref.MarketingEnabled = true
	pref.Timezone = "America/New_York"
	pref.UpdatedAt = time.Now()

	assert.False(t, pref.EmailEnabled)
	assert.True(t, pref.MarketingEnabled)
	assert.Equal(t, "America/New_York", pref.Timezone)
	assert.True(t, pref.UpdatedAt.After(originalUpdatedAt))
}
