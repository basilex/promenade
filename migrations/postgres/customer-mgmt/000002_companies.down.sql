-- ============================================================================
-- Migration Rollback: 000002_companies
-- Description: Drop customer_companies table and related objects
-- Context: Customer Management
-- ============================================================================

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_companies_updated_at ON customer_companies;

-- Drop indexes
DROP INDEX IF EXISTS idx_companies_type_size_active;
DROP INDEX IF EXISTS idx_companies_updated_at;
DROP INDEX IF EXISTS idx_companies_created_at;
DROP INDEX IF EXISTS idx_companies_deleted_at;
DROP INDEX IF EXISTS idx_companies_parent_company_id;
DROP INDEX IF EXISTS idx_companies_size;
DROP INDEX IF EXISTS idx_companies_industry;
DROP INDEX IF EXISTS idx_companies_tax_id;
DROP INDEX IF EXISTS idx_companies_name_unique;

-- Drop table
DROP TABLE IF EXISTS customer_companies;

-- Drop enums
DROP TYPE IF EXISTS company_size;
DROP TYPE IF EXISTS company_type;
