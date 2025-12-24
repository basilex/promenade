# Promenade Architektur - Schnellreferenz

[🇬🇧 English](ARCHITECTURE_QUICKREF.de.md) | [🇺🇦 Українська](../uk/ARCHITECTURE_QUICKREF.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/ARCHITECTURE_QUICKREF.pt.md) | [🇪🇸 Español](../es/ARCHITECTURE_QUICKREF.es.md)

## Core vs Module: Einfache Regel

**CORE** = Infrastruktur + Referenzdaten + Auth

- Immer aktiviert, stellt Dienste bereit

**MODULE** = Geschäftslogik

- Optional, lizenzierbar, unabhängig

---

## Was Gehört in Core?

### GEHÖRT in Core:

1. **Infrastrukturdienste**

   - Datenbank-Verbindungsverwaltung
   - Event Bus (Memory/Redis-Adapter)
   - Scheduler (Cron-Jobs)
   - Konfigurationslader
   - Logger
   - E-Mail-Dienst
   - JWT-Manager

2. **Sicherheitsgrundlage**

   - Benutzerauthentifizierung (Login, Registrierung, Passwort)
   - RBAC (Rollen, Berechtigungen, Zugriffskontrolle)
   - Sitzungen (JWT-Token)

3. **Referenzdaten**

   - Länder (145 Länder, ISO 3166-1-Codes, Regionen)
   - Währungen (124 Währungen, ISO 4217, Symbole)
   - Regionen (30 Verwaltungsregionen: Staaten, Oblaste, Provinzen, Länder)
   - Städte (17 Großstädte mit Koordinaten, Bevölkerung, Hauptstädte)
   - Zahlungsmethoden (40+ Methoden: Karten, Wallets, Krypto, BNPL)
   - Zeitzonen (IANA-Zeitzonendatenbank)
   - Sprachen (ISO 639-Codes)
   - _Stabile, selten ändernde Daten, die von Modulen gemeinsam genutzt werden_

4. **Verwaltungsschnittstellen**
   - Modul-Registry
   - Purge-Handler-Registry
   - Purge-Policy-Registry
   - Event-Bus-Schnittstelle

### GEHÖRT NICHT in Core:

- Geschäftsentitäten (Post, Comment, Profile usw.)
- Geschäfts-Use Cases
- Geschäfts-HTTP-Handler
- Geschäftsrouten
- Geschäftskonfiguration
- Entitätsspezifische Logik

**Faustregel:** Wenn es ein Geschäftskonzept ist, das separat verkauft werden könnte, ist es ein MODUL.

---

## Was Gehört in Module?

### Modul-Struktur

```
internal/modules/mymodule/
├── module.go              # Modul-Implementierung
├── register.go            # Auto-Registrierung via init()
├── config/                # Eigene YAML-Configs pro Umgebung
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── entity/                # Domain-Entitäten
├── usecase/               # Geschäftslogik
└── adapter/
    ├── http/              # Handler, DTOs, Routen
    ├── repository/        # Postgres-Implementierungen
    └── purge/             # Purge-Handler (falls benötigt)
```

### Modul-Checkliste

- [ ] Hat eigene entity/usecase/adapter-Struktur
- [ ] Lädt eigene Config aus `config/config.*.yaml`
- [ ] Registriert Routen in `RegisterRoutes()`
- [ ] Registriert Berechtigungen in `RegisterPermissions()`
- [ ] Registriert Purge-Handler (bei Soft-Delete-Entitäten)
- [ ] Keine Imports aus `internal/domain` oder `internal/usecase`
- [ ] Verwendet nur `pkg/*` Pakete

---

## Modul-Entwicklungs-Workflow

### 1. Modul Erstellen

```bash
mkdir -p internal/modules/mymodule/{config,entity,usecase,adapter/http/handler}
```

### 2. Modul-Interface Implementieren

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type MyModule struct {
    *module.BaseModule
    db *sqlx.DB
    // ... andere Felder
}

func New() module.Module {
    return &MyModule{
        BaseModule: module.NewBaseModule(module.Metadata{
            Name:        "mymodule",
            DisplayName: "My Module",
            Version:     "1.0.0",
            Description: "Does something useful",
        }),
    }
}

func (m *MyModule) Initialize(ctx context.Context, core *module.Core) error {
    // 1. Modul-Config laden
    cfg := moduleconfig.Load("internal/modules/mymodule/config", os.Getenv("ENVIRONMENT"))

    // 2. Repositories, Use Cases, Handler einrichten
    m.db = core.DB

    // 3. Purge-Handler registrieren (falls benötigt)
    // 4. Retention-Policies registrieren (falls benötigt)

    return nil
}

func (m *MyModule) RegisterRoutes(router *gin.RouterGroup) {
    group := router.Group("/mymodule")
    {
        group.GET("", m.handler.List)
        group.POST("", m.handler.Create)
    }
}

func (m *MyModule) RegisterPermissions() []module.Permission {
    return []module.Permission{
        {Resource: "mymodule", Action: "read", Description: "View items"},
        {Resource: "mymodule", Action: "create", Description: "Create items"},
    }
}

// ... andere Interface-Methoden implementieren
```

### 3. Auto-Registrierung

```go
// internal/modules/mymodule/register.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 4. Konfiguration Hinzufügen

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true
  version: "1.0.0"

mymodule:
  max_items: 100
  allow_public: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

### 5. In Haupt-Config Aktivieren

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - mymodule # Hier hinzufügen
```

### 6. In main.go Importieren

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/mymodule"  // Hier hinzufügen
)
```

---

## Konfigurations-Regeln

### Core-Konfiguration

**Datei:** `config/app.{dev|test|prod}.yaml`

**Enthält NUR:**

- Infrastruktur-Einstellungen (DB, Server, JWT, Logging)
- Event-Bus-Konfiguration
- Purge-Infrastruktur (enabled, schedule, batch_size)
- CORS-Einstellungen
- E-Mail-Dienst-Einstellungen

**Enthält NICHT:**

- Entitätsspezifische Aufbewahrungstage → Module
- Modulspezifische Einstellungen → Module
- Geschäftslogik-Konfiguration → Module

### Modul-Konfiguration

**Datei:** `internal/modules/{name}/config/config.{dev|test|prod}.yaml`

**Enthält:**

- Modul-Metadaten (Name, Version)
- Modulspezifische Einstellungen
- Purge-Retention-Policies (falls zutreffend)
- Feature-Flags (falls zutreffend)

**Geladen von:** Jedem Modul via `pkg/module/config.Load()`

---

## Purge-System-Regeln

### ALTER WEG (Core kennt Entitäten)

```go
// ❌ FALSCH - Core hat entitätsspezifische Config
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

### NEUER WEG (Core orchestriert nur)

**Core-Config:**

```yaml
purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**Modul-Config:**

```yaml
purge:
  user_posts:
    retention_days: 90
    enabled: true
```

**Modul-Registrierung:**

```go
// Modul registriert Handler
handler := purge.NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)

// Modul registriert Policy
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

**Core-Orchestrierung:**

```go
// Core holt ALLE Policies aus der Registry
policies := purge.DefaultPolicyRegistry.GetAllPolicies()

// Core erstellt Use Case
useCase := usecase.NewPurgeUseCase(
    purge.DefaultRegistry,  // Handler
    policies,               // von Modulen
    batchSize,
    eventBus,
)

// Core startet Scheduler
scheduler.Start(ctx)
```

**Ergebnis:** Core weiß nichts über `user_posts` oder Aufbewahrungstage!

---

## Kommunikations-Muster

### Event-Driven (Bevorzugt)

```go
// Modul A publiziert
event := &PostCreatedEvent{PostID: id}
eventBus.Publish(ctx, "post.created", event)

// Modul B abonniert
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Asynchron behandeln
})
```

**Vorteile:** Lose Kopplung, asynchrone Verarbeitung

### Direkte Registry (Sparsam Verwenden)

```go
// Anderes Modul holen
postsModule := core.Registry.Get("posts")

// Typ prüfen und aufrufen
if api, ok := postsModule.(PostsAPI); ok {
    stats := api.GetStats(userID)
}
```

**Nur verwenden wenn:** Synchrone Antwort benötigt, Events nicht möglich

---

## Häufige Fehler Vermeiden

### Core-Pakete in Modulen Importieren

```go
// ❌ FALSCH
import "github.com/basilex/promenade/internal/domain/entity"
import "github.com/basilex/promenade/internal/usecase"
```

**Behebung:** Entitäten im eigenen `entity/` Paket des Moduls definieren.

### Geschäftswerte im Code Hart Codieren

```go
// ❌ FALSCH
const maxCommentLength = 2000
```

**Behebung:** Aus Modul-Config laden.

### Geschäftslogik in Core Platzieren

```go
// ❌ FALSCH - PostUseCase in internal/usecase/
```

**Behebung:** In `usecase/` Paket des Moduls verschieben.

### Core Kennt Modul-Entitäten

```go
// ❌ FALSCH - Core hat Aufbewahrungstage für Posts
type PurgeConfig struct {
    RetentionDaysUserPosts int
}
```

**Behebung:** Modul registriert Retention-Policy via Registry.

---

## Test-Strategie

### Core-Tests

- Unit-Tests für Infrastrukturdienste
- Integrationstests für Auth/RBAC
- Tests für Referenzdaten-Repositories

### Modul-Tests

- Unit-Tests für Geschäftslogik (Use Cases)
- Integrationstests für Repositories
- Handler-Tests mit Mock-Use-Cases

### Smoke-Tests

- End-to-End kritische Flows
- Tests für Intermodul-Kommunikation via Events
- Verifizierung, dass Modul-Routen funktionieren

---

## Schnellbefehle

```bash
# Build
make build

# Dev-Modus starten
make dev

# Tests ausführen
make test

# Spezifische Modul-Tests ausführen
go test ./internal/modules/posts/...

# Smoke-Tests ausführen
make test-smoke

# Modul-Boilerplate generieren
make generate ENTITY=MyEntity

# Migration erstellen
make migrate-create NAME=add_my_table
```

---

## Entscheidungsbaum: Core oder Modul?

```
Ist es Infrastruktur (DB, Logger, Event Bus)?
└─> JA → CORE

Ist es Sicherheit (Auth, RBAC)?
└─> JA → CORE

Sind es Referenzdaten (Länder, Währungen, Regionen, Städte, Zahlungsmethoden)?
└─> JA → CORE

Ist es stabil und wird von mehreren Modulen verwendet?
└─> JA → CORE erwägen (oder gemeinsames pkg)

Ist es Geschäftslogik?
└─> JA → MODUL

Kann es separat verkauft werden?
└─> JA → MODUL

Ist es entitätsspezifisch?
└─> JA → MODUL

Im Zweifel?
└─> MODUL (einfacher später zu Core zu verschieben als umgekehrt)
```

---

## Ressourcen

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.de.md) - Detaillierter Architektur-Review
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.de.md) - Visuelle Architektur
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.de.md) - Modul-Entwicklungs-Leitfaden
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.de.md) - Unabhängigkeitsprinzipien
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.de.md) - Purge-System-Details
- [../../internal/CORE.md](../../internal/CORE.md) - Core-Komponenten-Dokumentation

---

**Merken Sie sich:**

- Core = Infrastruktur + Referenzdaten + Auth
- Module = Geschäftslogik (unabhängig, lizenzierbar)
- Registries für lose Kopplung verwenden
- Events für asynchrone Kommunikation
- Konfigurations-Autonomie für jedes Modul
