# Database Adapters Guide

**Multi-Database Support** for Promenade Platform - PostgreSQL, SQLite, MySQL, SQL Server.

---

## Overview

Promenade implements **database-agnostic architecture** through the **Dialect pattern**, allowing the same Go code to work with multiple SQL databases without vendor lock-in.

**Implementation**: `pkg/database` package with adapters for each database.

---

## Architecture

### Dialect Interface

```go
// Dialect abstracts database-specific SQL syntax
type Dialect interface {
    // Name returns the database name (e.g., "postgres", "sqlite")
    Name() string

    // Placeholder returns SQL placeholder for position n (1-indexed)
    // PostgreSQL: $1, $2, ...
    // SQLite/MySQL: ?, ?, ...
    Placeholder(n int) string

    // ConvertPlaceholders converts ? to database-specific placeholders
    ConvertPlaceholders(sql string) string

    // SupportsJSONB returns true if database has native JSONB type
    SupportsJSONB() bool

    // JSONType returns the column type for JSON storage
    // PostgreSQL: "JSONB"
    // SQLite/MySQL: "TEXT"
    JSONType() string
}
```

**Key Insight**: The Dialect interface abstracts **SQL syntax differences**, not ORM-level abstractions. We still write SQL, just portable SQL.

---

## Supported Databases

### PostgreSQL (Primary)

**Implementation**: `pkg/database/postgres/dialect.go`

**Features**:
- Native JSONB support
- UUID v7 generation (via `uuid_generate_v7()` extension)
- Full-text search (tsvector, tsquery)
- Advanced indexing (GIN, GiST)
- Transactional DDL

**Placeholder Style**: `$1`, `$2`, `$3`

**JSON Storage**: `JSONB` (binary format, indexable)

**Usage**:
```go
import "github.com/basilex/promenade/pkg/database/postgres"

dialect := postgres.NewDialect()
sql := dialect.ConvertPlaceholders("SELECT * FROM users WHERE email = ?")
// Result: "SELECT * FROM users WHERE email = $1"
```

**Recommended For**:
- Production deployments
- Large datasets (>1GB)
- Complex queries with JSONB
- High concurrency
- Full ACID requirements

---

### SQLite (Development/Testing)

**Implementation**: `pkg/database/sqlite/dialect.go`

**Features**:
- Embedded database (no server)
- Zero configuration
- Fast for small datasets (<1GB)
- Full SQL support
- WAL mode for concurrency

**Placeholder Style**: `?`, `?`, `?`

**JSON Storage**: `TEXT` (stored as JSON string)

**Usage**:
```go
import "github.com/basilex/promenade/pkg/database/sqlite"

dialect := sqlite.NewDialect()
sql := dialect.ConvertPlaceholders("SELECT * FROM users WHERE email = ?")
// Result: "SELECT * FROM users WHERE email = ?" (no change)
```

**Recommended For**:
- Local development
- Integration tests
- CI/CD pipelines
- Embedded applications
- Quick prototypes

**Limitations**:
- No concurrent writes (WAL helps but limited)
- No native UUID type (stored as TEXT)
- JSON operations less efficient (no indexing)
- No full-text search extensions

---

### MySQL (Planned - Phase 8)

**Status**: Not yet implemented

**Placeholder Style**: `?`, `?`, `?`

**JSON Storage**: `JSON` (native type since MySQL 5.7)

**Planned Support**:
- MySQL 8.0+
- MariaDB 10.5+

---

### SQL Server (Planned - Phase 9)

**Status**: Not yet implemented

**Placeholder Style**: `@p1`, `@p2`, `@p3`

**JSON Storage**: `NVARCHAR(MAX)` with JSON functions

---

## How It Works

### 1. Query Placeholders

**Problem**: Different databases use different placeholder syntax.

**Solution**: Write queries with `?`, convert at runtime.

**Example**:
```go
// Write SQL with ?
query := "SELECT * FROM users WHERE email = ? AND status = ?"

// Convert to database-specific placeholders
query = dialect.ConvertPlaceholders(query)

// PostgreSQL: "SELECT * FROM users WHERE email = $1 AND status = $2"
// SQLite: "SELECT * FROM users WHERE email = ? AND status = ?"
```

### 2. JSON Storage

**Problem**: PostgreSQL has `JSONB`, others use `TEXT`.

**Solution**: Use `jsonstore.Field[T]` wrapper that adapts to database.

**Example**:
```go
type Customer struct {
    Tags jsonstore.Field[[]string] `db:"tags"`
}

// PostgreSQL migration:
// ALTER TABLE customers ADD COLUMN tags JSONB;

// SQLite migration:
// ALTER TABLE customers ADD COLUMN tags TEXT;

// Go code works identically for both
```

**See**: [JSONB Strategy Guide](jsonb-strategy.md) for details.

### 3. UUID Generation

**Problem**: PostgreSQL has `uuid_generate_v7()`, SQLite doesn't.

**Solution**: Generate UUIDs in Go code, not database defaults.

**Example**:
```go
// ❌ Old (database-specific)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7()
);

// ✅ New (database-agnostic)
CREATE TABLE users (
    id UUID PRIMARY KEY
);

// Generate in Go
user := &User{
    BaseAggregate: aggregate.NewBaseAggregate(), // ID = uuidv7.New()
}
```

---

## Usage Patterns

### Repository Pattern

**BaseRepository** provides database-agnostic operations:

```go
type orderRepository struct {
    *BaseRepository
}

func NewOrderRepository(db *sqlx.DB) order.IRepository {
    return &orderRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

func (r *orderRepository) Create(ctx context.Context, o *order.Order) error {
    // Write query with ?
    query := `
        INSERT INTO order_orders (id, order_number, customer_id, total_amount)
        VALUES (?, ?, ?, ?)
    `

    // BaseRepository handles placeholder conversion
    _, err := r.Exec(ctx, query, o.ID, o.OrderNumber, o.CustomerID, o.Total.Amount)
    return err
}
```

**BaseRepository** methods:
- `Exec(ctx, query, args)` - Execute with placeholder conversion
- `Get(ctx, dest, query, args)` - Fetch single row
- `Select(ctx, dest, query, args)` - Fetch multiple rows
- `NamedExec(ctx, query, arg)` - Named parameter execution

### Transaction Pattern

```go
err := tm.WithTransaction(ctx, func(ctx context.Context) error {
    // All queries use same transaction
    if err := orderRepo.Create(ctx, order); err != nil {
        return err // Auto-rollback
    }
    return customerRepo.Update(ctx, customer) // Auto-commit
})
```

**Context-aware**: Transaction stored in context, retrieved by `getExecutor(ctx)`.

---

## Configuration

### PostgreSQL

**config/app.prod.yaml**:
```yaml
database:
  postgres:
    host: "postgres.prod.example.com"
    port: 5432
    user: "promenade_app"
    password: "${DB_PASSWORD}"
    database: "promenade_prod"
    ssl_mode: "require"
```

### SQLite

**config/app.test.yaml**:
```yaml
database:
  sqlite:
    path: ":memory:" # In-memory for tests
    # OR
    path: "./data/promenade.db" # File-based
```

---

## Migrations

**Namespace-based migrations** ensure correct order:

```
migrations/
  core/                     # Core infrastructure
  shared/                   # Shared context (reference data)
  identity/                 # Identity context
  customer-mgmt/            # Customer Management
  order-mgmt/               # Order Management
```

**Run Order**: `core → shared → identity → customer-mgmt → order-mgmt`

**Database-Agnostic Rules**:
1. ❌ No UUID defaults in CREATE TABLE
2. ❌ No UPDATE triggers for timestamps
3. ❌ No database functions in constraints
4. ✅ Generate UUIDs in Go (`uuidv7.New()`)
5. ✅ Update timestamps in Go (`.Touch()`)
6. ✅ Use TEXT for JSON (SQLite compatible)

**Example Migration**:
```sql
-- PostgreSQL
CREATE TABLE customers (
    id UUID PRIMARY KEY,  -- No DEFAULT
    name VARCHAR(255) NOT NULL,
    tags TEXT,  -- Not JSONB (SQLite compatible)
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- SQLite (same DDL)
CREATE TABLE customers (
    id TEXT PRIMARY KEY,  -- UUID as TEXT
    name TEXT NOT NULL,
    tags TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

---

## Testing

### Integration Tests

**Run with PostgreSQL**:
```bash
make test-integration  # Uses PostgreSQL on :5433
```

**Run with SQLite** (planned):
```bash
ENVIRONMENT=test DB_ADAPTER=sqlite make test-integration
```

### Benchmark Tests

**Current Results** (PostgreSQL, Apple M4 Max):
- User ListUsers: ~640-750μs
- Interaction Create: ~413μs
- Interaction List: ~904μs

---

## Performance Characteristics

### PostgreSQL vs SQLite

| Feature               | PostgreSQL         | SQLite             |
| --------------------- | ------------------ | ------------------ |
| **Concurrency**       | Excellent          | Limited (WAL mode) |
| **Dataset Size**      | >1GB optimal       | <1GB optimal       |
| **JSON Performance**  | Excellent (JSONB)  | Good (TEXT)        |
| **UUID Storage**      | Native (16 bytes)  | TEXT (36 bytes)    |
| **Full-Text Search**  | Built-in (tsvector)| Extension required |
| **Setup Complexity**  | High (server)      | Zero (embedded)    |

### Recommendations

- **Production**: PostgreSQL (scalability, concurrency)
- **Development**: SQLite (fast setup, zero config)
- **CI/CD**: SQLite (parallel tests, no server)
- **Embedded Apps**: SQLite (single binary)

---

## Troubleshooting

### Issue: Placeholder Mismatch

**Symptom**: `pq: syntax error near "$1"`

**Cause**: Query not converted via `dialect.ConvertPlaceholders()`

**Fix**: Use `BaseRepository` methods (Exec, Get, Select) instead of raw sqlx.

### Issue: JSONB Type Error

**Symptom**: `ERROR: type "jsonb" does not exist`

**Cause**: Using JSONB in SQLite migration

**Fix**: Use `TEXT` for JSON columns (works in both databases).

### Issue: UUID Type Mismatch

**Symptom**: `invalid input syntax for type uuid`

**Cause**: SQLite stores UUID as TEXT, PostgreSQL as UUID

**Fix**: Use `uuidv7.UUID` type in Go, let driver handle conversion.

---

## Roadmap

### Phase 8: MySQL Support (Q1 2026)

- [ ] MySQL 8.0+ dialect
- [ ] MariaDB 10.5+ testing
- [ ] JSON type support
- [ ] Migration compatibility

### Phase 9: SQL Server Support (Q2 2026)

- [ ] SQL Server 2019+ dialect
- [ ] Azure SQL Database testing
- [ ] NVARCHAR(MAX) JSON support
- [ ] Unique placeholder syntax (@p1)

### Future Considerations

- **CockroachDB**: Postgres-compatible, add retry logic
- **TiDB**: MySQL-compatible, test compatibility
- **YugabyteDB**: Postgres-compatible, test compatibility
- **MongoDB**: Deferred (NoSQL, different architecture)

---

## Related Documentation

- [Main README](../../README.md)
- [JSONB Strategy Guide](jsonb-strategy.md)
- [Database Conventions Guide](database-conventions.md)
- [Migration System](../../migrations/README.md)
- [Testing Guide](../../test/README.md)

---

**Last Updated**: January 3, 2026  
**Status**: PostgreSQL Production-Ready, SQLite Planned Phase 8  
**Test Coverage**: 101 tests (pkg/database + pkg/jsonstore), 100% passing  
**Maintainer**: Promenade Team
