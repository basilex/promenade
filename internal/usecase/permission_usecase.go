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
	ErrPermissionAlreadyExists = errors.New("permission with this resource:action already exists")
	ErrInvalidPermissionFormat = errors.New("invalid permission format (expected resource:action)")
	ErrPermissionNotFound      = errors.New("permission not found")
)

type PermissionUseCase interface {
	// Permission management
	CreatePermission(ctx context.Context, resource, action, description string) (*entity.Permission, error)
	CreatePermissionFromString(ctx context.Context, permissionString, description string) (*entity.Permission, error)
	GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*entity.Permission, error)
	GetPermissionByResourceAction(ctx context.Context, resource, action string) (*entity.Permission, error)
	ListPermissions(ctx context.Context) ([]*entity.Permission, error)
	UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*entity.Permission, error)
	DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error

	// Bulk operations
	CreateMany(ctx context.Context, permissions []*entity.Permission) error
	FindByResource(ctx context.Context, resource string) ([]*entity.Permission, error)
}

type permissionUseCase struct {
	permissionRepo repository.PermissionRepository
}

func NewPermissionUseCase(permissionRepo repository.PermissionRepository) PermissionUseCase {
	return &permissionUseCase{
		permissionRepo: permissionRepo,
	}
}

func (uc *permissionUseCase) CreatePermission(ctx context.Context, resource, action, description string) (*entity.Permission, error) {
	// Check if permission already exists
	existing, err := uc.permissionRepo.GetByResourceAction(ctx, resource, action)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing permission: %w", err)
	}
	if existing != nil {
		return nil, ErrPermissionAlreadyExists
	}

	permission := &entity.Permission{
		ID:          uuidv7.New(),
		Resource:    resource,
		Action:      action,
		Description: &description,
		CreatedAt:   time.Now(),
	}

	if err := permission.Validate(); err != nil {
		return nil, fmt.Errorf("permission validation failed: %w", err)
	}

	if err := uc.permissionRepo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

func (uc *permissionUseCase) CreatePermissionFromString(ctx context.Context, permissionString, description string) (*entity.Permission, error) {
	perm, err := entity.NewPermission(permissionString)
	if err != nil {
		return nil, ErrInvalidPermissionFormat
	}

	// Check if permission already exists
	existing, err := uc.permissionRepo.GetByResourceAction(ctx, perm.Resource, perm.Action)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing permission: %w", err)
	}
	if existing != nil {
		return nil, ErrPermissionAlreadyExists
	}

	perm.Description = &description

	if err := uc.permissionRepo.Create(ctx, perm); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return perm, nil
}

func (uc *permissionUseCase) GetPermission(ctx context.Context, permissionID uuidv7.UUID) (*entity.Permission, error) {
	permission, err := uc.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return permission, nil
}

func (uc *permissionUseCase) GetPermissionByResourceAction(ctx context.Context, resource, action string) (*entity.Permission, error) {
	permission, err := uc.permissionRepo.GetByResourceAction(ctx, resource, action)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return permission, nil
}

func (uc *permissionUseCase) ListPermissions(ctx context.Context) ([]*entity.Permission, error) {
	permissions, err := uc.permissionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return permissions, nil
}

func (uc *permissionUseCase) UpdatePermission(ctx context.Context, permissionID uuidv7.UUID, description string) (*entity.Permission, error) {
	permission, err := uc.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	permission.Description = &description

	if err := uc.permissionRepo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

func (uc *permissionUseCase) DeletePermission(ctx context.Context, permissionID uuidv7.UUID) error {
	if err := uc.permissionRepo.Delete(ctx, permissionID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrPermissionNotFound
		}
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

func (uc *permissionUseCase) CreateMany(ctx context.Context, permissions []*entity.Permission) error {
	// Validate all permissions
	for _, perm := range permissions {
		if err := perm.Validate(); err != nil {
			return fmt.Errorf("permission validation failed for %s: %w", perm.String(), err)
		}
	}

	if err := uc.permissionRepo.CreateMany(ctx, permissions); err != nil {
		return fmt.Errorf("failed to create many permissions: %w", err)
	}

	return nil
}

func (uc *permissionUseCase) FindByResource(ctx context.Context, resource string) ([]*entity.Permission, error) {
	permissions, err := uc.permissionRepo.FindByResource(ctx, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to find permissions by resource: %w", err)
	}
	return permissions, nil
}
