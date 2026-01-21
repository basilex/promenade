-- Drop indexes
DROP INDEX IF EXISTS idx_order_contracts_created_at;
DROP INDEX IF EXISTS idx_order_contracts_expires_at;
DROP INDEX IF EXISTS idx_order_contracts_status;
DROP INDEX IF EXISTS idx_order_contracts_customer_id;
DROP INDEX IF EXISTS idx_order_contracts_order_id;

-- Drop table
DROP TABLE IF EXISTS order_contracts;
