package mssql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	countryerrors "github.com/basilex/promenade/internal/contexts/shared/country"
	"github.com/basilex/promenade/internal/contexts/shared/country/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/country/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type countryRepository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new MS SQL Server country repository
func NewRepository(db sqlx.ExtContext) repository.ICountryRepository {
	return &countryRepository{db: db}
}

func (r *countryRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Country, error) {
	var c aggregate.Country
	query := `
		SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active,
		       created_at, updated_at
		FROM shared_countries
		WHERE id = @p1 AND is_active = 1
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, id)
	if err == sql.ErrNoRows {
		return nil, countryerrors.ErrCountryNotFound
	}
	return &c, err
}

func (r *countryRepository) GetByCode(ctx context.Context, code string) (*aggregate.Country, error) {
	var c aggregate.Country
	query := `
		SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active,
		       created_at, updated_at
		FROM shared_countries
		WHERE code = @p1 AND is_active = 1
	`
	err := sqlx.GetContext(ctx, r.db, &c, query, code)
	if err == sql.ErrNoRows {
		return nil, countryerrors.ErrCountryNotFound
	}
	return &c, err
}

func (r *countryRepository) List(ctx context.Context) ([]*aggregate.Country, error) {
	var countries []*aggregate.Country
	query := `
		SELECT id, code, code3, numeric_code, name, name_local, phone_code, is_active,
		       created_at, updated_at
		FROM shared_countries
		WHERE is_active = 1
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &countries, query)
	return countries, err
}

func (r *countryRepository) Create(ctx context.Context, c *aggregate.Country) error {
	query := `
		INSERT INTO shared_countries (id, code, code3, numeric_code, name, name_local, phone_code, is_active, created_at, updated_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10)
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.Code3, c.NumericCode, c.Name, c.NameLocal, c.PhoneCode, c.IsActive, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *countryRepository) Update(ctx context.Context, c *aggregate.Country) error {
	query := `
		UPDATE shared_countries
		SET code = @p2, code3 = @p3, numeric_code = @p4, name = @p5, name_local = @p6, phone_code = @p7, is_active = @p8, updated_at = @p9
		WHERE id = @p1
	`
	_, err := r.db.ExecContext(ctx, query, c.ID, c.Code, c.Code3, c.NumericCode, c.Name, c.NameLocal, c.PhoneCode, c.IsActive, c.UpdatedAt)
	return err
}

func (r *countryRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_countries SET is_active = 0 WHERE id = @p1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
