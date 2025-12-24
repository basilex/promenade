# Modul-Konfigurationsarchitektur

🇬🇧 [English](../MODULE_CONFIG_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/MODULE_CONFIG_ARCHITECTURE.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/MODULE_CONFIG_ARCHITECTURE.pt.md) | [🇪🇸 Español](../es/MODULE_CONFIG_ARCHITECTURE.es.md)

## Überblick

Module in Promenade sind **vollständig autonome vertikale Schnitte**, die ihre eigene Konfiguration verwalten. Jedes Modul lädt seine Konfiguration aus seinem eigenen Verzeichnisbaum und gewährleistet so vollständige Unabhängigkeit vom Kernsystem.

## Architekturprinzipien

### 1. Modulautonomie

- Jedes Modul ist für das Laden seiner eigenen Konfiguration verantwortlich
- Module speichern Konfigurationen in ihrem eigenen Verzeichnis: `internal/modules/{modulname}/config/`
- Der Kern lädt oder verwaltet keine Modulkonfigurationen
- Der Kern stellt nur Infrastruktur (DB, EventBus, JWT) und Umgebungskontext bereit

### 2. Umgebungsunterstützung

- Jedes Modul hat drei umgebungsspezifische Konfigurationsdateien:
  - `config.dev.yaml` - Entwicklungseinstellungen
  - `config.test.yaml` - Testeinstellungen
  - `config.prod.yaml` - Produktionseinstellungen
- Das Modul-SDK lädt automatisch die richtige Konfiguration basierend auf der `ENVIRONMENT`-Variable

### 3. Konfigurationslade-Ablauf

```
Anwendungsstart
    ↓
Kern lädt app.{env}.yaml
    ↓
Kern initialisiert Infrastruktur (DB, EventBus usw.)
    ↓
Modul-Registry entdeckt Module
    ↓
Für jedes aktivierte Modul:
    Module.Initialize(ctx, core) wird aufgerufen
        ↓
    Modul lädt internal/modules/{name}/config/config.{env}.yaml
        ↓
    Modul validiert Einstellungen (Lizenz, Funktionen usw.)
        ↓
    Modul initialisiert Repositories, Use Cases, Handler
        ↓
    Module.RegisterRoutes() registriert HTTP-Endpunkte
```

## Verzeichnisstruktur

```
config/
├── app.dev.yaml          # Kern-Konfiguration - Entwicklung
├── app.test.yaml         # Kern-Konfiguration - Test
├── app.prod.yaml         # Kern-Konfiguration - Produktion
└── modules.yaml          # Modul-Registry (aktiviert/deaktiviert)

internal/modules/
├── posts/
│   ├── config/
│   │   ├── config.dev.yaml   # Posts-Modul Dev-Konfiguration
│   │   ├── config.test.yaml  # Posts-Modul Test-Konfiguration
│   │   └── config.prod.yaml  # Posts-Modul Prod-Konfiguration
│   ├── entity/
│   ├── repository/
│   ├── usecase/
│   ├── handler/
│   └── module.go            # Lädt eigene Konfiguration in Initialize()
│
├── warehouse/
│   ├── config/
│   │   ├── config.dev.yaml   # Warehouse-Modul Dev-Konfiguration
│   │   ├── config.test.yaml  # Warehouse-Modul Test-Konfiguration
│   │   └── config.prod.yaml  # Warehouse-Modul Prod-Konfiguration
│   └── module.go            # Lädt eigene Konfiguration in Initialize()
│
└── profiles/
    ├── config/
    │   ├── config.dev.yaml   # Profiles-Modul Dev-Konfiguration
    │   ├── config.test.yaml  # Profiles-Modul Test-Konfiguration
    │   └── config.prod.yaml  # Profiles-Modul Prod-Konfiguration
    └── ...
```

## Konfigurationsbereiche

### Kern-Konfiguration (`config/app.{env}.yaml`)

**Verwaltet von**: `internal/infrastructure/config/yaml_config.go`

Enthält:

- Anwendungsmetadaten (Name, Version, Umgebung)
- Servereinstellungen (Host, Port, Timeouts)
- Datenbankverbindung (Host, Port, Anmeldedaten)
- JWT-Einstellungen (Secret, Ablauf)
- Logging-Konfiguration
- CORS-Einstellungen
- Event-Bus-Adapter (memory/redis)
- Ratenbegrenzung
- E-Mail-Dienst

**Enthält niemals**: Modulspezifische Geschäftslogik-Einstellungen

### Modul-Konfiguration (`internal/modules/{name}/config/config.{env}.yaml`)

**Verwaltet von**: Jedem Modul mit `pkg/module/config`

Enthält:

- Modulmetadaten (Name, Version, Aktivierungsflag)
- Modulspezifische Einstellungen
- Bereinigungsrichtlinien (Aufbewahrungstage, Batch-Größen)
- Berechtigungsdefinitionen
- Feature-Flags
- Lizenzschlüssel (für kommerzielle Module)

**Enthält niemals**: Kern-Infrastruktur-Einstellungen

## Implementierungsbeispiel

### Konfigurationsladen des Posts-Moduls

```go
// internal/modules/posts/module.go
package posts

import (
    "context"
    "github.com/basilex/promenade/pkg/module"
    moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

type PostsModule struct {
    *module.BaseModule
    config *moduleconfig.Config
    // ... weitere Felder
}

func (m *PostsModule) Initialize(ctx context.Context, core *module.Core) error {
    // Basis-Implementierung aufrufen
    if err := m.BaseModule.Initialize(ctx, core); err != nil {
        return err
    }

    // Eigene Konfiguration des Moduls aus eigenem Verzeichnis laden
    config, err := moduleconfig.Load("internal/modules/posts/config", core.Config.AppName)
    if err != nil {
        return fmt.Errorf("failed to load posts config: %w", err)
    }
    m.config = config

    // Konfigurationseinstellungen verwenden
    retentionDays := m.config.GetRetentionDays("posts", 30)
    batchSize := m.config.GetBatchSize("posts", 100)

    // ... Repositories, Use Cases, Handler initialisieren

    return nil
}
```

### Struktur der Modul-Konfigurationsdatei

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  enabled: true
  name: "posts"
  version: "1.0.0"

settings:
  max_content_length: 10000
  allow_markdown: true

purge:
  enabled: true
  schedule: "0 2 * * *"
  settings:
    posts:
      retention_days: 30
      batch_size: 100
    comments:
      retention_days: 60
      batch_size: 50

permissions:
  - resource: "posts"
    actions: ["create", "read", "update", "delete"]
  - resource: "comments"
    actions: ["create", "read", "update", "delete"]

features:
  enable_likes: true
  enable_sharing: true
```

## Modul-Konfigurations-SDK

### Konfiguration laden

```go
import moduleconfig "github.com/basilex/promenade/pkg/module/config"

// Konfiguration für aktuelle Umgebung laden
config, err := moduleconfig.Load("internal/modules/{name}/config", environment)
```

### Auf Einstellungen zugreifen

```go
// Aufbewahrungstage für Bereinigung mit Standard-Fallback abrufen
retentionDays := config.GetRetentionDays("entitaetsname", 30)

// Batch-Größe für Bereinigung mit Standard-Fallback abrufen
batchSize := config.GetBatchSize("entitaetsname", 100)

// Prüfen, ob Feature aktiviert ist
enabled := config.GetFeature("enable_likes", true)

// Verschachtelte Einstellung abrufen
value := config.GetNestedSetting("settings", "max_content_length")
```

## Migration von zentralisierten Konfigurationen

### Alte Architektur (Veraltet)

```
config/modules/
├── posts.dev.yaml          Zentralisiert
├── posts.test.yaml         Zentralisiert
├── posts.prod.yaml         Zentralisiert
└── ...

Kern lädt alle Modulkonfigurationen   Enge Kopplung
Kern übergibt Konfigurationen an Module   Abhängigkeit
```

### Neue Architektur (Aktuell)

```
internal/modules/posts/config/
├── config.dev.yaml         Modul-eigentum
├── config.test.yaml        Modul-eigentum
└── config.prod.yaml        Modul-eigentum

Modul lädt eigene Konfiguration   Autonom
Modul verwaltet eigene Einstellungen   Unabhängig
```

## Vorteile

### 1. Echte Modulunabhängigkeit

- Module können unabhängig entwickelt, getestet und bereitgestellt werden
- Keine Kernänderungen erforderlich beim Hinzufügen/Ändern von Modulkonfigurationen
- Module sind wirklich autonome Plugins

### 2. Bessere Kapselung

- Konfiguration befindet sich beim Code, den sie konfiguriert
- Klare Eigentums- und Verantwortungsverhältnisse
- Einfacher zu verstehende Modulfähigkeiten

### 3. Vereinfachter Kern

- Kern verwaltet nur Infrastruktur
- Keine Geschäftslogik in Kernkonfigurationen
- Klarere Trennung der Verantwortlichkeiten

### 4. Einfacheres Testen

- Jedes Modul kann unterschiedliche Testkonfigurationen haben
- Keine globale Konfigurationsverschmutzung
- Modultests sind isoliert

### 5. Flexible Bereitstellung

- Module aktivieren/deaktivieren ohne Kernänderungen
- Unterschiedliche Umgebungen können unterschiedliche Modulkonfigurationen haben
- Kommerzielle Module können Lizenzvalidierung haben

## Kommerzielle Module

Kommerzielle Module (z.B. warehouse) erfordern Lizenzschlüssel:

```yaml
# internal/modules/warehouse/config/config.prod.yaml
module:
  enabled: true
  name: "warehouse"
  version: "1.2.0"
  license_key: "ihr-kommerzieller-lizenzschluessel-hier"

settings:
  max_items: 100000
  enable_barcode_scanner: true
```

Das Modul validiert die Lizenz während `Initialize()`:

```go
func (m *WarehouseModule) verifyLicense() error {
    licenseKey := m.config.GetNestedSetting("module", "license_key").(string)
    if licenseKey == "" {
        return fmt.Errorf("warehouse module requires license key")
    }
    // Lizenz validieren...
    return nil
}
```

## Best Practices

### TUN

- Modulkonfigurationen in `internal/modules/{name}/config/` speichern
- Konfiguration in der `Initialize()`-Methode des Moduls laden
- Umgebungsspezifische Konfigurationsdateien verwenden
- Kritische Einstellungen während der Initialisierung validieren
- Hilfsmethoden von `pkg/module/config` verwenden
- Vernünftige Standardwerte für optionale Einstellungen bereitstellen

### NICHT TUN

- Modulkonfigurationen in `config/modules/` ablegen (veraltet)
- Modulkonfigurationen im Kern laden
- Moduleinstellungen in Kernkonfiguration ablegen
- Umgebungsspezifische Werte fest kodieren
- Konfigurationsvalidierung überspringen
- Auf Konfigurationen anderer Module zugreifen

## Fehlersuche

### Prüfen, welche Konfiguration geladen wurde:

```go
logger.FromContext(ctx).Info("Module config loaded",
    "module", m.GetMetadata().Name,
    "config_path", "internal/modules/posts/config",
    "environment", os.Getenv("ENVIRONMENT"),
)
```

### Umgebung überprüfen:

```bash
echo $ENVIRONMENT  # Sollte sein: development, test oder production
```

### Fehler beim Laden der Konfiguration:

- Prüfen Sie, ob die Datei existiert: `internal/modules/{name}/config/config.{env}.yaml`
- YAML-Syntax überprüfen
- Dateiberechtigungen prüfen
- Sicherstellen, dass die Umgebung korrekt gesetzt ist

## Verwandte Dokumentation

- [Modulentwicklungsleitfaden](MODULE_DEVELOPMENT.de.md)
- [Modulunabhängigkeit](MODULE_INDEPENDENCE.de.md)
- [Test-Infrastruktur](TESTING_INFRASTRUCTURE.de.md)
- [Konfigurationsmigration](CONFIG_MIGRATION.de.md)
