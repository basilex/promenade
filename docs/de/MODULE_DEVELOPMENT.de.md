# Modul-Entwicklungs-Leitfaden

[🇬🇧 English](../MODULE_DEVELOPMENT.md) | [🇺🇦 Українська](../uk/MODULE_DEVELOPMENT.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/MODULE_DEVELOPMENT.pt.md) | [🇪🇸 Español](../es/MODULE_DEVELOPMENT.es.md)

Dieser Leitfaden erklärt, wie man benutzerdefinierte Module für Promenade unter Verwendung der Plugin-Architektur entwickelt.

## Inhaltsverzeichnis

- [Überblick](#überblick)
- [Modul-Struktur](#modul-struktur)
- [Erstellen eines Moduls](#erstellen-eines-moduls)
- [Modul-Lebenszyklus](#modul-lebenszyklus)
- [Inter-Modul-Kommunikation](#inter-modul-kommunikation)
- [Best Practices](#best-practices)
- [Beispiele](#beispiele)

---

## Überblick

Promenade verwendet eine **Plugin-Architektur**, die es ermöglicht:

- Geschäftslogik als unabhängige, wiederverwendbare Module zu verpacken
- Kern-Infrastruktur zu teilen (DB, Event Bus, Auth, RBAC)
- Module über Konfiguration zu aktivieren/deaktivieren
- Module kommerziell mit Lizenzierung zu verkaufen
- Mehrere Module zu kombinieren (z.B. Warehouse + Fleet)

### Core vs Module

**Core** (immer aktiviert):

- Authentifizierung & JWT
- RBAC & Berechtigungen
- Event Bus & Benachrichtigungen
- Referenzdaten (Länder, Währungen)
- Audit-Logging

**Module** (optional):

- Posts, Comments, Profiles (soziale Funktionen)
- Warehouse Management (Bestandsverwaltung)
- Fleet Management (Fahrzeugverwaltung)
- Finance (Abrechnung, Rechnungen)
- Benutzerdefinierte Geschäftslogik

---

## Modul-Struktur

Ein typisches Modul folgt Clean Architecture:

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
└── module.go  # Modul-Registrierung
```

---

## Erstellen eines Moduls

### Schritt 1: Modul-Metadaten Definieren

```go
// modules/warehouse/module.go
package warehouse

import (
	"github.com/basilex/promenade/pkg/module"
)

type WarehouseModule struct {
	*module.BaseModule

	// Abhängigkeiten
	itemRepo repository.ItemRepository
	itemUC   usecase.ItemUseCase
}

func New() module.Module {
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

### Schritt 2: Modul-Interface Implementieren

```go
// Dependencies (optional - leeres Array zurückgeben wenn keine)
func (m *WarehouseModule) Dependencies() []string {
	return []string{} // Keine Abhängigkeiten
}

// Initialize - Repositories und Use Cases einrichten
func (m *WarehouseModule) Initialize(ctx context.Context, core *module.Core) error {
	// Basis-Implementierung aufrufen
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	// Repositories initialisieren
	m.itemRepo = postgres.NewItemRepository(core.DB)

	// Use Cases initialisieren
	m.itemUC = usecase.NewItemUseCase(m.itemRepo, core.EventBus)

	return nil
}

// RegisterRoutes - HTTP-Endpunkte registrieren
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

// RegisterMigrations - Datenbank-Migrationen zurückgeben
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

// RegisterEventHandlers - Auf Events abonnieren
func (m *WarehouseModule) RegisterEventHandlers(bus bus.Bus) error {
	// Auf Order-Events abonnieren
	return bus.Subscribe(ctx, "order.created", m.handleOrderCreated)
}

func (m *WarehouseModule) handleOrderCreated(ctx context.Context, event bus.Event) error {
	// Lagerbestand reduzieren
	// ...
	return nil
}

// RegisterPermissions - RBAC-Berechtigungen definieren
func (m *WarehouseModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "warehouse:items", Action: "read", Description: "View warehouse items"},
		{Resource: "warehouse:items", Action: "create", Description: "Create warehouse items"},
		{Resource: "warehouse:items", Action: "update", Description: "Update warehouse items"},
		{Resource: "warehouse:items", Action: "delete", Description: "Delete warehouse items"},
	}
}

// Start - Hintergrund-Worker starten (optional)
func (m *WarehouseModule) Start(ctx context.Context) error {
	// Inventar-Sync-Worker starten
	go m.syncInventoryWorker(ctx)
	return nil
}

// Stop - Graceful Shutdown (optional)
func (m *WarehouseModule) Stop(ctx context.Context) error {
	// Worker stoppen, Verbindungen schließen, usw.
	return nil
}

// HealthCheck - Modul-Gesundheit prüfen (optional)
func (m *WarehouseModule) HealthCheck(ctx context.Context) error {
	// Datenbank-Konnektivität, externe APIs prüfen, usw.
	return nil
}
```

### Schritt 3: Modul Registrieren

```go
// modules/warehouse/register.go
package warehouse

import "github.com/basilex/promenade/pkg/module"

func init() {
	// Modul beim Import auto-registrieren
	module.DefaultRegistry.Register(New())
}
```

### Schritt 4: Modul Aktivieren

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

## Modul-Lebenszyklus

**Start-Sequenz:**

1. **Anwendungsstart**

   - `config/modules.yaml` lesen
   - Aktivierte Module laden
   - Pro-Modul-Einstellungen laden

2. **Core Initialisieren**

   - Datenbank-Verbindung
   - Event Bus
   - JWT-Manager
   - Logger

3. **Module.Initialize()** (in Abhängigkeitsreihenfolge)

   - Abhängigkeiten auflösen
   - Repositories initialisieren
   - Use Cases initialisieren

4. **Module.RegisterRoutes()**

   - HTTP-Endpunkte registrieren

5. **Module.RegisterMigrations()**

   - Datenbank-Migrationen anwenden

6. **Module.RegisterEventHandlers()**

   - Events abonnieren

7. **Module.RegisterPermissions()**

   - RBAC-Berechtigungen einfügen

8. **Module.Start()**

   - Hintergrund-Worker starten
   - Externe Verbindungen initialisieren

9. **Server Läuft**

**Shutdown-Sequenz:**

10. **Module.Stop()** (in umgekehrter Reihenfolge)
    - Worker stoppen
    - Verbindungen schließen
    - Ressourcen aufräumen

---

## Inter-Modul-Kommunikation

Module sollten über **Events** kommunizieren (lose Kopplung):

### Events Veröffentlichen

```go
// Im Warehouse-Modul
func (uc *ItemUseCase) CreateItem(ctx context.Context, item *entity.Item) error {
	// In Datenbank speichern
	if err := uc.repo.Create(ctx, item); err != nil {
		return err
	}

	// Event veröffentlichen
	event := &events.ItemCreatedEvent{
		ItemID:   item.ID,
		SKU:      item.SKU,
		Quantity: item.Quantity,
	}
	uc.eventBus.Publish(ctx, "warehouse.item.created", event)

	return nil
}
```

### Events Abonnieren

```go
// Im Fleet-Modul (benötigt Warehouse-Items für Ersatzteile)
func (m *FleetModule) RegisterEventHandlers(bus bus.Bus) error {
	return bus.Subscribe(ctx, "warehouse.item.created", m.handleItemCreated)
}

func (m *FleetModule) handleItemCreated(ctx context.Context, event bus.Event) error {
	itemEvent := event.(*events.ItemCreatedEvent)

	// Ersatzteil-Inventar aktualisieren
	// ...

	return nil
}
```

---

## Best Practices

### TUN

1. **BaseModule Verwenden** - `module.BaseModule` einbetten, um Implementierung jeder Methode zu vermeiden
2. **Clean Architecture** - domain/usecase/adapter-Struktur folgen
3. **UUID v7** - `uuidv7.New()` für Primärschlüssel verwenden
4. **Event-Driven** - Zwischen Modulen über Events kommunizieren
5. **RBAC** - Berechtigungen für alle Endpunkte registrieren
6. **Migrationen** - Immer Up- und Down-Migrationen bereitstellen
7. **Health Checks** - `HealthCheck()` für Monitoring implementieren
8. **Graceful Shutdown** - Ressourcen in `Stop()` aufräumen

### NICHT TUN

1. **Direkte Abhängigkeiten** - Niemals Code anderer Module direkt importieren
2. **Gemeinsame Tabellen** - Jedes Modul besitzt seine Tabellen
3. **Core-Modifikationen** - Core-Code nicht für Modul-Features ändern
4. **Synchrone Aufrufe** - Blockierende Inter-Modul-Kommunikation vermeiden
5. **Globaler State** - Keine globalen Variablen verwenden (stattdessen Core)

---

## Beispiele

### Beispiel 1: Einfaches Modul (Keine Abhängigkeiten)

```go
package hello

type HelloModule struct {
	*module.BaseModule
}

func New() module.Module {
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

### Beispiel 2: Modul mit Abhängigkeiten

```go
package fleet

type FleetModule struct {
	*module.BaseModule
}

func (m *FleetModule) Dependencies() []string {
	return []string{"warehouse"} // Fleet hängt von Warehouse ab
}

func (m *FleetModule) Initialize(ctx context.Context, core *module.Core) error {
	// Warehouse muss zuerst initialisiert sein (garantiert durch Registry)
	warehouseModule := core.Registry.Get("warehouse")
	if warehouseModule == nil {
		return fmt.Errorf("warehouse module required but not loaded")
	}

	// Fleet-spezifische Logik initialisieren
	// ...

	return nil
}
```

---

## Module Testen

```go
// modules/warehouse/module_test.go
func TestWarehouseModule(t *testing.T) {
	// Test-Core einrichten
	core := setupTestCore(t)

	// Modul erstellen
	mod := warehouse.New()

	// Initialisieren
	err := mod.Initialize(context.Background(), core)
	assert.NoError(t, err)

	// Routen testen
	router := gin.New()
	group := router.Group("/api/v1")
	mod.RegisterRoutes(group)

	// Endpunkt testen
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/warehouse/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
```

---

## Nächste Schritte

1. Vorhandene Module in `internal/modules/` ansehen
2. `make generate-module NAME=mymodule` verwenden (TODO)
3. `pkg/module/` für vollständige API-Referenz prüfen
4. [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) für Clean-Architecture-Muster lesen

---

Für Fragen siehe [README.md](../../README.md) oder öffnen Sie ein Issue.
