package aggregate

import (
	"testing"

	"github.com/basilex/promenade/internal/contexts/accounting/account"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccount(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tests := []struct {
		name         string
		code         string
		accountName  string
		accountType  AccountType
		currencyCode string
		expectError  error
	}{
		{
			name:         "valid asset account",
			code:         "311",
			accountName:  "Рахунки в банках",
			accountType:  AccountTypeAsset,
			currencyCode: "UAH",
			expectError:  nil,
		},
		{
			name:         "valid liability account",
			code:         "631",
			accountName:  "Розрахунки з постачальниками",
			accountType:  AccountTypeLiability,
			currencyCode: "UAH",
			expectError:  nil,
		},
		{
			name:         "empty code",
			code:         "",
			accountName:  "Test Account",
			accountType:  AccountTypeAsset,
			currencyCode: "UAH",
			expectError:  account.ErrAccountCodeEmpty,
		},
		{
			name:         "empty name",
			code:         "100",
			accountName:  "",
			accountType:  AccountTypeAsset,
			currencyCode: "UAH",
			expectError:  account.ErrAccountNameEmpty,
		},
		{
			name:         "invalid account type",
			code:         "100",
			accountName:  "Test Account",
			accountType:  AccountType("invalid"),
			currencyCode: "UAH",
			expectError:  account.ErrAccountInvalidType,
		},
		{
			name:         "default currency",
			code:         "100",
			accountName:  "Test Account",
			accountType:  AccountTypeAsset,
			currencyCode: "",
			expectError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := NewAccount(orgID, tt.code, tt.accountName, tt.accountType, tt.currencyCode, userID)

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
				assert.Nil(t, acc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, acc)
				assert.Equal(t, orgID, acc.OrganizationID)
				assert.Equal(t, tt.code, acc.Code)
				assert.Equal(t, tt.accountName, acc.Name)
				assert.Equal(t, tt.accountType, acc.Type)
				assert.True(t, acc.IsActive)
				assert.Equal(t, userID, acc.LastUpdatedBy)
				assert.NotEqual(t, uuidv7.Nil, acc.ID)
				assert.Equal(t, 1, acc.Level)
				if tt.currencyCode == "" {
					assert.Equal(t, "UAH", acc.CurrencyCode)
				} else {
					assert.Equal(t, tt.currencyCode, acc.CurrencyCode)
				}
			}
		})
	}
}

func TestNewChildAccount(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parentID := uuidv7.New()

	t.Run("valid child account", func(t *testing.T) {
		acc, err := NewChildAccount(orgID, "311", "Поточні рахунки", AccountTypeAsset, "UAH", parentID, 1, userID)
		require.NoError(t, err)
		require.NotNil(t, acc)
		require.NotNil(t, acc.ParentID)
		assert.Equal(t, parentID, *acc.ParentID)
		assert.Equal(t, 2, acc.Level)
	})

	t.Run("nil parent ID", func(t *testing.T) {
		acc, err := NewChildAccount(orgID, "311", "Поточні рахунки", AccountTypeAsset, "UAH", uuidv7.Nil, 1, userID)
		assert.ErrorIs(t, err, account.ErrAccountParentNotFound)
		assert.Nil(t, acc)
	})
}

func TestAccount_UpdateName(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("update name", func(t *testing.T) {
		acc, err := NewAccount(orgID, "311", "Рахунки в банках", AccountTypeAsset, "UAH", userID)
		require.NoError(t, err)

		newName := "Розрахункові рахунки"
		err = acc.UpdateName(newName)
		require.NoError(t, err)
		assert.Equal(t, newName, acc.Name)
	})

	t.Run("update with empty name", func(t *testing.T) {
		acc, err := NewAccount(orgID, "311", "Рахунки в банках", AccountTypeAsset, "UAH", userID)
		require.NoError(t, err)

		err = acc.UpdateName("")
		assert.ErrorIs(t, err, account.ErrAccountNameEmpty)
	})
}

func TestAccount_ActivateDeactivate(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	acc, err := NewAccount(orgID, "311", "Рахунки в банках", AccountTypeAsset, "UAH", userID)
	require.NoError(t, err)
	assert.True(t, acc.IsActive)

	t.Run("deactivate", func(t *testing.T) {
		err := acc.Deactivate()
		require.NoError(t, err)
		assert.False(t, acc.IsActive)
	})

	t.Run("deactivate already inactive", func(t *testing.T) {
		err := acc.Deactivate()
		require.NoError(t, err) // Returns nil, doesn't error
		assert.False(t, acc.IsActive)
	})

	t.Run("activate", func(t *testing.T) {
		err := acc.Activate()
		require.NoError(t, err)
		assert.True(t, acc.IsActive)
	})

	t.Run("activate already active", func(t *testing.T) {
		err := acc.Activate()
		require.NoError(t, err) // Returns nil, doesn't error
		assert.True(t, acc.IsActive)
	})
}

func TestIsValidAccountType(t *testing.T) {
	validTypes := []AccountType{
		AccountTypeAsset,
		AccountTypeLiability,
		AccountTypeEquity,
		AccountTypeRevenue,
		AccountTypeExpense,
	}

	for _, at := range validTypes {
		t.Run(string(at), func(t *testing.T) {
			assert.True(t, isValidAccountType(at))
		})
	}

	t.Run("invalid type", func(t *testing.T) {
		assert.False(t, isValidAccountType(AccountType("invalid")))
		assert.False(t, isValidAccountType(AccountType("")))
	})
}

func TestAccount_IsAssetOrExpense(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tests := []struct {
		accountType AccountType
		expected    bool
	}{
		{AccountTypeAsset, true},
		{AccountTypeExpense, true},
		{AccountTypeLiability, false},
		{AccountTypeEquity, false},
		{AccountTypeRevenue, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.accountType), func(t *testing.T) {
			acc, err := NewAccount(orgID, "100", "Test", tt.accountType, "UAH", userID)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.IsAssetOrExpense())
		})
	}
}

func TestAccount_IsLiabilityEquityOrRevenue(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	tests := []struct {
		accountType AccountType
		expected    bool
	}{
		{AccountTypeLiability, true},
		{AccountTypeEquity, true},
		{AccountTypeRevenue, true},
		{AccountTypeAsset, false},
		{AccountTypeExpense, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.accountType), func(t *testing.T) {
			acc, err := NewAccount(orgID, "100", "Test", tt.accountType, "UAH", userID)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, acc.IsLiabilityEquityOrRevenue())
		})
	}
}
