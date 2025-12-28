-- Rollback Customer Management Context - Customers Table

-- Drop trigger
DROP TRIGGER IF EXISTS trg_customers_updated_at ON customer_mgmt_customers;

-- Drop indexes
DROP INDEX IF EXISTS idx_customers_email_unique;
DROP INDEX IF EXISTS idx_customers_tags;
DROP INDEX IF EXISTS idx_customers_created_at;
DROP INDEX IF EXISTS idx_customers_assigned_to;
DROP INDEX IF EXISTS idx_customers_tier;
DROP INDEX IF EXISTS idx_customers_status;
DROP INDEX IF EXISTS idx_customers_email;
DROP INDEX IF EXISTS idx_customers_company_id;
DROP INDEX IF EXISTS idx_customers_user_id;

-- Drop table
DROP TABLE IF EXISTS customer_mgmt_customers;

-- Drop enums
DROP TYPE IF EXISTS customer_tier;
DROP TYPE IF EXISTS customer_status;
