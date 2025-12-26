# Promenade Architektur-Überblick

[ English](../ARCHITECTURE_OVERVIEW.md) | [ Українська](../uk/ARCHITECTURE_OVERVIEW.uk.md) |  **Deutsch** | [ Português](../pt/ARCHITECTURE_OVERVIEW.pt.md) | [ Español](../es/ARCHITECTURE_OVERVIEW.es.md)

Dieses Dokument bietet einen Überblick über die Promenade-Architektur, organisiert um **Clean Architecture-Prinzipien** und ein **Plugin-basiertes Modulsystem**.

---

## Architekturebenen

### 1. Anwendungsebene

**Ort:** `cmd/api/main.go`

**Verantwortlichkeiten:**

- Infrastruktur initialisieren (DB, EventBus, Config, Logger)
- Core-Konfiguration laden
- Modulsystem initialisieren (Erkennung, Registrierung, Lebenszyklus)
- HTTP-Server starten

---

### 2. Core-Infrastruktur

**Ort:** `internal/infrastructure/`

**Komponenten:**

| Komponente    | Zweck                            |
| ------------- | -------------------------------- |
| Database      | PostgreSQL-Verbindungsverwaltung |
| Event IBus     | Memory/Redis Pub/Sub-Adapter     |
| Scheduler     | Cron-basierte Aufgabenplanung    |
| Config        | YAML-Konfigurationslader         |
| Logger        | Strukturiertes Logging mit slog  |
| Email Service | Asynchrones E-Mail-Versenden     |

**Bietet:** Gemeinsame Dienste für alle IModule (DB, EventBus, Config, JWT, Logger)

---

### 3. Core-Domain

**Ort:** `internal/domain/entity/`

**Sicherheit & Authentifizierung** (immer aktiviert):

- User (nur Authentifizierung: E-Mail, Passwort, Rollen)
- Session (JWT-Tokens)
- Role (RBAC-Rollen)
- Permission (Format resource:action)

**Referenzdaten** (stabil, gemeinsam):

- Country (195+ ISO-Codes, Regionen)
- Currency (170+ ISO 4217-Codes)
- Timezone (500+ IANA-Zeitzonen)
- Language (180+ ISO 639-Codes)

**Zweck:** Core-Entities, die von ALLEN Modulen benötigt werden. Keine Geschäftslogik.

---

### 4. Core Use Cases

**Ort:** `internal/usecase/`

**Verfügbare Use Cases:**

| Use Case           | Zweck                                  |
| ------------------ | -------------------------------------- |
| Auth UseCase       | Registrierung, Login, Logout           |
| Role UseCase       | CRUD-Rollen, Rollenzuweisung           |
| Permission UseCase | CRUD-Berechtigungen, Zugriffskontrolle |
| Country UseCase    | Länder auflisten, nach Code abrufen    |
| Currency UseCase   | Währungen auflisten, nach Code abrufen |
| Purge UseCase      | Purge-Jobs orchestrieren               |

**Hinweis:** Core Use Cases enthalten KEINE Geschäftslogik - nur Infrastruktur und Sicherheit.

---

### 5. Core-API-Routen

**Ort:** `internal/adapter/http/v1/`

**Endpunkte:**

| Route                  | Zweck                        |
| ---------------------- | ---------------------------- |
| /api/v1/health         | Gesundheitschecks            |
| /api/v1/auth/\*        | Login, Registrierung, Logout |
| /api/v1/countries/\*   | Referenzdaten                |
| /api/v1/currencies/\*  | Referenzdaten                |
| /api/v1/roles/\*       | RBAC-Verwaltung              |
| /api/v1/permissions/\* | RBAC-Verwaltung              |
| /api/v1/admin/\*       | Purge, Systemverwaltung      |

---

## Modulsystem

### Modul-Registry & Manager

**Ort:** `pkg/module/`

**Zweck:** Orchestriert Modul-Lebenszyklus

**Registry-Funktionen:**

- `Register(module)` - Auto-Registrierung via `init()`
- `GetEnabled(config)` - Aktivierte IModule aus Config filtern
- `InitializeAll()` - Initialisierung in Abhängigkeitsreihenfolge
- `StartAll()` - Hintergrund-Worker starten
- `StopAll()` - Graceful Shutdown

**Bietet Modulen:**

- Gemeinsame DB-Verbindung
- Gemeinsamer EventBus
- Gemeinsamer JWT-Manager
- Gemeinsamer Config-Loader

---

### Modul: Posts

**Ort:** `internal/modules/posts/`

**Status:** Aktiviert (Kostenlos)

**Entities:** Post, Comment, Like

**Funktionen:**

- Posts erstellen, aktualisieren, löschen
- Thread-Kommentare (konfigurierbare max. Tiefe)
- Posts und Kommentare liken
- Soft Delete mit konfigurierbarer Aufbewahrung

**Konfiguration:**

- max_content_length: 10000
- comments.max_depth: 10
- purge.user_posts.retention_days: 90
- purge.post_comments.retention_days: 30

**Routen:** /api/v1/posts/_, /api/v1/comments/_, /api/v1/likes/\*

---

### Modul: Profiles

**Ort:** `internal/modules/profiles/`

**Status:** Aktiviert (Kostenlos)

**Entities:** Profile, Contact

**Funktionen:**

- Benutzerprofilverwaltung
- Kontaktinformationen (E-Mail, Telefon usw.)
- Kontaktverifizierung
- Primärkontakt-Bezeichnung

**Konfiguration:**

- profiles.max_per_user: 1
- contacts.max_per_user: 5
- contacts.verification_required: true

**Routen:** /api/v1/profiles/_, /api/v1/contacts/_

---

### Modul: Warehouse

**Ort:** `internal/modules/warehouse/`

**Status:** Deaktiviert (Kommerziell - erfordert Lizenz)

**Funktionen** (bei Lizenzierung):

- Bestandsverwaltung
- Lagerbestandsverfolgung
- Barcode-Scannen
- Lagerorte

**Routen:** /api/v1/warehouse/\* (wenn aktiviert)

---

## Inter-Modul-Kommunikation

### Event IBus

**Ort:** `pkg/bus/`

**Adapter:**

| Adapter | Anwendungsfall               | Funktionen                    |
| ------- | ---------------------------- | ----------------------------- |
| Memory  | Dev/Test/Einzelinstanz       | Schnell, keine Abhängigkeiten |
| Redis   | Production/Mehrere Instanzen | Verteilt, persistent          |

**Event-Fluss:**

1. Modul A veröffentlicht Event zum EventBus
2. EventBus verteilt an alle Abonnenten
3. IModule B, C, D verarbeiten Event asynchron

**Beispiele:**

- user.registered → Willkommens-E-Mail senden (asynchron)
- post.created → Benutzerstatistiken aktualisieren (asynchron)
- purge.completed → In Audit protokollieren (asynchron)

---

## Konfigurationsarchitektur

### Core-Konfiguration

```
config/
 app.dev.yaml     - Core-Infrastruktur (dev)
 app.test.yaml    - Core-Infrastruktur (test)
 app.prod.yaml    - Core-Infrastruktur (prod)
 modules.yaml     - Welche IModule zu laden sind
```

### Modul-Konfiguration

```
internal/modules/posts/config/
 config.dev.yaml  - Posts-Moduleinstellungen (dev)
 config.test.yaml - Posts-Moduleinstellungen (test)
 config.prod.yaml - Posts-Moduleinstellungen (prod)

internal/modules/profiles/config/
 config.dev.yaml  - Profiles-Moduleinstellungen (dev)
 config.test.yaml - Profiles-Moduleinstellungen (test)
 config.prod.yaml - Profiles-Moduleinstellungen (prod)
```

**Umgebungsüberschreibungen:** `.env.example` (optional)

---

## Modul-Lebenszyklus

### 1. Auto-Registrierung (via init())

Modul registriert sich beim Paket-Import:

```go
package posts

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 2. Erkennung & Filterung

- `config/modules.yaml` lesen
- Aktivierte IModule filtern
- Abhängigkeiten auflösen (topologische Sortierung)

### 3. Initialisierung (in Abhängigkeitsreihenfolge)

Für jedes Modul:

- Modulkonfiguration aus `config/config.*.yaml` laden
- Repositories einrichten
- Use Cases einrichten
- Handler einrichten
- Purge-Handler registrieren
- Retention-Policies registrieren
- Berechtigungen registrieren

### 4. Routen-Registrierung

Für jedes Modul:

- Modulrouten zum Router mounten
- Middleware anwenden (Auth, RBAC usw.)

### 5. Event-Abonnement

Für jedes Modul:

- Relevante Events abonnieren
- Asynchrone Event-Handler einrichten

### 6. Start (Hintergrund-Worker)

Für jedes Modul:

- Cron-Jobs starten
- Hintergrund-Worker starten

### 7. Laufzeit

- IModule verarbeiten HTTP-Requests
- Veröffentlichen/Abonnieren von Events
- Geplante Aufgaben ausführen

### 8. Herunterfahren (bei SIGTERM/SIGINT)

Für jedes Modul (umgekehrte Reihenfolge):

- Worker graceful stoppen
- Verbindungen schließen
- Ressourcen aufräumen

---

## Wichtige Design-Prinzipien

### 1. CORE = INFRASTRUKTUR + REFERENZDATEN

- Core bietet Dienste (DB, EventBus, Config, JWT)
- Core enthält stabile Referenzdaten (Länder, Währungen)
- Core verwaltet Sicherheit (Auth, RBAC)
- Core enthält KEINE Geschäftslogik

### 2. MODULE = GESCHÄFTSLOGIK

- IModule sind selbstständige vertikale Slices
- IModule besitzen ihre Entities, Use Cases, Adapter
- IModule registrieren Handler, Policies, Berechtigungen
- IModule können via Config aktiviert/deaktiviert werden
- IModule importieren NICHT aus `internal/domain` oder `internal/usecase`
- IModule hängen NICHT direkt voneinander ab (verwenden Events)

### 3. PLUGIN-ARCHITEKTUR

- Dynamisches Laden via Modul-Registry
- Abhängigkeitsauflösung (topologische Sortierung)
- Lebenszyklusverwaltung (init → start → stop)
- Auto-Registrierung via `init()`

### 4. KONFIGURATIONSAUTONOMIE

- Core: `config/app.*.yaml` (nur Infrastruktur)
- IModule: `internal/modules/{name}/config/config.*.yaml`
- Jedes Modul lädt seine eigene Config
- Umgebungsspezifische Configs (dev, test, prod)

### 5. LOSE KOPPLUNG

- Event-basierte Kommunikation (Pub/Sub)
- Registry-Pattern (Handler, Policies, Berechtigungen)
- Interface-basierte Abhängigkeiten
- Keine direkten Modul-zu-Modul-Imports

### 6. LIZENZIERUNGSUNTERSTÜTZUNG

- Lizenzschlüssel pro Modul
- Lizenzvalidierung in `Initialize()`
- Graceful Degradation bei ungültiger Lizenz

---

## Vorteile

### Modularität

- Neue IModule hinzufügen ohne Core zu berühren
- IModule entfernen ohne andere zu brechen
- IModule unabhängig testen

### Skalierbarkeit

- Kommerzielle IModule (warehouse, fleet, finance)
- Lizenzbasierte Feature-Aktivierung
- Einfaches Hinzufügen neuer Vertikalen

### Wartbarkeit

- Klare Grenzen (Core vs IModule)
- Single Responsibility (jedes Modul besitzt seine Domain)
- Konfigurationsklarheit (keine monolithische Config)

### Testbarkeit

- Unit-Test von Modulen isoliert
- Integrationstest mit echtem/Mock-Core
- Smoke-Test kritischer Flows

### Deployability

- Nur benötigte IModule pro Deployment aktivieren
- A/B-Testing neuer IModule
- Schrittweiser Feature-Rollout

---

## Aktueller Status

### Core

- Infrastruktur (DB, EventBus, Scheduler, Config, Logger)
- Sicherheit (Auth, RBAC, JWT, Sessions)
- Referenzdaten (Countries, Currencies, Timezones, Languages)
- Modulverwaltung (Registry, Lifecycle, Config Loader)
- Purge-Orchestrierung (Registry-basiert, keine Entity-Kenntnisse)

### IModule

- **Posts** (posts + comments + likes) - Vollständig unabhängig
- **Profiles** (profiles + contacts) - Vollständig unabhängig
- **Warehouse** (inventory management) - Kommerziell, deaktiviert

### Architektur-Compliance

- Core enthält NUR Infrastruktur + Referenzdaten
- IModule sind VOLLSTÄNDIG unabhängig (keine Core-Imports)
- Konfiguration ist AUTONOM (jedes Modul besitzt Config)
- Lizenzierung wird UNTERSTÜTZT (bereit für kommerzielle IModule)

### Aktuelle Verbesserungen

- Purge-System refactored (Core = Orchestrator, IModule = Worker)
- Profiles/Contacts zu Modul migriert (aus Core entfernt)
- 15.000+ Codezeilen aus Core entfernt
- Vollständige Modulunabhängigkeit erreicht

---

## Verwandte Dokumentation

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Architektur-Compliance-Audit
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) - Schnellreferenz
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - Neue IModule erstellen
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Modulunabhängigkeitsprinzipien
- [../internal/CORE.md](../internal/CORE.md) - Core-Komponenten-Details
