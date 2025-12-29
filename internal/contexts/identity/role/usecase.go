package role

import (
	"context"
	"errors"
	"fmt"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Additional errors for use cases
var (
	ErrRoleNameRequired    = errors.New("role name is required")
	ErrRoleDisplayRequired = errors.New("role display name is required")
	ErrRoleNameExists      = errors.New("role name already exists")
)

// IUseCase defines the business logic interface for Role operations
type IUseCase interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, name, displayName, description string) (*Role, error)

	// GetRole retrieves a role by ID
	GetRole(ctx context.Context, roleID uuidv7.UUID) (*Role, error)

	// GetRoleByName retrieves a role by name
	GetRoleByName(ctx context.Context, name string) (*Role, error)

	// UpdateRole updates an existing role
	UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*Role, error)

	// DeleteRole soft deletes a role (protects system roles)
	DeleteRole(ctx context.Context, roleID uuidv7.UUID) error

	// ListRoles retrieves all active roles with pagination
	ListRoles(ctx context.Context, limit, offset int) ([]*Role, int, error)

	// GetUserRoles retrieves all roles assigned to a user
	GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*Role, error)
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

// CreateRole creates a new role
func (uc *UseCase) CreateRole(ctx context.Context, name, displayName, description string) (*Role, error) {
	// Validation
	if name == "" {
		return nil, ErrRoleNameRequired
	}
	if displayName == "" {
		return nil, ErrRoleDisplayRequired
	}

	// Check if role name already exists
	exists, err := uc.repo.ExistsByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check role existence: %w", err)
	}
	if exists {
		return nil, ErrRoleNameExists
	}

	// Create role entity
	role, err := NewRole(name, displayName, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create role entity: %w", err)
	}

	// Validate
	if err := role.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to repository
	if err := uc.repo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to save role: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (uc *UseCase) GetRole(ctx context.Context, roleID uuidv7.UUID) (*Role, error) {
	role, err := uc.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

// GetRoleByName retrieves a role by name
func (uc *UseCase) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	if name == "" {
		return nil, ErrRoleNameRequired
	}

	role, err := uc.repo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}
	return role, nil
}

// UpdateRole updates an existing role
func (uc *UseCase) UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*Role, error) {
	// Validation
	if displayName == "" {
		return nil, ErrRoleDisplayRequired
	}

	// Get existing role
	role, err := uc.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return nil, ErrRoleNotFound
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
	if err := uc.repo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole soft deletes a role (protects system roles)
func (uc *UseCase) DeleteRole(ctx context.Context, roleID uuidv7.UUID) error {
	// Get role to check if it's a system role
	role, err := uc.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Check if can delete (system role protection)
	if err := role.CanDelete(); err != nil {
		return err
	}

	// Soft delete
	if err := uc.repo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// ListRoles retrieves all active roles with pagination
func (uc *UseCase) ListRoles(ctx context.Context, limit, offset int) ([]*Role, int, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	roles, total, err := uc.repo.ListRoles(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, total, nil
}

// GetUserRoles retrieves all roles assigned to a user
func (uc *UseCase) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*Role, error) {
	roles, err := uc.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return roles, nil
}
