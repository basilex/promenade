package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type repository struct {
	db sqlx.ExtContext // Works with both *sqlx.DB and *sqlx.Tx
}

// NewRepository creates a new PostgreSQL timezone repository
func NewRepository(db sqlx.ExtContext) timezone.IRepository {
	return &repository{db: db}
}

func (r *repository) GetByID(ctx context.Context, id uuidv7.UUID) (*timezone.Timezone, error) {
	var t timezone.Timezone
	query := `
		SELECT id, name, abbreviation, utc_offset, is_active
		FROM shared_timezones
		WHERE id = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &t, query, id)
	if err == sql.ErrNoRows {
		return nil, timezone.ErrTimezoneNotFound
	}
	return &t, err
}

func (r *repository) GetByName(ctx context.Context, name string) (*timezone.Timezone, error) {
	var t timezone.Timezone
	query := `
		SELECT id, name, abbreviation, utc_offset, is_active
		FROM shared_timezones
		WHERE name = $1 AND is_active = TRUE
	`
	err := sqlx.GetContext(ctx, r.db, &t, query, name)
	if err == sql.ErrNoRows {
		return nil, timezone.ErrTimezoneNotFound
	}
	return &t, err
}

func (r *repository) List(ctx context.Context) ([]*timezone.Timezone, error) {
	var timezones []*timezone.Timezone
	query := `
		SELECT id, name, abbreviation, utc_offset, is_active
		FROM shared_timezones
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	err := sqlx.SelectContext(ctx, r.db, &timezones, query)
	return timezones, err
}

func (r *repository) Create(ctx context.Context, t *timezone.Timezone) error {
	query := `
		INSERT INTO shared_timezones (id, name, abbreviation, utc_offset, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Abbreviation, t.UTCOffset, t.IsActive)
	return err
}

func (r *repository) Update(ctx context.Context, t *timezone.Timezone) error {
	query := `
		UPDATE shared_timezones
		SET name = $2, abbreviation = $3, utc_offset = $4, is_active = $5
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Abbreviation, t.UTCOffset, t.IsActive)
	return err
}

func (r *repository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE shared_timezones SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
