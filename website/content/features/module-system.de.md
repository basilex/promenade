---
title: "Modulares System"
description: "Unabhängige Geschäftsmodule mit Plugin-Architektur"
weight: 2
---

## Modulares Plugin-System

Das **Modulsystem** von Promenade ermöglicht den Aufbau von Anwendungen als Sammlung unabhängiger, wiederverwendbarer Module.

### Was ist ein Modul?

Jedes Modul ist ein **vollständiger vertikaler Schnitt**:

- Eigene Domain-Entitäten
- Eigene Repositories
- Eigene Use Cases
- Eigene HTTP-Handler
- Eigene Datenbankmigrationen
- Eigene Konfiguration

### Verfügbare Module

**Kostenlose Module:**

- **Posts** - Benutzergenerierte Inhalte (Posts, Kommentare, Likes)
- **Profiles** - Benutzerprofile und Kontakte
- **Analytics** - Metriken, Berichte, Dashboards

**Kommerzielle Module:**

- **Billing** - Abonnementverwaltung, Rechnungsstellung, Zahlungen
- **Audit** - Unveränderliche Audit-Logs mit kryptografischen Signaturen
- **Warehouse** - Bestandsverwaltung (in Kürze)

### Modulunabhängigkeit

```go
// Jedes Modul implementiert dieses Interface
type IModule interface {
    Initialize(ctx, core) error
    RegisterRoutes(router)
    RegisterMigrations() []Migration
    RegisterPermissions() []Permission
    Start(ctx) error
    Stop(ctx) error
}
```

### Module Aktivieren/Deaktivieren

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics
    # - warehouse  # Deaktivieren durch Auskommentieren
```

### Vorteile

✅ **Echte Unabhängigkeit** - Module importieren sich nicht gegenseitig  
✅ **Dynamisches Laden** - Aktivieren/Deaktivieren ohne Codeänderungen  
✅ **Eigene Migrationen** - Jedes Modul hat unabhängige Schema-Historie  
✅ **Lizenzierbar** - Kommerzielle Module separat verkaufen

[Modulentwicklungs-Leitfaden →](/promenade/docs/MODULE_DEVELOPMENT)
