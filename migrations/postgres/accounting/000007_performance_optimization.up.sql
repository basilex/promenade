-- ============================================================================
-- Accounting Context: Performance Optimization
-- ============================================================================
-- Composite indices, materialized views, and query optimization
-- ============================================================================

-- ============================================================================
-- Composite Indices for Common Query Patterns
-- ============================================================================

-- Chart of Accounts: organization + status queries
CREATE INDEX IF NOT EXISTS idx_accounting_accounts_org_active 
    ON accounting_chart_of_accounts(organization_id, is_active) 
    WHERE deleted_at IS NULL;

-- Chart of Accounts: organization + type + active (for filtering)
CREATE INDEX IF NOT EXISTS idx_accounting_accounts_org_type_active 
    ON accounting_chart_of_accounts(organization_id, type, is_active) 
    WHERE deleted_at IS NULL;

-- Journal Entries: organization + status + date (most common query)
CREATE INDEX IF NOT EXISTS idx_accounting_entries_org_status_date 
    ON accounting_journal_entries(organization_id, status, entry_date DESC) 
    WHERE deleted_at IS NULL;

-- Journal Entry Lines: entry + account (for balance calculations)
CREATE INDEX IF NOT EXISTS idx_accounting_entry_lines_entry_account 
    ON accounting_journal_entry_lines(journal_entry_id, account_id) 
    WHERE deleted_at IS NULL;

-- Ledger: account + period (primary lookup pattern)
CREATE INDEX IF NOT EXISTS idx_accounting_ledger_account_period 
    ON accounting_ledger(account_id, fiscal_period_id, period_code) 
    WHERE deleted_at IS NULL;

-- Fiscal Periods: organization + status + dates (for validation)
CREATE INDEX IF NOT EXISTS idx_accounting_periods_org_status_dates 
    ON accounting_fiscal_periods(organization_id, status, start_date, end_date) 
    WHERE deleted_at IS NULL;

-- Tax Codes: organization + active + type (for invoice calculations)
CREATE INDEX IF NOT EXISTS idx_accounting_tax_codes_org_active_type 
    ON accounting_tax_codes(organization_id, is_active, tax_type) 
    WHERE deleted_at IS NULL;

-- Budgets: organization + status + fiscal year
CREATE INDEX IF NOT EXISTS idx_accounting_budgets_org_status_year 
    ON accounting_budgets(organization_id, status, fiscal_year) 
    WHERE deleted_at IS NULL;

-- Budget Lines: budget + account (for lookups)
CREATE INDEX IF NOT EXISTS idx_accounting_budget_lines_budget_account 
    ON accounting_budget_lines(budget_id, account_id);

-- Cost Centers: organization + active + type
CREATE INDEX IF NOT EXISTS idx_accounting_cost_centers_org_active_type 
    ON accounting_cost_centers(organization_id, is_active, center_type) 
    WHERE deleted_at IS NULL;

-- Bank Reconciliations: organization + status + date
CREATE INDEX IF NOT EXISTS idx_accounting_bank_recon_org_status_date 
    ON accounting_bank_reconciliations(organization_id, status, reconciliation_date DESC) 
    WHERE deleted_at IS NULL;

-- Bank Reconciliation Items: reconciliation + matched status
CREATE INDEX IF NOT EXISTS idx_accounting_bank_recon_items_recon_matched 
    ON accounting_bank_reconciliation_items(reconciliation_id, is_matched) 
    WHERE is_matched = FALSE;

COMMENT ON INDEX idx_accounting_accounts_org_active IS 'Optimize list active accounts by organization';
COMMENT ON INDEX idx_accounting_entries_org_status_date IS 'Optimize journal entry queries by org + status + date';
COMMENT ON INDEX idx_accounting_ledger_account_period IS 'Optimize balance queries for specific account in period';

-- ============================================================================
-- Materialized Views for Reporting
-- ============================================================================

-- Account Balances Summary (for fast balance sheet generation)
CREATE MATERIALIZED VIEW IF NOT EXISTS accounting_mv_account_balances AS
SELECT 
    l.organization_id,
    l.account_id,
    a.code AS account_code,
    a.name AS account_name,
    a.type AS account_type,
    l.fiscal_period_id,
    fp.period_code,
    fp.period_name,
    l.opening_balance_cents,
    l.debit_cents,
    l.credit_cents,
    l.closing_balance_cents,
    l.currency_code,
    l.updated_at AS last_calculated_at
FROM accounting_ledger l
JOIN accounting_chart_of_accounts a ON l.account_id = a.id AND a.deleted_at IS NULL
JOIN accounting_fiscal_periods fp ON l.fiscal_period_id = fp.id AND fp.deleted_at IS NULL;

CREATE UNIQUE INDEX idx_accounting_mv_balances_unique 
    ON accounting_mv_account_balances(account_id, fiscal_period_id);
CREATE INDEX idx_accounting_mv_balances_org 
    ON accounting_mv_account_balances(organization_id);
CREATE INDEX idx_accounting_mv_balances_period 
    ON accounting_mv_account_balances(fiscal_period_id);
CREATE INDEX idx_accounting_mv_balances_type 
    ON accounting_mv_account_balances(account_type);

COMMENT ON MATERIALIZED VIEW accounting_mv_account_balances IS 
'Denormalized account balances for fast financial statement generation. Refresh after period close.';

-- Trial Balance View (for quick validation)
CREATE MATERIALIZED VIEW IF NOT EXISTS accounting_mv_trial_balance AS
SELECT 
    organization_id,
    fiscal_period_id,
    period_code,
    SUM(CASE WHEN account_type IN ('asset', 'expense') THEN closing_balance_cents ELSE 0 END) AS total_debit_cents,
    SUM(CASE WHEN account_type IN ('liability', 'equity', 'income') THEN closing_balance_cents ELSE 0 END) AS total_credit_cents,
    SUM(CASE WHEN account_type IN ('asset', 'expense') THEN closing_balance_cents ELSE 0 END) -
    SUM(CASE WHEN account_type IN ('liability', 'equity', 'income') THEN closing_balance_cents ELSE 0 END) AS balance_difference_cents,
    COUNT(*) AS account_count,
    MAX(last_calculated_at) AS last_calculated_at
FROM accounting_mv_account_balances
GROUP BY organization_id, fiscal_period_id, period_code;

CREATE UNIQUE INDEX idx_accounting_mv_trial_org_period 
    ON accounting_mv_trial_balance(organization_id, fiscal_period_id);

COMMENT ON MATERIALIZED VIEW accounting_mv_trial_balance IS 
'Trial balance summary by period. Refresh after posting entries or period close.';

-- Budget vs Actual View (for management reports)
CREATE MATERIALIZED VIEW IF NOT EXISTS accounting_mv_budget_vs_actual AS
SELECT 
    b.organization_id,
    b.id AS budget_id,
    b.name AS budget_name,
    b.fiscal_year,
    b.status AS budget_status,
    bl.account_id,
    a.code AS account_code,
    a.name AS account_name,
    bl.budget_amount_cents,
    bl.actual_amount_cents,
    bl.variance_percent,
    b.updated_at AS last_updated_at
FROM accounting_budgets b
JOIN accounting_budget_lines bl ON b.id = bl.budget_id
JOIN accounting_chart_of_accounts a ON bl.account_id = a.id AND a.deleted_at IS NULL
WHERE b.deleted_at IS NULL;

CREATE INDEX idx_accounting_mv_budget_actual_org 
    ON accounting_mv_budget_vs_actual(organization_id);
CREATE INDEX idx_accounting_mv_budget_actual_budget 
    ON accounting_mv_budget_vs_actual(budget_id);
CREATE INDEX idx_accounting_mv_budget_actual_year 
    ON accounting_mv_budget_vs_actual(fiscal_year);

COMMENT ON MATERIALIZED VIEW accounting_mv_budget_vs_actual IS 
'Budget vs actual analysis. Refresh after budget updates or actual posting.';

-- ============================================================================
-- Functions for Materialized View Refresh
-- ============================================================================

-- Refresh all accounting materialized views
CREATE OR REPLACE FUNCTION accounting_refresh_materialized_views()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY accounting_mv_account_balances;
    REFRESH MATERIALIZED VIEW CONCURRENTLY accounting_mv_trial_balance;
    REFRESH MATERIALIZED VIEW CONCURRENTLY accounting_mv_budget_vs_actual;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION accounting_refresh_materialized_views() IS 
'Refresh all accounting materialized views. Call after period close or significant data changes.';

-- Refresh specific period views (more efficient)
CREATE OR REPLACE FUNCTION accounting_refresh_period_views(p_organization_id TEXT, p_fiscal_period_id TEXT)
RETURNS void AS $$
BEGIN
    -- For now, refresh all views (could be optimized to update specific rows)
    REFRESH MATERIALIZED VIEW CONCURRENTLY accounting_mv_account_balances;
    REFRESH MATERIALIZED VIEW CONCURRENTLY accounting_mv_trial_balance;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION accounting_refresh_period_views(TEXT, TEXT) IS 
'Refresh materialized views for specific organization and period.';
