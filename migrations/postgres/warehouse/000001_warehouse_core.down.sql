-- ============================================================================
-- Warehouse Context: Rollback Core Schema
-- ============================================================================
-- Drop all warehouse tables in reverse dependency order
-- ============================================================================

-- 4. Drop stock movements (has foreign key to inventory)
DROP TABLE IF EXISTS warehouse_stock_movements CASCADE;

-- 3. Drop inventory
DROP TABLE IF EXISTS warehouse_inventory CASCADE;

-- 2. Drop locations (has self-reference)
DROP TABLE IF EXISTS warehouse_locations CASCADE;

-- 1. Drop products (base table)
DROP TABLE IF EXISTS warehouse_products CASCADE;
