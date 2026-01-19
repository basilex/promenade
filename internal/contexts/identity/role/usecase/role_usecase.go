package usecase

import (
	"context"
	"errors"
	"fmt"

	roleerrors "github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/internal/contexts/identity/role/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/role/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRoleUseCase defines the business logic interface for Role operations
type IRoleUseCase interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, name, displayName, description string) (*aggregate.Role, error)

	// GetRole retrieves a role by ID
	GetRole(ctx context.Context, roleID uuidv7.UUID) (*aggregate.Role, error)

	// GetRoleByName retrieves a role by name
	GetRoleByName(ctx context.Context, name string) (*aggregate.Role, error)

	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*aggregate.Role, error)

	// DeleteRole soft deletes a role (protects system roles)
	DeleteRole(ctx context.Context, roleID uuidv7.UUID) error

	// ListRoles retrieves all active roles with pagination
	ListRoles(ctx context.Context, limit, offset int) ([]*aggregate.Role, int, error)

	// GetUserRoles retrieves all roles assigned to a user
	GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*aggregate.Role, error)
}

// RoleUseCase implements IRoleUseCase
type RoleUseCase struct {
	repo repository.IRoleRepository
}

// NewRoleUseCase creates a new RoleUseCase
func NewRoleUseCase(repo repository.IRoleRepository) IRoleUseCase {
	return &RoleUseCase{
		repo: repo,
	}
}

// CreateRole creates a new role
func (u *RoleUseCase) CreateRole(ctx context.Context, name, displayName, description string) (*aggregate.Role, error) {
	// Validation
	if name == "" {
		return nil, roleerrors.ErrRoleNameRequired
	}
	if displayName == "" {
		return nil, roleerrors.ErrRoleDisplayRequired
	}

	// Check if role name already exists
	exists, err := u.repo.ExistsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check role existence: %w", err)
	}
	if exists {
		return nil, roleerrors.ErrRoleNameExists
	}

	// Create role entity
	role, err := aggregate.NewRole(name, displayName, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create role entity: %w", err)
	}

	// Validate
	if err := role.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to repository
	if err := u.repo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to save role: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (u *RoleUseCase) GetRole(ctx context.Context, roleID uuidv7.UUID) (*aggregate.Role, error) {
	role, err := u.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, roleerrors.ErrRoleNotFound) {
			return nil, roleerrors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

// GetRoleByName retrieves a role by name
func (u *RoleUseCase) GetRoleByName(ctx context.Context, name string) (*aggregate.Role, error) {
	if name == "" {
		return nil, roleerrors.ErrRoleNameRequired
	}

	role, err := u.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, roleerrors.ErrRoleNotFound) {
			return nil, roleerrors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}
	return role, nil
}

// UpdateRole updates an existing role
func (u *RoleUseCase) UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*aggregate.Role, error) {
	// Validation
	if displayName == "" {
		return nil, roleerrors.ErrRoleDisplayRequired
	}

	// Get existing role
	role, err := u.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, roleerrors.ErrRoleNotFound) {
			return nil, roleerrors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Update fields
	if err := role.UpdateDisplayName(displayName); err != nil {
		return nil, fmt.Errorf("failed to update display name: %w", err)
	}
	role.UpdateDescription(description)

	// Validate
	if err := role.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save changes
	if err := u.repo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole soft deletes a role (protects system roles)
func (u *RoleUseCase) DeleteRole(ctx context.Context, roleID uuidv7.UUID) error {
	// Get role to check if it's a system role
	role, err := u.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, roleerrors.ErrRoleNotFound) {
			return roleerrors.ErrRoleNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Check if can delete (system role protection)
	if err := role.CanDelete(); err != nil {
		return err
	}

	// Soft delete
	if err := u.repo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// ListRoles retrieves all active roles with pagination
func (u *RoleUseCase) ListRoles(ctx context.Context, limit, offset int) ([]*aggregate.Role, int, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	roles, total, err := u.repo.ListRoles(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, total, nil
}

// GetUserRoles retrieves all roles assigned to a user
func (u *RoleUseCase) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*aggregate.Role, error) {
	roles, err := u.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}
