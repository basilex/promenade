package product

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines business operations for Product aggregate.
//
// Business rules enforced:
// - SKU uniqueness across all products
// - Product lifecycle state transitions
// - Inventory settings validation
// - Physical properties validation
type IUseCase interface {
	// CreateProduct creates a new product with required fields.
	// Validates SKU uniqueness before creating.
	CreateProduct(ctx context.Context, sku, name string) (*Product, error)

	// GetProduct retrieves a product by ID.
	GetProduct(ctx context.Context, id uuidv7.UUID) (*Product, error)

	// GetProductBySKU retrieves a product by SKU.
	GetProductBySKU(ctx context.Context, sku string) (*Product, error)

	// UpdateProduct updates product properties.
	// Validates business rules before saving.
	UpdateProduct(ctx context.Context, product *Product) error

	// DeleteProduct soft-deletes a product.
	// Business rule: Cannot delete if active inventory exists (check in handler).
	DeleteProduct(ctx context.Context, id uuidv7.UUID) error

	// ListProducts retrieves paginated list of products.
	ListProducts(ctx context.Context, page, pageSize int) ([]*Product, error)

	// ListProductsByCategory retrieves products by category.
	ListProductsByCategory(ctx context.Context, category string, page, pageSize int) ([]*Product, error)

	// ListProductsByBrand retrieves products by brand.
	ListProductsByBrand(ctx context.Context, brand string, page, pageSize int) ([]*Product, error)

	// ListProductsByStatus retrieves products by status.
	ListProductsByStatus(ctx context.Context, status ProductStatus, page, pageSize int) ([]*Product, error)

	// SearchProducts performs full-text search.
	SearchProducts(ctx context.Context, query string, page, pageSize int) ([]*Product, error)

	// CountProducts returns total number of products.
	CountProducts(ctx context.Context) (int, error)

	// ActivateProduct transitions product to Active status.
	ActivateProduct(ctx context.Context, id uuidv7.UUID) error

	// DeactivateProduct transitions product to OutOfStock status.
	DeactivateProduct(ctx context.Context, id uuidv7.UUID) error

	// DiscontinueProduct marks product as discontinued.
	DiscontinueProduct(ctx context.Context, id uuidv7.UUID) error

	// UpdateInventorySettings updates inventory tracking settings.
	UpdateInventorySettings(ctx context.Context, id uuidv7.UUID, trackInventory, allowBackorder bool) error

	// SetReorderPoint sets reorder threshold and quantity.
	SetReorderPoint(ctx context.Context, id uuidv7.UUID, point, quantity int) error

	// SetPhysicalProperties sets weight and dimensions.
	SetPhysicalProperties(ctx context.Context, id uuidv7.UUID, weight float64, dimensions Dimensions) error

	// ListLowStockProducts retrieves products below reorder point.
	ListLowStockProducts(ctx context.Context, page, pageSize int) ([]*Product, error)
}

// useCase implements IUseCase interface.
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new Product use case.
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateProduct creates a new product with required fields.
func (uc *useCase) CreateProduct(ctx context.Context, sku, name string) (*Product, error) {
	// Check SKU uniqueness
	exists, err := uc.repo.ExistsBySKU(ctx, sku)
	if err != nil {
		return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
	}
	if exists {
		return nil, ErrProductSKUDuplicate
	}

	// Create product
	product, err := NewProduct(sku, name)
	if err != nil {
		return nil, err
	}

	// Validate business rules
	if err := product.Validate(); err != nil {
		return nil, err
	}

	// Persist
	if err := uc.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetProduct retrieves a product by ID.
func (uc *useCase) GetProduct(ctx context.Context, id uuidv7.UUID) (*Product, error) {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// GetProductBySKU retrieves a product by SKU.
func (uc *useCase) GetProductBySKU(ctx context.Context, sku string) (*Product, error) {
	product, err := uc.repo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// UpdateProduct updates product properties.
func (uc *useCase) UpdateProduct(ctx context.Context, product *Product) error {
	// Validate business rules
	if err := product.Validate(); err != nil {
		return err
	}

	// Persist
	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// DeleteProduct soft-deletes a product.
func (uc *useCase) DeleteProduct(ctx context.Context, id uuidv7.UUID) error {
	// Retrieve product
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete
	product.SoftDelete()

	// Persist
	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// ListProducts retrieves paginated list of products.
func (uc *useCase) ListProducts(ctx context.Context, page, pageSize int) ([]*Product, error) {
	products, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	return products, nil
}

// ListProductsByCategory retrieves products by category.
func (uc *useCase) ListProductsByCategory(ctx context.Context, category string, page, pageSize int) ([]*Product, error) {
	products, err := uc.repo.ListByCategory(ctx, category, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by category: %w", err)
	}
	return products, nil
}

// ListProductsByBrand retrieves products by brand.
func (uc *useCase) ListProductsByBrand(ctx context.Context, brand string, page, pageSize int) ([]*Product, error) {
	products, err := uc.repo.ListByBrand(ctx, brand, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by brand: %w", err)
	}
	return products, nil
}

// ListProductsByStatus retrieves products by status.
func (uc *useCase) ListProductsByStatus(ctx context.Context, status ProductStatus, page, pageSize int) ([]*Product, error) {
	products, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by status: %w", err)
	}
	return products, nil
}

// SearchProducts performs full-text search.
func (uc *useCase) SearchProducts(ctx context.Context, query string, page, pageSize int) ([]*Product, error) {
	products, err := uc.repo.Search(ctx, query, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	return products, nil
}

// CountProducts returns total number of products.
func (uc *useCase) CountProducts(ctx context.Context) (int, error) {
	count, err := uc.repo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
}

// ActivateProduct transitions product to Active status.
func (uc *useCase) ActivateProduct(ctx context.Context, id uuidv7.UUID) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.Activate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to activate product: %w", err)
	}

	return nil
}

// DeactivateProduct transitions product to OutOfStock status.
func (uc *useCase) DeactivateProduct(ctx context.Context, id uuidv7.UUID) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.Deactivate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to deactivate product: %w", err)
	}

	return nil
}

// DiscontinueProduct marks product as discontinued.
func (uc *useCase) DiscontinueProduct(ctx context.Context, id uuidv7.UUID) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.Discontinue(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to discontinue product: %w", err)
	}

	return nil
}

// UpdateInventorySettings updates inventory tracking settings.
func (uc *useCase) UpdateInventorySettings(ctx context.Context, id uuidv7.UUID, trackInventory, allowBackorder bool) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	product.SetInventorySettings(trackInventory, allowBackorder)

	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to update inventory settings: %w", err)
	}

	return nil
}

// SetReorderPoint sets reorder threshold and quantity.
func (uc *useCase) SetReorderPoint(ctx context.Context, id uuidv7.UUID, point, quantity int) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.SetReorderPoint(point, quantity); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to set reorder point: %w", err)
	}

	return nil
}

// SetPhysicalProperties sets weight and dimensions.
func (uc *useCase) SetPhysicalProperties(ctx context.Context, id uuidv7.UUID, weight float64, dimensions Dimensions) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.SetPhysicalProperties(weight, dimensions); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to set physical properties: %w", err)
	}

	return nil
}

// ListLowStockProducts retrieves products below reorder point.
func (uc *useCase) ListLowStockProducts(ctx context.Context, page, pageSize int) ([]*Product, error) {
	products, err := uc.repo.ListLowStock(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list low stock products: %w", err)
	}
	return products, nil
}
