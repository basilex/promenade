# UUID v7 Migration Guide

## Quick Start

```bash
# 1. Apply migrations
make migrate-up

# 2. Run tests
go test ./pkg/uuidv7/

# 3. Restart application
make dev
```

## What Was Done

### 1. PostgreSQL

- ✅ Added `uuid_generate_v7()` function (migration 000005)
- ✅ Updated table defaults for `users`, `products`, `roles` (migration 000006)
- ✅ Existing records kept UUID v4, new ones will get v7

### 2. Go Code

- ✅ Created `pkg/uuidv7` package with full implementation
- ✅ Updated usecases: `auth_usecase.go`, `product_usecase.go`, `role_usecase.go`
- ✅ Updated generator template `scripts/templates/usecase.tmpl`

### 3. Tests and Performance

```
BenchmarkNew-16 (UUID v7):    87.43 ns/op   0 allocs
BenchmarkNewV4-16 (UUID v4):  185.9 ns/op   1 alloc

→ UUID v7 is 2.1x faster with zero allocations!
```

## Applying Migrations

```bash
# Check current version
make migrate-version

# Apply new migrations
make migrate-up

# Rollback if needed
make migrate-down  # rollback last migration
```

After migrations, PostgreSQL will have:

```sql
-- Generate UUID v7
SELECT uuid_generate_v7();

-- Check UUID version
SELECT
    id,
    substring(id::text from 15 for 1) as version  -- should be '7' for new records
FROM users
LIMIT 5;
```

## Usage in New Code

### Creating New Entity

```go
import "github.com/basilex/promenade/pkg/uuidv7"

// Generate new ID
user := &entity.User{
    ID:    uuidv7.New(),  // instead of uuid.New()
    Email: "user@example.com",
}

// Extract creation time from ID
createdAt := uuidv7.ExtractTime(user.ID)
fmt.Printf("User created at: %v\n", createdAt)
```

### Entity Generator

Now `scripts/generate-interactive.sh` automatically uses UUID v7:

```bash
make generate-interactive
# or
./scripts/generate-interactive.sh
```

## UUID v7 Advantages

### 1. Database Performance

```
UUID v4 (random):
  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐
  │ aef │ │ 1b2 │ │ 9cd │ │ 3de │  ❌ Chaos in B-tree index
  └─────┘ └─────┘ └─────┘ └─────┘

UUID v7 (time-ordered):
  ┌─────┬─────┬─────┬─────┐
  │ 1   │ 2   │ 3   │ 4   │        ✅ Sequential writes
  └─────┴─────┴─────┴─────┘
```

**Results:**

- 20-50% faster INSERT in benchmarks
- Fewer page splits in indexes
- Better cache locality
- Less index fragmentation

### 2. Time-based Sorting

```sql
-- Natural sorting by ID = sort by creation time
SELECT * FROM users
ORDER BY id DESC  -- newest first
LIMIT 10;

-- No need for created_at index for simple queries
```

### 3. Timestamp Extraction

```go
// Get creation time without database query
id := uuidv7.MustParse("018d2f07-0f3c-7000-8000-123456789abc")
createdAt := uuidv7.ExtractTime(id)

// Useful for:
// - Debugging
// - Logging
// - Analytics
// - Partitioning
```

## Compatibility

### Existing Data

Old UUID v4 will work normally:

```go
// Code works with both v4 and v7
func GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    // id can be any UUID version
    return repo.GetByID(ctx, id)
}
```

### API Compatibility

API hasn't changed - clients won't notice the difference:

```json
{
  "id": "018d2f07-0f3c-7000-8000-123456789abc", // v7
  "email": "user@example.com"
}
```

## Post-Deploy Monitoring

### Check v7 Generation

```sql
-- All new records should have version 7
SELECT
    COUNT(*) as total,
    substring(id::text from 15 for 1) as uuid_version
FROM users
WHERE created_at > NOW() - INTERVAL '1 day'
GROUP BY uuid_version;

-- Expected result:
-- total | uuid_version
-- ------+--------------
-- 1523  | 7
```

### INSERT Performance

```sql
-- Before migration (baseline)
EXPLAIN ANALYZE
INSERT INTO users (email, name, password)
VALUES ('test@example.com', 'Test', 'hash');

-- Execution Time: ~2.5ms (example)

-- After migration (expect 20-50% improvement)
-- Execution Time: ~1.5-2.0ms (example)
```

### Index Size

```sql
SELECT
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as size
FROM pg_indexes
WHERE tablename IN ('users', 'products', 'roles')
ORDER BY pg_relation_size(indexname::regclass) DESC;
```

## Rollback in Case of Issues

### Step 1: Rollback DB Migrations

```bash
make migrate-down  # rollback 000006
make migrate-down  # rollback 000005
```

### Step 2: Revert Code to UUID v4

```bash
git revert <commit-hash>
# or create hotfix replacing uuidv7.New() with uuid.New()
```

### Step 3: Redeploy

```bash
make build
make docker-build
# deploy...
```

## Next Steps

- [ ] Monitor performance metrics for the first week
- [ ] Update API documentation (if needed)
- [ ] Add UUID v7 to existing microservices
- [ ] Consider partitioning by timestamp from UUID v7

## Questions?

See full documentation: `docs/UUID_V7_GUIDE.md`
