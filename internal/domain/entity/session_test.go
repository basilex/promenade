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
			name:      "not expired",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := Session{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.want, session.IsExpired())
		})
	}
}

func TestSession_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		session Session
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid session",
			session: Session{
				UserID:       uuidv7.New(),
				RefreshToken: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
				CreatedAt:    now,
				ExpiresAt:    now.Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "empty user_id",
			session: Session{
				RefreshToken: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
				CreatedAt:    now,
				ExpiresAt:    now.Add(24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "empty refresh_token",
			session: Session{
				UserID:    uuidv7.New(),
				CreatedAt: now,
				ExpiresAt: now.Add(24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "refresh_token is required",
		},
		{
			name: "refresh_token too short",
			session: Session{
				UserID:       uuidv7.New(),
				RefreshToken: "short",
				CreatedAt:    now,
				ExpiresAt:    now.Add(24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "refresh_token must be at least 32 characters",
		},
		{
			name: "expires_at before created_at",
			session: Session{
				UserID:       uuidv7.New(),
				RefreshToken: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
				CreatedAt:    now,
				ExpiresAt:    now.Add(-1 * time.Hour),
			},
			wantErr: true,
			errMsg:  "expires_at must be after created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
