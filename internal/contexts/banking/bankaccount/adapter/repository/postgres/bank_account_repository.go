package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/banking/bankaccount"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/aggregate"
	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// BankAccountRepository implements bank account persistence
type BankAccountRepository struct {
	db *sqlx.DB
}

// NewBankAccountRepository creates a new bank account repository
func NewBankAccountRepository(db *sqlx.DB) repository.IBankAccountRepository {
	return &BankAccountRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *BankAccountRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *BankAccountRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *BankAccountRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning rows
func (r *BankAccountRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows
func (r *BankAccountRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

// Create inserts a new bank account
func (r *BankAccountRepository) Create(ctx context.Context, account *aggregate.BankAccount) error {
	query := `
		INSERT INTO banking_bank_accounts (
			id, version, organization_id, name, bank_name, iban, account_number, currency_code,
			provider, provider_account_id, status, last_sync_at, balance_cents,
			metadata_json, last_updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	var lastSyncAt *time.Time
	if account.LastSyncAt != nil {
		lastSyncAt = account.LastSyncAt
	}

	_, err := r.Exec(ctx, query,
		account.GetID().String(),
		account.GetVersion(),
		account.OrganizationID.String(),
		account.Name,
		account.BankName,
		account.IBAN,
		account.AccountNumber,
		account.CurrencyCode,
		string(account.Provider),
		account.ProviderAccountID,
		string(account.Status),
		lastSyncAt,
		account.BalanceCents,
		func() string { v, _ := account.Metadata.Value(); return v.(string) }(),
		account.LastUpdatedBy.String(),
		account.GetCreatedAt(),
		account.GetUpdatedAt(),
	)
	return err
}

// Update updates an existing bank account
func (r *BankAccountRepository) Update(ctx context.Context, account *aggregate.BankAccount) error {
	query := `
		UPDATE banking_bank_accounts SET
			version = $2,
			name = $3,
			bank_name = $4,
			iban = $5,
			account_number = $6,
			provider = $7,
			provider_account_id = $8,
			status = $9,
			last_sync_at = $10,
			balance_cents = $11,
			metadata_json = $12,
			updated_at = $13
		WHERE id = $1 AND deleted_at IS NULL
	`

	var lastSyncAt *time.Time
	if account.LastSyncAt != nil {
		lastSyncAt = account.LastSyncAt
	}

	_, err := r.Exec(ctx, query,
		account.GetID().String(),
		account.GetVersion()+1,
		account.Name,
		account.BankName,
		account.IBAN,
		account.AccountNumber,
		string(account.Provider),
		account.ProviderAccountID,
		string(account.Status),
		lastSyncAt,
		account.BalanceCents,
		func() string { v, _ := account.Metadata.Value(); return v.(string) }(),
		time.Now(),
	)
	return err
}

// GetByID retrieves bank account by ID
func (r *BankAccountRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.BankAccount, error) {
	query := `
		SELECT id, version, organization_id, name, bank_name, iban, account_number, currency_code,
			provider, provider_account_id, status, last_sync_at, balance_cents,
			metadata_json, last_updated_by, created_at, updated_at
		FROM banking_bank_accounts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		dbID              string
		version           int
		organizationID    string
		name              string
		bankName          string
		iban              string
		accountNumber     string
		currencyCode      string
		provider          string
		providerAccountID string
		status            string
		lastSyncAt        *time.Time
		balanceCents      int64
		metadataJSON      string
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
		&dbID, &version, &organizationID, &name, &bankName, &iban, &accountNumber, &currencyCode,
		&provider, &providerAccountID, &status, &lastSyncAt, &balanceCents,
		&metadataJSON, &lastUpdatedBy, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, bankaccount.ErrBankAccountNotFound
		}
		return nil, err
	}

	accountID, _ := uuidv7.Parse(dbID)
	orgID, _ := uuidv7.Parse(organizationID)
	updatedByID, _ := uuidv7.Parse(lastUpdatedBy)

	account := &aggregate.BankAccount{
		OrganizationID:    orgID,
		Name:              name,
		BankName:          bankName,
		IBAN:              iban,
		AccountNumber:     accountNumber,
		CurrencyCode:      currencyCode,
		Provider:          aggregate.BankProvider(provider),
		ProviderAccountID: providerAccountID,
		Status:            aggregate.BankAccountStatus(status),
		LastSyncAt:        lastSyncAt,
		BalanceCents:      balanceCents,
		LastUpdatedBy:     updatedByID,
	}

	account.ID = accountID
	account.Version = version
	account.CreatedAt = createdAt
	account.UpdatedAt = updatedAt
	_ = account.Metadata.UnmarshalJSON([]byte(metadataJSON))

	return account, nil
}

// GetByOrganization retrieves all accounts for organization
func (r *BankAccountRepository) GetByOrganization(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.BankAccount, error) {
	query := `
		SELECT id, version, organization_id, name, bank_name, iban, account_number, currency_code,
			provider, provider_account_id, status, last_sync_at, balance_cents,
			metadata_json, last_updated_by, created_at, updated_at
		FROM banking_bank_accounts
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, organizationID.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() // Ignore close error
	}()

	var accounts []*aggregate.BankAccount
	for rows.Next() {
		var (
			dbID              string
			version           int
			orgID             string
			name              string
			bankName          string
			iban              string
			accountNumber     string
			currencyCode      string
			provider          string
			providerAccountID string
			status            string
			lastSyncAt        *time.Time
			balanceCents      int64
			metadataJSON      string
			lastUpdatedBy     string
			createdAt         time.Time
			updatedAt         time.Time
		)

		if err := rows.Scan(
			&dbID, &version, &orgID, &name, &bankName, &iban, &accountNumber, &currencyCode,
			&provider, &providerAccountID, &status, &lastSyncAt, &balanceCents,
			&metadataJSON, &lastUpdatedBy, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}

		accountID, _ := uuidv7.Parse(dbID)
		organizationUUID, _ := uuidv7.Parse(orgID)
		updatedByID, _ := uuidv7.Parse(lastUpdatedBy)

		account := &aggregate.BankAccount{
			OrganizationID:    organizationUUID,
			Name:              name,
			BankName:          bankName,
			IBAN:              iban,
			AccountNumber:     accountNumber,
			CurrencyCode:      currencyCode,
			Provider:          aggregate.BankProvider(provider),
			ProviderAccountID: providerAccountID,
			Status:            aggregate.BankAccountStatus(status),
			LastSyncAt:        lastSyncAt,
			BalanceCents:      balanceCents,
			LastUpdatedBy:     updatedByID,
		}

		account.ID = accountID
		account.Version = version
		account.CreatedAt = createdAt
		account.UpdatedAt = updatedAt
		_ = account.Metadata.UnmarshalJSON([]byte(metadataJSON))

		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// Delete soft deletes bank account
func (r *BankAccountRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE banking_bank_accounts
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.Exec(ctx, query, time.Now(), id.String())
	return err
}

// CountByOrganization returns count of bank accounts for organization
func (r *BankAccountRepository) CountByOrganization(ctx context.Context, organizationID uuidv7.UUID) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM banking_bank_accounts 
		WHERE organization_id = $1 AND deleted_at IS NULL
	`

	var count int
	if err := r.Get(ctx, &count, query, organizationID.String()); err != nil {
		return 0, err
	}
	return count, nil
}

// GetByProviderAccountID retrieves bank account by provider and external account ID
func (r *BankAccountRepository) GetByProviderAccountID(ctx context.Context, provider aggregate.BankProvider, providerAccountID string) (*aggregate.BankAccount, error) {
	query := `
		SELECT id, version, organization_id, name, bank_name, iban, account_number, currency_code,
			provider, provider_account_id, status, last_sync_at, balance_cents,
			metadata_json, last_updated_by, created_at, updated_at
		FROM banking_bank_accounts
		WHERE provider = $1 
		  AND provider_account_id = $2 
		  AND deleted_at IS NULL
	`

	var scannedRow struct {
		ID                string     `db:"id"`
		Version           int        `db:"version"`
		OrganizationID    string     `db:"organization_id"`
		Name              string     `db:"name"`
		BankName          string     `db:"bank_name"`
		IBAN              string     `db:"iban"`
		AccountNumber     string     `db:"account_number"`
		CurrencyCode      string     `db:"currency_code"`
		Provider          string     `db:"provider"`
		ProviderAccountID string     `db:"provider_account_id"`
		Status            string     `db:"status"`
		LastSyncAt        *time.Time `db:"last_sync_at"`
		BalanceCents      int64      `db:"balance_cents"`
		MetadataJSON      string     `db:"metadata_json"`
		LastUpdatedBy     string     `db:"last_updated_by"`
		CreatedAt         time.Time  `db:"created_at"`
		UpdatedAt         time.Time  `db:"updated_at"`
	}

	if err := r.Get(ctx, &scannedRow, query, string(provider), providerAccountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, bankaccount.ErrBankAccountNotFound
		}
		return nil, err
	}

	accountID, _ := uuidv7.Parse(scannedRow.ID)
	organizationUUID, _ := uuidv7.Parse(scannedRow.OrganizationID)
	updatedByID, _ := uuidv7.Parse(scannedRow.LastUpdatedBy)

	account := &aggregate.BankAccount{
		OrganizationID:    organizationUUID,
		Name:              scannedRow.Name,
		BankName:          scannedRow.BankName,
		IBAN:              scannedRow.IBAN,
		AccountNumber:     scannedRow.AccountNumber,
		CurrencyCode:      scannedRow.CurrencyCode,
		Provider:          aggregate.BankProvider(scannedRow.Provider),
		ProviderAccountID: scannedRow.ProviderAccountID,
		Status:            aggregate.BankAccountStatus(scannedRow.Status),
		LastSyncAt:        scannedRow.LastSyncAt,
		BalanceCents:      scannedRow.BalanceCents,
		LastUpdatedBy:     updatedByID,
	}

	account.ID = accountID
	account.Version = scannedRow.Version
	account.CreatedAt = scannedRow.CreatedAt
	account.UpdatedAt = scannedRow.UpdatedAt
	_ = account.Metadata.UnmarshalJSON([]byte(scannedRow.MetadataJSON))

	return account, nil
}

// List retrieves bank accounts with pagination
func (r *BankAccountRepository) List(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.BankAccount, error) {
	query := `
		SELECT id, version, organization_id, name, bank_name, iban, account_number, currency_code,
			provider, provider_account_id, status, last_sync_at, balance_cents,
			metadata_json, last_updated_by, created_at, updated_at
		FROM banking_bank_accounts
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.getExecutor(ctx).QueryContext(ctx, query, organizationID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var accounts []*aggregate.BankAccount
	for rows.Next() {
		var (
			dbID              string
			version           int
			orgID             string
			name              string
			bankName          string
			iban              string
			accountNumber     string
			currencyCode      string
			provider          string
			providerAccountID string
			status            string
			lastSyncAt        *time.Time
			balanceCents      int64
			metadataJSON      string
			lastUpdatedBy     string
			createdAt         time.Time
			updatedAt         time.Time
		)

		if err := rows.Scan(
			&dbID, &version, &orgID, &name, &bankName, &iban, &accountNumber, &currencyCode,
			&provider, &providerAccountID, &status, &lastSyncAt, &balanceCents,
			&metadataJSON, &lastUpdatedBy, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}

		accountID, _ := uuidv7.Parse(dbID)
		organizationUUID, _ := uuidv7.Parse(orgID)
		updatedByID, _ := uuidv7.Parse(lastUpdatedBy)

		account := &aggregate.BankAccount{
			OrganizationID:    organizationUUID,
			Name:              name,
			BankName:          bankName,
			IBAN:              iban,
			AccountNumber:     accountNumber,
			CurrencyCode:      currencyCode,
			Provider:          aggregate.BankProvider(provider),
			ProviderAccountID: providerAccountID,
			Status:            aggregate.BankAccountStatus(status),
			LastSyncAt:        lastSyncAt,
			BalanceCents:      balanceCents,
			LastUpdatedBy:     updatedByID,
		}

		account.ID = accountID
		account.Version = version
		account.CreatedAt = createdAt
		account.UpdatedAt = updatedAt
		_ = account.Metadata.UnmarshalJSON([]byte(metadataJSON))

		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}
