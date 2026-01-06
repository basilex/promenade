-- ============================================================================
-- Rollback: Warehouse Context - Products
-- ============================================================================
-- Drop Products table and related indexes
-- ============================================================================

-- Drop indexes
DROP INDEX IF EXISTS idx_warehouse_products_deleted_at;
DROP INDEX IF EXISTS idx_warehouse_products_updated_at;
DROP INDEX IF EXISTS idx_warehouse_products_created_at;
DROP INDEX IF EXISTS idx_warehouse_products_search;
DROP INDEX IF EXISTS idx_warehouse_products_brand;
DROP INDEX IF EXISTS idx_warehouse_products_category;
DROP INDEX IF EXISTS idx_warehouse_products_is_active;
DROP INDEX IF EXISTS idx_warehouse_products_status;
DROP INDEX IF EXISTS idx_warehouse_products_name;
DROP INDEX IF EXISTS idx_warehouse_products_sku;

-- Drop table
DROP TABLE IF EXISTS warehouse_products;
