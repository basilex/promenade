-- ============================================================================
-- Accounting Context: Audit Trail & Financial Reports
-- ============================================================================
-- Comprehensive audit logging and financial report definitions
-- ============================================================================

-- Audit log for all accounting changes
CREATE TABLE IF NOT EXISTS accounting_audit_log (
    id TEXT PRIMARY KEY,
    
    organization_id TEXT NOT NULL,
    
    -- Entity tracking
    entity_type VARCHAR(50) NOT NULL, -- 'journal_entry', 'account', 'fiscal_period', etc.
    entity_id TEXT NOT NULL,
    
    -- Action tracking
    action VARCHAR(30) NOT NULL CHECK (action IN ('create', 'update', 'delete', 'post', 'reverse', 'close', 'lock', 'unlock', 'approve', 'reject')),
    
    -- Changes
    old_values TEXT, -- JSON of previous values
    new_values TEXT, -- JSON of new values
    
    -- Context
    user_id TEXT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Metadata
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounting_audit_org ON accounting_audit_log(organization_id);
CREATE INDEX idx_accounting_audit_entity ON accounting_audit_log(entity_type, entity_id);
CREATE INDEX idx_accounting_audit_action ON accounting_audit_log(action);
CREATE INDEX idx_accounting_audit_user ON accounting_audit_log(user_id);
CREATE INDEX idx_accounting_audit_created ON accounting_audit_log(created_at DESC);

COMMENT ON TABLE accounting_audit_log IS 'Complete audit trail for all accounting operations';
COMMENT ON COLUMN accounting_audit_log.old_values IS 'JSON snapshot of entity before change';
COMMENT ON COLUMN accounting_audit_log.new_values IS 'JSON snapshot of entity after change';

-- ============================================================================
-- Financial Report Templates (Шаблони звітів)
-- ============================================================================
-- Configurable financial report templates (Balance Sheet, P&L, Cash Flow)
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_report_templates (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    report_code VARCHAR(50) NOT NULL,
    report_name VARCHAR(255) NOT NULL,
    
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN ('balance_sheet', 'income_statement', 'cash_flow', 'trial_balance', 'general_ledger', 'custom')),
    
    -- Report structure (JSON)
    structure_json TEXT NOT NULL,
    
    -- Formatting options
    show_zero_balances BOOLEAN NOT NULL DEFAULT FALSE,
    show_inactive_accounts BOOLEAN NOT NULL DEFAULT FALSE,
    group_by_section BOOLEAN NOT NULL DEFAULT TRUE,
    
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    description TEXT,
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_reports_code ON accounting_report_templates(organization_id, report_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_reports_org ON accounting_report_templates(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_reports_type ON accounting_report_templates(report_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_reports_default ON accounting_report_templates(is_default) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_report_templates IS 'Financial report template definitions';
COMMENT ON COLUMN accounting_report_templates.structure_json IS 'JSON defining report sections, groupings, formulas';

-- ============================================================================
-- Saved Report Snapshots (Збережені звіти)
-- ============================================================================
-- Generated financial reports saved for audit and comparison
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_report_snapshots (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    report_template_id TEXT NOT NULL REFERENCES accounting_report_templates(id),
    
    fiscal_period_id TEXT REFERENCES accounting_fiscal_periods(id),
    
    -- Report parameters
    report_date DATE NOT NULL,
    as_of_date DATE NOT NULL, -- Balance sheet: as of date, P&L: end date
    from_date DATE, -- For period reports (P&L, Cash Flow)
    
    -- Generated report data (JSON)
    report_data_json TEXT NOT NULL,
    
    -- Summary metrics
    total_assets_cents INTEGER,
    total_liabilities_cents INTEGER,
    total_equity_cents INTEGER,
    total_revenue_cents INTEGER,
    total_expenses_cents INTEGER,
    net_income_cents INTEGER,
    
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Generation tracking
    generated_by TEXT NOT NULL,
    generated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Approval workflow
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'final', 'approved', 'published')),
    approved_by TEXT,
    approved_at TIMESTAMP,
    
    notes TEXT,
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_accounting_snapshots_org ON accounting_report_snapshots(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_snapshots_template ON accounting_report_snapshots(report_template_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_snapshots_period ON accounting_report_snapshots(fiscal_period_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_snapshots_date ON accounting_report_snapshots(as_of_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_snapshots_status ON accounting_report_snapshots(status) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_report_snapshots IS 'Saved financial report snapshots for audit and comparison';
COMMENT ON COLUMN accounting_report_snapshots.report_data_json IS 'Complete report data with all line items, calculations, and formatting';

-- ============================================================================
-- Account Restrictions (Обмеження рахунків)
-- ============================================================================
-- Define which accounts can be used in which contexts
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_account_restrictions (
    id TEXT PRIMARY KEY,
    
    account_id TEXT NOT NULL REFERENCES accounting_chart_of_accounts(id),
    
    -- Posting restrictions
    allow_manual_entry BOOLEAN NOT NULL DEFAULT TRUE,
    allow_bank_transactions BOOLEAN NOT NULL DEFAULT TRUE,
    allow_invoice_postings BOOLEAN NOT NULL DEFAULT TRUE,
    require_cost_center BOOLEAN NOT NULL DEFAULT FALSE,
    require_project BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Date restrictions
    restrict_future_dates BOOLEAN NOT NULL DEFAULT FALSE,
    max_future_days INTEGER,
    
    -- Amount restrictions
    require_approval_above_cents INTEGER,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_restrictions ON accounting_account_restrictions(account_id);

COMMENT ON TABLE accounting_account_restrictions IS 'Control rules for account usage';

-- ============================================================================
-- Cost Centers & Profit Centers (Центри витрат / прибутку)
-- ============================================================================
-- Optional cost/profit center tracking for management accounting
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_cost_centers (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    
    center_type VARCHAR(30) NOT NULL CHECK (center_type IN ('cost_center', 'profit_center', 'investment_center')),
    
    -- Hierarchy
    parent_id TEXT REFERENCES accounting_cost_centers(id),
    level INTEGER NOT NULL DEFAULT 1,
    
    -- Responsibility tracking
    manager_id TEXT,
    
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT,
    
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_cost_centers ON accounting_cost_centers(organization_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_cost_centers_org ON accounting_cost_centers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_cost_centers_parent ON accounting_cost_centers(parent_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_cost_centers IS 'Cost/profit centers for management accounting';

-- Cost center allocation on journal entry lines
CREATE TABLE IF NOT EXISTS accounting_journal_entry_line_allocations (
    id TEXT PRIMARY KEY,
    
    journal_entry_line_id TEXT NOT NULL REFERENCES accounting_journal_entry_lines(id),
    cost_center_id TEXT NOT NULL REFERENCES accounting_cost_centers(id),
    
    -- Allocation details
    allocation_percent DECIMAL(5, 2) NOT NULL DEFAULT 100.00,
    allocated_amount_cents INTEGER NOT NULL,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounting_line_alloc_line ON accounting_journal_entry_line_allocations(journal_entry_line_id);
CREATE INDEX idx_accounting_line_alloc_center ON accounting_journal_entry_line_allocations(cost_center_id);

COMMENT ON TABLE accounting_journal_entry_line_allocations IS 'Cost center allocation for journal entry lines';
