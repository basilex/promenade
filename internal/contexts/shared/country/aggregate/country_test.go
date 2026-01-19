package aggregate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCountry(t *testing.T) {
	country, err := NewCountry("UA", "Ukraine", "+380")
	require.NoError(t, err)
	assert.Equal(t, "UA", country.Code)
}
