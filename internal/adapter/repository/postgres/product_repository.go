package postgres

import (
    "context"
    "database/sql"
    
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/basilex/promenade/internal/domain/entity"
    "github.com/basilex/promenade/internal/domain/repository"
)

type productRepository struct {
    *BaseRepository
}

func NewProductRepository(db *sqlx.DB) repository.ProductRepository {
    return &productRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
    query := `
        INSERT INTO products (id, name, active, created_at, updated_at)
        VALUES (:id, :name, :active, NOW(), NOW())
    `
    return r.NamedExec(ctx, query, product)
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
    var product entity.Product
    query := `
        SELECT id, name, active, created_at, updated_at
        FROM products
        WHERE id = $1
    `
    
    err := r.Get(ctx, &product, query, id)
    if err == sql.ErrNoRows {
        return nil, entity.ErrProductNotFound
    }
    if err != nil {
        return nil, err
    }
    
    return &product, nil
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
    query := `
        UPDATE products
        SET name = :name, active = :active, updated_at = NOW()
        WHERE id = :id
    `
    return r.NamedExec(ctx, query, product)
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM products WHERE id = $1`
    return r.Exec(ctx, query, id)
}

func (r *productRepository) List(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
    var products []*entity.Product
    query := `
        SELECT id, name, active, created_at, updated_at
        FROM products
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `
    
    err := r.Select(ctx, &products, query, limit, offset)
    if err != nil {
        return nil, err
    }
    
    return products, nil
}

func (r *productRepository) Count(ctx context.Context) (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM products`
    err := r.Get(ctx, &count, query)
    return count, err
}
