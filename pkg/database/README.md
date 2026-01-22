# Database Package

**Database adapter abstraction** for PostgreSQL 14+.

---

## Overview

The `database` package provides a **dialect interface** for PostgreSQL database abstraction in Promenade Platform.

**Driver:** We use **pgx/v5/stdlib** instead of lib/pq for:

- 20-30% performance improvement
- Better PostgreSQL feature support
- Active development (lib/pq is in maintenance mode)
- Full compatibility with database/sql and sqlx
- Future path to native pgx API (50-60% gains when needed)

---

## Features

- **Dialect Abstraction**: PostgreSQL placeholders ($1, $2, ...)
- **Type Mapping**: UUID, JSON, TIMESTAMP native types
- **Feature Detection**: RETURNING, JSON indexes, native UUID support
- **SQL Safety**: Identifier validation, quote escaping, injection prevention
- **Placeholder Conversion**: Numbered placeholders for queries

---

## Supported Databases

PostgreSQL 14+ is the only supported database, providing production-grade features including native UUID, JSONB, and GIN indexes.

---

## Quick Start

### 1. Create Dialect

```go
import (
    "github.com/basilex/promenade/pkg/database/postgres"
)

// PostgreSQL (production-ready)
pgDialect := postgres.NewDialect()
```

### 2. Use Dialect for Queries

```go
// Build query with dialect
query := fmt.Sprintf(
    "INSERT INTO users (id, email) VALUES (%s, %s)",
    dialect.Placeholder(1),
    dialect.Placeholder(2),
)

// PostgreSQL: INSERT INTO users (id, email) VALUES ($1, $2)
```

### 3. Convert Existing Queries

```go
// Write query in Postgres style (standard)
query := "SELECT * FROM users WHERE id = $1 AND email = $2"

// Convert to target database when needed
converted := database.ConvertPlaceholders(query, targetDialect)
// Result depends on target dialect placeholder style
```

---

## Dialect Methods

### Placeholder

Returns database-specific parameter placeholder:

```go
// PostgreSQL
pgDialect.Placeholder(1)  // "$1"
pgDialect.Placeholder(2)  // "$2"
```

### Type Mapping

```go
// UUID storage
pgDialect.UUIDType()       // "UUID"

// JSON storage
pgDialect.JSONType()       // "JSONB"

// Timestamp storage
pgDialect.TimestampType()  // "TIMESTAMP"

// Boolean storage
pgDialect.BoolType()       // "BOOLEAN"
```

### Feature Detection

```go
// Check feature support
if dialect.SupportsReturning() {
    query += " RETURNING id"
}

if dialect.SupportsJSON() {
    // Use native JSON type
    columnType = dialect.JSONType()
} else {
    // Fallback to TEXT
    columnType = "TEXT"
}

if dialect.SupportsJSONIndex() {
    // Create GIN index (Postgres)
    createIndex = "CREATE INDEX idx_tags USING gin(tags)"
} else {
    // Use standard index or none
}
```

### Identifier Safety

```go
// Quote identifiers to prevent SQL injection
tableName := dialect.QuoteIdentifier("users")
// PostgreSQL: "users"

// Validate identifier
if err := database.ValidateIdentifier(userInput); err != nil {
    return fmt.Errorf("invalid table name: %w", err)
}
```

---

## PostgreSQL Dialect

### JSON Operations

```go
pgDialect := postgres.NewDialect()

// Check if JSONB contains value
where := pgDialect.JSONContains("tags", `["vip"]`)
// Result: tags @> '["vip"]'::jsonb

// Extract value from JSONB
select := pgDialect.JSONExtract("metadata", "score")
// Result: metadata->>'score'

// Create GIN index
index := pgDialect.CreateJSONIndex("customers", "tags", "idx_tags")
// Result: CREATE INDEX "idx_tags" ON "customers" USING gin("tags")
```

---

### Repository with Dialect

```go
type CustomerRepository struct {
    db      *sqlx.DB
    dialect database.Dialect
}

func NewCustomerRepository(db *sqlx.DB, dialect database.Dialect) *CustomerRepository {
    return &CustomerRepository{
        db:      db,
        dialect: dialect,
    }
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
    query := fmt.Sprintf(
        "SELECT * FROM customers WHERE id = %s",
        r.dialect.Placeholder(1),
    )

    var customer Customer
    err := r.db.GetContext(ctx, &customer, query, id)
    return &customer, err
}
```

### Database-Specific Optimizations

```go
func (r *CustomerRepository) GetByTag(ctx context.Context, tag string) ([]*Customer, error) {
    // Use database-specific optimization when available
    if r.dialect.SupportsJSON() && r.dialect.Name() == "postgres" {
        // PostgreSQL: Use @> operator (fast GIN index)
        pgDialect := r.dialect.(*postgres.Dialect)
        where := pgDialect.JSONContains("tags", fmt.Sprintf(`["%s"]`, tag))
        query := fmt.Sprintf("SELECT * FROM customers WHERE %s", where)

        var customers []*Customer
        err := r.db.SelectContext(ctx, &customers, query)
        return customers, err
    }

    // Fallback: Fetch all and filter in Go (slower but works)
    query := "SELECT * FROM customers"
    var customers []*Customer
    err := r.db.SelectContext(ctx, &customers, query)
    if err != nil {
        return nil, err
    }

    // Filter in memory
    filtered := []*Customer{}
    for _, c := range customers {
        if c.HasTag(tag) {
            filtered = append(filtered, c)
        }
    }

    return filtered, nil
}
```

---

## Migration Example

### Before (PostgreSQL-only)

```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tags JSONB DEFAULT '[]'::jsonb
);

CREATE INDEX idx_tags ON customers USING gin(tags);
```

```go
type Customer struct {
    ID   uuid.UUID `db:"id"`
    Tags []string  `db:"tags"` // Relies on Postgres JSONB
}
```

### After (Database-Agnostic)

```sql
-- PostgreSQL
CREATE TABLE customers (
    id UUID PRIMARY KEY,  -- No DEFAULT (generated in Go)
    tags JSONB DEFAULT '[]'
);
```

```go
import "github.com/basilex/promenade/pkg/jsonstore"

type Customer struct {
    ID   uuid.UUID                `db:"id"`
    Tags jsonstore.Field[[]string] `db:"tags"` // Works with any DB
}

func NewCustomer(name string) *Customer {
    return &Customer{
        ID:   uuidv7.New(),  // Generate in Go
        Tags: jsonstore.NewField([]string{}),
    }
}
```

---

## Testing

Repository-wide testing strategy and baseline budgets are documented in [docs/guides/testing-patterns.md](../../docs/guides/testing-patterns.md).

```bash
# Run dialect tests
go test ./pkg/database -v

# Test PostgreSQL dialect
go test ./pkg/database/postgres -v

# Benchmark placeholder conversion
go test -bench=. ./pkg/database
```

---

## Best Practices

### DO

- Use `dialect.Placeholder(n)` for all parameterized queries
- Check feature support with `SupportsXxx()` methods
- Quote identifiers with `QuoteIdentifier()` when using user input
- Validate identifiers with `ValidateIdentifier()` before use
- Use database-specific optimizations when available (fallback for others)

### DON'T

- Don't hardcode $1, $2 placeholders (use `Placeholder()`)
- Don't assume all databases support RETURNING
- Don't rely on native JSON type (use `jsonstore` package)
- Don't skip identifier validation for user input
- Don't write database-specific code without feature detection

---

## Future Enhancements

- [ ] Query builder for complex queries
- [ ] Migration generator from dialect
- [ ] Performance benchmarks
- [ ] Multi-database testing

---

Query builder for complex queries

- [ ] Migration generator from dialect
- [ ] Performance benchmarks Multi-database testing
- [Database Adapters Guide](../../docs/guides/database-adapters.md) - Multi-database support strategy
- [JSONB Strategy Guide](../../docs/guides/jsonb-strategy.md) - Cross-database JSON handling

---

**Status**: Production Ready  
**Version**: 1.0.0  
**Maintainer**: Promenade Team  
**Last Updated**: January 3, 2026
