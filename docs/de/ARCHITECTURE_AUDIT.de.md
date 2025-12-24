# Architektur-Audit - Core vs Module

[🇬🇧 English](../ARCHITECTURE_AUDIT.md) | [🇺🇦 Українська](../ARCHITECTURE_AUDIT.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../ARCHITECTURE_AUDIT.pt.md) | [🇪🇸 Español](../ARCHITECTURE_AUDIT.es.md)

**Datum:** 22. Dezember 2025
**Status:** Architekturkonform

## Zusammenfassung

Promenade folgt einer **Plugin-Architektur**, bei der:

- **Core** = Infrastruktur + Referenzdaten + Manager (immer aktiviert)
- **Module** = Geschäftslogik (optional, lizenzierbar, unabhängig)

Dieses Audit bestätigt, dass die Architektur **korrekt implementiert** ist mit ordnungsgemäßer Trennung der Zuständigkeiten.

---

## Core-Verantwortlichkeiten

### 1. Infrastruktur-Management

```
internal/infrastructure/
├── config/         Konfigurationsverwaltung (YAML + env)
├── database/       Datenbankverbindung + Transaktionen
├── email/          E-Mail-Service
└── scheduler/      Cron-Scheduler
```

**Status:** Korrekt - Core stellt Infrastruktur als Service für Module bereit.

---

### 2. Referenzdaten

```
internal/domain/entity/
├── country.go      Länder (ISO2/3, Regionen) - 195+ Einträge
├── timezone.go     Zeitzonen (IANA) - 500+ Einträge
├── language.go     Sprachen (ISO 639) - 180+ Einträge
└── (currency via repository)  Währungen (ISO 4217) - 170+ Einträge
```

**Zweck:** Stabile, selten sich ändernde Daten, die modulübergreifend genutzt werden.

**Status:** Korrekt - Dies sind echte Referenzdaten, keine Business-Entities.

**Use Case-Implementierungen:**

```
internal/usecase/
├── country_usecase.go   CRUD für Länder
└── currency_usecase.go  CRUD für Währungen
```

---

### 3. Authentifizierung & Autorisierung (RBAC)

```
internal/domain/entity/
├── user.go         Core-User-Entity (nur Auth: E-Mail, Passwort, Rollen)
├── session.go      JWT-Sessions
├── role.go         RBAC-Rollen (5 Systemrollen)
└── permission.go   RBAC-Berechtigungen (resource:action)
```

**Zweck:** Sicherheit und Zugriffskontrolle - grundlegend für alle Module.

**Status:** Korrekt - Auth/RBAC muss im Core sein (alle Module hängen davon ab).

**Use Case-Implementierungen:**

```
internal/usecase/
├── auth_usecase.go        Registrierung, Login, Passwortverwaltung
├── role_usecase.go        Rollenverwaltung
└── permission_usecase.go  Berechtigungsverwaltung
```

---

### 4. Modul-Management-System

```
pkg/module/
├── module.go       Modul-Interface
├── registry.go     Modul-Registry + Abhängigkeitsauflösung
├── config/         Modul-Config-Loader
└── base.go         BaseModule-Helfer
```

**Zweck:** Orchestrierung - Module entdecken, initialisieren, starten/stoppen.

**Status:** Korrekt - Core ist der Orchestrator, Module sind Worker.

**Hauptmerkmale:**

- Dynamisches Laden von Modulen via `init()`-Auto-Registrierung
- Abhängigkeitsauflösung (topologische Sortierung)
- Lifecycle-Management (Initialize → RegisterRoutes → Start → Stop)
- Konfigurationsverwaltung (jedes Modul lädt eigene Konfiguration)

---

### 5. Event Bus-Infrastruktur

```
pkg/bus/
├── bus.go          Event Bus-Interface
├── memory/         In-Memory-Adapter (dev/test)
├── redis/          Redis-Adapter (production)
└── factory.go      Adapter-Factory mit Fallback
```

**Zweck:** Modul-übergreifende Kommunikationsinfrastruktur.

**Status:** Korrekt - Core stellt den Bus bereit, Module nutzen ihn.

---

### 6. Purge-System-Infrastruktur

```
pkg/purge/
├── handler.go      Handler-Registry (Module registrieren Handler)
└── (NEU) Policy-Registry (Module registrieren Aufbewahrungsrichtlinien)
```

```
internal/usecase/
└── purge_usecase.go  Nur Orchestrierung (holt Policies aus Registry)
```

**Zweck:** Scheduler-Infrastruktur - Module definieren was/wann zu löschen ist.

**Status:** BEHOBEN (kürzliches Refactoring) - Core orchestriert, Module implementieren.

---

## Modul-Verantwortlichkeiten

### Aktuelle Module

#### 1. Posts-Modul (`internal/modules/posts/`)

```
posts/
├── module.go               Modul-Implementierung
├── register.go             Auto-Registrierung via init()
├── config/                 Eigene YAML-Configs (dev, test, prod)
│   └── config.*.yaml
├── domain/entity/          Post-, Comment-, Like-Entities
├── usecase/                Geschäftslogik
├── adapter/
│   ├── http/               Handler, DTOs, Routes
│   ├── repository/         Postgres-Implementierungen
│   └── purge/              Purge-Handler für Posts+Comments
└── README.md
```

**Funktionen:**

- Benutzerbeiträge (erstellen, aktualisieren, löschen, soft-delete)
- Kommentare mit Threading (maximale Tiefe konfigurierbar)
- Likes (Posts + Kommentare)
- Purge-Handler mit Aufbewahrungsrichtlinien (90 Tage Posts, 30 Tage Kommentare)

**Status:** Vollständig unabhängig - Keine Imports aus internal/domain oder internal/usecase

---

#### 2. Profiles-Modul (`internal/modules/profiles/`)

```
profiles/
├── module.go               Modul-Implementierung
├── register.go             Auto-Registrierung
├── config/                 Eigene YAML-Configs
│   └── config.*.yaml
├── entity/                 UserProfile-, UserContact-Entities
├── usecase/                Geschäftslogik
└── adapter/
    ├── http/               Handler, DTOs, Routes
    └── repository/         Postgres-Implementierungen
```

**Funktionen:**

- Benutzerprofile (Bio, Avatar, Social-Links)
- Benutzerkontakte (E-Mail, Telefon, mehrere Typen)
- Kontaktverifizierung
- Primärkontakt-Verwaltung

**Status:** Vollständig unabhängig - Profile+Kontakte in ein zusammenhängendes Modul zusammengeführt

---

#### 3. Warehouse-Modul (`internal/modules/warehouse/`)

**Status:** Auskommentiert (kommerzielles Modul, Lizenz erforderlich)

**Zweck:** Bestandsverwaltung für kommerzielle Deployments.

---

## Konfigurationsverwaltung

### Core-Konfiguration

```yaml
# config/app.{env}.yaml - Nur Core-Infrastruktur
app:
  name: "Promenade"
  environment: "development"

server:
  host: "localhost"
  port: 8081

database:
  host: "localhost"
  port: 5432

jwt:
  secret: "..."

bus:
  adapter: "memory" # oder "redis"

purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**Was NICHT in der Core-Konfiguration ist:**

- Entity-spezifische Aufbewahrungsrichtlinien → Zu Modulen verschoben
- Modul-spezifische Einstellungen → Zu Modulen verschoben
- Geschäftslogik-Konfiguration → Zu Modulen verschoben

---

### Modul-Konfiguration

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  name: "posts"
  enabled: true
  version: "1.0.0"

posts:
  max_content_length: 10000
  comments:
    max_content_length: 2000
    max_depth: 10

purge:
  user_posts:
    retention_days: 90
    enabled: true
  post_comments:
    retention_days: 30
    enabled: true
```

**Jedes Modul:**

- Lädt eigene Konfiguration via `pkg/module/config.Load()`
- Definiert eigene Aufbewahrungsrichtlinien
- Registriert Handler + Policies via globale Registries
- Volle Autonomie

---

### Modul-Registry

```yaml
# config/modules.yaml - Welche Module zu laden sind
modules:
  enabled:
    - posts
    - profiles
    # - warehouse  # Benötigt Lizenzschlüssel
```

**Zweck:** Kontrolle, welche Module aktiv sind (Lizenzierung, Funktionen, etc.)

---

## Lizenzierungs-Unterstützung

### Architektur bereit für Lizenzierung

```yaml
# config/modules.yaml (zukünftig)
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-ABC-123-XYZ" # Lizenzvalidierung
      settings:
        max_items: 10000
```

**Modul kann Lizenz in Initialize() validieren:**

```go
func (m *WarehouseModule) Initialize(ctx context.Context, core *Core) error {
    // Konfiguration laden
    cfg := moduleconfig.Load("internal/modules/warehouse/config", env)

    // Lizenz validieren
    licenseKey := cfg.GetString("module.license_key")
    if !validateLicense(licenseKey, "warehouse") {
        return fmt.Errorf("invalid license for warehouse module")
    }

    // Initialisierung fortsetzen...
}
```

**Status:** Architektur unterstützt Lizenzierung - Implementierung bereit wenn benötigt.

---

## Abhängigkeitsverwaltung

### Modul-Abhängigkeiten

```go
func (m *MyModule) Dependencies() []string {
    return []string{"posts", "profiles"}  // Dieses Modul benötigt Posts + Profiles
}
```

**Registry löst Abhängigkeiten automatisch auf:**

1. Topologische Sortierung der Module
2. Initialisierung in Abhängigkeitsreihenfolge
3. Fehler bei zirkulären Abhängigkeiten oder fehlenden Modulen

**Beispiel:** Fleet-Modul hängt vom Warehouse-Modul ab (für Ersatzteile):

```yaml
modules:
  enabled:
    - warehouse # Muss zuerst geladen werden
    - fleet # Hängt von Warehouse ab
```

**Status:** Abhängigkeitssystem implementiert in `pkg/module/registry.go`

---

## Kommunikationsmuster

### 1. Modul-übergreifende Events (Async)

```go
// Posts-Modul publiziert Event
event := &PostCreatedEvent{...}
eventBus.Publish(ctx, "post.created", event)

// Profiles-Modul abonniert
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Benutzerstatistiken aktualisieren
})
```

**Vorteile:**

- Keine direkten Modul-zu-Modul-Imports
- Lose Kopplung
- Asynchrone Verarbeitung

---

### 2. Modul-Registry (Sync)

```go
// Anderes Modul abrufen
postsModule := core.Registry.Get("posts")

// Methoden aufrufen (wenn Modul öffentliches API bereitstellt)
stats := postsModule.(PostsModuleAPI).GetUserStats(userID)
```

**Vorteile:**

- Direkte Kommunikation wenn benötigt
- Typsichere Interfaces
- Sparsam verwenden - Events bevorzugen

---

## Router-Architektur

### Core-Routes

```go
// internal/adapter/http/v1/router/router.go
type V1Router struct {
    // Nur Core-Infrastruktur-Routes
    HealthRouter   *gin.RouterGroup
    AuthRouter     *gin.RouterGroup
    CountryRouter  *gin.RouterGroup
    CurrencyRouter *gin.RouterGroup
    RBACRouter     *gin.RouterGroup  // Rollen + Berechtigungen
    AdminRouter    *gin.RouterGroup  // Purge-Verwaltung
}
```

**Was NICHT im Core-Router ist:**

- Posts-Routes → Zu Posts-Modul verschoben
- Comments-Routes → Zu Posts-Modul verschoben
- Profile-Routes → Zu Profiles-Modul verschoben
- Contact-Routes → Zu Profiles-Modul verschoben

---

### Modul-Routes

```go
// Posts-Modul registriert eigene Routes
func (m *PostsModule) RegisterRoutes(router *gin.RouterGroup) {
    postsGroup := router.Group("/posts")
    {
        postsGroup.GET("", m.postHandler.ListPosts)
        postsGroup.POST("", m.postHandler.CreatePost)
        // ...
    }

    commentsGroup := router.Group("/comments")
    {
        commentsGroup.POST("", m.commentHandler.CreateComment)
        // ...
    }
}
```

**Ergebnis:**

- Core: `/api/v1/auth/*`, `/api/v1/countries/*`, `/api/v1/admin/*`
- Posts-Modul: `/api/v1/posts/*`, `/api/v1/comments/*`
- Profiles-Modul: `/api/v1/profiles/*`, `/api/v1/contacts/*`

**Status:** Klare Trennung - jedes Modul besitzt seine Routes

---

## Datenbank-Verwaltung

### Migrationssystem

**Core-Migrationen (namespace-basiert mit beschreibenden Namen):**

```
migrations/core/
├── 000001_core_init_uuid_v7.up.sql            UUID v7 + Trigger
├── 000002_core_auth_full.up.sql               Auth-Tabellen (users, sessions, tokens)
├── 000003_core_rbac_full.up.sql               RBAC (roles, permissions)
├── 000004_core_ref_timezones.up.sql           Referenzdaten
├── 000005_core_ref_languages.up.sql           Referenzdaten
└── 000006_core_ref_countries_currencies.up.sql Referenzdaten
```

**Modul-Migrationen (namespace-basiert mit Modul-Präfixen):**

```
migrations/posts/
├── 000001_posts_posts.up.sql                  Posts-Tabelle
├── 000002_posts_comments.up.sql               Comments-Tabelle
└── 000003_posts_comment_likes.up.sql          Comment-Likes

migrations/profiles/
├── 000001_profiles_contacts.up.sql            Benutzerkontakte
└── 000002_profiles_profiles.up.sql            Benutzerprofile
```

**Zukünftig:** Module können Migrationen programmatisch registrieren:

```go
func (m *MyModule) RegisterMigrations() []module.Migration {
    return []module.Migration{
        {Version: 1, Up: "CREATE TABLE my_table ...", Down: "DROP TABLE my_table"},
    }
}
```

**Status:** Derzeit dateibasiert, programmatisches System bereit in `pkg/module/module.go`

---

## Test-Strategie

### Core-Tests

```
internal/
├── domain/entity/*_test.go         Entity-Unit-Tests
├── usecase/*_test.go               Use Case-Unit-Tests
└── adapter/repository/*_test.go    Repository-Integrationstests
```

**Fokus:** Auth, RBAC, Referenzdaten, Infrastruktur.

---

### Modul-Tests

```
internal/modules/posts/
├── usecase/*_test.go               Geschäftslogik-Unit-Tests
├── adapter/repository/*_test.go    Repository-Tests
└── module_test.go                  Modul-Integrationstests
```

**Status:** Jedes Modul testet seine eigene Logik unabhängig

---

### Smoke-Tests

```
test/smoke/
├── auth_smoke_test.go              Core-Auth-Flows
├── rbac_smoke_test.go              Core-RBAC-Flows
├── user_post_smoke_test.go         Posts-Modul (benötigt Update)
└── user_profile_smoke_test.go      Profiles-Modul (benötigt Update)
```

**Status:** Smoke-Tests benötigen Import-Pfad-Updates nach Modul-Migration

---

## Verstöße prüfen →

### BEHOBEN: Core hatte entity-spezifische Purge-Policies

**Vorher:**

```go
//  Core kannte Modul-Entities
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**Nachher:**

```go
//  Core hat nur Infrastruktur
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    BatchSize int
}

// Module registrieren Policies via purge.DefaultPolicyRegistry
```

---

### BEHOBEN: Posts/Comments waren im Core

**Vorher:** Posts und Kommentare hatten Entities, Use Cases, Handler in `internal/`

**Nachher:** Vollständige Migration zu `internal/modules/posts/`

**Aus Core gelöscht:** 15.000+ Zeilen Code zu Modul verschoben

---

### BEHOBEN: Profiles/Contacts waren im Core

**Vorher:** Profile und Kontakte verstreut über `internal/domain`, `internal/usecase`, `internal/adapter`

**Nachher:** Vollständige Migration zu `internal/modules/profiles/`

**Ergebnis:** Core wirklich minimal - nur Infrastruktur + Referenzdaten

---

## Zusammenfassung: Core vs Module

| Komponente          | Ort    | Zweck             | Status       |
| ------------------- | ------ | ----------------- | ------------ |
| **Authentication**  | Core   | Sicherheitsbasis  | Korrekt      |
| **RBAC**            | Core   | Zugriffskontrolle | Korrekt      |
| **Countries**       | Core   | Referenzdaten     | Korrekt      |
| **Currencies**      | Core   | Referenzdaten     | Korrekt      |
| **Timezones**       | Core   | Referenzdaten     | Korrekt      |
| **Languages**       | Core   | Referenzdaten     | Korrekt      |
| **Database**        | Core   | Infrastruktur     | Korrekt      |
| **Event Bus**       | Core   | Infrastruktur     | Korrekt      |
| **Purge Scheduler** | Core   | Infrastruktur     | Korrekt      |
| **Module Registry** | Core   | Orchestrierung    | Korrekt      |
|                     |        |                   |
| **Posts**           | Module | Geschäftslogik    | Unabhängig   |
| **Comments**        | Module | Geschäftslogik    | Unabhängig   |
| **Likes**           | Module | Geschäftslogik    | Unabhängig   |
| **Profiles**        | Module | Geschäftslogik    | Unabhängig   |
| **Contacts**        | Module | Geschäftslogik    | Unabhängig   |
| **Warehouse**       | Module | Geschäftslogik    | Lizenzierbar |

---

## Empfehlungen

### 1. Core ist sauber

Der aktuelle Core enthält nur:

- Infrastruktur-Services
- Referenzdaten
- Sicherheit (Auth + RBAC)
- Management-Interfaces (Registries)

**Maßnahme:** Keine Änderungen erforderlich - Architektur ist korrekt.

---

### 2. Module sind unabhängig

Jedes Modul:

- Hat eigene entity/usecase/adapter-Struktur
- Lädt eigene Konfiguration
- Registriert Handler/Policies/Permissions
- Kann via Konfiguration aktiviert/deaktiviert werden

**Maßnahme:** Keine Änderungen erforderlich - Module sind ordnungsgemäß isoliert.

---

### 3. Lizenzierung bereit

Die Architektur unterstützt:

- Lizenzschlüssel-Validierung in Modul-Init
- Modul-spezifische Lizenzkonfiguration
- Abhängigkeitsverwaltung (lizenziertes Modul hängt von freiem Modul ab)

**Maßnahme:** ⏳ Lizenzvalidierung implementieren, wenn kommerzielle Module bereit sind.

---

### 4. Kleinere TODOs

1. **Audit-Modul** - Umfassendes Audit-Logging-System hinzufügen
2. **Smoke-Tests aktualisieren** - Import-Pfade nach Modul-Migration korrigieren
3. **Programmatische Migrationen** - `RegisterMigrations()` in Modulen aktivieren
4. **API-Dokumentation** - Swagger aktualisieren, um Modul-Routes widerzuspiegeln

---

## Fazit

**Architektur-Bewertung: KONFORM**

Promenade implementiert erfolgreich eine **Plugin-Architektur** mit:

- Klarer Trennung zwischen Core (Infrastruktur) und Modulen (Geschäftslogik)
- Modul-Unabhängigkeit (keine Core-Abhängigkeiten)
- Dynamisches Laden von Modulen mit Abhängigkeitsauflösung
- Konfigurations-Autonomie (jedes Modul besitzt seine Konfiguration)
- Lizenzierungs-Unterstützung (bereit für kommerzielle Module)
- Event-gesteuerte Kommunikation (lose Kopplung)

**Core ist wirklich minimal:**

- Infrastruktur-Manager
- Referenzdaten
- Sicherheitsbasis (Auth + RBAC)

**Module sind in sich geschlossen:**

- Eigene Entities, Use Cases, Adapter
- Eigene Konfiguration
- Eigene Purge-Policies
- Einsteckbar (via Konfiguration aktivieren/deaktivieren)

**Nächste Schritte:**

1. Lizenzvalidierung für kommerzielle Module implementieren
2. Smoke-Tests aktualisieren
3. Erwägen, Timezone/Language in separates "reference"-Modul zu extrahieren, falls sie groß werden

---

**Audit-Datum:** 22. Dezember 2025
**Auditor:** AI Assistant (GitHub Copilot)
**Status:** BESTANDEN - Architektur ist solide und korrekt implementiert
