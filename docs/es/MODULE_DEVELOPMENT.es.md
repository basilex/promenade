# Guía de Desarrollo de Módulos

[🇬🇧 English](../MODULE_DEVELOPMENT.md) | [🇺🇦 Українська](../uk/MODULE_DEVELOPMENT.uk.md) | [🇩🇪 Deutsch](../de/MODULE_DEVELOPMENT.de.md) | [🇵🇹 Português](../pt/MODULE_DEVELOPMENT.pt.md) | 🇪🇸 **Español**

Esta guía explica cómo desarrollar módulos personalizados para Promenade usando la Arquitectura de Plugin.

## Tabla de Contenidos

- [Visión General](#visión-general)
- [Estructura del Módulo](#estructura-del-módulo)
- [Creando un Módulo](#creando-un-módulo)
- [Ciclo de Vida del Módulo](#ciclo-de-vida-del-módulo)
- [Comunicación Inter-Módulos](#comunicación-inter-módulos)
- [Mejores Prácticas](#mejores-prácticas)
- [Ejemplos](#ejemplos)

---

## Visión General

Promenade usa una **Arquitectura de Plugin** que permite:

- Empaquetar lógica de negocio como módulos independientes y reutilizables
- Compartir infraestructura central (DB, Event IBus, Auth, RBAC)
- Habilitar/deshabilitar módulos vía configuración
- Vender módulos comercialmente con licenciamiento
- Combinar múltiples módulos (ej: Warehouse + Fleet)

### Core vs Módulos

**Core** (siempre habilitado):

- Autenticación & JWT
- RBAC & Permisos
- Event IBus & Notificaciones
- Datos de Referencia (Países, Monedas)
- Registro de Auditoría

**Módulos** (opcionales):

- Posts, Comments, Profiles (características sociales)
- Warehouse Management (gestión de inventario)
- Fleet Management (gestión de vehículos)
- Finance (facturación, facturas)
- Lógica de negocio personalizada

---

## Estructura del Módulo

Un módulo típico sigue Clean Architecture:

```
modules/warehouse/
├── domain/
│   ├── entity/
│   │   └── item.go
│   └── repository/
│       └── item_repository.go
│
├── usecase/
│   └── item_usecase.go
│
├── adapter/
│   ├── http/
│   │   ├── handler/
│   │   │   └── item_handler.go
│   │   └── dto/
│   │       └── item_dto.go
│   └── repository/
│       └── postgres/
│           └── item_repository.go
│
├── migrations/
│   ├── 001_create_warehouse_items.up.sql
│   └── 001_create_warehouse_items.down.sql
│
└── module.go  # Registro del módulo
```

---

## Creando un Módulo

### Paso 1: Definir Metadatos del Módulo

```go
// modules/warehouse/module.go
package warehouse

import (
	"github.com/basilex/promenade/pkg/module"
)

type WarehouseModule struct {
	*module.BaseModule

	// Dependencias
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

### Paso 2: Implementar Interfaz del Módulo

```go
// Dependencies (opcional - devolver array vacío si no hay)
func (m *WarehouseModule) Dependencies() []string {
	return []string{} // Sin dependencias
}

// Initialize - configurar repositorios y casos de uso
func (m *WarehouseModule) Initialize(ctx context.Context, core *module.Core) error {
	// Llamar implementación base
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	// Inicializar repositorios
	m.itemRepo = postgres.NewItemRepository(core.DB)

	// Inicializar casos de uso
	m.itemUC = usecase.NewItemUseCase(m.itemRepo, core.EventBus)

	return nil
}

// RegisterRoutes - registrar endpoints HTTP
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

// RegisterMigrations - devolver migraciones de base de datos
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

// RegisterEventHandlers - suscribirse a eventos
func (m *WarehouseModule) RegisterEventHandlers(bus bus.IBus) error {
	// Suscribirse a eventos de pedidos
	return bus.Subscribe(ctx, "order.created", m.handleOrderCreated)
}

func (m *WarehouseModule) handleOrderCreated(ctx context.Context, event bus.Event) error {
	// Reducir cantidad en stock
	// ...
	return nil
}

// RegisterPermissions - definir permisos RBAC
func (m *WarehouseModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "warehouse:items", Action: "read", Description: "View warehouse items"},
		{Resource: "warehouse:items", Action: "create", Description: "Create warehouse items"},
		{Resource: "warehouse:items", Action: "update", Description: "Update warehouse items"},
		{Resource: "warehouse:items", Action: "delete", Description: "Delete warehouse items"},
	}
}

// Start - iniciar workers en segundo plano (opcional)
func (m *WarehouseModule) Start(ctx context.Context) error {
	// Iniciar worker de sincronización de inventario
	go m.syncInventoryWorker(ctx)
	return nil
}

// Stop - apagado gracioso (opcional)
func (m *WarehouseModule) Stop(ctx context.Context) error {
	// Detener workers, cerrar conexiones, etc.
	return nil
}

// HealthCheck - verificar salud del módulo (opcional)
func (m *WarehouseModule) HealthCheck(ctx context.Context) error {
	// Verificar conectividad de base de datos, APIs externas, etc.
	return nil
}
```

### Paso 3: Registrar Módulo

```go
// modules/warehouse/register.go
package warehouse

import "github.com/basilex/promenade/pkg/module"

func init() {
	// Auto-registrar módulo al importar
	module.DefaultRegistry.Register(New())
}
```

### Paso 4: Habilitar Módulo

```yaml
# config/modules.yaml
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-ABC-123-XYZ" # Opcional
      settings:
        max_items: 10000
```

---

## Ciclo de Vida del Módulo

**Secuencia de Inicio:**

1. **Inicio de Aplicación**

   - Leer `config/modules.yaml`
   - Cargar módulos habilitados
   - Cargar configuraciones por módulo

2. **Inicializar Core**

   - Conexión a base de datos
   - Event bus
   - Gestor JWT
   - Logger

3. **IModule.Initialize()** (en orden de dependencia)

   - Resolver dependencias
   - Inicializar repositorios
   - Inicializar casos de uso

4. **IModule.RegisterRoutes()**

   - Registrar endpoints HTTP

5. **IModule.RegisterMigrations()**

   - Aplicar migraciones de base de datos

6. **IModule.RegisterEventHandlers()**

   - Suscribirse a eventos

7. **IModule.RegisterPermissions()**

   - Insertar permisos RBAC

8. **IModule.Start()**

   - Iniciar workers en segundo plano
   - Inicializar conexiones externas

9. **Servidor Ejecutándose**

**Secuencia de Apagado:**

10. **IModule.Stop()** (en orden inverso)
    - Detener workers
    - Cerrar conexiones
    - Limpiar recursos

---

## Comunicación Inter-Módulos

Los módulos deben comunicarse vía **eventos** (acoplamiento débil):

### Publicando Eventos

```go
// En el módulo Warehouse
func (uc *ItemUseCase) CreateItem(ctx context.Context, item *entity.Item) error {
	// Guardar en base de datos
	if err := uc.repo.Create(ctx, item); err != nil {
		return err
	}

	// Publicar evento
	event := &events.ItemCreatedEvent{
		ItemID:   item.ID,
		SKU:      item.SKU,
		Quantity: item.Quantity,
	}
	uc.eventBus.Publish(ctx, "warehouse.item.created", event)

	return nil
}
```

### Suscribiéndose a Eventos

```go
// En el módulo Fleet (necesita artículos del warehouse para piezas de repuesto)
func (m *FleetModule) RegisterEventHandlers(bus bus.IBus) error {
	return bus.Subscribe(ctx, "warehouse.item.created", m.handleItemCreated)
}

func (m *FleetModule) handleItemCreated(ctx context.Context, event bus.Event) error {
	itemEvent := event.(*events.ItemCreatedEvent)

	// Actualizar inventario de piezas de repuesto
	// ...

	return nil
}
```

---

## Mejores Prácticas

### HACER

1. **Usar BaseModule** - Incorporar `module.BaseModule` para evitar implementar cada método
2. **Clean Architecture** - Seguir estructura domain/usecase/adapter
3. **UUID v7** - Usar `uuidv7.New()` para claves primarias
4. **Orientado a Eventos** - Comunicar entre módulos vía eventos
5. **RBAC** - Registrar permisos para todos los endpoints
6. **Migraciones** - Siempre proporcionar migraciones Up y Down
7. **Health Checks** - Implementar `HealthCheck()` para monitoreo
8. **Apagado Gracioso** - Limpiar recursos en `Stop()`

### NO HACER

1. **Dependencias Directas** - Nunca importar código de otros módulos directamente
2. **Tablas Compartidas** - Cada módulo posee sus tablas
3. **Modificaciones del Core** - No modificar código core para características del módulo
4. **Llamadas Síncronas** - Evitar comunicación inter-módulos bloqueante
5. **Estado Global** - No usar variables globales (usar Core en su lugar)

---

## Ejemplos

### Ejemplo 1: Módulo Simple (Sin Dependencias)

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

### Ejemplo 2: Módulo con Dependencias

```go
package fleet

type FleetModule struct {
	*module.BaseModule
}

func (m *FleetModule) Dependencies() []string {
	return []string{"warehouse"} // Fleet depende de Warehouse
}

func (m *FleetModule) Initialize(ctx context.Context, core *module.Core) error {
	// Warehouse debe ser inicializado primero (garantizado por registry)
	warehouseModule := core.Registry.Get("warehouse")
	if warehouseModule == nil {
		return fmt.Errorf("warehouse module required but not loaded")
	}

	// Inicializar lógica específica de fleet
	// ...

	return nil
}
```

---

## Probando Módulos

```go
// modules/warehouse/module_test.go
func TestWarehouseModule(t *testing.T) {
	// Configurar core de prueba
	core := setupTestCore(t)

	// Crear módulo
	mod := warehouse.New()

	// Inicializar
	err := mod.Initialize(context.Background(), core)
	assert.NoError(t, err)

	// Probar rutas
	router := gin.New()
	group := router.Group("/api/v1")
	mod.RegisterRoutes(group)

	// Probar endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/warehouse/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
```

---

## Próximos Pasos

1. Ver módulos existentes en `internal/modules/`
2. Usar `make generate-module NAME=mymodule` (TODO)
3. Verificar `pkg/module/` para referencia completa de API
4. Leer [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) para patrones de Clean Architecture

---

Para preguntas, vea [README.md](../../README.md) o abra un issue.
