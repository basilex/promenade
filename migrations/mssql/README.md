# MS SQL Server Migrations (Planned)

## Status

**MS SQL Server support is planned for future implementation.**

See [docs/guides/database-strategy.md](../../docs/guides/database-strategy.md) for implementation roadmap and timeline.

---

## Structure (When Implemented)

```
migrations/mssql/
  core/              # Core infrastructure
  shared/            # Shared reference data
  identity/          # Identity module
  customer-mgmt/     # Customer management
  order-mgmt/        # Order management
  billing/           # Billing
  accounting/        # Accounting
  banking/           # Banking
  warehouse/         # Warehouse
  fiscal/            # Fiscal integration
  ui/                # UI metadata
  scripting/         # Scripting engine
```

---

## Migration Approach

MS SQL Server migrations will be created to achieve **feature parity** with PostgreSQL migrations:

### Technical Differences

| Feature                | PostgreSQL  | MS SQL Server                         |
| ---------------------- | ----------- | ------------------------------------- |
| **UUID Type**          | `UUID`      | `UNIQUEIDENTIFIER`                    |
| **JSON Storage**       | `JSONB`     | `NVARCHAR(MAX)` with CHECK constraint |
| **JSON Indexing**      | GIN indexes | Indexed Views or computed columns     |
| **Materialized Views** | Native      | Indexed Views (equivalent)            |
| **RETURNING Clause**   | `RETURNING` | `OUTPUT INSERTED`                     |
| **Stored Procedures**  | PL/pgSQL    | T-SQL                                 |

### Example: UUID Column

**PostgreSQL:**

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL
);
```

**MS SQL Server:**

```sql
CREATE TABLE users (
    id UNIQUEIDENTIFIER PRIMARY KEY DEFAULT NEWID(),
    name NVARCHAR(255) NOT NULL
);
```

### Example: JSONB Column

**PostgreSQL:**

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_orders_metadata ON orders USING GIN (metadata);
```

**MS SQL Server:**

```sql
CREATE TABLE orders (
    id UNIQUEIDENTIFIER PRIMARY KEY,
    metadata NVARCHAR(MAX) NOT NULL DEFAULT '{}',
    CONSTRAINT CK_orders_metadata_json CHECK (ISJSON(metadata) = 1)
);

-- Indexed computed column for JSON queries
ALTER TABLE orders ADD metadata_type AS JSON_VALUE(metadata, '$.type');
CREATE INDEX idx_orders_metadata_type ON orders(metadata_type);
```

---

## Implementation Timeline

**Estimated effort**: 2-3 months

**Phase 1: Core Infrastructure** (Week 1-2)

- Dialect implementation (`pkg/database/mssql/`)
- Connection pooling and configuration
- Basic CRUD operations

**Phase 2: Migration Conversion** (Week 3-6)

- Convert all PostgreSQL migrations to T-SQL
- Implement equivalent features (indexed views, JSON handling)
- Schema parity verification

**Phase 3: Testing & Documentation** (Week 7-8)

- Integration testing with MS SQL Server
- Docker Compose setup
- CI/CD integration
- Migration guide for existing deployments

---

## When Available

MS SQL Server support will be added when:

1. Enterprise customer demand justifies the investment
2. Development resources are allocated
3. PostgreSQL implementation is fully stable

**Current focus**: PostgreSQL perfection (production-ready, all features)

**Future**: MS SQL Server bridge for enterprise deployments

---

## Contributing

If you're interested in MS SQL Server support:

1. Review [database-strategy.md](../../docs/guides/database-strategy.md)
2. Open GitHub Discussion to express interest
3. Share your use case and requirements

**Note**: MS SQL Server implementation will maintain PostgreSQL as the reference implementation. All features will be optimized for PostgreSQL first, with MS SQL Server providing feature parity where possible.

---

**Last updated**: January 21, 2026  
**Status**: Not started  
**Next action**: Await enterprise customer demand
