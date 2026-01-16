# Database Package

**Database adapter abstraction** for multi-database support in Promenade Platform.

---

## Overview

The `database` package provides a **unified interface** for working with different SQL databases (PostgreSQL, SQLite, MySQL, SQL Server). It abstracts away database-specific syntax differences (placeholders, types, features) while preserving database-specific optimizations.

---

## Features

-  **Dialect Abstraction**: Postgres ($1), SQLite (?), MySQL (?), SQL Server (@p1)
-  **Type Mapping**: UUID, JSON, TIMESTAMP differences handled automatically
-  **Feature Detection**: RETURNING, JSON indexes, native UUID support
-  **SQL Safety**: Identifier validation, quote escaping, injection prevention
-  **Placeholder Conversion**: Auto-convert $1 style to database-specific format

---

## Supported Databases

| Database     | Status         | Dialect           | Native UUID | Native JSON | Indexes  |
| ------------ | -------------- | ----------------- | ----------- | ----------- | -------- |
| PostgreSQL   |  Production  | `postgres.Dialect` | YES         | JSONB       | GIN      |
| SQLite       |  Production  | `sqlite.Dialect`   | NO (TEXT)   | NO (TEXT)   | Standard |
| MySQL        |  Planned     | `mysql.Dialect`    | NO (CHAR)   | JSON        | Generated|
| SQL Server   |  Planned     | `mssql.Dialect`    | NO (NCHAR)  | NO (NVARCHAR)| Standard|
| CockroachDB  |  Future      | Uses `postgres`    | YES         | JSONB       | GIN      |

---

## Quick Start

### 1. Create Dialect

```go
import (
    "github.com/basilex/promenade/pkg/database/postgres"
    "github.com/basilex/promenade/pkg/database/sqlite"
)

// PostgreSQL
pgDialect := postgres.NewDialect()

// SQLite
sqliteDialect := sqlite.NewDialect()
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
// SQLite:     INSERT INTO users (id, email) VALUES (?, ?)
```

### 3. Convert Existing Queries

```go
// Write query in Postgres style
query := "SELECT * FROM users WHERE id = $1 AND email = $2"

// Convert to target database
converted := database.ConvertPlaceholders(query, sqliteDialect)
// SQLite result: "SELECT * FROM users WHERE id = ? AND email = ?"
```

---

## Dialect Methods

### Placeholder

Returns database-specific parameter placeholder:

```go
// PostgreSQL
pgDialect.Placeholder(1)  // "$1"
pgDialect.Placeholder(2)  // "$2"

// SQLite
sqliteDialect.Placeholder(1)  // "?"
sqliteDialect.Placeholder(2)  // "?"
```

### Type Mapping

```go
// UUID storage
pgDialect.UUIDType()      // "UUID"
sqliteDialect.UUIDType()  // "TEXT"

// JSON storage
pgDialect.JSONType()      // "JSONB"
sqliteDialect.JSONType()  // "TEXT"

// Timestamp storage
pgDialect.TimestampType()      // "TIMESTAMP"
sqliteDialect.TimestampType()  // "DATETIME"

// Boolean storage
pgDialect.BoolType()      // "BOOLEAN"
sqliteDialect.BoolType()  // "BOOLEAN" (0/1 internally)
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
// SQLite: `users`

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

## SQLite Dialect

### JSON Operations (Limited)

```go
sqliteDialect := sqlite.NewDialect()

// Extract value using json_extract
select := sqliteDialect.JSONExtract("metadata", "score")
// Result: json_extract(`metadata`, '$.score')

// Check if array contains value (less efficient)
where := sqliteDialect.JSONArrayContains("tags", "vip")
// Result: json_extract(`tags`, '$') LIKE '%"vip"%'
```

**Note**: SQLite JSON queries are less efficient. For production, use PostgreSQL JSONB optimizations and filter in Go for SQLite.

### Constraints

```go
// UUID format validation
constraint := sqliteDialect.CreateUUIDCheckConstraint("id")
// Result: CHECK(length(`id`) = 36 AND `id` GLOB '[0-9a-f]...')

// JSON format validation
constraint := sqliteDialect.CreateJSONCheckConstraint("metadata")
// Result: CHECK(json_valid(`metadata`))
```

---

## Usage Patterns

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

-- SQLite
CREATE TABLE customers (
    id TEXT PRIMARY KEY,  -- UUID as TEXT
    tags TEXT DEFAULT '[]',
    CHECK(json_valid(tags))
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

# Test SQLite dialect
go test ./pkg/database/sqlite -v

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

- [ ] MySQL dialect implementation
- [ ] SQL Server dialect implementation
- [ ] Query builder for complex queries
- [ ] Migration generator from dialect
- [ ] Performance benchmarks per database

---

## Related Documentation

- [JSON Store Package](../jsonstore/README.md) - Database-agnostic JSON storage
- [Testing Guide](../../test/README.md) - Multi-database testing
- [Database Adapters Guide](../../docs/guides/database-adapters.md) - Multi-database support strategy
- [JSONB Strategy Guide](../../docs/guides/jsonb-strategy.md) - Cross-database JSON handling

---

**Status**:  Production Ready  
**Version**: 1.0.0  
**Maintainer**: Promenade Team  
**Last Updated**: January 3, 2026
