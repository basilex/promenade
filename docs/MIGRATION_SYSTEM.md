# Migration System Summary

## Problem Solved

**Old System:**

- All migrations in single `migrations/` folder
- Global sequential numbering
- Core + module migrations mixed together
- Violates module independence

**New System:**

- Namespace-based migrations (core, posts, profiles, etc.)
- Independent versioning per namespace
- Each module owns its migrations
- Full module autonomy

---

## Quick Start

### 1. Create Migration

```bash
# Core migration
make migrate-create-core NAME=add_audit_tables

# Module migration
make migrate-create MODULE=posts NAME=add_post_views
```

**Creates:**

```
migrations/posts/
├── 000001_add_post_views.up.sql
└── 000001_add_post_views.down.sql
```

### 2. Run Migrations

```bash
# Migrate everything (core + enabled modules)
make migrate

# Migrate core only
make migrate-core

# Migrate specific module
make migrate-module MODULE=posts
```

### 3. Check Status

```bash
# Show status for all namespaces
make migrate-status

# Show version for specific namespace
make migrate-version MODULE=posts
```

### 4. Rollback

```bash
# Rollback 1 step
make migrate-rollback MODULE=posts STEPS=1

# Rollback 3 steps
make migrate-rollback MODULE=posts STEPS=3
```

---

## Directory Structure

```
migrations/
├── core/                          # Core infrastructure migrations
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_auth_tables.up.sql
│   ├── 000002_auth_tables.down.sql
│   └── ...
│
├── posts/                         # Posts module migrations
│   ├── 000001_create_posts.up.sql
│   ├── 000001_create_posts.down.sql
│   ├── 000002_create_comments.up.sql
│   ├── 000002_create_comments.down.sql
│   └── ...
│
└── profiles/                      # Profiles module migrations
    ├── 000001_create_profiles.up.sql
    ├── 000001_create_profiles.down.sql
    └── ...
```

**Each namespace has independent versioning:**

- Core: 1, 2, 3, 4, ...
- Posts: 1, 2, 3, ...
- Profiles: 1, 2, ...

---

## Database Schema

```sql
CREATE TABLE schema_migrations (
    version     BIGINT       NOT NULL,
    namespace   VARCHAR(50)  NOT NULL,  -- 'core', 'posts', 'profiles'
    dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, version)
);
```

**Example data:**

```
| namespace | version | dirty | applied_at          |
|-----------|---------|-------|---------------------|
| core      | 1       | false | 2024-12-22 10:00:00 |
| core      | 2       | false | 2024-12-22 10:00:01 |
| posts     | 1       | false | 2024-12-22 10:00:03 |
| posts     | 2       | false | 2024-12-22 10:00:04 |
| profiles  | 1       | false | 2024-12-22 10:00:05 |
```

---

## CLI Tool

```bash
# Run migrations
go run cmd/migrate/main.go -command=up -all
go run cmd/migrate/main.go -command=up -namespace=posts

# Rollback
go run cmd/migrate/main.go -command=down -namespace=posts -steps=1

# Status
go run cmd/migrate/main.go -command=status

# Version
go run cmd/migrate/main.go -command=version -namespace=posts
```

---

## Integration with Modules

Modules can optionally provide migrations programmatically:

```go
// pkg/module/module.go
type Module interface {
    // ... existing methods ...

    // RegisterMigrations returns migrations for this module
    // Return nil to use file-based migrations from migrations/{namespace}/
    RegisterMigrations() []Migration
}
```

**Example:**

```go
func (m *PostsModule) RegisterMigrations() []Migration {
    // Use file-based migrations
    return nil

    // OR embed in code
    return []Migration{
        {Version: 1, Up: `CREATE TABLE...`, Down: `DROP TABLE...`},
    }
}
```

---

## Benefits

**Module Independence**

- Each module owns its migrations
- Enable/disable modules without conflicts
- Clear ownership boundaries

  **Version Control**

- Independent versioning per namespace
- No global numbering conflicts
- Easy version tracking

  **Flexible Deployment**

- Deploy with only needed modules
- Add new modules without touching existing migrations
- Roll back module migrations independently

  **Developer Experience**

- Clear where to put migrations
- Auto-incrementing version numbers
- Namespace-specific commands

---

## Migration from Old System

**One-time reorganization needed:**

1. Create namespace directories:

   ```bash
   mkdir -p migrations/{core,posts,profiles}
   ```

2. Move existing migrations to namespaces and rename with descriptive names:

   ```bash
   # Core migrations (with new descriptive names)
   mv migrations/000001_init_schema_deps.* migrations/core/000001_core_init_uuid_v7.*
   mv migrations/000002_create_auth_schema.* migrations/core/000002_core_auth_full.*
   mv migrations/000009_create_rbac_tables.* migrations/core/000003_core_rbac_full.*
   # etc.

   # Posts module migrations (with module prefix)
   mv migrations/000006_create_user_posts.* migrations/posts/000001_posts_posts.*
   mv migrations/000007_create_post_comments.* migrations/posts/000002_posts_comments.*
   # etc.

   # Profiles module migrations (with module prefix)
   mv migrations/000004_create_user_contacts.* migrations/profiles/000001_create_user_contacts.*
   # etc.
   ```

3. Update schema_migrations table:

   ```sql
   -- Backup
   CREATE TABLE schema_migrations_backup AS SELECT * FROM schema_migrations;

   -- Drop old table
   DROP TABLE schema_migrations;

   -- Create new table (will be auto-created by migration manager)
   ```

4. Run new migration system:
   ```bash
   make migrate
   ```

**The migration manager will:**

- Create new schema_migrations table with namespace support
- Mark existing migrations as applied

---

## Implementation Files

- **`pkg/migration/manager.go`** - Migration manager implementation
- **`cmd/migrate/main.go`** - CLI tool
- **`scripts/create-migration.sh`** - Helper script for creating migrations
- **`Makefile.prod.mk`** - Updated with new commands
- **`docs/MIGRATION_ARCHITECTURE.md`** - Full architecture documentation

---

## Common Workflows

### Adding a New Module

1. Create module migrations directory:

   ```bash
   mkdir -p migrations/mymodule
   ```

2. Create first migration:

   ```bash
   make migrate-create MODULE=mymodule NAME=initial_schema
   ```

3. Edit SQL files:

   ```sql
   -- migrations/mymodule/000001_initial_schema.up.sql
   CREATE TABLE my_entities (...);

   -- migrations/mymodule/000001_initial_schema.down.sql
   DROP TABLE my_entities;
   ```

4. Run migration:

   ```bash
   make migrate-module MODULE=mymodule
   ```

5. Enable module in `config/modules.yaml`:
   ```yaml
   modules:
     enabled:
       - posts
       - profiles
       - mymodule # Add here
   ```

---

### Deploying to New Environment

```bash
# 1. Check current status
make migrate-status

# 2. Run all migrations
make migrate

# 3. Verify
make migrate-status
```

---

### Rolling Back a Failed Migration

```bash
# 1. Check status (will show "dirty" flag)
make migrate-status

# 2. Fix the migration SQL file

# 3. Manually mark as clean if needed (SQL)
UPDATE schema_migrations
SET dirty = false
WHERE namespace = 'posts' AND version = 3;

# 4. Re-run migration
make migrate-module MODULE=posts
```

---

## FAQ

**Q: What if I need to change a migration that's already applied?**
A: Don't modify applied migrations. Create a new migration with the changes.

**Q: Can I delete old migrations?**
A: No, keep all migrations in git history for reproducibility.

**Q: What happens if a module is disabled?**
A: Its migrations stay in the database. Re-enabling the module will apply any new migrations.

**Q: Can migrations have dependencies on other modules?**
A: Core is always migrated first. Modules should not depend on each other's tables.

**Q: What if two modules need the same table?**
A: Move that table to core, or use events for inter-module communication.

---

## Next Steps

1. **Review** [MIGRATION_ARCHITECTURE.md](MIGRATION_ARCHITECTURE.md) for full details
2. **Reorganize** existing migrations into namespaces
3. **Test** migration system in development
4. **Deploy** to staging/production

---

**Status:** Implemented and ready for use

**Documentation:**

- Full architecture: [MIGRATION_ARCHITECTURE.md](MIGRATION_ARCHITECTURE.md)
- Module development: [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md)
- Architecture overview: [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md)
