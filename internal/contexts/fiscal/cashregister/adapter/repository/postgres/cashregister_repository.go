package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// cashRegisterRepository implements cashregister.IRepository using PostgreSQL
type cashRegisterRepository struct {
	*BaseRepository
}

// NewCashRegisterRepository creates a new PostgreSQL cash register repository
func NewCashRegisterRepository(db *sqlx.DB) cashregister.IRepository {
	return &cashRegisterRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// cashRegisterRow represents database row structure for cash register table
type cashRegisterRow struct {
	ID                     string         `db:"id"`
	Version                int            `db:"version"`
	OrganizationID         string         `db:"organization_id"`
	FiscalNumber           string         `db:"fiscal_number"`
	Model                  string         `db:"model"`
	Status                 string         `db:"status"`
	LicenseKey             sql.NullString `db:"license_key"`
	LastSyncAt             sql.NullTime   `db:"last_sync_at"`
	ProviderCashRegisterID sql.NullString `db:"provider_cash_register_id"`
	ActiveShiftID          sql.NullString `db:"active_shift_id"`
	ShiftOpenedAt          sql.NullTime   `db:"shift_opened_at"`
	ShiftClosedAt          sql.NullTime   `db:"shift_closed_at"`
	LastZReportID          sql.NullString `db:"last_z_report_id"`
	LastZReportAt          sql.NullTime   `db:"last_z_report_at"`
	LastUpdatedBy          string         `db:"last_updated_by"`
	CreatedAt              time.Time      `db:"created_at"`
	UpdatedAt              time.Time      `db:"updated_at"`
	DeletedAt              sql.NullTime   `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *cashRegisterRow) toEntity() (*cashregister.CashRegister, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	organizationID, err := uuidv7.Parse(r.OrganizationID)
	if err != nil {
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	lastUpdatedBy, err := uuidv7.Parse(r.LastUpdatedBy)
	if err != nil {
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	// Reconstruct entity with all fields
	cr := &cashregister.CashRegister{
		OrganizationID: organizationID,
		FiscalNumber:   r.FiscalNumber,
		Model:          r.Model,
		Status:         cashregister.CashRegisterStatus(r.Status),
		LicenseKey:     "",
		LastUpdatedBy:  lastUpdatedBy,
	}

	// Set BaseAggregate fields directly
	cr.ID = id
	cr.Version = r.Version
	cr.CreatedAt = r.CreatedAt
	cr.UpdatedAt = r.UpdatedAt

	if r.LicenseKey.Valid {
		cr.LicenseKey = r.LicenseKey.String
	}

	if r.LastSyncAt.Valid {
		cr.LastSyncAt = &r.LastSyncAt.Time
	}
	if r.ProviderCashRegisterID.Valid {
		cr.ProviderCashRegisterID = r.ProviderCashRegisterID.String
	}
	if r.ActiveShiftID.Valid {
		cr.ActiveShiftID = r.ActiveShiftID.String
	}
	if r.ShiftOpenedAt.Valid {
		cr.ShiftOpenedAt = &r.ShiftOpenedAt.Time
	}
	if r.ShiftClosedAt.Valid {
		cr.ShiftClosedAt = &r.ShiftClosedAt.Time
	}
	if r.LastZReportID.Valid {
		cr.LastZReportID = r.LastZReportID.String
	}
	if r.LastZReportAt.Valid {
		cr.LastZReportAt = &r.LastZReportAt.Time
	}

	return cr, nil
}

// fromEntity converts domain entity to database row
func fromEntity(cr *cashregister.CashRegister) *cashRegisterRow {
	row := &cashRegisterRow{
		ID:             cr.GetID().String(),
		Version:        cr.GetVersion(),
		OrganizationID: cr.OrganizationID.String(),
		FiscalNumber:   cr.FiscalNumber,
		Model:          cr.Model,
		Status:         string(cr.Status),
		LastUpdatedBy:  cr.LastUpdatedBy.String(),
		CreatedAt:      cr.CreatedAt,
		UpdatedAt:      cr.UpdatedAt,
	}

	if cr.LicenseKey != "" {
		row.LicenseKey = sql.NullString{String: cr.LicenseKey, Valid: true}
	}

	if cr.LastSyncAt != nil {
		row.LastSyncAt = sql.NullTime{Time: *cr.LastSyncAt, Valid: true}
	}
	if cr.ProviderCashRegisterID != "" {
		row.ProviderCashRegisterID = sql.NullString{String: cr.ProviderCashRegisterID, Valid: true}
	}
	if cr.ActiveShiftID != "" {
		row.ActiveShiftID = sql.NullString{String: cr.ActiveShiftID, Valid: true}
	}
	if cr.ShiftOpenedAt != nil {
		row.ShiftOpenedAt = sql.NullTime{Time: *cr.ShiftOpenedAt, Valid: true}
	}
	if cr.ShiftClosedAt != nil {
		row.ShiftClosedAt = sql.NullTime{Time: *cr.ShiftClosedAt, Valid: true}
	}
	if cr.LastZReportID != "" {
		row.LastZReportID = sql.NullString{String: cr.LastZReportID, Valid: true}
	}
	if cr.LastZReportAt != nil {
		row.LastZReportAt = sql.NullTime{Time: *cr.LastZReportAt, Valid: true}
	}

	if cr.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *cr.DeletedAt, Valid: true}
	}

	return row
}

// Create creates a new cash register
func (r *cashRegisterRepository) Create(ctx context.Context, cr *cashregister.CashRegister) error {
	query := `
		INSERT INTO fiscal_cash_registers (
			id, version, organization_id, fiscal_number, model, status,
			license_key, last_sync_at, provider_cash_register_id, active_shift_id,
			shift_opened_at, shift_closed_at, last_z_report_id, last_z_report_at,
			last_updated_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	row := fromEntity(cr)
	_, err := r.Exec(ctx, query,
		row.ID, row.Version, row.OrganizationID, row.FiscalNumber, row.Model, row.Status,
		row.LicenseKey, row.LastSyncAt, row.ProviderCashRegisterID, row.ActiveShiftID,
		row.ShiftOpenedAt, row.ShiftClosedAt, row.LastZReportID, row.LastZReportAt,
		row.LastUpdatedBy, row.CreatedAt, row.UpdatedAt,
	)

	if err != nil {
		return cashregister.ErrCashRegisterCreateFailed
	}

	return nil
}

// GetByID retrieves cash register by ID
func (r *cashRegisterRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*cashregister.CashRegister, error) {
	query := `
		SELECT id, version, organization_id, fiscal_number, model, status,
		       license_key, last_sync_at, provider_cash_register_id, active_shift_id,
		       shift_opened_at, shift_closed_at, last_z_report_id, last_z_report_at, last_updated_by,
		       created_at, updated_at, deleted_at
		FROM fiscal_cash_registers
		WHERE id = $1 AND deleted_at IS NULL`

	var row cashRegisterRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, cashregister.ErrCashRegisterNotFound
		}
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	return row.toEntity()
}

// GetByFiscalNumber retrieves cash register by fiscal number
func (r *cashRegisterRepository) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*cashregister.CashRegister, error) {
	query := `
		SELECT id, version, organization_id, fiscal_number, model, status,
		       license_key, last_sync_at, provider_cash_register_id, active_shift_id,
		       shift_opened_at, shift_closed_at, last_z_report_id, last_z_report_at, last_updated_by,
		       created_at, updated_at, deleted_at
		FROM fiscal_cash_registers
		WHERE fiscal_number = $1 AND deleted_at IS NULL`

	var row cashRegisterRow
	if err := r.Get(ctx, &row, query, fiscalNumber); err != nil {
		if err == sql.ErrNoRows {
			return nil, cashregister.ErrCashRegisterNotFound
		}
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	return row.toEntity()
}

// List retrieves cash registers with filters
func (r *cashRegisterRepository) List(ctx context.Context, filters *cashregister.ListFilters) ([]*cashregister.CashRegister, error) {
	query := `
		SELECT id, version, organization_id, fiscal_number, model, status,
		       license_key, last_sync_at, provider_cash_register_id, active_shift_id,
		       shift_opened_at, shift_closed_at, last_z_report_id, last_z_report_at, last_updated_by,
		       created_at, updated_at, deleted_at
		FROM fiscal_cash_registers
		WHERE deleted_at IS NULL`

	args := []interface{}{}
	if filters != nil && filters.OrganizationID != nil {
		query += ` AND organization_id = $1`
		args = append(args, filters.OrganizationID.String())
	}

	query += ` ORDER BY created_at DESC`

	var rows []cashRegisterRow
	if err := r.Select(ctx, &rows, query, args...); err != nil {
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	cashRegisters := make([]*cashregister.CashRegister, len(rows))
	for i, row := range rows {
		cr, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		cashRegisters[i] = cr
	}

	return cashRegisters, nil
}

// GetByLocation retrieves all cash registers for a location
func (r *cashRegisterRepository) GetByLocation(ctx context.Context, locationID uuidv7.UUID) ([]*cashregister.CashRegister, error) {
	query := `
		SELECT id, version, organization_id, fiscal_number, model, status,
		       license_key, last_sync_at, provider_cash_register_id, active_shift_id,
		       shift_opened_at, shift_closed_at, last_z_report_id, last_z_report_at, last_updated_by,
		       created_at, updated_at, deleted_at
		FROM fiscal_cash_registers
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var rows []cashRegisterRow
	if err := r.Select(ctx, &rows, query, locationID.String()); err != nil {
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	cashRegisters := make([]*cashregister.CashRegister, len(rows))
	for i, row := range rows {
		cr, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		cashRegisters[i] = cr
	}

	return cashRegisters, nil
}

// ListActive retrieves all active cash registers
func (r *cashRegisterRepository) ListActive(ctx context.Context) ([]*cashregister.CashRegister, error) {
	query := `
		SELECT id, version, organization_id, fiscal_number, model, status,
		       license_key, last_sync_at, provider_cash_register_id, active_shift_id,
		       shift_opened_at, shift_closed_at, last_z_report_id, last_z_report_at, last_updated_by,
		       created_at, updated_at, deleted_at
		FROM fiscal_cash_registers
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var rows []cashRegisterRow
	if err := r.Select(ctx, &rows, query, string(cashregister.StatusActive)); err != nil {
		return nil, cashregister.ErrCashRegisterCreateFailed
	}

	cashRegisters := make([]*cashregister.CashRegister, len(rows))
	for i, row := range rows {
		cr, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		cashRegisters[i] = cr
	}

	return cashRegisters, nil
}

// Update updates cash register
func (r *cashRegisterRepository) Update(ctx context.Context, cr *cashregister.CashRegister) error {
	query := `
		UPDATE fiscal_cash_registers
		SET version = $1, model = $2, status = $3,
		    license_key = $4, last_sync_at = $5,
		    provider_cash_register_id = $6, active_shift_id = $7,
		    shift_opened_at = $8, shift_closed_at = $9,
		    last_z_report_id = $10, last_z_report_at = $11,
		    last_updated_by = $12,
		    updated_at = $13
		WHERE id = $14 AND version = $15 AND deleted_at IS NULL`

	row := fromEntity(cr)
	result, err := r.Exec(ctx, query,
		row.Version+1, row.Model, row.Status,
		row.LicenseKey, row.LastSyncAt,
		row.ProviderCashRegisterID, row.ActiveShiftID,
		row.ShiftOpenedAt, row.ShiftClosedAt,
		row.LastZReportID, row.LastZReportAt,
		row.LastUpdatedBy,
		row.UpdatedAt,
		row.ID, row.Version,
	)

	if err != nil {
		return cashregister.ErrCashRegisterUpdateFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return cashregister.ErrCashRegisterUpdateFailed
	}

	if rowsAffected == 0 {
		return cashregister.ErrCashRegisterNotFound
	}

	// Update version directly
	cr.Version = cr.GetVersion() + 1
	return nil
}

// Delete soft-deletes cash register
func (r *cashRegisterRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE fiscal_cash_registers
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, time.Now(), id.String())
	if err != nil {
		return cashregister.ErrCashRegisterDeleteFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return cashregister.ErrCashRegisterDeleteFailed
	}

	if rowsAffected == 0 {
		return cashregister.ErrCashRegisterNotFound
	}

	return nil
}
