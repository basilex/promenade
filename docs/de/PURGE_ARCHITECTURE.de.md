# Purge-System-Architektur

🇬🇧 [English](PURGE_ARCHITECTURE.de.md) | [🇺🇦 Українська](../uk/PURGE_ARCHITECTURE.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/PURGE_ARCHITECTURE.pt.md) | [🇪🇸 Español](../es/PURGE_ARCHITECTURE.es.md)

## Überblick

Das Purge-System wurde entwickelt, um soft-gelöschte Datensätze basierend auf Aufbewahrungsrichtlinien automatisch zu bereinigen. Es folgt dem **Orchestrator-Muster**, wobei:

- **Kern** die Infrastruktur verwaltet (Scheduler, Use Case, Event Bus)
- **Module** die Geschäftslogik besitzen (Handler, Aufbewahrungsrichtlinien)

## Architekturprinzipien

### 1. Modulunabhängigkeit

Jedes Modul ist verantwortlich für:

- Registrierung von Purge-Handlern für ihre Entitäten
- Definition von Aufbewahrungsrichtlinien für ihre Entitäten
- Implementierung der tatsächlichen Purge-Logik

Der Kern kennt **niemals** spezifische Entitätstypen oder Aufbewahrungsrichtlinien.

### 2. Registry-Muster

Zwei globale Registries ermöglichen Modulunabhängigkeit:

#### Handler-Registry (`purge.DefaultRegistry`)

```go
// Modul registriert Handler während der Initialisierung
handler := NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)
```

#### Richtlinien-Registry (`purge.DefaultPolicyRegistry`)

```go
// Modul registriert Aufbewahrungsrichtlinie
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

### 3. Kern-Orchestrierung

Die Verantwortlichkeiten des Kerns beschränken sich auf:

1. **Infrastruktur**: Purge-Konfiguration aus YAML laden (enabled, schedule, dry_run, batch_size)
2. **Orchestrierung**: Richtlinien aus Registry sammeln, Use Case erstellen, Scheduler starten
3. **Ausführung**: Purge-Operationen nach Zeitplan auslösen
4. **Ereignisse**: Purge-Ereignisse veröffentlichen (Erfolg/Fehler)

Der Kern macht **NICHT**:

- Spezifische Entitäten kennen
- Aufbewahrungsrichtlinien definieren
- Purge-Logik implementieren

## Konfigurationsstruktur

### Kern-Konfiguration (`config/app.*.yaml`)

```yaml
# Nur Kern-Purge-Infrastruktur
purge:
  enabled: true # Hauptschalter
  schedule: "0 2 * * *" # Cron-Zeitplan (täglich um 2 Uhr)
  dry_run: false # Vorschaumodus
  batch_size: 1000 # Datensätze pro Batch
```

### Modul-Konfiguration (`internal/modules/{name}/config/config.*.yaml`)

```yaml
# Aufbewahrungsrichtlinien des Moduls
purge:
  user_posts:
    retention_days: 90 # 90 Tage aufbewahren
    enabled: true

  post_comments:
    retention_days: 30 # 30 Tage aufbewahren
    enabled: true
```

## Implementierungsablauf

### 1. Modul-Initialisierung

```go
func (m *PostsModule) Initialize(core *Core) error {
    // Modul-Konfiguration laden
    cfg := moduleconfig.Load("internal/modules/posts/config", environment)

    // Aufbewahrungseinstellungen abrufen
    postsRetentionDays := cfg.GetRetentionDays("purge.user_posts.retention_days")
    commentsRetentionDays := cfg.GetRetentionDays("purge.post_comments.retention_days")

    // Purge-Handler registrieren
    postHandler := NewPostPurgeHandler(db)
    purge.DefaultRegistry.Register(postHandler)

    // Aufbewahrungsrichtlinie registrieren
    policy := purge.RetentionPolicy{
        EntityName:    "user_posts",
        RetentionDays: postsRetentionDays,
        Enabled:       true,
    }
    purge.DefaultPolicyRegistry.RegisterPolicy(policy)

    return nil
}
```

### 2. Kern-Initialisierung

```go
func InitPurgeModule(purgeConfig config.PurgeConfig, eventBus bus.Bus) {
    // Alle registrierten Handler abrufen
    handlerRegistry := purge.DefaultRegistry

    // Alle registrierten Richtlinien abrufen
    policyRegistry := purge.DefaultPolicyRegistry
    policies := policyRegistry.GetAllPolicies()

    // In Domain-Entitäten konvertieren
    domainPolicies := convertToEntityPolicies(policies, purgeConfig.Enabled)

    // Use Case erstellen
    useCase := usecase.NewPurgeUseCase(
        handlerRegistry,
        domainPolicies,
        purgeConfig.BatchSize,
        eventBus,
    )

    // Scheduler erstellen und starten
    scheduler := scheduler.NewScheduler(useCase, purgeConfig.Schedule, ...)
    scheduler.Start(ctx)
}
```

### 3. Purge-Ausführung

```go
// Scheduler löst Purge nach Cron-Zeitplan aus
func (s *Scheduler) runPurge(ctx context.Context) {
    // Use Case iteriert über Richtlinien
    for _, policy := range policies {
        // Handler aus Registry abrufen
        handler, ok := registry.Get(policy.EntityName)
        if !ok {
            continue // Überspringen, wenn kein Handler
        }

        // Purge ausführen
        cutoffDate := time.Now().AddDate(0, 0, -policy.RetentionDays)
        recordsPurged, err := handler.Purge(ctx, cutoffDate, batchSize, dryRun)

        // Ereignisse veröffentlichen
        if err != nil {
            eventBus.Publish(ctx, bus.TopicPurgeFailed, ...)
        } else {
            eventBus.Publish(ctx, bus.TopicPurgeCompleted, ...)
        }
    }
}
```

## Erstellen eines neuen Purge-Handlers

### 1. Handler-Interface implementieren

```go
// internal/modules/mymodule/adapter/purge/handler.go
package purge

import (
    "context"
    "time"
    "github.com/jmoiron/sqlx"
)

type MyEntityPurgeHandler struct {
    db *sqlx.DB
}

func NewMyEntityPurgeHandler(db *sqlx.DB) *MyEntityPurgeHandler {
    return &MyEntityPurgeHandler{db: db}
}

func (h *MyEntityPurgeHandler) EntityName() string {
    return "my_entities"
}

func (h *MyEntityPurgeHandler) Purge(
    ctx context.Context,
    cutoffDate time.Time,
    batchSize int,
    dryRun bool,
) (int64, error) {
    query := `
        SELECT id FROM my_entities
        WHERE deleted_at IS NOT NULL
        AND deleted_at < $1
        LIMIT $2
    `

    var ids []string
    if err := h.db.SelectContext(ctx, &ids, query, cutoffDate, batchSize); err != nil {
        return 0, err
    }

    if dryRun {
        return int64(len(ids)), nil // Nur Vorschau
    }

    deleteQuery := `DELETE FROM my_entities WHERE id = ANY($1)`
    result, err := h.db.ExecContext(ctx, deleteQuery, pq.Array(ids))
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}
```

### 2. Im Modul registrieren

```go
// internal/modules/mymodule/module.go
func (m *MyModule) Initialize(core *Core) error {
    // Konfiguration laden
    cfg := moduleconfig.Load("internal/modules/mymodule/config", env)
    retentionDays := cfg.GetRetentionDays("purge.my_entities.retention_days")

    // Handler registrieren
    handler := purge.NewMyEntityPurgeHandler(m.db)
    if err := purge.DefaultRegistry.Register(handler); err != nil {
        return err
    }

    // Richtlinie registrieren
    policy := purge.RetentionPolicy{
        EntityName:    "my_entities",
        RetentionDays: retentionDays,
        Enabled:       true,
    }
    if err := purge.DefaultPolicyRegistry.RegisterPolicy(policy); err != nil {
        return err
    }

    slog.Info("Registered purge for my_entities", "retention_days", retentionDays)
    return nil
}
```

### 3. Modul-Konfiguration hinzufügen

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

## Vorteile

### ✅ Modulunabhängigkeit

- Module besitzen ihre Purge-Logik vollständig
- Keine Kern-Abhängigkeiten von Modul-Entitäten
- Einfaches Hinzufügen/Entfernen von Modulen

### ✅ Konfigurationsklarheit

- Kern: Infrastruktureinstellungen
- Module: Geschäftsrichtlinien
- Klare Trennung der Verantwortlichkeiten

### ✅ Testbarkeit

- Handler können unabhängig getestet werden
- Richtlinien können ohne Code-Änderungen geändert werden
- Registry-Muster ermöglicht einfaches Mocken

### ✅ Wartbarkeit

- Änderungen an Aufbewahrungsrichtlinien erfordern keine Kernänderungen
- Neue Entitäten werden automatisch über Registry erkannt
- Zentralisierte Orchestrierungslogik

## Überwachung

### Veröffentlichte Ereignisse

- `purge.entity.completed` - Entitäts-Purge erfolgreich
- `purge.entity.failed` - Entitäts-Purge fehlgeschlagen
- `purge.all.completed` - Vollständiger Purge-Zyklus abgeschlossen

### Logs

```
level=info msg="Registered purge for user_posts" retention_days=90
level=info msg="Starting purge operation" entity=user_posts cutoff_date=2024-09-22
level=info msg="Purge operation completed" entity=user_posts records_purged=150 duration=2.3s
```

### Admin-Endpunkte

- `GET /api/v1/admin/purge/policies` - Alle Aufbewahrungsrichtlinien auflisten
- `POST /api/v1/admin/purge/preview/:entity` - Purge-Vorschau für Entität
- `POST /api/v1/admin/purge/execute/:entity` - Purge manuell auslösen

## Migration vom alten System

**Vorher** (Kern kannte Entitäten):

```go
// ❌ Kern hatte entitätsspezifische Konfiguration
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**Nachher** (Kern hat nur Infrastruktur):

```go
// ✅ Kern hat nur Infrastruktureinstellungen
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    DryRun    bool
    BatchSize int
}
```

Module registrieren jetzt ihre eigenen Richtlinien über `purge.DefaultPolicyRegistry`.

## Siehe auch

- [Modulunabhängigkeit](MODULE_INDEPENDENCE.de.md)
- [Modulentwicklung](MODULE_DEVELOPMENT.de.md)
- [Testleitfaden](TESTING_GUIDE.de.md)
