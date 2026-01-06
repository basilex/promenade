package warehouse

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	inventoryHTTP "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/http"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
)

// Router handles all Warehouse context routes
type Router struct {
	inventoryHandler *inventoryHTTP.InventoryHandler
}

// NewRouter creates a new Warehouse context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Inventory aggregate
	inventoryRepository := inventoryRepo.NewInventoryRepository(db)
	inventoryUseCase := inventory.NewUseCase(inventoryRepository)
	inventoryHandler := inventoryHTTP.NewInventoryHandler(inventoryUseCase)

	return &Router{
		inventoryHandler: inventoryHandler,
	}
}

// RegisterRoutes registers all Warehouse context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	warehouse := api.Group("/warehouse")
	{
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
	}
}
