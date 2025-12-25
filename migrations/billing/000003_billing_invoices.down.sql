-- Migration: Drop billing_invoices table
-- Description: Rollback invoices table

DROP INDEX IF EXISTS idx_billing_invoices_deleted_at;
DROP INDEX IF EXISTS idx_billing_invoices_due_date;
DROP INDEX IF EXISTS idx_billing_invoices_status;
DROP INDEX IF EXISTS idx_billing_invoices_invoice_number;
DROP INDEX IF EXISTS idx_billing_invoices_subscription_id;

DROP TABLE IF EXISTS billing_invoices;
