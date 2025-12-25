---
title: "Architektur"
description: "Umfassende Architekturdokumentation mit Diagrammen"
weight: 10
---

## Architekturüberblick

Promenade folgt den Prinzipien der **Clean Architecture** mit einem **modularen Plugin-System**.

### Systemschichten

```mermaid
graph TB
    A[Application Entry Point] --> B[Core Infrastructure]
    B --> C[Core Domain]
    B --> D[Module System]
    C --> E[Core Use Cases]
    D --> F[Module Domains]
    F --> G[Module Use Cases]
    E --> H[Core Adapters]
    G --> I[Module Adapters]
    H --> J[Database/HTTP/Events]
    I --> J
```

### Core vs. Module

**Core = Infrastruktur + Sicherheit + Referenzdaten**

- Immer aktiviert
- Authentifizierung (JWT)
- RBAC (Rollen, Berechtigungen)
- Referenzdaten (Länder, Währungen, Regionen, Städte, Zahlungsmethoden, Zeitzonen, Sprachen)
- Event Bus, Datenbank, Logger, Scheduler

**Module = Geschäftslogik**

- Optional, aktivierbar/deaktivierbar
- Vollständige vertikale Schichten (Entity → UseCase → Adapter)
- Eigene Migrationen, Konfigurationen, Berechtigungen
- Separat lizenzierbar (kostenlos oder kommerziell)

---

## Clean Architecture Schichten

### Abhängigkeitsfluss

```mermaid
graph LR
    A[Domain Layer] --> B[Use Case Layer]
    B --> C[Adapter Layer]
    C --> D[Infrastructure Layer]

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#a78bfa
    style D fill:#c084fc
```

**Kernregel:** Innere Schichten hängen niemals von äußeren ab!

### Schichtenverantwortlichkeiten

**1. Domain Layer** (`internal/domain/entity/`)

- Reine Geschäftsentitäten
- Validierungsmethoden
- Keine Framework-Abhängigkeiten
- Beispiel: `User`, `Post`, `Comment`

**2. Use Case Layer** (`internal/usecase/`)

- Nur Geschäftslogik
- Abhängig von Domain-Interfaces
- Beispiel: `RegisterUser`, `CreatePost`

**3. Adapter Layer** (`internal/adapter/`)

- HTTP-Handler
- Repository-Implementierungen
- Framework-Integration
- Beispiel: `UserHandler`, `PostgresUserRepo`

**4. Infrastructure Layer** (`internal/infrastructure/`)

- Datenbankverbindungen
- Event Bus
- Externe Dienste
- Beispiel: `PostgreSQL`, `Redis`, `SMTP`

---

## Modularchitektur

### Modulstruktur

```
internal/modules/posts/
├── module.go              # Modul-Interface-Implementierung
├── register.go            # Auto-Registrierung
├── config/                # Eigene Umgebungskonfigurationen
├── entity/                # Domain-Entitäten
├── usecase/               # Geschäftslogik
└── adapter/
    ├── http/              # Handler, DTO, Routen
    └── repository/        # Datenzugriff
```

### Modul-Lebenszyklus

```mermaid
sequenceDiagram
    participant A as Application
    participant R as Registry
    participant M as Module
    participant C as Core

    A->>R: Register modules (init)
    R->>M: Initialize(core)
    M->>C: Use DB, EventBus, etc.
    M->>R: RegisterRoutes()
    M->>R: RegisterPermissions()
    A->>M: Start()
    Note over M: Background workers
    A->>M: Stop() on shutdown
```

### Modulunabhängigkeit

**Module importieren NIEMALS:**

- `internal/domain` (Core-Domain)
- `internal/usecase` (Core Use Cases)
- `internal/adapter` (Core-Adapter)
- Andere Module

**Module KÖNNEN verwenden:**

- `pkg/*` (gemeinsame Pakete)
- Core-Infrastruktur über `module.Core`
- Events für Modul-zu-Modul-Kommunikation

---

## Kommunikationsmuster

### Event-Driven Communication

```mermaid
graph LR
    A[Posts Module] -->|user.registered| B[Event Bus]
    B -->|Subscribe| C[Email Module]
    B -->|Subscribe| D[Analytics Module]
    B -->|Subscribe| E[Audit Module]

    style A fill:#38bdf8
    style B fill:#fbbf24
    style C fill:#818cf8
    style D fill:#a78bfa
    style E fill:#c084fc
```

**Vorteile:**

- Lose Kopplung zwischen Modulen
- Asynchrone Verarbeitung
- Einfaches Hinzufügen neuer Abonnenten
- Keine direkten Modulabhängigkeiten

---

## Datenbankarchitektur

### UUID v7 Primärschlüssel

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Vorteile:**

- ⚡ 2x schnellere Einfügungen (B-Tree-Lokalität)
- 📉 Reduzierte Index-Fragmentierung
- 🔍 Zeitbasierte Sortierung
- 🆔 Global eindeutig

### Namespace-basierte Migrationen

```
migrations/
├── core/                    # Core wird immer zuerst ausgeführt
│   ├── 000001_init.sql
│   ├── 000002_auth.sql
│   └── 000003_rbac.sql
├── posts/                   # Posts-Modul
│   ├── 000001_posts.sql
│   └── 000002_comments.sql
└── billing/                 # Billing-Modul
    └── 000001_tables.sql
```

**Jeder Namespace hat eine unabhängige Versionshistorie!**

---

## Purge-System

### Registry-basiertes Design

```mermaid
graph TB
    A[Core Scheduler] -->|Runs cron| B[Purge Use Case]
    B -->|Gets policies| C[Policy Registry]
    B -->|Gets handlers| D[Handler Registry]

    E[Posts Module] -->|Register| C
    E -->|Register| D

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#fbbf24
    style D fill:#fbbf24
```

**Kernpunkte:**

- Core weiß WANN bereinigt wird (Schedule, Batch-Größe)
- Module definieren WAS bereinigt wird (Entitäten, Aufbewahrungstage)
- Echte Modulunabhängigkeit durch Registries

---

## Testen

### Test-Pyramide

```mermaid
graph TB
    A[Smoke Tests<br/>5-10 tests<br/>E2E critical flows]
    B[Integration Tests<br/>50+ tests<br/>Real DB]
    C[Unit Tests<br/>400+ tests<br/>Manual mocks]

    A --> B
    B --> C

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#a78bfa
```

**Abdeckung:** Core: 275 Tests | Posts: 33 | Profiles: 21 | Billing: 375 | Analytics: 11

**Gesamt: 400+ Tests laufen in ~20 Sekunden**

---

## Deployment

### Entwicklung

```mermaid
graph LR
    A[Developer] -->|make dev| B[Docker Compose]
    B --> C[PostgreSQL:5432]
    B --> D[Redis:6379]
    E[Go App:8081] --> C
    E --> D
```

### Produktion

```mermaid
graph TB
    A[GitHub Actions] -->|Deploy| B[Docker Image]
    B --> C[Kubernetes Pod]
    C --> D[PostgreSQL RDS]
    C --> E[Redis ElastiCache]
    G[Load Balancer] --> C

    style A fill:#38bdf8
    style C fill:#818cf8
    style D fill:#fbbf24
    style E fill:#fbbf24
```

---

## Nächste Schritte

- [Modul-Entwicklungshandbuch](/promenade/docs/module-development)
- [Datenbankschema](/promenade/docs/database-schema)
- [Schnellstart](/promenade/docs/getting-started)
- [Features](/promenade/features)
