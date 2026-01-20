-- ============================================================================
-- Accounting Context: Journal Entry Adjustments & Reconciliation
-- ============================================================================
-- Supports adjustment entries, reversing entries, and bank reconciliation
-- ============================================================================

-- Journal entry relationships (for reversals, adjustments, corrections)
CREATE TABLE IF NOT EXISTS accounting_journal_entry_links (
    id TEXT PRIMARY KEY,
    
    source_entry_id TEXT NOT NULL REFERENCES accounting_journal_entries(id),
    target_entry_id TEXT NOT NULL REFERENCES accounting_journal_entries(id),
    
    link_type VARCHAR(30) NOT NULL CHECK (link_type IN ('reversal', 'adjustment', 'correction', 'allocation', 'reclassification')),
    
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounting_entry_links_source ON accounting_journal_entry_links(source_entry_id);
CREATE INDEX idx_accounting_entry_links_target ON accounting_journal_entry_links(target_entry_id);
CREATE INDEX idx_accounting_entry_links_type ON accounting_journal_entry_links(link_type);

COMMENT ON TABLE accounting_journal_entry_links IS 'Tracks relationships between journal entries (reversals, adjustments, etc.)';
COMMENT ON COLUMN accounting_journal_entry_links.link_type IS 'reversal: cancels original, adjustment: modifies, correction: fixes error';

-- ============================================================================
-- Bank Reconciliation (Звірка банку)
-- ============================================================================
-- Tracks reconciliation of bank statements with accounting records
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_bank_reconciliations (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    bank_account_id TEXT NOT NULL, -- Reference to banking context
    account_id TEXT NOT NULL REFERENCES accounting_chart_of_accounts(id), -- 311 account
    
    reconciliation_date DATE NOT NULL,
    statement_date DATE NOT NULL,
    
    -- Balances
    bank_statement_balance_cents INTEGER NOT NULL,
    book_balance_cents INTEGER NOT NULL,
    
    -- Reconciliation status
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'completed', 'approved')),
    
    -- Differences
    outstanding_deposits_cents INTEGER NOT NULL DEFAULT 0,
    outstanding_checks_cents INTEGER NOT NULL DEFAULT 0,
    bank_fees_cents INTEGER NOT NULL DEFAULT 0,
    interest_earned_cents INTEGER NOT NULL DEFAULT 0,
    
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Approval tracking
    reconciled_by TEXT,
    reconciled_at TIMESTAMP,
    approved_by TEXT,
    approved_at TIMESTAMP,
    
    notes TEXT,
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_accounting_bank_recon_org ON accounting_bank_reconciliations(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_bank_recon_account ON accounting_bank_reconciliations(account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_bank_recon_date ON accounting_bank_reconciliations(reconciliation_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_bank_recon_status ON accounting_bank_reconciliations(status) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_bank_reconciliations IS 'Bank statement reconciliation records';

-- Reconciliation line items (individual transactions)
CREATE TABLE IF NOT EXISTS accounting_bank_reconciliation_items (
    id TEXT PRIMARY KEY,
    
    reconciliation_id TEXT NOT NULL REFERENCES accounting_bank_reconciliations(id),
    
    -- Transaction reference
    transaction_type VARCHAR(30) NOT NULL CHECK (transaction_type IN ('bank_transaction', 'journal_entry', 'outstanding')),
    transaction_id TEXT, -- ID from banking or journal entry
    
    transaction_date DATE NOT NULL,
    description TEXT NOT NULL,
    
    amount_cents INTEGER NOT NULL,
    
    -- Matching status
    is_matched BOOLEAN NOT NULL DEFAULT FALSE,
    matched_at TIMESTAMP,
    
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_accounting_recon_items_recon ON accounting_bank_reconciliation_items(reconciliation_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_recon_items_matched ON accounting_bank_reconciliation_items(is_matched) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_bank_reconciliation_items IS 'Individual items in bank reconciliation';

-- ============================================================================
-- Tax Tracking (Податковий облік)
-- ============================================================================
-- Tracks tax-related information for accounts and journal entries
-- Supports VAT, income tax, payroll tax
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_tax_codes (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    
    tax_type VARCHAR(30) NOT NULL CHECK (tax_type IN ('vat', 'income_tax', 'payroll_tax', 'other')),
    tax_rate DECIMAL(10, 4), -- 20.0000 for 20% VAT
    
    -- Account associations
    tax_payable_account_id TEXT REFERENCES accounting_chart_of_accounts(id),
    tax_receivable_account_id TEXT REFERENCES accounting_chart_of_accounts(id),
    
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT,
    
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_tax_codes ON accounting_tax_codes(organization_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_tax_codes_org ON accounting_tax_codes(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_tax_codes_type ON accounting_tax_codes(tax_type) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_tax_codes IS 'Tax codes for VAT, income tax, payroll tax tracking';

-- Tax amounts on journal entry lines
CREATE TABLE IF NOT EXISTS accounting_journal_entry_line_taxes (
    id TEXT PRIMARY KEY,
    
    journal_entry_line_id TEXT NOT NULL REFERENCES accounting_journal_entry_lines(id),
    tax_code_id TEXT NOT NULL REFERENCES accounting_tax_codes(id),
    
    taxable_amount_cents INTEGER NOT NULL,
    tax_amount_cents INTEGER NOT NULL,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_accounting_line_taxes_line ON accounting_journal_entry_line_taxes(journal_entry_line_id);
CREATE INDEX idx_accounting_line_taxes_code ON accounting_journal_entry_line_taxes(tax_code_id);

COMMENT ON TABLE accounting_journal_entry_line_taxes IS 'Tax breakdown for journal entry lines';

-- ============================================================================
-- Account Budget (Бюджет)
-- ============================================================================
-- Budget tracking per account and period
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_budgets (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    budget_name VARCHAR(255) NOT NULL,
    fiscal_year INTEGER NOT NULL,
    
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'active', 'closed')),
    
    approved_by TEXT,
    approved_at TIMESTAMP,
    
    description TEXT,
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_accounting_budgets_org ON accounting_budgets(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_budgets_year ON accounting_budgets(fiscal_year) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_budgets_status ON accounting_budgets(status) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_budgets IS 'Budget master records';

-- Budget line items (per account, per period)
CREATE TABLE IF NOT EXISTS accounting_budget_lines (
    id TEXT PRIMARY KEY,
    
    budget_id TEXT NOT NULL REFERENCES accounting_budgets(id),
    account_id TEXT NOT NULL REFERENCES accounting_chart_of_accounts(id),
    fiscal_period_id TEXT REFERENCES accounting_fiscal_periods(id),
    
    -- Period identification
    period_code VARCHAR(20) NOT NULL, -- YYYY-MM
    
    -- Budget amounts
    budgeted_amount_cents INTEGER NOT NULL,
    actual_amount_cents INTEGER NOT NULL DEFAULT 0,
    variance_cents INTEGER NOT NULL DEFAULT 0,
    
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',
    
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_accounting_budget_lines_budget ON accounting_budget_lines(budget_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_budget_lines_account ON accounting_budget_lines(account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_budget_lines_period ON accounting_budget_lines(period_code) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_budget_lines IS 'Budget allocations per account and period';
COMMENT ON COLUMN accounting_budget_lines.variance_cents IS 'Actual - Budgeted (positive = over budget)';
