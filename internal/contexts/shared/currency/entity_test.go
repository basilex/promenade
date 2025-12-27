package currency

import (
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
)

func TestNewCurrency(t *testing.T) {
currency, err := NewCurrency("USD", "US Dollar", "$", 2)
require.NoError(t, err)
assert.Equal(t, "USD", currency.Code)
assert.Equal(t, 2, currency.DecimalPlaces)
}
