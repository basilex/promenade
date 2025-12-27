package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewAddress(t *testing.T) {
	t.Run("creates valid address", func(t *testing.T) {
		addr, err := valueobject.NewAddress(
			"123 Main St",
			"New York",
			"10001",
			"US",
		)
		
		require.NoError(t, err)
		assert.Equal(t, "123 Main St", addr.Street)
		assert.Equal(t, "New York", addr.City)
		assert.Equal(t, "10001", addr.PostalCode)
		assert.Equal(t, "US", addr.Country)
	})

	t.Run("normalizes country to uppercase", func(t *testing.T) {
		addr, err := valueobject.NewAddress(
			"123 Main St",
			"Kyiv",
			"01001",
			"ua",
		)
		
		require.NoError(t, err)
		assert.Equal(t, "UA", addr.Country)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		addr, err := valueobject.NewAddress(
			"  123 Main St  ",
			"  New York  ",
			"  10001  ",
			"  US  ",
		)
		
		require.NoError(t, err)
		assert.Equal(t, "123 Main St", addr.Street)
		assert.Equal(t, "New York", addr.City)
		assert.Equal(t, "10001", addr.PostalCode)
		assert.Equal(t, "US", addr.Country)
	})

	t.Run("rejects empty street", func(t *testing.T) {
		_, err := valueobject.NewAddress("", "New York", "10001", "US")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "street is required")
	})

	t.Run("rejects empty city", func(t *testing.T) {
		_, err := valueobject.NewAddress("123 Main St", "", "10001", "US")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "city is required")
	})

	t.Run("rejects empty postal code", func(t *testing.T) {
		_, err := valueobject.NewAddress("123 Main St", "New York", "", "US")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "postal code is required")
	})

	t.Run("rejects empty country", func(t *testing.T) {
		_, err := valueobject.NewAddress("123 Main St", "New York", "10001", "")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "country is required")
	})

	t.Run("rejects invalid country code length", func(t *testing.T) {
		_, err := valueobject.NewAddress("123 Main St", "New York", "10001", "USA")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be 2-letter ISO code")
	})

	t.Run("accepts optional street2", func(t *testing.T) {
		addr, err := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr.Street2 = "Apt 4B"
		
		require.NoError(t, err)
		assert.Equal(t, "Apt 4B", addr.Street2)
	})

	t.Run("accepts optional state", func(t *testing.T) {
		addr, err := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr.State = "NY"
		
		require.NoError(t, err)
		assert.Equal(t, "NY", addr.State)
	})
}

func TestAddress_IsDomestic(t *testing.T) {
	t.Run("returns true for same country", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		assert.True(t, addr.IsDomestic("US"))
	})

	t.Run("returns false for different country", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		assert.False(t, addr.IsDomestic("CA"))
	})

	t.Run("is case insensitive", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		assert.True(t, addr.IsDomestic("us"))
	})
}

func TestAddress_SingleLine(t *testing.T) {
	t.Run("formats address without street2", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		single := addr.SingleLine()
		assert.Contains(t, single, "123 Main St")
		assert.Contains(t, single, "New York")
		assert.Contains(t, single, "10001")
		assert.Contains(t, single, "US")
	})

	t.Run("formats address with street2", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr.Street2 = "Apt 4B"
		
		single := addr.SingleLine()
		assert.Contains(t, single, "123 Main St")
		assert.Contains(t, single, "Apt 4B")
		assert.Contains(t, single, "New York")
		assert.Contains(t, single, "10001")
		assert.Contains(t, single, "US")
	})

	t.Run("formats address with state", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr.State = "NY"
		
		single := addr.SingleLine()
		assert.Contains(t, single, "123 Main St")
		assert.Contains(t, single, "New York")
		assert.Contains(t, single, "NY")
		assert.Contains(t, single, "10001")
		assert.Contains(t, single, "US")
	})

	t.Run("separates parts with comma", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		single := addr.SingleLine()
		assert.Contains(t, single, ",")
	})
}

func TestAddress_StringFormat(t *testing.T) {
	t.Run("String returns multiline format", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		str := addr.String()
		assert.Contains(t, str, "\n") // multiline has newlines
		assert.Contains(t, str, "123 Main St")
		assert.Contains(t, str, "New York")
		assert.Contains(t, str, "10001")
		assert.Contains(t, str, "US")
	})

	t.Run("String formats address with street2", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr.Street2 = "Apt 4B"
		
		str := addr.String()
		assert.Contains(t, str, "123 Main St")
		assert.Contains(t, str, "Apt 4B")
		assert.Contains(t, str, "\n")
	})

	t.Run("String formats address with state", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr.State = "NY"
		
		str := addr.String()
		assert.Contains(t, str, "New York, NY")
		assert.Contains(t, str, "10001")
	})
}

func TestAddress_Equals(t *testing.T) {
	t.Run("equal addresses are equal", func(t *testing.T) {
		addr1, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr2, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		assert.True(t, addr1.Equals(addr2))
	})

	t.Run("different streets are not equal", func(t *testing.T) {
		addr1, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr2, _ := valueobject.NewAddress("456 Oak Ave", "New York", "10001", "US")
		
		assert.False(t, addr1.Equals(addr2))
	})

	t.Run("different cities are not equal", func(t *testing.T) {
		addr1, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr2, _ := valueobject.NewAddress("123 Main St", "Boston", "02101", "US")
		
		assert.False(t, addr1.Equals(addr2))
	})

	t.Run("different postal codes are not equal", func(t *testing.T) {
		addr1, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr2, _ := valueobject.NewAddress("123 Main St", "New York", "10002", "US")
		
		assert.False(t, addr1.Equals(addr2))
	})

	t.Run("different countries are not equal", func(t *testing.T) {
		addr1, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr2, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "CA")
		
		assert.False(t, addr1.Equals(addr2))
	})

	t.Run("different street2 are not equal", func(t *testing.T) {
		addr1, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr1.Street2 = "Apt 4B"
		
		addr2, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr2.Street2 = "Apt 5C"
		
		assert.False(t, addr1.Equals(addr2))
	})

	t.Run("different states are not equal", func(t *testing.T) {
		addr1, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr1.State = "NY"
		
		addr2, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		addr2.State = "NJ"
		
		assert.False(t, addr1.Equals(addr2))
	})
}

func TestAddress_IsEmpty(t *testing.T) {
	t.Run("non-empty address is not empty", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		assert.False(t, addr.IsEmpty())
	})

	t.Run("zero value is empty", func(t *testing.T) {
		var addr valueobject.Address
		
		assert.True(t, addr.IsEmpty())
	})
}

func TestAddress_String(t *testing.T) {
	t.Run("returns single line format", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		str := addr.String()
		assert.Contains(t, str, "123 Main St")
		assert.Contains(t, str, "New York")
		assert.Contains(t, str, "10001")
		assert.Contains(t, str, "US")
	})

	t.Run("SingleLine vs String format", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "New York", "10001", "US")
		
		singleLine := addr.SingleLine()
		multiLine := addr.String()
		
		// SingleLine has commas, no newlines
		assert.Contains(t, singleLine, ",")
		assert.NotContains(t, singleLine, "\n")
		
		// String (multiline) has newlines, no commas
		assert.Contains(t, multiLine, "\n")
	})
}

func TestAddress_Validation(t *testing.T) {
	t.Run("accepts various international addresses", func(t *testing.T) {
		addresses := []struct {
			name       string
			street     string
			city       string
			postalCode string
			country    string
		}{
			{"US", "123 Main St", "New York", "10001", "US"},
			{"UK", "10 Downing Street", "London", "SW1A 2AA", "GB"},
			{"Ukraine", "Khreshchatyk 1", "Kyiv", "01001", "UA"},
			{"Germany", "Unter den Linden 1", "Berlin", "10117", "DE"},
			{"Poland", "Marszałkowska 1", "Warsaw", "00-624", "PL"},
		}

		for _, tc := range addresses {
			t.Run(tc.name, func(t *testing.T) {
				addr, err := valueobject.NewAddress(tc.street, tc.city, tc.postalCode, tc.country)
				require.NoError(t, err)
				assert.Equal(t, tc.country, addr.Country)
			})
		}
	})

	t.Run("handles unicode characters", func(t *testing.T) {
		addr, err := valueobject.NewAddress(
			"Хрещатик 1",
			"Київ",
			"01001",
			"UA",
		)
		
		require.NoError(t, err)
		assert.Equal(t, "Хрещатик 1", addr.Street)
		assert.Equal(t, "Київ", addr.City)
	})
}
