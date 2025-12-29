package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// permissionRepository implements permission.IRepository for PostgreSQL
type permissionRepository struct {
	*BaseRepository
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *sqlx.DB) permission.IRepository {
	return &permissionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// permissionRow represents database row structure for identity_permissions table
type permissionRow struct {
	ID          uuidv7.UUID `db:"id"`
	Name        string      `db:"name"`
	Resource    string      `db:"resource"`
	Action      string      `db:"action"`
	Description string      `db:"description"`
	CreatedAt   string      `db:"created_at"`
}

// toEntity converts database row to domain entity
func (r *permissionRow) toEntity() (*permission.Permission, error) {
	return &permission.Permission{
		ID:          r.ID,
		Name:        r.Name,
		Resource:    r.Resource,
		Action:      r.Action,
		Description: r.Description,
	}, nil
}

// fromEntity converts domain entity to database row
func fromPermissionEntity(p *permission.Permission) *permissionRow {
	return &permissionRow{
		ID:          p.ID,
		Name:        p.Name,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
	}
}

// Create creates a new permission in the database
func (r *permissionRepository) Create(ctx context.Context, perm *permission.Permission) error {
	row := fromPermissionEntity(perm)

	query := `
		INSERT INTO identity_permissions (id, name, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
	`

	_, err := r.Exec(ctx, query, row.ID, row.Name, row.Resource, row.Action, row.Description)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*permission.Permission, error) {
	var row permissionRow

	query := `
		SELECT id, name, resource, action, description, created_at
		FROM identity_permissions
		WHERE id = $1
	`

	err := r.Get(ctx, &row, query, id)
	if err == sql.ErrNoRows {
		return nil, permission.ErrPermissionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return row.toEntity()
}

// GetByName retrieves a permission by name (resource:action)
func (r *permissionRepository) GetByName(ctx context.Context, name string) (*permission.Permission, error) {
	var row permissionRow

	query := `
		SELECT id, name, resource, action, description, created_at
		FROM identity_permissions
		WHERE name = $1
	`

	err := r.Get(ctx, &row, query, name)
	if err == sql.ErrNoRows {
		return nil, permission.ErrPermissionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing permission
func (r *permissionRepository) Update(ctx context.Context, perm *permission.Permission) error {
	row := fromPermissionEntity(perm)

	query := `
		UPDATE identity_permissions
		SET description = $2
		WHERE id = $1
	`

	result, err := r.Exec(ctx, query, row.ID, row.Description)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return permission.ErrPermissionNotFound
	}

	return nil
}

// Delete deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		DELETE FROM identity_permissions
		WHERE id = $1
	`

	result, err := r.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return permission.ErrPermissionNotFound
	}

	return nil
}

// ExistsByName checks if a permission with the given name exists
func (r *permissionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM identity_permissions
			WHERE name = $1
		)
	`

	err := r.Get(ctx, &exists, query, name)
	if err != nil {
		return false, fmt.Errorf("failed to check permission existence: %w", err)
	}

	return exists, nil
}

// ListPermissions lists all permissions
func (r *permissionRepository) ListPermissions(ctx context.Context) ([]*permission.Permission, error) {
	var rows []permissionRow

	query := `
		SELECT id, name, resource, action, description, created_at
		FROM identity_permissions
		ORDER BY resource, action
	`

	err := r.Select(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		permissions[i] = entity
	}

	return permissions, nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*permission.Permission, error) {
	var rows []permissionRow

	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at
		FROM identity_permissions p
		INNER JOIN identity_role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`

	err := r.Select(ctx, &rows, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		permissions[i] = entity
	}

	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *permissionRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	query := `
		INSERT INTO identity_role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`

	_, err := r.Exec(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *permissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	query := `
		DELETE FROM identity_role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`

	_, err := r.Exec(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	return nil
}
