package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/basilex/promenade/internal/domain/entity"
)

type RoleRepository interface {
    Create(ctx context.Context, role *entity.Role) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error)
    Update(ctx context.Context, role *entity.Role) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*entity.Role, error)
    Count(ctx context.Context) (int, error)
}
