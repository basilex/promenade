package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/product/aggregate"
	producterrors "github.com/basilex/promenade/internal/contexts/warehouse/product"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

// MockRepository is a manual mock implementation of IRepository for testing.
type MockRepository struct {
	CreateFunc              func(ctx context.Context, p *aggregate.Product) error
	GetByIDFunc             func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error)
	GetBySKUFunc            func(ctx context.Context, sku string) (*aggregate.Product, error)
	UpdateFunc              func(ctx context.Context, p *aggregate.Product) error
	DeleteFunc              func(ctx context.Context, id uuidv7.UUID) error
	ListFunc                func(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error)
	ListByCategoryFunc      func(ctx context.Context, category string, page, pageSize int) ([]*aggregate.Product, error)
	ListByBrandFunc         func(ctx context.Context, brand string, page, pageSize int) ([]*aggregate.Product, error)
	ListByStatusFunc        func(ctx context.Context, status aggregate.ProductStatus, page, pageSize int) ([]*aggregate.Product, error)
	SearchFunc              func(ctx context.Context, query string, page, pageSize int) ([]*aggregate.Product, error)
	CountFunc               func(ctx context.Context) (int, error)
	ExistsBySKUFunc         func(ctx context.Context, sku string) (bool, error)
	GetByIDsFunc            func(ctx context.Context, ids []uuidv7.UUID) ([]*aggregate.Product, error)
	ListLowStockFunc        func(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error)
	BulkUpdateStatusFunc    func(ctx context.Context, ids []uuidv7.UUID, status aggregate.ProductStatus) error
}

func (m *MockRepository) Create(ctx context.Context, p *aggregate.Product) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, p)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockRepository) GetBySKU(ctx context.Context, sku string) (*aggregate.Product, error) {
	if m.GetBySKUFunc != nil {
		return m.GetBySKUFunc(ctx, sku)
	}
	return nil, errors.New("GetBySKUFunc not implemented")
}

func (m *MockRepository) Update(ctx context.Context, p *aggregate.Product) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, p)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockRepository) List(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, page, pageSize)
	}
	return nil, errors.New("ListFunc not implemented")
}

func (m *MockRepository) ListByCategory(ctx context.Context, category string, page, pageSize int) ([]*aggregate.Product, error) {
	if m.ListByCategoryFunc != nil {
		return m.ListByCategoryFunc(ctx, category, page, pageSize)
	}
	return nil, errors.New("ListByCategoryFunc not implemented")
}

func (m *MockRepository) ListByBrand(ctx context.Context, brand string, page, pageSize int) ([]*aggregate.Product, error) {
	if m.ListByBrandFunc != nil {
		return m.ListByBrandFunc(ctx, brand, page, pageSize)
	}
	return nil, errors.New("ListByBrandFunc not implemented")
}

func (m *MockRepository) ListByStatus(ctx context.Context, status aggregate.ProductStatus, page, pageSize int) ([]*aggregate.Product, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, errors.New("ListByStatusFunc not implemented")
}

func (m *MockRepository) Search(ctx context.Context, query string, page, pageSize int) ([]*aggregate.Product, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, page, pageSize)
	}
	return nil, errors.New("SearchFunc not implemented")
}

func (m *MockRepository) Count(ctx context.Context) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx)
	}
	return 0, errors.New("CountFunc not implemented")
}

func (m *MockRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	if m.ExistsBySKUFunc != nil {
		return m.ExistsBySKUFunc(ctx, sku)
	}
	return false, errors.New("ExistsBySKUFunc not implemented")
}

func (m *MockRepository) GetByIDs(ctx context.Context, ids []uuidv7.UUID) ([]*aggregate.Product, error) {
	if m.GetByIDsFunc != nil {
		return m.GetByIDsFunc(ctx, ids)
	}
	return nil, errors.New("GetByIDsFunc not implemented")
}

func (m *MockRepository) ListLowStock(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
	if m.ListLowStockFunc != nil {
		return m.ListLowStockFunc(ctx, page, pageSize)
	}
	return nil, errors.New("ListLowStockFunc not implemented")
}

func (m *MockRepository) BulkUpdateStatus(ctx context.Context, ids []uuidv7.UUID, status aggregate.ProductStatus) error {
	if m.BulkUpdateStatusFunc != nil {
		return m.BulkUpdateStatusFunc(ctx, ids, status)
	}
	return errors.New("BulkUpdateStatusFunc not implemented")
}

// ============================================================================
// Test Helpers
// ============================================================================

func createTestProduct(t *testing.T) *aggregate.Product {
	t.Helper()
	p, err := aggregate.NewProduct("TEST-SKU-001", "Test aggregate.Product")
	require.NoError(t, err)
	return p
}

func createTestProductWithDetails(t *testing.T, sku, name, category, brand string) *aggregate.Product {
	t.Helper()
	p, err := aggregate.NewProduct(sku, name)
	require.NoError(t, err)
	p.Category = category
	p.Brand = brand
	return p
}

// ============================================================================
// CreateProduct Tests
// ============================================================================

func TestUseCase_CreateProduct_Success(t *testing.T) {
	repo := &MockRepository{
		ExistsBySKUFunc: func(ctx context.Context, sku string) (bool, error) {
			return false, nil // SKU doesn't exist
		},
		CreateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil // Success
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.CreateProduct(context.Background(), "TEST-SKU-001", "Test aggregate.Product")

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "TEST-SKU-001", product.SKU)
	assert.Equal(t, "Test aggregate.Product", product.Name)
	assert.Equal(t, aggregate.ProductStatusDraft, product.Status)
	assert.True(t, product.TrackInventory)
}

func TestUseCase_CreateProduct_DuplicateSKU(t *testing.T) {
	repo := &MockRepository{
		ExistsBySKUFunc: func(ctx context.Context, sku string) (bool, error) {
			return true, nil // SKU already exists
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.CreateProduct(context.Background(), "TEST-SKU-001", "Test aggregate.Product")

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.ErrorIs(t, err, producterrors.ErrProductSKUDuplicate)
}

func TestUseCase_CreateProduct_RepositoryExistsError(t *testing.T) {
	repo := &MockRepository{
		ExistsBySKUFunc: func(ctx context.Context, sku string) (bool, error) {
			return false, errors.New("database connection error")
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.CreateProduct(context.Background(), "TEST-SKU-001", "Test aggregate.Product")

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.Is(err, producterrors.ErrCheckSKUFailed))
}

func TestUseCase_CreateProduct_RepositoryCreateError(t *testing.T) {
	repo := &MockRepository{
		ExistsBySKUFunc: func(ctx context.Context, sku string) (bool, error) {
			return false, nil
		},
		CreateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return errors.New("database write error")
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.CreateProduct(context.Background(), "TEST-SKU-001", "Test aggregate.Product")

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.True(t, errors.Is(err, producterrors.ErrCreateFailed))
}

// ============================================================================
// GetProduct Tests
// ============================================================================

func TestUseCase_GetProduct_Success(t *testing.T) {
	expectedProduct := createTestProduct(t)
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return expectedProduct, nil
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.GetProduct(context.Background(), expectedProduct.ID)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, expectedProduct.ID, product.ID)
	assert.Equal(t, expectedProduct.SKU, product.SKU)
}

func TestUseCase_GetProduct_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return nil, producterrors.ErrProductNotFound
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.GetProduct(context.Background(), uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.ErrorIs(t, err, producterrors.ErrProductNotFound)
}

func TestUseCase_GetProduct_RepositoryError(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return nil, errors.New("database read error")
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.GetProduct(context.Background(), uuidv7.New())

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "database read error")
}

// ============================================================================
// GetProductBySKU Tests
// ============================================================================

func TestUseCase_GetProductBySKU_Success(t *testing.T) {
	expectedProduct := createTestProduct(t)
	
	repo := &MockRepository{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Product, error) {
			return expectedProduct, nil
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.GetProductBySKU(context.Background(), "TEST-SKU-001")

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "TEST-SKU-001", product.SKU)
}

func TestUseCase_GetProductBySKU_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetBySKUFunc: func(ctx context.Context, sku string) (*aggregate.Product, error) {
			return nil, producterrors.ErrProductNotFound
		},
	}
	uc := NewProductUseCase(repo)

	product, err := uc.GetProductBySKU(context.Background(), "NONEXISTENT-SKU")

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.ErrorIs(t, err, producterrors.ErrProductNotFound)
}

// ============================================================================
// UpdateProduct Tests
// ============================================================================

func TestUseCase_UpdateProduct_Success(t *testing.T) {
	product := createTestProduct(t)
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	product.Name = "Updated Name"
	product.Description = "Updated Description"
	product.Category = "Electronics"
	product.Brand = "TestBrand"

	err := uc.UpdateProduct(context.Background(), product)

	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", product.Name)
	assert.Equal(t, "Updated Description", product.Description)
	assert.Equal(t, "Electronics", product.Category)
	assert.Equal(t, "TestBrand", product.Brand)
}

func TestUseCase_UpdateProduct_NotFound(t *testing.T) {
	repo := &MockRepository{
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return producterrors.ErrProductNotFound
		},
	}
	uc := NewProductUseCase(repo)

	p := createTestProduct(t)
	err := uc.UpdateProduct(context.Background(), p)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update product")
}

func TestUseCase_UpdateProduct_RepositoryError(t *testing.T) {
	product := createTestProduct(t)
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return errors.New("database write error")
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.UpdateProduct(context.Background(), product)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, producterrors.ErrUpdateFailed))
}

// ============================================================================
// DeleteProduct Tests
// ============================================================================

func TestUseCase_DeleteProduct_Success(t *testing.T) {
	product := createTestProduct(t)
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.DeleteProduct(context.Background(), product.ID)

	assert.NoError(t, err)
}

func TestUseCase_DeleteProduct_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return nil, producterrors.ErrProductNotFound
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.DeleteProduct(context.Background(), uuidv7.New())

	assert.Error(t, err)
	assert.ErrorIs(t, err, producterrors.ErrProductNotFound)
}

func TestUseCase_DeleteProduct_RepositoryError(t *testing.T) {
	product := createTestProduct(t)
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return errors.New("database delete error")
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.DeleteProduct(context.Background(), product.ID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, producterrors.ErrDeleteFailed))
}

// ============================================================================
// ListProducts Tests
// ============================================================================

func TestUseCase_ListProducts_Success(t *testing.T) {
	expectedProducts := []*aggregate.Product{
		createTestProduct(t),
		createTestProduct(t),
	}
	
	repo := &MockRepository{
		ListFunc: func(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
			return expectedProducts, nil
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.ListProducts(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestUseCase_ListProducts_EmptyResult(t *testing.T) {
	repo := &MockRepository{
		ListFunc: func(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
			return []*aggregate.Product{}, nil
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.ListProducts(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Empty(t, products)
}

func TestUseCase_ListProducts_RepositoryError(t *testing.T) {
	repo := &MockRepository{
		ListFunc: func(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
			return nil, errors.New("database query error")
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.ListProducts(context.Background(), 1, 10)

	assert.Error(t, err)
	assert.Nil(t, products)
	assert.True(t, errors.Is(err, producterrors.ErrListFailed))
}

// ============================================================================
// ListProductsByCategory Tests
// ============================================================================

func TestUseCase_ListProductsByCategory_Success(t *testing.T) {
	expectedProducts := []*aggregate.Product{
		createTestProductWithDetails(t, "SKU-1", "aggregate.Product 1", "Electronics", "Brand A"),
		createTestProductWithDetails(t, "SKU-2", "aggregate.Product 2", "Electronics", "Brand B"),
	}
	
	repo := &MockRepository{
		ListByCategoryFunc: func(ctx context.Context, category string, page, pageSize int) ([]*aggregate.Product, error) {
			return expectedProducts, nil
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.ListProductsByCategory(context.Background(), "Electronics", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

// ============================================================================
// ListProductsByBrand Tests
// ============================================================================

func TestUseCase_ListProductsByBrand_Success(t *testing.T) {
	expectedProducts := []*aggregate.Product{
		createTestProductWithDetails(t, "SKU-1", "aggregate.Product 1", "Electronics", "BrandA"),
	}
	
	repo := &MockRepository{
		ListByBrandFunc: func(ctx context.Context, brand string, page, pageSize int) ([]*aggregate.Product, error) {
			return expectedProducts, nil
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.ListProductsByBrand(context.Background(), "BrandA", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "BrandA", products[0].Brand)
}

// ============================================================================
// ListProductsByStatus Tests
// ============================================================================

func TestUseCase_ListProductsByStatus_Success(t *testing.T) {
	product := createTestProduct(t)
	product.Status = aggregate.ProductStatusActive
	
	repo := &MockRepository{
		ListByStatusFunc: func(ctx context.Context, status aggregate.ProductStatus, page, pageSize int) ([]*aggregate.Product, error) {
			return []*aggregate.Product{product}, nil
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.ListProductsByStatus(context.Background(), aggregate.ProductStatusActive, 1, 10)

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, aggregate.ProductStatusActive, products[0].Status)
}

// ============================================================================
// SearchProducts Tests
// ============================================================================

func TestUseCase_SearchProducts_Success(t *testing.T) {
	expectedProducts := []*aggregate.Product{
		createTestProduct(t),
	}
	
	repo := &MockRepository{
		SearchFunc: func(ctx context.Context, query string, page, pageSize int) ([]*aggregate.Product, error) {
			return expectedProducts, nil
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.SearchProducts(context.Background(), "test", 1, 10)

	assert.NoError(t, err)
	assert.Len(t, products, 1)
}

func TestUseCase_SearchProducts_EmptyQuery(t *testing.T) {
	repo := &MockRepository{
		SearchFunc: func(ctx context.Context, query string, page, pageSize int) ([]*aggregate.Product, error) {
			return []*aggregate.Product{}, nil
		},
	}
	uc := NewProductUseCase(repo)

	products, err := uc.SearchProducts(context.Background(), "", 1, 10)

	assert.NoError(t, err)
	assert.Empty(t, products)
}

// ============================================================================
// CountProducts Tests
// ============================================================================

func TestUseCase_CountProducts_Success(t *testing.T) {
	repo := &MockRepository{
		CountFunc: func(ctx context.Context) (int, error) {
			return 42, nil
		},
	}
	uc := NewProductUseCase(repo)

	count, err := uc.CountProducts(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 42, count)
}

func TestUseCase_CountProducts_RepositoryError(t *testing.T) {
	repo := &MockRepository{
		CountFunc: func(ctx context.Context) (int, error) {
			return 0, errors.New("database count error")
		},
	}
	uc := NewProductUseCase(repo)

	count, err := uc.CountProducts(context.Background())

	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.True(t, errors.Is(err, producterrors.ErrCountFailed))
}

// ============================================================================
// ActivateProduct Tests
// ============================================================================

func TestUseCase_ActivateProduct_Success(t *testing.T) {
	product := createTestProduct(t)
	product.Status = aggregate.ProductStatusDraft
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.ActivateProduct(context.Background(), product.ID)

	assert.NoError(t, err)
	assert.Equal(t, aggregate.ProductStatusActive, product.Status)
}

func TestUseCase_ActivateProduct_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return nil, producterrors.ErrProductNotFound
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.ActivateProduct(context.Background(), uuidv7.New())

	assert.Error(t, err)
	assert.ErrorIs(t, err, producterrors.ErrProductNotFound)
}

func TestUseCase_ActivateProduct_InvalidStatus(t *testing.T) {
	product := createTestProduct(t)
	product.Status = aggregate.ProductStatusActive
	product.IsActive = true
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.ActivateProduct(context.Background(), product.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, producterrors.ErrProductAlreadyActive)
}

// ============================================================================
// DeactivateProduct Tests
// ============================================================================

func TestUseCase_DeactivateProduct_Success(t *testing.T) {
	product := createTestProduct(t)
	product.Status = aggregate.ProductStatusActive // Start with Active status
	product.IsActive = true // Must set IsActive flag
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.DeactivateProduct(context.Background(), product.ID)

	assert.NoError(t, err)
	assert.Equal(t, aggregate.ProductStatusOutOfStock, product.Status)
}

func TestUseCase_DeactivateProduct_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return nil, producterrors.ErrProductNotFound
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.DeactivateProduct(context.Background(), uuidv7.New())

	assert.Error(t, err)
	assert.ErrorIs(t, err, producterrors.ErrProductNotFound)
}

// ============================================================================
// DiscontinueProduct Tests
// ============================================================================

func TestUseCase_DiscontinueProduct_Success(t *testing.T) {
	product := createTestProduct(t)
	product.Status = aggregate.ProductStatusActive
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.DiscontinueProduct(context.Background(), product.ID)

	assert.NoError(t, err)
	assert.Equal(t, aggregate.ProductStatusDiscontinued, product.Status)
}

// ============================================================================
// UpdateInventorySettings Tests
// ============================================================================

func TestUseCase_UpdateInventorySettings_EnableTracking(t *testing.T) {
	product := createTestProduct(t)
	product.TrackInventory = false
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.UpdateInventorySettings(context.Background(), product.ID, true, false)

	assert.NoError(t, err)
	assert.True(t, product.TrackInventory)
	assert.False(t, product.AllowBackorder)
}

func TestUseCase_UpdateInventorySettings_DisableTracking(t *testing.T) {
	product := createTestProduct(t)
	product.TrackInventory = true
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.UpdateInventorySettings(context.Background(), product.ID, false, false)

	assert.NoError(t, err)
	assert.False(t, product.TrackInventory)
}

func TestUseCase_UpdateInventorySettings_EnableBackorder(t *testing.T) {
	product := createTestProduct(t)
	product.AllowBackorder = false
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.UpdateInventorySettings(context.Background(), product.ID, true, true)

	assert.NoError(t, err)
	assert.True(t, product.AllowBackorder)
}

// ============================================================================
// SetReorderPoint Tests
// ============================================================================

func TestUseCase_SetReorderPoint_Success(t *testing.T) {
	product := createTestProduct(t)
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
		UpdateFunc: func(ctx context.Context, p *aggregate.Product) error {
			return nil
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.SetReorderPoint(context.Background(), product.ID, 10, 100)

	assert.NoError(t, err)
	assert.Equal(t, 10, product.ReorderPoint)
	assert.Equal(t, 100, product.ReorderQuantity)
}

func TestUseCase_SetReorderPoint_InvalidValues(t *testing.T) {
	product := createTestProduct(t)
	
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return product, nil
		},
	}
	uc := NewProductUseCase(repo)

	// Test reorder point > reorder quantity (invalid)
	err := uc.SetReorderPoint(context.Background(), product.ID, 100, 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reorder point must be less than reorder quantity")
}

func TestUseCase_SetReorderPoint_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
			return nil, producterrors.ErrProductNotFound
		},
	}
	uc := NewProductUseCase(repo)

	err := uc.SetReorderPoint(context.Background(), uuidv7.New(), 10, 100)

	assert.Error(t, err)
	assert.ErrorIs(t, err, producterrors.ErrProductNotFound)
}
