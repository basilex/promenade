# PostgreSQL Driver Migration: lib/pq → pgx

**Status**: Planned (Phase 2 - Q3 2026)  
**Priority**: MEDIUM  
**Target Date**: Q3 2026  
**Last Updated**: January 22, 2026

---

## Overview

Migration plan from `github.com/lib/pq` (maintenance mode) to `pgx/v5` for better PostgreSQL performance and modern features.

---

## Current State

### lib/pq (Current Driver)

**Version**: `github.com/lib/pq` v1.10.9  
**Status**:  **Maintenance mode** (no new features, security fixes only)

**Pros**:

-  Stable and battle-tested
-  Works through `database/sql` stdlib interface
-  Compatible with all ORM/query builders
-  Simple migration from other SQL databases
-  Well-documented

**Cons**:

-  No new features (frozen since 2021)
-  Slower than pgx (15-20% performance gap)
-  No connection pooling optimizations
-  No native PostgreSQL-specific features (COPY, LISTEN/NOTIFY, etc.)
-  Higher memory allocation

**Announcement**: https://github.com/lib/pq/issues/1008

> "This package is in maintenance mode and is not actively developed. For new projects, consider using pgx instead."

---

## Proposed: pgx/v5

**Version**: `github.com/jackc/pgx/v5`  
**Status**:  **Active development**

**Pros**:

-  15-20% faster than lib/pq
-  Lower memory allocations
-  Native PostgreSQL features (COPY, arrays, JSON)
-  Built-in connection pooling (pgxpool)
-  Better error messages
-  Prepared statement cache
-  LISTEN/NOTIFY support
-  Active development and community

**Cons**:

-  Two APIs: stdlib compatible vs native
-  Breaking changes from v4 to v5
-  Migration effort required

**Benchmarks** (from pgx README):

```
BenchmarkPgxStdlib          10000    187251 ns/op    3298 B/op    108 allocs/op
BenchmarkPq                 10000    213907 ns/op    5691 B/op    160 allocs/op
BenchmarkPgxNative          20000     93527 ns/op    2034 B/op     51 allocs/op
```

**Key Insight**: pgx through stdlib is 14% faster, native pgx is 56% faster

---

## Migration Strategy

### Phase 1: pgx via database/sql (Low Risk)

**Timeline**: Q3 2026 (2 months)  
**Risk**: **LOW** (drop-in replacement)  
**Performance Gain**: 14-20%

Use pgx through `database/sql` stdlib interface (minimal code changes):

```go
// Before (lib/pq)
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", dsn)

// After (pgx via stdlib)
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

db, err := sql.Open("pgx", dsn) // Only change: driver name
```

**Benefits**:

- Minimal code changes (driver name only)
- Keep existing sqlx code
- 14-20% performance improvement
- Future-proof (maintained driver)
- Easy rollback if issues found

**Migration Steps**:

1. **Update dependencies** (week 1)

   ```bash
   go get github.com/jackc/pgx/v5/stdlib
   go mod tidy
   ```

2. **Change driver name** (week 2)

   ```go
   // pkg/database/postgres.go
   func Connect(dsn string) (*sqlx.DB, error) {
       db, err := sqlx.Connect("pgx", dsn) // Changed from "postgres"
       if err != nil {
           return nil, err
       }
       return db, nil
   }
   ```

3. **Update DSN format** (week 3)

   ```go
   // Before (lib/pq)
   dsn := "postgresql://user:pass@localhost:5432/db?sslmode=disable"

   // After (pgx) - same format, works!
   dsn := "postgresql://user:pass@localhost:5432/db?sslmode=disable"
   ```

4. **Test thoroughly** (week 4-6)
   - Run all unit tests
   - Run all integration tests
   - Run benchmarks (verify performance gain)
   - Test in staging for 1 week

5. **Deploy to production** (week 7)
   - Canary deployment (10% traffic)
   - Monitor performance metrics
   - Full rollout if stable

6. **Remove lib/pq** (week 8)
   ```bash
   go mod tidy
   ```

---

### Phase 2: Native pgx (High Risk, High Reward)

**Timeline**: Q4 2026 - Q1 2027 (4-6 months)  
**Risk**: **HIGH** (major refactoring)  
**Performance Gain**: 50-60% (vs lib/pq)

Use native pgx API for maximum performance:

```go
// Before (sqlx)
type CustomerRepository struct {
    db *sqlx.DB
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
    var customer Customer
    err := r.db.GetContext(ctx, &customer, "SELECT * FROM customers WHERE id = $1", id)
    return &customer, err
}

// After (pgx native)
type CustomerRepository struct {
    pool *pgxpool.Pool
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
    var customer Customer
    err := r.pool.QueryRow(ctx, "SELECT * FROM customers WHERE id = $1", id).Scan(
        &customer.ID, &customer.Name, &customer.Email, // ...
    )
    return &customer, err
}
```

**Benefits**:

- 50-60% faster than lib/pq
- Lower memory usage (40% reduction)
- Native PostgreSQL features
- Built-in connection pooling
- Prepared statement cache

**Costs**:

- Major refactoring required
- All repositories must be rewritten
- sqlx dependency removed
- Higher learning curve
- Longer migration timeline

**Decision**: **DEFER to Phase 2** (not needed immediately)

**Rationale**:

- Phase 1 gives 80% of benefits with 20% of effort
- Application not yet performance-bottlenecked
- Premature optimization
- Focus on feature development first

---

## Performance Comparison

### Benchmarks (PostgreSQL SELECT)

| Driver       | Ops/sec | Time/op | Memory/op | Allocs/op |
| ------------ | ------- | ------- | --------- | --------- |
| lib/pq       | 4,676   | 213μs   | 5.7 KB    | 160       |
| pgx (stdlib) | 5,342   | 187μs   | 3.3 KB    | 108       |
| pgx (native) | 10,695  | 93μs    | 2.0 KB    | 51        |

**Improvement**:

- **pgx stdlib**: 14% faster, 42% less memory, 32% fewer allocations
- **pgx native**: 56% faster, 65% less memory, 68% fewer allocations

### Real-World Impact

**Current Performance** (lib/pq):

- Customer.GetByID: 1.2ms (query + app)
- Order.ListByCustomer: 8.5ms (query + app)
- Invoice.Generate: 45ms (multiple queries)

**Expected After Phase 1** (pgx stdlib):

- Customer.GetByID: 1.0ms (-16%)
- Order.ListByCustomer: 7.2ms (-15%)
- Invoice.Generate: 38ms (-16%)

**Expected After Phase 2** (pgx native):

- Customer.GetByID: 0.7ms (-42%)
- Order.ListByCustomer: 4.8ms (-44%)
- Invoice.Generate: 25ms (-44%)

**Conclusion**: Phase 1 is sufficient for now (16% gain is significant)

---

## Migration Checklist

### Phase 1 (pgx via stdlib)

#### Pre-Migration

- [ ] Review pgx v5 documentation
- [ ] Update dependencies in go.mod
- [ ] Test DSN format compatibility
- [ ] Create rollback plan
- [ ] Schedule staging deployment

#### Migration

- [ ] Update driver name in `pkg/database/postgres.go`
- [ ] Update connection string (if needed)
- [ ] Update test setup (use pgx driver)
- [ ] Run all unit tests
- [ ] Run all integration tests
- [ ] Run benchmark tests (verify performance gain)

#### Validation

- [ ] Smoke tests pass
- [ ] Integration tests pass (all contexts)
- [ ] Benchmark shows 10-20% improvement
- [ ] No connection pool issues
- [ ] No query compatibility issues

#### Staging Deployment

- [ ] Deploy to staging
- [ ] Monitor for 1 week
- [ ] Check error logs
- [ ] Verify metrics (latency, throughput)
- [ ] Load test

#### Production Deployment

- [ ] Canary deployment (10% traffic)
- [ ] Monitor for 24 hours
- [ ] Gradual rollout (25%, 50%, 100%)
- [ ] Final verification
- [ ] Remove lib/pq dependency

#### Post-Migration

- [ ] Update documentation
- [ ] Update README.md
- [ ] Update CI/CD pipelines
- [ ] Team training (if needed)
- [ ] Celebrate 

---

## Rollback Plan

If issues found after migration:

### Quick Rollback (< 1 hour)

```bash
# 1. Revert driver name
sed -i 's/"pgx"/"postgres"/g' pkg/database/postgres.go

# 2. Revert dependencies
go get github.com/lib/pq@v1.10.9
go mod tidy

# 3. Rebuild
make build

# 4. Deploy previous version
kubectl rollout undo deployment/promenade-api
```

### Persistent Issues

If rollback needed after several days:

1. Create revert PR
2. Test thoroughly in staging
3. Deploy to production
4. Document lessons learned
5. Plan future retry

---

## Known Issues & Mitigations

### Issue 1: Query Compatibility

**Problem**: Some queries may behave differently

**Mitigation**:

- Run full integration test suite
- Test edge cases (NULL values, arrays, JSON)
- Compare query results between drivers

### Issue 2: Connection Pooling

**Problem**: pgx has different pooling behavior

**Mitigation**:

- Use same connection pool settings
- Monitor connection metrics
- Adjust pool size if needed

### Issue 3: Error Messages

**Problem**: Error messages formatted differently

**Mitigation**:

- Update error handling tests
- Verify error discrimination still works
- Update error message documentation

---

## Alternative: Stay on lib/pq

### Arguments FOR Staying

-  Stable and working
-  No migration risk
-  Zero development time
-  Security fixes still applied

### Arguments AGAINST Staying

-  15-20% performance left on table
-  No future improvements
-  Eventually forced to migrate anyway
-  Missing modern PostgreSQL features

**Recommendation**: Migrate in Phase 1 (low risk, high reward)

---

## Timeline

### Q3 2026: Phase 1 (pgx via stdlib)

- **Month 1**: Planning, dependency updates, testing
- **Month 2**: Staging deployment, monitoring
- **Month 3**: Production deployment, verification

### Q4 2026 - Q1 2027: Phase 2 (Optional - Native pgx)

- Evaluate if performance improvement needed
- If yes: Plan repository refactoring
- If no: Defer indefinitely

---

## Success Criteria

### Phase 1

-  All tests passing
-  10-20% performance improvement measured
-  No production incidents
-  Zero customer-facing issues
-  Stable for 30 days

### Phase 2 (If pursued)

-  40-50% performance improvement measured
-  Lower memory usage
-  All repositories refactored
-  No regressions
-  Team trained on pgx native API

---

## Related Documentation

- [Database Strategy](database-strategy.md) - Overall database approach
- [Connection Pooling](database-connection-pooling.md) - Pool configuration
- [Testing Patterns](testing-patterns.md) - Integration test setup
- [Migration Rollback](migration-testing-rollback.md) - Rollback procedures

---

## External Resources

- [pgx GitHub](https://github.com/jackc/pgx)
- [lib/pq Maintenance Mode Announcement](https://github.com/lib/pq/issues/1008)
- [pgx vs lib/pq Comparison](https://github.com/jackc/pgx/wiki/PostgreSQL-Driver-Comparison)
- [Go database/sql Documentation](https://pkg.go.dev/database/sql)

---

**Status**: Planned (Phase 1 - Q3 2026)  
**Owner**: Platform Team  
**Sponsor**: CTO  
**Maintainer**: Promenade Team
