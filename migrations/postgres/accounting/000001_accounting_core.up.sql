-- ============================================================================
-- Accounting Context: Core Tables
-- ============================================================================
-- Chart of Accounts, Journal Entries, Posting Rules
-- Part of Accounting Bounded Context
-- ============================================================================

-- Chart of accounts
CREATE TABLE IF NOT EXISTS accounting_chart_of_accounts (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    code VARCHAR(30) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('asset', 'liability', 'equity', 'income', 'expense')),
    parent_id TEXT,
    level INTEGER NOT NULL DEFAULT 1,

    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description TEXT,

    created_by TEXT NOT NULL,
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_accounts_code ON accounting_chart_of_accounts(organization_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_accounts_org ON accounting_chart_of_accounts(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_accounts_type ON accounting_chart_of_accounts(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_accounts_level ON accounting_chart_of_accounts(level) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_chart_of_accounts IS 'Chart of accounts for double-entry bookkeeping';
COMMENT ON COLUMN accounting_chart_of_accounts.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN accounting_chart_of_accounts.level IS 'Hierarchy level (1 = top-level, 2 = child, etc.)';
COMMENT ON COLUMN accounting_chart_of_accounts.currency_code IS 'Default currency code (ISO 4217)';
COMMENT ON COLUMN accounting_chart_of_accounts.deleted_at IS 'Soft delete timestamp - NULL means not deleted';

-- Journal entries (header)
CREATE TABLE IF NOT EXISTS accounting_journal_entries (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    entry_number VARCHAR(50),
    description TEXT,
    entry_date DATE NOT NULL,

    source_type VARCHAR(50),
    source_id TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'posted', 'reversed')),

    total_debit_cents INTEGER NOT NULL DEFAULT 0,
    total_credit_cents INTEGER NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',

    posted_at TIMESTAMP,
    posted_by TEXT,
    reversed_at TIMESTAMP,
    reversed_by TEXT,

    created_by TEXT NOT NULL,
    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_entries_number ON accounting_journal_entries(organization_id, entry_number) WHERE deleted_at IS NULL AND entry_number IS NOT NULL;
CREATE INDEX idx_accounting_entries_org ON accounting_journal_entries(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_entries_status ON accounting_journal_entries(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_entries_date ON accounting_journal_entries(entry_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_entries_source ON accounting_journal_entries(source_type, source_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_journal_entries IS 'Journal entry headers (double-entry)';
COMMENT ON COLUMN accounting_journal_entries.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN accounting_journal_entries.entry_number IS 'Entry number (nullable for drafts, required before posting)';
COMMENT ON COLUMN accounting_journal_entries.posted_by IS 'User who posted the entry';
COMMENT ON COLUMN accounting_journal_entries.reversed_by IS 'User who reversed the entry';
COMMENT ON COLUMN accounting_journal_entries.deleted_at IS 'Soft delete timestamp - NULL means not deleted';

-- Journal entry lines
CREATE TABLE IF NOT EXISTS accounting_journal_entry_lines (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    journal_entry_id TEXT NOT NULL,
    account_id TEXT NOT NULL,

    debit_cents BIGINT NOT NULL DEFAULT 0,
    credit_cents BIGINT NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',

    description TEXT,
    line_order INTEGER NOT NULL DEFAULT 0,

    last_updated_by TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_accounting_entry_lines_entry ON accounting_journal_entry_lines(journal_entry_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_entry_lines_account ON accounting_journal_entry_lines(account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_entry_lines_order ON accounting_journal_entry_lines(journal_entry_id, line_order) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_journal_entry_lines IS 'Journal entry lines for debit/credit postings';
COMMENT ON COLUMN accounting_journal_entry_lines.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN accounting_journal_entry_lines.debit_cents IS 'Debit amount in cents (0 if credit line)';
COMMENT ON COLUMN accounting_journal_entry_lines.credit_cents IS 'Credit amount in cents (0 if debit line)';
COMMENT ON COLUMN accounting_journal_entry_lines.line_order IS 'Display order within journal entry';
COMMENT ON COLUMN accounting_journal_entry_lines.deleted_at IS 'Soft delete timestamp - NULL means not deleted';

-- Posting rules
CREATE TABLE IF NOT EXISTS accounting_posting_rules (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,

    debit_account_id TEXT NOT NULL,
    credit_account_id TEXT NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    rule_json TEXT,

    last_updated_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_accounting_rules_org ON accounting_posting_rules(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_rules_event ON accounting_posting_rules(event_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounting_rules_active ON accounting_posting_rules(is_active) WHERE deleted_at IS NULL;

COMMENT ON TABLE accounting_posting_rules IS 'Mapping rules from domain events to accounting postings';
COMMENT ON COLUMN accounting_posting_rules.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN accounting_posting_rules.deleted_at IS 'Soft delete timestamp - NULL means not deleted';
