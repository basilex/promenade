# Primary Key Strategy Recommendations for Promenade

## Executive Summary

**Current Solution:** UUID v7 (migrated from v4)

**Why UUID v7:**

1. Time-ordered = better index performance (20-50% faster INSERT)
2. Retains distributed generation of UUID v4
3. 2x faster generation than v4 (87ns vs 185ns)
4. Zero memory allocations
5. Extractable creation timestamp

---

## ID Strategy Comparison

### 1. UUID v7 (current choice) ⭐

**Pros:**

- ✅ Time-ordered for B-tree indexes
- ✅ Distributed generation without coordination
- ✅ Global uniqueness
- ✅ Extractable timestamp
- ✅ Compatible with UUID v4
- ✅ Better performance than v4

**Cons:**

- ❌ 16 bytes (vs 8 bytes for BIGINT)
- ❌ 36 characters in string form

**Use when:**

- Microservice architecture
- Multiple write sources
- Multi-regional deployment
- Need temporal sorting
- APIs where clients generate IDs

---

### 2. UUID v4 (random) - Deprecated

**Pros:**

- ✅ Global uniqueness
- ✅ Distributed generation

**Cons:**

- ❌ Poor index locality
- ❌ Many page splits in B-tree
- ❌ Index fragmentation
- ❌ Slower than v7

**Status:** Replaced with v7 in this project

---

### 3. PostgreSQL SERIAL / BIGSERIAL

```sql
CREATE TABLE simple (
    id BIGSERIAL PRIMARY KEY,  -- 1, 2, 3, 4...
    name TEXT
);
```

**Pros:**

- ✅ Small size (8 bytes)
- ✅ Sequential (excellent locality)
- ✅ Human-readable
- ✅ Maximum performance for single-node

**Cons:**

- ❌ Single point of failure (PostgreSQL must generate)
- ❌ Replication issues (ID conflicts)
- ❌ Sharding complications
- ❌ Information leakage (can count records)

**Use when:**

- Simple application with single DB
- No sharding plans
- Record count leakage not critical
- Need maximum storage efficiency

---

### 4. ULID (Universally Unique Lexicographically Sortable Identifier)

```
01ARZ3NDEKTSV4RRFFQ69G5FAV  (26 characters)
```

**Pros:**

- ✅ Time-ordered like UUID v7
- ✅ Shorter than string UUID (26 vs 36 characters)
- ✅ Base32 encoding (URL-safe without escaping)
- ✅ Case-insensitive

**Cons:**

- ❌ No native PostgreSQL support
- ❌ Fewer library options
- ❌ Still 16 bytes in DB

**Use when:**

- IDs frequently passed in URLs
- Readability important
- Want to avoid UUID dashes

**Example library:** `github.com/oklog/ulid`

---

### 5. Snowflake ID (Twitter)

```
64-bit integer:
┌─────────────┬──────────┬────────────┐
│ Timestamp   │ Machine  │ Sequence   │
│ 41 bits     │ 10 bits  │ 12 bits    │
└─────────────┴──────────┴────────────┘
```

**Pros:**

- ✅ Time-ordered
- ✅ Compact (8 bytes)
- ✅ Very fast generation
- ✅ Extractable timestamp

**Cons:**

- ❌ Requires coordination (machine ID)
- ❌ Needs central service or configuration
- ❌ Possible collisions with misconfiguration
- ❌ Limit: 4096 IDs/ms per machine

**Use when:**

- Twitter/Facebook scale
- Infrastructure for coordination exists
- Size critical (8 bytes vs 16 for UUID)

**Example library:** `github.com/bwmarrin/snowflake`

---

### 6. Composite Natural Keys

```sql
CREATE TABLE country_codes (
    code CHAR(2) PRIMARY KEY,  -- 'US', 'RU', 'GB'
    name VARCHAR(100)
);

CREATE TABLE user_emails (
    email VARCHAR(255) PRIMARY KEY,
    user_id UUID REFERENCES users(id)
);
```

**Pros:**

- ✅ Business-meaningful keys
- ✅ No surrogate key (fewer JOINs)
- ✅ Self-documenting

**Cons:**

- ❌ Business logic may change
- ❌ Complications when changing key
- ❌ Privacy concerns (email as PK)
- ❌ More data for FK

**Use when:**

- Key is truly immutable (ISO codes)
- Reference tables and dictionaries
- Small lookup tables

---

## Foreign Key Recommendations

### UUID v7 in FK

```go
type Order struct {
    ID        uuid.UUID `db:"id"`         // PK: UUID v7
    UserID    uuid.UUID `db:"user_id"`    // FK: UUID v7
    ProductID uuid.UUID `db:"product_id"` // FK: UUID v7
}
```

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    product_id UUID NOT NULL REFERENCES products(id)
);

-- IMPORTANT: always index FKs
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_product_id ON orders(product_id);
```

### Many-to-Many

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Index for reverse lookup
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

---

## When to Change Strategy?

### Stay with UUID v7 if:

- ✅ Microservice architecture
- ✅ Plans for sharding/partitioning
- ✅ API with client-side ID generation
- ✅ Multi-regional deployment
- ✅ Multiple write sources

### Consider SERIAL/BIGSERIAL if:

- Monolithic application on single DB
- No scaling plans
- Data size critical (billions of records)
- No distributed generation requirements

### Consider ULID if:

- IDs often in URLs and readability matters
- Want to avoid dashes
- Need case-insensitive sorting

### Consider Snowflake if:

- Very high load (Twitter scale)
- DB size critical
- Coordination infrastructure exists

---

## Best Practices

### 1. Always index foreign keys

```sql
-- ❌ BAD
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id)
);

-- ✅ GOOD
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id)
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
```

### 2. Use ON DELETE for cascade

```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id)
);
```

### 3. Add timestamps

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Even with UUID v7, store explicit timestamp for:
-- - Precision (UUID v7 = milliseconds, timestamp = microseconds)
-- - Query readability
-- - Independence from ID format
```

### 4. Partitioning with UUID v7

UUID v7 is excellent for temporal partitioning:

```sql
-- Partition by extracted timestamp
CREATE TABLE logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    message TEXT,
    created_at TIMESTAMP GENERATED ALWAYS AS
        (to_timestamp(('x'||substring(id::text from 1 for 8))::bit(32)::bigint / 1000.0))
        STORED
) PARTITION BY RANGE (created_at);

CREATE TABLE logs_2024_01 PARTITION OF logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

---

## Final Recommendations for Promenade

1. **UUID v7 - current standard**: Use everywhere by default
2. **Always index FKs**: For fast JOINs
3. **Store timestamps explicitly**: Even if UUID v7 contains time
4. **Use pkg/uuidv7**: Don't generate manually
5. **Document exceptions**: If using another strategy, explain why

## Migration

See [UUID_V7_MIGRATION.md](UUID_V7_MIGRATION.md) for migration instructions.

## Questions?

Open an issue or ask in team chat.
