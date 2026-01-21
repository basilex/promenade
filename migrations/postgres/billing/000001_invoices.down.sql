-- Drop indexes first
DROP INDEX IF EXISTS idx_billing_invoice_lines_invoice_id;
DROP INDEX IF EXISTS idx_billing_invoices_invoice_no;
DROP INDEX IF EXISTS idx_billing_invoices_deleted_at;
DROP INDEX IF EXISTS idx_billing_invoices_paid_date;
DROP INDEX IF EXISTS idx_billing_invoices_due_date;
DROP INDEX IF EXISTS idx_billing_invoices_status;
DROP INDEX IF EXISTS idx_billing_invoices_order_id;
DROP INDEX IF EXISTS idx_billing_invoices_customer_id;

-- Drop tables
DROP TABLE IF EXISTS billing_invoice_lines;
DROP TABLE IF EXISTS billing_invoices;
