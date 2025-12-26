---
title: "Database & Migrations"
description: "PostgreSQL with UUID v7 and namespace-based migrations"
weight: 4
---

## Database Architecture

Promenade uses **PostgreSQL 16** with **UUID v7** primary keys and **namespace-based migrations**.

### UUID v7 - Time-Ordered IDs

Unlike random UUID v4, **UUID v7 is time-ordered**:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),  -- Not uuid_generate_v4()!
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Benefits:**

- ⚡ **2x faster inserts** - better B-tree locality
- 📉 **Reduced index fragmentation**
- 🔍 **Sortable by creation time**
- 🆔 **Still globally unique**

### Namespace-Based Migrations

Each module has **independent migration history**:

```
migrations/
├── core/                    # Core infrastructure
│   ├── 000001_auth.sql
│   └── 000002_rbac.sql
├── posts/                   # Posts module
│   ├── 000001_posts.sql
│   └── 000002_comments.sql
└── billing/                 # Billing module
    └── 000001_tables.sql
```

### True Module Autonomy

```bash
# Migrate all enabled modules
make migrate

# Migrate specific module
make migrate-module MODULE=posts

# Create new migration
make migrate-create MODULE=posts NAME=add_views_count
```

### No ORM - Raw SQL

Promenade uses **sqlx** (not ORM):

```go
// Clean, explicit queries
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
    return &user, r.Get(ctx, &user, query, email)
}
```

**Why No ORM:**

- Full SQL control
- No hidden N+1 queries
- Explicit performance tuning
- Zero abstraction overhead

### BaseRepository Pattern

```go
type BaseRepository struct {
    db *sqlx.DB
}

// Shared methods: Get, Select, Exec, NamedExec
// Each repo embeds BaseRepository
```

### Benefits

✅ **Fast** - UUID v7 improves insert performance  
✅ **Independent** - Module migrations don't conflict  
✅ **Explicit** - No ORM magic, full SQL control  
✅ **Transactional** - Context-aware transaction support

[UUID v7 Guide →](/promenade/docs/UUID_V7_GUIDE)  
[Migration Architecture →](/promenade/docs/MIGRATION_ARCHITECTURE)
