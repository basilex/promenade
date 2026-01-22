package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/accounting/taxcode"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type taxCodeRepository struct {
	db *sqlx.DB
}

// NewTaxCodeRepository creates a new PostgreSQL tax code repository
func NewTaxCodeRepository(db *sqlx.DB) repository.ITaxCodeRepository {
	return &taxCodeRepository{db: db}
}

func (r *taxCodeRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

func (r *taxCodeRepository) Create(ctx context.Context, taxCode *aggregate.TaxCode) error {
	query := `
        INSERT INTO accounting_tax_codes (
            id, version, organization_id, code, name, tax_type, rate,
            tax_payable_account_id, tax_receivable_account_id, is_active,
            last_updated_by, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        )`

	_, err := r.getExecutor(ctx).ExecContext(
		ctx, query,
		taxCode.ID.String(),
		taxCode.Version,
		taxCode.OrganizationID.String(),
		taxCode.Code,
		taxCode.Name,
		taxCode.TaxType,
		taxCode.Rate,
		uuidPtrToString(taxCode.TaxPayableAccountID),
		uuidPtrToString(taxCode.TaxReceivableAccountID),
		taxCode.IsActive,
		taxCode.LastUpdatedBy.String(),
		taxCode.CreatedAt,
		taxCode.UpdatedAt,
	)

	return err
}

func (r *taxCodeRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
	query := `
        SELECT id, version, organization_id, code, name, tax_type, rate,
               tax_payable_account_id, tax_receivable_account_id, is_active,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_tax_codes
        WHERE id = $1 AND deleted_at IS NULL`

	var tc aggregate.TaxCode
	var orgID, lastUpdatedBy string
	var taxPayableAccountID, taxReceivableAccountID sql.NullString

	err := r.getExecutor(ctx).QueryRowxContext(ctx, query, id.String()).Scan(
		&tc.ID,
		&tc.Version,
		&orgID,
		&tc.Code,
		&tc.Name,
		&tc.TaxType,
		&tc.Rate,
		&taxPayableAccountID,
		&taxReceivableAccountID,
		&tc.IsActive,
		&lastUpdatedBy,
		&tc.CreatedAt,
		&tc.UpdatedAt,
		&tc.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, taxcode.ErrTaxCodeNotFound
	}
	if err != nil {
		return nil, err
	}

	tc.OrganizationID, _ = uuidv7.Parse(orgID)
	tc.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

	if taxPayableAccountID.Valid {
		id, _ := uuidv7.Parse(taxPayableAccountID.String)
		tc.TaxPayableAccountID = &id
	}
	if taxReceivableAccountID.Valid {
		id, _ := uuidv7.Parse(taxReceivableAccountID.String)
		tc.TaxReceivableAccountID = &id
	}

	return &tc, nil
}

func (r *taxCodeRepository) GetByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.TaxCode, error) {
	query := `
        SELECT id, version, organization_id, code, name, tax_type, rate,
               tax_payable_account_id, tax_receivable_account_id, is_active,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_tax_codes
        WHERE organization_id = $1 AND code = $2 AND deleted_at IS NULL`

	var tc aggregate.TaxCode
	var orgID, lastUpdatedBy string
	var taxPayableAccountID, taxReceivableAccountID sql.NullString

	err := r.getExecutor(ctx).QueryRowxContext(ctx, query, organizationID.String(), code).Scan(
		&tc.ID,
		&tc.Version,
		&orgID,
		&tc.Code,
		&tc.Name,
		&tc.TaxType,
		&tc.Rate,
		&taxPayableAccountID,
		&taxReceivableAccountID,
		&tc.IsActive,
		&lastUpdatedBy,
		&tc.CreatedAt,
		&tc.UpdatedAt,
		&tc.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, taxcode.ErrTaxCodeNotFound
	}
	if err != nil {
		return nil, err
	}

	tc.OrganizationID, _ = uuidv7.Parse(orgID)
	tc.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

	if taxPayableAccountID.Valid {
		id, _ := uuidv7.Parse(taxPayableAccountID.String)
		tc.TaxPayableAccountID = &id
	}
	if taxReceivableAccountID.Valid {
		id, _ := uuidv7.Parse(taxReceivableAccountID.String)
		tc.TaxReceivableAccountID = &id
	}

	return &tc, nil
}

func (r *taxCodeRepository) Update(ctx context.Context, taxCode *aggregate.TaxCode) error {
	query := `
        UPDATE accounting_tax_codes
        SET version = $2,
            name = $3,
            rate = $4,
            tax_payable_account_id = $5,
            tax_receivable_account_id = $6,
            is_active = $7,
            last_updated_by = $8,
            updated_at = $9
        WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.getExecutor(ctx).ExecContext(
		ctx, query,
		taxCode.ID.String(),
		taxCode.Version,
		taxCode.Name,
		taxCode.Rate,
		uuidPtrToString(taxCode.TaxPayableAccountID),
		uuidPtrToString(taxCode.TaxReceivableAccountID),
		taxCode.IsActive,
		taxCode.LastUpdatedBy.String(),
		taxCode.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return taxcode.ErrTaxCodeNotFound
	}

	return nil
}

func (r *taxCodeRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
        UPDATE accounting_tax_codes
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
		return taxcode.ErrTaxCodeNotFound
	}

	return nil
}

func (r *taxCodeRepository) ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.TaxCode, error) {
	query := `
        SELECT id, version, organization_id, code, name, tax_type, rate,
               tax_payable_account_id, tax_receivable_account_id, is_active,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_tax_codes
        WHERE organization_id = $1 AND deleted_at IS NULL
        ORDER BY code
        LIMIT $2 OFFSET $3`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanTaxCodes(rows)
}

func (r *taxCodeRepository) ListByType(ctx context.Context, organizationID uuidv7.UUID, taxType aggregate.TaxType) ([]*aggregate.TaxCode, error) {
	query := `
        SELECT id, version, organization_id, code, name, tax_type, rate,
               tax_payable_account_id, tax_receivable_account_id, is_active,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_tax_codes
        WHERE organization_id = $1 AND tax_type = $2 AND deleted_at IS NULL
        ORDER BY code`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), taxType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanTaxCodes(rows)
}

func (r *taxCodeRepository) ListActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.TaxCode, error) {
	query := `
        SELECT id, version, organization_id, code, name, tax_type, rate,
               tax_payable_account_id, tax_receivable_account_id, is_active,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_tax_codes
        WHERE organization_id = $1 AND is_active = true AND deleted_at IS NULL
        ORDER BY code`

	rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanTaxCodes(rows)
}

func (r *taxCodeRepository) scanTaxCodes(rows *sqlx.Rows) ([]*aggregate.TaxCode, error) {
	var taxCodes []*aggregate.TaxCode

	for rows.Next() {
		var tc aggregate.TaxCode
		var orgID, lastUpdatedBy string
		var taxPayableAccountID, taxReceivableAccountID sql.NullString

		err := rows.Scan(
			&tc.ID,
			&tc.Version,
			&orgID,
			&tc.Code,
			&tc.Name,
			&tc.TaxType,
			&tc.Rate,
			&taxPayableAccountID,
			&taxReceivableAccountID,
			&tc.IsActive,
			&lastUpdatedBy,
			&tc.CreatedAt,
			&tc.UpdatedAt,
			&tc.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		tc.OrganizationID, _ = uuidv7.Parse(orgID)
		tc.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

		if taxPayableAccountID.Valid {
			id, _ := uuidv7.Parse(taxPayableAccountID.String)
			tc.TaxPayableAccountID = &id
		}
		if taxReceivableAccountID.Valid {
			id, _ := uuidv7.Parse(taxReceivableAccountID.String)
			tc.TaxReceivableAccountID = &id
		}

		taxCodes = append(taxCodes, &tc)
	}

	return taxCodes, nil
}

func uuidPtrToString(id *uuidv7.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
