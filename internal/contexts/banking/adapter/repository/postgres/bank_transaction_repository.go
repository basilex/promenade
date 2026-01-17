package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/banking/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// BankTransactionRepository implements bank transaction persistence
type BankTransactionRepository struct {
	*BaseRepository
}

// NewBankTransactionRepository creates a new bank transaction repository
func NewBankTransactionRepository(db *sqlx.DB) *BankTransactionRepository {
	return &BankTransactionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create inserts a new bank transaction
func (r *BankTransactionRepository) Create(ctx context.Context, tx *aggregate.BankTransaction) error {
	query := `
		INSERT INTO banking_bank_transactions (
			id, version, bank_account_id, external_id, direction, amount_cents, currency_code,
			counterparty_name, counterparty_iban, description, transaction_at, status,
			raw_payload_json, last_updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	return r.Exec(ctx, query,
		tx.GetID().String(),
		tx.GetVersion(),
		tx.BankAccountID.String(),
		tx.ExternalID,
		string(tx.Direction),
		tx.AmountCents,
		tx.CurrencyCode,
		tx.CounterpartyName,
		tx.CounterpartyIBAN,
		tx.Description,
		tx.TransactionAt,
		string(tx.Status),
		func() string { v, _ := tx.RawPayload.Value(); return v.(string) }(),
		tx.LastUpdatedBy.String(),
		tx.GetCreatedAt(),
		tx.GetUpdatedAt(),
	)
}

// GetByID retrieves transaction by ID
func (r *BankTransactionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.BankTransaction, error) {
	query := `
		SELECT id, version, bank_account_id, external_id, direction, amount_cents, currency_code,
			counterparty_name, counterparty_iban, description, transaction_at, booked_at, status,
			matched_entity_type, matched_entity_id, matched_at,
			raw_payload_json, last_updated_by, created_at, updated_at
		FROM banking_bank_transactions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID              string
		version           int
		bankAccountID     string
		externalID        string
		direction         string
		amountCents       int64
		currencyCode      string
		counterpartyName  string
		counterpartyIBAN  string
		description       string
		transactionAt     time.Time
		bookedAt          *time.Time
		status            string
		matchedEntityType *string
		matchedEntityID   *string
		matchedAt         *time.Time
		rawPayloadJSON    string
		lastUpdatedBy     string
		createdAt         time.Time
		updatedAt         time.Time
	)

	if err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&dbID, &version, &bankAccountID, &externalID, &direction, &amountCents, &currencyCode,
		&counterpartyName, &counterpartyIBAN, &description, &transactionAt, &bookedAt, &status,
		&matchedEntityType, &matchedEntityID, &matchedAt,
		&rawPayloadJSON, &lastUpdatedBy, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, aggregate.ErrBankTransactionNotFound
		}
		return nil, err
	}

	txID, _ := uuidv7.Parse(dbID)
	accountID, _ := uuidv7.Parse(bankAccountID)
	updatedByID, _ := uuidv7.Parse(lastUpdatedBy)

	transaction := &aggregate.BankTransaction{
		BankAccountID:    accountID,
		ExternalID:       externalID,
		Direction:        aggregate.TransactionDirection(direction),
		AmountCents:      amountCents,
		CurrencyCode:     currencyCode,
		CounterpartyName: counterpartyName,
		CounterpartyIBAN: counterpartyIBAN,
		Description:      description,
		TransactionAt:    transactionAt,
		BookedAt:         bookedAt,
		Status:           aggregate.TransactionStatus(status),
		LastUpdatedBy:    updatedByID,
	}

	transaction.ID = txID
	transaction.Version = version
	transaction.CreatedAt = createdAt
	transaction.UpdatedAt = updatedAt

	if matchedEntityType != nil {
		et := aggregate.MatchedEntityType(*matchedEntityType)
		transaction.MatchedEntityType = &et
	}
	if matchedEntityID != nil {
		eid, _ := uuidv7.Parse(*matchedEntityID)
		transaction.MatchedEntityID = &eid
	}
	transaction.MatchedAt = matchedAt

	_ = transaction.RawPayload.UnmarshalJSON([]byte(rawPayloadJSON))

	return transaction, nil
}

// ListByAccount lists transactions for bank account
func (r *BankTransactionRepository) ListByAccount(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error) {
	query := `
		SELECT id, version, bank_account_id, external_id, direction, amount_cents, currency_code,
			counterparty_name, counterparty_iban, description, transaction_at, booked_at, status,
			matched_entity_type, matched_entity_id, matched_at,
			raw_payload_json, last_updated_by, created_at, updated_at
		FROM banking_bank_transactions
		WHERE bank_account_id = $1 AND deleted_at IS NULL
		ORDER BY transaction_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, accountID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*aggregate.BankTransaction
	for rows.Next() {
		var (
			dbID              string
			version           int
			bankAccountIDStr  string
			externalID        string
			direction         string
			amountCents       int64
			currencyCode      string
			counterpartyName  string
			counterpartyIBAN  string
			description       string
			transactionAt     time.Time
			bookedAt          *time.Time
			status            string
			matchedEntityType *string
			matchedEntityID   *string
			matchedAt         *time.Time
			rawPayloadJSON    string
			lastUpdatedBy     string
			createdAt         time.Time
			updatedAt         time.Time
		)

		if err := rows.Scan(
			&dbID, &version, &bankAccountIDStr, &externalID, &direction, &amountCents, &currencyCode,
			&counterpartyName, &counterpartyIBAN, &description, &transactionAt, &bookedAt, &status,
			&matchedEntityType, &matchedEntityID, &matchedAt,
			&rawPayloadJSON, &lastUpdatedBy, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}

		txID, _ := uuidv7.Parse(dbID)
		accID, _ := uuidv7.Parse(bankAccountIDStr)
		updatedByID, _ := uuidv7.Parse(lastUpdatedBy)

		transaction := &aggregate.BankTransaction{
			BankAccountID:    accID,
			ExternalID:       externalID,
			Direction:        aggregate.TransactionDirection(direction),
			AmountCents:      amountCents,
			CurrencyCode:     currencyCode,
			CounterpartyName: counterpartyName,
			CounterpartyIBAN: counterpartyIBAN,
			Description:      description,
			TransactionAt:    transactionAt,
			BookedAt:         bookedAt,
			Status:           aggregate.TransactionStatus(status),
			LastUpdatedBy:    updatedByID,
		}

		transaction.ID = txID
		transaction.Version = version
		transaction.CreatedAt = createdAt
		transaction.UpdatedAt = updatedAt

		if matchedEntityType != nil {
			et := aggregate.MatchedEntityType(*matchedEntityType)
			transaction.MatchedEntityType = &et
		}
		if matchedEntityID != nil {
			eid, _ := uuidv7.Parse(*matchedEntityID)
			transaction.MatchedEntityID = &eid
		}
		transaction.MatchedAt = matchedAt

		_ = transaction.RawPayload.UnmarshalJSON([]byte(rawPayloadJSON))

		transactions = append(transactions, transaction)
	}

	return transactions, rows.Err()
}

// Delete soft deletes transaction
func (r *BankTransactionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE banking_bank_transactions
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	return r.Exec(ctx, query, time.Now(), id.String())
}
