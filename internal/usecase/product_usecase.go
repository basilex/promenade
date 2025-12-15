package usecase

import (
    "context"
    
    "github.com/google/uuid"
    "github.com/basilex/promenade/internal/domain/entity"
    "github.com/basilex/promenade/internal/domain/repository"
    "github.com/basilex/promenade/internal/infrastructure/database"
)

type ProductUseCase struct {
    productRepo repository.ProductRepository
    txManager database.TransactionManager
}

func NewProductUseCase(
    productRepo repository.ProductRepository,
    txManager database.TransactionManager,
) *ProductUseCase {
    return &ProductUseCase{
        productRepo: productRepo,
        txManager:  txManager,
    }
}

func (uc *ProductUseCase) CreateProduct(ctx context.Context, name string) (*entity.Product, error) {
    if name == "" {
        return nil, entity.ErrInvalidInput
    }

    product := &entity.Product{
        ID:     uuid.New(),
        Name:   name,
        Active: true,
    }

    if err := uc.productRepo.Create(ctx, product); err != nil {
        return nil, err
    }

    return product, nil
}

func (uc *ProductUseCase) GetProduct(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
    return uc.productRepo.GetByID(ctx, id)
}

func (uc *ProductUseCase) UpdateProduct(ctx context.Context, product *entity.Product) error {
    existing, err := uc.productRepo.GetByID(ctx, product.ID)
    if err != nil {
        return err
    }

    if existing == nil {
        return entity.ErrProductNotFound
    }

    return uc.productRepo.Update(ctx, product)
}

func (uc *ProductUseCase) DeleteProduct(ctx context.Context, id uuid.UUID) error {
    return uc.productRepo.Delete(ctx, id)
}

func (uc *ProductUseCase) ListProducts(ctx context.Context, limit, offset int) ([]*entity.Product, int, error) {
    products, err := uc.productRepo.List(ctx, limit, offset)
    if err != nil {
        return nil, 0, err
    }

    total, err := uc.productRepo.Count(ctx)
    if err != nil {
        return nil, 0, err
    }

    return products, total, nil
}
