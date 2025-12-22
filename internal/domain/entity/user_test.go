package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_HashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty, but validation should catch it
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
				assert.NotEqual(t, tt.password, user.Password)
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	err := user.HashPassword("correctpassword")
	assert.NoError(t, err)

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: "correctpassword",
			want:     true,
		},
		{
			name:     "incorrect password",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.password)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestUser_Status(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   UserStatus
	}{
		{
			name:   "active user",
			status: UserStatusActive,
			want:   UserStatusActive,
		},
		{
			name:   "suspended user",
			status: UserStatusSuspended,
			want:   UserStatusSuspended,
		},
		{
			name:   "banned user",
			status: UserStatusBanned,
			want:   UserStatusBanned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Status: tt.status}
			assert.Equal(t, tt.want, user.Status)
		})
	}
}
