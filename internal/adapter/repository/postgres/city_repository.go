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

type cityRepository struct {
	*BaseRepository
}

// NewCityRepository creates a new city repository
func NewCityRepository(db *sqlx.DB) repository.CityRepository {
	return &cityRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new city
func (r *cityRepository) Create(ctx context.Context, city *entity.City) error {
	query := `
		INSERT INTO core_cities (id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		city.ID,
		city.RegionID,
		city.CountryID,
		city.Name,
		city.NameLocal,
		city.Latitude,
		city.Longitude,
		city.Population,
		city.IsCapital,
		city.IsRegionalCapital,
		city.IsActive,
		city.SortOrder,
		city.CreatedAt,
		city.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create city: %w", err)
	}
	return nil
}

// GetByID retrieves a city by ID
func (r *cityRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.City, error) {
	query := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		WHERE id = $1
	`
	var city entity.City
	err := r.Get(ctx, &city, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get city: %w", err)
	}
	return &city, nil
}

// List retrieves all cities with pagination
func (r *cityRepository) List(ctx context.Context, offset, limit int) ([]entity.City, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_cities`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count cities: %w", err)
	}

	// Get cities
	query := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		ORDER BY sort_order ASC, population DESC NULLS LAST, name ASC
		LIMIT $1 OFFSET $2
	`
	var cities []entity.City
	if err := r.Select(ctx, &cities, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list cities: %w", err)
	}

	return cities, total, nil
}

// ListByCountry retrieves cities for a specific country
func (r *cityRepository) ListByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.City, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_cities WHERE country_id = $1`
	if err := r.Get(ctx, &total, countQuery, countryID); err != nil {
		return nil, 0, fmt.Errorf("failed to count cities by country: %w", err)
	}

	// Get cities
	query := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		WHERE country_id = $1
		ORDER BY sort_order ASC, population DESC NULLS LAST, name ASC
		LIMIT $2 OFFSET $3
	`
	var cities []entity.City
	if err := r.Select(ctx, &cities, query, countryID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list cities by country: %w", err)
	}

	return cities, total, nil
}

// ListByRegion retrieves cities for a specific region
func (r *cityRepository) ListByRegion(ctx context.Context, regionID uuidv7.UUID, offset, limit int) ([]entity.City, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_cities WHERE region_id = $1`
	if err := r.Get(ctx, &total, countQuery, regionID); err != nil {
		return nil, 0, fmt.Errorf("failed to count cities by region: %w", err)
	}

	// Get cities
	query := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		WHERE region_id = $1
		ORDER BY sort_order ASC, population DESC NULLS LAST, name ASC
		LIMIT $2 OFFSET $3
	`
	var cities []entity.City
	if err := r.Select(ctx, &cities, query, regionID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list cities by region: %w", err)
	}

	return cities, total, nil
}

// ListActive retrieves only active cities
func (r *cityRepository) ListActive(ctx context.Context, offset, limit int) ([]entity.City, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_cities WHERE is_active = TRUE`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count active cities: %w", err)
	}

	// Get cities
	query := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, population DESC NULLS LAST, name ASC
		LIMIT $1 OFFSET $2
	`
	var cities []entity.City
	if err := r.Select(ctx, &cities, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list active cities: %w", err)
	}

	return cities, total, nil
}

// ListActiveByCountry retrieves active cities for a specific country
func (r *cityRepository) ListActiveByCountry(ctx context.Context, countryID uuidv7.UUID, offset, limit int) ([]entity.City, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_cities WHERE country_id = $1 AND is_active = TRUE`
	if err := r.Get(ctx, &total, countQuery, countryID); err != nil {
		return nil, 0, fmt.Errorf("failed to count active cities by country: %w", err)
	}

	// Get cities
	query := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		WHERE country_id = $1 AND is_active = TRUE
		ORDER BY sort_order ASC, population DESC NULLS LAST, name ASC
		LIMIT $2 OFFSET $3
	`
	var cities []entity.City
	if err := r.Select(ctx, &cities, query, countryID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list active cities by country: %w", err)
	}

	return cities, total, nil
}

// ListCapitals retrieves all capital cities
func (r *cityRepository) ListCapitals(ctx context.Context, offset, limit int) ([]entity.City, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_cities WHERE is_capital = TRUE`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count capital cities: %w", err)
	}

	// Get cities
	query := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		WHERE is_capital = TRUE
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`
	var cities []entity.City
	if err := r.Select(ctx, &cities, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list capital cities: %w", err)
	}

	return cities, total, nil
}

// SearchByName searches cities by name (supports partial matching)
func (r *cityRepository) SearchByName(ctx context.Context, query string, offset, limit int) ([]entity.City, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM core_cities WHERE name ILIKE $1 OR name_local ILIKE $1`
	searchPattern := "%" + query + "%"
	if err := r.Get(ctx, &total, countQuery, searchPattern); err != nil {
		return nil, 0, fmt.Errorf("failed to count cities by name: %w", err)
	}

	// Get cities
	searchQuery := `
		SELECT id, region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, is_active, sort_order, created_at, updated_at
		FROM core_cities
		WHERE name ILIKE $1 OR name_local ILIKE $1
		ORDER BY population DESC NULLS LAST, name ASC
		LIMIT $2 OFFSET $3
	`
	var cities []entity.City
	if err := r.Select(ctx, &cities, searchQuery, searchPattern, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to search cities by name: %w", err)
	}

	return cities, total, nil
}

// Update updates an existing city
func (r *cityRepository) Update(ctx context.Context, city *entity.City) error {
	query := `
		UPDATE core_cities
		SET region_id = $2, name = $3, name_local = $4, latitude = $5, longitude = $6, 
		    population = $7, is_capital = $8, is_regional_capital = $9, is_active = $10, 
		    sort_order = $11, updated_at = $12
		WHERE id = $1
	`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		city.ID,
		city.RegionID,
		city.Name,
		city.NameLocal,
		city.Latitude,
		city.Longitude,
		city.Population,
		city.IsCapital,
		city.IsRegionalCapital,
		city.IsActive,
		city.SortOrder,
		city.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update city: %w", err)
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

// Delete deletes a city by ID
func (r *cityRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM core_cities WHERE id = $1`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete city: %w", err)
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
