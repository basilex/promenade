package warehouse

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/module"
	moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

// WarehouseModule implements the Module interface for warehouse management
// This is a COMMERCIAL module (requires license)
type WarehouseModule struct {
	*module.BaseModule

	// Dependencies
	db              *sqlx.DB
	authMiddleware  *middleware.AuthMiddleware
	authzMiddleware *middleware.AuthorizationMiddleware

	// Module configuration
	config *moduleconfig.Config

	// Module-specific configuration
	maxItems             int
	enableBarcodeScanner bool
}

// New creates a new warehouse module
func New() module.Module {
	meta := module.Metadata{
		Name:        "warehouse",
		DisplayName: "Warehouse Management",
		Version:     "1.2.0",
		Author:      "Promenade Commercial",
		Description: "Complete inventory and warehouse management system with barcode scanning, stock tracking, and automated reordering",
		License:     "Commercial", // Requires license key
		Tags:        []string{"inventory", "logistics", "commercial"},
	}

	return &WarehouseModule{
		BaseModule: module.NewBaseModule(meta),
	}
}

// Dependencies returns empty slice (no dependencies on other modules)
func (m *WarehouseModule) Dependencies() []string {
	return []string{} // Independent module
}

// Initialize sets up the warehouse module
func (m *WarehouseModule) Initialize(ctx context.Context, core *module.Core) error {
	// Call base implementation
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	// Load module configuration from its own directory
	config, err := moduleconfig.Load("internal/modules/warehouse/config", "promenade")
	if err != nil {
		return fmt.Errorf("failed to load warehouse config: %w", err)
	}
	m.config = config

	// License verification (commercial module)
	if err := m.verifyLicense(); err != nil {
		return fmt.Errorf("warehouse module license verification failed: %w", err)
	}

	// Load module-specific settings
	if err := m.loadSettings(); err != nil {
		return fmt.Errorf("failed to load warehouse settings: %w", err)
	}

	// Initialize dependencies
	m.db = core.DB
	m.authMiddleware = middleware.NewAuthMiddleware(core.JWT)

	// TODO: Initialize warehouse-specific repositories and use cases
	// warehouseRepo := postgres.NewWarehouseItemRepository(core.DB)
	// warehouseUseCase := usecase.NewWarehouseUseCase(warehouseRepo, core.EventBus)

	return nil
}

// verifyLicense checks if valid license key is provided
func (m *WarehouseModule) verifyLicense() error {
	if m.config == nil {
		return fmt.Errorf("warehouse module configuration not loaded")
	}

	licenseKey, ok := m.config.Settings["module"].(map[string]interface{})["license_key"].(string)
	if !ok || licenseKey == "" {
		return fmt.Errorf("warehouse module requires license key (set in internal/modules/warehouse/config/config.*.yaml)")
	}

	// TODO: Implement JWT-based license verification
	// For now, just check if key is present
	// In production: verify with license server or validate JWT signature

	// Example validation (simplified)
	if len(licenseKey) < 10 {
		return fmt.Errorf("invalid license key format")
	}

	return nil
}

// loadSettings loads module-specific settings
func (m *WarehouseModule) loadSettings() error {
	if m.config == nil {
		// Use defaults
		m.maxItems = 10000
		m.enableBarcodeScanner = true
		return nil
	}

	// Load settings using config helper methods
	if maxItems, ok := m.config.GetNestedSetting("settings", "max_items").(int); ok {
		m.maxItems = maxItems
	} else {
		m.maxItems = 10000
	}

	if enableBarcode, ok := m.config.GetNestedSetting("settings", "enable_barcode_scanner").(bool); ok {
		m.enableBarcodeScanner = enableBarcode
	} else {
		m.enableBarcodeScanner = true
	}

	return nil
}

// RegisterRoutes registers HTTP routes for warehouse
func (m *WarehouseModule) RegisterRoutes(router *gin.RouterGroup) {
	// Warehouse routes (protected by auth)
	warehouse := router.Group("/warehouse")
	warehouse.Use(m.authMiddleware.RequireAuth())
	{
		// Items
		items := warehouse.Group("/items")
		{
			items.GET("", m.listItems)         // GET /api/v1/warehouse/items
			items.POST("", m.createItem)       // POST /api/v1/warehouse/items
			items.GET("/:id", m.getItem)       // GET /api/v1/warehouse/items/:id
			items.PUT("/:id", m.updateItem)    // PUT /api/v1/warehouse/items/:id
			items.DELETE("/:id", m.deleteItem) // DELETE /api/v1/warehouse/items/:id
		}

		// Stock operations
		stock := warehouse.Group("/stock")
		{
			stock.POST("/in", m.stockIn)             // POST /api/v1/warehouse/stock/in
			stock.POST("/out", m.stockOut)           // POST /api/v1/warehouse/stock/out
			stock.POST("/transfer", m.stockTransfer) // POST /api/v1/warehouse/stock/transfer
		}

		// Reports
		reports := warehouse.Group("/reports")
		{
			reports.GET("/inventory", m.inventoryReport) // GET /api/v1/warehouse/reports/inventory
			reports.GET("/movements", m.movementReport)  // GET /api/v1/warehouse/reports/movements
			reports.GET("/low-stock", m.lowStockReport)  // GET /api/v1/warehouse/reports/low-stock
		}

		// Barcode scanner (if enabled)
		if m.enableBarcodeScanner {
			warehouse.POST("/scan", m.scanBarcode) // POST /api/v1/warehouse/scan
		}
	}
}

// Placeholder handlers (to be implemented with full business logic)
func (m *WarehouseModule) listItems(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Warehouse items list",
		"module":  "warehouse",
		"version": m.Metadata().Version,
	})
}

func (m *WarehouseModule) createItem(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func (m *WarehouseModule) getItem(c *gin.Context)    { c.JSON(501, gin.H{"error": "not implemented"}) }
func (m *WarehouseModule) updateItem(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func (m *WarehouseModule) deleteItem(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func (m *WarehouseModule) stockIn(c *gin.Context)    { c.JSON(501, gin.H{"error": "not implemented"}) }
func (m *WarehouseModule) stockOut(c *gin.Context)   { c.JSON(501, gin.H{"error": "not implemented"}) }
func (m *WarehouseModule) stockTransfer(c *gin.Context) {
	c.JSON(501, gin.H{"error": "not implemented"})
}
func (m *WarehouseModule) inventoryReport(c *gin.Context) {
	c.JSON(501, gin.H{"error": "not implemented"})
}
func (m *WarehouseModule) movementReport(c *gin.Context) {
	c.JSON(501, gin.H{"error": "not implemented"})
}
func (m *WarehouseModule) lowStockReport(c *gin.Context) {
	c.JSON(501, gin.H{"error": "not implemented"})
}
func (m *WarehouseModule) scanBarcode(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }

// RegisterMigrations returns database migrations for warehouse
func (m *WarehouseModule) RegisterMigrations() []module.Migration {
	return []module.Migration{
		{
			Version:     100,
			Description: "Create warehouse_items table",
			Up: `
				CREATE TABLE IF NOT EXISTS warehouse_items (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
					sku VARCHAR(100) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					description TEXT,
					category VARCHAR(100),
					quantity INTEGER NOT NULL DEFAULT 0,
					min_quantity INTEGER NOT NULL DEFAULT 0,
					max_quantity INTEGER,
					unit_price DECIMAL(10,2) NOT NULL,
					barcode VARCHAR(50),
					location VARCHAR(100),
					supplier VARCHAR(255),
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMP NOT NULL DEFAULT NOW()
				);

				CREATE INDEX idx_warehouse_items_sku ON warehouse_items(sku);
				CREATE INDEX idx_warehouse_items_category ON warehouse_items(category);
				CREATE INDEX idx_warehouse_items_barcode ON warehouse_items(barcode);

				COMMENT ON TABLE warehouse_items IS 'Warehouse inventory items';
			`,
			Down: `DROP TABLE IF EXISTS warehouse_items;`,
		},
		{
			Version:     101,
			Description: "Create warehouse_movements table",
			Up: `
				CREATE TABLE IF NOT EXISTS warehouse_movements (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
					item_id UUID NOT NULL REFERENCES warehouse_items(id) ON DELETE CASCADE,
					type VARCHAR(20) NOT NULL CHECK (type IN ('in', 'out', 'transfer', 'adjustment')),
					quantity INTEGER NOT NULL,
					from_location VARCHAR(100),
					to_location VARCHAR(100),
					reason TEXT,
					user_id UUID NOT NULL REFERENCES users(id),
					created_at TIMESTAMP NOT NULL DEFAULT NOW()
				);

				CREATE INDEX idx_warehouse_movements_item_id ON warehouse_movements(item_id);
				CREATE INDEX idx_warehouse_movements_type ON warehouse_movements(type);
				CREATE INDEX idx_warehouse_movements_created_at ON warehouse_movements(created_at DESC);

				COMMENT ON TABLE warehouse_movements IS 'Warehouse stock movement history';
			`,
			Down: `DROP TABLE IF EXISTS warehouse_movements;`,
		},
	}
}

// RegisterEventHandlers subscribes to domain events
func (m *WarehouseModule) RegisterEventHandlers(eventBus bus.Bus) error {
	// Subscribe to order events (if orders module exists)
	// return eventBus.Subscribe(ctx, "order.created", m.handleOrderCreated)
	return nil
}

// RegisterPermissions returns RBAC permissions for warehouse
func (m *WarehouseModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "warehouse:items", Action: "read", Description: "View warehouse items"},
		{Resource: "warehouse:items", Action: "create", Description: "Create warehouse items"},
		{Resource: "warehouse:items", Action: "update", Description: "Update warehouse items"},
		{Resource: "warehouse:items", Action: "delete", Description: "Delete warehouse items"},
		{Resource: "warehouse:stock", Action: "in", Description: "Receive stock"},
		{Resource: "warehouse:stock", Action: "out", Description: "Issue stock"},
		{Resource: "warehouse:stock", Action: "transfer", Description: "Transfer stock between locations"},
		{Resource: "warehouse:reports", Action: "read", Description: "View warehouse reports"},
	}
}

// Start is called after initialization (start background workers)
func (m *WarehouseModule) Start(ctx context.Context) error {
	// Start low-stock monitor worker
	go m.monitorLowStock(ctx)
	return nil
}

// monitorLowStock checks for items below minimum quantity
func (m *WarehouseModule) monitorLowStock(ctx context.Context) {
	// TODO: Implement periodic check for low stock items
	// and send notifications
}

// Stop is called during graceful shutdown
func (m *WarehouseModule) Stop(ctx context.Context) error {
	// Stop background workers
	return nil
}

// HealthCheck checks if warehouse module is healthy
func (m *WarehouseModule) HealthCheck(ctx context.Context) error {
	if m.db == nil {
		return module.ErrModuleNotInitialized
	}
	// Check license is still valid
	return nil
}
