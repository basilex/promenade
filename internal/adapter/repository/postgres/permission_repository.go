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

type permissionRepository struct {
	*BaseRepository
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *sqlx.DB) *permissionRepository {
	return &permissionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *permissionRepository) Create(ctx context.Context, permission *entity.Permission) error {
	query := `
		INSERT INTO permissions (id, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		permission.ID,
		permission.Resource,
		permission.Action,
		permission.Description,
		permission.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

func (r *permissionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Permission, error) {
	query := `
		SELECT id, resource, action, description, created_at
		FROM permissions
		WHERE id = $1
	`
	var permission entity.Permission
	err := r.Get(ctx, &permission, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return &permission, nil
}

func (r *permissionRepository) GetByResourceAction(ctx context.Context, resource, action string) (*entity.Permission, error) {
	query := `
		SELECT id, resource, action, description, created_at
		FROM permissions
		WHERE resource = $1 AND action = $2
	`
	var permission entity.Permission
	err := r.Get(ctx, &permission, query, resource, action)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return &permission, nil
}

func (r *permissionRepository) List(ctx context.Context) ([]*entity.Permission, error) {
	query := `
		SELECT id, resource, action, description, created_at
		FROM permissions
		ORDER BY resource, action
	`
	var permissions []*entity.Permission
	err := r.Select(ctx, &permissions, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return permissions, nil
}

func (r *permissionRepository) Update(ctx context.Context, permission *entity.Permission) error {
	query := `
		UPDATE permissions
		SET resource = $2, action = $3, description = $4
		WHERE id = $1
	`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query,
		permission.ID,
		permission.Resource,
		permission.Action,
		permission.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
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

func (r *permissionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	executor := r.getExecutor(ctx)
	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
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

func (r *permissionRepository) CreateMany(ctx context.Context, permissions []*entity.Permission) error {
	if len(permissions) == 0 {
		return nil
	}

	query := `
		INSERT INTO permissions (id, resource, action, description, created_at)
		VALUES (:id, :resource, :action, :description, :created_at)
	`
	err := r.NamedExec(ctx, query, permissions)
	if err != nil {
		return fmt.Errorf("failed to create many permissions: %w", err)
	}
	return nil
}

func (r *permissionRepository) GetByIDs(ctx context.Context, ids []uuidv7.UUID) ([]*entity.Permission, error) {
	if len(ids) == 0 {
		return []*entity.Permission{}, nil
	}

	query := `
		SELECT id, resource, action, description, created_at
		FROM permissions
		WHERE id = ANY($1)
		ORDER BY resource, action
	`
	var permissions []*entity.Permission
	err := r.Select(ctx, &permissions, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by IDs: %w", err)
	}
	return permissions, nil
}

func (r *permissionRepository) FindByResource(ctx context.Context, resource string) ([]*entity.Permission, error) {
	query := `
		SELECT id, resource, action, description, created_at
		FROM permissions
		WHERE resource = $1
		ORDER BY action
	`
	var permissions []*entity.Permission
	err := r.Select(ctx, &permissions, query, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by resource: %w", err)
	}
	return permissions, nil
}
