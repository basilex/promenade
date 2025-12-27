package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewPhone(t *testing.T) {
	t.Run("creates valid phone with country code", func(t *testing.T) {
		phone, err := valueobject.NewPhone("+380501234567")
		
		require.NoError(t, err)
		assert.Equal(t, "+380501234567", phone.String())
	})

	t.Run("accepts phone with spaces", func(t *testing.T) {
		phone, err := valueobject.NewPhone("+380 50 123 4567")
		
		require.NoError(t, err)
		assert.Equal(t, "+380501234567", phone.String())
	})

	t.Run("accepts phone with dashes", func(t *testing.T) {
		phone, err := valueobject.NewPhone("+380-50-123-4567")
		
		require.NoError(t, err)
		assert.Equal(t, "+380501234567", phone.String())
	})

	t.Run("accepts phone with parentheses", func(t *testing.T) {
		phone, err := valueobject.NewPhone("+1 (555) 123-4567")
		
		require.NoError(t, err)
		assert.Equal(t, "+15551234567", phone.String())
	})

	t.Run("rejects empty phone", func(t *testing.T) {
		_, err := valueobject.NewPhone("")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone number is required")
	})

	t.Run("rejects phone without plus sign", func(t *testing.T) {
		_, err := valueobject.NewPhone("380501234567")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must start with +")
	})

	t.Run("rejects invalid phone format", func(t *testing.T) {
		_, err := valueobject.NewPhone("+0123456789") // starts with 0 after +, invalid per E.164
		
		assert.Error(t, err)
	})

	t.Run("rejects phone too long", func(t *testing.T) {
		_, err := valueobject.NewPhone("+12345678901234567") // > 15 digits
		
		assert.Error(t, err)
	})

	t.Run("accepts various country codes", func(t *testing.T) {
		tests := []struct {
			name  string
			phone string
		}{
			{"US", "+12025551234"},
			{"UK", "+442079460958"},
			{"Germany", "+4930123456"},
			{"France", "+33123456789"},
			{"Poland", "+48123456789"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				phone, err := valueobject.NewPhone(tt.phone)
				require.NoError(t, err)
				assert.NotEmpty(t, phone.String())
			})
		}
	})
}

func TestPhone_CountryCode(t *testing.T) {
	t.Run("extracts Ukraine country code", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+380501234567")
		
		assert.Equal(t, "380", phone.CountryCode())
	})

	t.Run("extracts US country code", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+12025551234")
		
		assert.Equal(t, "1", phone.CountryCode())
	})

	t.Run("extracts UK country code", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+442079460958")
		
		assert.Equal(t, "44", phone.CountryCode())
	})

	t.Run("extracts Germany country code", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+4930123456")
		
		assert.Equal(t, "49", phone.CountryCode())
	})

	t.Run("returns empty for unrecognized code", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+999123456789")
		
		assert.Equal(t, "", phone.CountryCode())
	})
}

func TestPhone_Formatted(t *testing.T) {
	t.Run("formats Ukraine phone", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+380501234567")
		
		formatted := phone.Formatted()
		assert.Contains(t, formatted, "+380")
		assert.Contains(t, formatted, "50")
	})

	t.Run("formats US phone", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+12025551234")
		
		formatted := phone.Formatted()
		assert.Contains(t, formatted, "+1")
		assert.Contains(t, formatted, " ") // has spaces for readability
	})

	t.Run("formatted output is human readable", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+380501234567")
		
		formatted := phone.Formatted()
		assert.NotEqual(t, phone.String(), formatted)
		assert.Contains(t, formatted, " ")
	})
}

func TestPhone_Equals(t *testing.T) {
	t.Run("equal phones are equal", func(t *testing.T) {
		phone1, _ := valueobject.NewPhone("+380501234567")
		phone2, _ := valueobject.NewPhone("+380 50 123 4567")
		
		assert.True(t, phone1.Equals(phone2))
	})

	t.Run("different phones are not equal", func(t *testing.T) {
		phone1, _ := valueobject.NewPhone("+380501234567")
		phone2, _ := valueobject.NewPhone("+380509876543")
		
		assert.False(t, phone1.Equals(phone2))
	})
}

func TestPhone_IsEmpty(t *testing.T) {
	t.Run("non-empty phone is not empty", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+380501234567")
		
		assert.False(t, phone.IsEmpty())
	})

	t.Run("zero value is empty", func(t *testing.T) {
		var phone valueobject.Phone
		
		assert.True(t, phone.IsEmpty())
	})
}

func TestPhone_String(t *testing.T) {
	t.Run("returns E.164 format", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+380 50 123 4567")
		
		assert.Equal(t, "+380501234567", phone.String())
	})

	t.Run("string is normalized", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+1 (555) 123-4567")
		
		str := phone.String()
		assert.NotContains(t, str, " ")
		assert.NotContains(t, str, "(")
		assert.NotContains(t, str, ")")
		assert.NotContains(t, str, "-")
	})
}
