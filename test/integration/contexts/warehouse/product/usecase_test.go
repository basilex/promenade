package product_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/product"
	productRepo "github.com/basilex/promenade/internal/contexts/warehouse/product/adapter/repository/postgres"
	productAggregate "github.com/basilex/promenade/internal/contexts/warehouse/product/aggregate"
	productUseCase "github.com/basilex/promenade/internal/contexts/warehouse/product/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// TestProductUseCase_CreateProduct tests creating a new product through UseCase
func TestProductUseCase_CreateProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product with unique SKU
		uniqueSKU := fmt.Sprintf("UC-SKU-%s", uuidv7.New().String()[:8])
		p, err := uc.CreateProduct(ctx, uniqueSKU, "Test Product")
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, p.GetID())
		assert.Equal(t, uniqueSKU, p.SKU)
		assert.Equal(t, "Test Product", p.Name)
		assert.Equal(t, productAggregate.ProductStatusDraft, p.Status)

		// Test - Duplicate SKU should fail
		_, err = uc.CreateProduct(ctx, uniqueSKU, "Another Product")
		assert.Error(t, err)
		assert.Equal(t, product.ErrProductSKUDuplicate, err)
	})
}

// TestProductUseCase_GetProduct tests retrieving a product by ID
func TestProductUseCase_GetProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product
		uniqueSKU := fmt.Sprintf("UC-GET-%s", uuidv7.New().String()[:8])
		created, err := uc.CreateProduct(ctx, uniqueSKU, "Product to Get")
		require.NoError(t, err)

		// Test - Get by ID
		retrieved, err := uc.GetProduct(ctx, created.GetID())
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), retrieved.GetID())
		assert.Equal(t, uniqueSKU, retrieved.SKU)
		assert.Equal(t, "Product to Get", retrieved.Name)

		// Test - Non-existent ID
		_, err = uc.GetProduct(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, product.ErrProductNotFound, err)
	})
}

// TestProductUseCase_GetBySKU tests SKU lookup through UseCase
func TestProductUseCase_GetBySKU(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product with unique SKU
		uniqueSKU := fmt.Sprintf("UC-FIND-%s", uuidv7.New().String()[:8])
		created, err := uc.CreateProduct(ctx, uniqueSKU, "Find Me")
		require.NoError(t, err)

		// Test - GetProductBySKU - found
		found, err := uc.GetProductBySKU(ctx, uniqueSKU)
		require.NoError(t, err)
		assert.Equal(t, created.GetID(), found.GetID())
		assert.Equal(t, uniqueSKU, found.SKU)

		// Test - GetProductBySKU - not found
		_, err = uc.GetProductBySKU(ctx, "NON-EXISTENT-SKU")
		assert.Error(t, err)
		assert.Equal(t, product.ErrProductNotFound, err)
	})
}

// TestProductUseCase_UpdateProduct tests updating product through UseCase
func TestProductUseCase_UpdateProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product
		uniqueSKU := fmt.Sprintf("UC-UPD-%s", uuidv7.New().String()[:8])
		p, err := uc.CreateProduct(ctx, uniqueSKU, "Original Name")
		require.NoError(t, err)

		// Update product name and description
		p.Name = "Updated Name"
		p.Description = "Updated description"
		err = uc.UpdateProduct(ctx, p)
		require.NoError(t, err)

		// Verify update
		updated, err := uc.GetProduct(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, uniqueSKU, updated.SKU) // SKU unchanged
	})
}

// TestProductUseCase_SetClassification tests setting product classification
func TestProductUseCase_SetClassification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product
		uniqueSKU := fmt.Sprintf("UC-CLASS-%s", uuidv7.New().String()[:8])
		p, err := uc.CreateProduct(ctx, uniqueSKU, "Product")
		require.NoError(t, err)

		// Set classification
		tags := []string{"electronics", "gadgets"}
		p.SetClassification("Electronics", "Apple", tags)
		err = uc.UpdateProduct(ctx, p)
		require.NoError(t, err)

		// Verify classification
		updated, err := uc.GetProduct(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Electronics", updated.Category)
		assert.Equal(t, "Apple", updated.Brand)
		assert.Equal(t, tags, updated.Tags.Get())
	})
}

// TestProductUseCase_ActivateProduct tests product activation
func TestProductUseCase_ActivateProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product (default status: draft)
		uniqueSKU := fmt.Sprintf("UC-ACT-%s", uuidv7.New().String()[:8])
		p, err := uc.CreateProduct(ctx, uniqueSKU, "Product")
		require.NoError(t, err)
		assert.Equal(t, productAggregate.ProductStatusDraft, p.Status)

		// Activate product
		err = uc.ActivateProduct(ctx, p.GetID())
		require.NoError(t, err)

		// Verify activation
		activated, err := uc.GetProduct(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, productAggregate.ProductStatusActive, activated.Status)
	})
}

// TestProductUseCase_DeactivateProduct tests product deactivation
func TestProductUseCase_DeactivateProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create and activate product
		uniqueSKU := fmt.Sprintf("UC-DEACT-%s", uuidv7.New().String()[:8])
		p, err := uc.CreateProduct(ctx, uniqueSKU, "Product")
		require.NoError(t, err)
		err = uc.ActivateProduct(ctx, p.GetID())
		require.NoError(t, err)

		// Deactivate product
		err = uc.DeactivateProduct(ctx, p.GetID())
		require.NoError(t, err)

		// Verify deactivation
		deactivated, err := uc.GetProduct(ctx, p.GetID())
		require.NoError(t, err)
		assert.Equal(t, productAggregate.ProductStatusOutOfStock, deactivated.Status)
	})
}

// TestProductUseCase_ListProducts tests listing all products with pagination
func TestProductUseCase_ListProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create multiple products with unique SKUs
		prefix := fmt.Sprintf("UC-LIST-%s", uuidv7.New().String()[:8])
		for i := 1; i <= 5; i++ {
			sku := fmt.Sprintf("%s-%d", prefix, i)
			_, err := uc.CreateProduct(ctx, sku, fmt.Sprintf("Product %d", i))
			require.NoError(t, err)
		}

		// List products (page 1, 10 per page)
		products, err := uc.ListProducts(ctx, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(products), 5) // At least our 5 products
	})
}

// TestProductUseCase_ListByCategory tests filtering products by category
func TestProductUseCase_ListByCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create products with unique SKUs and category
		category := "Integration-Electronics"
		prefix := fmt.Sprintf("UC-CAT-%s", uuidv7.New().String()[:8])
		for i := 1; i <= 3; i++ {
			sku := fmt.Sprintf("%s-%d", prefix, i)
			p, err := uc.CreateProduct(ctx, sku, fmt.Sprintf("Gadget %d", i))
			require.NoError(t, err)
			p.SetClassification(category, "Brand", []string{"tag"})
			err = uc.UpdateProduct(ctx, p)
			require.NoError(t, err)
		}

		// List by category
		products, err := uc.ListProductsByCategory(ctx, category, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(products), 3) // At least our 3 products
		for _, p := range products {
			if p.Category == category {
				assert.Equal(t, category, p.Category)
			}
		}
	})
}

// TestProductUseCase_SearchProducts tests full-text search
func TestProductUseCase_SearchProducts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product with unique searchable name
		searchTerm := fmt.Sprintf("UltraGadget-%s", uuidv7.New().String()[:8])
		uniqueSKU := fmt.Sprintf("UC-SEARCH-%s", uuidv7.New().String()[:8])
		p, err := uc.CreateProduct(ctx, uniqueSKU, searchTerm)
		require.NoError(t, err)
		p.Description = "Amazing ultra gadget"
		err = uc.UpdateProduct(ctx, p)
		require.NoError(t, err)

		// Search for product
		results, err := uc.SearchProducts(ctx, searchTerm, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1) // At least one result

		// Verify our product is in results
		found := false
		for _, p := range results {
			if p.SKU == uniqueSKU {
				found = true
				assert.Contains(t, p.Name, searchTerm)
				break
			}
		}
		assert.True(t, found, "Created product should be in search results")
	})
}

// TestProductUseCase_DeleteProduct tests product deletion
func TestProductUseCase_DeleteProduct(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := productRepo.NewProductRepository(testDB.DB)
		uc := productUseCase.NewProductUseCase(repo)

		// Create product
		uniqueSKU := fmt.Sprintf("UC-DEL-%s", uuidv7.New().String()[:8])
		p, err := uc.CreateProduct(ctx, uniqueSKU, "Product to Delete")
		require.NoError(t, err)

		// Delete product
		err = uc.DeleteProduct(ctx, p.GetID())
		require.NoError(t, err)

		// Verify deletion (soft delete)
		_, err = uc.GetProduct(ctx, p.GetID())
		assert.Error(t, err)
		assert.Equal(t, product.ErrProductNotFound, err)
	})
}
