package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	timezoneerrors "github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/internal/contexts/shared/timezone/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/timezone/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type timezoneRepository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new PostgreSQL timezone repository
func NewRepository(db sqlx.ExtContext) repository.ITimezoneRepository {
	return &timezoneRepository{db: db}
}

func (r *timezoneRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Timezone, error) {
	var t aggregate.Timezone
	query := `
		SELECT id, name, abbreviation, utc_offset, is_active, created_at, updated_at
		FROM shared_timezones
		WHERE id = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &t, query, id)
	if err == sql.ErrNoRows {
		return nil, timezoneerrors.ErrTimezoneNotFound
	}
	return &t, err
}

func (r *timezoneRepository) GetByName(ctx context.Context, name string) (*aggregate.Timezone, error) {
	var t aggregate.Timezone
	query := `
		SELECT id, name, abbreviation, utc_offset, is_active, created_at, updated_at
		FROM shared_timezones
		WHERE name = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &t, query, name)
	if err == sql.ErrNoRows {
		return nil, timezoneerrors.ErrTimezoneNotFound
	}
	return &t, err
}

func (r *timezoneRepository) List(ctx context.Context) ([]*aggregate.Timezone, error) {
	var timezones []*aggregate.Timezone
	query := `
		SELECT id, name, abbreviation, utc_offset, is_active, created_at, updated_at
		FROM shared_timezones
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &timezones, query)
	return timezones, err
}

func (r *timezoneRepository) Create(ctx context.Context, t *aggregate.Timezone) error {
	query := `
		INSERT INTO shared_timezones (id, name, abbreviation, utc_offset, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Abbreviation, t.UTCOffset, t.IsActive, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *timezoneRepository) Update(ctx context.Context, t *aggregate.Timezone) error {
	query := `
		UPDATE shared_timezones
		SET name = $2, abbreviation = $3, utc_offset = $4, is_active = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Abbreviation, t.UTCOffset, t.IsActive, t.UpdatedAt)
	return err
}

func (r *timezoneRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_timezones SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
