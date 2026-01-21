# Namespace-Based Migration System

## Overview

Promenade uses a **namespace-based migration system** where each module maintains its own independent migration history. This ensures true module autonomy - modules can be enabled/disabled without affecting other modules' database state.

**Multi-database support**: Migrations are organized by database driver, allowing different SQL syntax for PostgreSQL and MS SQL Server.

## Structure

```
migrations/
 postgres/       # PostgreSQL migrations (production-ready)
    core/        # Core infrastructure (auth, RBAC, reference data)
       000001_init_schema_deps.up.sql
       000001_init_schema_deps.down.sql
       000002_create_auth_schema.up.sql
       000003_create_rbac_tables.up.sql
       ...
    identity/    # Identity module (users, contacts, profiles)
       000001_create_users_table.up.sql
       000001_create_users_table.down.sql
       ...
    customer-mgmt/  # Customer management
    order-mgmt/     # Order management
    billing/        # Billing
    warehouse/      # Warehouse
    accounting/     # Accounting
    banking/        # Banking
    fiscal/         # Fiscal integration
    shared/         # Shared reference data
    ui/             # UI metadata
    scripting/      # Scripting engine

 mssql/          # MS SQL Server migrations (planned)
    core/        # (To be implemented)
    identity/
    ...
```

**Note**: Migration manager automatically selects the correct directory based on `DATABASE_DRIVER` configuration.

## Key Features

### Independent Versioning

Each namespace maintains its own version sequence starting from 000001:

- **core**: v1-v8 (schema deps, auth, RBAC, timezones, languages, countries/currencies, regions/cities, payment methods)
- **posts**: v1-v3 (user_posts, post_comments, comment_likes)
- **profiles**: v1-v2 (user_contacts, user_profiles)

### schema_migrations Table

```sql
CREATE TABLE schema_migrations (
    version     BIGINT       NOT NULL,
    namespace   VARCHAR(50)  NOT NULL,
    dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, version)
);
```

## Usage

### CLI Tool Commands

```bash
# Check migration status for all namespaces
make migrate-status

# Run all migrations (core + enabled modules)
make migrate

# Run migrations for specific namespace
make migrate-core                    # Core only
make migrate-module MODULE=posts     # Specific module

# Rollback migrations
make migrate-rollback MODULE=posts STEPS=1

# Check version for namespace
make migrate-version MODULE=posts

# Create new migration
make migrate-create MODULE=posts NAME=add_post_tags
make migrate-create-core NAME=add_audit_log
```

### Automatic Migrations

Migrations run automatically on application startup:

- **Core** migrations always run first
- **Module** migrations run for all namespaces in sequence

The migration manager automatically uses the correct database-specific migrations based on `DATABASE_DRIVER` configuration.

See [cmd/api/bootstrap.go](../cmd/api/bootstrap.go):

```go
driver := cfg.Database.Driver  // "postgres" or "mssql"
migrationManager := migration.NewManager(db, driver, "migrations")
// Runs migrations from migrations/{driver}/{namespace}/
```

### Manual Migration Management

```bash
# Build migration CLI tool
go build -o bin/migrate ./cmd/migrate/main.go

# Run migrations (uses DATABASE_DRIVER from config)
./bin/migrate -cmd=up -all                # All namespaces
./bin/migrate -cmd=up -namespace=core     # Specific namespace

# Rollback
./bin/migrate -cmd=down -namespace=core -steps=2

# Check status
./bin/migrate -cmd=status

# Check version
./bin/migrate -cmd=version -namespace=core
```

## Creating New Migrations

### Using Script (Recommended)

```bash
# Create migration for module
./scripts/create-migration.sh posts add_post_views

# Output:
# migrations/posts/000004_add_post_views.up.sql
# migrations/posts/000004_add_post_views.down.sql
```

### Manually

1. Determine next version number for your namespace:

   ```bash
   ls migrations/posts/*.up.sql | tail -1
   # 000003_create_comment_likes_table.up.sql → next is 000004
   ```

2. Create migration files in the correct database directory:

   ```bash
   # PostgreSQL
   touch migrations/postgres/core/000004_add_audit_log.up.sql
   touch migrations/postgres/core/000004_add_audit_log.down.sql

   # MS SQL Server (when implemented)
   touch migrations/mssql/core/000004_add_audit_log.up.sql
   touch migrations/mssql/core/000004_add_audit_log.down.sql
   ```

3. Write SQL:
   - **UP**: DDL to apply changes
   - **DOWN**: DDL to revert changes
   - **Note**: Use database-specific syntax appropriate for the driver

## Database-Specific Considerations

### PostgreSQL (migrations/postgres/)

- Use PostgreSQL-native features: JSONB, UUID, materialized views
- PL/pgSQL for stored procedures
- `RETURNING` clause for insert/update operations
- GIN indexes for JSONB columns

### MS SQL Server (migrations/mssql/)

- Use T-SQL syntax
- `UNIQUEIDENTIFIER` for UUIDs
- `NVARCHAR(MAX)` for JSON storage
- `OUTPUT INSERTED` instead of `RETURNING`
- Indexed Views instead of materialized views

**Migration manager automatically selects the correct directory based on DATABASE_DRIVER configuration.**

## Migration Naming Convention

```
NNNNNN_description.{up|down}.sql

NNNNNN      - 6-digit version number (000001, 000002, etc.)
description - Snake_case description (create_table, add_column, etc.)
up          - Forward migration
down        - Rollback migration
```

Examples:

- `000001_create_user_posts.up.sql`
- `000002_add_post_status_index.up.sql`
- `000003_alter_comments_add_edited_at.up.sql`

## Best Practices

### 1. Always Write Down Migrations

Every UP migration must have a corresponding DOWN migration for safe rollbacks.

### 2. Transactional Migrations

Migrations automatically run in transactions. If any part fails, the entire migration is rolled back and marked as "dirty".

### 3. Test Before Committing

```bash
# Apply migration
make migrate-module MODULE=posts

# Verify
docker exec promenade_postgres psql -U system -d promenade_dev -c '\dt'

# Test rollback
make migrate-rollback MODULE=posts STEPS=1

# Re-apply
make migrate-module MODULE=posts
```

### 4. IModule Dependencies

If your module depends on tables from another module:

- Document the dependency in module README
- Consider if the table should be in `core` instead
- Use foreign keys carefully (respect module boundaries)

### 5. Data Migrations

For complex data transformations:

```sql
-- 000005_migrate_old_format_data.up.sql
BEGIN;

-- Transform data
UPDATE user_posts
SET new_column = old_column::jsonb
WHERE new_column IS NULL;

-- Cleanup
ALTER TABLE user_posts DROP COLUMN old_column;

COMMIT;
```

## Migration States

### Clean State

```
Namespace: posts
  Current Version: 3
  Pending: 0
  Status: OK
```

### Dirty State (Failed Migration)

```
Namespace: posts
  Current Version: 3
  Pending: 1
  Status: DIRTY (manual intervention required)
```

**Recovery:**

1. Fix the SQL error in migration file
2. Manually delete the dirty record:
   ```sql
   DELETE FROM schema_migrations
   WHERE namespace='posts' AND version=4 AND dirty=true;
   ```
3. Re-run migration: `make migrate-module MODULE=posts`

## Troubleshooting

### "No migrations found for namespace"

- Check directory exists: `ls migrations/posts/`
- Verify filename format: `NNNNNN_description.{up|down}.sql`

### "Migration failed: namespace X is in dirty state"

See "Dirty State" section above for recovery steps.

### "Foreign key constraint violation"

- Ensure dependent tables exist (check migration order)
- Consider adding `IF NOT EXISTS` checks
- Verify namespace execution order (core → modules)

### IModule migrations not running

- Check module is enabled in `config/modules.yaml`
- Verify `cfg.Modules.Enabled` includes the module
- Run manually: `make migrate-module MODULE=yourmodule`

## Architecture

See full documentation:

- [docs/INDEX.md](../docs/INDEX.md) - Documentation index
- [docs/PHASE1_ARCHITECTURE_PREPARATION.md](../docs/PHASE1_ARCHITECTURE_PREPARATION.md) - Migration roadmap
- [pkg/migration/manager.go](../pkg/migration/manager.go) - Implementation

## Examples

### Creating a New IModule's First Migration

```bash
# 1. Create module directory
mkdir -p internal/modules/warehouse

# 2. Create migration directory
mkdir -p migrations/warehouse

# 3. Create initial migration
./scripts/create-migration.sh warehouse create_products_table

# 4. Write SQL
cat > migrations/warehouse/000001_create_products_table.up.sql <<EOF
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
EOF

# 5. Write DOWN migration
cat > migrations/warehouse/000001_create_products_table.down.sql <<EOF
DROP TABLE IF EXISTS products;
EOF

# 6. Test
make migrate-module MODULE=warehouse
```

## References

- **Migration Manager**: [pkg/migration/manager.go](../pkg/migration/manager.go)
- **CLI Tool**: [cmd/migrate/main.go](../cmd/migrate/main.go)
- **Helper Script**: [scripts/create-migration.sh](../scripts/create-migration.sh)
