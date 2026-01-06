package warehouse

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	inventoryHTTP "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/http"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/warehouse/product"
	productHTTP "github.com/basilex/promenade/internal/contexts/warehouse/product/adapter/http"
	productRepo "github.com/basilex/promenade/internal/contexts/warehouse/product/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	stockmovementHTTP "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/adapter/http"
	stockmovementRepo "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/adapter/repository/postgres"
)

// Router handles all Warehouse context routes
type Router struct {
	inventoryHandler     *inventoryHTTP.InventoryHandler
	productHandler       *productHTTP.ProductHandler
	stockMovementHandler *stockmovementHTTP.StockMovementHandler
}

// NewRouter creates a new Warehouse context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Inventory aggregate
	inventoryRepository := inventoryRepo.NewInventoryRepository(db)
	inventoryUseCase := inventory.NewUseCase(inventoryRepository)
	inventoryHandler := inventoryHTTP.NewInventoryHandler(inventoryUseCase)

	// Initialize Product aggregate
	productRepository := productRepo.NewProductRepository(db)
	productUseCase := product.NewUseCase(productRepository)
	productHandler := productHTTP.NewProductHandler(productUseCase)

	// Initialize StockMovement aggregate
	stockMovementRepository := stockmovementRepo.NewStockMovementRepository(db)
	stockMovementUseCase := stockmovement.NewUseCase(stockMovementRepository)
	stockMovementHandler := stockmovementHTTP.NewStockMovementHandler(stockMovementUseCase)

	return &Router{
		inventoryHandler:     inventoryHandler,
		productHandler:       productHandler,
		stockMovementHandler: stockMovementHandler,
	}
}

// RegisterRoutes registers all Warehouse context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	warehouse := api.Group("/warehouse")
	{
		// Product routes
		products := warehouse.Group("/products")
		{
			// Product CRUD
			products.POST("", r.productHandler.Create)                              // Create product
			products.GET("/:id", r.productHandler.GetByID)                          // Get by ID
			products.PUT("/:id", r.productHandler.Update)                           // Update product
			products.DELETE("/:id", r.productHandler.Delete)                        // Soft delete product
			products.GET("", r.productHandler.List)                                 // List products (paginated)

			// Product queries
			products.GET("/sku/:sku", r.productHandler.GetBySKU)                    // Get by SKU
			products.GET("/category/:category", r.productHandler.ListByCategory)    // List by category
			products.GET("/brand/:brand", r.productHandler.ListByBrand)             // List by brand
			products.GET("/status/:status", r.productHandler.ListByStatus)          // List by status
			products.GET("/search", r.productHandler.Search)                        // Search products

			// Product operations
			products.POST("/:id/activate", r.productHandler.Activate)                         // Activate product
			products.POST("/:id/deactivate", r.productHandler.Deactivate)                     // Deactivate product
			products.POST("/:id/discontinue", r.productHandler.Discontinue)                   // Discontinue product
			products.PUT("/:id/inventory-settings", r.productHandler.UpdateInventorySettings) // Update inventory settings
			products.PUT("/:id/reorder-point", r.productHandler.SetReorderPoint)              // Set reorder point
			products.PUT("/:id/physical", r.productHandler.SetPhysicalProperties)             // Set physical properties
		}

		// Inventory routes
		inventory := warehouse.Group("/inventory")
		{
			// Inventory CRUD
			inventory.POST("", r.inventoryHandler.Create)                                              // Create inventory item
			inventory.GET("/:id", r.inventoryHandler.GetByID)                                          // Get by ID
			inventory.PUT("/:id", r.inventoryHandler.Update)                                           // Update inventory item
			inventory.DELETE("/:id", r.inventoryHandler.Delete)                                        // Soft delete inventory item
			inventory.GET("", r.inventoryHandler.List)                                                 // List inventory (paginated)

			// Inventory queries
			inventory.GET("/sku/:sku", r.inventoryHandler.GetBySKU)                                    // Get by SKU
			inventory.GET("/product/:product_id", r.inventoryHandler.GetByProductID)                   // Get by product ID
			inventory.GET("/warehouse/:warehouse_id", r.inventoryHandler.GetByWarehouse)               // Get by warehouse
			inventory.GET("/location/:warehouse_id/:location_code", r.inventoryHandler.GetByLocation)  // Get by location
			inventory.GET("/low-stock", r.inventoryHandler.GetLowStock)                                // Get low stock items

			// Stock operations
			inventory.POST("/:id/receive", r.inventoryHandler.ReceiveStock)         // Receive stock
			inventory.POST("/:id/reserve", r.inventoryHandler.ReserveStock)         // Reserve stock for order
			inventory.POST("/:id/release", r.inventoryHandler.ReleaseReservation)   // Release reservation (Saga compensation)
			inventory.POST("/:id/commit", r.inventoryHandler.CommitStock)           // Commit stock
		}

		// StockMovement routes
		stockMovements := warehouse.Group("/stock-movements")
		{
			// Basic CRUD
			stockMovements.POST("", r.stockMovementHandler.RecordMovement)                              // Record generic movement
			stockMovements.GET("/:id", r.stockMovementHandler.GetByID)                                  // Get movement by ID
			
			// Queries
			stockMovements.GET("/inventory/:inventory_id", r.stockMovementHandler.GetByInventory)       // Get by inventory (paginated)
			stockMovements.GET("/reference", r.stockMovementHandler.GetByReference)                     // Get by reference (query params: type, id)
			stockMovements.GET("/type/:type", r.stockMovementHandler.GetByType)                         // Get by type (paginated, last 30 days)
			stockMovements.GET("/recent", r.stockMovementHandler.GetRecent)                             // Get recent movements (last 50)
			stockMovements.GET("/summary/:inventory_id", r.stockMovementHandler.GetInventorySummary)    // Get inventory summary (last 30 days)
		}
	}
}
