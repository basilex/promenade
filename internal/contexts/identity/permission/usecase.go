package permission

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Additional errors for use cases
var (
	ErrPermissionNameExists = errors.New("permission name already exists")
	ErrResourceRequired     = errors.New("resource is required")
	ErrActionRequired       = errors.New("action is required")
	ErrInvalidAction        = errors.New("invalid action format")
)

// IUseCase defines the business logic interface for Permission operations
type IUseCase interface {
	// CreatePermission creates a new permission
	CreatePermission(ctx context.Context, resource, action, description string) (*Permission, error)

	// GetPermission retrieves a permission by ID
	GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*Permission, error)

	// GetPermissionByName retrieves a permission by name (resource:action)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)

	// UpdatePermission updates an existing permission
	UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*Permission, error)

	// DeletePermission soft deletes a permission
	DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error

	// ListPermissions retrieves all active permissions with pagination
	ListPermissions(ctx context.Context, limit, offset int) ([]*Permission, int, error)

	// GetRolePermissions retrieves all permissions assigned to a role
	GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*Permission, error)
}

// UseCase implements IUseCase
type UseCase struct {
	repo IRepository
}

// NewUseCase creates a new UseCase
func NewUseCase(repo IRepository) IUseCase {
	return &UseCase{
		repo: repo,
	}
}

// CreatePermission creates a new permission
func (uc *UseCase) CreatePermission(ctx context.Context, resource, action, description string) (*Permission, error) {
	// Validation
	if resource == "" {
		return nil, ErrResourceRequired
	}
	if action == "" {
		return nil, ErrActionRequired
	}

	// Create permission entity
	perm, err := NewPermission(resource, action, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission entity: %w", err)
	}

	// Check if permission name already exists
	name := perm.Name
	exists, err := uc.repo.ExistsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission existence: %w", err)
	}
	if exists {
		return nil, ErrPermissionNameExists
	}

	// Validate
	if err := perm.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to repository
	if err := uc.repo.Create(ctx, perm); err != nil {
		return nil, fmt.Errorf("failed to save permission: %w", err)
	}

	return perm, nil
}

// GetPermission retrieves a permission by ID
func (uc *UseCase) GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*Permission, error) {
	perm, err := uc.repo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return perm, nil
}

// GetPermissionByName retrieves a permission by name (resource:action)
func (uc *UseCase) GetPermissionByName(ctx context.Context, name string) (*Permission, error) {
	if name == "" {
		return nil, ErrResourceRequired
	}

	perm, err := uc.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}
	return perm, nil
}

// UpdatePermission updates an existing permission
func (uc *UseCase) UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*Permission, error) {
	// Get existing permission
	perm, err := uc.repo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			return nil, ErrPermissionNotFound
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
	if err := uc.repo.Update(ctx, perm); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return perm, nil
}

// DeletePermission soft deletes a permission
func (uc *UseCase) DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error {
	// Check if permission exists
	_, err := uc.repo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, ErrPermissionNotFound) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// Soft delete
	if err := uc.repo.Delete(ctx, permissionID); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// ListPermissions retrieves all active permissions with pagination
func (uc *UseCase) ListPermissions(ctx context.Context, limit, offset int) ([]*Permission, int, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	perms, total, err := uc.repo.ListPermissions(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	return perms, total, nil
}

// GetRolePermissions retrieves all permissions assigned to a role
func (uc *UseCase) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*Permission, error) {
	perms, err := uc.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return perms, nil
}
