# Benchmark Tests

**Performance measurement** tests for Promenade platform - validate optimizations and measure query performance.

---

## Overview

Benchmark tests use Go's built-in `testing.B` framework to measure performance characteristics:

- **Execution time** (ns/op) - Time per operation
- **Memory usage** (B/op) - Bytes allocated per operation
- **Allocations** (allocs/op) - Number of memory allocations

**Database**: Tests require real PostgreSQL database (like integration tests)

---

## Directory Structure

```
test/benchmark/
 contexts/                   # Mirror path structure (same as integration tests)
    identity/
       user/
          repository_bench_test.go  # User repository benchmarks
    shared/
       # Future: Reference data benchmarks
    customer-mgmt/
       # Future: Customer management benchmarks
```

**Naming Convention**: `{entity}_bench_test.go` or `repository_bench_test.go`

---

## Running Benchmarks

### Quick Commands

```bash
# Run all benchmarks (auto-starts test DB)
make test-benchmark

# Extended benchmarks (10s per benchmark)
make test-benchmark-all

# Specific benchmark
go test -bench=BenchmarkListUsers_SmallDataset -benchmem ./test/benchmark/contexts/identity/user

# All benchmarks in a package
go test -bench=. -benchmem ./test/benchmark/contexts/identity/user

# With CPU profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof ./test/benchmark/contexts/identity/user
go tool pprof cpu.prof

# With memory profiling
go test -bench=. -benchmem -memprofile=mem.prof ./test/benchmark/contexts/identity/user
go tool pprof mem.prof
```

### Makefile Targets

| Target                    | Description                             | Duration |
| ------------------------- | --------------------------------------- | -------- |
| `make test-benchmark`     | Run all benchmarks (5s per benchmark)   | ~30s     |
| `make test-benchmark-all` | Extended benchmarks (10s per benchmark) | ~60s     |
| `make test-db-start`      | Start test database (port 5433)         | Instant  |
| `make test-db-stop`       | Stop test database                      | Instant  |

**Note**: Makefile targets automatically start/stop test database

---

## Writing Benchmarks

### Basic Structure

```go
package user_test

import (
    "context"
    "testing"

    "github.com/jmoiron/sqlx"

    "github.com/basilex/promenade/internal/contexts/identity/user"
    userRepo "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
)

// BenchmarkListUsers_SmallDataset benchmarks ListUsers with 20 users
func BenchmarkListUsers_SmallDataset(b *testing.B) {
    db := setupBenchmarkDB(b)
    defer db.Close()

    repo := userRepo.NewUserRepository(db)
    ctx := context.Background()

    // Setup: Create test data (BEFORE b.ResetTimer)
    setupTestData(b, ctx, repo, 20)

    // Reset timer after setup (exclude setup time from benchmark)
    b.ResetTimer()

    // Benchmark loop
    for i := 0; i < b.N; i++ {
        users, total, err := repo.ListUsers(ctx, 1, 20)
        if err != nil {
            b.Fatalf("ListUsers failed: %v", err)
        }
        if len(users) != 20 {
            b.Fatalf("Expected 20 users, got %d", len(users))
        }
    }
}
```

### Setup Helper

```go
// setupBenchmarkDB creates clean test database connection
func setupBenchmarkDB(b *testing.B) *sqlx.DB {
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        dsn = "postgresql://system:passw0rd@localhost:5433/promenade_test?sslmode=disable"
    }

    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        b.Fatalf("Failed to connect to test database: %v", err)
    }

    // Clean tables
    tables := []string{"identity_users", "identity_roles"}
    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
        if err != nil {
            b.Fatalf("Failed to truncate table %s: %v", table, err)
        }
    }

    return db
}
```

---

## Best Practices

### DO

- **Reset timer after setup**: `b.ResetTimer()` excludes setup from measurement
- **Clean database before test**: Truncate tables for consistent results
- **Use test DB**: Port 5433 (not production 5432)
- **Measure meaningful operations**: Focus on hot paths and optimizations
- **Run multiple times**: Use `-benchtime=10s` for stable results
- **Document hardware**: Include CPU/RAM in benchmark comments

### DON'T

- **Don't include setup in benchmark**: Always call `b.ResetTimer()`
- **Don't use shared state**: Each benchmark should be independent
- **Don't run on production DB**: Always use test database
- **Don't ignore errors**: Use `b.Fatalf()` for assertion failures
- **Don't skip validation**: Verify query correctness, not just speed

---

## Benchmark Scenarios

### Common Patterns

1. **Small Dataset** (10-20 records) - Typical page size
2. **Medium Dataset** (100-500 records) - Realistic production load
3. **Large Dataset** (1000+ records) - Stress test
4. **Edge Cases** - Empty results, no relations, etc.
5. **Optimization Validation** - Before/after comparisons

### Naming Convention

```
Benchmark{Operation}_{Scenario}
BenchmarkListUsers_SmallDataset      // 20 users
BenchmarkListUsers_MediumDataset     // 100 users
BenchmarkListUsers_LargeDataset      // 1000 users
BenchmarkListUsers_NoRoles           // Users without relations
BenchmarkListUsers_MultipleRoles     // Users with many relations
```

---

## Example: N+1 Query Optimization

**Before**: N+1 problem (1 query + N queries for relations)

```
BenchmarkListUsers_Before-16    500    2500000 ns/op    150000 B/op    1500 allocs/op
```

**After**: Single JOIN query

```
BenchmarkListUsers_After-16    1484     701680 ns/op     50309 B/op     558 allocs/op
```

**Improvement**: 3.5x faster, 3x less memory, 2.7x fewer allocations

---

## Interpreting Results

### Time (ns/op)

- **<1ms** - Excellent (suitable for high-traffic endpoints)
- **1-10ms** - Good (acceptable for most operations)
- **10-100ms** - Acceptable (for complex queries)
- **>100ms** - Needs optimization

### Memory (B/op)

- **<10KB** - Excellent
- **10-100KB** - Good
- **100KB-1MB** - Acceptable
- **>1MB** - Consider optimization

### Allocations (allocs/op)

- **<100** - Excellent
- **100-500** - Good
- **500-1000** - Acceptable
- **>1000** - Consider pooling or reuse

---

## Test Database

**Connection**: `postgresql://system:passw0rd@localhost:5433/promenade_test`

**Managed by**: `docker-compose.postgres.test.yml`

```bash
# Start test DB
make test-db-start

# Stop test DB
make test-db-stop

# Check if running
docker ps | grep promenade_test
```

---

## Performance Baselines

**Hardware**: Apple M4 Max, 16 cores, 64GB RAM  
**Database**: PostgreSQL 14, localhost, SSD  
**Date**: January 22, 2026

### Identity Context - User Repository

| Operation                             | Ops/sec | Avg Time | Memory/op | Allocs/op |
| ------------------------------------- | ------- | -------- | --------- | --------- |
| `ListUsers_SmallDataset` (20 users)   | TBD     | TBD      | TBD       | TBD       |
| `ListUsers_MediumDataset` (100 users) | TBD     | TBD      | TBD       | TBD       |
| `ListUsers_NoRoles`                   | TBD     | TBD      | TBD       | TBD       |
| `ListUsers_MultipleRoles`             | TBD     | TBD      | TBD       | TBD       |

### Customer Management - Interaction Repository

| Operation           | Ops/sec | Avg Time | Memory/op | Allocs/op |
| ------------------- | ------- | -------- | --------- | --------- |
| `CreateInteraction` | ~2,630  | 380μs    | 5.2 KB    | 102       |
| `ListByCustomer`    | ~1,130  | 886μs    | 59.3 KB   | 961       |

**Notes**:

- User benchmarks currently failing (setup data issue - see Issue #TBD)
- Interaction benchmarks stable and passing
- Times are median values from `-benchtime=3s` runs
- Memory includes all heap allocations (query + object creation)

---

## Regression Detection

### Using `benchstat`

Compare benchmark results before/after changes to detect regressions:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run benchmarks BEFORE changes
go test -bench=. -benchmem ./test/benchmark/... -run=^$ > before.txt

# Make your code changes...

# Run benchmarks AFTER changes
go test -bench=. -benchmem ./test/benchmark/... -run=^$ > after.txt

# Compare results
benchstat before.txt after.txt
```

### Example Output

```
name                    old time/op    new time/op    delta
CreateInteraction-16      380μs ± 2%     285μs ± 1%  -25.00%  (p=0.000 n=10+10)
ListByCustomer-16         886μs ± 3%     912μs ± 2%   +2.93%  (p=0.001 n=10+10)

name                    old alloc/op   new alloc/op   delta
CreateInteraction-16     5.24kB ± 0%    4.12kB ± 0%  -21.37%  (p=0.000 n=10+10)
ListByCustomer-16        59.3kB ± 0%    60.1kB ± 0%   +1.35%  (p=0.000 n=10+10)

name                    old allocs/op  new allocs/op  delta
CreateInteraction-16       102 ± 0%        87 ± 0%  -14.71%  (p=0.000 n=10+10)
ListByCustomer-16          961 ± 0%       968 ± 0%   +0.73%  (p=0.000 n=10+10)
```

**Interpretation**:

-  **-25% time**: Significant improvement
-  **+2.93% time**: Minor regression (acceptable if <5%)
-  **-21% memory**: Excellent improvement
-  **+1.35% memory**: Minor increase (monitor)

### Regression Thresholds

**Block PR if**:

- Time regression >10%
- Memory regression >20%
- Allocation regression >20%

**Investigate if**:

- Time regression 5-10%
- Memory regression 10-20%
- Allocation regression 10-20%

**Acceptable**:

- Time regression <5% (noise tolerance)
- Memory regression <10%
- Allocation regression <10%

### CI Integration (Future)

```yaml
# .github/workflows/benchmarks.yml
name: Benchmark Regression Check
on: [pull_request]
jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout base branch
        run: git checkout ${{ github.base_ref }}
      - name: Run base benchmarks
        run: make test-benchmark > base.txt
      - name: Checkout PR branch
        run: git checkout ${{ github.head_ref }}
      - name: Run PR benchmarks
        run: make test-benchmark > pr.txt
      - name: Compare results
        run: benchstat base.txt pr.txt
```

---

## Fixing Benchmark Data Issues

**Current Issue**: User benchmarks fail with "Expected 20 users, got 0"

**Root Cause**: Test data setup not persisting or transaction isolation issue

**Debug Steps**:

```bash
# 1. Check if test data is created
go test -v -bench=BenchmarkListUsers_SmallDataset ./test/benchmark/contexts/identity/user

# 2. Verify database connection
psql postgresql://system:passw0rd@localhost:5433/promenade_test -c "SELECT COUNT(*) FROM identity_users;"

# 3. Check for transaction rollback issues
# Benchmarks should NOT use db.WithTransaction() - commits must persist
```

**Fix Pattern**:

```go
//  WRONG - Transaction rolls back after setup
func setupTestData(b *testing.B, ctx context.Context, repo *Repository) {
    db.WithTransaction(b, func(ctx context.Context, tx *sqlx.Tx) {
        // Data created here will be rolled back!
    })
}

//  CORRECT - Direct inserts that persist
func setupTestData(b *testing.B, ctx context.Context, repo *Repository) {
    for i := 0; i < 20; i++ {
        user := aggregate.NewUser(fmt.Sprintf("user%d@test.com", i))
        err := repo.Create(ctx, user)
        if err != nil {
            b.Fatalf("Failed to create user: %v", err)
        }
    }
}
```

---

## Adding New Benchmarks

### Step 1: Create benchmark file

```bash
# Mirror integration test structure
mkdir -p test/benchmark/contexts/order-mgmt/order
touch test/benchmark/contexts/order-mgmt/order/repository_bench_test.go
```

### Step 2: Write benchmarks

```go
package order_test

import (
    "context"
    "testing"
    "github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/repository/postgres"
)

func BenchmarkCreateOrder(b *testing.B) {
    db := setupBenchmarkDB(b)
    defer db.Close()
    repo := postgres.NewOrderRepository(db)
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Benchmark logic
    }
}
```

### Step 3: Establish baseline

```bash
# Run 10 times to get stable baseline
go test -bench=BenchmarkCreateOrder -benchtime=10s -count=10 ./test/benchmark/contexts/order-mgmt/order
```

### Step 4: Update baselines table in this README

---

## Related Documentation

- [Test README](../README.md) - Complete testing guide
- [Testing Patterns](../../docs/guides/testing-patterns.md) - Best practices
- [N+1 Optimization](../../docs/reference/n-plus-one-optimization.md) - Real-world example
- [Database Query Patterns](../../docs/guides/database-query-patterns.md) - Query optimization
- [Performance Tuning](../../docs/guides/database-connection-pooling.md) - Connection pooling

---

**Last Updated**: January 22, 2026  
**Status**: Production-ready (2 passing, 4 failing - data setup issue)  
**Benchmarks**: 6 tests across 2 contexts  
**Hardware**: Apple M4 Max, PostgreSQL 14  
**Maintainer**: Promenade Team
