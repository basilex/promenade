package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/banking/banktransaction"
	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/aggregate"
	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// BankTransactionRepository implements bank transaction persistence
type BankTransactionRepository struct {
	db *sqlx.DB
}

// NewBankTransactionRepository creates a new bank transaction repository
func NewBankTransactionRepository(db *sqlx.DB) repository.IBankTransactionRepository {
	return &BankTransactionRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *BankTransactionRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *BankTransactionRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *BankTransactionRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning rows
func (r *BankTransactionRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows
func (r *BankTransactionRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

// Create inserts a new bank transaction
func (r *BankTransactionRepository) Create(ctx context.Context, tx *aggregate.BankTransaction) error {
	query := `
		INSERT INTO banking_bank_transactions (
			id, version, bank_account_id, external_id, direction, amount_cents, currency_code,
			counterparty_name, counterparty_iban, description, transaction_at, status,
			matched_entity_type, matched_entity_id, matched_at,
			raw_payload_json, last_updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`

	var matchedType, matchedID *string
	var matchedAt *time.Time
	if tx.MatchedEntityType != nil {
		s := string(*tx.MatchedEntityType)
		matchedType = &s
	}
	if tx.MatchedEntityID != nil {
		s := tx.MatchedEntityID.String()
		matchedID = &s
	}
	if tx.MatchedAt != nil {
		matchedAt = tx.MatchedAt
	}

	_, err := r.Exec(ctx, query,
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
		matchedType,
		matchedID,
		matchedAt,
		func() string { v, _ := tx.RawPayload.Value(); return v.(string) }(),
		tx.LastUpdatedBy.String(),
		tx.GetCreatedAt(),
		tx.GetUpdatedAt(),
	)
	return err
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

	executor := r.getExecutor(ctx)
	// Type assertion to access QueryRowContext
	queryer, ok := executor.(interface {
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	})
	if !ok {
		return nil, errors.New("executor does not support QueryRowContext")
	}

	if err := queryer.QueryRowContext(ctx, query, id.String()).Scan(
		&dbID, &version, &bankAccountID, &externalID, &direction, &amountCents, &currencyCode,
		&counterpartyName, &counterpartyIBAN, &description, &transactionAt, &bookedAt, &status,
		&matchedEntityType, &matchedEntityID, &matchedAt,
		&rawPayloadJSON, &lastUpdatedBy, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, banktransaction.ErrBankTransactionNotFound
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
func (r *BankTransactionRepository) Update(ctx context.Context, tx *aggregate.BankTransaction) error {
	query := `
		UPDATE banking_bank_transactions
		SET version = version + 1,
			amount_cents = $1,
			direction = $2,
			status = $3,
			description = $4,
			external_id = $5,
			counterparty_name = $6,
			matched_entity_type = $7,
			matched_entity_id = $8,
			matched_at = $9,
			raw_payload_json = $10,
			updated_at = $11
		WHERE id = $12 AND version = $13 AND deleted_at IS NULL
	`

	var matchedType, matchedID *string
	if tx.MatchedEntityType != nil {
		s := string(*tx.MatchedEntityType)
		matchedType = &s
	}
	if tx.MatchedEntityID != nil {
		s := tx.MatchedEntityID.String()
		matchedID = &s
	}

	rawPayloadJSON, _ := tx.RawPayload.MarshalJSON()

	result, err := r.Exec(ctx, query,
		tx.AmountCents,
		string(tx.Direction),
		string(tx.Status),
		tx.Description,
		tx.ExternalID,
		tx.CounterpartyName,
		matchedType,
		matchedID,
		tx.MatchedAt,
		string(rawPayloadJSON),
		time.Now(),
		tx.ID.String(),
		tx.Version,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	tx.Version++
	tx.UpdatedAt = time.Now()
	return nil
}

func (r *BankTransactionRepository) GetByExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string) (*aggregate.BankTransaction, error) {
	query := `
		SELECT id, version, bank_account_id, amount_cents, direction, status,
			transaction_at, external_id, counterparty_name,
			matched_entity_type, matched_entity_id, matched_at,
			raw_payload_json, created_at, updated_at
		FROM banking_bank_transactions
		WHERE bank_account_id = $1 AND external_id = $2 AND deleted_at IS NULL
	`

	var scannedRow struct {
		ID                string     `db:"id"`
		Version           int        `db:"version"`
		BankAccountID     string     `db:"bank_account_id"`
		AmountCents       int64      `db:"amount_cents"`
		Direction         string     `db:"direction"`
		Status            string     `db:"status"`
		TransactionDate   time.Time  `db:"transaction_at"`
		ExternalID        *string    `db:"external_id"`
		CounterpartyName  *string    `db:"counterparty_name"`
		MatchedEntityType *string    `db:"matched_entity_type"`
		MatchedEntityID   *string    `db:"matched_entity_id"`
		MatchedAt         *time.Time `db:"matched_at"`
		RawPayloadJSON    string     `db:"raw_payload_json"`
		CreatedAt         time.Time  `db:"created_at"`
		UpdatedAt         time.Time  `db:"updated_at"`
	}

	err := sqlx.GetContext(ctx, r.getExecutor(ctx), &scannedRow, query, accountID.String(), externalID)

	if err == sql.ErrNoRows {
		return nil, errors.New("transaction not found")
	}
	if err != nil {
		return nil, err
	}

	txID, _ := uuidv7.Parse(scannedRow.ID)
	accID, _ := uuidv7.Parse(scannedRow.BankAccountID)

	transaction := &aggregate.BankTransaction{
		BankAccountID: accID,
		AmountCents:   scannedRow.AmountCents,
		Direction:     aggregate.TransactionDirection(scannedRow.Direction),
		Status:        aggregate.TransactionStatus(scannedRow.Status),
		TransactionAt: scannedRow.TransactionDate,
		CurrencyCode:  "UAH",
	}

	if scannedRow.ExternalID != nil {
		transaction.ExternalID = *scannedRow.ExternalID
	}
	if scannedRow.CounterpartyName != nil {
		transaction.CounterpartyName = *scannedRow.CounterpartyName
	}

	transaction.ID = txID
	transaction.Version = scannedRow.Version
	transaction.CreatedAt = scannedRow.CreatedAt
	transaction.UpdatedAt = scannedRow.UpdatedAt

	if scannedRow.MatchedEntityType != nil {
		met := aggregate.MatchedEntityType(*scannedRow.MatchedEntityType)
		transaction.MatchedEntityType = &met
	}
	if scannedRow.MatchedEntityID != nil {
		eid, _ := uuidv7.Parse(*scannedRow.MatchedEntityID)
		transaction.MatchedEntityID = &eid
	}
	transaction.MatchedAt = scannedRow.MatchedAt

	_ = transaction.RawPayload.UnmarshalJSON([]byte(scannedRow.RawPayloadJSON))

	return transaction, nil
}

func (r *BankTransactionRepository) ListUnmatched(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*aggregate.BankTransaction, error) {
	query := `
		SELECT id, version, bank_account_id, amount_cents, direction, status,
			transaction_at, external_id, counterparty_name,
			matched_entity_type, matched_entity_id, matched_at,
			raw_payload_json, created_at, updated_at
		FROM banking_bank_transactions
		WHERE bank_account_id = $1 
		  AND matched_entity_id IS NULL 
		  AND deleted_at IS NULL
		ORDER BY transaction_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, accountID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var transactions []*aggregate.BankTransaction
	for rows.Next() {
		var (
			dbID                          string
			version                       int
			accountID_str                 string
			amountCents                   int64
			direction                     string
			status                        string
			transactionDate               time.Time
			externalID                    *string
			counterpartyName              *string
			matchedEntityType             *string
			matchedEntityID               *string
			matchedAt                     *time.Time
			rawPayloadJSON                string
			createdAt                     time.Time
			updatedAt                     time.Time
		)

		if err := rows.Scan(
			&dbID, &version, &accountID_str, &amountCents, &direction, &status,
			&transactionDate, &externalID, &counterpartyName,
			&matchedEntityType, &matchedEntityID, &matchedAt,
			&rawPayloadJSON, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}

		txID, _ := uuidv7.Parse(dbID)
		accID, _ := uuidv7.Parse(accountID_str)

		transaction := &aggregate.BankTransaction{
			BankAccountID:    accID,
			AmountCents:      amountCents,
			Direction:        aggregate.TransactionDirection(direction),
			Status:           aggregate.TransactionStatus(status),
			TransactionAt:    transactionDate,
			CurrencyCode:     "UAH",
		}

		if externalID != nil {
			transaction.ExternalID = *externalID
		}
		if counterpartyName != nil {
			transaction.CounterpartyName = *counterpartyName
		}

		transaction.ID = txID
		transaction.Version = version
		transaction.CreatedAt = createdAt
		transaction.UpdatedAt = updatedAt

		if matchedEntityType != nil {
			met := aggregate.MatchedEntityType(*matchedEntityType)
			transaction.MatchedEntityType = &met
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

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, accountID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() // Ignore close error
	}()

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

	_, err := r.Exec(ctx, query, time.Now(), id.String())
	return err
}

// CountByAccount returns count of transactions for account
func (r *BankTransactionRepository) CountByAccount(ctx context.Context, accountID uuidv7.UUID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM banking_bank_transactions 
		WHERE bank_account_id = $1 AND deleted_at IS NULL
	`
	
	var count int
	if err := r.Get(ctx, &count, query, accountID.String()); err != nil {
		return 0, err
	}
	return count, nil
}
