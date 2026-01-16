# Migration Package

**Purpose:** Namespace-based database migrations management  
**Status:** Production-ready  
**Tests:** 8 tests, 90% coverage

---

## Overview

The `migration` package provides a **namespace-based migration system** for managing database schema changes across multiple bounded contexts. Unlike traditional migration tools, it supports **isolated namespaces** (core, shared, identity, customer-mgmt, etc.) allowing contexts to evolve independently.

## Features

- **Namespace Isolation:** Each context has its own migration namespace
- **Sequential Execution:** Migrations run in order: core → shared → identity → customer-mgmt → ...
- **Versioning:** Tracks applied migrations per namespace
- **Rollback Support:** Safely revert migrations
- **Dirty State Detection:** Prevents partial migrations
- **Transactional:** Each migration runs in a transaction

---

## Installation

```go
import "github.com/basilex/promenade/pkg/migration"
```

---

## Quick Start

### Initialize Migration Manager

```go
package main

import (
    "context"
    "github.com/basilex/promenade/pkg/migration"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to database
    db, _ := sqlx.Connect("postgres", "postgresql://user:pass@localhost/db")
    
    // Create migration manager
    manager := migration.NewManager(db, "migrations")
    
    // Run migrations for a namespace
    ctx := context.Background()
    err := manager.MigrateNamespace(ctx, "core")
    if err != nil {
        panic(err)
    }
}
```

---

## Directory Structure

Migrations are organized by namespace:

```
migrations/
  core/                             # Core infrastructure
    000001_extensions.up.sql        # Create UUID v7 extension
    000001_extensions.down.sql      # Drop UUID v7 extension
    000002_auth.up.sql              # Authentication tables
    000002_auth.down.sql
  shared/                           # Shared context (reference data)
    000001_reference_data.up.sql
    000001_reference_data.down.sql
  identity/                         # Identity context
    000001_users.up.sql
    000001_users.down.sql
    000002_contacts.up.sql
    000002_contacts.down.sql
  customer-mgmt/                    # Customer Management context
    000001_customers.up.sql
    000001_customers.down.sql
```

### Naming Convention

Format: `{version}_{description}.{direction}.sql`

- **Version:** 6-digit zero-padded (000001, 000002, ...)
- **Description:** Snake_case description
- **Direction:** `up` (apply) or `down` (rollback)

**Examples:**
- `000001_create_users_table.up.sql`
- `000001_create_users_table.down.sql`
- `000002_add_email_verification.up.sql`

---

## Usage Examples

### 1. Run All Namespaces

```go
ctx := context.Background()

// Run core migrations first
manager.MigrateNamespace(ctx, "core")

// Run shared kernel
manager.MigrateNamespace(ctx, "shared")

// Run bounded contexts
manager.MigrateNamespace(ctx, "identity")
manager.MigrateNamespace(ctx, "customer-mgmt")
manager.MigrateNamespace(ctx, "order-mgmt")
```

### 2. Run Single Namespace

```go
// Run only identity migrations
err := manager.MigrateNamespace(ctx, "identity")
if err != nil {
    log.Fatal("Migration failed:", err)
}
```

### 3. Check Migration Status

```go
status, err := manager.Status(ctx)
if err != nil {
    log.Fatal(err)
}

for namespace, info := range status {
    fmt.Printf("Namespace: %s\n", namespace)
    fmt.Printf("  Current Version: %d\n", info.CurrentVersion)
    fmt.Printf("  Pending: %d\n", info.PendingCount)
    fmt.Printf("  Dirty: %v\n", info.Dirty)
}
```

**Output:**
```
Namespace: core
  Current Version: 2
  Pending: 0
  Dirty: false
Namespace: identity
  Current Version: 5
  Pending: 1
  Dirty: false
```

### 4. Rollback Migrations

```go
// Rollback last 2 migrations in identity namespace
err := manager.Rollback(ctx, "identity", 2)
if err != nil {
    log.Fatal(err)
}
```

### 5. Get Current Version

```go
version, err := manager.Version(ctx, "identity")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Identity context version: %d\n", version)
```

---

## Migration File Examples

### 1. Create Table (UP)

**File:** `migrations/identity/000001_users.up.sql`

```sql
-- Create identity_users table
CREATE TABLE identity_users (
    id              UUID PRIMARY KEY DEFAULT uuid_v7_generate(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP NULL
);

-- Create indexes
CREATE INDEX idx_users_email ON identity_users(email);
CREATE INDEX idx_users_status ON identity_users(status);
CREATE INDEX idx_users_deleted_at ON identity_users(deleted_at);

-- Add comment
COMMENT ON TABLE identity_users IS 'User accounts for authentication';
```

### 2. Drop Table (DOWN)

**File:** `migrations/identity/000001_users.down.sql`

```sql
-- Drop indexes first
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email;

-- Drop table
DROP TABLE IF EXISTS identity_users;
```

### 3. Add Column (UP)

**File:** `migrations/identity/000002_add_last_login.up.sql`

```sql
-- Add last_login_at column
ALTER TABLE identity_users
ADD COLUMN last_login_at TIMESTAMP NULL;

-- Add index
CREATE INDEX idx_users_last_login 
ON identity_users(last_login_at) 
WHERE last_login_at IS NOT NULL;
```

### 4. Add Column Rollback (DOWN)

**File:** `migrations/identity/000002_add_last_login.down.sql`

```sql
-- Drop index
DROP INDEX IF EXISTS idx_users_last_login;

-- Drop column
ALTER TABLE identity_users
DROP COLUMN IF EXISTS last_login_at;
```

---

## CLI Tool

Promenade includes a migration CLI tool:

```bash
# Run all migrations
go run cmd/migrate/main.go --cmd=up

# Run specific namespace
go run cmd/migrate/main.go --cmd=up --namespace=identity

# Rollback last migration
go run cmd/migrate/main.go --cmd=down --namespace=identity --steps=1

# Check status
go run cmd/migrate/main.go --cmd=status

# Create new migration
./scripts/create-migration.sh identity add_email_verification
```

---

## Makefile Commands

The project includes convenient make targets:

```bash
# Run all migrations (core + shared + identity + customer-mgmt)
make migrate

# Run specific namespace
make migrate-core           # Core infrastructure
make migrate-shared         # Shared context
make migrate-identity       # Identity context
make migrate-customer-mgmt  # Customer Management

# Create new migration
make migrate-new CONTEXT=identity NAME=add_email_verification

# Check status
make migrate-status
```

---

## API Reference

### Manager Interface

```go
type Manager interface {
    // MigrateNamespace applies all pending migrations for a namespace
    MigrateNamespace(ctx context.Context, namespace string) error
    
    // Rollback rolls back N migrations for a namespace
    Rollback(ctx context.Context, namespace string, steps int) error
    
    // Version returns current version for a namespace
    Version(ctx context.Context, namespace string) (int, error)
    
    // Status returns migration status for all namespaces
    Status(ctx context.Context) (map[string]MigrationStatus, error)
}
```

### MigrationStatus

```go
type MigrationStatus struct {
    Namespace      string  // Namespace name
    CurrentVersion int     // Current applied version
    PendingCount   int     // Number of pending migrations
    Dirty          bool    // True if last migration failed
}
```

---

## Best Practices

### DO

- **One change per migration** - Keep migrations focused and atomic
  ```sql
  -- Good: Single table creation
  CREATE TABLE users (...);
  
  -- Bad: Multiple unrelated changes
  CREATE TABLE users (...);
  CREATE TABLE orders (...);
  ALTER TABLE products ...;
  ```

- **Always write DOWN migrations** - Enable rollback capability
  ```sql
  -- 000001_create_users.down.sql
  DROP TABLE IF EXISTS users;
  ```

- **Use transactions** - The manager wraps each migration in a transaction
  ```sql
  -- No need for BEGIN/COMMIT - handled automatically
  CREATE TABLE users (...);
  CREATE INDEX idx_users_email ON users(email);
  ```

- **Test migrations on copy of production data**
  ```bash
  # Dump production
  pg_dump production > prod_dump.sql
  
  # Restore to test DB
  psql test_db < prod_dump.sql
  
  # Test migration
  make migrate
  ```

- **Use meaningful names**
  ```
  000001_create_users_table.up.sql           Clear
  000002_add_email_verification.up.sql       Descriptive
  000001_changes.up.sql                      Vague
  ```

### DON'T

- **Don't modify applied migrations** - Create new migration instead
  ```
   Edit: 000001_create_users.up.sql
   Create: 000002_modify_users_table.up.sql
  ```

- **Don't skip version numbers** - Keep sequential order
  ```
  000001_create_users.up.sql        
  000002_add_contacts.up.sql        
  000005_add_profiles.up.sql         Skipped 3 and 4
  ```

- **Don't use DROP DATABASE/DROP SCHEMA** - Too dangerous
  ```sql
  -- NEVER in migrations
  DROP DATABASE promenade;
  DROP SCHEMA public;
  ```

- **Don't forget indexes** - Create indexes in same migration as table
  ```sql
  CREATE TABLE users (...);
  
  -- Always add relevant indexes immediately
  CREATE INDEX idx_users_email ON users(email);
  CREATE INDEX idx_users_created_at ON users(created_at);
  ```

---

## Namespace Execution Order

Migrations must run in specific order to respect dependencies:

```
1. core/           # Core infrastructure (extensions, functions)
2. shared/         # Shared kernel (reference data)
3. identity/       # Identity context (users, auth)
4. customer-mgmt/  # Customer Management (depends on identity)
5. order-mgmt/     # Order Management (depends on customer)
6. billing/        # Billing (depends on order)
```

**Implementation** (`cmd/api/main.go`):

```go
// Run core migrations (extensions: uuid_v7, pgcrypto)
if err := manager.MigrateNamespace(ctx, "core"); err != nil {
    log.Fatal("Core migrations failed:", err)
}

// Run shared kernel migrations (countries, currencies, etc.)
if err := manager.MigrateNamespace(ctx, "shared"); err != nil {
    log.Fatal("Shared migrations failed:", err)
}

// Run identity context migrations (users, contacts)
if err := manager.MigrateNamespace(ctx, "identity"); err != nil {
    log.Fatal("Identity migrations failed:", err)
}

// Run customer management context migrations
if err := manager.MigrateNamespace(ctx, "customer-mgmt"); err != nil {
    log.Fatal("Customer-mgmt migrations failed:", err)
}
```

---

## Schema Migrations Table

The manager creates a `schema_migrations` table to track applied migrations:

```sql
CREATE TABLE schema_migrations (
    version     BIGINT       NOT NULL,
    namespace   VARCHAR(50)  NOT NULL,
    dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, version)
);
```

**Example data:**

| namespace      | version | dirty | applied_at          |
|----------------|---------|-------|---------------------|
| core           | 1       | false | 2025-12-28 10:00:00 |
| core           | 2       | false | 2025-12-28 10:00:01 |
| shared         | 1       | false | 2025-12-28 10:00:02 |
| identity       | 1       | false | 2025-12-28 10:00:03 |
| identity       | 2       | false | 2025-12-28 10:00:04 |
| customer-mgmt  | 1       | false | 2025-12-28 10:00:05 |

---

## Error Handling

### Dirty State

If a migration fails mid-execution, the namespace is marked as "dirty":

```go
status, _ := manager.Status(ctx)
if status["identity"].Dirty {
    fmt.Println("Identity namespace is dirty - manual intervention required")
    // 1. Inspect database state
    // 2. Fix any partial changes
    // 3. Mark as clean or rollback
}
```

**Recovery:**

```sql
-- Option 1: Mark as clean if manually fixed
UPDATE schema_migrations 
SET dirty = false 
WHERE namespace = 'identity' AND version = 5;

-- Option 2: Delete failed version and retry
DELETE FROM schema_migrations 
WHERE namespace = 'identity' AND version = 5;
```

---

## Testing

Repository-wide testing strategy and baseline budgets are documented in [docs/guides/testing-patterns.md](../../docs/guides/testing-patterns.md).

```bash
# Run migration tests
go test ./pkg/migration -v

# Test with real database
go test ./test/integration -v
```

---

## Related Packages

- `pkg/logger` - Structured logging (used for migration progress)
- `pkg/uuidv7` - UUID v7 generator (used in migrations)
- `internal/infrastructure/database` - Database connection management

---

**Last Updated:** 2025-12-28  
**Status:** Production-ready  
**Maintainer:** Promenade Team
