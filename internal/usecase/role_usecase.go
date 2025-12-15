package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
)

type RoleUseCase struct {
	roleRepo  repository.RoleRepository
	txManager database.TransactionManager
}

func NewRoleUseCase(
	roleRepo repository.RoleRepository,
	txManager database.TransactionManager,
) *RoleUseCase {
	return &RoleUseCase{
		roleRepo:  roleRepo,
		txManager: txManager,
	}
}

func (uc *RoleUseCase) CreateRole(ctx context.Context, name string) (*entity.Role, error) {
	if name == "" {
		return nil, entity.ErrInvalidInput
	}

	role := &entity.Role{
		ID:     uuid.New(),
		Name:   name,
		Active: true,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (uc *RoleUseCase) GetRole(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	return uc.roleRepo.GetByID(ctx, id)
}

func (uc *RoleUseCase) UpdateRole(ctx context.Context, role *entity.Role) error {
	existing, err := uc.roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		return err
	}

	if existing == nil {
		return entity.ErrRoleNotFound
	}

	return uc.roleRepo.Update(ctx, role)
}

func (uc *RoleUseCase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return uc.roleRepo.Delete(ctx, id)
}

func (uc *RoleUseCase) ListRoles(ctx context.Context, limit, offset int) ([]*entity.Role, int, error) {
	roles, err := uc.roleRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := uc.roleRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}
