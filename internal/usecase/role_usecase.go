package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var (
	ErrRoleAlreadyExists      = errors.New("role with this name already exists")
	ErrCannotDeleteSystemRole = errors.New("cannot delete system role")
	ErrRoleNotFound           = errors.New("role not found")
)

type IRoleUseCase interface {
	// Role management
	CreateRole(ctx context.Context, name, displayName, description string) (*entity.Role, error)
	GetRole(ctx context.Context, roleID uuidv7.UUID) (*entity.Role, error)
	GetRoleByName(ctx context.Context, name string) (*entity.Role, error)
	ListRoles(ctx context.Context) ([]*entity.Role, error)
	UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*entity.Role, error)
	DeleteRole(ctx context.Context, roleID uuidv7.UUID) error

	// Permission management for roles
	AddPermissionToRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*entity.Permission, error)
	SyncRolePermissions(ctx context.Context, roleID uuidv7.UUID, permissionIDs []uuidv7.UUID) error

	// User role assignment
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuidv7.UUID, expiresAt *time.Time) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error
	GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error)
	GetUserActiveRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error)
	GetUsersWithRole(ctx context.Context, roleID uuidv7.UUID) ([]uuidv7.UUID, error)

	// User permissions (aggregated from roles)
	GetUserPermissions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Permission, error)
	HasPermission(ctx context.Context, userID uuidv7.UUID, permission string) (bool, error)
	HasAnyPermission(ctx context.Context, userID uuidv7.UUID, permissions []string) (bool, error)
	HasAllPermissions(ctx context.Context, userID uuidv7.UUID, permissions []string) (bool, error)
}

type roleUseCase struct {
	roleRepo       repository.IRoleRepository
	permissionRepo repository.IPermissionRepository
}

func NewRoleUseCase(
	roleRepo repository.IRoleRepository,
	permissionRepo repository.IPermissionRepository,
) IRoleUseCase {
	return &roleUseCase{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

func (uc *roleUseCase) CreateRole(ctx context.Context, name, displayName, description string) (*entity.Role, error) {
	// Check if role already exists
	existingRole, err := uc.roleRepo.GetByName(ctx, name)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing role: %w", err)
	}
	if existingRole != nil {
		return nil, ErrRoleAlreadyExists
	}

	now := time.Now()
	role := &entity.Role{
		ID:          uuidv7.New(),
		Name:        name,
		DisplayName: displayName,
		Description: &description,
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := role.Validate(); err != nil {
		return nil, fmt.Errorf("role validation failed: %w", err)
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

func (uc *roleUseCase) GetRole(ctx context.Context, roleID uuidv7.UUID) (*entity.Role, error) {
	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Load permissions
	permissions, err := uc.roleRepo.GetPermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	role.Permissions = permissions

	return role, nil
}

func (uc *roleUseCase) GetRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	role, err := uc.roleRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Load permissions
	permissions, err := uc.roleRepo.GetPermissions(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	role.Permissions = permissions

	return role, nil
}

func (uc *roleUseCase) ListRoles(ctx context.Context) ([]*entity.Role, error) {
	roles, err := uc.roleRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	// Load permissions for each role
	for _, role := range roles {
		permissions, err := uc.roleRepo.GetPermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %w", role.Name, err)
		}
		role.Permissions = permissions
	}

	return roles, nil
}

func (uc *roleUseCase) UpdateRole(ctx context.Context, roleID uuidv7.UUID, displayName, description string) (*entity.Role, error) {
	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	if role.IsSystem {
		return nil, ErrCannotDeleteSystemRole
	}

	role.DisplayName = displayName
	role.Description = &description
	role.UpdatedAt = time.Now()

	if err := role.Validate(); err != nil {
		return nil, fmt.Errorf("role validation failed: %w", err)
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

func (uc *roleUseCase) DeleteRole(ctx context.Context, roleID uuidv7.UUID) error {
	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role.IsSystem {
		return ErrCannotDeleteSystemRole
	}

	if err := uc.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

func (uc *roleUseCase) AddPermissionToRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	// Verify role exists
	_, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Verify permission exists
	_, err = uc.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to get permission: %w", err)
	}

	if err := uc.roleRepo.AddPermission(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to add permission to role: %w", err)
	}

	return nil
}

func (uc *roleUseCase) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error {
	if err := uc.roleRepo.RemovePermission(ctx, roleID, permissionID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}
	return nil
}

func (uc *roleUseCase) GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*entity.Permission, error) {
	permissions, err := uc.roleRepo.GetPermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return permissions, nil
}

func (uc *roleUseCase) SyncRolePermissions(ctx context.Context, roleID uuidv7.UUID, permissionIDs []uuidv7.UUID) error {
	// Verify role exists
	_, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Verify all permissions exist
	if len(permissionIDs) > 0 {
		existingPerms, err := uc.permissionRepo.GetByIDs(ctx, permissionIDs)
		if err != nil {
			return fmt.Errorf("failed to verify permissions: %w", err)
		}
		if len(existingPerms) != len(permissionIDs) {
			return ErrPermissionNotFound
		}
	}

	if err := uc.roleRepo.SyncPermissions(ctx, roleID, permissionIDs); err != nil {
		return fmt.Errorf("failed to sync permissions: %w", err)
	}

	return nil
}

func (uc *roleUseCase) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuidv7.UUID, expiresAt *time.Time) error {
	// Verify role exists
	_, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to get role: %w", err)
	}

	userRole := &entity.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedAt: time.Now(),
		AssignedBy: &assignedBy,
		ExpiresAt:  expiresAt,
	}

	if err := uc.roleRepo.AssignToUser(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	return nil
}

func (uc *roleUseCase) RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error {
	if err := uc.roleRepo.RemoveFromUser(ctx, userID, roleID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	return nil
}

func (uc *roleUseCase) GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error) {
	roles, err := uc.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Load permissions for each role
	for _, role := range roles {
		permissions, err := uc.roleRepo.GetPermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %w", role.Name, err)
		}
		role.Permissions = permissions
	}

	return roles, nil
}

func (uc *roleUseCase) GetUserActiveRoles(ctx context.Context, userID uuidv7.UUID) ([]*entity.Role, error) {
	roles, err := uc.roleRepo.GetUserActiveRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user active roles: %w", err)
	}

	// Load permissions for each role
	for _, role := range roles {
		permissions, err := uc.roleRepo.GetPermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %w", role.Name, err)
		}
		role.Permissions = permissions
	}

	return roles, nil
}

func (uc *roleUseCase) GetUsersWithRole(ctx context.Context, roleID uuidv7.UUID) ([]uuidv7.UUID, error) {
	userIDs, err := uc.roleRepo.GetUsersWithRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with role: %w", err)
	}
	return userIDs, nil
}

func (uc *roleUseCase) GetUserPermissions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Permission, error) {
	roles, err := uc.GetUserActiveRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Collect unique permissions from all roles
	permMap := make(map[string]*entity.Permission)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			key := perm.String()
			if _, exists := permMap[key]; !exists {
				permMap[key] = perm
			}
		}
	}

	// Convert map to slice
	permissions := make([]*entity.Permission, 0, len(permMap))
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

func (uc *roleUseCase) HasPermission(ctx context.Context, userID uuidv7.UUID, permission string) (bool, error) {
	permissions, err := uc.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Matches(permission) {
			return true, nil
		}
	}

	return false, nil
}

func (uc *roleUseCase) HasAnyPermission(ctx context.Context, userID uuidv7.UUID, requiredPermissions []string) (bool, error) {
	permissions, err := uc.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, required := range requiredPermissions {
		for _, perm := range permissions {
			if perm.Matches(required) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (uc *roleUseCase) HasAllPermissions(ctx context.Context, userID uuidv7.UUID, requiredPermissions []string) (bool, error) {
	permissions, err := uc.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, required := range requiredPermissions {
		found := false
		for _, perm := range permissions {
			if perm.Matches(required) {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}
