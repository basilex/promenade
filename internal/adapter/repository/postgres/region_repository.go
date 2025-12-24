package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/jmoiron/sqlx"
)

type regionRepository struct {
	*BaseRepository
}

// NewRegionRepository creates a new region repository
func NewRegionRepository(db *sqlx.DB) repository.IRegionRepository {
	return &regionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new region
func (r *regionRepository) Create(ctx context.Context, region *entity.Region) error {
	query := `
		INSERT INTO core_regions (id, country_id, name, code, region_type, latitude, longitude, is_active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		region.ID,
		region.CountryID,
		region.Name,
		region.Code,
		region.RegionType,
		region.Latitude,
		region.Longitude,
		region.IsActive,
		region.SortOrder,
		region.CreatedAt,
		region.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create region: %w", err)
	}
	return nil
}

// GetByID retrieves a region by ID
func (r *regionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Region, error) {
	query := `
		SELECT id, country_id, name, code, region_type, latitude, longitude, is_active, sort_order, created_at, updated_at
		FROM core_regions
		WHERE id = $1
	`
	var region entity.Region
	err := r.Get(ctx, &region, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get region: %w", err)
	}
	return &region, nil
}

// GetByCode retrieves a region by country and code
func (r *regionRepository) GetByCode(ctx context.Context, countryID uuidv7.UUID, code string) (*entity.Region, error) {
	query := `
		SELECT id, country_id, name, code, region_type, latitude, longitude, is_active, sort_order, created_at, updated_at
		FROM core_regions
		WHERE country_id = $1 AND code = $2
	`
	var region entity.Region
	err := r.Get(ctx, &region, query, countryID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get region by code: %w", err)
	}
	return &region, nil
}

// List retrieves all regions with pagination
func (r *regionRepository) List(ctx context.Context, offset, limit int) ([]entity.Region, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_regions`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count regions: %w", err)
	}

	// Get regions
	query := `
		SELECT id, country_id, name, code, region_type, latitude, longitude, is_active, sort_order, created_at, updated_at
		FROM core_regions
		ORDER BY sort_order ASC, name ASC
		LIMIT $1 OFFSET $2
	`
	var regions []entity.Region
	if err := r.Select(ctx, &regions, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list regions: %w", err)
	}

	return regions, total, nil
}

// ListByCountry retrieves regions for a specific country
func (r *regionRepository) ListByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.Region, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_regions WHERE country_id = $1`
	if err := r.Get(ctx, &total, countQuery, countryID); err != nil {
		return nil, 0, fmt.Errorf("failed to count regions by country: %w", err)
	}

	// Get regions
	query := `
		SELECT id, country_id, name, code, region_type, latitude, longitude, is_active, sort_order, created_at, updated_at
		FROM core_regions
		WHERE country_id = $1
		ORDER BY sort_order ASC, name ASC
		LIMIT $2 OFFSET $3
	`
	var regions []entity.Region
	if err := r.Select(ctx, &regions, query, countryID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list regions by country: %w", err)
	}

	return regions, total, nil
}

// ListActive retrieves only active regions
func (r *regionRepository) ListActive(ctx context.Context, offset, limit int) ([]entity.Region, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_regions WHERE is_active = TRUE`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count active regions: %w", err)
	}

	// Get regions
	query := `
		SELECT id, country_id, name, code, region_type, latitude, longitude, is_active, sort_order, created_at, updated_at
		FROM core_regions
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, name ASC
		LIMIT $1 OFFSET $2
	`
	var regions []entity.Region
	if err := r.Select(ctx, &regions, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list active regions: %w", err)
	}

	return regions, total, nil
}

// ListActiveByCountry retrieves active regions for a specific country
func (r *regionRepository) ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.Region, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_regions WHERE country_id = $1 AND is_active = TRUE`
	if err := r.Get(ctx, &total, countQuery, countryID); err != nil {
		return nil, 0, fmt.Errorf("failed to count active regions by country: %w", err)
	}

	// Get regions
	query := `
		SELECT id, country_id, name, code, region_type, latitude, longitude, is_active, sort_order, created_at, updated_at
		FROM core_regions
		WHERE country_id = $1 AND is_active = TRUE
		ORDER BY sort_order ASC, name ASC
		LIMIT $2 OFFSET $3
	`
	var regions []entity.Region
	if err := r.Select(ctx, &regions, query, countryID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list active regions by country: %w", err)
	}

	return regions, total, nil
}

// Update updates an existing region
func (r *regionRepository) Update(ctx context.Context, region *entity.Region) error {
	query := `
		UPDATE core_regions
		SET name = $2, code = $3, region_type = $4, latitude = $5, longitude = $6, is_active = $7, sort_order = $8, updated_at = $9
		WHERE id = $1
	`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		region.ID,
		region.Name,
		region.Code,
		region.RegionType,
		region.Latitude,
		region.Longitude,
		region.IsActive,
		region.SortOrder,
		region.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update region: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}

// Delete deletes a region by ID
func (r *regionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM core_regions WHERE id = $1`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}
