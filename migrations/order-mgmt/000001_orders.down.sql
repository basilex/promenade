-- Order Management Context: Rollback orders and order lines tables
-- Migration: 000001_orders.down.sql

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS order_mgmt_order_lines CASCADE;
DROP TABLE IF EXISTS order_mgmt_orders CASCADE;
