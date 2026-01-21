-- ============================================================================
-- Accounting Context: Fiscal Periods (Облікові періоди)
-- ============================================================================
-- Manages accounting periods for period-end closing and financial reporting
-- Supports monthly/quarterly/yearly periods with open/closed status
-- ============================================================================

-- Fiscal periods table
CREATE TABLE IF NOT EXISTS accounting_fiscal_periods (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    period_type VARCHAR(20) NOT NULL CHECK (period_type IN ('month', 'quarter', 'year')),
    period_code VARCHAR(20) NOT NULL, -- YYYY-MM, YYYY-Q1, YYYY
    period_name VARCHAR(100) NOT NULL, -- "Січень 2026", "Q1 2026", "2026"
    
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'locked')),
    
    -- Period close tracking
    closed_at TIMESTAMP,
    closed_by TEXT,
    lock_date DATE, -- Transactions before this date cannot be modified
    
    -- Financial summary (denormalized for performance)
    total_revenue_cents INTEGER NOT NULL DEFAULT 0,
    total_expense_cents INTEGER NOT NULL DEFAULT 0,
    net_income_cents INTEGER NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Metadata
    notes TEXT,
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_periods_code ON accounting_fiscal_periods(organization_id, period_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_periods_org ON accounting_fiscal_periods(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_periods_status ON accounting_fiscal_periods(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_periods_dates ON accounting_fiscal_periods(start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_periods_type ON accounting_fiscal_periods(period_type) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_fiscal_periods IS 'Fiscal periods for period-end closing and financial reporting';
COMMENT ON COLUMN accounting_fiscal_periods.status IS 'open: can post entries, closed: cannot post but can adjust, locked: completely frozen';
COMMENT ON COLUMN accounting_fiscal_periods.lock_date IS 'Transactions dated before this cannot be modified in closed periods';

-- ============================================================================
-- Account Ledger (Головна книга) - Account balances by period
-- ============================================================================
-- Stores running balances for each account in each period
-- Denormalized for fast balance queries and reports
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_ledger (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    account_id TEXT NOT NULL REFERENCES accounting_chart_of_accounts(id),
    fiscal_period_id TEXT NOT NULL REFERENCES accounting_fiscal_periods(id),
    
    -- Period identification (denormalized)
    period_code VARCHAR(20) NOT NULL,
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,
    
    -- Balances in cents
    opening_balance_cents INTEGER NOT NULL DEFAULT 0,
    debit_cents INTEGER NOT NULL DEFAULT 0,
    credit_cents INTEGER NOT NULL DEFAULT 0,
    closing_balance_cents INTEGER NOT NULL DEFAULT 0,
    
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Audit
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_ledger_account_period ON accounting_ledger(account_id, fiscal_period_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_ledger_org ON accounting_ledger(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_ledger_account ON accounting_ledger(account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_ledger_period ON accounting_ledger(fiscal_period_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_ledger_period_code ON accounting_ledger(period_code) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_ledger IS 'Account balances by fiscal period for fast reporting';
COMMENT ON COLUMN accounting_ledger.opening_balance_cents IS 'Balance at start of period (closing balance from previous period)';
COMMENT ON COLUMN accounting_ledger.closing_balance_cents IS 'Balance at end of period (opening + debits - credits for assets/expenses)';

-- ============================================================================
-- Account Groups (Групування рахунків)
-- ============================================================================
-- Logical grouping of accounts for financial reporting
-- Supports Ukrainian P(S)BO and IFRS reporting structures
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_account_groups (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    
    -- Group hierarchy
    parent_id TEXT REFERENCES accounting_account_groups(id),
    level INTEGER NOT NULL DEFAULT 1,
    
    -- Reporting classification
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN ('balance_sheet', 'income_statement', 'cash_flow')),
    section VARCHAR(100), -- "Current Assets", "Operating Revenue", etc.
    
    -- Display order in reports
    display_order INTEGER NOT NULL DEFAULT 0,
    
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT,
    
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_groups_code ON accounting_account_groups(organization_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_groups_org ON accounting_account_groups(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_groups_parent ON accounting_account_groups(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_groups_report ON accounting_account_groups(report_type) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_account_groups IS 'Logical grouping of accounts for financial statement reporting';

-- ============================================================================
-- Account to Group Mapping
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_account_group_members (
    id TEXT PRIMARY KEY,
    
    account_id TEXT NOT NULL REFERENCES accounting_chart_of_accounts(id),
    group_id TEXT NOT NULL REFERENCES accounting_account_groups(id),
    
    -- Multiple accounts can be in same group
    -- One account can be in multiple groups (e.g., reporting vs tax groups)
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_group_members ON accounting_account_group_members(account_id, group_id);
CREATE INDEX idx_accounting_group_members_account ON accounting_account_group_members(account_id);
CREATE INDEX idx_accounting_group_members_group ON accounting_account_group_members(group_id);

COMMENT ON TABLE accounting_account_group_members IS 'Many-to-many relationship between accounts and groups';
