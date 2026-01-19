package usecase

import (
	"context"

	producterrors "github.com/basilex/promenade/internal/contexts/warehouse/product"
	"github.com/basilex/promenade/internal/contexts/warehouse/product/aggregate"
	"github.com/basilex/promenade/internal/contexts/warehouse/product/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IProductUseCase defines business operations for Product aggregate.
//
// Business rules enforced:
// - SKU uniqueness across all products
// - Product lifecycle state transitions
// - Inventory settings validation
// - Physical properties validation
type IProductUseCase interface {
	// CreateProduct creates a new product with required fields.
	// Validates SKU uniqueness before creating.
	CreateProduct(ctx context.Context, sku, name string) (*aggregate.Product, error)

	// GetProduct retrieves a product by ID.
	GetProduct(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error)

	// GetProductBySKU retrieves a product by SKU.
	GetProductBySKU(ctx context.Context, sku string) (*aggregate.Product, error)

	// UpdateProduct updates product properties.
	// Validates business rules before saving.
	UpdateProduct(ctx context.Context, product *aggregate.Product) error

	// DeleteProduct soft-deletes a product.
	// Business rule: Cannot delete if active inventory exists (check in handler).
	DeleteProduct(ctx context.Context, id uuidv7.UUID) error

	// ListProducts retrieves paginated list of products.
	ListProducts(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error)

	// ListProductsByCategory retrieves products by category.
	ListProductsByCategory(ctx context.Context, category string, page, pageSize int) ([]*aggregate.Product, error)

	// ListProductsByBrand retrieves products by brand.
	ListProductsByBrand(ctx context.Context, brand string, page, pageSize int) ([]*aggregate.Product, error)

	// ListProductsByStatus retrieves products by status.
	ListProductsByStatus(ctx context.Context, status aggregate.ProductStatus, page, pageSize int) ([]*aggregate.Product, error)

	// SearchProducts performs full-text search.
	SearchProducts(ctx context.Context, query string, page, pageSize int) ([]*aggregate.Product, error)

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
	SetPhysicalProperties(ctx context.Context, id uuidv7.UUID, weight float64, dimensions aggregate.Dimensions) error

	// ListLowStockProducts retrieves products below reorder point.
	ListLowStockProducts(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error)
}

// ProductUseCase implements IProductUseCase interface.
type ProductUseCase struct {
	repo repository.IProductRepository
}

// NewProductUseCase creates a new Product use case.
func NewProductUseCase(repo repository.IProductRepository) IProductUseCase {
	return &ProductUseCase{
		repo: repo,
	}
}

// CreateProduct creates a new product with required fields.
func (uc *ProductUseCase) CreateProduct(ctx context.Context, sku, name string) (*aggregate.Product, error) {
	// Check SKU uniqueness
	exists, err := uc.repo.ExistsBySKU(ctx, sku)
	if err != nil {
		return nil, producterrors.ErrCheckSKUFailed
	}
	if exists {
		return nil, producterrors.ErrProductSKUDuplicate
	}

	// Create product
	product, err := aggregate.NewProduct(sku, name)
	if err != nil {
		return nil, err
	}

	// Validate business rules
	if err := product.Validate(); err != nil {
		return nil, err
	}

	// Persist
	if err := uc.repo.Create(ctx, product); err != nil {
		return nil, producterrors.ErrCreateFailed
	}

	return product, nil
}

// GetProduct retrieves a product by ID.
func (uc *ProductUseCase) GetProduct(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// GetProductBySKU retrieves a product by SKU.
func (uc *ProductUseCase) GetProductBySKU(ctx context.Context, sku string) (*aggregate.Product, error) {
	product, err := uc.repo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// UpdateProduct updates product properties.
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, product *aggregate.Product) error {
	// Validate business rules
	if err := product.Validate(); err != nil {
		return err
	}

	// Persist
	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrUpdateFailed
	}

	return nil
}

// DeleteProduct soft-deletes a product.
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, id uuidv7.UUID) error {
	// Retrieve product
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete
	product.SoftDelete()

	// Persist
	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrDeleteFailed
	}

	return nil
}

// ListProducts retrieves paginated list of products.
func (uc *ProductUseCase) ListProducts(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
	products, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, producterrors.ErrListFailed
	}
	return products, nil
}

// ListProductsByCategory retrieves products by category.
func (uc *ProductUseCase) ListProductsByCategory(ctx context.Context, category string, page, pageSize int) ([]*aggregate.Product, error) {
	products, err := uc.repo.ListByCategory(ctx, category, page, pageSize)
	if err != nil {
		return nil, producterrors.ErrListByCategoryFailed
	}
	return products, nil
}

// ListProductsByBrand retrieves products by brand.
func (uc *ProductUseCase) ListProductsByBrand(ctx context.Context, brand string, page, pageSize int) ([]*aggregate.Product, error) {
	products, err := uc.repo.ListByBrand(ctx, brand, page, pageSize)
	if err != nil {
		return nil, producterrors.ErrListByBrandFailed
	}
	return products, nil
}

// ListProductsByStatus retrieves products by status.
func (uc *ProductUseCase) ListProductsByStatus(ctx context.Context, status aggregate.ProductStatus, page, pageSize int) ([]*aggregate.Product, error) {
	products, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, producterrors.ErrListByStatusFailed
	}
	return products, nil
}

// SearchProducts performs full-text search.
func (uc *ProductUseCase) SearchProducts(ctx context.Context, query string, page, pageSize int) ([]*aggregate.Product, error) {
	products, err := uc.repo.Search(ctx, query, page, pageSize)
	if err != nil {
		return nil, producterrors.ErrSearchFailed
	}
	return products, nil
}

// CountProducts returns total number of products.
func (uc *ProductUseCase) CountProducts(ctx context.Context) (int, error) {
	count, err := uc.repo.Count(ctx)
	if err != nil {
		return 0, producterrors.ErrCountFailed
	}
	return count, nil
}

// ActivateProduct transitions product to Active status.
func (uc *ProductUseCase) ActivateProduct(ctx context.Context, id uuidv7.UUID) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.Activate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrActivateFailed
	}

	return nil
}

// DeactivateProduct transitions product to OutOfStock status.
func (uc *ProductUseCase) DeactivateProduct(ctx context.Context, id uuidv7.UUID) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.Deactivate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrDeactivateFailed
	}

	return nil
}

// DiscontinueProduct marks product as discontinued.
func (uc *ProductUseCase) DiscontinueProduct(ctx context.Context, id uuidv7.UUID) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.Discontinue(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrDiscontinueFailed
	}

	return nil
}

// UpdateInventorySettings updates inventory tracking settings.
func (uc *ProductUseCase) UpdateInventorySettings(ctx context.Context, id uuidv7.UUID, trackInventory, allowBackorder bool) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	product.SetInventorySettings(trackInventory, allowBackorder)

	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrUpdateInventorySettingsFailed
	}

	return nil
}

// SetReorderPoint sets reorder threshold and quantity.
func (uc *ProductUseCase) SetReorderPoint(ctx context.Context, id uuidv7.UUID, point, quantity int) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.SetReorderPoint(point, quantity); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrSetReorderPointFailed
	}

	return nil
}

// SetPhysicalProperties sets weight and dimensions.
func (uc *ProductUseCase) SetPhysicalProperties(ctx context.Context, id uuidv7.UUID, weight float64, dimensions aggregate.Dimensions) error {
	product, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := product.SetPhysicalProperties(weight, dimensions); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, product); err != nil {
		return producterrors.ErrSetPhysicalPropertiesFailed
	}

	return nil
}

// ListLowStockProducts retrieves products below reorder point.
func (uc *ProductUseCase) ListLowStockProducts(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
	products, err := uc.repo.ListLowStock(ctx, page, pageSize)
	if err != nil {
		return nil, producterrors.ErrListLowStockFailed
	}
	return products, nil
}
