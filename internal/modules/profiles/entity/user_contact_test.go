package entity

import (
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
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
