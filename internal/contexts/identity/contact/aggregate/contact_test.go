package aggregate

import (
	"errors"
	"testing"

	contacterrors "github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailContact(t *testing.T) {
	userID := uuidv7.New()

	t.Run("valid email contact", func(t *testing.T) {
		contact, err := NewEmailContact(userID, "test@example.com", "Work")

		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, contact.ID)
		assert.Equal(t, userID, contact.UserID)
		assert.Equal(t, ContactTypeEmail, contact.Type)
		assert.Equal(t, "Work", contact.Label)
		assert.NotNil(t, contact.Email)
		assert.Equal(t, "test@example.com", contact.Email.Value())
		assert.Nil(t, contact.Phone)
		assert.Nil(t, contact.Address)
		assert.False(t, contact.IsPrimary)
		assert.False(t, contact.IsVerified)
		assert.False(t, contact.IsPublic)
	})

	t.Run("invalid email", func(t *testing.T) {
		_, err := NewEmailContact(userID, "invalid-email", "Work")
		assert.Error(t, err)
	})

	t.Run("empty label", func(t *testing.T) {
		_, err := NewEmailContact(userID, "test@example.com", "")
		assert.Error(t, err)
	})
}

func TestNewPhoneContact(t *testing.T) {
	userID := uuidv7.New()

	t.Run("valid phone contact", func(t *testing.T) {
		contact, err := NewPhoneContact(userID, "+380501234567", "Mobile")

		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, contact.ID)
		assert.Equal(t, userID, contact.UserID)
		assert.Equal(t, ContactTypePhone, contact.Type)
		assert.Equal(t, "Mobile", contact.Label)
		assert.NotNil(t, contact.Phone)
		assert.Equal(t, "+380501234567", contact.Phone.Value())
		assert.Nil(t, contact.Email)
		assert.Nil(t, contact.Address)
	})

	t.Run("invalid phone", func(t *testing.T) {
		_, err := NewPhoneContact(userID, "123", "Mobile")
		assert.Error(t, err)
	})
}

func TestNewAddressContact(t *testing.T) {
	userID := uuidv7.New()

	t.Run("valid address contact", func(t *testing.T) {
		contact, err := NewAddressContact(
			userID,
			"123 Main St",
			"Kyiv",
			"UA",
			"01001",
			"Home",
		)

		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, contact.ID)
		assert.Equal(t, userID, contact.UserID)
		assert.Equal(t, ContactTypeAddress, contact.Type)
		assert.Equal(t, "Home", contact.Label)
		assert.NotNil(t, contact.Address)
		assert.Equal(t, "123 Main St", contact.Address.Street)
		assert.Equal(t, "Kyiv", contact.Address.City)
		assert.Equal(t, "UA", contact.Address.Country)
		assert.Equal(t, "01001", contact.Address.PostalCode)
		assert.Nil(t, contact.Email)
		assert.Nil(t, contact.Phone)
	})
}

func TestContact_SetEmail(t *testing.T) {
	userID := uuidv7.New()

	t.Run("set email on email contact", func(t *testing.T) {
		contact, _ := NewEmailContact(userID, "old@example.com", "Work")
		err := contact.SetEmail("new@example.com")

		require.NoError(t, err)
		assert.Equal(t, "new@example.com", contact.Email.Value())
	})

	t.Run("cannot set email on phone contact", func(t *testing.T) {
		contact, _ := NewPhoneContact(userID, "+380501234567", "Mobile")
		err := contact.SetEmail("test@example.com")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, contacterrors.ErrCannotSetEmailOnNonEmailContact))
	})

	t.Run("invalid email", func(t *testing.T) {
		contact, _ := NewEmailContact(userID, "old@example.com", "Work")
		err := contact.SetEmail("invalid")

		assert.Error(t, err)
	})
}

func TestContact_SetPhone(t *testing.T) {
	userID := uuidv7.New()

	t.Run("set phone on phone contact", func(t *testing.T) {
		contact, _ := NewPhoneContact(userID, "+380501234567", "Mobile")
		err := contact.SetPhone("+380509876543")

		require.NoError(t, err)
		assert.Equal(t, "+380509876543", contact.Phone.Value())
	})

	t.Run("cannot set phone on email contact", func(t *testing.T) {
		contact, _ := NewEmailContact(userID, "test@example.com", "Work")
		err := contact.SetPhone("+380501234567")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, contacterrors.ErrCannotSetPhoneOnNonPhoneContact))
	})
}

func TestContact_SetAddress(t *testing.T) {
	userID := uuidv7.New()

	t.Run("set address on address contact", func(t *testing.T) {
		contact, _ := NewAddressContact(userID, "Old St", "Kyiv", "UA", "01001", "Home")
		err := contact.SetAddress("New St", "Lviv", "UA", "79000")

		require.NoError(t, err)
		assert.Equal(t, "New St", contact.Address.Street)
		assert.Equal(t, "Lviv", contact.Address.City)
	})

	t.Run("cannot set address on email contact", func(t *testing.T) {
		contact, _ := NewEmailContact(userID, "test@example.com", "Work")
		err := contact.SetAddress("Street", "City", "UA", "12345")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, contacterrors.ErrCannotSetAddressOnNonAddressContact))
	})
}

func TestContact_SetAsPrimary(t *testing.T) {
	userID := uuidv7.New()
	contact, _ := NewEmailContact(userID, "test@example.com", "Work")

	assert.False(t, contact.IsPrimary)

	contact.SetAsPrimary()
	assert.True(t, contact.IsPrimary)

	contact.UnsetAsPrimary()
	assert.False(t, contact.IsPrimary)
}

func TestContact_Verify(t *testing.T) {
	userID := uuidv7.New()
	contact, _ := NewEmailContact(userID, "test@example.com", "Work")

	assert.False(t, contact.IsVerified)

	contact.Verify()
	assert.True(t, contact.IsVerified)

	contact.Unverify()
	assert.False(t, contact.IsVerified)
}

func TestContact_MakePublic(t *testing.T) {
	userID := uuidv7.New()
	contact, _ := NewEmailContact(userID, "test@example.com", "Work")

	assert.False(t, contact.IsPublic)

	contact.MakePublic()
	assert.True(t, contact.IsPublic)

	contact.MakePrivate()
	assert.False(t, contact.IsPublic)
}

func TestContact_UpdateLabel(t *testing.T) {
	userID := uuidv7.New()
	contact, _ := NewEmailContact(userID, "test@example.com", "Work")

	err := contact.UpdateLabel("Personal")
	require.NoError(t, err)
	assert.Equal(t, "Personal", contact.Label)

	err = contact.UpdateLabel("")
	assert.Error(t, err)
}

func TestContact_Validate(t *testing.T) {
	userID := uuidv7.New()

	t.Run("valid email contact", func(t *testing.T) {
		contact, _ := NewEmailContact(userID, "test@example.com", "Work")
		err := contact.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid phone contact", func(t *testing.T) {
		contact, _ := NewPhoneContact(userID, "+380501234567", "Mobile")
		err := contact.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid address contact", func(t *testing.T) {
		contact, _ := NewAddressContact(userID, "Street", "City", "US", "12345", "Home")
		err := contact.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid - empty label", func(t *testing.T) {
		contact, _ := NewEmailContact(userID, "test@example.com", "Work")
		contact.Label = ""
		err := contact.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid - no email for email contact", func(t *testing.T) {
		contact, _ := NewContact(userID, ContactTypeEmail, "Work")
		err := contact.Validate()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, contacterrors.ErrEmailRequiredForEmailContact))
	})
}

func TestContact_GetValue(t *testing.T) {
	userID := uuidv7.New()

	t.Run("email contact value", func(t *testing.T) {
		contact, _ := NewEmailContact(userID, "test@example.com", "Work")
		assert.Equal(t, "test@example.com", contact.GetValue())
	})

	t.Run("phone contact value", func(t *testing.T) {
		contact, _ := NewPhoneContact(userID, "+380501234567", "Mobile")
		assert.Equal(t, "+380501234567", contact.GetValue())
	})

	t.Run("address contact value", func(t *testing.T) {
		contact, _ := NewAddressContact(userID, "123 Main St", "Kyiv", "UA", "01001", "Home")
		value := contact.GetValue()
		assert.Contains(t, value, "123 Main St")
		assert.Contains(t, value, "Kyiv")
	})
}

func TestValidateContactType(t *testing.T) {
	tests := []struct {
		name    string
		input   ContactType
		wantErr bool
	}{
		{"valid email", ContactTypeEmail, false},
		{"valid phone", ContactTypePhone, false},
		{"valid address", ContactTypeAddress, false},
		{"invalid type", ContactType("invalid"), true},
		{"empty type", ContactType(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContactType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
