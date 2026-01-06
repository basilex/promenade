-- Migration: add_stock_movements_table
-- Context: warehouse
-- Created: 2026-01-06 09:12:25

-- Drop indexes first
DROP INDEX IF EXISTS idx_stock_movements_created_at;
DROP INDEX IF EXISTS idx_stock_movements_movement_date;
DROP INDEX IF EXISTS idx_stock_movements_reference;
DROP INDEX IF EXISTS idx_stock_movements_type;
DROP INDEX IF EXISTS idx_stock_movements_inventory_id;

-- Drop table
DROP TABLE IF EXISTS warehouse_stock_movements;
