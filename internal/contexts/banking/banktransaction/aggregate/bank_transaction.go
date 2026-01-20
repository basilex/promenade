package aggregate

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/banking/banktransaction"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TransactionDirection represents debit or credit
type TransactionDirection string

const (
	DirectionDebit  TransactionDirection = "debit"
	DirectionCredit TransactionDirection = "credit"
)

// TransactionStatus represents transaction status
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusBooked     TransactionStatus = "booked"
	TransactionStatusCanceled   TransactionStatus = "canceled"
	TransactionStatusReconciled TransactionStatus = "reconciled"
)

// MatchedEntityType represents type of matched entity
type MatchedEntityType string

const (
	MatchedEntityInvoice MatchedEntityType = "invoice"
	MatchedEntityOrder   MatchedEntityType = "order"
	MatchedEntityPayment MatchedEntityType = "payment"
)

// BankTransaction is an aggregate root representing a bank transaction
type BankTransaction struct {
	aggregate.BaseAggregate

	// Account and Statement
	BankAccountID uuidv7.UUID
	StatementID   *uuidv7.UUID
	ExternalID    string // Provider's transaction ID

	// Transaction Details
	Direction        TransactionDirection
	AmountCents      int64
	CurrencyCode     string
	CounterpartyName string
	CounterpartyIBAN string
	Description      string

	// Timestamps
	TransactionAt time.Time
	BookedAt      *time.Time

	// Status
	Status TransactionStatus

	// Matching
	MatchedEntityType *MatchedEntityType
	MatchedEntityID   *uuidv7.UUID
	MatchedAt         *time.Time

	// Raw Data
	RawPayload    jsonstore.Field[map[string]interface{}]
	LastUpdatedBy uuidv7.UUID
}

// NewBankTransaction creates a new bank transaction
func NewBankTransaction(
	bankAccountID uuidv7.UUID,
	direction TransactionDirection,
	amountCents int64,
	currencyCode string,
	transactionAt time.Time,
	description string,
	lastUpdatedBy uuidv7.UUID,
) (*BankTransaction, error) {
	if bankAccountID == uuidv7.Nil {
		return nil, banktransaction.ErrInvalidBankAccountID
	}
	if direction != DirectionDebit && direction != DirectionCredit {
		return nil, banktransaction.ErrInvalidDirection
	}
	if amountCents <= 0 {
		return nil, banktransaction.ErrInvalidAmount
	}
	if currencyCode == "" {
		currencyCode = "UAH"
	}
	if lastUpdatedBy == uuidv7.Nil {
		return nil, banktransaction.ErrInvalidLastUpdatedBy
	}

	return &BankTransaction{
		BaseAggregate: aggregate.NewBaseAggregate(),
		BankAccountID: bankAccountID,
		Direction:     direction,
		AmountCents:   amountCents,
		CurrencyCode:  currencyCode,
		TransactionAt: transactionAt,
		Description:   description,
		Status:        TransactionStatusPending,
		RawPayload:    jsonstore.NewField(map[string]interface{}{}),
		LastUpdatedBy: lastUpdatedBy,
	}, nil
}

// SetExternalID sets the external transaction ID from provider
func (t *BankTransaction) SetExternalID(externalID string) {
	t.ExternalID = externalID
	t.Touch()
}

// SetCounterparty sets counterparty information
func (t *BankTransaction) SetCounterparty(name, iban string) {
	t.CounterpartyName = name
	t.CounterpartyIBAN = iban
	t.Touch()
}

// Book marks transaction as booked
func (t *BankTransaction) Book() error {
	if t.Status == TransactionStatusBooked || t.Status == TransactionStatusReconciled {
		return nil
	}

	t.Status = TransactionStatusBooked
	now := time.Now()
	t.BookedAt = &now
	t.Touch()
	return nil
}

// Cancel marks transaction as canceled
func (t *BankTransaction) Cancel() error {
	if t.Status == TransactionStatusCanceled {
		return nil
	}
	if t.Status == TransactionStatusReconciled {
		return banktransaction.ErrCannotBookCanceledTx
	}

	t.Status = TransactionStatusCanceled
	t.Touch()
	return nil
}

// Match matches transaction to an entity (invoice, order, payment)
func (t *BankTransaction) Match(entityType MatchedEntityType, entityID uuidv7.UUID) error {
	if t.IsMatched() {
		return banktransaction.ErrAlreadyMatched
	}
	if entityID == uuidv7.Nil {
		return banktransaction.ErrInvalidMatchedEntity
	}

	t.MatchedEntityType = &entityType
	t.MatchedEntityID = &entityID
	now := time.Now()
	t.MatchedAt = &now
	t.Status = TransactionStatusReconciled
	t.Touch()
	return nil
}

// Unmatch removes matching from transaction
func (t *BankTransaction) Unmatch() error {
	if !t.IsMatched() {
		return banktransaction.ErrBankTransactionNotFound
	}

	t.MatchedEntityType = nil
	t.MatchedEntityID = nil
	t.MatchedAt = nil
	t.Status = TransactionStatusBooked
	t.Touch()
	return nil
}

// SetRawPayload sets raw provider payload
func (t *BankTransaction) SetRawPayload(payload map[string]interface{}) {
	t.RawPayload = jsonstore.NewField(payload)
	t.Touch()
}

// LinkToStatement links transaction to statement
func (t *BankTransaction) LinkToStatement(statementID uuidv7.UUID) error {
	if statementID == uuidv7.Nil {
		return banktransaction.ErrInvalidBankAccountID
	}

	t.StatementID = &statementID
	t.Touch()
	return nil
}

// IsMatched checks if transaction is matched to an entity
func (t *BankTransaction) IsMatched() bool {
	return t.MatchedEntityType != nil && t.MatchedEntityID != nil
}

// IsReconciled checks if transaction is reconciled
func (t *BankTransaction) IsReconciled() bool {
	return t.Status == TransactionStatusReconciled
}

// CanBeMatched checks if transaction can be matched
func (t *BankTransaction) CanBeMatched() bool {
	return t.Status == TransactionStatusBooked && !t.IsMatched()
}

// GetMatchInfo returns matching information
func (t *BankTransaction) GetMatchInfo() (entityType MatchedEntityType, entityID uuidv7.UUID, matchedAt time.Time, matched bool) {
	if !t.IsMatched() {
		return "", uuidv7.Nil, time.Time{}, false
	}
	return *t.MatchedEntityType, *t.MatchedEntityID, *t.MatchedAt, true
}
