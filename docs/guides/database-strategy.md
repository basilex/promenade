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

## Planned: MS SQL Server

**Status:** Planned for enterprise deployments

**Timeline:** Implementation based on market demand

**Rationale:**

- Large enterprise footprint (many companies already have licenses)
- Strong presence in banking, insurance, government sectors
- Integration with existing Microsoft infrastructure (.NET, Azure)
- IT departments familiar with MS SQL ecosystem
- Docker support (Linux containers available)

**Feature compatibility:**

```
MS SQL Server 2022 capabilities:
- Materialized views (Indexed Views)
- JSON support (OPENJSON, FOR JSON)
- CTEs and window functions
- Full-text search
- Partitioning
- OUTPUT clause (equivalent to RETURNING)
- UNIQUEIDENTIFIER type (UUIDs)
- Advanced enterprise features
```

**Implementation approach:**

- PostgreSQL remains reference implementation
- MS SQL gets feature parity where possible
- Clear documentation of any differences
- Graceful degradation for missing features

**Estimated effort:** 2-3 months for production-ready support

---

## Database Comparison Matrix

| Feature                | PostgreSQL | MS SQL Server         | Notes                                |
| ---------------------- | ---------- | --------------------- | ------------------------------------ |
| **Status**             | Production | Planned               | -                                    |
| **Native UUID**        | UUID       | UNIQUEIDENTIFIER      | Different syntax                     |
| **JSON Storage**       | JSONB      | NVARCHAR(MAX) + CHECK | Validation via constraints           |
| **JSON Indexing**      | GIN        | Indexed Views         | Different approach, similar result   |
| **Materialized Views** | Native     | Indexed Views         | MS SQL equivalent                    |
| **RETURNING/OUTPUT**   | RETURNING  | OUTPUT INSERTED       | Different syntax, same functionality |
| **Full-text Search**   | Native     | Native                | Both supported                       |
| **CTEs**               | Yes        | Yes                   | Same syntax                          |
| **Window Functions**   | Yes        | Yes                   | Same syntax                          |
| **Partitioning**       | Native     | Native                | Both supported                       |
| **Cost**               | Free       | Licensed              | PostgreSQL = $0, MS SQL = $$$        |
| **Docker Support**     | Excellent  | Good                  | Both have official images            |

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
- No significant market advantage over PostgreSQL or MS SQL

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

### Phase 2: MS SQL Server Support (Future)

**When:** Based on customer demand from enterprise sector

**Steps:**

1. Dialect layer implementation (pkg/database/mssql/)
2. Migration adapter (PostgreSQL → MS SQL syntax)
3. Repository testing with MS SQL
4. Docker Compose setup for local development
5. CI/CD integration
6. Documentation and examples

**Estimated timeline:** 2-3 months

**Success criteria:**

- Core features work identically
- Clear documentation of differences
- Automated testing for both databases
- Zero performance regression for PostgreSQL

---

### Phase 3: Advanced Features (Long-term)

**Conditional on Phase 2 success:**

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

- PostgreSQL: YES (primary database)
- MS SQL Server: YES (when demand justifies it)
- SQLite: NO (not production-grade for ERP)
- MySQL: NO (no compelling advantages)
- Oracle: NO (cost prohibitive)

---

## Migration Path

### For users wanting MS SQL Server:

**Phase 1:** Use PostgreSQL

- Deploy with Docker Compose
- All features available
- Reference implementation

**Phase 2:** When MS SQL support ships

- Migration tools provided
- Schema conversion automated
- Data migration scripts included
- Rollback capability maintained

**Phase 3:** Production on MS SQL

- Feature parity documented
- Known differences clearly stated
- Support available

---

## Technical Architecture

### Dialect Pattern (pkg/database/)

```
database/
  dialect.go         # Interface definition
  postgres/          # PostgreSQL implementation (current)
    dialect.go
    dialect_test.go
  mssql/            # MS SQL Server (future)
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

// Works with both PostgreSQL and MS SQL
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
  driver: postgres # or: mssql (when available)
  postgres:
    host: localhost
    # ...
  mssql: # future
    host: localhost
    instance: SQLEXPRESS
    # ...
```

---

## Maintenance Philosophy

1. **PostgreSQL First:** Always optimize for PostgreSQL
2. **Feature Parity:** MS SQL gets equivalent functionality, not limited by it
3. **Clear Documentation:** Any differences clearly documented
4. **Test Coverage:** Both databases tested in CI/CD
5. **No Compromise:** Adding databases never compromises existing quality

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

**Current:** PostgreSQL only - production-ready, all features

**Near future:** MS SQL Server support when market demands

**Long-term:** Maintain focus on quality over quantity of databases

**Never:** SQLite, MySQL, Oracle, or databases without compelling business case

---

**Last updated:** January 21, 2026  
**Next review:** When MS SQL Server implementation begins
