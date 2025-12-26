---
title: "Module"
description: "Vollständiger Leitfaden zur modularen Architektur von Promenade und verfügbaren Modulen"
aliases:
  - /de/docs/modules/
---

## Modulsystem

Promenades revolutionäre modulare Architektur, bei der Business-Logik in unabhängigen, lizenzierbaren Modulen lebt.

---

## Kernkonzept

**Core = Orchestrator** (Infrastruktur, keine Business-Logik)

- Authentifizierung & Autorisierung (JWT, RBAC)
- Event Bus (Memory/Redis-Adapter)
- Datenbankverwaltung (Connection Pool, Transaktionen)
- Referenzdaten (Länder, Währungen, Regionen, Städte, Zahlungsmethoden)
- Logging, Konfiguration, Scheduler

**Module = Worker** (Business-Domänen)

- Vollständige vertikale Slices (entity → usecase → adapter)
- Unabhängiger Lebenszyklus (Aktivierung/Deaktivierung in Konfiguration)
- Eigene Migrationen (namespace-basiert)
- Eigene Berechtigungen (RBAC-Integration)
- Event-gesteuerte Kommunikation

---

## Verfügbare Module

### Kostenlose Module

<div class="docs-grid">

<div class="feature-card">

#### Posts-Modul

Verwaltung benutzergenerierter Inhalte.

**Funktionen:**

- Erstellen/Bearbeiten von Posts mit Markdown
- Verschachtelte Kommentare (Threading)
- Like-System
- Soft Delete mit Löschrichtlinien
- Slug-Generierung

**Entitäten:** `UserPost`, `Comment`, `CommentLike`

[Quellcode anzeigen](https://github.com/basilex/promenade/tree/dev/internal/modules/posts)

</div>

<div class="feature-card">

#### Profiles-Modul

Benutzerprofil- und Kontaktverwaltung.

**Funktionen:**

- Benutzerprofile mit Bio, Avatar, Standort
- Mehrere Kontaktmethoden (E-Mail, Telefon, Social Media)
- Datenschutzeinstellungen (öffentlich/privat/Freunde)
- Kontaktverifizierung

**Entitäten:** `UserProfile`, `UserContact`

[Quellcode anzeigen](https://github.com/basilex/promenade/tree/dev/internal/modules/profiles)

</div>

<div class="feature-card">

#### Analytics-Modul

Metrikerfassung und Reporting.

**Funktionen:**

- Benutzerdefinierte Metrikaufzeichnung
- Zeitreihen-Aggregation
- Dashboard-Daten
- 90-Tage-Aufbewahrung

**Entitäten:** `Metric`, `MetricAggregate`

[Quellcode anzeigen](https://github.com/basilex/promenade/tree/dev/internal/modules/analytics)

</div>

</div>

---

### Kommerzielle Module

<div class="docs-grid">

<div class="feature-card">

#### Notifications-Modul 🔔

Mehrkanaliges Benachrichtigungssystem.

**Funktionen:**

- Mehrkanalige Zustellung (E-Mail, SMS, Push, In-App)
- Benutzerpräferenzen (pro Kanal, pro Typ)
- Ruhezonen mit Zeitzonenunterstützung
- Statusverfolgung (gesendet → zugestellt → geöffnet → geklickt)
- Systembenachrichtigungen umgehen Ruhezonen
- Flexible JSONB-Datenspeicherung

**Entitäten:** `Notification`, `UserPreference`

**Abdeckung:** 48 Tests (20 Entity + 15 Usecase + 13 Integration)

[Quellcode anzeigen](https://github.com/basilex/promenade/tree/dev/internal/modules/notifications)

</div>

<div class="feature-card">

#### Billing-Modul 💰

Produktionsreife Abonnementverwaltung.

**Funktionen:**

- Flexible Abrechnungspläne (monatlich/vierteljährlich/jährlich)
- Testphasen
- Abonnement-Lebenszyklus (aktiv/pausiert/gekündigt)
- Rechnungserstellung
- Zahlungsverfolgung
- Mehrere Zahlungsmethoden

**Entitäten:** `Plan`, `Subscription`, `Invoice`, `Payment`

**Abdeckung:** 375 Tests, 100% Entity + Usecase Coverage

[Quellcode anzeigen](https://github.com/basilex/promenade/tree/dev/internal/modules/billing)

</div>

<div class="feature-card">

#### Audit-Modul 🔒

Unveränderliche Audit-Logs für Compliance.

**Funktionen:**

- Unveränderlicher Audit-Trail
- HMAC-SHA256-Signaturen
- SOC 2 / GDPR-Konformität
- Manipulationserkennung
- Langzeitaufbewahrung

**Anwendungsfälle:** Finanzanwendungen, Gesundheitswesen, Regulierte Branchen

</div>

<div class="feature-card">

#### Warehouse-Modul 📦

Inventar- und Produktverwaltung (geplant).

**Funktionen:**

- Produktkatalog
- Bestandsverfolgung
- Bestandsanpassungen
- Multi-Standort-Unterstützung

**Status:** Geplant für Q1 2026

</div>

</div>

---

## Modul-Entwicklung

### Eigenes Modul Erstellen

Folgen Sie unserem [Modul-Entwicklungsleitfaden](/promenade/de/docs/module-development) zum Erstellen benutzerdefinierter Module.

**Schnellstruktur:**

```
internal/modules/mymodule/
├── module.go              # IModule-Interface-Implementierung
├── register.go            # Auto-Registrierung via init()
├── config/                # Umgebungsspezifische Konfigurationen
├── entity/                # Domain-Entitäten
├── usecase/               # Business-Logik
└── adapter/
    ├── http/              # HTTP-Handler & DTOs
    └── repository/        # PostgreSQL-Implementierungen
```

### Modul-Funktionen

✅ **Auto-Registrierung** - `init()`-Funktion registriert Modul  
✅ **Eigene Konfiguration** - YAML-Konfigurationen pro Umgebung  
✅ **Eigene Migrationen** - Namespace-basiert, unabhängige Historie  
✅ **Eigene Berechtigungen** - RBAC-Integration  
✅ **Event-Kommunikation** - Publish/Subscribe zu Domain-Events  
✅ **Hintergrund-Worker** - Cron-Jobs, Queue-Prozessoren  
✅ **Health-Checks** - Graceful Startup/Shutdown

---

## Modul-Unabhängigkeit

**Module importieren NIEMALS:**

- `internal/domain` (Core-Domain)
- `internal/usecase` (Core-Use-Cases)
- `internal/adapter` (Core-Adapter)
- Andere Module

**Module KÖNNEN verwenden:**

- `pkg/*` (gemeinsame Pakete)
- Core-Infrastruktur via `module.Core`
- Events für Modul-zu-Modul-Kommunikation

Dies erzwingt echte architektonische Autonomie.

---

## Module Aktivieren

### Konfiguration

Bearbeiten Sie `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics
    - notifications # Lizenz erforderlich
    - billing # Lizenz erforderlich
    - audit # Lizenz erforderlich
    # - warehouse  # Noch nicht verfügbar
```

### Import in main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/analytics"
    _ "github.com/basilex/promenade/internal/modules/notifications"
    _ "github.com/basilex/promenade/internal/modules/billing"
)
```

### Migrationen Ausführen

```bash
make migrate                     # Alle aktivierten Module
make migrate-module MODULE=posts # Spezifisches Modul
```

---

## Kommerzielle Lizenzierung

Kommerzielle Module (billing, audit, warehouse) benötigen Lizenzschlüssel.

**Lizenz Generieren:**

```bash
./scripts/generate-license.sh billing PRO 365
```

**Umgebungsvariable Setzen:**

```bash
export BILLING_LICENSE_KEY="PROMENADE-BILLING-PRO-20261225-xxxxx"
```

**Oder in YAML Konfigurieren:**

```yaml
# config/modules.yaml
modules:
  config:
    billing:
      license_key: "PROMENADE-BILLING-PRO-20261225-xxxxx"
```

Kontakt: alexander.vasilenko@gmail.com für Enterprise-Lizenzierung.

---

## Mehr Erfahren

- [Modul-Entwicklungsleitfaden](/promenade/de/docs/module-development) - Vollständiges Tutorial
- [Architekturübersicht](/promenade/de/docs/architecture) - Systemdesign
- [Datenbankschema](/promenade/de/docs/database-schema) - Modul-Tabellen
- [Beispiel-Module](https://github.com/basilex/promenade/tree/dev/internal/modules) - Quellcode

---

## Support

🐛 [Probleme Melden](https://github.com/basilex/promenade/issues)  
💬 [Diskussionen](https://github.com/basilex/promenade/discussions)  
📧 alexander.vasilenko@gmail.com
