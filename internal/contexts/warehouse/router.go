package warehouse

import (
	inventoryHTTP "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/http"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	inventoryusecase "github.com/basilex/promenade/internal/contexts/warehouse/inventory/usecase"
	locationHTTP "github.com/basilex/promenade/internal/contexts/warehouse/location/adapter/http"
	locationRepo "github.com/basilex/promenade/internal/contexts/warehouse/location/adapter/repository/postgres"
	locationusecase "github.com/basilex/promenade/internal/contexts/warehouse/location/usecase"
	productHTTP "github.com/basilex/promenade/internal/contexts/warehouse/product/adapter/http"
	productRepo "github.com/basilex/promenade/internal/contexts/warehouse/product/adapter/repository/postgres"
	productusecase "github.com/basilex/promenade/internal/contexts/warehouse/product/usecase"
	stockmovementHTTP "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/adapter/http"
	stockmovementRepo "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/adapter/repository/postgres"
	stockmovementusecase "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/usecase"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Router handles all Warehouse context routes
type Router struct {
	inventoryHandler     *inventoryHTTP.InventoryHandler
	productHandler       *productHTTP.ProductHandler
	stockMovementHandler *stockmovementHTTP.StockMovementHandler
	locationHandler      *locationHTTP.LocationHandler
}

// NewRouter creates a new Warehouse context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Inventory aggregate
	inventoryRepository := inventoryRepo.NewInventoryRepository(db)
	inventoryUseCase := inventoryusecase.NewInventoryUseCase(inventoryRepository)
	inventoryHandler := inventoryHTTP.NewInventoryHandler(inventoryUseCase)

	// Initialize Product aggregate
	productRepository := productRepo.NewProductRepository(db)
	productUseCase := productusecase.NewProductUseCase(productRepository)
	productHandler := productHTTP.NewProductHandler(productUseCase)

	// Initialize StockMovement aggregate
	stockMovementRepository := stockmovementRepo.NewStockMovementRepository(db)
	stockMovementUseCase := stockmovementusecase.NewStockMovementUseCase(stockMovementRepository)
	stockMovementHandler := stockmovementHTTP.NewStockMovementHandler(stockMovementUseCase)

	// Initialize Location aggregate
	locationRepository := locationRepo.NewLocationRepository(db)
	locationUseCase := locationusecase.NewLocationUseCase(locationRepository)
	locationHandler := locationHTTP.NewLocationHandler(locationUseCase)

	return &Router{
		inventoryHandler:     inventoryHandler,
		productHandler:       productHandler,
		stockMovementHandler: stockMovementHandler,
		locationHandler:      locationHandler,
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
			products.POST("", r.productHandler.Create)       // Create product
			products.GET("/:id", r.productHandler.GetByID)   // Get by ID
			products.PUT("/:id", r.productHandler.Update)    // Update product
			products.DELETE("/:id", r.productHandler.Delete) // Soft delete product
			products.GET("", r.productHandler.List)          // List products (paginated)

			// Product queries
			products.GET("/sku/:sku", r.productHandler.GetBySKU)                 // Get by SKU
			products.GET("/category/:category", r.productHandler.ListByCategory) // List by category
			products.GET("/brand/:brand", r.productHandler.ListByBrand)          // List by brand
			products.GET("/status/:status", r.productHandler.ListByStatus)       // List by status
			products.GET("/search", r.productHandler.Search)                     // Search products

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
			inventory.POST("", r.inventoryHandler.Create)       // Create inventory item
			inventory.GET("/:id", r.inventoryHandler.GetByID)   // Get by ID
			inventory.PUT("/:id", r.inventoryHandler.Update)    // Update inventory item
			inventory.DELETE("/:id", r.inventoryHandler.Delete) // Soft delete inventory item
			inventory.GET("", r.inventoryHandler.List)          // List inventory (paginated)

			// Inventory queries
			inventory.GET("/sku/:sku", r.inventoryHandler.GetBySKU)                                   // Get by SKU
			inventory.GET("/product/:product_id", r.inventoryHandler.GetByProductID)                  // Get by product ID
			inventory.GET("/warehouse/:warehouse_id", r.inventoryHandler.GetByWarehouse)              // Get by warehouse
			inventory.GET("/location/:warehouse_id/:location_code", r.inventoryHandler.GetByLocation) // Get by location
			inventory.GET("/low-stock", r.inventoryHandler.GetLowStock)                               // Get low stock items

			// Stock operations
			inventory.POST("/:id/receive", r.inventoryHandler.ReceiveStock)       // Receive stock
			inventory.POST("/:id/reserve", r.inventoryHandler.ReserveStock)       // Reserve stock for order
			inventory.POST("/:id/release", r.inventoryHandler.ReleaseReservation) // Release reservation (Saga compensation)
			inventory.POST("/:id/commit", r.inventoryHandler.CommitStock)         // Commit stock
		}

		// StockMovement routes
		stockMovements := warehouse.Group("/stock-movements")
		{
			// Basic CRUD
			stockMovements.POST("", r.stockMovementHandler.RecordMovement) // Record generic movement
			stockMovements.GET("/:id", r.stockMovementHandler.GetByID)     // Get movement by ID

			// Queries
			stockMovements.GET("/inventory/:inventory_id", r.stockMovementHandler.GetByInventory)    // Get by inventory (paginated)
			stockMovements.GET("/reference", r.stockMovementHandler.GetByReference)                  // Get by reference (query params: type, id)
			stockMovements.GET("/type/:type", r.stockMovementHandler.GetByType)                      // Get by type (paginated, last 30 days)
			stockMovements.GET("/recent", r.stockMovementHandler.GetRecent)                          // Get recent movements (last 50)
			stockMovements.GET("/summary/:inventory_id", r.stockMovementHandler.GetInventorySummary) // Get inventory summary (last 30 days)
		}

		// Location routes
		locations := warehouse.Group("/locations")
		{
			// Location CRUD
			locations.POST("", r.locationHandler.Create)       // Create location
			locations.GET("/:id", r.locationHandler.GetByID)   // Get by ID
			locations.PUT("/:id", r.locationHandler.Update)    // Update location
			locations.DELETE("/:id", r.locationHandler.Delete) // Soft delete location
			locations.GET("", r.locationHandler.List)          // List locations (paginated with filters)

			// Location queries
			locations.GET("/code/:code", r.locationHandler.GetByCode) // Get by unique code

			// Hierarchy operations
			locations.GET("/:id/children", r.locationHandler.GetChildren)   // Get child locations
			locations.GET("/:id/hierarchy", r.locationHandler.GetHierarchy) // Get hierarchy path

			// Status management
			locations.PUT("/:id/activate", r.locationHandler.Activate)     // Activate location
			locations.PUT("/:id/deactivate", r.locationHandler.Deactivate) // Deactivate location

			// Capacity management
			locations.PUT("/:id/capacity", r.locationHandler.UpdateCapacity)     // Update capacity settings
			locations.PUT("/:id/dimensions", r.locationHandler.UpdateDimensions) // Update dimensions (width, height, depth)
			locations.PUT("/:id/flags", r.locationHandler.UpdateFlags)           // Update operational flags (pickable, putawayable)
			locations.PUT("/:id/maintenance", r.locationHandler.SetMaintenance)  // Set maintenance mode
		}
	}
}
