# Module Development Guide

This guide explains how to develop custom modules for Promenade using the Plugin Architecture.

## Table of Contents

- [Overview](#overview)
- [Module Structure](#module-structure)
- [Creating a Module](#creating-a-module)
- [Module Lifecycle](#module-lifecycle)
- [Inter-Module Communication](#inter-module-communication)
- [Best Practices](#best-practices)
- [Examples](#examples)

---

## Overview

Promenade uses a **Plugin Architecture** that allows you to:

- Package business logic as independent, reusable modules
- Share core infrastructure (DB, Event IBus, Auth, RBAC)
- Enable/disable modules via configuration
- Sell modules commercially with licensing
- Combine multiple modules (e.g., Warehouse + Fleet)

### Core vs Modules

**Core** (always enabled):

- Authentication & JWT
- RBAC & Permissions
- Event IBus & Notifications
- Reference Data (Countries, Currencies)
- Audit Logging

**Modules** (optional):

- Posts, Comments, Profiles (social features)
- Warehouse Management (inventory)
- Fleet Management (vehicles)
- Finance (billing, invoices)
- Custom business logic

---

## Module Structure

A typical module follows Clean Architecture:

```
modules/warehouse/
 domain/
    entity/
       item.go
    repository/
        item_repository.go

 usecase/
    item_usecase.go

 adapter/
    http/
       handler/
          item_handler.go
       dto/
           item_dto.go
    repository/
        postgres/
            item_repository.go

 migrations/
    001_create_warehouse_items.up.sql
    001_create_warehouse_items.down.sql

 module.go  # Module registration
```

---

## Creating a Module

### Step 1: Define Module Metadata

```go
// modules/warehouse/module.go
package warehouse

import (
	"github.com/basilex/promenade/pkg/module"
)

type WarehouseModule struct {
	*module.BaseModule

	// Dependencies
	itemRepo repository.ItemRepository
	itemUC   usecase.ItemUseCase
}

func New() module.IModule {
	meta := module.Metadata{
		Name:        "warehouse",
		DisplayName: "Warehouse Management",
		Version:     "1.2.0",
		Author:      "Your Company",
		Description: "Inventory and stock management system",
		License:     "Commercial",
		Tags:        []string{"inventory", "logistics"},
	}

	return &WarehouseModule{
		BaseModule: module.NewBaseModule(meta),
	}
}
```

### Step 2: Implement Module Interface

```go
// Dependencies (optional - return empty slice if none)
func (m *WarehouseModule) Dependencies() []string {
	return []string{} // No dependencies
}

// Initialize - set up repositories and use cases
func (m *WarehouseModule) Initialize(ctx context.Context, core *module.Core) error {
	// Call base implementation
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	// Initialize repositories
	m.itemRepo = postgres.NewItemRepository(core.DB)

	// Initialize use cases
	m.itemUC = usecase.NewItemUseCase(m.itemRepo, core.EventBus)

	return nil
}

// RegisterRoutes - register HTTP endpoints
func (m *WarehouseModule) RegisterRoutes(router *gin.RouterGroup) {
	handler := handler.NewItemHandler(m.itemUC)

	items := router.Group("/warehouse/items")
	{
		items.GET("", handler.List)      // GET /api/v1/warehouse/items
		items.POST("", handler.Create)   // POST /api/v1/warehouse/items
		items.GET("/:id", handler.Get)   // GET /api/v1/warehouse/items/:id
		items.PUT("/:id", handler.Update)
		items.DELETE("/:id", handler.Delete)
	}
}

// RegisterMigrations - return database migrations
func (m *WarehouseModule) RegisterMigrations() []module.Migration {
	return []module.Migration{
		{
			Version:     1,
			Description: "Create warehouse_items table",
			Up: `
				CREATE TABLE warehouse_items (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
					name VARCHAR(255) NOT NULL,
					sku VARCHAR(100) UNIQUE NOT NULL,
					quantity INTEGER NOT NULL DEFAULT 0,
					price DECIMAL(10,2) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMP NOT NULL DEFAULT NOW()
				);
			`,
			Down: `DROP TABLE warehouse_items;`,
		},
	}
}

// RegisterEventHandlers - subscribe to events
func (m *WarehouseModule) RegisterEventHandlers(bus bus.IBus) error {
	// Subscribe to order events
	return bus.Subscribe(ctx, "order.created", m.handleOrderCreated)
}

func (m *WarehouseModule) handleOrderCreated(ctx context.Context, event bus.Event) error {
	// Reduce stock quantity
	// ...
	return nil
}

// RegisterPermissions - define RBAC permissions
func (m *WarehouseModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "warehouse:items", Action: "read", Description: "View warehouse items"},
		{Resource: "warehouse:items", Action: "create", Description: "Create warehouse items"},
		{Resource: "warehouse:items", Action: "update", Description: "Update warehouse items"},
		{Resource: "warehouse:items", Action: "delete", Description: "Delete warehouse items"},
	}
}

// Start - start background workers (optional)
func (m *WarehouseModule) Start(ctx context.Context) error {
	// Start inventory sync worker
	go m.syncInventoryWorker(ctx)
	return nil
}

// Stop - graceful shutdown (optional)
func (m *WarehouseModule) Stop(ctx context.Context) error {
	// Stop workers, close connections, etc.
	return nil
}

// HealthCheck - check module health (optional)
func (m *WarehouseModule) HealthCheck(ctx context.Context) error {
	// Check database connectivity, external APIs, etc.
	return nil
}
```

### Step 3: Register Module

```go
// modules/warehouse/register.go
package warehouse

import "github.com/basilex/promenade/pkg/module"

func init() {
	// Auto-register module on import
	module.DefaultRegistry.Register(New())
}
```

### Step 4: Enable Module

```yaml
# config/modules.yaml
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-ABC-123-XYZ" # Optional
      settings:
        max_items: 10000
```

---

## Module Lifecycle

**Startup Sequence:**

1. **Application Startup**

   - Read `config/modules.yaml`
   - Load enabled modules
   - Load per-module settings

2. **Initialize Core**

   - Database connection
   - Event bus
   - JWT manager
   - Logger

3. **Module.Initialize()** (in dependency order)

   - Resolve dependencies
   - Initialize repositories
   - Initialize use cases

4. **Module.RegisterRoutes()**

   - Register HTTP endpoints

5. **Module.RegisterMigrations()**

   - Apply database migrations

6. **Module.RegisterEventHandlers()**

   - Subscribe to events

7. **Module.RegisterPermissions()**

   - Insert RBAC permissions

8. **Module.Start()**

   - Start background workers
   - Initialize external connections

9. **Server Running**

**Shutdown Sequence:**

10. **Module.Stop()** (in reverse order)
    - Stop workers
    - Close connections
    - Cleanup resources

---

## Inter-Module Communication

Modules should communicate via **events** (loose coupling):

### Publishing Events

```go
// In Warehouse module
func (uc *ItemUseCase) CreateItem(ctx context.Context, item *entity.Item) error {
	// Save to database
	if err := uc.repo.Create(ctx, item); err != nil {
		return err
	}

	// Publish event
	event := &events.ItemCreatedEvent{
		ItemID:   item.ID,
		SKU:      item.SKU,
		Quantity: item.Quantity,
	}
	uc.eventBus.Publish(ctx, "warehouse.item.created", event)

	return nil
}
```

### Subscribing to Events

```go
// In Fleet module (needs warehouse items for spare parts)
func (m *FleetModule) RegisterEventHandlers(bus bus.IBus) error {
	return bus.Subscribe(ctx, "warehouse.item.created", m.handleItemCreated)
}

func (m *FleetModule) handleItemCreated(ctx context.Context, event bus.Event) error {
	itemEvent := event.(*events.ItemCreatedEvent)

	// Update spare parts inventory
	// ...

	return nil
}
```

---

## Best Practices

### DO

1. **Use BaseModule** - Embed `module.BaseModule` to avoid implementing every method
2. **Clean Architecture** - Follow domain/usecase/adapter structure
3. **UUID v7** - Use `uuidv7.New()` for primary keys
4. **Event-Driven** - Communicate between modules via events
5. **RBAC** - Register permissions for all endpoints
6. **Migrations** - Always provide Up and Down migrations
7. **Health Checks** - Implement `HealthCheck()` for monitoring
8. **Graceful Shutdown** - Cleanup resources in `Stop()`

### DON'T

1. **Direct Dependencies** - Never import other module's code directly
2. **Shared Tables** - Each module owns its tables
3. **Core Modifications** - Don't modify core code for module features
4. **Synchronous Calls** - Avoid blocking inter-module communication
5. **Global State** - Don't use global variables (use Core instead)

---

## Examples

### Example 1: Simple IModule (No Dependencies)

```go
package hello

type HelloModule struct {
	*module.BaseModule
}

func New() module.IModule {
	return &HelloModule{
		BaseModule: module.NewBaseModule(module.Metadata{
			Name:    "hello",
			Version: "1.0.0",
		}),
	}
}

func (m *HelloModule) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from module!"})
	})
}
```

### Example 2: IModule with Dependencies

```go
package fleet

type FleetModule struct {
	*module.BaseModule
}

func (m *FleetModule) Dependencies() []string {
	return []string{"warehouse"} // Fleet depends on Warehouse
}

func (m *FleetModule) Initialize(ctx context.Context, core *module.Core) error {
	// Warehouse must be initialized first (guaranteed by registry)
	warehouseModule := core.Registry.Get("warehouse")
	if warehouseModule == nil {
		return fmt.Errorf("warehouse module required but not loaded")
	}

	// Initialize fleet-specific logic
	// ...

	return nil
}
```

---

## Testing Modules

```go
// modules/warehouse/module_test.go
func TestWarehouseModule(t *testing.T) {
	// Setup test core
	core := setupTestCore(t)

	// Create module
	mod := warehouse.New()

	// Initialize
	err := mod.Initialize(context.Background(), core)
	assert.NoError(t, err)

	// Test routes
	router := gin.New()
	group := router.Group("/api/v1")
	mod.RegisterRoutes(group)

	// Test endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/warehouse/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
```

---

## Next Steps

1. See existing modules in `internal/modules/`
2. Use `make generate-module NAME=mymodule` (TODO)
3. Check `pkg/module/` for full API reference
4. Read [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) for Clean Architecture patterns

---

For questions, see [README.md](../README.md) or open an issue.
