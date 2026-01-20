package bankaccount

import "errors"

// Domain errors for BankAccount aggregate
var (
	ErrBankAccountNotFound      = errors.New("bank account not found")
	ErrBankAccountInactive      = errors.New("bank account is inactive")
	ErrBankAccountCodeExists    = errors.New("bank account code already exists")
	ErrInvalidProvider          = errors.New("invalid bank provider")
	ErrProviderAlreadyConnected = errors.New("provider already connected")
	ErrManualAccountCannotSync  = errors.New("manual accounts cannot be synced")
	ErrArchivedAccountOperation = errors.New("cannot operate on archived account")
	ErrInvalidCurrency          = errors.New("invalid currency code")
	ErrInvalidBankName          = errors.New("bank name is required")
	ErrInvalidAccountName       = errors.New("account name is required")
	ErrInvalidOrganizationID    = errors.New("organization ID is required")
	ErrInvalidLastUpdatedBy     = errors.New("last updated by is required")
)
