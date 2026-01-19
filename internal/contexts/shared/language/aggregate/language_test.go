package aggregate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLanguage(t *testing.T) {
	language, err := NewLanguage("en", "English", "English")
	require.NoError(t, err)
	assert.Equal(t, "en", language.Code)
	assert.Equal(t, "English", language.NativeName)
}
