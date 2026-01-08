package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type contractRepository struct {
	*BaseRepository
}

// NewContractRepository creates a new contract repository
func NewContractRepository(db *sqlx.DB) contract.IContractRepository {
	return &contractRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// contractRow represents database row for contract
type contractRow struct {
	ID                  string         `db:"id"`
	OrderID             string         `db:"order_id"`
	CustomerID          string         `db:"customer_id"`
	Status              string         `db:"status"`
	Terms               sql.NullString `db:"terms"`
	TermsURL            sql.NullString `db:"terms_url"`
	SignatureID         sql.NullString `db:"signature_id"`
	Version             int            `db:"version"`
	SignedAt            sql.NullTime   `db:"signed_at"`
	ActivatedAt         sql.NullTime   `db:"activated_at"`
	CompletedAt         sql.NullTime   `db:"completed_at"`
	TerminatedAt        sql.NullTime   `db:"terminated_at"`
	RenewedAt           sql.NullTime   `db:"renewed_at"`
	ExpiresAt           sql.NullTime   `db:"expires_at"`
	TerminationReason   sql.NullString `db:"termination_reason"`
	SignedByName        sql.NullString `db:"signed_by_name"`
	SignedByEmail       sql.NullString `db:"signed_by_email"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
	DeletedAt           sql.NullTime   `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (row *contractRow) toEntity() (*contract.Contract, error) {
	id, err := uuidv7.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("parse contract id: %w", err)
	}

	orderID, err := uuidv7.Parse(row.OrderID)
	if err != nil {
		return nil, fmt.Errorf("parse order id: %w", err)
	}

	customerID, err := uuidv7.Parse(row.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("parse customer id: %w", err)
	}

	terms := row.Terms.String
	if !row.Terms.Valid {
		terms = ""
	}

	c := contract.NewContract(orderID, customerID, terms)
	c.ID = id  // Set ID directly
	c.Status = contract.ContractStatus(row.Status)
	c.Version = row.Version
	c.SetCreatedAt(row.CreatedAt)
	c.SetUpdatedAt(row.UpdatedAt)

	if row.TermsURL.Valid {
		c.TermsURL = row.TermsURL.String
	}
	if row.SignatureID.Valid {
		c.SignatureID = row.SignatureID.String
	}
	if row.SignedAt.Valid {
		c.SignedAt = &row.SignedAt.Time
	}
	if row.ActivatedAt.Valid {
		c.ActivatedAt = &row.ActivatedAt.Time
	}
	if row.CompletedAt.Valid {
		c.CompletedAt = &row.CompletedAt.Time
	}
	if row.TerminatedAt.Valid {
		c.TerminatedAt = &row.TerminatedAt.Time
	}
	if row.RenewedAt.Valid {
		c.RenewedAt = &row.RenewedAt.Time
	}
	if row.ExpiresAt.Valid {
		c.ExpiresAt = &row.ExpiresAt.Time
	}
	if row.TerminationReason.Valid {
		c.TerminationReason = row.TerminationReason.String
	}
	if row.SignedByName.Valid {
		c.SignedByName = row.SignedByName.String
	}
	if row.SignedByEmail.Valid {
		c.SignedByEmail = row.SignedByEmail.String
	}
	if row.DeletedAt.Valid {
		c.DeletedAt = &row.DeletedAt.Time  // Set DeletedAt directly
	}

	return c, nil
}

// toRow converts domain entity to database row
func toRow(c *contract.Contract) *contractRow {
	row := &contractRow{
		ID:         c.GetID().String(),
		OrderID:    c.OrderID.String(),
		CustomerID: c.CustomerID.String(),
		Status:     string(c.Status),
		Version:    c.Version,
		CreatedAt:  c.GetCreatedAt(),
		UpdatedAt:  c.GetUpdatedAt(),
	}

	if c.Terms != "" {
		row.Terms = sql.NullString{String: c.Terms, Valid: true}
	}
	if c.TermsURL != "" {
		row.TermsURL = sql.NullString{String: c.TermsURL, Valid: true}
	}
	if c.SignatureID != "" {
		row.SignatureID = sql.NullString{String: c.SignatureID, Valid: true}
	}
	if c.SignedAt != nil && !c.SignedAt.IsZero() {
		row.SignedAt = sql.NullTime{Time: *c.SignedAt, Valid: true}
	}
	if c.ActivatedAt != nil && !c.ActivatedAt.IsZero() {
		row.ActivatedAt = sql.NullTime{Time: *c.ActivatedAt, Valid: true}
	}
	if c.CompletedAt != nil && !c.CompletedAt.IsZero() {
		row.CompletedAt = sql.NullTime{Time: *c.CompletedAt, Valid: true}
	}
	if c.TerminatedAt != nil && !c.TerminatedAt.IsZero() {
		row.TerminatedAt = sql.NullTime{Time: *c.TerminatedAt, Valid: true}
	}
	if c.RenewedAt != nil && !c.RenewedAt.IsZero() {
		row.RenewedAt = sql.NullTime{Time: *c.RenewedAt, Valid: true}
	}
	if c.ExpiresAt != nil && !c.ExpiresAt.IsZero() {
		row.ExpiresAt = sql.NullTime{Time: *c.ExpiresAt, Valid: true}
	}
	if c.TerminationReason != "" {
		row.TerminationReason = sql.NullString{String: c.TerminationReason, Valid: true}
	}
	if c.SignedByName != "" {
		row.SignedByName = sql.NullString{String: c.SignedByName, Valid: true}
	}
	if c.SignedByEmail != "" {
		row.SignedByEmail = sql.NullString{String: c.SignedByEmail, Valid: true}
	}
	if c.DeletedAt != nil && !c.DeletedAt.IsZero() {
		row.DeletedAt = sql.NullTime{Time: *c.DeletedAt, Valid: true}
	}

	return row
}

// Create inserts a new contract
func (r *contractRepository) Create(ctx context.Context, c *contract.Contract) error {
	row := toRow(c)

	query := `
		INSERT INTO order_contracts (
			id, order_id, customer_id, status, terms, terms_url, signature_id,
			version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
			expires_at, termination_reason, signed_by_name, signed_by_email,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`

	return r.Exec(ctx, query,
		row.ID, row.OrderID, row.CustomerID, row.Status, row.Terms, row.TermsURL,
		row.SignatureID, row.Version, row.SignedAt, row.ActivatedAt, row.CompletedAt,
		row.TerminatedAt, row.RenewedAt, row.ExpiresAt, row.TerminationReason,
		row.SignedByName, row.SignedByEmail, row.CreatedAt, row.UpdatedAt,
	)
}

// GetByID retrieves a contract by ID
func (r *contractRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*contract.Contract, error) {
	var row contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE id = $1 AND deleted_at IS NULL
	`

	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, contract.ErrContractNotFound
		}
		return nil, fmt.Errorf("get contract: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing contract
func (r *contractRepository) Update(ctx context.Context, c *contract.Contract) error {
	row := toRow(c)

	query := `
		UPDATE order_contracts SET
			status = $1, terms = $2, terms_url = $3, signature_id = $4,
			version = version + 1, signed_at = $5, activated_at = $6, completed_at = $7,
			terminated_at = $8, renewed_at = $9, expires_at = $10, termination_reason = $11,
			signed_by_name = $12, signed_by_email = $13, updated_at = $14
		WHERE id = $15 AND deleted_at IS NULL AND version = $16
	`

	result, err := r.getExecutor(ctx).ExecContext(ctx, query,
		row.Status, row.Terms, row.TermsURL, row.SignatureID, row.SignedAt,
		row.ActivatedAt, row.CompletedAt, row.TerminatedAt, row.RenewedAt,
		row.ExpiresAt, row.TerminationReason, row.SignedByName, row.SignedByEmail,
		time.Now(), row.ID, row.Version,
	)
	if err != nil {
		return fmt.Errorf("update contract: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if affected == 0 {
		return contract.ErrContractNotFound
	}

	return nil
}

// Delete soft-deletes a contract
func (r *contractRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE order_contracts
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.getExecutor(ctx).ExecContext(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("delete contract: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if affected == 0 {
		return contract.ErrContractNotFound
	}

	return nil
}

// List retrieves contracts with pagination
func (r *contractRepository) List(ctx context.Context, offset, limit int) ([]*contract.Contract, int, error) {
	var rows []contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	if err := r.Select(ctx, &rows, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list contracts: %w", err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM order_contracts WHERE deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("count contracts: %w", err)
	}

	contracts := make([]*contract.Contract, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("convert row to entity: %w", err)
		}
		contracts = append(contracts, c)
	}

	return contracts, total, nil
}

// GetByOrder retrieves contracts for a specific order
func (r *contractRepository) GetByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*contract.Contract, error) {
	var rows []contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE order_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	if err := r.Select(ctx, &rows, query, orderID.String()); err != nil {
		return nil, fmt.Errorf("get contracts by order: %w", err)
	}

	contracts := make([]*contract.Contract, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("convert row to entity: %w", err)
		}
		contracts = append(contracts, c)
	}

	return contracts, nil
}

// GetByCustomer retrieves contracts for a specific customer
func (r *contractRepository) GetByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*contract.Contract, error) {
	var rows []contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	if err := r.Select(ctx, &rows, query, customerID.String()); err != nil {
		return nil, fmt.Errorf("get contracts by customer: %w", err)
	}

	contracts := make([]*contract.Contract, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("convert row to entity: %w", err)
		}
		contracts = append(contracts, c)
	}

	return contracts, nil
}

// GetActiveContracts retrieves all active contracts
func (r *contractRepository) GetActiveContracts(ctx context.Context) ([]*contract.Contract, error) {
	var rows []contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	if err := r.Select(ctx, &rows, query, string(contract.ContractStatusActive)); err != nil {
		return nil, fmt.Errorf("get active contracts: %w", err)
	}

	contracts := make([]*contract.Contract, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("convert row to entity: %w", err)
		}
		contracts = append(contracts, c)
	}

	return contracts, nil
}

// ListByStatus retrieves contracts by status with pagination
func (r *contractRepository) ListByStatus(ctx context.Context, status contract.ContractStatus, offset, limit int) ([]*contract.Contract, int, error) {
	var rows []contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	if err := r.Select(ctx, &rows, query, string(status), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list contracts by status: %w", err)
	}

	// Get total count for this status
	var total int
	countQuery := `SELECT COUNT(*) FROM order_contracts WHERE status = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, string(status)); err != nil {
		return nil, 0, fmt.Errorf("count contracts by status: %w", err)
	}

	contracts := make([]*contract.Contract, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("convert row to entity: %w", err)
		}
		contracts = append(contracts, c)
	}

	return contracts, total, nil
}

// ListByCustomer retrieves contracts for a customer with pagination
func (r *contractRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, offset, limit int) ([]*contract.Contract, int, error) {
	var rows []contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	if err := r.Select(ctx, &rows, query, customerID.String(), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list contracts by customer: %w", err)
	}

	// Get total count for this customer
	var total int
	countQuery := `SELECT COUNT(*) FROM order_contracts WHERE customer_id = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, customerID.String()); err != nil {
		return nil, 0, fmt.Errorf("count contracts by customer: %w", err)
	}

	contracts := make([]*contract.Contract, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("convert row to entity: %w", err)
		}
		contracts = append(contracts, c)
	}

	return contracts, total, nil
}

// ListExpiringSoon retrieves contracts expiring within specified days
func (r *contractRepository) ListExpiringSoon(ctx context.Context, days int) ([]*contract.Contract, error) {
	var rows []contractRow

	query := `
		SELECT id, order_id, customer_id, status, terms, terms_url, signature_id,
		       version, signed_at, activated_at, completed_at, terminated_at, renewed_at,
		       expires_at, termination_reason, signed_by_name, signed_by_email,
		       created_at, updated_at, deleted_at
		FROM order_contracts
		WHERE expires_at IS NOT NULL
		  AND expires_at > NOW()
		  AND expires_at <= NOW() + ($1 || ' days')::INTERVAL
		  AND deleted_at IS NULL
		ORDER BY expires_at ASC
	`

	if err := r.Select(ctx, &rows, query, days); err != nil {
		return nil, fmt.Errorf("list expiring contracts: %w", err)
	}

	contracts := make([]*contract.Contract, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("convert row to entity: %w", err)
		}
		contracts = append(contracts, c)
	}

	return contracts, nil
}
