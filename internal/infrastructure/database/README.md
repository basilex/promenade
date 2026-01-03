**Navigation**: [Home](../../../README.md) > [Internal](../../README.md) > [Infrastructure](../README.md) > Database

---

# Database Package

**PostgreSQL connection and transaction management** for Promenade Platform.

---

## Overview

The `database` package provides:

- **PostgreSQL Connection**: Connection pooling with sqlx
- **Transaction Management**: Context-aware transactions
- **BaseRepository Pattern**: Shared repository functionality
- **Migration Support**: Namespace-based migrations

---

## Quick Start

### Connect to Database

```go
import (
    "github.com/basilex/promenade/internal/infrastructure/database"
    "github.com/basilex/promenade/internal/infrastructure/config"
)

// Load configuration
cfg, _ := config.Load()

// Connect to PostgreSQL
db, err := database.NewPostgresConnection(&cfg.Database.Postgres)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}
defer db.Close()

// Database is ready
fmt.Println("Connected to PostgreSQL")
```

---

## Features

### Connection Pooling

**sqlx** provides enhanced database functionality:

```go
// Configuration
type PostgresConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
    SSLMode  string
}

// Connection pool settings (default)
MaxOpenConns: 25
MaxIdleConns: 5
ConnMaxLifetime: 5 minutes
```

### Transaction Management

**TransactionManager** provides context-aware transactions:

```go
import "github.com/basilex/promenade/internal/infrastructure/database"

tm := database.NewTransactionManager(db)

// Execute in transaction
err := tm.WithTransaction(ctx, func(ctx context.Context) error {
    // All repository calls within this function use the same transaction
    if err := userRepo.Create(ctx, user); err != nil {
        return err // Auto-rollback
    }
    
    if err := profileRepo.Create(ctx, profile); err != nil {
        return err // Auto-rollback
    }
    
    return nil // Auto-commit if no error
})
```

**Key features**:
- Automatic commit on success
- Automatic rollback on error
- Context propagation (transaction stored in context)
- Nested transaction detection

### BaseRepository Pattern

**BaseRepository** provides common database operations for all repositories:

```go
// Each context has its own BaseRepository
type BaseRepository struct {
    db *sqlx.DB
}

// Common methods
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) error
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) error
func (r *BaseRepository) getExecutor(ctx context.Context) database.Executor
```

**Repository implementation**:

```go
// Identity context repository
type contactRepository struct {
    *BaseRepository
}

func NewContactRepository(db *sqlx.DB) IRepository {
    return &contactRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

func (r *contactRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error) {
    var row contactRow
    query := `SELECT * FROM identity_contacts WHERE id = $1 AND deleted_at IS NULL`
    
    // getExecutor automatically uses transaction if present in context
    if err := r.Get(ctx, &row, query, id); err != nil {
        return nil, err
    }
    
    return row.toEntity()
}
```

---

## Database Schema

### Namespace-Based Tables

Each context owns its tables with namespace prefix:

**Shared Context**: `shared_countries`, `shared_currencies`, `shared_languages`, `shared_timezones`  
**Identity Context**: `identity_users`, `identity_contacts`, `identity_profiles`, `identity_roles`, `identity_permissions`  
**Customer Management**: `customer_customers`, `customer_companies`, `customer_deals`

**Benefits**:
- Clear ownership
- No table name conflicts
- Easy to identify context boundaries

### Migration System

**Namespace-based migrations** with separate directories per context:

```
migrations/
 core/              # Core infrastructure (UUID v7, extensions)
 shared/            # Shared context tables
 identity/          # Identity context tables
 customer-mgmt/     # Customer Management tables
```

**Run migrations**:

```bash
make migrate                  # All namespaces
make migrate-identity         # Identity context only
make migrate-customer-mgmt    # Customer Management only
```

**See**: [migrations/README.md](../../../migrations/README.md) for details

---

## Best Practices

### DO ✅

- **Always pass context** to repository methods
- **Use transactions** for multi-table operations
- **Filter soft-deleted records** (`WHERE deleted_at IS NULL`)
- **Use UUID v7** for primary keys (`uuidv7.New()`)
- **Use BaseRepository** for common operations
- **Close connections** with `defer db.Close()`

### DON'T ❌

- **DON'T use uuid.New()** (use uuidv7.New() instead)
- **DON'T forget soft delete filter** in SELECT queries
- **DON'T create multiple connections** (reuse db instance)
- **DON'T ignore context cancellation**
- **DON'T manually commit/rollback** (TransactionManager handles it)

---

## Examples

### Repository with Transaction

```go
// Use case with transaction
func (uc *UserUseCase) Register(ctx context.Context, email, name, password string) (*User, error) {
    return uc.tm.WithTransaction(ctx, func(ctx context.Context) (*User, error) {
        // Create user
        user, err := entity.NewUser(email, name, password)
        if err != nil {
            return nil, err
        }
        
        if err := uc.userRepo.Create(ctx, user); err != nil {
            return nil, err
        }
        
        // Create default profile (same transaction)
        profile, _ := profile.NewProfile(user.ID, name)
        if err := uc.profileRepo.Create(ctx, profile); err != nil {
            return nil, err // Rollback user creation
        }
        
        return user, nil // Commit both
    })
}
```

### Soft Delete Query

```go
func (r *contactRepository) GetActiveContacts(ctx context.Context, userID uuidv7.UUID) ([]*Contact, error) {
    var rows []contactRow
    
    // Always filter soft-deleted records
    query := `
        SELECT * FROM identity_contacts 
        WHERE user_id = $1 
          AND deleted_at IS NULL
        ORDER BY created_at DESC
    `
    
    if err := r.Select(ctx, &rows, query, userID); err != nil {
        return nil, err
    }
    
    return rowsToEntities(rows)
}
```

---

## Configuration

### PostgreSQL Config

```yaml
# config/app.postgres-dev.yaml or config/app.sqlite-dev.yaml
database:
  postgres:
    host: "localhost"
    port: 5432
    user: "system"
    password: "passw0rd"
    database: "promenade_dev"
    ssl_mode: "disable"
```

### Environment Variables

```bash
# Production overrides
export DB_HOST="production-db.example.com"
export DB_USER="promenade_app"
export DB_PASSWORD="secure_password"
export DB_NAME="promenade_prod"
```

---

## Files

```
internal/infrastructure/database/
 postgres.go          # PostgreSQL connection
 transaction.go       # Transaction management
 base_repository.go   # BaseRepository pattern (per context)
```

---

## Related Documentation

- [Main README](../../../README.md)
- [Documentation Index](../../../docs/INDEX.md)
- [Configuration](../config/README.md)
- [Migrations](../../../migrations/README.md)
- [Identity Context](../../contexts/identity/README.md)

---

**Status**: Production-ready  
**Database**: PostgreSQL 16  
**Driver**: sqlx  
**Maintainer**: Promenade Team
