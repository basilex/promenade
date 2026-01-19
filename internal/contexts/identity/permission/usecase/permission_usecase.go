package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	permissionerrors "github.com/basilex/promenade/internal/contexts/identity/permission"
	"github.com/basilex/promenade/internal/contexts/identity/permission/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/permission/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IPermissionUseCase defines the business logic interface for Permission operations
type IPermissionUseCase interface {
	// CreatePermission creates a new permission
	CreatePermission(ctx context.Context, resource, action, description string) (*aggregate.Permission, error)

	// GetPermission retrieves a permission by ID
	GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*aggregate.Permission, error)

	// GetPermissionByName retrieves a permission by name (resource:action)
	GetPermissionByName(ctx context.Context, name string) (*aggregate.Permission, error)

	// UpdatePermission updates an existing permission
	UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*aggregate.Permission, error)

	// DeletePermission soft deletes a permission
	DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error

	// ListPermissions retrieves all active permissions with pagination
	ListPermissions(ctx context.Context, limit, offset int) ([]*aggregate.Permission, int, error)

	// GetRolePermissions retrieves all permissions assigned to a role
	GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*aggregate.Permission, error)
}

// PermissionUseCase implements IPermissionUseCase
type PermissionUseCase struct {
	repo repository.IPermissionRepository
}

// NewPermissionUseCase creates a new PermissionUseCase
func NewPermissionUseCase(repo repository.IPermissionRepository) IPermissionUseCase {
	return &PermissionUseCase{
		repo: repo,
	}
}

// CreatePermission creates a new permission
func (u *PermissionUseCase) CreatePermission(ctx context.Context, resource, action, description string) (*aggregate.Permission, error) {
	// Validation
	if resource == "" {
		return nil, permissionerrors.ErrResourceRequired
	}
	if action == "" {
		return nil, permissionerrors.ErrActionRequired
	}

	// Create permission entity
	perm, err := aggregate.NewPermission(resource, action, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission entity: %w", err)
	}

	// Check if permission name already exists
	name := perm.Name
	exists, err := u.repo.ExistsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission existence: %w", err)
	}
	if exists {
		return nil, permissionerrors.ErrPermissionNameExists
	}

	// Validate
	if err := perm.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to repository
	if err := u.repo.Create(ctx, perm); err != nil {
		return nil, fmt.Errorf("failed to save permission: %w", err)
	}

	return perm, nil
}

// GetPermission retrieves a permission by ID
func (u *PermissionUseCase) GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*aggregate.Permission, error) {
	perm, err := u.repo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, permissionerrors.ErrPermissionNotFound) {
			return nil, permissionerrors.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return perm, nil
}

// GetPermissionByName retrieves a permission by name (resource:action)
func (u *PermissionUseCase) GetPermissionByName(ctx context.Context, name string) (*aggregate.Permission, error) {
	if name == "" {
		return nil, permissionerrors.ErrResourceRequired
	}

	perm, err := u.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, permissionerrors.ErrPermissionNotFound) {
			return nil, permissionerrors.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}
	return perm, nil
}

// UpdatePermission updates an existing permission
func (u *PermissionUseCase) UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*aggregate.Permission, error) {
	// Get existing permission
	perm, err := u.repo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, permissionerrors.ErrPermissionNotFound) {
			return nil, permissionerrors.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	// Update description
	perm.Description = description
	perm.UpdatedAt = time.Now()

	// Validate
	if err := perm.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save changes
	if err := u.repo.Update(ctx, perm); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return perm, nil
}

// DeletePermission soft deletes a permission
func (u *PermissionUseCase) DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error {
	// Check if permission exists
	_, err := u.repo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, permissionerrors.ErrPermissionNotFound) {
			return permissionerrors.ErrPermissionNotFound
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// Soft delete
	if err := u.repo.Delete(ctx, permissionID); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// ListPermissions retrieves all active permissions with pagination
func (u *PermissionUseCase) ListPermissions(ctx context.Context, limit, offset int) ([]*aggregate.Permission, int, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	perms, total, err := u.repo.ListPermissions(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	return perms, total, nil
}

// GetRolePermissions retrieves all permissions assigned to a role
func (u *PermissionUseCase) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*aggregate.Permission, error) {
	perms, err := u.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return perms, nil
}
