package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_HashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "SecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "long password",
			password: "ThisIsAVeryLongPasswordThatExceeds72BytesInLengthWhichIsTheMaximumForBcrypt1234567890",
			wantErr:  true, // bcrypt returns error for passwords longer than 72 bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{}
			err := user.HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, user.Password)
				assert.NotEqual(t, tt.password, user.Password) // Password should be hashed
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "CorrectPassword123!"
	err := user.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name         string
		testPassword string
		want         bool
	}{
		{
			name:         "correct password",
			testPassword: password,
			want:         true,
		},
		{
			name:         "incorrect password",
			testPassword: "WrongPassword",
			want:         false,
		},
		{
			name:         "empty password",
			testPassword: "",
			want:         false,
		},
		{
			name:         "case sensitive password",
			testPassword: "correctpassword123!",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.testPassword)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{
			name:   "active user",
			status: UserStatusActive,
			want:   true,
		},
		{
			name:   "unverified user",
			status: UserStatusUnverified,
			want:   true,
		},
		{
			name:   "suspended user",
			status: UserStatusSuspended,
			want:   false,
		},
		{
			name:   "banned user",
			status: UserStatusBanned,
			want:   false,
		},
		{
			name:   "inactive user",
			status: UserStatusInactive,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			assert.Equal(t, tt.want, user.IsActive())
		})
	}
}

func TestUser_IsEmailVerified(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		emailVerifiedAt *time.Time
		want            bool
	}{
		{
			name:            "verified email",
			emailVerifiedAt: &now,
			want:            true,
		},
		{
			name:            "unverified email",
			emailVerifiedAt: nil,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{EmailVerifiedAt: tt.emailVerifiedAt}
			assert.Equal(t, tt.want, user.IsEmailVerified())
		})
	}
}

func TestUser_CanLogin(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{
			name:   "active user can login",
			status: UserStatusActive,
			want:   true,
		},
		{
			name:   "unverified user can login",
			status: UserStatusUnverified,
			want:   true,
		},
		{
			name:   "suspended user cannot login",
			status: UserStatusSuspended,
			want:   false,
		},
		{
			name:   "banned user cannot login",
			status: UserStatusBanned,
			want:   false,
		},
		{
			name:   "inactive user cannot login",
			status: UserStatusInactive,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			assert.Equal(t, tt.want, user.CanLogin())
		})
	}
}

func TestUser_Activate(t *testing.T) {
	user := &User{
		ID:              uuidv7.New(),
		Status:          UserStatusUnverified,
		EmailVerifiedAt: nil,
	}

	user.Activate()

	assert.Equal(t, UserStatusActive, user.Status)
	assert.NotNil(t, user.EmailVerifiedAt)
	assert.WithinDuration(t, time.Now(), *user.EmailVerifiedAt, 1*time.Second)
}

func TestUser_Suspend(t *testing.T) {
	user := &User{
		ID:     uuidv7.New(),
		Status: UserStatusActive,
	}

	reason := "Violation of terms"
	until := time.Now().Add(7 * 24 * time.Hour)

	user.Suspend(reason, &until)

	assert.Equal(t, UserStatusSuspended, user.Status)
	assert.NotNil(t, user.SuspendedReason)
	assert.Equal(t, reason, *user.SuspendedReason)
	assert.NotNil(t, user.SuspendedUntil)
	assert.Equal(t, until, *user.SuspendedUntil)
}

func TestUser_Suspend_Permanent(t *testing.T) {
	user := &User{
		ID:     uuidv7.New(),
		Status: UserStatusActive,
	}

	reason := "Multiple violations"

	user.Suspend(reason, nil)

	assert.Equal(t, UserStatusSuspended, user.Status)
	assert.NotNil(t, user.SuspendedReason)
	assert.Equal(t, reason, *user.SuspendedReason)
	assert.Nil(t, user.SuspendedUntil)
}

func TestUser_Ban(t *testing.T) {
	user := &User{
		ID:     uuidv7.New(),
		Status: UserStatusActive,
	}

	reason := "Severe policy violation"

	user.Ban(reason)

	assert.Equal(t, UserStatusBanned, user.Status)
	assert.NotNil(t, user.SuspendedReason)
	assert.Equal(t, reason, *user.SuspendedReason)
}

func TestUser_Deactivate(t *testing.T) {
	user := &User{
		ID:     uuidv7.New(),
		Status: UserStatusActive,
	}

	user.Deactivate()

	assert.Equal(t, UserStatusInactive, user.Status)
}

func TestUser_Reactivate(t *testing.T) {
	reason := "Previously suspended"
	until := time.Now().Add(7 * 24 * time.Hour)

	user := &User{
		ID:              uuidv7.New(),
		Status:          UserStatusSuspended,
		SuspendedReason: &reason,
		SuspendedUntil:  &until,
	}

	user.Reactivate()

	assert.Equal(t, UserStatusActive, user.Status)
	assert.Nil(t, user.SuspendedReason)
	assert.Nil(t, user.SuspendedUntil)
}

func TestUser_UpdateLastLogin(t *testing.T) {
	user := &User{
		ID:          uuidv7.New(),
		LastLoginAt: nil,
	}

	user.UpdateLastLogin()

	assert.NotNil(t, user.LastLoginAt)
	assert.WithinDuration(t, time.Now(), *user.LastLoginAt, 1*time.Second)

	// Test update of existing last login
	firstLogin := *user.LastLoginAt
	time.Sleep(10 * time.Millisecond)
	user.UpdateLastLogin()

	assert.NotEqual(t, firstLogin, *user.LastLoginAt)
	assert.True(t, user.LastLoginAt.After(firstLogin))
}

func TestUser_StatusConstants(t *testing.T) {
	// Ensure status constants have expected values
	assert.Equal(t, UserStatus("unverified"), UserStatusUnverified)
	assert.Equal(t, UserStatus("active"), UserStatusActive)
	assert.Equal(t, UserStatus("suspended"), UserStatusSuspended)
	assert.Equal(t, UserStatus("banned"), UserStatusBanned)
	assert.Equal(t, UserStatus("inactive"), UserStatusInactive)
}
