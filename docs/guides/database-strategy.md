# Database Strategy

**Database support strategy** for Promenade Platform - focused on production quality over quantity.

---

## Current Status (January 2026)

### Production: PostgreSQL 14+

**Status:** Fully supported, production-ready

**Features:**

- Native UUID type
- JSONB with GIN indexes
- Materialized views
- CTEs and window functions
- Full-text search
- Partitioning
- Advanced query optimization
- Excellent Docker support

**Why PostgreSQL:**

- Industry-standard for complex applications
- All advanced features needed for ERP
- Strong ACID compliance
- Active development and community
- Cost-effective (open source)

---

---

## Rejected Databases

### SQLite

**Rejected:** Not suitable for production ERP

**Reasons:**

- Single-writer limitation (concurrency issues)
- No native JSONB support
- Limited full-text search
- No partitioning
- Not designed for multi-user applications
- No materialized views

**Conclusion:** Good for embedded apps, not for enterprise ERP

---

### MySQL/MariaDB

**Rejected:** Insufficient feature set

**Reasons:**

- Limited JSON capabilities (worse than PostgreSQL)
- No materialized views
- Weaker window functions support
- Less advanced query optimizer
- No significant market advantage over PostgreSQL

**Conclusion:** Offers no compelling reason to support

---

### Oracle Database

**Rejected:** Cost prohibitive

**Reasons:**

- Extremely expensive licensing ($$$$$)
- Complex deployment
- Overkill for most use cases
- PostgreSQL provides 90% of functionality at $0 cost

**Conclusion:** Not worth the investment for our market

---

### PostgreSQL-compatible (CockroachDB, YugabyteDB)

**Status:** Low priority

**Consideration:** May support in future for:

- Cloud-native deployments
- Multi-region geo-distribution
- Horizontal scaling needs

**Note:** These are PostgreSQL wire protocol compatible, so support would be nearly automatic if needed.

---

## Implementation Strategy

### Phase 1: PostgreSQL Perfection (Current)

**Focus:** Make PostgreSQL implementation world-class

- All advanced features
- Optimal performance
- Complete test coverage
- Production hardening

**Status:** COMPLETE

---

### Phase 2: Advanced Features (Long-term)

- Query builder (database-agnostic)
- Migration generator
- Performance comparison tools
- Database selection wizard

---

## Decision Framework

### When to add database support:

**YES - Add support if:**

1. Strong enterprise customer demand
2. Technical feature parity (>80% of PostgreSQL features)
3. ROI positive (large market segment)
4. Reasonable implementation effort (<6 months)
5. Long-term maintenance feasible

**NO - Reject if:**

1. Niche database with small user base
2. Missing critical features (JSONB, materialized views)
3. High implementation cost vs limited market
4. Would compromise PostgreSQL quality
5. Licensing issues

### Current decisions:

- PostgreSQL: YES (primary and only supported database)
- SQLite: NO (not production-grade for ERP)
- MySQL: NO (no compelling advantages)
- Oracle: NO (cost prohibitive)
- MS SQL Server: NO (no market demand)

---

---

## Technical Architecture

### Dialect Pattern (pkg/database/)

```
database/
  dialect.go         # Interface definition
  postgres/          # PostgreSQL implementation
    dialect.go
    dialect_test.go
```

### Repository Pattern

All repositories use dialect abstraction:

```go
type Repository struct {
    db      *sqlx.DB
    dialect database.Dialect
}

func (r *Repository) Create(ctx context.Context, entity Entity) error {
    query := fmt.Sprintf(
        "INSERT INTO entities (...) VALUES (%s, %s) %s",
        r.dialect.Placeholder(1),
        r.dialect.Placeholder(2),
        r.dialect.SupportsReturning() ? "RETURNING id" : "",
    )
    // ...
}
```

### Configuration

```yaml
database:
  driver: postgres
  postgres:
    host: localhost
    port: 5432
    # ...
```

---

## Maintenance Philosophy

1. **PostgreSQL Excellence:** Focus on making PostgreSQL implementation world-class
2. **Feature Completeness:** Use all PostgreSQL advanced features
3. **Clear Documentation:** Comprehensive guides and examples
4. **Test Coverage:** Full CI/CD coverage
5. **No Compromise:** Quality over quantity

---

## Future Considerations

### Cloud-native databases:

If customer demand emerges:

- CockroachDB (PostgreSQL-compatible)
- YugabyteDB (PostgreSQL-compatible)
- Amazon Aurora PostgreSQL (AWS)

**Note:** PostgreSQL-compatible means minimal work to support

### Specialized databases:

**Not planned:**

- Time-series databases (use PostgreSQL TimescaleDB extension)
- NoSQL databases (not suitable for ERP transactional data)
- In-memory databases (use PostgreSQL + Redis caching)

---

## Summary

**Current:** PostgreSQL only - production-ready, all features, only supported database

**Philosophy:** Focus on quality over quantity - maintain excellence with PostgreSQL rather than spreading efforts across multiple databases

**Never:** SQLite, MySQL, Oracle, MS SQL Server, or databases without compelling business case

---

**Last updated:** January 2026  
**Next review:** As needed
