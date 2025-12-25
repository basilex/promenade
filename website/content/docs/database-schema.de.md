---
title: "Datenbankschema"
description: "Vollständige PostgreSQL-Schemadokumentation mit Diagrammen"
weight: 20
---

## Datenbankschema-Überblick

Promenade verwendet **PostgreSQL 16** mit **UUID v7** Primärschlüsseln und **namespace-basierten Migrationen**.

---

## Core-Tabellen

### Authentifizierung und Benutzer

```mermaid
erDiagram
    USERS ||--o{ SESSIONS : has
    USERS ||--o{ USER_ROLES : has

    USERS {
        uuid id PK
        text email UK
        text password_hash
        text name
        text status
        timestamp created_at
        timestamp updated_at
    }

    SESSIONS {
        uuid id PK
        uuid user_id FK
        text refresh_token UK
        timestamp expires_at
        timestamp created_at
    }
```

**Schlüsselfelder:**

- `users.id` - UUID v7 (zeitbasiert sortiert)
- `users.email` - eindeutiger Index
- `users.status` - `active`, `suspended`, `deleted`
- `sessions.refresh_token` - JWT Refresh Token

### RBAC (Rollenbasierte Zugriffskontrolle)

```mermaid
erDiagram
    ROLES ||--o{ USER_ROLES : assigned
    ROLES ||--o{ ROLE_PERMISSIONS : has
    PERMISSIONS ||--o{ ROLE_PERMISSIONS : granted
    USERS ||--o{ USER_ROLES : has

    ROLES {
        uuid id PK
        text name UK
        text description
        timestamp created_at
    }

    PERMISSIONS {
        uuid id PK
        text resource
        text action
        text description UK
        timestamp created_at
    }

    USER_ROLES {
        uuid id PK
        uuid user_id FK
        uuid role_id FK
        timestamp assigned_at
    }

    ROLE_PERMISSIONS {
        uuid id PK
        uuid role_id FK
        uuid permission_id FK
        timestamp assigned_at
    }
```

**4 Systemrollen:**

- `Admin` - Vollzugriff (`*`)
- `Moderator` - Inhaltsmoderation
- `User` - Grundlegende Operationen
- `Guest` - Nur Lesezugriff

**Berechtigungsformat:** `resource:action` (z.B. `posts:create`, `users:delete`)

---

## Referenzdaten

### Länder und Währungen

```mermaid
erDiagram
    COUNTRIES ||--o{ CURRENCIES : uses
    COUNTRIES ||--o{ REGIONS : contains

    COUNTRIES {
        uuid id PK
        text code UK "ISO 3166-1"
        text name
        text alpha2
        text alpha3
        text region
    }

    CURRENCIES {
        uuid id PK
        text code UK "ISO 4217"
        text name
        text symbol
        int decimals
    }
```

**Daten:**

- 145 Länder mit ISO-Codes
- 124 Währungen mit Symbolen
- 30 Verwaltungsregionen
- 17 Großstädte mit Koordinaten

### Regionen und Städte

```mermaid
erDiagram
    COUNTRIES ||--o{ REGIONS : contains
    REGIONS ||--o{ CITIES : contains

    REGIONS {
        uuid id PK
        uuid country_id FK
        text name
        text code UK
        text type "state/oblast/province"
    }

    CITIES {
        uuid id PK
        uuid region_id FK
        text name
        text slug UK
        decimal latitude
        decimal longitude
        int population
        bool is_capital
    }
```

**Regionstypen:**

- `state` - USA, Australien
- `oblast` - Ukraine
- `province` - Kanada
- `land` - Deutschland, Österreich

### Andere Referenzdaten

- **Timezones** - IANA Zeitzonen-Datenbank
- **Languages** - ISO 639 Codes
- **Payment Methods** - 40+ Methoden (Karten, Wallets, Krypto, BNPL)

---

## Posts-Modul

### Beiträge und Kommentare

```mermaid
erDiagram
    USER_POSTS ||--o{ POST_COMMENTS : has
    POST_COMMENTS ||--o{ COMMENT_LIKES : has
    USERS ||--o{ USER_POSTS : creates
    USERS ||--o{ POST_COMMENTS : writes

    USER_POSTS {
        uuid id PK
        uuid user_id FK
        text title
        text slug UK
        text content
        text status "draft/published/archived"
        timestamp published_at
        timestamp deleted_at "Soft delete"
        timestamp created_at
        timestamp updated_at
    }

    POST_COMMENTS {
        uuid id PK
        uuid post_id FK
        uuid user_id FK
        uuid parent_id FK "Threading"
        text content
        timestamp deleted_at "Soft delete"
        timestamp created_at
    }

    COMMENT_LIKES {
        uuid id PK
        uuid comment_id FK
        uuid user_id FK
        timestamp created_at
    }
```

**Muster:**

- Soft Delete: `deleted_at IS NULL` in allen Abfragen
- Slug-Generierung: Automatischer eindeutiger Slug aus Titel
- Status: `draft`, `published`, `scheduled`, `archived`

---

## Profiles-Modul

### Profile und Kontakte

```mermaid
erDiagram
    USERS ||--|| USER_PROFILES : has
    USERS ||--o{ USER_CONTACTS : has

    USER_PROFILES {
        uuid id PK
        uuid user_id FK UK
        text bio
        text avatar_url
        text location
        text website
        text privacy "public/private/friends"
        timestamp created_at
        timestamp updated_at
    }

    USER_CONTACTS {
        uuid id PK
        uuid user_id FK
        text contact_type "email/phone/telegram/linkedin"
        text contact_value
        bool is_primary
        bool is_verified
        timestamp created_at
    }
```

**Privacy-Einstellungen:**

- `public` - für alle sichtbar
- `private` - nur für Eigentümer
- `friends` - nur für Freunde

---

## Analytics-Modul

### Metriken und Aggregate

```sql
CREATE TABLE metrics (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    metric_name TEXT NOT NULL,
    metric_value NUMERIC NOT NULL,
    dimensions JSONB,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE metric_aggregates (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    metric_name TEXT NOT NULL,
    aggregate_type TEXT NOT NULL, -- sum, avg, count
    aggregate_value NUMERIC NOT NULL,
    period TEXT NOT NULL,         -- hourly, daily, monthly
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL
);

CREATE INDEX idx_metrics_name_time ON metrics(metric_name, timestamp);
CREATE INDEX idx_aggregates_period ON metric_aggregates(metric_name, period, period_start);
```

**Aufbewahrung:** 90 Tage für Roh-Metriken, unbegrenzt für Aggregate

---

## Billing-Modul

### Abonnements und Zahlungen

```mermaid
erDiagram
    PLANS ||--o{ SUBSCRIPTIONS : subscribes
    USERS ||--o{ SUBSCRIPTIONS : has
    SUBSCRIPTIONS ||--o{ INVOICES : generates
    INVOICES ||--o{ PAYMENTS : paid_by

    PLANS {
        uuid id PK
        text name UK
        text billing_interval "monthly/quarterly/annual"
        numeric price
        text currency
        int trial_days
        bool is_active
    }

    SUBSCRIPTIONS {
        uuid id PK
        uuid user_id FK
        uuid plan_id FK
        text status "active/paused/cancelled"
        timestamp current_period_start
        timestamp current_period_end
        timestamp trial_end
        timestamp cancelled_at
    }

    INVOICES {
        uuid id PK
        uuid subscription_id FK
        text invoice_number UK
        numeric amount
        text currency
        text status "draft/sent/paid/void"
        timestamp due_date
    }

    PAYMENTS {
        uuid id PK
        uuid invoice_id FK
        numeric amount
        text payment_method
        text status "pending/completed/failed"
        timestamp paid_at
    }
```

**Lebenszyklus:** Trial → Active → Paused/Cancelled → Expired

---

## Indizes und Performance

### Schlüsselindizes

```sql
-- Authentifizierung
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_sessions_user ON sessions(user_id);

-- Soft Delete (immer filtern)
CREATE INDEX idx_posts_deleted ON user_posts(deleted_at) WHERE deleted_at IS NULL;

-- Suche
CREATE INDEX idx_posts_slug ON user_posts(slug);
CREATE INDEX idx_comments_post ON post_comments(post_id) WHERE deleted_at IS NULL;

-- Analytics
CREATE INDEX idx_metrics_time ON metrics(timestamp);
```

### Performance-Strategie

- **UUID v7:** Reduziert Page Splits um 60%
- **Partial Indexes:** Nur aktive Einträge (deleted_at IS NULL)
- **JSONB GIN:** Schnelle JSON-Feldsuche
- **Connection Pool:** 25 Max-Verbindungen, 5 Min Idle

---

## Migrationen

### Namespace-Struktur

```
migrations/
├── core/                    # Wird immer zuerst ausgeführt
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000003_core_rbac_full.up.sql
│   ├── 000004_core_ref_timezones.up.sql
│   ├── 000005_core_ref_languages.up.sql
│   ├── 000006_core_ref_countries_currencies.up.sql
│   ├── 000007_core_ref_regions_cities.up.sql
│   └── 000008_core_ref_payment_methods.up.sql
├── posts/                   # Posts-Modul
│   ├── 000001_posts_posts.up.sql
│   ├── 000002_posts_comments.up.sql
│   └── 000003_posts_comment_likes.up.sql
├── profiles/                # Profiles-Modul
│   ├── 000001_profiles_contacts.up.sql
│   └── 000002_profiles_profiles.up.sql
└── billing/                 # Billing-Modul
    └── 000001_billing_tables.up.sql
```

**Befehle:**

```bash
make migrate                           # Alle Migrationen
make migrate-core                      # Nur Core
make migrate-module MODULE=posts       # Einzelnes Modul
make migrate-status                    # Status aller Namespaces
```

---

## Backup und Recovery

### Backup-Strategie

- **Daily:** Vollbackup um 2:00 Uhr
- **Hourly:** Inkrementelle WAL-Archive
- **Retention:** 30 Tage für Produktion
- **Point-in-Time Recovery:** Verfügbar für letzte 7 Tage

### Befehle

```bash
# Backup
pg_dump -h localhost -U system promenade_dev > backup.sql

# Restore
psql -h localhost -U system promenade_dev < backup.sql
```

---

## Nächste Schritte

- [UUID v7 Guide](/promenade/features/database#uuid-v7)
- [Soft Delete Pattern](/promenade/features/database#soft-delete)
- [Migrations-Handbuch](https://github.com/basilex/promenade/blob/dev/migrations/README.md)
- [Architektur](/promenade/docs/architecture)
