package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/jmoiron/sqlx"
)

type fiscalPeriodRepository struct {
	db *sqlx.DB
}

// NewFiscalPeriodRepository creates a new PostgreSQL fiscal period repository
func NewFiscalPeriodRepository(db *sqlx.DB) repository.IFiscalPeriodRepository {
	return &fiscalPeriodRepository{db: db}
}

func (r *fiscalPeriodRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

func (r *fiscalPeriodRepository) Create(ctx context.Context, period *aggregate.FiscalPeriod) error {
	query := `
        INSERT INTO accounting_fiscal_periods (
            id, version, organization_id, period_type, period_code, period_name,
            start_date, end_date, status, lock_date,
            last_updated_by, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        )`

	var lockDate *string
	if period.LockDate != nil {
		ld := period.LockDate.Format("2006-01-02")
		lockDate = &ld
	}

	_, err := r.getExecutor(ctx).ExecContext(
		ctx, query,
		period.ID.String(),
		period.Version,
		period.OrganizationID.String(),
		period.PeriodType,
		period.Code,
		period.Name,
		period.StartDate.Format("2006-01-02"),
		period.EndDate.Format("2006-01-02"),
		period.Status,
		lockDate,
		period.LastUpdatedBy.String(),
		period.CreatedAt,
		period.UpdatedAt,
	)

	return err
}

func (r *fiscalPeriodRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	query := `
        SELECT id, version, organization_id, period_type, period_code, period_name,
               start_date, end_date, status, lock_date,
               closed_by, closed_at, reopened_by, reopened_at, locked_by, locked_at,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_fiscal_periods
        WHERE id = $1 AND deleted_at IS NULL`

	var period aggregate.FiscalPeriod
	var orgID, lastUpdatedBy string
	var closedBy, reopenedBy, lockedBy sql.NullString
	var closedAt, reopenedAt, lockedAt sql.NullTime
	var lockDate sql.NullString

	err := r.getExecutor(ctx).QueryRowxContext(ctx, query, id.String()).Scan(
		&period.ID,
		&period.Version,
		&orgID,
		&period.PeriodType,
		&period.Code,
		&period.Name,
		&period.StartDate,
		&period.EndDate,
		&period.Status,
		&lockDate,
		&closedBy,
		&closedAt,
		&reopenedBy,
		&reopenedAt,
		&lockedBy,
		&lockedAt,
		&lastUpdatedBy,
		&period.CreatedAt,
		&period.UpdatedAt,
		&period.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fiscalperiod.ErrPeriodNotFound
	}
	if err != nil {
		return nil, err
	}

	period.OrganizationID, _ = uuidv7.Parse(orgID)
	period.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

	if lockDate.Valid {
		ld, _ := time.Parse("2006-01-02", lockDate.String)
		period.LockDate = &ld
	}

	if closedBy.Valid {
		id, _ := uuidv7.Parse(closedBy.String)
		period.ClosedBy = &id
	}
	if closedAt.Valid {
		period.ClosedAt = &closedAt.Time
	}
	if reopenedBy.Valid {
		id, _ := uuidv7.Parse(reopenedBy.String)
		period.ReopenedBy = &id
	}
	if reopenedAt.Valid {
		period.ReopenedAt = &reopenedAt.Time
	}
	if lockedBy.Valid {
		id, _ := uuidv7.Parse(lockedBy.String)
		period.LockedBy = &id
	}
	if lockedAt.Valid {
		period.LockedAt = &lockedAt.Time
	}

	return &period, nil
}

func (r *fiscalPeriodRepository) GetByOrganizationAndDate(ctx context.Context, organizationID uuidv7.UUID, date string) (*aggregate.FiscalPeriod, error) {
	query := `
        SELECT id, version, organization_id, period_type, period_code, period_name,
               start_date, end_date, status, lock_date,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_fiscal_periods
        WHERE organization_id = $1 
          AND start_date <= $2 
          AND end_date >= $2 
          AND deleted_at IS NULL`

	var period aggregate.FiscalPeriod
	var orgID, lastUpdatedBy string
	var lockDate sql.NullString

	err := r.getExecutor(ctx).QueryRowxContext(ctx, query, organizationID.String(), date).Scan(
		&period.ID,
		&period.Version,
		&orgID,
		&period.PeriodType,
		&period.Code,
		&period.Name,
		&period.StartDate,
		&period.EndDate,
		&period.Status,
		&lockDate,
		&lastUpdatedBy,
		&period.CreatedAt,
		&period.UpdatedAt,
		&period.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fiscalperiod.ErrPeriodNotFound
	}
	if err != nil {
		return nil, err
	}

	period.OrganizationID, _ = uuidv7.Parse(orgID)
	period.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

	if lockDate.Valid {
		ld, _ := time.Parse("2006-01-02", lockDate.String)
		period.LockDate = &ld
	}

	return &period, nil
}

func (r *fiscalPeriodRepository) Update(ctx context.Context, period *aggregate.FiscalPeriod) error {
	query := `
        UPDATE accounting_fiscal_periods
        SET version = $2,
            status = $3,
            lock_date = $4,
            closed_by = $5,
            closed_at = $6,
            reopened_by = $7,
            reopened_at = $8,
            locked_by = $9,
            locked_at = $10,
            last_updated_by = $11,
            updated_at = $12
        WHERE id = $1 AND deleted_at IS NULL`

	var lockDate *string
	if period.LockDate != nil {
		ld := period.LockDate.Format("2006-01-02")
		lockDate = &ld
	}

	var closedBy, reopenedBy, lockedBy *string
	if period.ClosedBy != nil {
		cb := period.ClosedBy.String()
		closedBy = &cb
	}
	if period.ReopenedBy != nil {
		rb := period.ReopenedBy.String()
		reopenedBy = &rb
	}
	if period.LockedBy != nil {
		lb := period.LockedBy.String()
		lockedBy = &lb
	}

	result, err := r.getExecutor(ctx).ExecContext(
		ctx, query,
		period.ID.String(),
		period.Version,
		period.Status,
		lockDate,
		closedBy,
		period.ClosedAt,
		reopenedBy,
		period.ReopenedAt,
		lockedBy,
		period.LockedAt,
		period.LastUpdatedBy.String(),
		period.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fiscalperiod.ErrPeriodNotFound
	}

	return nil
}

func (r *fiscalPeriodRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
        UPDATE accounting_fiscal_periods
        SET deleted_at = $2
        WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.getExecutor(ctx).ExecContext(ctx, query, id.String(), time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fiscalperiod.ErrPeriodNotFound
	}

	return nil
}

func (r *fiscalPeriodRepository) ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error) {
	query := `
        SELECT id, version, organization_id, period_type, period_code, period_name,
               start_date, end_date, status, lock_date,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_fiscal_periods
        WHERE organization_id = $1 AND deleted_at IS NULL
        ORDER BY start_date DESC
        LIMIT $2 OFFSET $3`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanPeriods(rows)
}

func (r *fiscalPeriodRepository) ListByYear(ctx context.Context, organizationID uuidv7.UUID, fiscalYear int) ([]*aggregate.FiscalPeriod, error) {
	query := `
        SELECT id, version, organization_id, period_type, period_code, period_name,
               start_date, end_date, status, lock_date,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_fiscal_periods
        WHERE organization_id = $1 
          AND period_code LIKE $2
          AND deleted_at IS NULL
        ORDER BY start_date`

	yearPrefix := fmt.Sprintf("%d%%", fiscalYear)
	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), yearPrefix)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanPeriods(rows)
}

func (r *fiscalPeriodRepository) ListOpen(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error) {
	query := `
        SELECT id, version, organization_id, period_type, period_code, period_name,
               start_date, end_date, status, lock_date,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_fiscal_periods
        WHERE organization_id = $1 
          AND status = 'open' 
          AND deleted_at IS NULL
        ORDER BY start_date DESC`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanPeriods(rows)
}

func (r *fiscalPeriodRepository) scanPeriods(rows *sqlx.Rows) ([]*aggregate.FiscalPeriod, error) {
	var periods []*aggregate.FiscalPeriod

	for rows.Next() {
		var period aggregate.FiscalPeriod
		var orgID, lastUpdatedBy string
		var lockDate sql.NullString

		err := rows.Scan(
			&period.ID,
			&period.Version,
			&orgID,
			&period.PeriodType,
			&period.Code,
			&period.Name,
			&period.StartDate,
			&period.EndDate,
			&period.Status,
			&lockDate,
			&lastUpdatedBy,
			&period.CreatedAt,
			&period.UpdatedAt,
			&period.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		period.OrganizationID, _ = uuidv7.Parse(orgID)
		period.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

		if lockDate.Valid {
			ld, _ := time.Parse("2006-01-02", lockDate.String)
			period.LockDate = &ld
		}

		periods = append(periods, &period)
	}

	return periods, nil
}
