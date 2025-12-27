package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewMoney(t *testing.T) {
	t.Run("creates valid money", func(t *testing.T) {
		money, err := valueobject.NewMoney(1050, "USD")

		require.NoError(t, err)
		assert.Equal(t, int64(1050), money.Amount)
		assert.Equal(t, "USD", money.Currency)
	})

	t.Run("accepts zero amount", func(t *testing.T) {
		money, err := valueobject.NewMoney(0, "USD")

		require.NoError(t, err)
		assert.Equal(t, int64(0), money.Amount)
	})

	t.Run("accepts negative amount", func(t *testing.T) {
		money, err := valueobject.NewMoney(-1000, "USD")

		require.NoError(t, err)
		assert.Equal(t, int64(-1000), money.Amount)
	})

	t.Run("rejects empty currency", func(t *testing.T) {
		_, err := valueobject.NewMoney(1000, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "currency is required")
	})

	t.Run("rejects invalid currency code length", func(t *testing.T) {
		_, err := valueobject.NewMoney(1000, "INVALID")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "3-letter")
	})

	t.Run("accepts all major currencies", func(t *testing.T) {
		currencies := []string{"USD", "EUR", "UAH", "GBP", "JPY", "CNY", "CHF", "CAD", "AUD", "PLN"}

		for _, curr := range currencies {
			t.Run(curr, func(t *testing.T) {
				money, err := valueobject.NewMoney(1000, curr)
				require.NoError(t, err)
				assert.Equal(t, curr, money.Currency)
			})
		}
	})
}

func TestFromFloat(t *testing.T) {
	t.Run("converts float to cents", func(t *testing.T) {
		money, err := valueobject.FromFloat(10.50, "USD")

		require.NoError(t, err)
		assert.Equal(t, int64(1050), money.Amount)
	})

	t.Run("handles zero", func(t *testing.T) {
		money, err := valueobject.FromFloat(0.0, "USD")

		require.NoError(t, err)
		assert.Equal(t, int64(0), money.Amount)
	})

	t.Run("handles negative amounts", func(t *testing.T) {
		money, err := valueobject.FromFloat(-25.99, "EUR")

		require.NoError(t, err)
		assert.Equal(t, int64(-2599), money.Amount)
	})

	t.Run("rounds correctly", func(t *testing.T) {
		money, err := valueobject.FromFloat(10.555, "USD")

		require.NoError(t, err)
		assert.Equal(t, int64(1056), money.Amount) // rounds to nearest cent
	})

	t.Run("handles large amounts", func(t *testing.T) {
		money, err := valueobject.FromFloat(1000000.00, "USD")

		require.NoError(t, err)
		assert.Equal(t, int64(100000000), money.Amount)
	})
}

func TestMoney_Add(t *testing.T) {
	t.Run("adds same currency", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(500, "USD")

		result, err := m1.Add(m2)

		require.NoError(t, err)
		assert.Equal(t, int64(1500), result.Amount)
		assert.Equal(t, "USD", result.Currency)
	})

	t.Run("rejects different currencies", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(500, "EUR")

		_, err := m1.Add(m2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "different currencies")
	})

	t.Run("handles negative amounts", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(-300, "USD")

		result, err := m1.Add(m2)

		require.NoError(t, err)
		assert.Equal(t, int64(700), result.Amount)
	})

	t.Run("handles zero", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(0, "USD")

		result, err := m1.Add(m2)

		require.NoError(t, err)
		assert.Equal(t, int64(1000), result.Amount)
	})
}

func TestMoney_Subtract(t *testing.T) {
	t.Run("subtracts same currency", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(300, "USD")

		result, err := m1.Subtract(m2)

		require.NoError(t, err)
		assert.Equal(t, int64(700), result.Amount)
	})

	t.Run("rejects different currencies", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(300, "EUR")

		_, err := m1.Subtract(m2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "different currencies")
	})

	t.Run("can result in negative", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(500, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		result, err := m1.Subtract(m2)

		require.NoError(t, err)
		assert.Equal(t, int64(-500), result.Amount)
	})
}

func TestMoney_Multiply(t *testing.T) {
	t.Run("multiplies by positive factor", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1000, "USD")

		result := money.Multiply(2.5)

		assert.Equal(t, int64(2500), result.Amount)
		assert.Equal(t, "USD", result.Currency)
	})

	t.Run("multiplies by zero", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1000, "USD")

		result := money.Multiply(0)

		assert.Equal(t, int64(0), result.Amount)
	})

	t.Run("multiplies by negative factor", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1000, "USD")

		result := money.Multiply(-1.5)

		assert.Equal(t, int64(-1500), result.Amount)
	})

	t.Run("rounds correctly", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1000, "USD")

		result := money.Multiply(1.555)

		// math.Round(1000 * 1.555) = math.Round(1555.0) = 1555
		assert.Equal(t, int64(1555), result.Amount)
	})

	t.Run("handles fractional amounts", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1000, "USD")

		result := money.Multiply(0.15)

		assert.Equal(t, int64(150), result.Amount)
	})
}

func TestMoney_Compare(t *testing.T) {
	t.Run("less than", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(500, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		cmp, err := m1.Compare(m2)

		require.NoError(t, err)
		assert.Equal(t, -1, cmp)
	})

	t.Run("greater than", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1500, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		cmp, err := m1.Compare(m2)

		require.NoError(t, err)
		assert.Equal(t, 1, cmp)
	})

	t.Run("equal", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		cmp, err := m1.Compare(m2)

		require.NoError(t, err)
		assert.Equal(t, 0, cmp)
	})

	t.Run("rejects different currencies", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(1000, "EUR")

		_, err := m1.Compare(m2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "different currencies")
	})
}

func TestMoney_LessThan(t *testing.T) {
	t.Run("returns true when less", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(500, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		result, err := m1.LessThan(m2)

		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("returns false when equal", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		result, err := m1.LessThan(m2)

		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("returns false when greater", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1500, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		result, err := m1.LessThan(m2)

		require.NoError(t, err)
		assert.False(t, result)
	})
}

func TestMoney_GreaterThan(t *testing.T) {
	t.Run("returns true when greater", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1500, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		result, err := m1.GreaterThan(m2)

		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("returns false when equal", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		result, err := m1.GreaterThan(m2)

		require.NoError(t, err)
		assert.False(t, result)
	})

	t.Run("returns false when less", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(500, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		result, err := m1.GreaterThan(m2)

		require.NoError(t, err)
		assert.False(t, result)
	})
}

func TestMoney_Equals(t *testing.T) {
	t.Run("equal money is equal", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(1000, "USD")

		assert.True(t, m1.Equals(m2))
	})

	t.Run("different amounts are not equal", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(500, "USD")

		assert.False(t, m1.Equals(m2))
	})

	t.Run("different currencies are not equal", func(t *testing.T) {
		m1, _ := valueobject.NewMoney(1000, "USD")
		m2, _ := valueobject.NewMoney(1000, "EUR")

		assert.False(t, m1.Equals(m2))
	})
}

func TestMoney_IsZero(t *testing.T) {
	t.Run("zero amount is zero", func(t *testing.T) {
		money, _ := valueobject.NewMoney(0, "USD")

		assert.True(t, money.IsZero())
	})

	t.Run("positive amount is not zero", func(t *testing.T) {
		money, _ := valueobject.NewMoney(100, "USD")

		assert.False(t, money.IsZero())
	})

	t.Run("negative amount is not zero", func(t *testing.T) {
		money, _ := valueobject.NewMoney(-100, "USD")

		assert.False(t, money.IsZero())
	})
}

func TestMoney_IsPositive(t *testing.T) {
	t.Run("positive amount is positive", func(t *testing.T) {
		money, _ := valueobject.NewMoney(100, "USD")

		assert.True(t, money.IsPositive())
	})

	t.Run("zero is not positive", func(t *testing.T) {
		money, _ := valueobject.NewMoney(0, "USD")

		assert.False(t, money.IsPositive())
	})

	t.Run("negative amount is not positive", func(t *testing.T) {
		money, _ := valueobject.NewMoney(-100, "USD")

		assert.False(t, money.IsPositive())
	})
}

func TestMoney_IsNegative(t *testing.T) {
	t.Run("negative amount is negative", func(t *testing.T) {
		money, _ := valueobject.NewMoney(-100, "USD")

		assert.True(t, money.IsNegative())
	})

	t.Run("zero is not negative", func(t *testing.T) {
		money, _ := valueobject.NewMoney(0, "USD")

		assert.False(t, money.IsNegative())
	})

	t.Run("positive amount is not negative", func(t *testing.T) {
		money, _ := valueobject.NewMoney(100, "USD")

		assert.False(t, money.IsNegative())
	})
}

func TestMoney_ToFloat(t *testing.T) {
	t.Run("converts cents to float", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1050, "USD")

		assert.Equal(t, 10.50, money.ToFloat())
	})

	t.Run("handles zero", func(t *testing.T) {
		money, _ := valueobject.NewMoney(0, "USD")

		assert.Equal(t, 0.0, money.ToFloat())
	})

	t.Run("handles negative", func(t *testing.T) {
		money, _ := valueobject.NewMoney(-2599, "EUR")

		assert.Equal(t, -25.99, money.ToFloat())
	})

	t.Run("handles large amounts", func(t *testing.T) {
		money, _ := valueobject.NewMoney(100000000, "USD")

		assert.Equal(t, 1000000.00, money.ToFloat())
	})
}

func TestMoney_Format(t *testing.T) {
	t.Run("formats USD", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1050, "USD")

		formatted := money.Format()
		assert.Equal(t, "$10.50", formatted)
	})

	t.Run("formats EUR", func(t *testing.T) {
		money, _ := valueobject.NewMoney(2599, "EUR")

		formatted := money.Format()
		assert.Equal(t, "€25.99", formatted)
	})

	t.Run("formats UAH", func(t *testing.T) {
		money, _ := valueobject.NewMoney(25000, "UAH")

		formatted := money.Format()
		assert.Equal(t, "₴250.00", formatted)
	})

	t.Run("formats negative amounts", func(t *testing.T) {
		money, _ := valueobject.NewMoney(-1050, "USD")

		formatted := money.Format()
		assert.Equal(t, "-$10.50", formatted)
	})

	t.Run("formats zero", func(t *testing.T) {
		money, _ := valueobject.NewMoney(0, "USD")

		formatted := money.Format()
		assert.Equal(t, "$0.00", formatted)
	})

	t.Run("formats unknown currency with code", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1000, "PLN")

		formatted := money.Format()
		assert.Contains(t, formatted, "PLN")
		assert.Contains(t, formatted, "10.00")
	})
}

func TestMoney_String(t *testing.T) {
	t.Run("returns formatted string", func(t *testing.T) {
		money, _ := valueobject.NewMoney(1050, "USD")

		// String() returns "USD 10.50" format
		str := money.String()
		assert.Contains(t, str, "USD")
		assert.Contains(t, str, "10.50")
	})

	t.Run("Format uses currency symbols", func(t *testing.T) {
		money, _ := valueobject.NewMoney(2599, "EUR")

		// Format() returns "€25.99" with symbol
		formatted := money.Format()
		assert.Equal(t, "€25.99", formatted)

		// String() returns "EUR 25.99" with code
		str := money.String()
		assert.Contains(t, str, "EUR")
		assert.Contains(t, str, "25.99")
	})
}
