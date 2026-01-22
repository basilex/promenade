-- ============================================================================
-- Billing Context: Rollback Core Schema
-- ============================================================================
-- Drop all billing tables in reverse dependency order
-- ============================================================================

-- 3. Drop subscriptions (no dependencies)
DROP TABLE IF EXISTS billing_subscriptions CASCADE;

-- 2. Drop payments (has FK to invoices and customers)
DROP TABLE IF EXISTS billing_payments CASCADE;

-- 1. Drop invoice lines and invoices
DROP TABLE IF EXISTS billing_invoice_lines CASCADE;
DROP TABLE IF EXISTS billing_invoices CASCADE;
