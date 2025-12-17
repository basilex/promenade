package entity

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestContactType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		ct       ContactType
		expected bool
	}{
		{"valid email", ContactTypeEmail, true},
		{"valid phone", ContactTypePhone, true},
		{"valid telegram", ContactTypeTelegram, true},
		{"valid whatsapp", ContactTypeWhatsApp, true},
		{"valid viber", ContactTypeViber, true},
		{"valid signal", ContactTypeSignal, true},
		{"valid skype", ContactTypeSkype, true},
		{"valid discord", ContactTypeDiscord, true},
		{"valid linkedin", ContactTypeLinkedIn, true},
		{"valid other", ContactTypeOther, true},
		{"invalid empty", ContactType(""), false},
		{"invalid unknown", ContactType("unknown"), false},
		{"invalid facebook", ContactType("facebook"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ct.IsValid())
		})
	}
}

func TestUserContact_Validate(t *testing.T) {
	userID := uuidv7.New()
	contactID := uuidv7.New()
	label := "Work"
	timezone := "UTC"
	notes := "Available during work hours"

	t.Run("valid contact", func(t *testing.T) {
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactTypeEmail,
			ContactValue: "test@example.com",
			Label:        &label,
			IsVerified:   false,
			IsPrimary:    false,
			IsActive:     true,
			IsPublic:     true,
			Timezone:     &timezone,
			Notes:        &notes,
		}

		err := contact.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid contact type", func(t *testing.T) {
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactType("invalid"),
			ContactValue: "test@example.com",
		}

		err := contact.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("empty contact value", func(t *testing.T) {
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactTypeEmail,
			ContactValue: "",
		}

		err := contact.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("contact value too long", func(t *testing.T) {
		longValue := string(make([]byte, 256))
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactTypeEmail,
			ContactValue: longValue,
		}

		err := contact.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("label too long", func(t *testing.T) {
		longLabel := string(make([]byte, 101))
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactTypeEmail,
			ContactValue: "test@example.com",
			Label:        &longLabel,
		}

		err := contact.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("timezone too long", func(t *testing.T) {
		longTimezone := string(make([]byte, 51))
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactTypeEmail,
			ContactValue: "test@example.com",
			Timezone:     &longTimezone,
		}

		err := contact.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("notes too long", func(t *testing.T) {
		longNotes := string(make([]byte, 1001))
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactTypeEmail,
			ContactValue: "test@example.com",
			Notes:        &longNotes,
		}

		err := contact.Validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})
}

func TestUserContact_IsAvailable(t *testing.T) {
	userID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("always available - no restrictions", func(t *testing.T) {
		contact := &UserContact{
			ID:           contactID,
			UserID:       userID,
			ContactType:  ContactTypeEmail,
			ContactValue: "test@example.com",
		}

		// Tuesday 10:00 AM
		testTime := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))
	})

	t.Run("available during time range", func(t *testing.T) {
		from := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		to := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)

		contact := &UserContact{
			ID:            contactID,
			UserID:        userID,
			ContactType:   ContactTypeEmail,
			ContactValue:  "test@example.com",
			AvailableFrom: &from,
			AvailableTo:   &to,
		}

		// 10:00 AM - within range
		testTime := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))

		// 8:00 AM - before range
		testTime = time.Date(2024, 1, 2, 8, 0, 0, 0, time.UTC)
		assert.False(t, contact.IsAvailable(testTime))

		// 18:00 - after range
		testTime = time.Date(2024, 1, 2, 18, 0, 0, 0, time.UTC)
		assert.False(t, contact.IsAvailable(testTime))
	})

	t.Run("available on specific days", func(t *testing.T) {
		contact := &UserContact{
			ID:            contactID,
			UserID:        userID,
			ContactType:   ContactTypeEmail,
			ContactValue:  "test@example.com",
			AvailableDays: pq.StringArray{"monday", "wednesday", "friday"},
		}

		// Monday - available
		testTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))

		// Tuesday - not available
		testTime = time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
		assert.False(t, contact.IsAvailable(testTime))

		// Wednesday - available
		testTime = time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))

		// Thursday - not available
		testTime = time.Date(2024, 1, 4, 10, 0, 0, 0, time.UTC)
		assert.False(t, contact.IsAvailable(testTime))

		// Friday - available
		testTime = time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))
	})

	t.Run("available during time range on specific days", func(t *testing.T) {
		from := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		to := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)

		contact := &UserContact{
			ID:            contactID,
			UserID:        userID,
			ContactType:   ContactTypeEmail,
			ContactValue:  "test@example.com",
			AvailableFrom: &from,
			AvailableTo:   &to,
			AvailableDays: []string{"monday", "wednesday", "friday"},
		}

		// Monday 10:00 AM - available (right day and time)
		testTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))

		// Monday 8:00 AM - not available (right day, wrong time)
		testTime = time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
		assert.False(t, contact.IsAvailable(testTime))

		// Tuesday 10:00 AM - not available (wrong day, right time)
		testTime = time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)
		assert.False(t, contact.IsAvailable(testTime))
	})

	t.Run("case insensitive day matching", func(t *testing.T) {
		contact := &UserContact{
			ID:            contactID,
			UserID:        userID,
			ContactType:   ContactTypeEmail,
			ContactValue:  "test@example.com",
			AvailableDays: pq.StringArray{"Monday", "WEDNESDAY", "FriDay"},
		}

		// Monday
		testTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))

		// Wednesday
		testTime = time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))

		// Friday
		testTime = time.Date(2024, 1, 5, 10, 0, 0, 0, time.UTC)
		assert.True(t, contact.IsAvailable(testTime))
	})

	t.Run("empty available days means always available", func(t *testing.T) {
		contact := &UserContact{
			ID:            contactID,
			UserID:        userID,
			ContactType:   ContactTypeEmail,
			ContactValue:  "test@example.com",
			AvailableDays: pq.StringArray{},
		}

		// Any day should be available
		for day := 1; day <= 7; day++ {
			testTime := time.Date(2024, 1, day, 10, 0, 0, 0, time.UTC)
			assert.True(t, contact.IsAvailable(testTime), "day %d should be available", day)
		}
	})
}

func TestContactTypeConstants(t *testing.T) {
	t.Run("all contact types defined", func(t *testing.T) {
		types := []ContactType{
			ContactTypeEmail,
			ContactTypePhone,
			ContactTypeTelegram,
			ContactTypeWhatsApp,
			ContactTypeViber,
			ContactTypeSignal,
			ContactTypeSkype,
			ContactTypeDiscord,
			ContactTypeLinkedIn,
			ContactTypeOther,
		}

		for _, ct := range types {
			assert.True(t, ct.IsValid(), "contact type %s should be valid", ct)
			assert.NotEmpty(t, string(ct), "contact type should not be empty")
		}
	})

	t.Run("contact types are unique", func(t *testing.T) {
		types := []ContactType{
			ContactTypeEmail,
			ContactTypePhone,
			ContactTypeTelegram,
			ContactTypeWhatsApp,
			ContactTypeViber,
			ContactTypeSignal,
			ContactTypeSkype,
			ContactTypeDiscord,
			ContactTypeLinkedIn,
			ContactTypeOther,
		}

		seen := make(map[ContactType]bool)
		for _, ct := range types {
			require.False(t, seen[ct], "contact type %s should be unique", ct)
			seen[ct] = true
		}
	})
}
