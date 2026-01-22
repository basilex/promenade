# ADR-0002: Use sqlx Instead of GORM

**Status**: Accepted  
**Date**: 2026-01-15  
**Deciders**: Core Architecture Team  
**Tags**: database, orm, performance, clean-architecture

**Update (January 2026):** Migrated from `lib/pq` to `pgx/v5/stdlib` driver for 20-30% performance improvement while maintaining full sqlx compatibility.

---

## Context

Promenade requires a database access layer for PostgreSQL. The application follows **Clean Architecture** with:

- Domain aggregates (business logic)
- Repository interfaces (domain layer)
- Repository implementations (infrastructure layer)

**Requirements**:

- Raw SQL control (complex joins, CTEs, window functions)
- High performance (10K+ TPS target)
- Clean separation of concerns (no ORM "magic")
- PostgreSQL 14+ support
- Transaction propagation via context

**Concerns**:

- GORM provides convenience but hides SQL complexity
- ORMs often generate inefficient queries (N+1 problem)
- Clean Architecture requires explicit repository interfaces (ORMs complicate this)

---

## Considered Options

### Option 1: database/sql (Standard Library)

- **Pros**:
  - Zero dependencies
  - Maximum control
  - Fastest performance
- **Cons**:
  - Manual row scanning (boilerplate: `Scan(&id, &name, &email, ...)`)
  - No struct mapping
  - Verbose error handling
- **Example**:
  ```go
  rows, err := db.Query("SELECT id, name FROM customers")
  for rows.Next() {
      var id uuid.UUID
      var name string
      rows.Scan(&id, &name)  // Manual scanning
  }
  ```

### Option 2: GORM (Full ORM)

- **Pros**:
  - Auto-migrations
  - Associations (has-many, belongs-to)
  - Query builder
- **Cons**:
  - **Performance**: 40-60% slower than raw SQL (query builder overhead)
  - **Magic**: Implicit behavior (joins, lazy loading) breaks Clean Architecture
  - **N+1 Problem**: Easy to write inefficient queries
  - **Limited SQL**: CTEs, window functions require raw SQL anyway
  - **Version risk**: Breaking changes between major versions
- **Example**:
  ```go
  db.Where("status = ?", "active").Find(&customers)  // Hidden SQL
  ```

### Option 3: sqlx (SQL Extension)

- **Pros**:
  - **Struct scanning**: `sqlx.Get(&customer, query)` (no manual Scan)
  - **Named parameters**: `:name` instead of `$1, $2, ...`
  - **Raw SQL control**: Write exact SQL needed
  - **Transaction propagation**: `sqlx.ExtContext` interface
  - **Performance**: Near-native (5-10% overhead vs database/sql)
  - **Multi-DB**: PostgreSQL, MySQL, SQLite
- **Cons**:
  - No query builder (write SQL manually)
  - No auto-migrations (use dedicated migration tool)
- **Example**:
  ```go
  var customer Customer
  err := db.Get(&customer, "SELECT * FROM customers WHERE id = $1", id)
  ```

### Option 4: sqlc (SQL Compiler)

- **Pros**:
  - Type-safe SQL (generated from .sql files)
  - Fast (compiled queries)
- **Cons**:
  - Build-time code generation (adds complexity)
  - Less flexible (queries must be in .sql files)
  - Overkill for Clean Architecture (repositories already type-safe)

---

## Decision

**We will use sqlx** as the database access layer instead of GORM.

**Rationale**:

1. **Clean Architecture fit**: Explicit repository implementations (no ORM "magic")
2. **Performance**: Near-native speed (5-10% overhead vs 40-60% for GORM)
3. **SQL control**: Complex queries (CTEs, window functions, lateral joins) without fighting ORM
4. **Struct scanning**: Reduces boilerplate vs database/sql
5. **Transaction propagation**: `sqlx.ExtContext` allows context-based transactions

**Implementation**:

```go
// internal/contexts/customer-mgmt/customer/adapter/repository/postgres/customer_repository.go
import "github.com/jmoiron/sqlx"

type CustomerRepository struct {
    db *sqlx.DB
}

func (r *CustomerRepository) FindByID(ctx context.Context, id uuid.UUID) (*aggregate.Customer, error) {
    var customer aggregate.Customer

    query := `
        SELECT id, name, email, status, created_at, updated_at
        FROM customers
        WHERE id = $1 AND deleted_at IS NULL
    `

    exec := r.getExecutor(ctx)  // Transaction propagation
    if err := exec.GetContext(ctx, &customer, query, id); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrCustomerNotFound
        }
        return nil, err
    }

    return &customer, nil
}

// Transaction propagation helper
func (r *CustomerRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
    if tx, ok := database.GetTx(ctx); ok {
        return tx  // Use active transaction
    }
    return r.db  // Fallback to connection pool
}
```

**Transaction Example**:

```go
// usecase/customer_usecase.go
func (uc *UseCase) CreateCustomer(ctx context.Context, req CreateCustomerRequest) error {
    return database.WithTransaction(ctx, uc.db, func(ctx context.Context) error {
        // Create customer
        customer := aggregate.NewCustomer(req.Name, req.Email)
        if err := uc.customerRepo.Create(ctx, customer); err != nil {
            return err
        }

        // Create contact (same transaction)
        contact := aggregate.NewContact(customer.ID, req.ContactName)
        if err := uc.contactRepo.Create(ctx, contact); err != nil {
            return err  // Automatic rollback
        }

        return nil  // Automatic commit
    })
}
```

---

## Consequences

### Positive

- **Performance**: 5-10% overhead (vs 40-60% for GORM)
- **SQL control**: Write exact queries needed (no query builder limitations)
- **Clean Architecture**: Explicit repository implementations (no hidden behavior)
- **Debugging**: Raw SQL visible in code (no generated queries)
- **Flexibility**: Easy to optimize hot paths (add indexes, rewrite queries)
- **Multi-DB**: PostgreSQL (production), SQLite (tests)

### Negative

- **Manual SQL**: No query builder (must write SQL by hand)
- Mitigation: Repository pattern abstracts SQL from use cases
- **No auto-migrations**: Use dedicated tool (`golang-migrate/migrate`)
- Mitigation: Already implemented (`migrations/` directory, `make migrate`)
- **Struct tags**: Requires `db:` tags on aggregates
- Mitigation: Standard practice, same as JSON tags

### Neutral

- ℹ Team must know SQL (acceptable, SQL is core skill)
- ℹ No associations helper (must write JOINs manually)
- ℹ Repository tests require real database (integration tests with Docker)

---

## Implementation Notes

**Rollout**:

- **Phase 1** (2026-01-15): `pkg/database` wrapper created
- **Phase 2** (2026-01-16): All repositories migrated to sqlx
- **Phase 3** (2026-01-17): Transaction propagation tested (102+ integration tests)

**Affected Components**:

- All repository implementations (`adapter/repository/postgres/`)
- `pkg/database` - sqlx wrapper with transaction helpers
- Transaction propagation via `database.WithTransaction()`

**Rollback Plan**:

- Switch to GORM (unlikely, would require full rewrite)
- Estimated effort: 2-3 weeks (all repositories)

---

## References

- [sqlx Documentation](https://github.com/jmoiron/sqlx)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Robert C. Martin
- [ORM Performance Comparison](https://github.com/efectn/go-orm-benchmarks) - GORM vs sqlx benchmarks
- [pkg/database/README.md](../../pkg/database/README.md) - Implementation details

---

## Revisit Criteria

**Reconsider this decision if**:

- Query performance degrades by >20% compared to baseline
- Team struggles with raw SQL (evidence: >50% of PR feedback is SQL-related)
- New requirement for complex associations (has-many-through, polymorphic)
- sqlx maintenance stalls (last commit >2 years ago)
- Competing library emerges with better performance + query builder

---

## Changelog

| Date       | Change                                             | Author              |
| ---------- | -------------------------------------------------- | ------------------- |
| 2026-01-15 | Initial version                                    | Alexander Vasilenko |
| 2026-01-22 | Updated with integration test results (102+ tests) | Alexander Vasilenko |
