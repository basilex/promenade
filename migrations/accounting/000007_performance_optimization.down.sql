-- ============================================================================
-- Accounting Context: Performance Optimization - Rollback
-- ============================================================================

-- Drop functions
DROP FUNCTION IF EXISTS accounting_refresh_period_views(TEXT, TEXT);
DROP FUNCTION IF EXISTS accounting_refresh_materialized_views();

-- Drop materialized views
DROP MATERIALIZED VIEW IF EXISTS accounting_mv_budget_vs_actual;
DROP MATERIALIZED VIEW IF EXISTS accounting_mv_trial_balance;
DROP MATERIALIZED VIEW IF EXISTS accounting_mv_account_balances;

-- Drop composite indices
DROP INDEX IF EXISTS idx_accounting_bank_recon_items_recon_matched;
DROP INDEX IF EXISTS idx_accounting_bank_recon_org_status_date;
DROP INDEX IF EXISTS idx_accounting_cost_centers_org_active_type;
DROP INDEX IF EXISTS idx_accounting_budget_lines_budget_account;
DROP INDEX IF EXISTS idx_accounting_budgets_org_status_year;
DROP INDEX IF EXISTS idx_accounting_tax_codes_org_active_type;
DROP INDEX IF EXISTS idx_accounting_periods_org_status_dates;
DROP INDEX IF EXISTS idx_accounting_ledger_account_period;
DROP INDEX IF EXISTS idx_accounting_entry_lines_entry_account;
DROP INDEX IF EXISTS idx_accounting_entries_org_status_date;
DROP INDEX IF EXISTS idx_accounting_accounts_org_type_active;
DROP INDEX IF EXISTS idx_accounting_accounts_org_active;
