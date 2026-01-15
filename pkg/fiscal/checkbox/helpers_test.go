package checkbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConvertPaymentType tests payment type conversion
func TestConvertPaymentType(t *testing.T) {
	tests := []struct {
		name        string
		paymentType string
		want        string
		wantErr     bool
	}{
		{
			name:        "Cash",
			paymentType: "cash",
			want:        "CASH",
			wantErr:     false,
		},
		{
			name:        "Card",
			paymentType: "card",
			want:        "CARD",
			wantErr:     false,
		},
		{
			name:        "Cashless",
			paymentType: "cashless",
			want:        "CASHLESS",
			wantErr:     false,
		},
		{
			name:        "Invalid",
			paymentType: "bitcoin",
			want:        "",
			wantErr:     true,
		},
		{
			name:        "Empty",
			paymentType: "",
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertPaymentType(tt.paymentType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported payment type")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestConvertToKopiyky tests conversion from float to kopiyky
func TestConvertToKopiyky(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   int
	}{
		{
			name:   "Whole number",
			amount: 100.0,
			want:   10000,
		},
		{
			name:   "With cents",
			amount: 123.45,
			want:   12345,
		},
		{
			name:   "Small amount",
			amount: 0.99,
			want:   99,
		},
		{
			name:   "Zero",
			amount: 0.0,
			want:   0,
		},
		{
			name:   "Large amount",
			amount: 9999.99,
			want:   999999,
		},
		{
			name:   "One kopiyky",
			amount: 0.01,
			want:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToKopiyky(tt.amount)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestConvertFromKopiyky tests conversion from kopiyky to float
func TestConvertFromKopiyky(t *testing.T) {
	tests := []struct {
		name    string
		kopiyky int
		want    float64
	}{
		{
			name:    "Whole number",
			kopiyky: 10000,
			want:    100.0,
		},
		{
			name:    "With cents",
			kopiyky: 12345,
			want:    123.45,
		},
		{
			name:    "Small amount",
			kopiyky: 99,
			want:    0.99,
		},
		{
			name:    "Zero",
			kopiyky: 0,
			want:    0.0,
		},
		{
			name:    "Large amount",
			kopiyky: 999999,
			want:    9999.99,
		},
		{
			name:    "One kopiyky",
			kopiyky: 1,
			want:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertFromKopiyky(tt.kopiyky)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRoundTrip tests conversion round trip
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
	}{
		{name: "100 UAH", amount: 100.0},
		{name: "123.45 UAH", amount: 123.45},
		{name: "0.99 UAH", amount: 0.99},
		{name: "9999.99 UAH", amount: 9999.99},
		{name: "0.01 UAH", amount: 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to kopiyky and back
			kopiyky := ConvertToKopiyky(tt.amount)
			result := ConvertFromKopiyky(kopiyky)

			assert.InDelta(t, tt.amount, result, 0.0001)
		})
	}
}

// TestPaymentTypeValidation tests all valid payment types
func TestPaymentTypeValidation(t *testing.T) {
	validTypes := []string{"cash", "card", "cashless"}

	for _, paymentType := range validTypes {
		t.Run(paymentType, func(t *testing.T) {
			result, err := ConvertPaymentType(paymentType)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

// TestPaymentTypeCase tests case sensitivity
func TestPaymentTypeCase(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{input: "cash", wantErr: false},
		{input: "CASH", wantErr: true},  // Case sensitive
		{input: "Cash", wantErr: true},  // Case sensitive
		{input: "card", wantErr: false},
		{input: "CARD", wantErr: true},
		{input: "Card", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ConvertPaymentType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestKopiykyCents tests precision with cents
func TestKopiykyCents(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		kopiyky int
	}{
		{"99 cents", 0.99, 99},
		{"50 cents", 0.50, 50},
		{"1 cent", 0.01, 1},
		{"25 cents", 0.25, 25},
		{"75 cents", 0.75, 75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToKopiyky(tt.amount)
			assert.Equal(t, tt.kopiyky, got)

			// Reverse conversion
			reverse := ConvertFromKopiyky(tt.kopiyky)
			assert.InDelta(t, tt.amount, reverse, 0.0001)
		})
	}
}

// TestNegativeAmounts tests negative amounts
func TestNegativeAmounts(t *testing.T) {
	t.Run("Negative float to kopiyky", func(t *testing.T) {
		result := ConvertToKopiyky(-100.0)
		assert.Equal(t, -10000, result)
	})

	t.Run("Negative kopiyky to float", func(t *testing.T) {
		result := ConvertFromKopiyky(-10000)
		assert.Equal(t, -100.0, result)
	})
}

// TestLargeAmounts tests large amounts
func TestLargeAmounts(t *testing.T) {
	t.Run("Large amount to kopiyky", func(t *testing.T) {
		result := ConvertToKopiyky(1000000.0)
		assert.Equal(t, 100000000, result)
	})

	t.Run("Large kopiyky to float", func(t *testing.T) {
		result := ConvertFromKopiyky(100000000)
		assert.Equal(t, 1000000.0, result)
	})
}
