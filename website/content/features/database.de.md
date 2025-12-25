---
title: "Datenbank & Migrationen"
description: "PostgreSQL mit UUID v7 und namensraumbasierten Migrationen"
weight: 4
---

## Datenbankarchitektur

Promenade verwendet **PostgreSQL 16** mit **UUID v7** Primärschlüsseln und **namensraumbasierten Migrationen**.

### UUID v7 - Zeitgeordnete IDs

Im Gegensatz zu zufälligen UUID v4 sind **UUID v7 zeitgeordnet**:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),  -- Nicht uuid_generate_v4()!
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Vorteile:**

- ⚡ **2x schnellere Einfügungen** - bessere B-Baum-Lokalität
- 📉 **Reduzierte Indexfragmentierung**
- 🔍 **Nach Erstellungszeit sortierbar**
- 🆔 **Dennoch global eindeutig**

### Namensraumbasierte Migrationen

Jedes Modul hat **unabhängige Migrationshistorie**:

```
migrations/
├── core/                    # Kerninfrastruktur
│   ├── 000001_auth.sql
│   └── 000002_rbac.sql
├── posts/                   # Posts-Modul
│   ├── 000001_posts.sql
│   └── 000002_comments.sql
└── billing/                 # Billing-Modul
    └── 000001_tables.sql
```

### Echte Modulautonomie

```bash
# Alle aktivierten Module migrieren
make migrate

# Spezifisches Modul migrieren
make migrate-module MODULE=posts

# Neue Migration erstellen
make migrate-create MODULE=posts NAME=add_views_count
```

### Kein ORM - Reines SQL

Promenade verwendet **sqlx** (kein ORM):

```go
// Saubere, explizite Abfragen
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
    return &user, r.Get(ctx, &user, query, email)
}
```

**Warum Kein ORM:**

- Volle SQL-Kontrolle
- Keine versteckten N+1-Abfragen
- Explizites Performance-Tuning
- Null Abstraktions-Overhead

### BaseRepository-Muster

```go
type BaseRepository struct {
    db *sqlx.DB
}

// Geteilte Methoden: Get, Select, Exec, NamedExec
// Jedes Repo bettet BaseRepository ein
```

### Vorteile

✅ **Schnell** - UUID v7 verbessert Einfügeleistung  
✅ **Unabhängig** - Modulmigrationen kollidieren nicht  
✅ **Explizit** - Keine ORM-Magie, volle SQL-Kontrolle  
✅ **Transaktional** - Kontextbewusste Transaktionsunterstützung

[Ausführlicher Leitfaden →](/docs/UUID_V7_GUIDE)
