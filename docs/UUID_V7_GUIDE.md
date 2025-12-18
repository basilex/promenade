# UUID v7 Implementation Guide

## Overview

This project has been migrated from UUID v4 to UUID v7 for primary key generation. UUID v7 provides significant performance benefits while maintaining the global uniqueness and distributed generation capabilities of UUID v4.

## Why UUID v7?

### Benefits over UUID v4

1. **Time-ordered**: UUIDs are naturally sorted by creation time
2. **Better B-tree performance**: Reduces page splits in PostgreSQL indexes (20-50% faster INSERTs in benchmarks)
3. **Reduced fragmentation**: Sequential-ish IDs minimize index fragmentation
4. **Cache-friendly**: Better locality of reference for recently created records
5. **Extractable timestamp**: Can retrieve creation time from the UUID itself
6. **Still globally unique**: Maintains UUID v4's uniqueness guarantees

### Comparison with other strategies

| Strategy         | Pros                                          | Cons                                           |
| ---------------- | --------------------------------------------- | ---------------------------------------------- |
| **UUID v7**      | [+] Time-ordered, globally unique, distributed | Slightly larger than BIGINT (16 bytes)         |
| UUID v4          | Globally unique, distributed                  | [X] Random = poor index locality                |
| SERIAL/BIGSERIAL | Small, sequential, fast                       | [X] Single point of failure, replication issues |
| ULID             | Similar to v7, base32 encoded                 | Less standard, limited library support         |
| Snowflake ID     | Fast, time-ordered                            | Requires coordination service                  |

## Implementation

### PostgreSQL Setup

The project includes two migrations:

1. **000005_add_uuidv7_support.sql**: Adds the `uuid_generate_v7()` function to PostgreSQL
2. **000006_switch_tables_to_uuidv7.sql**: Updates table defaults to use v7

```sql
-- Generate a UUID v7 in PostgreSQL
SELECT uuid_generate_v7();
-- Example: 018d2f07-0f3c-7000-8000-123456789abc
```

### Go Code

Use the `pkg/uuidv7` package:

```go
import "github.com/basilex/promenade/pkg/uuidv7"

// Generate a new UUID v7
id := uuidv7.New()

// Generate with specific timestamp (useful for testing)
id := uuidv7.NewWithTime(time.Now())

// Extract creation timestamp from UUID
timestamp := uuidv7.ExtractTime(id)

// Check if UUID is version 7
if uuidv7.IsV7(id) {
    // ...
}
```

### Migration Path

Existing records with UUID v4 will continue to work. The system supports mixed UUIDs:

```
┌─────────────────────────────────────┬──────────────┐
│ Old records: UUID v4 (random)       │ Pre-migration│
│ New records: UUID v7 (time-ordered) │ Post-migration│
└─────────────────────────────────────┴──────────────┘
```

## Performance Considerations

### When to Use UUID v7

[+] **Good fit:**

- Distributed systems where multiple nodes generate IDs
- Multi-region deployments
- Microservices architecture
- APIs where clients need to generate IDs
- Tables with high INSERT rates
- When you need natural time-based ordering

[X] **Reconsider if:**

- Single-database system with no distribution needs
- Storage is extremely constrained (UUID = 16 bytes vs BIGINT = 8 bytes)
- You need sequential gaps analysis (use SERIAL instead)

### PostgreSQL Index Optimization

UUID v7 works best with B-tree indexes due to its time-ordered nature:

```sql
-- Good: Primary key automatically uses B-tree
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7()
);

-- Good: Time-based queries benefit from UUID v7 ordering
CREATE INDEX idx_users_created ON users(created_at DESC);

-- Optimal: Combine UUID v7 with timestamp for range queries
SELECT * FROM users
WHERE id >= uuid_generate_v7_from_timestamp('2024-01-01'::timestamp)
ORDER BY id DESC;
```

### Monitoring Performance

Compare INSERT performance before/after migration:

```sql
-- Check index bloat
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE tablename IN ('users', 'products', 'roles')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Monitor INSERT performance
EXPLAIN ANALYZE
INSERT INTO users (email, name, password)
VALUES ('test@example.com', 'Test', 'hash');
```

## Recommendations for Foreign Keys

### Use UUID v7 for Foreign Keys Too

```go
type Product struct {
    ID        uuid.UUID `db:"id"`        // UUID v7 primary key
    UserID    uuid.UUID `db:"user_id"`   // UUID v7 foreign key
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
}
```

### Indexing Foreign Keys

Always index foreign keys for JOIN performance:

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL
);

-- Critical: Index the foreign key
CREATE INDEX idx_products_user_id ON products(user_id);
```

### Composite Keys with UUID v7

For many-to-many relationships:

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Index for reverse lookup
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

## Alternative Strategies (Future Consideration)

### Option 1: ULID (Universally Unique Lexicographically Sortable Identifier)

Similar to UUID v7 but uses base32 encoding (shorter string representation):

```
01ARZ3NDEKTSV4RRFFQ69G5FAV  (ULID - 26 chars)
018d2f07-0f3c-7000-8000-123456789abc  (UUID v7 - 36 chars)
```

**When to use**: If you frequently display IDs to users and want shorter strings.

### Option 2: Snowflake IDs

Twitter's Snowflake: 64-bit integers with timestamp, datacenter ID, and sequence.

```go
// Example structure (not implemented in this project)
// 41 bits: timestamp
// 10 bits: datacenter/worker ID
// 12 bits: sequence number
```

**When to use**: Extremely high-throughput systems (Twitter scale) where coordination is acceptable.

### Option 3: PostgreSQL SERIAL/BIGSERIAL

Traditional auto-incrementing integers:

```sql
CREATE TABLE simple_table (
    id BIGSERIAL PRIMARY KEY,  -- 1, 2, 3, 4, ...
    name VARCHAR(255)
);
```

**When to use**: Simple single-database applications with no distribution needs.

### Option 4: Composite Natural Keys

Use meaningful business data as keys:

```sql
CREATE TABLE country_codes (
    code CHAR(2) PRIMARY KEY,  -- 'US', 'GB', 'FR'
    name VARCHAR(100)
);
```

**When to use**: When natural keys are stable, short, and truly unique.

## Testing

Run the test suite to verify UUID v7 implementation:

```bash
# Test UUID v7 package
go test -v ./pkg/uuidv7/

# Run with benchmarks
go test -bench=. ./pkg/uuidv7/

# Test integration with database
make test-integration
```

Expected benchmark results (approximate):

```
BenchmarkNew-8      3000000    450 ns/op    (UUID v7)
BenchmarkNewV4-8    2800000    480 ns/op    (UUID v4)
```

## Migration Checklist

- [x] Add `uuid_generate_v7()` function to PostgreSQL
- [x] Update table defaults to use UUID v7
- [x] Update Go code to use `pkg/uuidv7` package
- [x] Update all usecases (auth, product, role)
- [ ] Run migrations: `make migrate-up`
- [ ] Test UUID generation: `go test ./pkg/uuidv7/`
- [ ] Monitor production metrics after deployment
- [ ] Update templates in `scripts/templates/` to use UUID v7 by default

## Further Reading

- [RFC 9562 - UUID Version 7 (Draft)](https://datatracker.ietf.org/doc/draft-ietf-uuidrev-rfc4122bis/)
- [PostgreSQL UUID Performance](https://www.2ndquadrant.com/en/blog/sequential-uuid-generators/)
- [UUID v7 in Production (Blog post)](https://buildkite.com/blog/goodbye-integers-hello-uuids)

## Questions?

See `.github/copilot-instructions.md` for project conventions or ask in team chat.
