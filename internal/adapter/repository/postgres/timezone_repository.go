package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/jmoiron/sqlx"
)

type timezoneRepository struct {
	*BaseRepository
}

// NewTimezoneRepository creates a new timezone repository
func NewTimezoneRepository(db *sqlx.DB) repository.TimezoneRepository {
	return &timezoneRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new timezone
func (r *timezoneRepository) Create(ctx context.Context, timezone *entity.Timezone) error {
	query := `
		INSERT INTO timezones (id, name, abbreviation, utc_offset, utc_dst_offset, description, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	executor := r.getExecutor(ctx)
	return executor.QueryRowxContext(ctx, query,
		timezone.ID,
		timezone.Name,
		timezone.Abbreviation,
		timezone.UtcOffset,
		timezone.UtcDstOffset,
		timezone.Description,
		timezone.IsActive,
	).Scan(&timezone.CreatedAt, &timezone.UpdatedAt)
}

// GetByID retrieves a timezone by ID
func (r *timezoneRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Timezone, error) {
	query := `
		SELECT id, name, abbreviation, utc_offset, utc_dst_offset, description, is_active, created_at, updated_at
		FROM timezones
		WHERE id = $1
	`

	var timezone entity.Timezone
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &timezone, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &timezone, nil
}

// GetByName retrieves a timezone by name (e.g., "Europe/Moscow")
func (r *timezoneRepository) GetByName(ctx context.Context, name string) (*entity.Timezone, error) {
	query := `
		SELECT id, name, abbreviation, utc_offset, utc_dst_offset, description, is_active, created_at, updated_at
		FROM timezones
		WHERE name = $1
	`

	var timezone entity.Timezone
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &timezone, query, name); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &timezone, nil
}

// List retrieves all timezones with pagination
func (r *timezoneRepository) List(ctx context.Context, params pagination.Params) ([]*entity.Timezone, *pagination.Metadata, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM timezones`
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &total, countQuery); err != nil {
		return nil, nil, err
	}

	// Get paginated timezones
	query := `
		SELECT id, name, abbreviation, utc_offset, utc_dst_offset, description, is_active, created_at, updated_at
		FROM timezones
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	var timezones []*entity.Timezone
	limit := params.GetLimit()
	offset := params.GetOffset()
	if err := sqlx.SelectContext(ctx, executor, &timezones, query, limit, offset); err != nil {
		return nil, nil, err
	}

	metadata := pagination.NewMetadata(total, limit, offset)
	return timezones, metadata, nil
}

// ListActive retrieves all active timezones
func (r *timezoneRepository) ListActive(ctx context.Context) ([]*entity.Timezone, error) {
	query := `
		SELECT id, name, abbreviation, utc_offset, utc_dst_offset, description, is_active, created_at, updated_at
		FROM timezones
		WHERE is_active = true
		ORDER BY name ASC
	`

	var timezones []*entity.Timezone
	executor := r.getExecutor(ctx)
	if err := sqlx.SelectContext(ctx, executor, &timezones, query); err != nil {
		return nil, err
	}

	return timezones, nil
}

// Update updates an existing timezone
func (r *timezoneRepository) Update(ctx context.Context, timezone *entity.Timezone) error {
	query := `
		UPDATE timezones
		SET name = $2, abbreviation = $3, utc_offset = $4, utc_dst_offset = $5, description = $6, is_active = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	executor := r.getExecutor(ctx)
	err := executor.QueryRowxContext(ctx, query,
		timezone.ID,
		timezone.Name,
		timezone.Abbreviation,
		timezone.UtcOffset,
		timezone.UtcDstOffset,
		timezone.Description,
		timezone.IsActive,
	).Scan(&timezone.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return entity.ErrNotFound
		}
		return err
	}

	return nil
}

// Delete deletes a timezone by ID
func (r *timezoneRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM timezones WHERE id = $1`

	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

// Exists checks if a timezone exists by name
func (r *timezoneRepository) Exists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM timezones WHERE name = $1)`

	var exists bool
	executor := r.getExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &exists, query, name); err != nil {
		return false, fmt.Errorf("failed to check timezone existence: %w", err)
	}

	return exists, nil
}
