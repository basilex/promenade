package product_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/product"
	productRepo "github.com/basilex/promenade/internal/contexts/warehouse/product/adapter/repository/postgres"
	"github.com/basilex/promenade/test/integration"
)

// TestProductRepository_Create tests creating a new product
func TestProductRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	p, _ := product.NewProduct("TEST-SKU-001", "Test Product")
	err := repo.Create(ctx, p)

	require.NoError(t, err)
	assert.NotEqual(t, "", p.GetID().String())
}

// TestProductRepository_GetByID tests retrieving product by ID
func TestProductRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create product
	p, _ := product.NewProduct("TEST-SKU-002", "Test Product 2")
	_ = repo.Create(ctx, p)

	// Retrieve product
	retrieved, err := repo.GetByID(ctx, p.GetID())

	require.NoError(t, err)
	assert.Equal(t, p.GetID(), retrieved.GetID())
	assert.Equal(t, "TEST-SKU-002", retrieved.SKU)
	assert.Equal(t, "Test Product 2", retrieved.Name)
}

// TestProductRepository_GetBySKU tests retrieving product by SKU
func TestProductRepository_GetBySKU(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create product
	p, _ := product.NewProduct("TEST-SKU-003", "Test Product 3")
	_ = repo.Create(ctx, p)

	// Retrieve by SKU
	retrieved, err := repo.GetBySKU(ctx, "TEST-SKU-003")

	require.NoError(t, err)
	assert.Equal(t, p.GetID(), retrieved.GetID())
	assert.Equal(t, "TEST-SKU-003", retrieved.SKU)
}

// TestProductRepository_Update tests updating a product
func TestProductRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create product
	p, _ := product.NewProduct("TEST-SKU-004", "Test Product 4")
	_ = repo.Create(ctx, p)

	// Update product
	p.SetDescription("Updated description")
	err := repo.Update(ctx, p)

	require.NoError(t, err)

	// Verify update
	retrieved, _ := repo.GetByID(ctx, p.GetID())
	assert.Equal(t, "Updated description", retrieved.Description)
}

// TestProductRepository_Delete tests soft deleting a product
func TestProductRepository_Delete(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create product
	p, _ := product.NewProduct("TEST-SKU-005", "Test Product 5")
	_ = repo.Create(ctx, p)

	// Delete product
	err := repo.Delete(ctx, p.GetID())
	require.NoError(t, err)

	// Verify deletion (should not find)
	_, err = repo.GetByID(ctx, p.GetID())
	assert.Error(t, err)
}

// TestProductRepository_List tests listing products with pagination
func TestProductRepository_List(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create 3 products
	for i := 1; i <= 3; i++ {
		p, _ := product.NewProduct(integration.FakeSKU(i), integration.FakeName("Product", i))
		_ = repo.Create(ctx, p)
	}

	// List products
	products, err := repo.List(ctx, 1, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(products), 3)
}

// TestProductRepository_ListByCategory tests listing products by category
func TestProductRepository_ListByCategory(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create product with category
	p, _ := product.NewProduct("TEST-SKU-006", "Test Product 6")
	p.Category = "Electronics"
	_ = repo.Create(ctx, p)

	// List by category
	products, err := repo.ListByCategory(ctx, "Electronics", 1, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(products), 1)
}

// TestProductRepository_ListByStatus tests listing products by status
func TestProductRepository_ListByStatus(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create active product
	p, _ := product.NewProduct("TEST-SKU-007", "Test Product 7")
	_ = p.Activate()
	_ = repo.Create(ctx, p)

	// List by status
	products, err := repo.ListByStatus(ctx, product.ProductStatusActive, 1, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(products), 1)
	assert.Equal(t, product.ProductStatusActive, products[0].Status)
}

// TestProductRepository_Search tests searching products
func TestProductRepository_Search(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create product with searchable name
	p, _ := product.NewProduct("TEST-SKU-008", "Unique Widget")
	_ = repo.Create(ctx, p)

	// Search by name
	products, err := repo.Search(ctx, "Unique", 1, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(products), 1)
}

// TestProductRepository_Count tests counting products
func TestProductRepository_Count(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := productRepo.NewProductRepository(db.DB)
	ctx := context.Background()

	// Create 2 products
	p1, _ := product.NewProduct(integration.FakeSKU(1), "Product 1")
	p2, _ := product.NewProduct(integration.FakeSKU(2), "Product 2")
	_ = repo.Create(ctx, p1)
	_ = repo.Create(ctx, p2)

	// Count products
	count, err := repo.Count(ctx)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 2)
}
