package aggregate

import (
	"testing"

	"github.com/basilex/promenade/internal/contexts/accounting/taxcode"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaxCode(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tests := []struct {
		name        string
		code        string
		taxName     string
		taxType     TaxType
		rate        int
		expectError error
	}{
		{
			name:        "valid VAT tax 20%",
			code:        "VAT20",
			taxName:     "VAT 20%",
			taxType:     TaxTypeVAT,
			rate:        2000,
			expectError: nil,
		},
		{
			name:        "valid payroll tax 10%",
			code:        "PT10",
			taxName:     "Payroll Tax 10%",
			taxType:     TaxTypePayrollTax,
			rate:        1000,
			expectError: nil,
		},
		{
			name:        "valid income tax 15%",
			code:        "IT15",
			taxName:     "Income Tax 15%",
			taxType:     TaxTypeIncomeTax,
			rate:        1500,
			expectError: nil,
		},
		{
			name:        "empty code",
			code:        "",
			taxName:     "Test Tax",
			taxType:     TaxTypeVAT,
			rate:        2000,
			expectError: taxcode.ErrTaxCodeEmpty,
		},
		{
			name:        "empty name",
			code:        "VAT20",
			taxName:     "",
			taxType:     TaxTypeVAT,
			rate:        2000,
			expectError: taxcode.ErrTaxNameEmpty,
		},
		{
			name:        "invalid tax type",
			code:        "TAX",
			taxName:     "Test Tax",
			taxType:     TaxType("invalid"),
			rate:        2000,
			expectError: taxcode.ErrInvalidTaxType,
		},
		{
			name:        "negative rate",
			code:        "VAT",
			taxName:     "Negative Tax",
			taxType:     TaxTypeVAT,
			rate:        -500,
			expectError: taxcode.ErrInvalidTaxRate,
		},
		{
			name:        "rate over 10000",
			code:        "VAT",
			taxName:     "Over Tax",
			taxType:     TaxTypeVAT,
			rate:        15000,
			expectError: taxcode.ErrInvalidTaxRate,
		},
		{
			name:        "zero rate valid",
			code:        "VAT0",
			taxName:     "Zero VAT",
			taxType:     TaxTypeVAT,
			rate:        0,
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc, err := NewTaxCode(orgID, tt.code, tt.taxName, tt.taxType, tt.rate, userID)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, tc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tc)
				assert.Equal(t, orgID, tc.OrganizationID)
				assert.Equal(t, tt.code, tc.Code)
				assert.Equal(t, tt.taxName, tc.Name)
				assert.Equal(t, tt.taxType, tc.TaxType)
				assert.Equal(t, tt.rate, tc.Rate)
				assert.True(t, tc.IsActive)
			}
		})
	}
}

func TestTaxCode_CalculateTax(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tc, err := NewTaxCode(orgID, "VAT20", "VAT 20%", TaxTypeVAT, 2000, userID)
	require.NoError(t, err)

	tests := []struct {
		name              string
		netAmount         int64
		expectedTaxAmount int64
	}{
		{
			name:              "calculate 20% on 10000 cents",
			netAmount:         10000,
			expectedTaxAmount: 2000,
		},
		{
			name:              "calculate 20% on 25050 cents",
			netAmount:         25050,
			expectedTaxAmount: 5010,
		},
		{
			name:              "calculate on zero",
			netAmount:         0,
			expectedTaxAmount: 0,
		},
		{
			name:              "calculate on 100000 cents",
			netAmount:         100000,
			expectedTaxAmount: 20000,
		},
		{
			name:              "calculate on small amount",
			netAmount:         100,
			expectedTaxAmount: 20,
		},
		{
			name:              "calculate negative (refund)",
			netAmount:         -10000,
			expectedTaxAmount: -2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxAmount := tc.CalculateTax(tt.netAmount)
			assert.Equal(t, tt.expectedTaxAmount, taxAmount)
		})
	}
}

func TestTaxCode_CalculateTaxableBase(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tc, err := NewTaxCode(orgID, "VAT20", "VAT 20%", TaxTypeVAT, 2000, userID)
	require.NoError(t, err)

	tests := []struct {
		name                string
		grossAmount         int64
		expectedTaxableBase int64
	}{
		{
			name:                "gross 12000 cents -> base 10000 cents",
			grossAmount:         12000,
			expectedTaxableBase: 10000,
		},
		{
			name:                "gross 30060 cents -> base 25050 cents",
			grossAmount:         30060,
			expectedTaxableBase: 25050,
		},
		{
			name:                "gross 0 -> base 0",
			grossAmount:         0,
			expectedTaxableBase: 0,
		},
		{
			name:                "gross 120000 cents -> base 100000 cents",
			grossAmount:         120000,
			expectedTaxableBase: 100000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tc.CalculateTaxableBase(tt.grossAmount)
			assert.Equal(t, tt.expectedTaxableBase, base)
		})
	}
}

func TestTaxCode_CalculateTaxFromGross(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tc, err := NewTaxCode(orgID, "VAT20", "VAT 20%", TaxTypeVAT, 2000, userID)
	require.NoError(t, err)

	tests := []struct {
		name              string
		grossAmount       int64
		expectedTaxAmount int64
	}{
		{
			name:              "gross 12000 cents -> tax 2000 cents",
			grossAmount:       12000,
			expectedTaxAmount: 2000,
		},
		{
			name:              "gross 30060 cents -> tax 5010 cents",
			grossAmount:       30060,
			expectedTaxAmount: 5010,
		},
		{
			name:              "gross 120000 cents -> tax 20000 cents",
			grossAmount:       120000,
			expectedTaxAmount: 20000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxAmount := tc.CalculateTaxFromGross(tt.grossAmount)
			assert.Equal(t, tt.expectedTaxAmount, taxAmount)
		})
	}
}

func TestTaxCode_UpdateRate(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tc, err := NewTaxCode(orgID, "VAT20", "VAT 20%", TaxTypeVAT, 2000, userID)
	require.NoError(t, err)

	t.Run("update to valid rate", func(t *testing.T) {
		err := tc.UpdateRate(1500)
		require.NoError(t, err)
		assert.Equal(t, 1500, tc.Rate)
	})

	t.Run("update to negative rate", func(t *testing.T) {
		err := tc.UpdateRate(-500)
		assert.ErrorIs(t, err, taxcode.ErrInvalidTaxRate)
	})

	t.Run("update to rate over 10000", func(t *testing.T) {
		err := tc.UpdateRate(15000)
		assert.ErrorIs(t, err, taxcode.ErrInvalidTaxRate)
	})

	t.Run("update to zero rate", func(t *testing.T) {
		err := tc.UpdateRate(0)
		require.NoError(t, err)
		assert.Equal(t, 0, tc.Rate)
	})
}

func TestTaxCode_SetTaxPayableAccount(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	tc, err := NewTaxCode(orgID, "VAT20", "VAT 20%", TaxTypeVAT, 2000, userID)
	require.NoError(t, err)
	assert.Nil(t, tc.TaxPayableAccountID)

	t.Run("set tax payable account", func(t *testing.T) {
		err := tc.SetTaxPayableAccount(accountID)
		require.NoError(t, err)
		assert.NotNil(t, tc.TaxPayableAccountID)
		assert.Equal(t, accountID, *tc.TaxPayableAccountID)
	})

	t.Run("update tax payable account", func(t *testing.T) {
		newAccountID := uuidv7.New()
		err := tc.SetTaxPayableAccount(newAccountID)
		require.NoError(t, err)
		assert.Equal(t, newAccountID, *tc.TaxPayableAccountID)
	})

	t.Run("nil account ID", func(t *testing.T) {
		err := tc.SetTaxPayableAccount(uuidv7.Nil)
		assert.ErrorIs(t, err, taxcode.ErrGLAccountRequired)
	})
}

func TestTaxCode_SetTaxReceivableAccount(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	tc, err := NewTaxCode(orgID, "VAT20", "VAT 20%", TaxTypeVAT, 2000, userID)
	require.NoError(t, err)
	assert.Nil(t, tc.TaxReceivableAccountID)

	t.Run("set tax receivable account", func(t *testing.T) {
		err := tc.SetTaxReceivableAccount(accountID)
		require.NoError(t, err)
		assert.NotNil(t, tc.TaxReceivableAccountID)
		assert.Equal(t, accountID, *tc.TaxReceivableAccountID)
	})

	t.Run("update tax receivable account", func(t *testing.T) {
		newAccountID := uuidv7.New()
		err := tc.SetTaxReceivableAccount(newAccountID)
		require.NoError(t, err)
		assert.Equal(t, newAccountID, *tc.TaxReceivableAccountID)
	})

	t.Run("nil account ID", func(t *testing.T) {
		err := tc.SetTaxReceivableAccount(uuidv7.Nil)
		assert.ErrorIs(t, err, taxcode.ErrGLAccountRequired)
	})
}

func TestTaxCode_ActivateDeactivate(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tc, err := NewTaxCode(orgID, "VAT20", "VAT 20%", TaxTypeVAT, 2000, userID)
	require.NoError(t, err)
	require.True(t, tc.IsActive)

	t.Run("deactivate", func(t *testing.T) {
		tc.Deactivate()
		assert.False(t, tc.IsActive)
	})

	t.Run("deactivate already inactive", func(t *testing.T) {
		tc.Deactivate()
		assert.False(t, tc.IsActive)
	})

	t.Run("activate", func(t *testing.T) {
		tc.Activate()
		assert.True(t, tc.IsActive)
	})

	t.Run("activate already active", func(t *testing.T) {
		tc.Activate()
		assert.True(t, tc.IsActive)
	})
}

func TestTaxCode_isValidTaxType(t *testing.T) {
	tests := []struct {
		name     string
		taxType  TaxType
		expected bool
	}{
		{
			name:     "valid VAT",
			taxType:  TaxTypeVAT,
			expected: true,
		},
		{
			name:     "valid income tax",
			taxType:  TaxTypeIncomeTax,
			expected: true,
		},
		{
			name:     "valid payroll tax",
			taxType:  TaxTypePayrollTax,
			expected: true,
		},
		{
			name:     "valid withholding",
			taxType:  TaxTypeWithholding,
			expected: true,
		},
		{
			name:     "invalid type",
			taxType:  TaxType("invalid"),
			expected: false,
		},
		{
			name:     "empty type",
			taxType:  TaxType(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTaxType(tt.taxType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
