package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewEmail(t *testing.T) {
	t.Run("creates valid email", func(t *testing.T) {
		email, err := valueobject.NewEmail("user@example.com")

		require.NoError(t, err)
		assert.Equal(t, "user@example.com", email.String())
	})

	t.Run("normalizes to lowercase", func(t *testing.T) {
		email, err := valueobject.NewEmail("User@Example.COM")

		require.NoError(t, err)
		assert.Equal(t, "user@example.com", email.String())
	})

	t.Run("trims whitespace", func(t *testing.T) {
		email, err := valueobject.NewEmail("  user@example.com  ")

		require.NoError(t, err)
		assert.Equal(t, "user@example.com", email.String())
	})

	t.Run("rejects empty email", func(t *testing.T) {
		_, err := valueobject.NewEmail("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("rejects email without @", func(t *testing.T) {
		_, err := valueobject.NewEmail("userexample.com")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("rejects email without domain", func(t *testing.T) {
		_, err := valueobject.NewEmail("user@")

		assert.Error(t, err)
	})

	t.Run("rejects email too long", func(t *testing.T) {
		longEmail := string(make([]byte, 255)) + "@example.com"
		_, err := valueobject.NewEmail(longEmail)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
	})

	t.Run("accepts email with plus sign", func(t *testing.T) {
		email, err := valueobject.NewEmail("user+tag@example.com")

		require.NoError(t, err)
		assert.Equal(t, "user+tag@example.com", email.String())
	})

	t.Run("accepts email with dots", func(t *testing.T) {
		email, err := valueobject.NewEmail("first.last@example.com")

		require.NoError(t, err)
		assert.Equal(t, "first.last@example.com", email.String())
	})
}

func TestEmail_Domain(t *testing.T) {
	t.Run("extracts domain correctly", func(t *testing.T) {
		email, _ := valueobject.NewEmail("user@example.com")

		assert.Equal(t, "example.com", email.Domain())
	})

	t.Run("handles subdomain", func(t *testing.T) {
		email, _ := valueobject.NewEmail("user@mail.example.com")

		assert.Equal(t, "mail.example.com", email.Domain())
	})
}

func TestEmail_LocalPart(t *testing.T) {
	t.Run("extracts local part correctly", func(t *testing.T) {
		email, _ := valueobject.NewEmail("user@example.com")

		assert.Equal(t, "user", email.LocalPart())
	})

	t.Run("handles complex local part", func(t *testing.T) {
		email, _ := valueobject.NewEmail("first.last+tag@example.com")

		assert.Equal(t, "first.last+tag", email.LocalPart())
	})
}

func TestEmail_Equals(t *testing.T) {
	t.Run("equal emails are equal", func(t *testing.T) {
		email1, _ := valueobject.NewEmail("user@example.com")
		email2, _ := valueobject.NewEmail("user@example.com")

		assert.True(t, email1.Equals(email2))
	})

	t.Run("different emails are not equal", func(t *testing.T) {
		email1, _ := valueobject.NewEmail("user1@example.com")
		email2, _ := valueobject.NewEmail("user2@example.com")

		assert.False(t, email1.Equals(email2))
	})
}

func TestEmail_IsEmpty(t *testing.T) {
	t.Run("non-empty email is not empty", func(t *testing.T) {
		email, _ := valueobject.NewEmail("user@example.com")

		assert.False(t, email.IsEmpty())
	})

	t.Run("zero value is empty", func(t *testing.T) {
		var email valueobject.Email

		assert.True(t, email.IsEmpty())
	})
}
