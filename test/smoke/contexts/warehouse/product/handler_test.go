package product_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/warehouse/product"
	productHTTP "github.com/basilex/promenade/internal/contexts/warehouse/product/adapter/http"
	productAggregate "github.com/basilex/promenade/internal/contexts/warehouse/product/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// ============================================================================
// Mock UseCase
// ============================================================================

// MockProductUseCase implements product.IUseCase for testing
type MockProductUseCase struct {
	CreateProductFunc           func(ctx context.Context, sku, name string) (*productAggregate.Product, error)
	GetProductFunc              func(ctx context.Context, id uuidv7.UUID) (*productAggregate.Product, error)
	GetProductBySKUFunc         func(ctx context.Context, sku string) (*productAggregate.Product, error)
	UpdateProductFunc           func(ctx context.Context, p *productAggregate.Product) error
	DeleteProductFunc           func(ctx context.Context, id uuidv7.UUID) error
	ListProductsFunc            func(ctx context.Context, page, pageSize int) ([]*productAggregate.Product, error)
	CountProductsFunc           func(ctx context.Context) (int, error)
	ListProductsByCategoryFunc  func(ctx context.Context, category string, page, pageSize int) ([]*productAggregate.Product, error)
	ListProductsByBrandFunc     func(ctx context.Context, brand string, page, pageSize int) ([]*productAggregate.Product, error)
	ListProductsByStatusFunc    func(ctx context.Context, status productAggregate.ProductStatus, page, pageSize int) ([]*productAggregate.Product, error)
	SearchProductsFunc          func(ctx context.Context, query string, page, pageSize int) ([]*productAggregate.Product, error)
	ActivateProductFunc         func(ctx context.Context, id uuidv7.UUID) error
	DeactivateProductFunc       func(ctx context.Context, id uuidv7.UUID) error
	DiscontinueProductFunc      func(ctx context.Context, id uuidv7.UUID) error
	UpdateInventorySettingsFunc func(ctx context.Context, id uuidv7.UUID, trackInventory, allowBackorder bool) error
	SetReorderPointFunc         func(ctx context.Context, id uuidv7.UUID, reorderPoint, reorderQuantity int) error
	SetPhysicalPropertiesFunc   func(ctx context.Context, id uuidv7.UUID, weight float64, dimensions productAggregate.Dimensions) error
	ListLowStockProductsFunc    func(ctx context.Context, page, pageSize int) ([]*productAggregate.Product, error)
}

func (m *MockProductUseCase) CreateProduct(ctx context.Context, sku, name string) (*productAggregate.Product, error) {
	if m.CreateProductFunc != nil {
		return m.CreateProductFunc(ctx, sku, name)
	}
	return nil, errors.New("CreateProductFunc not implemented")
}

func (m *MockProductUseCase) GetProduct(ctx context.Context, id uuidv7.UUID) (*productAggregate.Product, error) {
	if m.GetProductFunc != nil {
		return m.GetProductFunc(ctx, id)
	}
	return nil, errors.New("GetProductFunc not implemented")
}

func (m *MockProductUseCase) GetProductBySKU(ctx context.Context, sku string) (*productAggregate.Product, error) {
	if m.GetProductBySKUFunc != nil {
		return m.GetProductBySKUFunc(ctx, sku)
	}
	return nil, errors.New("GetProductBySKUFunc not implemented")
}

func (m *MockProductUseCase) UpdateProduct(ctx context.Context, p *productAggregate.Product) error {
	if m.UpdateProductFunc != nil {
		return m.UpdateProductFunc(ctx, p)
	}
	return errors.New("UpdateProductFunc not implemented")
}

func (m *MockProductUseCase) DeleteProduct(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteProductFunc != nil {
		return m.DeleteProductFunc(ctx, id)
	}
	return errors.New("DeleteProductFunc not implemented")
}

func (m *MockProductUseCase) ListProducts(ctx context.Context, page, pageSize int) ([]*productAggregate.Product, error) {
	if m.ListProductsFunc != nil {
		return m.ListProductsFunc(ctx, page, pageSize)
	}
	return nil, errors.New("ListProductsFunc not implemented")
}

func (m *MockProductUseCase) CountProducts(ctx context.Context) (int, error) {
	if m.CountProductsFunc != nil {
		return m.CountProductsFunc(ctx)
	}
	return 0, errors.New("CountProductsFunc not implemented")
}

func (m *MockProductUseCase) ListProductsByCategory(ctx context.Context, category string, page, pageSize int) ([]*productAggregate.Product, error) {
	if m.ListProductsByCategoryFunc != nil {
		return m.ListProductsByCategoryFunc(ctx, category, page, pageSize)
	}
	return nil, errors.New("ListProductsByCategoryFunc not implemented")
}

func (m *MockProductUseCase) ListProductsByBrand(ctx context.Context, brand string, page, pageSize int) ([]*productAggregate.Product, error) {
	if m.ListProductsByBrandFunc != nil {
		return m.ListProductsByBrandFunc(ctx, brand, page, pageSize)
	}
	return nil, errors.New("ListProductsByBrandFunc not implemented")
}

func (m *MockProductUseCase) ListProductsByStatus(ctx context.Context, status productAggregate.ProductStatus, page, pageSize int) ([]*productAggregate.Product, error) {
	if m.ListProductsByStatusFunc != nil {
		return m.ListProductsByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, errors.New("ListProductsByStatusFunc not implemented")
}

func (m *MockProductUseCase) SearchProducts(ctx context.Context, query string, page, pageSize int) ([]*productAggregate.Product, error) {
	if m.SearchProductsFunc != nil {
		return m.SearchProductsFunc(ctx, query, page, pageSize)
	}
	return nil, errors.New("SearchProductsFunc not implemented")
}

func (m *MockProductUseCase) ActivateProduct(ctx context.Context, id uuidv7.UUID) error {
	if m.ActivateProductFunc != nil {
		return m.ActivateProductFunc(ctx, id)
	}
	return errors.New("ActivateProductFunc not implemented")
}

func (m *MockProductUseCase) DeactivateProduct(ctx context.Context, id uuidv7.UUID) error {
	if m.DeactivateProductFunc != nil {
		return m.DeactivateProductFunc(ctx, id)
	}
	return errors.New("DeactivateProductFunc not implemented")
}

func (m *MockProductUseCase) DiscontinueProduct(ctx context.Context, id uuidv7.UUID) error {
	if m.DiscontinueProductFunc != nil {
		return m.DiscontinueProductFunc(ctx, id)
	}
	return errors.New("DiscontinueProductFunc not implemented")
}

func (m *MockProductUseCase) UpdateInventorySettings(ctx context.Context, id uuidv7.UUID, trackInventory, allowBackorder bool) error {
	if m.UpdateInventorySettingsFunc != nil {
		return m.UpdateInventorySettingsFunc(ctx, id, trackInventory, allowBackorder)
	}
	return errors.New("UpdateInventorySettingsFunc not implemented")
}

func (m *MockProductUseCase) SetReorderPoint(ctx context.Context, id uuidv7.UUID, reorderPoint, reorderQuantity int) error {
	if m.SetReorderPointFunc != nil {
		return m.SetReorderPointFunc(ctx, id, reorderPoint, reorderQuantity)
	}
	return errors.New("SetReorderPointFunc not implemented")
}

func (m *MockProductUseCase) SetPhysicalProperties(ctx context.Context, id uuidv7.UUID, weight float64, dimensions productAggregate.Dimensions) error {
	if m.SetPhysicalPropertiesFunc != nil {
		return m.SetPhysicalPropertiesFunc(ctx, id, weight, dimensions)
	}
	return errors.New("SetPhysicalPropertiesFunc not implemented")
}

func (m *MockProductUseCase) ListLowStockProducts(ctx context.Context, page, pageSize int) ([]*productAggregate.Product, error) {
	if m.ListLowStockProductsFunc != nil {
		return m.ListLowStockProductsFunc(ctx, page, pageSize)
	}
	return nil, errors.New("ListLowStockProductsFunc not implemented")
}

// Helper to create fake product
func fakeProduct() *productAggregate.Product {
	p, _ := productAggregate.NewProduct("TEST-SKU", "Test Product")
	return p
}

// ============================================================================
// Tests
// ============================================================================

func TestProductHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		CreateProductFunc: func(ctx context.Context, sku, name string) (*productAggregate.Product, error) {
			return fakeProduct(), nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.POST("/products", handler.Create)

	body := map[string]interface{}{
		"sku":  "TEST-SKU",
		"name": "Test Product",
	}
	w := smoke.MakeRequest(t, router, "POST", "/products", body)

	smoke.AssertSuccessResponse(t, w, http.StatusCreated)
}

func TestProductHandler_Create_DuplicateSKU(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		CreateProductFunc: func(ctx context.Context, sku, name string) (*productAggregate.Product, error) {
			return nil, product.ErrProductSKUDuplicate
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.POST("/products", handler.Create)

	body := map[string]interface{}{
		"sku":  "TEST-SKU",
		"name": "Test Product",
	}
	w := smoke.MakeRequest(t, router, "POST", "/products", body)

	smoke.AssertErrorResponse(t, w, http.StatusConflict, "CONFLICT")
}

func TestProductHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		GetProductFunc: func(ctx context.Context, id uuidv7.UUID) (*productAggregate.Product, error) {
			return fakeProduct(), nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.GET("/products/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/products/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestProductHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		GetProductFunc: func(ctx context.Context, id uuidv7.UUID) (*productAggregate.Product, error) {
			return nil, product.ErrProductNotFound
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.GET("/products/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/products/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestProductHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		ListProductsFunc: func(ctx context.Context, page, pageSize int) ([]*productAggregate.Product, error) {
			return []*productAggregate.Product{fakeProduct()}, nil
		},
		CountProductsFunc: func(ctx context.Context) (int, error) {
			return 1, nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.GET("/products", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/products?page=1&page_size=20", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestProductHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		GetProductFunc: func(ctx context.Context, id uuidv7.UUID) (*productAggregate.Product, error) {
			return fakeProduct(), nil
		},
		UpdateProductFunc: func(ctx context.Context, p *productAggregate.Product) error {
			return nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.PUT("/products/:id", handler.Update)

	body := map[string]interface{}{
		"name": "Updated Product",
	}
	w := smoke.MakeRequest(t, router, "PUT", "/products/"+smoke.FakeUUID(), body)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestProductHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		DeleteProductFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.DELETE("/products/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/products/"+smoke.FakeUUID(), nil)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestProductHandler_Activate_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		ActivateProductFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
		GetProductFunc: func(ctx context.Context, id uuidv7.UUID) (*productAggregate.Product, error) {
			p := fakeProduct()
			_ = p.Activate()
			return p, nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.POST("/products/:id/activate", handler.Activate)

	w := smoke.MakeRequest(t, router, "POST", "/products/"+smoke.FakeUUID()+"/activate", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestProductHandler_Search_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		SearchProductsFunc: func(ctx context.Context, query string, page, pageSize int) ([]*productAggregate.Product, error) {
			return []*productAggregate.Product{fakeProduct()}, nil
		},
		CountProductsFunc: func(ctx context.Context) (int, error) {
			return 1, nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.GET("/products/search", handler.Search)

	w := smoke.MakeRequest(t, router, "GET", "/products/search?q=test", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestProductHandler_ListByCategory_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProductUseCase{
		ListProductsByCategoryFunc: func(ctx context.Context, category string, page, pageSize int) ([]*productAggregate.Product, error) {
			return []*productAggregate.Product{fakeProduct()}, nil
		},
		CountProductsFunc: func(ctx context.Context) (int, error) {
			return 1, nil
		},
	}

	handler := productHTTP.NewProductHandler(mockUC)
	router.GET("/products/category/:category", handler.ListByCategory)

	w := smoke.MakeRequest(t, router, "GET", "/products/category/electronics", nil)

	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}
