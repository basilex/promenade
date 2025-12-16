package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired session - past",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "expired session - 1 second ago",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
		{
			name:      "valid session - future",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "valid session - 1 second from now",
			expiresAt: time.Now().Add(1 * time.Second),
			want:      false,
		},
		{
			name:      "valid session - far future",
			expiresAt: time.Now().Add(24 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &Session{
				ID:           uuidv7.New(),
				UserID:       uuidv7.New(),
				RefreshToken: "hashed_token",
				ExpiresAt:    tt.expiresAt,
				CreatedAt:    time.Now(),
			}

			result := session.IsExpired()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSession_IsExpired_EdgeCase(t *testing.T) {
	session := &Session{
		ID:           uuidv7.New(),
		UserID:       uuidv7.New(),
		RefreshToken: "hashed_token",
		ExpiresAt:    time.Now().Add(10 * time.Millisecond),
		CreatedAt:    time.Now(),
	}

	assert.False(t, session.IsExpired())

	time.Sleep(20 * time.Millisecond)

	assert.True(t, session.IsExpired())
}

func TestSession_Fields(t *testing.T) {
	userID := uuidv7.New()
	sessionID := uuidv7.New()
	refreshToken := "hashed_refresh_token_value"
	userAgent := "Mozilla/5.0"
	ipAddress := "192.168.1.1"
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	createdAt := time.Now()

	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    &userAgent,
		IPAddress:    &ipAddress,
		ExpiresAt:    expiresAt,
		CreatedAt:    createdAt,
	}

	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, refreshToken, session.RefreshToken)
	assert.NotNil(t, session.UserAgent)
	assert.Equal(t, userAgent, *session.UserAgent)
	assert.NotNil(t, session.IPAddress)
	assert.Equal(t, ipAddress, *session.IPAddress)
	assert.Equal(t, expiresAt, session.ExpiresAt)
	assert.Equal(t, createdAt, session.CreatedAt)
}

func TestSession_OptionalFields(t *testing.T) {
	session := &Session{
		ID:           uuidv7.New(),
		UserID:       uuidv7.New(),
		RefreshToken: "hashed_token",
		UserAgent:    nil,
		IPAddress:    nil,
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		CreatedAt:    time.Now(),
	}

	assert.Nil(t, session.UserAgent)
	assert.Nil(t, session.IPAddress)
	assert.False(t, session.IsExpired())
}
