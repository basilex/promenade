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

| Target                 | Description                                    | Duration  |
| ---------------------- | ---------------------------------------------- | --------- |
| `make test-benchmark`  | Run all benchmarks (5s per benchmark)         | ~30s      |
| `make test-benchmark-all` | Extended benchmarks (10s per benchmark)     | ~60s      |
| `make test-db-start`   | Start test database (port 5433)                | Instant   |
| `make test-db-stop`    | Stop test database                             | Instant   |

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

**Managed by**: `docker-compose.test.yml`

```bash
# Start test DB
make test-db-start

# Stop test DB
make test-db-stop

# Check if running
docker ps | grep promenade_test
```

---

## Related Documentation

- [Test README](../README.md) - Complete testing guide
- [Testing Patterns](../../docs/guides/testing-patterns.md) - Best practices
- [N+1 Optimization](../../docs/work-in-progress/N+1_OPTIMIZATION.md) - Real-world example

---

**Last Updated**: December 31, 2025  
**Status**: Production-ready  
**Benchmarks**: 4 tests (Identity User ListUsers)  
**Maintainer**: Promenade Team
