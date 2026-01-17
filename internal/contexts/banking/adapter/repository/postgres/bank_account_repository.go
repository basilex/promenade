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

// BankAccountRepository implements bank account persistence
type BankAccountRepository struct {
	*BaseRepository
}

// NewBankAccountRepository creates a new bank account repository
func NewBankAccountRepository(db *sqlx.DB) *BankAccountRepository {
	return &BankAccountRepository{
		BaseRepository: NewBaseRepository(db),
	}
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

	return r.Exec(ctx, query,
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

	return r.Exec(ctx, query,
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

	if err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&dbID, &version, &organizationID, &name, &bankName, &iban, &accountNumber, &currencyCode,
		&provider, &providerAccountID, &status, &lastSyncAt, &balanceCents,
		&metadataJSON, &lastUpdatedBy, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, aggregate.ErrBankAccountNotFound
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

	rows, err := r.db.QueryContext(ctx, query, organizationID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	return r.Exec(ctx, query, time.Now(), id.String())
}
