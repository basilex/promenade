# Database Conventions Guide

**Comprehensive database naming standards** for tables, columns, indexes, constraints, and migrations in Promenade Platform.

---

## Table of Contents

- [Table Naming](#table-naming)
- [Column Naming](#column-naming)
- [Index Naming](#index-naming)
- [Constraint Naming](#constraint-naming)
- [Migration Naming](#migration-naming)
- [SQL Conventions](#sql-conventions)
- [Examples](#examples)

---

## Table Naming

### Pattern: `<context>_<aggregate>`

**Rule**: Table names follow the pattern `{context}_{aggregate}` in `snake_case` with NO `mgmt` suffix.

**Examples**:

```sql
-- Identity Context (11 tables)
identity_users
identity_contacts
identity_profiles
identity_roles
identity_permissions
identity_user_roles

-- Shared Context (4 tables)
shared_countries
shared_currencies
shared_languages
shared_timezones

-- Customer Management Context (4 tables)
customer_customers     -- NOT customer_mgmt_customers 
customer_companies
customer_deals
customer_interactions

-- Order Management Context (2 tables)
order_orders          -- NOT order_mgmt_orders 
order_lines           -- NOT order_mgmt_order_lines 
```

### Rationale

**Consistency**: 15 existing tables (Identity + Shared) already use `{context}_{aggregate}` pattern  
**Shorter Names**: Easier to type, read, and maintain  
**Clear Context**: Context prefix provides namespace isolation  
**No Ambiguity**: Aggregate name is clear without `mgmt` suffix

### Migration Strategy

When changing table names (e.g., from `customer_mgmt_customers` to `customer_customers`):

1. **Edit existing migrations in-place** (no production data yet)
2. Drop and recreate databases
3. Update all SQL queries in Go code
4. Update foreign key references
5. Run all tests to verify

**See**: [Table Naming Strategy Document](../reference/table-naming-strategy.md)

---

## Column Naming

### Pattern: `snake_case`

**Rule**: All column names use `snake_case` with descriptive names.

**Standard Columns**:

```sql
-- Primary Key
id UUID PRIMARY KEY DEFAULT uuid_v7()

-- Foreign Keys (with {table}_id pattern)
user_id UUID NOT NULL
customer_id UUID NOT NULL
company_id UUID NOT NULL

-- Lifecycle Timestamps
created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
deleted_at TIMESTAMP WITH TIME ZONE NULL  -- Soft delete

-- Status/State
status VARCHAR(20) NOT NULL DEFAULT 'active'
is_active BOOLEAN NOT NULL DEFAULT TRUE
is_verified BOOLEAN NOT NULL DEFAULT FALSE
is_primary BOOLEAN NOT NULL DEFAULT FALSE

-- Metadata
tags JSONB DEFAULT '[]'::jsonb
metadata JSONB DEFAULT '{}'::jsonb
```

### Column Type Patterns

| Type              | Pattern                   | Example                           |
|-------------------|---------------------------|-----------------------------------|
| **ID**            | `UUID DEFAULT uuid_v7()`  | `id UUID PRIMARY KEY DEFAULT uuid_v7()` |
| **Foreign Key**   | `{table}_id UUID`         | `user_id UUID NOT NULL`           |
| **Status**        | `VARCHAR(20)`             | `status VARCHAR(20) NOT NULL`     |
| **Boolean Flag**  | `is_{condition} BOOLEAN`  | `is_active BOOLEAN DEFAULT TRUE`  |
| **Timestamp**     | `TIMESTAMP WITH TIME ZONE`| `created_at TIMESTAMP WITH TIME ZONE` |
| **Money**         | `DECIMAL(15, 2)`          | `total_amount DECIMAL(15, 2)`     |
| **Currency**      | `VARCHAR(3)`              | `currency VARCHAR(3) NOT NULL`    |
| **JSONB**         | `JSONB`                   | `tags JSONB DEFAULT '[]'::jsonb`  |
| **Email**         | `VARCHAR(255)`            | `email VARCHAR(255) NOT NULL`     |
| **Name**          | `VARCHAR(100)`            | `name VARCHAR(100) NOT NULL`      |

### Naming Rules

**Foreign Keys**:
```sql
-- Pattern: {referenced_table}_id
user_id UUID NOT NULL                    -- References identity_users
customer_id UUID NOT NULL                -- References customer_customers
company_id UUID NOT NULL                 -- References customer_companies
assigned_to UUID NULL                    -- References identity_users (sales rep)
```

**Boolean Flags**:
```sql
-- Pattern: is_{condition}
is_active BOOLEAN NOT NULL DEFAULT TRUE
is_verified BOOLEAN NOT NULL DEFAULT FALSE
is_primary BOOLEAN NOT NULL DEFAULT FALSE
is_public BOOLEAN NOT NULL DEFAULT FALSE
is_deleted BOOLEAN NOT NULL DEFAULT FALSE
```

**Status/State Columns**:
```sql
-- Use descriptive names
status VARCHAR(20) NOT NULL              -- Generic status
customer_status VARCHAR(20) NOT NULL     -- Specific entity status
order_status VARCHAR(20) NOT NULL
payment_status VARCHAR(20) NOT NULL
```

**Timestamps**:
```sql
-- Standard lifecycle timestamps
created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
deleted_at TIMESTAMP WITH TIME ZONE NULL

-- Business timestamps
verified_at TIMESTAMP WITH TIME ZONE NULL
confirmed_at TIMESTAMP WITH TIME ZONE NULL
fulfilled_at TIMESTAMP WITH TIME ZONE NULL
cancelled_at TIMESTAMP WITH TIME ZONE NULL
```

---

## Index Naming

### Pattern: `idx_<table>_<column(s)>`

**Single Column Index**:
```sql
-- Pattern: idx_{table}_{column}
CREATE INDEX idx_customers_email ON customer_customers (email) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customers_status ON customer_customers (status) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_deals_stage ON customer_deals (stage) 
    WHERE deleted_at IS NULL;
```

**Multi-Column Index**:
```sql
-- Pattern: idx_{table}_{column1}_{column2}
CREATE INDEX idx_customers_status_tier ON customer_customers (status, tier) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_orders_customer_status ON order_orders (customer_id, status) 
    WHERE deleted_at IS NULL;
```

**Unique Index**:
```sql
-- Pattern: idx_{table}_{column}_unique
CREATE UNIQUE INDEX idx_customers_email_unique ON customer_customers (email) 
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_users_email_unique ON identity_users (email) 
    WHERE deleted_at IS NULL;
```

**Partial Index** (with WHERE clause):
```sql
-- Active records only
CREATE INDEX idx_customers_active ON customer_customers (status) 
    WHERE deleted_at IS NULL AND is_active = TRUE;

-- Soft delete tracking
CREATE INDEX idx_customers_deleted_at ON customer_customers (deleted_at) 
    WHERE deleted_at IS NOT NULL;
```

**GIN Index** (for JSONB):
```sql
-- Pattern: idx_{table}_{column}_gin
CREATE INDEX idx_customers_tags_gin ON customer_customers USING GIN (tags);

CREATE INDEX idx_interactions_attendees_gin ON customer_interactions USING GIN (attendees);
```

### Index Strategy

**Always Index**:
- Primary keys (automatic)
- Foreign keys
- Unique constraints
- Columns in WHERE clauses
- Columns in ORDER BY clauses
- JSONB columns with GIN

**Partial Indexes for Soft Delete**:
```sql
-- All queries filter by deleted_at IS NULL
CREATE INDEX idx_{table}_{column} ON {table} ({column}) 
    WHERE deleted_at IS NULL;
```

---

## Constraint Naming

### Pattern: `{type}_{table}_{column(s)}`

**Primary Key**:
```sql
-- Automatic: {table}_pkey
-- customer_customers_pkey
CONSTRAINT customer_customers_pkey PRIMARY KEY (id)
```

**Foreign Key**:
```sql
-- Pattern: fk_{table}_{ref_table}
CONSTRAINT fk_deals_customers FOREIGN KEY (customer_id) 
    REFERENCES customer_customers(id);

CONSTRAINT fk_orders_customers FOREIGN KEY (customer_id) 
    REFERENCES customer_customers(id);

CONSTRAINT fk_interactions_companies FOREIGN KEY (company_id) 
    REFERENCES customer_companies(id);
```

**Check Constraint**:
```sql
-- Pattern: chk_{table}_{description}
CONSTRAINT chk_customer_status CHECK (
    status IN ('lead', 'prospect', 'customer', 'churned')
)

CONSTRAINT chk_deal_amount_positive CHECK (amount >= 0)

CONSTRAINT chk_order_quantity_positive CHECK (quantity > 0)

CONSTRAINT chk_currency_length CHECK (LENGTH(currency) = 3)
```

**Unique Constraint**:
```sql
-- Pattern: uq_{table}_{column}
CONSTRAINT uq_customers_email UNIQUE (email)

CONSTRAINT uq_users_email UNIQUE (email)

CONSTRAINT uq_orders_number UNIQUE (order_number)
```

---

## Migration Naming

### Pattern: `{number}_{description}.{up|down}.sql`

**Namespace-Based Organization**:

```
migrations/
  core/
      000001_extensions.up.sql
      000001_extensions.down.sql
  shared/
      000001_reference_data.up.sql
      000001_reference_data.down.sql
  identity/
      000001_users.up.sql
      000001_users.down.sql
      000002_authentication.up.sql
      000002_authentication.down.sql
  customer-mgmt/
      000001_customers.up.sql
      000001_customers.down.sql
      000002_companies.up.sql
      000002_companies.down.sql
  order-mgmt/
      000001_orders.up.sql
      000001_orders.down.sql
```

**Migration Order**:
1. `core` - Extensions (UUID v7, pgcrypto)
2. `shared` - Reference data (countries, currencies, languages, timezones)
3. `identity` - Users, authentication, RBAC
4. `customer-mgmt` - Customer management tables
5. `order-mgmt` - Order management tables

**Naming Rules**:
- Sequential numbers: `000001`, `000002`, `000003`
- Descriptive names: `users`, `authentication`, `customers`
- Always create both `.up.sql` and `.down.sql`
- Use underscores for multi-word names: `reference_data`, `order_lines`

---

## SQL Conventions

### CREATE TABLE Template

```sql
-- ============================================================================
-- {Context} Context: {Aggregate} Table
-- ============================================================================

-- Enum types (if needed)
CREATE TYPE customer_status AS ENUM ('lead', 'prospect', 'customer', 'churned');
CREATE TYPE customer_tier AS ENUM ('free', 'basic', 'pro', 'enterprise');

-- Table creation
CREATE TABLE IF NOT EXISTS customer_customers (
    -- Identity
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- Core Fields
    email VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    status customer_status NOT NULL DEFAULT 'lead',
    tier customer_tier NOT NULL DEFAULT 'free',
    
    -- Metadata
    tags JSONB DEFAULT '[]'::jsonb,
    
    -- Relationships (Foreign Keys)
    user_id UUID NULL,
    company_id UUID NULL,
    assigned_to UUID NULL,  -- Sales rep
    
    -- Lifecycle Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    
    -- Constraints
    CONSTRAINT chk_customer_status CHECK (
        status IN ('lead', 'prospect', 'customer', 'churned')
    ),
    CONSTRAINT chk_customer_tier CHECK (
        tier IN ('free', 'basic', 'pro', 'enterprise')
    )
);

-- Indexes
CREATE UNIQUE INDEX idx_customers_email_unique ON customer_customers (email) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_status ON customer_customers (status) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_tier ON customer_customers (tier) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_user_id ON customer_customers (user_id) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_assigned_to ON customer_customers (assigned_to) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_tags_gin ON customer_customers USING GIN (tags);
CREATE INDEX idx_customers_deleted_at ON customer_customers (deleted_at) 
    WHERE deleted_at IS NOT NULL;

-- Trigger for updated_at
CREATE TRIGGER update_customers_updated_at
    BEFORE UPDATE ON customer_customers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE customer_customers IS 'Customer aggregate: manages customer lifecycle from lead to churned';
COMMENT ON COLUMN customer_customers.status IS 'Customer status: lead → prospect → customer → churned';
COMMENT ON COLUMN customer_customers.tier IS 'Customer tier: free → basic → pro → enterprise';
COMMENT ON COLUMN customer_customers.tags IS 'Flexible JSONB tags for segmentation';
```

### DROP TABLE Template

```sql
-- Rollback migration
DROP TRIGGER IF EXISTS update_customers_updated_at ON customer_customers;
DROP INDEX IF EXISTS idx_customers_deleted_at;
DROP INDEX IF EXISTS idx_customers_tags_gin;
DROP INDEX IF EXISTS idx_customers_assigned_to;
DROP INDEX IF EXISTS idx_customers_user_id;
DROP INDEX IF EXISTS idx_customers_tier;
DROP INDEX IF EXISTS idx_customers_status;
DROP INDEX IF EXISTS idx_customers_email_unique;
DROP TABLE IF EXISTS customer_customers CASCADE;
DROP TYPE IF EXISTS customer_tier;
DROP TYPE IF EXISTS customer_status;
```

---

## Examples

### Complete Migration Example

**File**: `migrations/customer-mgmt/000001_customers.up.sql`

```sql
-- Customer Management Context - Customers Table
-- Aggregate: Customer (Lead → Prospect → Customer → Churned lifecycle)

-- Customer Status enum
CREATE TYPE customer_status AS ENUM ('lead', 'prospect', 'customer', 'churned');

-- Customer Tier enum
CREATE TYPE customer_tier AS ENUM ('free', 'basic', 'pro', 'enterprise');

-- Customers table
CREATE TABLE IF NOT EXISTS customer_customers (
    -- Identity
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- Core customer information
    email VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(50) NULL,
    status customer_status NOT NULL DEFAULT 'lead',
    tier customer_tier NOT NULL DEFAULT 'free',
    
    -- Business fields
    source VARCHAR(50) NULL,
    tags JSONB DEFAULT '[]'::jsonb,
    lifetime_value DECIMAL(15, 2) DEFAULT 0,
    
    -- Relationships
    user_id UUID NULL,
    company_id UUID NULL,
    assigned_to UUID NULL,
    
    -- Lifecycle timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    
    -- Constraints
    CONSTRAINT chk_customer_status CHECK (
        status IN ('lead', 'prospect', 'customer', 'churned')
    ),
    CONSTRAINT chk_customer_tier CHECK (
        tier IN ('free', 'basic', 'pro', 'enterprise')
    ),
    CONSTRAINT chk_lifetime_value_positive CHECK (lifetime_value >= 0)
);

-- Indexes for performance
CREATE UNIQUE INDEX idx_customers_email_unique ON customer_customers (email) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_status ON customer_customers (status) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_tier ON customer_customers (tier) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_user_id ON customer_customers (user_id) 
    WHERE deleted_at IS NULL AND user_id IS NOT NULL;
CREATE INDEX idx_customers_company_id ON customer_customers (company_id) 
    WHERE deleted_at IS NULL AND company_id IS NOT NULL;
CREATE INDEX idx_customers_assigned_to ON customer_customers (assigned_to) 
    WHERE deleted_at IS NULL AND assigned_to IS NOT NULL;
CREATE INDEX idx_customers_tags_gin ON customer_customers USING GIN (tags);
CREATE INDEX idx_customers_created_at ON customer_customers (created_at DESC) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_deleted_at ON customer_customers (deleted_at) 
    WHERE deleted_at IS NOT NULL;

-- Trigger for updated_at
CREATE TRIGGER update_customers_updated_at
    BEFORE UPDATE ON customer_customers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE customer_customers IS 'Customer aggregate: manages customer lifecycle from lead to churned';
COMMENT ON COLUMN customer_customers.status IS 'Customer status: lead → prospect → customer → churned';
COMMENT ON COLUMN customer_customers.tier IS 'Customer tier: free → basic → pro → enterprise';
COMMENT ON COLUMN customer_customers.tags IS 'JSONB array for flexible customer segmentation';
COMMENT ON COLUMN customer_customers.user_id IS 'Optional link to identity_users for authenticated customers';
COMMENT ON COLUMN customer_customers.assigned_to IS 'Sales rep assigned to this customer (identity_users.id)';
```

---

## Anti-Patterns to Avoid

###  DON'T

```sql
-- DON'T: Use mgmt suffix
CREATE TABLE customer_mgmt_customers   -- Use customer_customers 

-- DON'T: Use camelCase
CREATE TABLE customerCustomers         -- Use customer_customers 

-- DON'T: Abbreviate
CREATE TABLE cust_customers            -- Use customer_customers 
CREATE INDEX idx_cust_email            -- Use idx_customers_email 

-- DON'T: Use generic names
CREATE INDEX idx_email                 -- Use idx_customers_email 
CREATE INDEX idx_status                -- Use idx_customers_status 

-- DON'T: Forget WHERE clause for soft delete
CREATE INDEX idx_customers_status ON customer_customers (status);
-- Use with WHERE deleted_at IS NULL 

-- DON'T: Use TEXT for fixed-length strings
currency TEXT                          -- Use VARCHAR(3) 
email TEXT                             -- Use VARCHAR(255) 
```

###  DO

```sql
-- DO: Use {context}_{aggregate} pattern
CREATE TABLE customer_customers
CREATE TABLE order_orders
CREATE TABLE order_lines

-- DO: Use snake_case
CREATE TABLE customer_customers
CREATE INDEX idx_customers_email

-- DO: Use full words
CREATE TABLE customer_customers        -- NOT cust_customers
CREATE INDEX idx_customers_email       -- NOT idx_cust_email

-- DO: Include table name in index
CREATE INDEX idx_customers_email
CREATE INDEX idx_customers_status

-- DO: Add WHERE clause for soft delete
CREATE INDEX idx_customers_status ON customer_customers (status) 
    WHERE deleted_at IS NULL;

-- DO: Use appropriate types
currency VARCHAR(3)
email VARCHAR(255)
```

---

## Summary Table

| Component       | Pattern                          | Example                           |
|-----------------|----------------------------------|-----------------------------------|
| **Table**       | `<context>_<aggregate>`          | `customer_customers`              |
| **Column**      | `snake_case`                     | `user_id`, `created_at`           |
| **Index**       | `idx_<table>_<column>`           | `idx_customers_email`             |
| **FK**          | `fk_<table>_<ref_table>`         | `fk_orders_customers`             |
| **Check**       | `chk_<table>_<description>`      | `chk_customer_status`             |
| **Unique**      | `uq_<table>_<column>`            | `uq_customers_email`              |
| **Migration**   | `{number}_{name}.{up\|down}.sql` | `000001_customers.up.sql`         |

---

**See Also**:
- [Naming Conventions Guide](naming-conventions.md) - Files, directories, Go code
- [Architecture Patterns Guide](architecture-patterns.md) - Repository, UseCase, Handler patterns
- [Table Naming Strategy](../reference/table-naming-strategy.md) - Migration decisions

---

**Last Updated**: January 1, 2026  
**Status**: Production Standard  
**Maintainer**: Promenade Team
