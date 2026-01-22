# Migration Testing & Rollback Strategy

**Status**: Planned (Phase 2)  
**Priority**: MEDIUM  
**Target Date**: Q2 2026  
**Last Updated**: January 22, 2026

---

## Overview

Strategy for testing database migration reversibility and ensuring safe rollback capabilities in production.

---

## Current State

###  Implemented

- **Up migrations**: All contexts have up.sql files
- **Down migrations**: All contexts have down.sql files
- **Namespace isolation**: 11 migration namespaces (core, shared, identity, customer-mgmt, etc.)
- **Version tracking**: Applied migrations tracked in `schema_migrations` table
- **Make targets**: `make migrate`, `make migrate-core`, etc.

###  Missing

- **Rollback command**: No `make migrate-rollback` target
- **Rollback tests**: No automated testing that down migrations work
- **CI validation**: No verification in CI that migrations are reversible
- **Rollback documentation**: No runbook for production rollback scenarios

---

## Risks Without Rollback Testing

### 1. Broken Down Migrations

**Problem**: Down migration doesn't reverse up migration

```sql
-- up.sql
ALTER TABLE customers ADD COLUMN tier VARCHAR(50);
UPDATE customers SET tier = 'standard';

-- down.sql (WRONG - doesn't restore data)
ALTER TABLE customers DROP COLUMN tier;
```

**Impact**:

- Data loss on rollback
- Cannot safely revert deployment
- Production incidents

### 2. Dependency Order

**Problem**: Migrations have dependencies

```sql
-- Migration 001: Create customers table
-- Migration 002: Create orders table with FK to customers
-- Migration 003: Add index on orders.customer_id

-- Rollback 003: OK
-- Rollback 002: FAILS (FK constraint from migration 003 still exists)
```

**Impact**:

- Rollback fails mid-way
- Database in inconsistent state
- Manual intervention required

### 3. Data Transformation Reversibility

**Problem**: Some transformations are not reversible

```sql
-- up.sql: Merge first_name + last_name → full_name
UPDATE customers SET full_name = first_name || ' ' || last_name;
ALTER TABLE customers DROP COLUMN first_name, DROP COLUMN last_name;

-- down.sql: Cannot split full_name back to first_name + last_name accurately
```

**Impact**:

- Information loss
- Rollback impossible without backup
- Application broken on rollback

---

## Proposed Solution

### Phase 2A: Add Rollback Command

#### Implementation

```makefile
# Makefile.dev.mk

migrate-rollback: validate-env  ## Rollback last N migrations
	@if [ -z "$(MODULE)" ]; then \
		echo " Error: MODULE required. Usage: make migrate-rollback MODULE=core STEPS=1"; \
		exit 1; \
	fi; \
	if [ -z "$(STEPS)" ]; then \
		echo " Error: STEPS required. Usage: make migrate-rollback MODULE=core STEPS=1"; \
		exit 1; \
	fi; \
	echo "Rolling back $(STEPS) migration(s) from $(MODULE)..."; \
	go run cmd/migrate/main.go --cmd=down --namespace=$(MODULE) --steps=$(STEPS)

migrate-rollback-all: validate-env  ## Rollback all migrations (DANGEROUS)
	@echo " WARNING: This will rollback ALL migrations!"; \
	read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "Rolling back all migrations..."; \
		go run cmd/migrate/main.go --cmd=down --namespace=all; \
	else \
		echo " Cancelled"; \
	fi
```

#### Usage

```bash
# Rollback last migration from core
make migrate-rollback MODULE=core STEPS=1

# Rollback last 3 migrations from customer-mgmt
make migrate-rollback MODULE=customer-mgmt STEPS=3

# Rollback all migrations (dev only)
make migrate-rollback-all
```

### Phase 2B: Migration Reversibility Tests

#### CI Test Workflow

```bash
#!/bin/bash
# test/scripts/test-migration-reversibility.sh

set -e

echo "Testing migration reversibility..."

# Start test database
make test-db-start

# Apply all migrations
echo "1. Applying all migrations..."
make migrate

# Run integration tests (verify schema is correct)
echo "2. Running integration tests..."
make test-integration

# Rollback last migration from each namespace
NAMESPACES="core shared identity customer-mgmt order-mgmt billing warehouse scripting fiscal ui"

for namespace in $NAMESPACES; do
    echo "3. Rolling back $namespace..."
    make migrate-rollback MODULE=$namespace STEPS=1

    echo "4. Re-applying $namespace..."
    make migrate MODULE=$namespace

    echo "5. Running integration tests again..."
    make test-integration CONTEXT=$namespace

    if [ $? -ne 0 ]; then
        echo " FAILED: $namespace migration not reversible"
        exit 1
    fi
done

echo " SUCCESS: All migrations are reversible"
```

#### Add to CI

```yaml
# .github/workflows/migrations.yml (future)
name: Migration Tests
on: [push, pull_request]

jobs:
  test-reversibility:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.23"

      - name: Test migration reversibility
        run: |
          chmod +x test/scripts/test-migration-reversibility.sh
          ./test/scripts/test-migration-reversibility.sh
```

### Phase 2C: Migration Best Practices

#### Guidelines

**DO**:

-  Always create both up.sql AND down.sql
-  Test down migration locally before committing
-  Use transactions for data migrations
-  Document irreversible migrations
-  Keep migrations small and focused
-  Use safe column additions (NULL or DEFAULT)

**DON'T**:

-  Drop columns without backup
-  Delete data without archiving
-  Transform data irreversibly
-  Skip down migration implementation
-  Mix DDL and DML in one migration
-  Rename columns (drop + add instead)

#### Safe Patterns

**Adding Column (Reversible)**:

```sql
-- up.sql
ALTER TABLE customers
ADD COLUMN tier VARCHAR(50) DEFAULT 'standard' NOT NULL;

-- down.sql
ALTER TABLE customers
DROP COLUMN tier;
```

**Removing Column (Use Soft Delete)**:

```sql
-- up.sql (mark as deprecated)
COMMENT ON COLUMN customers.old_field IS 'DEPRECATED: Removed in v2.0. Use new_field instead.';

-- Phase 1: Deploy code that doesn't use old_field (wait 1 week)
-- Phase 2: Run this migration to drop column

-- down.sql
-- Cannot restore data! Document this:
-- IMPORTANT: This rollback will NOT restore data.
-- Restore from backup if needed.
ALTER TABLE customers
ADD COLUMN old_field VARCHAR(255);
```

**Data Transformation (Backup First)**:

```sql
-- up.sql
-- Backup data before transformation
CREATE TABLE customers_backup_20260122 AS
SELECT id, first_name, last_name FROM customers;

-- Transform
ALTER TABLE customers ADD COLUMN full_name VARCHAR(255);
UPDATE customers SET full_name = first_name || ' ' || last_name;

-- down.sql
-- Restore from backup
UPDATE customers c
SET first_name = b.first_name,
    last_name = b.last_name
FROM customers_backup_20260122 b
WHERE c.id = b.id;

ALTER TABLE customers DROP COLUMN full_name;
DROP TABLE customers_backup_20260122;
```

---

## Production Rollback Runbook

### Prerequisites

- [ ] Backup created and verified
- [ ] Rollback tested in staging
- [ ] Database maintenance window scheduled
- [ ] Rollback SQL reviewed
- [ ] Deployment can be reverted
- [ ] Team on standby

### Steps

#### 1. Verify Current State

```bash
# Check applied migrations
psql $DATABASE_URL -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 10;"

# Identify problematic migration
# Example: 20260120_000001_add_customer_tier.up.sql
```

#### 2. Stop Application

```bash
# Production
kubectl scale deployment promenade-api --replicas=0

# Or with downtime page
kubectl apply -f maintenance-mode.yaml
```

#### 3. Backup Database

```bash
# Full backup
pg_dump $DATABASE_URL > backup_before_rollback_$(date +%Y%m%d_%H%M%S).sql

# Verify backup
ls -lh backup_*.sql
```

#### 4. Rollback Migration

```bash
# Rollback single migration
make migrate-rollback MODULE=customer-mgmt STEPS=1

# Verify rollback
psql $DATABASE_URL -c "SELECT * FROM schema_migrations WHERE namespace='customer-mgmt';"
```

#### 5. Verify Database State

```bash
# Run integration tests
make test-integration

# Check critical queries
psql $DATABASE_URL -c "SELECT COUNT(*) FROM customers WHERE tier IS NOT NULL;"
```

#### 6. Redeploy Previous Version

```bash
# Revert to previous deployment
kubectl rollout undo deployment/promenade-api

# Or deploy specific version
kubectl set image deployment/promenade-api app=promenade:v0.8.5
```

#### 7. Restart Application

```bash
kubectl scale deployment promenade-api --replicas=3

# Monitor logs
kubectl logs -f deployment/promenade-api
```

#### 8. Verify Application

```bash
# Health check
curl https://api.promenade.example.com/health

# Smoke tests
curl https://api.promenade.example.com/api/v1/customers
```

### Rollback Failure Recovery

If rollback fails:

```bash
# 1. Restore from backup
psql $DATABASE_URL < backup_before_rollback_20260122_143000.sql

# 2. Re-apply all migrations
make migrate

# 3. Deploy current version
kubectl set image deployment/promenade-api app=promenade:latest
```

---

## Testing Checklist

Before merging migration PR:

- [ ] up.sql creates schema correctly
- [ ] down.sql reverses up.sql completely
- [ ] Tested locally: `make migrate` → `make migrate-rollback MODULE=X STEPS=1` → `make migrate`
- [ ] Integration tests pass after rollback + re-apply
- [ ] Data transformation preserves information (or backup created)
- [ ] Foreign key dependencies handled correctly
- [ ] Migration documented in PR description
- [ ] Reviewed by DBA (if data migration)

---

## Related Issues

### Issue 3.12: lib/pq vs pgx Migration

**Current**: Using `github.com/lib/pq` (maintenance mode)  
**Proposed**: Migrate to `pgx/v5` (Phase 2)

**See**: [pgx-migration-plan.md](pgx-migration-plan.md)

---

## Related Documentation

- [Migration README](../../migrations/README.md) - Migration structure
- [Database Strategy](database-strategy.md) - Overall database approach
- [Testing Patterns](testing-patterns.md) - Integration test patterns

---

**Status**: Planned (Phase 2)  
**Target Date**: Q2 2026  
**Owner**: Platform Team  
**Maintainer**: Promenade Team
