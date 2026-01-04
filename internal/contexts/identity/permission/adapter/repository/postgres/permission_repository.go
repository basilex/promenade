package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	ID          uuidv7.UUID  `db:"id"`
	Name        string       `db:"name"`
	Resource    string       `db:"resource"`
	Action      string       `db:"action"`
	Description string       `db:"description"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
	DeletedAt   sql.NullTime `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *permissionRow) toEntity() (*permission.Permission, error) {
	perm := &permission.Permission{
		Name:        r.Name,
		Resource:    r.Resource,
		Action:      r.Action,
		Description: r.Description,
	}
	// Initialize BaseAggregate fields
	perm.ID = r.ID
	perm.CreatedAt = r.CreatedAt
	perm.UpdatedAt = r.UpdatedAt
	if r.DeletedAt.Valid {
		deletedTime := r.DeletedAt.Time
		perm.DeletedAt = &deletedTime
	}

	return perm, nil
}

// fromEntity converts domain entity to database row
func fromPermissionEntity(p *permission.Permission) *permissionRow {
	row := &permissionRow{
		ID:          p.GetID(),
		Name:        p.Name,
		Resource:    p.Resource,
		Action:      p.Action,
		Description: p.Description,
		CreatedAt:   p.GetCreatedAt(),
		UpdatedAt:   p.GetUpdatedAt(),
	}

	if p.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *p.DeletedAt, Valid: true}
	}

	return row
}

// Create creates a new permission in the database
func (r *permissionRepository) Create(ctx context.Context, perm *permission.Permission) error {
	row := fromPermissionEntity(perm)

	query := `
		INSERT INTO identity_permissions (id, resource, action, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := r.Exec(ctx, query, row.ID, row.Resource, row.Action, row.Description)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*permission.Permission, error) {
	var row permissionRow

	query := `
		SELECT id, name, resource, action, description, created_at, updated_at, deleted_at
		FROM identity_permissions
		WHERE id = $1 AND deleted_at IS NULL
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
		SELECT id, name, resource, action, description, created_at, updated_at, deleted_at
		FROM identity_permissions
		WHERE name = $1 AND deleted_at IS NULL
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
		SET description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
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

// Delete soft deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE identity_permissions
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
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
			WHERE name = $1 AND deleted_at IS NULL
		)
	`

	err := r.Get(ctx, &exists, query, name)
	if err != nil {
		return false, fmt.Errorf("failed to check permission existence: %w", err)
	}

	return exists, nil
}

// ListPermissions lists all permissions with pagination
func (r *permissionRepository) ListPermissions(ctx context.Context, limit, offset int) ([]*permission.Permission, int, error) {
	var rows []permissionRow

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM identity_permissions
		WHERE deleted_at IS NULL
	`
	err := r.Get(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at, deleted_at
		FROM identity_permissions
		WHERE deleted_at IS NULL
		ORDER BY resource, action
		LIMIT $1 OFFSET $2
	`

	err = r.Select(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	permissions := make([]*permission.Permission, len(rows))
	for i, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		permissions[i] = entity
	}

	return permissions, total, nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*permission.Permission, error) {
	var rows []permissionRow

	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at, p.deleted_at
		FROM identity_permissions p
		INNER JOIN identity_role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1 AND p.deleted_at IS NULL
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
