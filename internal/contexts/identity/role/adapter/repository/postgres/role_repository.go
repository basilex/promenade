package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	roleentity "github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// roleRepository implements roleentity.IRepository for PostgreSQL
type roleRepository struct {
	*BaseRepository
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sqlx.DB) roleentity.IRepository {
	return &roleRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// roleRow represents database row structure for identity_roles table
type roleRow struct {
	ID          uuidv7.UUID  `db:"id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	IsSystem    bool         `db:"is_system"`
	CreatedAt   string       `db:"created_at"`
	UpdatedAt   string       `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *roleRow) toEntity() (*roleentity.Role, error) {
	return &roleentity.Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
	}, nil
}

// fromEntity converts domain entity to database row
func fromRoleEntity(r *roleentity.Role) *roleRow {
	row := &roleRow{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
	}

	if r.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *r.DeletedAt, Valid: true}
	}

	return row
}

// Create creates a new role in the database
func (r *roleRepository) Create(ctx context.Context, role *roleentity.Role) error {
	row := fromRoleEntity(role)

	query := `
		INSERT INTO identity_roles (id, name, description, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := r.Exec(ctx, query, row.ID, row.Name, row.Description, row.IsSystem)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// GetByID retrieves a role by ID
func (r *roleRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*roleentity.Role, error) {
	var row roleRow

	query := `
		SELECT id, name, description, is_system, created_at, updated_at, deleted_at
		FROM identity_roles
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.Get(ctx, &row, query, id)
	if err == sql.ErrNoRows {
		return nil, roleentity.ErrRoleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return row.toEntity()
}

// GetByName retrieves a role by name
func (r *roleRepository) GetByName(ctx context.Context, name string) (*roleentity.Role, error) {
	var row roleRow

	query := `
		SELECT id, name, description, is_system, created_at, updated_at
		FROM identity_roles
		WHERE name = $1
	`

	err := r.Get(ctx, &row, query, name)
	if err == sql.ErrNoRows {
		return nil, roleentity.ErrRoleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing role
func (r *roleRepository) Update(ctx context.Context, role *roleentity.Role) error {
	row := fromRoleEntity(role)

	query := `
		UPDATE identity_roles
		SET description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.Exec(ctx, query, row.ID, row.Description)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return roleentity.ErrRoleNotFound
	}

	return nil
}

// Delete soft deletes a role
func (r *roleRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
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
		return roleentity.ErrCannotDeleteSystem
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

// ListRoles lists all roles
func (r *roleRepository) ListRoles(ctx context.Context) ([]*roleentity.Role, error) {
	var rows []roleRow

	query := `
		SELECT id, name, description, is_system, created_at, updated_at, deleted_at
		FROM identity_roles
		WHERE deleted_at IS NULL
		ORDER BY name
	`

	err := r.Select(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	roles := make([]*roleentity.Role, len(rows))
	for i, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		roles[i] = entity
	}

	return roles, nil
}

// GetUserRoles retrieves all roles for a user
func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*roleentity.Role, error) {
	var rows []roleRow

	query := `
		SELECT r.id, r.name, r.description, r.is_system, r.created_at, r.updated_at
		FROM identity_roles r
		INNER JOIN identity_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	err := r.Select(ctx, &rows, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roles := make([]*roleentity.Role, len(rows))
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
