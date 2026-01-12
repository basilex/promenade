package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/warehouse/location"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// locationRepository implements location.IRepository
type locationRepository struct {
	db *sqlx.DB
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *sqlx.DB) location.IRepository {
	return &locationRepository{db: db}
}

// locationRow represents a database row
type locationRow struct {
	ID               string         `db:"id"`
	Code             string         `db:"code"`
	Name             string         `db:"name"`
	Type             string         `db:"type"`
	Description      string         `db:"description"`
	ParentID         sql.NullString `db:"parent_id"`
	Path             string         `db:"path"`
	Level            int            `db:"level"`
	Capacity         int            `db:"capacity"`
	CurrentOccupancy int            `db:"current_occupancy"`
	IsLimited        bool           `db:"is_limited"`
	Width            float64        `db:"width"`
	Height           float64        `db:"height"`
	Depth            float64        `db:"depth"`
	Status           string         `db:"status"`
	IsPickable       bool           `db:"is_pickable"`
	IsPutawayable    bool           `db:"is_putawayable"`
	Notes            string         `db:"notes"`
	CreatedAt        string         `db:"created_at"`
	UpdatedAt        string         `db:"updated_at"`
	Version          int            `db:"version"`
}

// toEntity converts database row to domain entity
func (r *locationRow) toEntity() (*location.Location, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID: %w", err)
	}

	// Hydrate entity fields directly (bypass business logic for database loading)
	loc := &location.Location{}
	loc.ID = id
	loc.Code = r.Code
	loc.Name = r.Name
	loc.Type = location.LocationType(r.Type)
	loc.Description = r.Description

	if r.ParentID.Valid {
		parentID, err := uuidv7.Parse(r.ParentID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid parent ID: %w", err)
		}
		loc.ParentID = &parentID
	}

	loc.Path = r.Path
	loc.Level = r.Level
	loc.Capacity = r.Capacity
	loc.CurrentOccupancy = r.CurrentOccupancy
	loc.IsLimited = r.IsLimited
	loc.Width = r.Width
	loc.Height = r.Height
	loc.Depth = r.Depth
	loc.Status = location.LocationStatus(r.Status)
	loc.IsPickable = r.IsPickable
	loc.IsPutawayable = r.IsPutawayable
	loc.Notes = r.Notes
	loc.Version = r.Version

	return loc, nil
}

// fromEntity converts domain entity to database row
func fromEntity(loc *location.Location) *locationRow {
	var parentID sql.NullString
	if loc.ParentID != nil {
		parentID = sql.NullString{String: loc.ParentID.String(), Valid: true}
	}

	row := &locationRow{
		ID:                loc.GetID().String(),
		Code:              loc.Code,
		ParentID:          parentID,
		Name:             loc.Name,
		Type:             string(loc.Type),
		Description:      loc.Description,
		Path:             loc.Path,
		Level:            loc.Level,
		Capacity:         loc.Capacity,
		CurrentOccupancy: loc.CurrentOccupancy,
		IsLimited:        loc.IsLimited,
		Width:            loc.Width,
		Height:           loc.Height,
		Depth:            loc.Depth,
		Status:           string(loc.Status),
		IsPickable:       loc.IsPickable,
		IsPutawayable:    loc.IsPutawayable,
		Notes:            loc.Notes,
		Version:          loc.GetVersion(),
	}

	if loc.ParentID != nil {
		row.ParentID = sql.NullString{String: loc.ParentID.String(), Valid: true}
	}

	return row
}

// getExecutor returns either a transaction or the regular database connection
func (r *locationRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Create inserts a new location
func (r *locationRepository) Create(ctx context.Context, loc *location.Location) error {
	query := `
		INSERT INTO warehouse_locations (
			id, code, name, type, description, parent_id, path, level,
			capacity, current_occupancy, is_limited, width, height, depth,
			status, is_pickable, is_putawayable, notes, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	row := fromEntity(loc)
	_, err := r.getExecutor(ctx).ExecContext(ctx, query,
		row.ID, row.Code, row.Name, row.Type, row.Description,
		row.ParentID,
		row.Path, row.Level, row.Capacity, row.CurrentOccupancy, row.IsLimited,
		row.Width, row.Height, row.Depth, row.Status,
		row.IsPickable, row.IsPutawayable, row.Notes, row.Version,
	)

	return err
}

// Update updates an existing location
func (r *locationRepository) Update(ctx context.Context, loc *location.Location) error {
	query := `
		UPDATE warehouse_locations SET
			code = $2, name = $3, type = $4, description = $5,
			parent_id = $6, path = $7, level = $8,
			capacity = $9, current_occupancy = $10, is_limited = $11,
			width = $12, height = $13, depth = $14,
			status = $15, is_pickable = $16, is_putawayable = $17,
			notes = $18, version = $19, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND version = $20`

	row := fromEntity(loc)
	oldVersion := loc.GetVersion()
	loc.IncrementVersion()

	result, err := r.getExecutor(ctx).ExecContext(ctx, query,
		row.ID, row.Code, row.Name, row.Type, row.Description,
		row.ParentID,
		row.Path, row.Level, row.Capacity, row.CurrentOccupancy, row.IsLimited,
		row.Width, row.Height, row.Depth, row.Status,
		row.IsPickable, row.IsPutawayable, row.Notes, loc.GetVersion(), oldVersion,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// Check if location exists to distinguish between not found vs version mismatch
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM warehouse_locations WHERE id = $1 AND deleted_at IS NULL)`
		
		// Use the executor with type assertion for Get method
		exec := r.getExecutor(ctx)
		if db, ok := exec.(*sqlx.DB); ok {
			err := db.Get(&exists, checkQuery, row.ID)
			if err != nil {
				return fmt.Errorf("failed to check location existence: %w", err)
			}
		} else if tx, ok := exec.(*sqlx.Tx); ok {
			err := tx.Get(&exists, checkQuery, row.ID)
			if err != nil {
				return fmt.Errorf("failed to check location existence: %w", err)
			}
		}

		if !exists {
			return location.ErrLocationNotFound
		}
		// Location exists but version doesn't match
		return location.ErrVersionMismatch
	}

	return nil
}

// Delete soft-deletes a location by ID
func (r *locationRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE warehouse_locations SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.getExecutor(ctx).ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("location not found")
	}

	return nil
}

// GetByID retrieves a location by ID
func (r *locationRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*location.Location, error) {
	query := `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE id = $1 AND deleted_at IS NULL`

	var row locationRow
	err := sqlx.GetContext(ctx, r.getExecutor(ctx), &row, query, id.String())
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("location not found")
	}
	if err != nil {
		return nil, err
	}

	return row.toEntity()
}

// GetByCode retrieves a location by code
func (r *locationRepository) GetByCode(ctx context.Context, code string) (*location.Location, error) {
	query := `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE code = $1 AND deleted_at IS NULL`

	var row locationRow
	err := sqlx.GetContext(ctx, r.getExecutor(ctx), &row, query, code)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("location not found")
	}
	if err != nil {
		return nil, err
	}

	return row.toEntity()
}

// List retrieves paginated locations
func (r *locationRepository) List(ctx context.Context, limit, offset int) ([]*location.Location, error) {
	query := `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE deleted_at IS NULL
		ORDER BY path, code
		LIMIT $1 OFFSET $2`

	var rows []locationRow
	err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, query, limit, offset)
	if err != nil {
		return nil, err
	}

	locations := make([]*location.Location, 0, len(rows))
	for _, row := range rows {
		loc, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// Count returns total number of locations
func (r *locationRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM warehouse_locations WHERE deleted_at IS NULL`
	var count int
	exec := r.getExecutor(ctx)
	if queryer, ok := exec.(interface{ QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row }); ok {
		err := queryer.QueryRowxContext(ctx, query).Scan(&count)
		return count, err
	}
	// Fallback for plain db
	err := r.db.QueryRowx(query).Scan(&count)
	return count, err
}

// ListByType retrieves locations by type
func (r *locationRepository) ListByType(ctx context.Context, locType location.LocationType, limit, offset int) ([]*location.Location, error) {
	query := `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE type = $1 AND deleted_at IS NULL
		ORDER BY path, code
		LIMIT $2 OFFSET $3`

	var rows []locationRow
	err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, query, string(locType), limit, offset)
	if err != nil {
		return nil, err
	}

	locations := make([]*location.Location, 0, len(rows))
	for _, row := range rows {
		loc, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// CountByType returns count of locations by type
func (r *locationRepository) CountByType(ctx context.Context, locType location.LocationType) (int, error) {
	query := `SELECT COUNT(*) FROM warehouse_locations WHERE type = $1 AND deleted_at IS NULL`
	var count int
	exec := r.getExecutor(ctx)
	if queryer, ok := exec.(interface{ QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row }); ok {
		err := queryer.QueryRowxContext(ctx, query, string(locType)).Scan(&count)
		return count, err
	}
	err := r.db.QueryRowx(query, string(locType)).Scan(&count)
	return count, err
}

// ListByParent retrieves direct children of a parent location
func (r *locationRepository) ListByParent(ctx context.Context, parentID uuidv7.UUID) ([]*location.Location, error) {
	query := `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY code`

	var rows []locationRow
	err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, query, parentID.String())
	if err != nil {
		return nil, err
	}

	locations := make([]*location.Location, 0, len(rows))
	for _, row := range rows {
		loc, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// ListChildren retrieves all descendants of a parent location
func (r *locationRepository) ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*location.Location, error) {
	query := `SELECT path FROM warehouse_locations WHERE id = $1 AND deleted_at IS NULL`
	var parentPath string
	var err error
	exec := r.getExecutor(ctx)
	if queryer, ok := exec.(interface{ QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row }); ok {
		err = queryer.QueryRowxContext(ctx, query, parentID.String()).Scan(&parentPath)
		if err != nil {
			return nil, err
		}
	} else {
		err = r.db.QueryRowx(query, parentID.String()).Scan(&parentPath)
		if err != nil {
			return nil, err
		}
	}

	query = `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE path LIKE $1 AND id != $2 AND deleted_at IS NULL
		ORDER BY path, code`

	var rows []locationRow
	err = sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, query, parentPath+"/%", parentID.String())
	if err != nil {
		return nil, err
	}

	locations := make([]*location.Location, 0, len(rows))
	for _, row := range rows {
		loc, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// ListByStatus retrieves locations by status
func (r *locationRepository) ListByStatus(ctx context.Context, status location.LocationStatus, limit, offset int) ([]*location.Location, error) {
	query := `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY path, code
		LIMIT $2 OFFSET $3`

	var rows []locationRow
	err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, query, string(status), limit, offset)
	if err != nil {
		return nil, err
	}

	locations := make([]*location.Location, 0, len(rows))
	for _, row := range rows {
		loc, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// CountByStatus returns count of locations by status
func (r *locationRepository) CountByStatus(ctx context.Context, status location.LocationStatus) (int, error) {
	query := `SELECT COUNT(*) FROM warehouse_locations WHERE status = $1 AND deleted_at IS NULL`
	var count int
	exec := r.getExecutor(ctx)
	if queryer, ok := exec.(interface{ QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row }); ok {
		err := queryer.QueryRowxContext(ctx, query, string(status)).Scan(&count)
		return count, err
	}
	err := r.db.QueryRowx(query, string(status)).Scan(&count)
	return count, err
}

// ListAvailable retrieves available locations
func (r *locationRepository) ListAvailable(ctx context.Context, limit, offset int) ([]*location.Location, error) {
	query := `
		SELECT id, code, name, type, description, parent_id, path, level,
		       capacity, current_occupancy, is_limited, width, height, depth,
		       status, is_pickable, is_putawayable, notes,
		       created_at, updated_at, version
		FROM warehouse_locations
		WHERE status = $1 
		  AND is_putawayable = true
		  AND (is_limited = false OR current_occupancy < capacity)
		  AND deleted_at IS NULL
		ORDER BY path, code
		LIMIT $2 OFFSET $3`

	var rows []locationRow
	err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, query, string(location.LocationStatusActive), limit, offset)
	if err != nil {
		return nil, err
	}

	locations := make([]*location.Location, 0, len(rows))
	for _, row := range rows {
		loc, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// CountAvailable returns count of available locations
func (r *locationRepository) CountAvailable(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM warehouse_locations 
		WHERE status = $1 
		  AND is_putawayable = true
		  AND (is_limited = false OR current_occupancy < capacity)
		  AND deleted_at IS NULL`
	var count int
	exec := r.getExecutor(ctx)
	if queryer, ok := exec.(interface{ QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row }); ok {
		err := queryer.QueryRowxContext(ctx, query, string(location.LocationStatusActive)).Scan(&count)
		return count, err
	}
	err := r.db.QueryRowx(query, string(location.LocationStatusActive)).Scan(&count)
	return count, err
}
