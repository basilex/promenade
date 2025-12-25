-- Migration: Drop billing_payments table
-- Description: Rollback payments table

DROP INDEX IF EXISTS idx_billing_payments_user_status;
DROP INDEX IF EXISTS idx_billing_payments_deleted_at;
DROP INDEX IF EXISTS idx_billing_payments_payment_method;
DROP INDEX IF EXISTS idx_billing_payments_status;
DROP INDEX IF EXISTS idx_billing_payments_transaction_id;
DROP INDEX IF EXISTS idx_billing_payments_user_id;
DROP INDEX IF EXISTS idx_billing_payments_invoice_id;

DROP TABLE IF EXISTS billing_payments;
