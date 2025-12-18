package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type roleRepository struct {
	*BaseRepository
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sqlx.DB) *roleRepository {
	return &roleRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	query := `
		INSERT INTO roles (id, name, display_name, description, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		role.ID,
		role.Name,
		role.DisplayName,
		role.Description,
		role.IsSystem,
		role.CreatedAt,
		role.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at
		FROM roles
		WHERE id = $1
	`
	var role entity.Role
	err := r.Get(ctx, &role, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at
		FROM roles
		WHERE name = $1
	`
	var role entity.Role
	err := r.Get(ctx, &role, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return &role, nil
}

func (r *roleRepository) List(ctx context.Context) ([]*entity.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at
		FROM roles
		ORDER BY name
	`
	var roles []*entity.Role
	err := r.Select(ctx, &roles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	query := `
		UPDATE roles
		SET name = $2, display_name = $3, description = $4, updated_at = $5
		WHERE id = $1
	`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		role.ID,
		role.Name,
		role.DisplayName,
		role.Description,
		role.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
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

func (r *roleRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM roles WHERE id = $1 AND is_system = FALSE`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
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

func (r *roleRepository) AddPermission(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	err := r.Exec(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to add permission to role: %w", err)
	}
	return nil
}

func (r *roleRepository) RemovePermission(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
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

func (r *roleRepository) GetPermissions(ctx context.Context, roleID uuidv7.UUID) ([]*entity.Permission, error) {
	query := `
		SELECT p.id, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`
	var permissions []*entity.Permission
	err := r.Select(ctx, &permissions, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return permissions, nil
}

func (r *roleRepository) SyncPermissions(ctx context.Context, roleID uuidv7.UUID, permissionIDs []uuidv7.UUID) error {
	// Delete all existing permissions
	deleteQuery := `DELETE FROM role_permissions WHERE role_id = $1`
	err := r.Exec(ctx, deleteQuery, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete existing permissions: %w", err)
	}

	// Insert new permissions
	if len(permissionIDs) > 0 {
		insertQuery := `
			INSERT INTO role_permissions (role_id, permission_id)
			VALUES ($1, unnest($2::uuid[]))
		`
		err = r.Exec(ctx, insertQuery, roleID, pq.Array(permissionIDs))
		if err != nil {
			return fmt.Errorf("failed to insert new permissions: %w", err)
		}
	}

	return nil
}

func (r *roleRepository) AssignToUser(ctx context.Context, userRole *entity.UserRole) error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, role_id) DO UPDATE
		SET assigned_at = EXCLUDED.assigned_at,
		    assigned_by = EXCLUDED.assigned_by,
		    expires_at = EXCLUDED.expires_at
	`
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		userRole.UserID,
		userRole.RoleID,
		userRole.AssignedAt,
		userRole.AssignedBy,
		userRole.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}
	return nil
}

func (r *roleRepository) RemoveFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
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

func (r *roleRepository) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error) {
	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`
	var roles []*entity.Role
	err := r.Select(ctx, &roles, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) GetUserActiveRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error) {
	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		ORDER BY r.name
	`
	var roles []*entity.Role
	err := r.Select(ctx, &roles, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user active roles: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) GetUsersWithRole(ctx context.Context, roleID uuidv7.UUID) ([]uuidv7.UUID, error) {
	query := `
		SELECT user_id
		FROM user_roles
		WHERE role_id = $1
		ORDER BY assigned_at DESC
	`
	var userIDs []uuidv7.UUID
	err := r.Select(ctx, &userIDs, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with role: %w", err)
	}
	return userIDs, nil
}

func (r *roleRepository) CreateMany(ctx context.Context, roles []*entity.Role) error {
	if len(roles) == 0 {
		return nil
	}

	query := `
		INSERT INTO roles (id, name, display_name, description, is_system, created_at, updated_at)
		VALUES (:id, :name, :display_name, :description, :is_system, :created_at, :updated_at)
	`
	err := r.NamedExec(ctx, query, roles)
	if err != nil {
		return fmt.Errorf("failed to create many roles: %w", err)
	}
	return nil
}
