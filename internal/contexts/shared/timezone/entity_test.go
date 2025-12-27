package timezone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimezone(t *testing.T) {
	t.Run("valid timezone", func(t *testing.T) {
		timezone, err := NewTimezone("America/New_York", "EST", -18000) // -05:00 = -18000 seconds
		require.NoError(t, err)
		assert.Equal(t, "America/New_York", timezone.Name)
		assert.Equal(t, "EST", timezone.Abbreviation)
		assert.Equal(t, -18000, timezone.UTCOffset)
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := NewTimezone("", "EST", -18000)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("empty abbreviation", func(t *testing.T) {
		_, err := NewTimezone("America/New_York", "", -18000)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "abbreviation")
	})

	t.Run("extreme positive offset", func(t *testing.T) {
		timezone, err := NewTimezone("Pacific/Kiritimati", "LINT", 50400) // +14:00
		require.NoError(t, err)
		assert.Equal(t, 50400, timezone.UTCOffset)
	})

	t.Run("extreme negative offset", func(t *testing.T) {
		timezone, err := NewTimezone("Pacific/Midway", "SST", -39600) // -11:00
		require.NoError(t, err)
		assert.Equal(t, -39600, timezone.UTCOffset)
	})
}

func TestTimezone_Validate(t *testing.T) {
	t.Run("valid timezone", func(t *testing.T) {
		timezone, _ := NewTimezone("Europe/London", "GMT", 0)
		err := timezone.Validate()
		assert.NoError(t, err)
	})

	t.Run("whitespace name", func(t *testing.T) {
		timezone, _ := NewTimezone("Europe/London", "GMT", 0)
		timezone.Name = "   "
		err := timezone.Validate()
		assert.Error(t, err)
	})

	t.Run("whitespace abbreviation", func(t *testing.T) {
		timezone, _ := NewTimezone("Europe/London", "GMT", 0)
		timezone.Abbreviation = "   "
		err := timezone.Validate()
		assert.Error(t, err)
	})
}
