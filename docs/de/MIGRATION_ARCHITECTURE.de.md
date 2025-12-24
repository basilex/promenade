# Modul-Migrationsarchitektur

[🇬🇧 English](../MIGRATION_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/MIGRATION_ARCHITECTURE.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/MIGRATION_ARCHITECTURE.pt.md) | [🇪🇸 Español](../es/MIGRATION_ARCHITECTURE.es.md)

## Problem

Das aktuelle Migrationssystem verletzt die Modulunabhängigkeit:

- Alle Migrationen in einem einzigen `migrations/` Ordner
- Globale sequenzielle Nummerierung (000001, 000002, ...)
- Core-Migrationen vermischt mit Modul-Migrationen
- Keine Möglichkeit, Modul-Migrationen unabhängig zu aktivieren/deaktivieren

## Lösung: Namespace-basierte Migrationen

### 1. Verzeichnisstruktur

```
migrations/
├── core/                          # Core Infrastruktur-Migrationen
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_auth_tables.up.sql
│   ├── 000002_auth_tables.down.sql
│   └── ...
│
├── posts/                         # Posts Modul-Migrationen
│   ├── 000001_create_posts.up.sql
│   ├── 000001_create_posts.down.sql
│   ├── 000002_create_comments.up.sql
│   └── ...
│
└── profiles/                      # Profiles Modul-Migrationen
    ├── 000001_create_profiles.up.sql
    └── ...
```

**Jeder Namespace hat unabhängige Versionierung:**

- Core: 1, 2, 3, 4, ...
- Posts: 1, 2, 3, ...
- Profiles: 1, 2, ...

---

### 2. Migrationstabellen-Schema

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     BIGINT       NOT NULL,
    namespace   VARCHAR(50)  NOT NULL,  -- 'core', 'posts', 'profiles', etc.
    dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, version)
);

CREATE INDEX idx_schema_migrations_namespace ON schema_migrations(namespace);
```

**Beispieldaten:**

```
| version | namespace | dirty | applied_at          |
|---------|-----------|-------|---------------------|
| 1       | core      | false | 2024-12-22 10:00:00 |
| 2       | core      | false | 2024-12-22 10:00:01 |
| 3       | core      | false | 2024-12-22 10:00:02 |
| 1       | posts     | false | 2024-12-22 10:00:03 |
| 2       | posts     | false | 2024-12-22 10:00:04 |
| 1       | profiles  | false | 2024-12-22 10:00:05 |
```

---

### 3. Migration Manager (pkg/migration/)

#### Interface

```go
package migration

type Manager interface {
    // MigrateNamespace wendet alle ausstehenden Migrationen für einen Namespace an
    MigrateNamespace(ctx context.Context, namespace string) error

    // MigrateAll wendet alle ausstehenden Migrationen an (Core + aktivierte Module)
    MigrateAll(ctx context.Context, enabledModules []string) error

    // Rollback rollt N Migrationen für einen Namespace zurück
    Rollback(ctx context.Context, namespace string, steps int) error

    // Version gibt die aktuelle Version für einen Namespace zurück
    Version(ctx context.Context, namespace string) (int, error)

    // Status gibt den Migrationsstatus für alle Namespaces zurück
    Status(ctx context.Context) (map[string]MigrationStatus, error)
}

type MigrationStatus struct {
    Namespace      string
    CurrentVersion int
    PendingCount   int
    Dirty          bool
}
```

#### Implementierung

```go
package migration

type manager struct {
    db            *sqlx.DB
    migrationsDir string  // "migrations/" standardmäßig
}

func NewManager(db *sqlx.DB, migrationsDir string) Manager {
    return &manager{db: db, migrationsDir: migrationsDir}
}

func (m *manager) MigrateNamespace(ctx context.Context, namespace string) error {
    // 1. Prüfen, ob schema_migrations Tabelle existiert
    // 2. Aktuelle Version für Namespace abrufen
    // 3. Migrationsdateien aus migrations/{namespace}/ lesen
    // 4. Ausstehende Migrationen in Transaktion anwenden
    // 5. schema_migrations Tabelle aktualisieren
}

func (m *manager) MigrateAll(ctx context.Context, enabledModules []string) error {
    // 1. Immer zuerst Core migrieren
    if err := m.MigrateNamespace(ctx, "core"); err != nil {
        return err
    }

    // 2. Jedes aktivierte Modul migrieren
    for _, module := range enabledModules {
        if err := m.MigrateNamespace(ctx, module); err != nil {
            return err
        }
    }

    return nil
}
```

---

### 4. Modulintegration

Module können optional Migrationen programmatisch bereitstellen:

```go
// pkg/module/module.go
type Module interface {
    // ... bestehende Methoden ...

    // RegisterMigrations gibt eingebettete Migrationen für dieses Modul zurück
    // Gibt nil zurück, um dateibasierte Migrationen aus migrations/{namespace}/ zu verwenden
    RegisterMigrations() []Migration
}

type Migration struct {
    Version     int
    Description string
    Up          string  // Anzuwendende SQL
    Down        string  // Zurückzurollende SQL
}
```

Beispiel im Modul:

```go
// internal/modules/posts/module.go
func (m *PostsModule) RegisterMigrations() []Migration {
    // Option 1: nil zurückgeben, um dateibasierte Migrationen zu verwenden
    return nil

    // Option 2: Migrationen im Code einbetten (für Libraries)
    return []Migration{
        {
            Version:     1,
            Description: "Create posts table",
            Up:          `CREATE TABLE user_posts (...)`,
            Down:        `DROP TABLE user_posts`,
        },
    }
}
```

---

### 5. Verwendung in main.go

```go
// cmd/api/main.go
func main() {
    // ... Setup ...

    // Migration Manager initialisieren
    migrationMgr := migration.NewManager(db, "migrations")

    // Aktivierte Module aus Config abrufen
    enabledModules := cfg.Modules.Enabled  // ["posts", "profiles"]

    // Migrationen für Core + aktivierte Module ausführen
    if err := migrationMgr.MigrateAll(ctx, enabledModules); err != nil {
        log.Fatal("Failed to run migrations", "error", err)
    }

    // ... Startup fortsetzen ...
}
```

---

### 6. CLI-Befehle

```bash
# Nur Core migrieren
make migrate-core

# Spezifisches Modul migrieren
make migrate-module MODULE=posts

# Alles migrieren (Core + aktivierte Module)
make migrate-all

# Modulmigrationen zurückrollen
make migrate-rollback MODULE=posts STEPS=1

# Migrationsstatus anzeigen
make migrate-status
```

**Makefile:**

```makefile
migrate-core:
	go run cmd/migrate/main.go up --namespace=core

migrate-module:
	go run cmd/migrate/main.go up --namespace=$(MODULE)

migrate-all:
	go run cmd/migrate/main.go up --all

migrate-rollback:
	go run cmd/migrate/main.go down --namespace=$(MODULE) --steps=$(STEPS)

migrate-status:
	go run cmd/migrate/main.go status
```

---

### 7. Migrationsreorganisation

**Aktuell (flach - VERALTET, verwenden Sie stattdessen Namespace-basiert):**

```
migrations/
├── 000001_init_schema_deps.up.sql         # core
├── 000002_create_auth_schema.up.sql       # core
├── 000003_create_countries_currencies.up.sql  # core
├── 000004_create_user_contacts.up.sql     # profiles Modul
├── 000005_create_user_profiles.up.sql     # profiles Modul
├── 000006_create_user_posts.up.sql        # posts Modul
├── 000007_create_post_comments.up.sql     # posts Modul
├── 000008_create_comment_likes_table.up.sql  # posts Modul
├── 000009_create_rbac_tables.up.sql       # core
├── 000014_create_timezones_table.up.sql   # core
└── 000015_create_languages_table.up.sql   # core
```

**Neu (namespace-basiert mit beschreibenden Namen):**

```
migrations/
├── core/
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000001_core_init_uuid_v7.down.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000002_core_auth_full.down.sql
│   ├── 000003_core_rbac_full.up.sql            # war 000009
│   ├── 000003_core_rbac_full.down.sql
│   ├── 000004_core_ref_timezones.up.sql        # war 000014
│   ├── 000004_core_ref_timezones.down.sql
│   ├── 000005_core_ref_languages.up.sql        # war 000015
│   ├── 000005_core_ref_languages.down.sql
│   ├── 000006_core_ref_countries_currencies.up.sql  # war 000003
│   └── 000006_core_ref_countries_currencies.down.sql
│
├── posts/
│   ├── 000001_posts_posts.up.sql               # war 000006_create_user_posts
│   ├── 000001_posts_posts.down.sql
│   ├── 000002_posts_comments.up.sql            # war 000007_create_post_comments
│   ├── 000002_posts_comments.down.sql
│   ├── 000003_create_comment_likes.up.sql    # war 000008
│   └── 000003_create_comment_likes.down.sql
│
└── profiles/
    ├── 000001_create_user_contacts.up.sql    # war 000004
    ├── 000001_create_user_contacts.down.sql
    ├── 000002_create_user_profiles.up.sql    # war 000005
    └── 000002_create_user_profiles.down.sql
```

---

### 8. Vorteile

**Modulunabhängigkeit**

- Jedes Modul besitzt seine Migrationen
- Module aktivieren/deaktivieren ohne Migrationskonflikte
- Klare Eigentumsgrenzen

**Versionskontrolle**

- Jeder Namespace hat unabhängige Versionierung
- Keine globalen Nummerierungskonflikte
- Einfach zu verstehen, auf welcher Version ein Modul ist

**Flexibles Deployment**

- Deployment nur mit benötigten Modulen
- Neue Module hinzufügen ohne bestehende Migrationen zu berühren
- Modulmigrationen unabhängig zurückrollen

**Entwicklererfahrung**

- Klar, wo neue Migrationen platziert werden
- Kein Raten der nächsten globalen Nummer
- Modulspezifische Migrationsbefehle

---

### 9. Migrations-Workflow

#### Erstellen einer neuen Migration

```bash
# Core Migration
make migrate-create-core NAME=add_audit_tables

# Modul Migration
make migrate-create MODULE=posts NAME=add_post_views
```

**Generierte Dateien:**

```
migrations/posts/
├── 000004_add_post_views.up.sql    # Auto-inkrementiert
└── 000004_add_post_views.down.sql
```

#### Anwenden von Migrationen

```bash
# Entwicklung: Alles migrieren
make migrate-all

# Produktion: Core + spezifische Module migrieren
MODULES=posts,profiles make migrate-all

# Selektiv: Nur neues Modul migrieren
make migrate-module MODULE=warehouse
```

---

### 10. Rückwärtskompatibilität

Für bestehende Deployments:

1. **Einmaliges Migrationsskript** das:
   - Aktuelle `schema_migrations` Tabelle sichert
   - Neue `schema_migrations` mit Namespace erstellt
   - Alte Versionen auf Namespace-Versionen abbildet
   - Alle als angewendet markiert

```sql
-- Backup
CREATE TABLE schema_migrations_backup AS SELECT * FROM schema_migrations;

-- Alte Tabelle löschen
DROP TABLE schema_migrations;

-- Neue Tabelle mit Namespace erstellen
CREATE TABLE schema_migrations (...);

-- Gemappte Versionen einfügen
INSERT INTO schema_migrations (namespace, version, dirty, applied_at)
VALUES
    ('core', 1, false, NOW()),  -- war 000001
    ('core', 2, false, NOW()),  -- war 000002
    ('core', 3, false, NOW()),  -- war 000003
    ('profiles', 1, false, NOW()),  -- war 000004
    ('profiles', 2, false, NOW()),  -- war 000005
    ('posts', 1, false, NOW()),  -- war 000006
    ...
```

2. **Migrationsreorganisationsskript** das Dateien in Namespace-Ordner verschiebt

---

### 11. Implementierungsplan

**Phase 1: Infrastruktur (Woche 1)**

1. `pkg/migration/` Package erstellen
2. `Manager` Interface implementieren
3. Namespace-Unterstützung zur schema_migrations Tabelle hinzufügen
4. Tests schreiben

**Phase 2: CLI & Tooling (Woche 1)**

1. `cmd/migrate/main.go` CLI Tool erstellen
2. Makefile-Befehle hinzufügen
3. Dokumentation aktualisieren

**Phase 3: Migration (Woche 2)**

1. Bestehende Migrationen in Namespaces reorganisieren
2. Rückwärtskompatibilitätsmigration erstellen
3. main.go aktualisieren, um neuen Manager zu verwenden
4. Auf Staging testen

**Phase 4: Modulintegration (Woche 2)**

1. `RegisterMigrations()` zum Modul-Interface hinzufügen
2. Bestehende Module aktualisieren
3. Dokumentation & Beispiele

---

### 12. Alternative: golang-migrate Fork

Falls wir die `golang-migrate` Library verwenden wollen:

```go
import "github.com/golang-migrate/migrate/v4"

// Custom Source Driver, der aus Namespace-Ordnern liest
type NamespaceSource struct {
    namespace string
    basePath  string
}

func (s *NamespaceSource) First() (version uint, err error) {
    // Aus migrations/{namespace}/ lesen
}

// Custom Source registrieren
migrate.Register("namespace", &NamespaceSource{})
```

---

## Zusammenfassung

**Beste Lösung:** Custom Migration Manager mit Namespace-Unterstützung.

**Warum?**

- Volle Kontrolle über Namespace-Logik
- Unabhängige Modul-Versionierung
- Einfache Implementierung von Modul aktivieren/deaktivieren
- Klare Eigentumsgrenzen
- Keine Einschränkungen durch externe Libraries

**Migrationspfad:**

1. `pkg/migration/` Manager implementieren
2. Namespace zu schema_migrations hinzufügen
3. Bestehende Migrationen reorganisieren
4. Startup-Code aktualisieren
5. Workflow dokumentieren

**Geschätzter Aufwand:** 2-3 Tage für vollständige Implementierung + Testing
