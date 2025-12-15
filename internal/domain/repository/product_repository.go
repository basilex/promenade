package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/basilex/promenade/internal/domain/entity"
)

type ProductRepository interface {
    Create(ctx context.Context, product *entity.Product) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
    Update(ctx context.Context, product *entity.Product) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*entity.Product, error)
    Count(ctx context.Context) (int, error)
}
