package entity

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestContactType_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		contactType ContactType
		want        bool
	}{
		{"email", ContactTypeEmail, true},
		{"phone", ContactTypePhone, true},
		{"telegram", ContactTypeTelegram, true},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.contactType.IsValid())
		})
	}
}

func TestUserContact_Creation(t *testing.T) {
	contact := &UserContact{
		ID:           uuidv7.New(),
		UserID:       uuidv7.New(),
		ContactType:  ContactTypeEmail,
		ContactValue: "test@example.com",
		IsPrimary:    true,
		IsVerified:   false,
	}

	assert.NotNil(t, contact)
	assert.Equal(t, ContactTypeEmail, contact.ContactType)
	assert.True(t, contact.IsPrimary)
	assert.False(t, contact.IsVerified)
}

func TestUserContact_Flags(t *testing.T) {
	contact := &UserContact{
		ID:           uuidv7.New(),
		UserID:       uuidv7.New(),
		ContactType:  ContactTypePhone,
		ContactValue: "+1234567890",
		IsVerified:   true,
		IsPrimary:    false,
		IsActive:     true,
		IsPublic:     false,
	}

	assert.True(t, contact.IsVerified)
	assert.False(t, contact.IsPrimary)
	assert.True(t, contact.IsActive)
	assert.False(t, contact.IsPublic)
}

func TestUserContact_Validate(t *testing.T) {
	userID := uuidv7.New()

	tests := []struct {
		name    string
		contact *UserContact
		wantErr bool
	}{
		{
			name: "valid email",
			contact: &UserContact{
				ID:           uuidv7.New(),
				UserID:       userID,
				ContactType:  ContactTypeEmail,
				ContactValue: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			contact: &UserContact{
				ID:           uuidv7.New(),
				ContactType:  ContactTypeEmail,
				ContactValue: "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid contact type",
			contact: &UserContact{
				ID:           uuidv7.New(),
				UserID:       userID,
				ContactType:  "invalid",
				ContactValue: "value",
			},
			wantErr: true,
		},
		{
			name: "empty contact value",
			contact: &UserContact{
				ID:          uuidv7.New(),
				UserID:      userID,
				ContactType: ContactTypeEmail,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contact.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserContact_IsAvailable(t *testing.T) {
	// Test time: Tuesday, 10:00 AM
	testTime := time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC)

	availFrom := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	availTo := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		contact *UserContact
		want    bool
	}{
		{
			name: "no restrictions",
			contact: &UserContact{
				ID:           uuidv7.New(),
				UserID:       uuidv7.New(),
				ContactType:  ContactTypeEmail,
				ContactValue: "test@example.com",
			},
			want: true,
		},
		{
			name: "available in time window",
			contact: &UserContact{
				ID:            uuidv7.New(),
				UserID:        uuidv7.New(),
				ContactType:   ContactTypeEmail,
				ContactValue:  "test@example.com",
				AvailableFrom: &availFrom,
				AvailableTo:   &availTo,
			},
			want: true,
		},
		{
			name: "available on tuesday",
			contact: &UserContact{
				ID:            uuidv7.New(),
				UserID:        uuidv7.New(),
				ContactType:   ContactTypeEmail,
				ContactValue:  "test@example.com",
				AvailableDays: pq.StringArray{"monday", "tuesday", "friday"},
			},
			want: true,
		},
		{
			name: "not available on tuesday",
			contact: &UserContact{
				ID:            uuidv7.New(),
				UserID:        uuidv7.New(),
				ContactType:   ContactTypeEmail,
				ContactValue:  "test@example.com",
				AvailableDays: pq.StringArray{"monday", "wednesday", "friday"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.contact.IsAvailable(testTime))
		})
	}
}
