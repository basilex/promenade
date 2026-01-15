# Table Naming Strategy - Analysis & Decision

**Date**: 2026-01-01  
**Status**:  CRITICAL - Requires decision before further development

---

## Current State (Inconsistent)

###  Consistent Contexts

**Identity Context** (11 tables):
```
identity_users
identity_contacts
identity_profiles
identity_permissions
identity_roles
identity_role_permissions
identity_user_roles
identity_user_sessions
identity_password_reset_tokens
identity_email_verification_tokens
identity_login_attempts
```
Pattern: `identity_<aggregate>`

**Shared Context** (4 tables):
```
shared_countries
shared_currencies
shared_languages
shared_timezones
```
Pattern: `shared_<aggregate>`

###  Inconsistent Contexts

**Customer Management Context** (4 tables):
```
customer_mgmt_customers    ← includes mgmt (migration 000001)
customer_deals             ← missing mgmt (migration 000003)
customer_companies         ← missing mgmt (migration 000002)
customer_interactions      ← missing mgmt (migration 000004)
```
Pattern: **INCONSISTENT**

**Order Management Context** (2 tables):
```
order_mgmt_orders
order_mgmt_order_lines
```
Pattern: `order_mgmt_<aggregate>`

---

## Problem Analysis

### 1. Why Inconsistency Happened?

Looking at migration order:
- `000001_customers.up.sql` - created `customer_mgmt_customers` (СТАРИЙ підхід з mgmt)
- `000002_companies.up.sql` - created `customer_companies` (НОВИЙ підхід без mgmt)
- `000003_deals.up.sql` - created `customer_deals` (НОВИЙ підхід без mgmt)
- `000004_interactions.up.sql` - created `customer_interactions` (НОВИЙ підхід без mgmt)

**Root cause**: Migration 000001 was created BEFORE we established naming convention. Migrations 000002-000004 followed new pattern but 000001 was never updated.

### 2. Impact

**Code affected**:
- All SQL queries in `customer` aggregate reference `customer_mgmt_customers`
- All SQL queries in other aggregates reference `customer_*` (no mgmt)
- Analytics queries (just implemented) use mixed approach

**Migration count**: 4 migrations in customer-mgmt context

---

## Options Analysis

### Option 1: `<context>_<aggregate>` (SHORT - without mgmt)  RECOMMENDED

**Pattern**:
```
customer_customers
customer_deals
customer_companies
customer_interactions

order_orders
order_lines

identity_users      (already correct)
shared_countries    (already correct)
```

**Pros**:
-  Shorter, cleaner
-  Consistent with Identity & Shared contexts
-  Logical: `customer_deals` = deals in customer context
-  No redundancy (`customer_mgmt` sounds like "customer management management")
-  Future-proof: easier to type, less DB storage

**Cons**:
-  Requires renaming 3 tables:
  - `customer_mgmt_customers` → `customer_customers`
  - `order_mgmt_orders` → `order_orders`
  - `order_mgmt_order_lines` → `order_lines`
-  Breaking change (but early in development)

**Migration effort**: 3 tables × 1 migration each = **3 new migrations**

---

### Option 2: `<context_full>_<aggregate>` (FULL - with mgmt)

**Pattern**:
```
customer_mgmt_customers
customer_mgmt_deals
customer_mgmt_companies
customer_mgmt_interactions

order_mgmt_orders
order_mgmt_lines

identity_users              (exception)
shared_countries            (exception)
```

**Pros**:
-  Keeps `customer_mgmt_customers` and `order_mgmt_*` as-is
-  No breaking changes for existing tables

**Cons**:
-  Longer, more verbose
-  Inconsistent with Identity & Shared (11+4=15 tables without mgmt)
-  Redundant (`customer_mgmt` = customer management context)
-  Requires renaming 3 customer tables:
  - `customer_deals` → `customer_mgmt_deals`
  - `customer_companies` → `customer_mgmt_companies`
  - `customer_interactions` → `customer_mgmt_interactions`

**Migration effort**: 3 tables × 1 migration each = **3 new migrations**

**Result**: Same migration effort but WORSE naming convention

---

### Option 3: `<short_context>_<aggregate>` (ABBR - abbreviated)

**Pattern**:
```
crm_customers
crm_deals
crm_companies
crm_interactions

orders_orders       (weird duplication)
orders_lines        (or just "orders" table?)

identity_users      (already correct)
shared_countries    (already correct)
```

**Pros**:
-  Very short (`crm_` instead of `customer_`)
-  Domain-driven (CRM is ubiquitous language)

**Cons**:
-  Breaks established pattern (Identity & Shared use full context name)
-  `orders_orders` looks weird
-  Less clear for new developers (what is "crm"?)
-  Requires renaming ALL 6 tables in customer-mgmt and order-mgmt

**Migration effort**: 6 tables × 1 migration each = **6 new migrations**

---

## Recommendation: Option 1 

### Rationale

1. **Consistency First**: 15 tables (Identity + Shared) already use `<context>_<aggregate>` pattern
2. **Future Growth**: We'll add more contexts (Billing, Warehouse) - shorter is better
3. **Developer Experience**: Less typing, cleaner queries
4. **Database Performance**: Shorter names = less storage in indexes
5. **Same Migration Effort**: Options 1 & 2 both need 3 migrations

### Proposed Final State

```sql
-- Identity Context (11 tables)  No changes
identity_users
identity_contacts
identity_profiles
identity_permissions
identity_roles
identity_role_permissions
identity_user_roles
identity_user_sessions
identity_password_reset_tokens
identity_email_verification_tokens
identity_login_attempts

-- Shared Context (4 tables)  No changes
shared_countries
shared_currencies
shared_languages
shared_timezones

-- Customer Management Context (4 tables)  1 rename needed
customer_customers         ← rename from customer_mgmt_customers
customer_deals              already correct
customer_companies          already correct
customer_interactions       already correct

-- Order Management Context (2 tables)  2 renames needed
order_orders               ← rename from order_mgmt_orders
order_lines                ← rename from order_mgmt_order_lines

-- Future: Billing Context (0 tables)  Will follow pattern
billing_invoices
billing_payments
billing_subscriptions

-- Future: Warehouse Context (0 tables)  Will follow pattern
warehouse_inventory
warehouse_stock_movements
```

**Total**: 21 current tables, 3 renames needed

---

## Migration Plan

### Phase 1: Rename `customer_mgmt_customers` → `customer_customers`

**New migration**: `migrations/customer-mgmt/000005_rename_customers_table.up.sql`

```sql
-- Rename table
ALTER TABLE customer_mgmt_customers RENAME TO customer_customers;

-- Update indexes (PostgreSQL auto-renames, but explicit for clarity)
-- customer_mgmt_customers_pkey → customer_customers_pkey (auto)
-- idx_customer_mgmt_customers_email → idx_customer_customers_email
ALTER INDEX IF EXISTS idx_customer_mgmt_customers_email 
    RENAME TO idx_customer_customers_email;

-- ... (rename all indexes)
```

**Down migration**: `migrations/customer-mgmt/000005_rename_customers_table.down.sql`

```sql
ALTER TABLE customer_customers RENAME TO customer_mgmt_customers;
-- ... (reverse index renames)
```

**Code changes**:
- Update all SQL queries in `internal/contexts/customer-mgmt/customer/adapter/repository/postgres/`
- Update analytics queries in `internal/contexts/customer-mgmt/analytics/usecase.go`
- Update tests

**Estimated effort**: 2 hours

---

### Phase 2: Rename `order_mgmt_orders` → `order_orders`

**New migration**: `migrations/order-mgmt/000002_rename_orders_table.up.sql`

```sql
ALTER TABLE order_mgmt_orders RENAME TO order_orders;
-- ... (rename indexes)
```

**Code changes**:
- Update all SQL queries in `internal/contexts/order-mgmt/order/adapter/repository/postgres/`
- Update tests

**Estimated effort**: 1 hour

---

### Phase 3: Rename `order_mgmt_order_lines` → `order_lines`

**New migration**: `migrations/order-mgmt/000003_rename_order_lines_table.up.sql`

```sql
ALTER TABLE order_mgmt_order_lines RENAME TO order_lines;
-- ... (rename indexes)
```

**Code changes**:
- Update all SQL queries in order aggregate
- Update tests

**Estimated effort**: 1 hour

---

## Total Effort

- **3 new migrations** (up + down for each)
- **~4-6 hours** development time
- **Breaking change**: YES (but early in development, no production data)
- **Test coverage**: All existing tests will catch issues

---

## Decision Checkpoint

### Questions to Answer

1. **Do we proceed with Option 1** (`<context>_<aggregate>` without mgmt)?
   - [ ] YES - Rename 3 tables now (4-6 hours work)
   - [ ] NO - Keep current inconsistency (technical debt)

2. **When to execute?**
   - [ ] NOW - Before more code depends on current names
   - [ ] LATER - After Phase 5 (more migrations to manage)

3. **Documentation update?**
   - [ ] YES - Update all READMEs with new table names
   - [ ] Create `.github/copilot-instructions.md` section on naming

---

## Related Files

- Migrations: `migrations/customer-mgmt/000001_customers.up.sql`
- Migrations: `migrations/order-mgmt/000001_orders.up.sql`
- Code: `internal/contexts/customer-mgmt/customer/adapter/repository/postgres/*.go`
- Code: `internal/contexts/customer-mgmt/analytics/usecase.go`
- Code: `internal/contexts/order-mgmt/order/adapter/repository/postgres/*.go`
- Tests: All integration tests for customer and order contexts

---

**Status**: Awaiting decision  
**Priority**:  HIGH - Should decide before next aggregate implementation  
**Impact**: Breaking change but fixable with migrations
