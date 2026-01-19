package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	roleerrors "github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/internal/contexts/identity/role/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/role/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// roleRepository implements repository.IRoleRepository for PostgreSQL
type roleRepository struct {
	db *sqlx.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sqlx.DB) repository.IRoleRepository {
	return &roleRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *roleRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *roleRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *roleRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning rows
func (r *roleRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows
func (r *roleRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

// roleRow represents database row structure for identity_roles table
type roleRow struct {
	ID          uuidv7.UUID  `db:"id"`
	Name        string       `db:"name"`
	DisplayName string       `db:"display_name"`
	Description string       `db:"description"`
	IsSystem    bool         `db:"is_system"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *roleRow) toEntity() (*aggregate.Role, error) {
	role := &aggregate.Role{
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		IsSystem:    r.IsSystem,
	}
	// Initialize BaseAggregate fields
	role.ID = r.ID
	role.CreatedAt = r.CreatedAt
	role.UpdatedAt = r.UpdatedAt
	if r.DeletedAt.Valid {
		deletedTime := r.DeletedAt.Time
		role.DeletedAt = &deletedTime
	}

	return role, nil
}

// fromEntity converts domain entity to database row
func fromRoleEntity(r *aggregate.Role) *roleRow {
	row := &roleRow{
		ID:          r.GetID(),
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.GetCreatedAt(),
		UpdatedAt:   r.GetUpdatedAt(),
	}

	if r.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *r.DeletedAt, Valid: true}
	}

	return row
}

// Create creates a new role in the database
func (r *roleRepository) Create(ctx context.Context, role *aggregate.Role) error {
	row := fromRoleEntity(role)

	query := `
		INSERT INTO identity_roles (id, name, display_name, description, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := r.Exec(ctx, query, row.ID, row.Name, row.DisplayName, row.Description, row.IsSystem)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// GetByID retrieves a role by ID
func (r *roleRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Role, error) {
	var row roleRow

	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at, deleted_at
		FROM identity_roles
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.Get(ctx, &row, query, id)
	if err == sql.ErrNoRows {
		return nil, roleerrors.ErrRoleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return row.toEntity()
}

// GetByName retrieves a role by name
func (r *roleRepository) GetByName(ctx context.Context, name string) (*aggregate.Role, error) {
	var row roleRow

	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at, deleted_at
		FROM identity_roles
		WHERE name = $1 AND deleted_at IS NULL
	`

	err := r.Get(ctx, &row, query, name)
	if err == sql.ErrNoRows {
		return nil, roleerrors.ErrRoleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing role
func (r *roleRepository) Update(ctx context.Context, role *aggregate.Role) error {
	row := fromRoleEntity(role)

	query := `
		UPDATE identity_roles
		SET display_name = $2, description = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.Exec(ctx, query, row.ID, row.DisplayName, row.Description)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return roleerrors.ErrRoleNotFound
	}

	return nil
}

// Delete soft deletes a role
func (r *roleRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	// First check if role exists and get its is_system flag
	var row roleRow
	checkQuery := `SELECT id, is_system, deleted_at FROM identity_roles WHERE id = $1`
	err := r.Get(ctx, &row, checkQuery, id)
	if err == sql.ErrNoRows {
		return roleerrors.ErrRoleNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to check role: %w", err)
	}

	// Check if already deleted
	if row.DeletedAt.Valid {
		return roleerrors.ErrRoleNotFound
	}

	// Check if system role
	if row.IsSystem {
		return roleerrors.ErrCannotDeleteSystem
	}

	// Perform soft delete
	query := `
		UPDATE identity_roles
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL AND is_system = false
	`

	result, err := r.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// This shouldn't happen given our checks above, but handle defensively
		return roleerrors.ErrCannotDeleteSystem
	}

	return nil
}

// ExistsByName checks if a role with the given name exists
func (r *roleRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM identity_roles
			WHERE name = $1 AND deleted_at IS NULL
		)
	`

	err := r.Get(ctx, &exists, query, name)
	if err != nil {
		return false, fmt.Errorf("failed to check role existence: %w", err)
	}

	return exists, nil
}

// ListRoles lists all roles with pagination
func (r *roleRepository) ListRoles(ctx context.Context, limit, offset int) ([]*aggregate.Role, int, error) {
	var rows []roleRow

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM identity_roles
		WHERE deleted_at IS NULL
	`
	err := r.Get(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at, deleted_at
		FROM identity_roles
		WHERE deleted_at IS NULL
		ORDER BY name
		LIMIT $1 OFFSET $2
	`

	err = r.Select(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	roles := make([]*aggregate.Role, len(rows))
	for i, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		roles[i] = entity
	}

	return roles, total, nil
}

// GetUserRoles retrieves all roles for a user
func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*aggregate.Role, error) {
	var rows []roleRow

	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at, r.deleted_at
		FROM identity_roles r
		INNER JOIN identity_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.name
	`

	err := r.Select(ctx, &rows, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roles := make([]*aggregate.Role, len(rows))
	for i, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		roles[i] = entity
	}

	return roles, nil
}

// AssignRoleToUser assigns a role to a user
func (r *roleRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuidv7.UUID, assignedBy *uuidv7.UUID) error {
	query := `
		INSERT INTO identity_user_roles (user_id, role_id, assigned_by, assigned_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := r.Exec(ctx, query, userID, roleID, assignedBy)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error {
	query := `
		DELETE FROM identity_user_roles
		WHERE user_id = $1 AND role_id = $2
	`

	_, err := r.Exec(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	return nil
}
