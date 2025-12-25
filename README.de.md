# Promenade

[🇬🇧 English](README.md) | [🇺🇦 Українська](README.uk.md) | 🇩🇪 **Deutsch**

> **Hinweis zu Übersetzungen**: Einige technische Dokumente aus internen Verzeichnissen (migrations/, internal/, pkg/, test/) sind derzeit nur auf Englisch verfügbar. **Alle Dokumente aus docs/ sind vollständig auf Deutsch übersetzt.** Siehe [docs/de/INDEX.de.md](docs/de/INDEX.de.md) für die vollständige Liste der verfügbaren Übersetzungen.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Produktionsreife REST API** basierend auf **Clean Architecture**, **modularem Plugin-System** und **namespace-basierten Datenbankmigrationen**. Bietet autonome Geschäftsmodule, PostgreSQL mit UUID v7, ereignisgesteuerte Architektur und umfassende Testing-Infrastruktur.

---

## Architekturübersicht

Promenade folgt einer **strikten Schichtenarchitektur**, bei der der **Core orchestriert** und **IModule die Geschäftslogik ausführen**:

**Core-Schicht** - Orchestrator + Infrastruktur + Gemeinsame Dienste

- Authentifizierung & Autorisierung (RBAC)
- Event IBus (Memory/Redis)
- Datenbank & Transaktionen
- Logging & Konfiguration
- Modul-Registry & Lebenszyklus
- Referenzdaten (Länder, Währungen, Regionen, Städte, Zeitzonen, Sprachen, Zahlungsmethoden)

**Modul-Schicht** - Unabhängige Vertikale Slices (Domänen)

| Modul         | Entitäten                          | Beschreibung                        | Status          |
| ------------- | ---------------------------------- | ----------------------------------- | --------------- |
| Posts         | posts, comments, likes             | Benutzergenerierte Inhalte          | Frei            |
| Profiles      | contacts, profiles                 | Benutzerprofile                     | Frei            |
| **Analytics** | **metrics, reports, dashboards**   | **Analytics und Berichterstattung** | **Frei**        |
| **Billing**   | **plans, subscriptions, invoices** | **Abonnement-Abrechnungssystem**    | **Kommerziell** |
| Warehouse     | products, inventory                | Lagerverwaltung (zukünftig)         | Geplant         |

Jedes Modul ist eigenständig mit:

- Eigenen Domain-Entitäten & Geschäftslogik
- Eigenen Datenbankmigrationen (namespace-basiert)
- Eigenen Repositories & Use Cases
- Eigenen HTTP-Handlern & Routen
- Eigener Konfiguration & Lebenszyklus
- Optional: Purge-Policies, Berechtigungen, Events

### Grundprinzipien

1. **Core als Orchestrator**

   - Core **verwaltet** den Modul-Lebenszyklus (init, start, stop)
   - Core **stellt** gemeinsame Dienste bereit (Auth, Events, DB, Logging)
   - Core **weiß WANN** IModule aufgerufen werden, aber nicht **WIE** sie arbeiten
   - Core **importiert niemals** modulspezifischen Code

2. **IModule als Worker**

   - IModule **implementieren** domänenspezifische Geschäftslogik
   - IModule **registrieren** sich selbst über `init()`-Funktionen
   - IModule sind **unabhängig** - können aktiviert/deaktiviert werden ohne andere zu beeinflussen
   - IModule **importieren niemals** Code anderer IModule (nur `pkg/*`)
   - Jedes Modul = vollständige vertikale Slice (Entitäten → Handler)

3. **Clean Architecture Schichten**

   ```
   Domain (Entitäten, Interfaces) → Use Case (Geschäftslogik)
      ↓                                      ↓
   Adapter (Repos, Handler) → Infrastructure (DB, HTTP, Events)
   ```

   - **Dependency Rule**: Innere Schichten hängen niemals von äußeren ab
   - Use Cases hängen nur von Domain-Interfaces ab, nie von konkreten Implementierungen

4. **Namespace-Basierte Migrationen**
   - Jeder Namespace (core, posts, profiles) hat unabhängige Versionsgeschichte
   - Migrationen liegen in `migrations/{namespace}/NNNNNN_beschreibung.{up|down}.sql`
   - Core-Migrationen laufen zuerst, dann aktivierte IModule
   - Echte Modul-Autonomie - aktivieren/deaktivieren ohne Schema-Konflikte

---

## Schnellstart

### Voraussetzungen

- **Go 1.25+**
- **Docker & Docker Compose** (für PostgreSQL, Redis)
- **Make** (für Automatisierung)

### 1. Klonen & Einrichten

```bash
git clone https://github.com/basilex/promenade.git
cd promenade

# Entwicklungsabhängigkeiten installieren
make install

# PostgreSQL + Redis über Docker starten
make docker-up
```

### 2. Migrationen Ausführen

Migrationen laufen **automatisch** beim Anwendungsstart, können aber auch manuell ausgeführt werden:

```bash
# Migrationsstatus für alle Namespaces prüfen
make migrate-status

# Alle Migrationen ausführen (core + aktivierte IModule)
make migrate

# Spezifischen Namespace ausführen
make migrate-core                    # Nur Core
make migrate-module MODULE=posts     # Spezifisches Modul
```

### 3. Anwendung Starten

```bash
# Entwicklungsmodus (Hot Reload, Debug-Logging)
make dev

# Oder Binary bauen und ausführen
make build
./bin/promenade
```

Server startet auf **http://localhost:8081**

---

## Dokumentationsstruktur

### Kerndokumentation

| Dokument                                                                       | Beschreibung                                                             |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| **[docs/de/ARCHITECTURE_OVERVIEW.de.md](docs/de/ARCHITECTURE_OVERVIEW.de.md)** | Visuelle Architekturdiagramme, Schichtverantwortlichkeiten, Lebenszyklus |
| **[docs/de/ARCHITECTURE_QUICKREF.de.md](docs/de/ARCHITECTURE_QUICKREF.de.md)** | Schnellreferenz, Entscheidungsbäume, häufige Fehler                      |
| **[docs/de/ARCHITECTURE_AUDIT.de.md](docs/de/ARCHITECTURE_AUDIT.de.md)**       | Architektur-Compliance-Audit, Verifikationscheckliste                    |

### Modulsystem

| Dokument                                                                                 | Beschreibung                                   |
| ---------------------------------------------------------------------------------------- | ---------------------------------------------- |
| **[docs/de/MODULE_DEVELOPMENT.de.md](docs/de/MODULE_DEVELOPMENT.de.md)**                 | Neue IModule erstellen, Best Practices         |
| **[docs/de/MODULE_INDEPENDENCE.de.md](docs/de/MODULE_INDEPENDENCE.de.md)**               | Modul-Autonomieregeln, Abhängigkeitsverwaltung |
| **[docs/de/MODULE_CONFIG_ARCHITECTURE.de.md](docs/de/MODULE_CONFIG_ARCHITECTURE.de.md)** | Modul-Konfigurationssystem                     |

### Infrastruktur & Systeme

| Dokument                                                                 | Beschreibung                                          |
| ------------------------------------------------------------------------ | ----------------------------------------------------- |
| **[docs/de/PURGE_ARCHITECTURE.de.md](docs/de/PURGE_ARCHITECTURE.de.md)** | Automatisiertes Daten-Purge-System (registry-basiert) |

### Entwicklungsleitfäden

| Dokument                                                                         | Beschreibung                     |
| -------------------------------------------------------------------------------- | -------------------------------- |
| **[docs/de/TESTING_GUIDE.de.md](docs/de/TESTING_GUIDE.de.md)**                   | Testing Best Practices, Patterns |
| **[docs/de/TESTING_INFRASTRUCTURE.de.md](docs/de/TESTING_INFRASTRUCTURE.de.md)** | Test-Infrastruktur-Setup         |

### Technische Referenzen

| Dokument                                                       | Beschreibung                                     |
| -------------------------------------------------------------- | ------------------------------------------------ |
| **[docs/de/UUID_V7_GUIDE.de.md](docs/de/UUID_V7_GUIDE.de.md)** | UUID v7 Implementierung und Vorteile             |
| **[docs/de/SOFT_DELETE.de.md](docs/de/SOFT_DELETE.de.md)**     | Soft-Delete-Pattern für Benutzerinhalte          |
| **[docs/de/AUTHORIZATION.de.md](docs/de/AUTHORIZATION.de.md)** | RBAC-System (4 Rollen, Wildcard-Berechtigungen)  |
| **[docs/de/LOGGING.de.md](docs/de/LOGGING.de.md)**             | Strukturiertes Logging mit slog                  |
| **[docs/de/VALIDATION.de.md](docs/de/VALIDATION.de.md)**       | Request-Validierungsmuster                       |
| **[docs/de/CREDENTIALS.de.md](docs/de/CREDENTIALS.de.md)**     | Standard-Testbenutzer und Anmeldedaten           |
| **[docs/de/INDEX.de.md](docs/de/INDEX.de.md)**                 | Vollständiger Dokumentationsindex mit Lernpfaden |

---

## Modulsystem

### Verfügbare IModule

#### **Posts-Modul** (`internal/modules/posts`)

Verwaltung benutzergenerierter Inhalte:

- **Entitäten**: Posts, Comments, Likes
- **Features**: Posts erstellen/bearbeiten, Kommentar-Threads, Like-System
- **Migrationen**: 3 Migrationen (Namespace: `posts`)
- **Konfiguration**: `config/modules.yaml` → `posts`

#### **Profiles-Modul** (`internal/modules/profiles`)

Benutzerprofil- und Kontaktverwaltung:

- **Entitäten**: UserProfiles, UserContacts
- **Features**: Profilverwaltung, Kontaktinformationen
- **Migrationen**: 2 Migrationen (Namespace: `profiles`)
- **Konfiguration**: `config/modules.yaml` → `profiles`

#### **Analytics-Modul** (`internal/modules/analytics`) - Kostenlos

Analytics, Metriken und Berichterstattung:

- **Status**: Kostenlos - Für alle Benutzer verfügbar
- **Entitäten**: Metrics, Reports, Dashboards
- **Features**: Metrikerfassung, benutzerdefinierte Berichte, visuelle Dashboards
- **Migrationen**: 1 Migration (Namespace: `analytics`)
- **Anwendungsfall**: Business Intelligence, Performance-Monitoring, Dateneinblicke

**Vollständige Dokumentation**: [internal/modules/analytics/README.md](internal/modules/analytics/README.md)

#### **Billing-Modul** (`internal/modules/billing`) - Kommerziell

Abonnement-Abrechnung und Zahlungsabwicklung:

- **Status**: Kommerziell - Produktionsreifes Abonnementverwaltungssystem
- **Entitäten**: Pläne, Abonnements, Rechnungen, Zahlungen
- **Features**:
  - Flexible Abrechnungspläne mit Testphasen
  - Abonnement-Lebenszyklus-Verwaltung (aktiv, pausiert, gekündigt)
  - Automatische Rechnungserstellung
  - Zahlungsverfolgung und -abgleich
  - Mehrere Abrechnungsintervalle (monatlich, vierteljährlich, jährlich)
- **Migrationen**: 4 Migrationen (Namespace: `billing`)
- **Testing**: 375 umfassende Tests (100% Entity + Usecase-Abdeckung)
- **Anwendungsfall**: SaaS-Plattformen, Abonnement-Dienste, wiederkehrende Abrechnung

**Vollständige Dokumentation**: [internal/modules/billing/README.md](internal/modules/billing/README.md)

#### **Notifications-Modul** (`internal/modules/notifications`) - Kommerziell

Mehrkanaliges Benachrichtigungssystem mit Benutzerpräferenzen:

- **Status**: Kommerziell - Produktionsreifes Benachrichtigungs-Liefersystem
- **Entitäten**: Benachrichtigungen, Benutzerpräferenzen
- **Features**:
  - Mehrkanalige Zustellung (Email, SMS, Push, In-App)
  - Benutzerpräferenzverwaltung (pro Kanal, pro Typ)
  - Ruhezeiten mit Zeitzonenunterstützung
  - Benachrichtigungs-Lebenszyklus-Tracking (gesendet, zugestellt, geöffnet, geklickt)
  - Ereignisgesteuerte Architektur
- **Migrationen**: 2 Migrationen (Namespace: `notifications`)
- **Testing**: 35 umfassende Tests (20 Entity + 15 Usecase-Abdeckung)
- **Anwendungsfall**: Transaktions-E-Mails, Marketingkampagnen, System-Alerts, Echtzeit-Benachrichtigungen

**Vollständige Dokumentation**: [internal/modules/notifications/README.md](internal/modules/notifications/README.md)

#### **Warehouse-Modul** (`internal/modules/warehouse`) - Zukünftiges Modul

Lager- und Produktverwaltung (geplant):

- **Status**: Geplant - Struktur existiert als Platzhalter, noch nicht implementiert
- **Anwendungsfall**: E-Commerce, Lagersysteme, Einzelhandel

**Geplante Dokumentation**: [internal/modules/warehouse/README.md](internal/modules/warehouse/README.md)

### Modulstruktur

Jedes Modul folgt einer konsistenten Struktur:

```
internal/modules/{module}/
├── module.go           # Modulregistrierung & Lebenszyklus
├── domain/
│   └── entity/         # Domain-Entitäten
├── repository/         # Datenzugriffs-Interfaces & Implementierungen
├── usecase/            # Geschäftslogik
├── adapter/
│   └── handler/        # HTTP-Handler & DTOs
└── README.md           # Modulspezifische Dokumentation
```

### IModule Aktivieren/Deaktivieren

Bearbeiten Sie `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts # Benutzergenerierte Inhalte
    - profiles # Benutzerprofile + Kontakte
    - analytics # Business Analytics (erfordert Lizenz)
    # - warehouse  # Zukünftig: Lagerverwaltung
```

IModule werden automatisch beim Anwendungsstart geladen.

---

## Datenbankmigrationen

### Namespace-Basiertes System

Jeder Namespace verwaltet **unabhängige Versionsgeschichte**:

```
migrations/
├── core/               # Core-Infrastruktur (läuft immer zuerst)
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000003_core_rbac_full.up.sql
│   ├── 000004_core_ref_timezones.up.sql
│   ├── 000005_core_ref_languages.up.sql
│   ├── 000006_core_ref_countries_currencies.up.sql    # 145 Länder, 124 Währungen
│   ├── 000007_core_ref_regions_cities.up.sql          # 30 Regionen, 17 Städte
│   └── 000008_core_ref_payment_methods.up.sql         # 40+ Zahlungsmethoden
├── posts/              # Posts-Modul-Migrationen
│   ├── 000001_posts_posts.up.sql
│   ├── 000002_posts_comments.up.sql
│   └── 000003_posts_comment_likes.up.sql
├── profiles/           # Profiles-Modul-Migrationen
│   ├── 000001_profiles_contacts.up.sql
│   └── 000002_profiles_profiles.up.sql
└── analytics/          # Analytics-Modul-Migrationen (kommerziell)
    └── 000001_analytics_tables.up.sql
```

### Migrationsbefehle

```bash
# Status für alle Namespaces
make migrate-status

# Alle ausführen (core + aktivierte IModule)
make migrate

# Spezifischen Namespace ausführen
make migrate-core
make migrate-module MODULE=posts

# Rollback
make migrate-rollback MODULE=posts STEPS=1

# Neue Migration erstellen
make migrate-create MODULE=posts NAME=add_post_views
make migrate-create-core NAME=add_audit_log
```

**Auto-Migrationen**: Migrationen laufen automatisch beim App-Start (core zuerst, dann aktivierte IModule).

---

## Authentifizierung & Autorisierung

### Standard-Testbenutzer

| E-Mail                          | Passwort   | Rolle     | Berechtigungen              |
| ------------------------------- | ---------- | --------- | --------------------------- |
| `system@promenade.com`          | `passw0rd` | Admin     | Vollzugriff (`*`)           |
| `admin@promenade.com`           | `passw0rd` | Admin     | Benutzer-/Inhaltsverwaltung |
| `moderator@promenade.com`       | `passw0rd` | Moderator | Inhaltsmoderation           |
| `alexander.vasilenko@gmail.com` | `03041965` | User      | Grundlegende Operationen    |

**Ändern Sie Passwörter vor dem Produktionseinsatz!**

### RBAC-System

- **4 Systemrollen**: Admin, Moderator, User, Guest
- **Wildcard-Berechtigungen**: `posts:*` (alle Post-Aktionen), `*` (Vollzugriff)
- **Ressource-Aktion-Format**: `posts:create`, `users:delete`, `comments:moderate`

**Vollständiger RBAC-Leitfaden**: [docs/de/AUTHORIZATION.de.md](docs/de/AUTHORIZATION.de.md)

---

## Testing

**400+ Tests** über alle Schichten (100% bestanden, ~20 Sekunden):

```bash
# Alle Tests ausführen
make test               # Alle Tests (~20s)

# Nach Modul ausführen
make test-core          # Core-Tests (275 Tests: 39 Entity + 236 Usecase)
make test-modules       # Alle Modul-Tests
make test-module-posts         # Posts-Modul Tests (33 Tests)
make test-module-profiles      # Profiles-Modul Tests (21 Tests)
make test-module-analytics     # Analytics-Modul Tests (11 Tests)
make test-module-notifications # Notifications-Modul Tests (48 Tests: 35 Unit + 13 Integration)
make test-module-billing       # Billing-Modul Tests (163 Tests)

# Coverage-Bericht
make test-coverage      # HTML Coverage-Bericht
```

### Test-Abdeckung

- **Core-Schicht**: 275 Tests
  - Domain-Entities: 39 Tests (Country, Currency, Language, Timezone, Permission, Role, User, Session, Purge)
  - Use Cases: 236 Tests (Auth, RBAC, Referenzdaten-CRUD, Purge-Operationen)
- **Posts-Modul**: 33 Tests, 83.3% Abdeckung (PostStatus, UserPost-Lebenszyklus, Validierung, Slug-Generierung)
- **Profiles-Modul**: 21 Tests, 80.4% Abdeckung (UserContact, UserProfile, Privatsphäre, Validierung)
- **Analytics-Modul**: 11 Tests (Metrics, MetricAggregate, Usecase-Operationen)
- **Notifications-Modul**: 48 Tests (20 Entity + 15 Usecase + 13 Integration) - Mehrkanalige Zustellung, Ruhezeiten, Präferenzen
- **Billing-Modul**: 163 Tests (56 Entity + 107 Usecase) - Plan, Subscription, Invoice, Payment Entitäten
- **Utilities**: 51 Tests, 89.5% durchschnittliche Abdeckung (response 100%, validator 80%, logger 83.8%, pagination 94.1%)

### Test-Ausführungszeit

- **Core-Tests**: 14.7s (Entity 4.4s + Usecase 10.2s)
- **Posts-Modul**: 1.4s
- **Profiles-Modul**: 1.5s
- **Analytics-Modul**: 2.7s (Entity 1.4s + Usecase 1.4s)
- **Notifications-Modul**: 1.8s (Entity 0.4s + Usecase 0.2s + Integration 1.2s)
- **Gesamt**: ~22 Sekunden für die vollständige Test-Suite

**Testing-Leitfäden**:

- [docs/de/TESTING_GUIDE.de.md](docs/de/TESTING_GUIDE.de.md) - Best Practices

---

## Event IBus

**Dual-Adapter Event IBus** für asynchrone Operationen:

### Memory-Adapter

- In-Memory Pub/Sub (Goroutines + Channels)
- **Anwendungsfall**: Entwicklung, Testing, Single-Instance-Deployments
- **Vorteile**: Keine Abhängigkeiten, schnell, einfach
- **Konfiguration**: `BUS_ADAPTER=memory` (Standard)

### Redis-Adapter

- Verteiltes Pub/Sub über Redis
- **Anwendungsfall**: Produktions-Multi-Instance-Deployments
- **Vorteile**: Persistent, skalierbar, fehlertolerant
- **Konfiguration**: `BUS_ADAPTER=redis` + Redis-Verbindungseinstellungen
- **Fallback**: Automatischer Fallback zu Memory, falls Redis nicht verfügbar

### Verwendungsbeispiel

```go
// Event veröffentlichen
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// Events abonnieren
eventBus.Subscribe(ctx, bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    evt := e.(*UserRegisteredEvent)
    // Willkommens-E-Mail senden
    return emailService.SendWelcome(ctx, evt.Email)
})
```

---

## Makefile-Befehle

### Entwicklung

```bash
make dev                # Dev-Server starten (Hot Reload)
make build              # Produktions-Binary bauen
make run                # Gebautes Binary ausführen
make lint               # Linter ausführen (golangci-lint)
make fmt                # Code formatieren
make config-show        # YAML-Konfiguration anzeigen (ENV=dev|test|prod)
```

### Testing

```bash
make test                      # Alle Tests (core + IModule)
make test-core                 # Nur Core-Tests (domain + usecase)
make test-modules              # Alle Modul-Tests
make test-module-posts         # Posts-Modul-Tests
make test-module-profiles      # Profiles-Modul-Tests
make test-coverage             # HTML-Coverage-Bericht generieren
```

### Datenbank

```bash
make migrate                   # Alle Migrationen ausführen (core + aktivierte IModule)
make migrate-status            # Migrationsstatus anzeigen
make migrate-core              # Nur Core migrieren
make migrate-module MODULE=posts          # Spezifisches Modul migrieren
make migrate-rollback MODULE=posts STEPS=1  # Rollback
make migrate-create MODULE=posts NAME=xxx  # Modul-Migration erstellen
make migrate-create-core NAME=xxx          # Core-Migration erstellen
```

### Docker

```bash
make docker-up          # PostgreSQL + Redis starten
make docker-down        # Dienste stoppen
make docker-clean       # Container + Volumes entfernen
make docker-logs        # Logs anzeigen
```

### Swagger

```bash
make swagger-all        # API-Docs generieren (v1 + v2)
make swagger-v1         # Nur v1-Docs generieren
make swagger-v2         # Nur v2-Docs generieren
```

**Vollständiger Makefile-Leitfaden**: [docs/de/MAKEFILE_ARCHITECTURE.de.md](docs/de/MAKEFILE_ARCHITECTURE.de.md)

---

## Docker

### Entwicklungs-Setup

```bash
# Dienste starten
make docker-up

# Logs anzeigen
make docker-logs

# Dienste stoppen
make docker-down

# Neuanfang (Volumes entfernen)
make docker-clean
```

### Dienste

- **PostgreSQL 16**: Port 5432, Benutzer `system`, Datenbank `promenade_dev`
- **Redis 7**: Port 6379 (für verteilten Event IBus)

---

## Wichtige Technische Features

### UUID v7 Primärschlüssel

Zeitgeordnete UUIDs für **2x schnellere Inserts** als UUID v4 und bessere B-Tree-Performance.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    ...
);
```

[docs/de/UUID_V7_GUIDE.de.md](docs/de/UUID_V7_GUIDE.de.md)

### Soft-Delete-Pattern

Benutzergenerierte Inhalte (Posts, Kommentare) verwenden `deleted_at`-Zeitstempel für sichere Löschung.

```go
// Soft-gelöschte Datensätze immer filtern
WHERE deleted_at IS NULL
```

[docs/de/SOFT_DELETE.de.md](docs/de/SOFT_DELETE.de.md)

### Automatisiertes Purge-System

Registry-basiertes System, bei dem IModule ihre Retention-Policies registrieren:

```go
purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
})
```

Core-Scheduler führt Purge-Jobs per Cron aus. Core kennt keine spezifischen Entitäten.

[docs/de/PURGE_ARCHITECTURE.de.md](docs/de/PURGE_ARCHITECTURE.de.md)

### Strukturiertes Logging

Context-aware Logging mit `slog`:

```go
log := logger.FromContext(ctx)  // Enthält request_id, user_id
log.Info("User registered", "email", user.Email)
```

[docs/de/LOGGING.de.md](docs/de/LOGGING.de.md)

---

## API-Dokumentation

### Swagger UI

- **v1 API**: http://localhost:8081/api/v1/docs/swagger/index.html
- **v2 API**: http://localhost:8081/api/v2/docs/swagger/index.html

### Health Check

```bash
curl http://localhost:8081/api/v1/health
```

Antwort:

```json
{
  "status": "success",
  "data": {
    "status": "healthy",
    "database": "connected",
    "timestamp": "2025-12-22T16:40:00Z"
  }
}
```

### API-Versionierung

- **v1**: Aktuelle stabile API (`internal/adapter/http/v1`)
- **v2**: Next-Generation API (`internal/adapter/http/v2`)

Beide Versionen haben:

- Isolierte Handler und DTOs
- Separate Swagger-Dokumentation
- Unabhängige Routen-Registrierung

---

## Projektstruktur

```
promenade/
├── cmd/
│   ├── api/                    # Hauptanwendungs-Einstiegspunkt
│   └── migrate/                # Migrations-CLI-Tool
├── internal/
│   ├── domain/                 # Core-Domain (Entitäten, Interfaces)
│   │   ├── entity/             # Domain-Entitäten (User, Session)
│   │   ├── event/              # Domain-Events (UserRegistered, etc.)
│   │   └── repository/         # Repository-Interfaces
│   ├── usecase/                # Core Use Cases (Auth, RBAC)
│   ├── adapter/                # Adapter (HTTP, Repositories)
│   │   ├── http/
│   │   │   ├── v1/             # API v1 (Handler, DTOs, Routen)
│   │   │   └── v2/             # API v2
│   │   └── repository/postgres/ # PostgreSQL-Implementierungen
│   ├── infrastructure/         # Infrastruktur (DB, Config, Scheduler)
│   │   ├── database/
│   │   ├── config/
│   │   ├── notification/
│   │   └── scheduler/
│   └── modules/                # Geschäftsmodule (Plugins)
│       ├── posts/              # Posts + Kommentare + Likes
│       ├── profiles/           # Benutzerprofile + Kontakte
│       ├── analytics/          # Analytics + Berichte (Kommerziell, aktiv)
│       └── warehouse/          # Lagerverwaltung (zukünftig)
├── pkg/                        # Gemeinsame Pakete (wiederverwendbar)
│   ├── bus/                    # Event IBus (memory/redis)
│   ├── jwt/                    # JWT-Manager
│   ├── logger/                 # Strukturierter Logger
│   ├── migration/              # Migrations-Manager
│   ├── module/                 # Modul-Registry
│   ├── purge/                  # Purge-Registry
│   ├── response/               # HTTP-Response-Helfer
│   ├── uuidv7/                 # UUID v7 Generator
│   └── validator/              # Request-Validierung
├── migrations/                 # Namespace-basierte Migrationen
│   ├── core/                   # Core-Migrationen (Auth, RBAC, Ref-Daten)
│   ├── posts/                  # Posts-Modul-Migrationen
│   └── profiles/               # Profiles-Modul-Migrationen
├── test/                       # Test-Infrastruktur
│   ├── helpers/                # Test-Helfer (Fixtures, DB-Setup)
│   ├── integration/            # Integrationstests
│   ├── smoke/                  # Smoke-Tests
│   └── mocks/                  # Mock-Implementierungen
├── config/                     # Konfigurationsdateien
│   ├── app.dev.yaml            # Dev-Umgebungskonfiguration
│   ├── app.test.yaml           # Test-Umgebungskonfiguration
│   ├── app.prod.yaml           # Produktionskonfiguration
│   └── modules.yaml            # Modul aktivieren/deaktivieren + Einstellungen
├── docs/                       # Dokumentation
├── scripts/                    # Hilfs-Skripte
├── templates/                  # E-Mail-Vorlagen
└── docker/                     # Docker-Konfigurationen
```

---

## Konfiguration

### Umgebungsspezifische Konfigurationen

Konfigurations-Priorität (YAML zuerst, `.env` als Fallback):

1. `config/app.{dev|test|prod}.yaml` - Core-Infrastruktureinstellungen
2. `config/modules.yaml` - Modul aktivieren/deaktivieren + modulspezifische Einstellungen
3. `.env.{env}.local` / `.env.{env}` / `.env` - Legacy-Unterstützung

### Beispiel: `config/app.dev.yaml`

```yaml
app:
  name: "Promenade API"
  environment: "development"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8081

database:
  host: "localhost"
  port: 5432
  user: "system"
  password: "passw0rd"
  database: "promenade_dev"

jwt:
  secret: "your-super-secret-jwt-key-change-in-production"
  access_token_duration: 15m
  refresh_token_duration: 168h

bus:
  adapter: "memory" # oder "redis"
  worker_pool_size: 4

purge:
  enabled: true
  schedule: "0 2 * * *" # Täglich um 2 Uhr
  batch_size: 1000
```

### Beispiel: `config/modules.yaml`

```yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics # Kommerzielles Modul (erfordert Lizenz)
    # - warehouse  # Zukünftig: Lagerverwaltung

  config:
    posts:
      version: "1.0.0"
      settings:
        max_post_length: 10000
        max_comment_depth: 10

    analytics:
      version: "1.0.0"
      license_key: "" # Über ANALYTICS_LICENSE_KEY Umgebungsvariable setzen
      settings:
        metrics_retention_days: 90
```

**Konfigurations-Leitfaden**: [docs/de/MODULE_CONFIG_ARCHITECTURE.de.md](docs/de/MODULE_CONFIG_ARCHITECTURE.de.md)

---

## Neue IModule Erstellen

### Schritt 1: Modulstruktur Erstellen

```bash
mkdir -p internal/modules/mymodule/{domain/entity,repository,usecase,adapter/handler}
```

### Schritt 2: Modul-Interface Implementieren

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type MyModule struct{}

func (m *MyModule) Name() string { return "mymodule" }

func (m *MyModule) Initialize(ctx context.Context, core module.Core) error {
    // Routen, Berechtigungen, Purge-Handler registrieren
    return nil
}

func (m *MyModule) Start(ctx context.Context) error {
    // Background-Worker starten
    return nil
}

func (m *MyModule) Stop(ctx context.Context) error {
    // Graceful Shutdown
    return nil
}

func init() {
    module.DefaultRegistry.Register(&MyModule{})
}
```

### Schritt 3: Migrationen Erstellen

```bash
make migrate-create MODULE=mymodule NAME=create_tables
```

### Schritt 4: Modul Aktivieren

Zu `config/modules.yaml` hinzufügen:

```yaml
modules:
  enabled:
    - mymodule
```

**Vollständiger Leitfaden**: [docs/de/MODULE_DEVELOPMENT.de.md](docs/de/MODULE_DEVELOPMENT.de.md)

---

## Lernpfade

### Für Neue Entwickler

1. **Start**: [docs/de/ARCHITECTURE_QUICKREF.de.md](docs/de/ARCHITECTURE_QUICKREF.de.md) - 15-Minuten-Übersicht
2. **Hands-on**: Ein einfaches Modul erstellen gemäß [docs/de/MODULE_DEVELOPMENT.de.md](docs/de/MODULE_DEVELOPMENT.de.md)

### Für DevOps/Deployment

1. **Makefile**: [docs/de/MAKEFILE_ARCHITECTURE.de.md](docs/de/MAKEFILE_ARCHITECTURE.de.md)
2. **Konfiguration**: [docs/de/MODULE_CONFIG_ARCHITECTURE.de.md](docs/de/MODULE_CONFIG_ARCHITECTURE.de.md)

### Für Architekten

1. **Architekturübersicht**: [docs/de/ARCHITECTURE_OVERVIEW.de.md](docs/de/ARCHITECTURE_OVERVIEW.de.md)
2. **Audit & Verifikation**: [docs/de/ARCHITECTURE_AUDIT.de.md](docs/de/ARCHITECTURE_AUDIT.de.md)
3. **Modul-Unabhängigkeit**: [docs/de/MODULE_INDEPENDENCE.de.md](docs/de/MODULE_INDEPENDENCE.de.md)
4. **Migrationssystem**: [docs/de/MIGRATION_ARCHITECTURE.de.md](docs/de/MIGRATION_ARCHITECTURE.de.md)

**Vollständiger Index**: [docs/de/INDEX.de.md](docs/de/INDEX.de.md)

---

## Mitwirken

1. Repository forken
2. Feature-Branch erstellen (`git checkout -b feature/amazing-feature`)
3. Architekturprinzipien befolgen (siehe [docs/de/ARCHITECTURE_QUICKREF.de.md](docs/de/ARCHITECTURE_QUICKREF.de.md))
4. Tests schreiben (100% Pass-Rate beibehalten)
5. Änderungen committen (`git commit -m 'Add amazing feature'`)
6. Branch pushen (`git push origin feature/amazing-feature`)
7. Pull Request öffnen

---

## Lizenz

Dieses Projekt ist unter der MIT-Lizenz lizenziert - siehe [LICENSE](LICENSE)-Datei für Details.

---

## Support

- **Dokumentation**: [docs/de/INDEX.de.md](docs/de/INDEX.de.md)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **E-Mail**: alexander.vasilenko@gmail.com

---

**Gebaut mit Clean Architecture und Go**
