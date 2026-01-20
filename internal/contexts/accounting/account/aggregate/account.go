package aggregate

import (
	"github.com/basilex/promenade/internal/contexts/accounting/account"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// AccountType represents the type of account in the chart of accounts
type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"     // Assets (Debits increase)
	AccountTypeLiability AccountType = "liability" // Liabilities (Credits increase)
	AccountTypeEquity    AccountType = "equity"    // Equity (Credits increase)
	AccountTypeRevenue   AccountType = "revenue"   // Revenue (Credits increase)
	AccountTypeExpense   AccountType = "expense"   // Expenses (Debits increase)
)

// Account is an aggregate root representing an account in the chart of accounts
type Account struct {
	aggregate.BaseAggregate

	OrganizationID uuidv7.UUID
	Code           string      // Account code (e.g., "311", "702")
	Name           string      // Account name
	Type           AccountType // Account type

	// Hierarchy
	ParentID *uuidv7.UUID // Parent account for hierarchical structure
	Level    int          // Hierarchy level (1 = top-level, 2 = child, etc.)

	// Currency
	CurrencyCode string // Default currency (e.g., "UAH", "USD")

	// Status
	IsActive      bool
	LastUpdatedBy uuidv7.UUID
}

// NewAccount creates a new account
func NewAccount(organizationID uuidv7.UUID, code, name string, accountType AccountType, currencyCode string, createdBy uuidv7.UUID) (*Account, error) {
	if code == "" {
		return nil, account.ErrAccountCodeEmpty
	}
	if name == "" {
		return nil, account.ErrAccountNameEmpty
	}
	if !isValidAccountType(accountType) {
		return nil, account.ErrAccountInvalidType
	}
	if currencyCode == "" {
		currencyCode = "UAH" // Default to UAH
	}

	return &Account{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		OrganizationID: organizationID,
		Code:           code,
		Name:           name,
		Type:           accountType,
		Level:          1, // Top-level by default
		CurrencyCode:   currencyCode,
		IsActive:       true,
		LastUpdatedBy:  createdBy,
	}, nil
}

// NewChildAccount creates a child account with parent reference
func NewChildAccount(organizationID uuidv7.UUID, code, name string, accountType AccountType, currencyCode string, parentID uuidv7.UUID, parentLevel int, createdBy uuidv7.UUID) (*Account, error) {
	acc, err := NewAccount(organizationID, code, name, accountType, currencyCode, createdBy)
	if err != nil {
		return nil, err
	}

	if parentID == uuidv7.Nil {
		return nil, account.ErrAccountParentNotFound
	}

	acc.ParentID = &parentID
	acc.Level = parentLevel + 1

	return acc, nil
}

// UpdateName updates account name
func (a *Account) UpdateName(name string) error {
	if name == "" {
		return account.ErrAccountNameEmpty
	}
	a.Name = name
	a.Touch()
	return nil
}

// Deactivate marks account as inactive
func (a *Account) Deactivate() error {
	if !a.IsActive {
		return nil // Already inactive
	}
	a.IsActive = false
	a.Touch()
	return nil
}

// Activate marks account as active
func (a *Account) Activate() error {
	if a.IsActive {
		return nil // Already active
	}
	a.IsActive = true
	a.Touch()
	return nil
}

// IsAssetOrExpense returns true if account is Asset or Expense type
// (Debits increase balance, Credits decrease)
func (a *Account) IsAssetOrExpense() bool {
	return a.Type == AccountTypeAsset || a.Type == AccountTypeExpense
}

// IsLiabilityEquityOrRevenue returns true if account is Liability, Equity, or Revenue type
// (Credits increase balance, Debits decrease)
func (a *Account) IsLiabilityEquityOrRevenue() bool {
	return a.Type == AccountTypeLiability || a.Type == AccountTypeEquity || a.Type == AccountTypeRevenue
}

// isValidAccountType validates account type
func isValidAccountType(t AccountType) bool {
	switch t {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue, AccountTypeExpense:
		return true
	default:
		return false
	}
}
