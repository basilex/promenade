-- ============================================================================
-- Order Management Context: Rollback Core Schema
-- ============================================================================
-- Drop all order management tables in reverse dependency order
-- ============================================================================

-- 3. Drop contracts (has FK to orders)
DROP TABLE IF EXISTS order_contracts CASCADE;

-- 2. Drop order lines (has FK to orders)
DROP TABLE IF EXISTS order_lines CASCADE;

-- 1. Drop orders (base table)
DROP TABLE IF EXISTS order_orders CASCADE;
