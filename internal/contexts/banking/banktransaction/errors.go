package banktransaction

import "errors"

// Domain errors for BankTransaction aggregate
var (
	ErrBankTransactionNotFound = errors.New("bank transaction not found")
	ErrInvalidDirection        = errors.New("invalid transaction direction")
	ErrInvalidAmount           = errors.New("invalid transaction amount")
	ErrAlreadyMatched          = errors.New("transaction already matched")
	ErrInvalidMatchedEntity    = errors.New("invalid matched entity")
	ErrInvalidBankAccountID    = errors.New("bank account ID is required")
	ErrInvalidTransactionDate  = errors.New("transaction date is required")
	ErrInvalidLastUpdatedBy    = errors.New("last updated by is required")
	ErrCannotMatchPendingTx    = errors.New("cannot match pending transaction")
	ErrCannotBookCanceledTx    = errors.New("cannot book canceled transaction")
	ErrExternalIDExists        = errors.New("external ID already exists")
	ErrInvalidCounterparty     = errors.New("invalid counterparty data")
)
