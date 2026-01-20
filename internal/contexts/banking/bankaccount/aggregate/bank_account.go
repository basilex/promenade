package aggregate

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/banking/bankaccount"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// BankProvider represents external bank provider
type BankProvider string

const (
	ProviderManual   BankProvider = "manual"
	ProviderMonobank BankProvider = "monobank"
	ProviderPrivat24 BankProvider = "privat24"
	ProviderPUMB     BankProvider = "pumb"
)

// BankAccountStatus represents account status
type BankAccountStatus string

const (
	BankAccountStatusActive   BankAccountStatus = "active"
	BankAccountStatusInactive BankAccountStatus = "inactive"
	BankAccountStatusArchived BankAccountStatus = "archived"
)

// BankAccount is an aggregate root representing a connected bank account
type BankAccount struct {
	aggregate.BaseAggregate

	OrganizationID    uuidv7.UUID
	Name              string
	BankName          string
	IBAN              string
	AccountNumber     string
	CurrencyCode      string
	Provider          BankProvider
	ProviderAccountID string
	Status            BankAccountStatus
	LastSyncAt        *time.Time
	BalanceCents      int64
	Metadata          jsonstore.Field[map[string]interface{}]
	LastUpdatedBy     uuidv7.UUID
}

// NewBankAccount creates a new manual bank account
func NewBankAccount(
	organizationID uuidv7.UUID,
	name, bankName, currencyCode string,
	lastUpdatedBy uuidv7.UUID,
) (*BankAccount, error) {
	if organizationID == uuidv7.Nil {
		return nil, bankaccount.ErrInvalidOrganizationID
	}
	if name == "" {
		return nil, bankaccount.ErrInvalidAccountName
	}
	if bankName == "" {
		return nil, bankaccount.ErrInvalidBankName
	}
	if currencyCode == "" {
		currencyCode = "UAH"
	}
	if lastUpdatedBy == uuidv7.Nil {
		return nil, bankaccount.ErrInvalidLastUpdatedBy
	}

	return &BankAccount{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		OrganizationID: organizationID,
		Name:           name,
		BankName:       bankName,
		CurrencyCode:   currencyCode,
		Provider:       ProviderManual,
		Status:         BankAccountStatusActive,
		BalanceCents:   0,
		Metadata:       jsonstore.NewField(map[string]interface{}{}),
		LastUpdatedBy:  lastUpdatedBy,
	}, nil
}

// ConnectProvider connects bank account to external provider
func (a *BankAccount) ConnectProvider(
	provider BankProvider,
	providerAccountID string,
	iban, accountNumber string,
) error {
	if provider == ProviderManual {
		return bankaccount.ErrInvalidProvider
	}
	if providerAccountID == "" {
		return bankaccount.ErrInvalidProvider
	}

	a.Provider = provider
	a.ProviderAccountID = providerAccountID
	a.IBAN = iban
	a.AccountNumber = accountNumber
	a.Touch()

	return nil
}

// UpdateBalance updates account balance
func (a *BankAccount) UpdateBalance(balanceCents int64) error {
	a.BalanceCents = balanceCents
	a.Touch()
	return nil
}

// RecordSync records successful synchronization
func (a *BankAccount) RecordSync() error {
	if a.Provider == ProviderManual {
		return bankaccount.ErrManualAccountCannotSync
	}
	if a.Status != BankAccountStatusActive {
		return bankaccount.ErrBankAccountInactive
	}

	now := time.Now()
	a.LastSyncAt = &now
	a.Touch()

	return nil
}

// CanSync checks if account can be synced
func (a *BankAccount) CanSync(minInterval time.Duration) bool {
	if a.Provider == ProviderManual {
		return false
	}
	if a.LastSyncAt == nil {
		return true
	}
	return time.Since(*a.LastSyncAt) >= minInterval
}

// Activate activates an inactive account
func (a *BankAccount) Activate() error {
	if a.Status == BankAccountStatusActive {
		return nil
	}
	if a.Status == BankAccountStatusArchived {
		return bankaccount.ErrArchivedAccountOperation
	}

	a.Status = BankAccountStatusActive
	a.Touch()
	return nil
}

// Deactivate deactivates an active account
func (a *BankAccount) Deactivate() error {
	if a.Status == BankAccountStatusInactive {
		return nil
	}

	a.Status = BankAccountStatusInactive
	a.Touch()
	return nil
}

// Archive archives the account
func (a *BankAccount) Archive() error {
	a.Status = BankAccountStatusArchived
	a.Touch()
	return nil
}

// IsActive checks if account is active
func (a *BankAccount) IsActive() bool {
	return a.Status == BankAccountStatusActive
}

// IsProviderConnected checks if account has provider integration
func (a *BankAccount) IsProviderConnected() bool {
	return a.Provider != ProviderManual && a.ProviderAccountID != ""
}
