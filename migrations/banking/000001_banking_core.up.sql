-- ============================================================================
-- Banking Context: Core Tables
-- ============================================================================
-- BankAccount, BankStatement, BankTransaction aggregates
-- Part of Banking Bounded Context
-- ============================================================================

-- Bank accounts
CREATE TABLE IF NOT EXISTS banking_bank_accounts (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    organization_id TEXT NOT NULL,
    name VARCHAR(150) NOT NULL,
    bank_name VARCHAR(150) NOT NULL,
    iban VARCHAR(34),
    account_number VARCHAR(50),
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',

    provider VARCHAR(30) NOT NULL DEFAULT 'manual' CHECK (provider IN ('manual', 'monobank', 'privat24', 'pumb')),
    provider_account_id TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'archived')),
    last_sync_at TIMESTAMP,
    balance_cents INTEGER NOT NULL DEFAULT 0,

    metadata_json TEXT,
    last_updated_by TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_banking_accounts_org ON banking_bank_accounts(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_accounts_provider ON banking_bank_accounts(provider, provider_account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_accounts_status ON banking_bank_accounts(status) WHERE deleted_at IS NULL;

COMMENT ON TABLE banking_bank_accounts IS 'Bank accounts connected to external providers or manual accounts';
COMMENT ON COLUMN banking_bank_accounts.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN banking_bank_accounts.deleted_at IS 'Soft delete timestamp - NULL means not deleted';

-- Bank statements
CREATE TABLE IF NOT EXISTS banking_bank_statements (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    bank_account_id TEXT NOT NULL,
    statement_date DATE NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,

    opening_balance_cents INTEGER NOT NULL DEFAULT 0,
    closing_balance_cents INTEGER NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',

    provider VARCHAR(30) NOT NULL DEFAULT 'manual' CHECK (provider IN ('manual', 'monobank', 'privat24', 'pumb')),
    provider_statement_id TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'imported' CHECK (status IN ('imported', 'reconciled', 'failed')),
    imported_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    raw_payload_json TEXT,
    last_updated_by TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_banking_statements_account ON banking_bank_statements(bank_account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_statements_period ON banking_bank_statements(bank_account_id, period_start, period_end) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_statements_provider ON banking_bank_statements(provider, provider_statement_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE banking_bank_statements IS 'Imported bank statements by account and period';
COMMENT ON COLUMN banking_bank_statements.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN banking_bank_statements.deleted_at IS 'Soft delete timestamp - NULL means not deleted';

-- Bank transactions
CREATE TABLE IF NOT EXISTS banking_bank_transactions (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    bank_account_id TEXT NOT NULL,
    statement_id TEXT,
    external_id TEXT,

    direction VARCHAR(10) NOT NULL CHECK (direction IN ('debit', 'credit')),
    amount_cents INTEGER NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'UAH',

    counterparty_name VARCHAR(255),
    counterparty_iban VARCHAR(34),
    description TEXT,

    transaction_at TIMESTAMP NOT NULL,
    booked_at TIMESTAMP,

    status VARCHAR(20) NOT NULL DEFAULT 'booked' CHECK (status IN ('pending', 'booked', 'canceled', 'reconciled')),

    matched_entity_type VARCHAR(50),
    matched_entity_id TEXT,
    matched_at TIMESTAMP,

    raw_payload_json TEXT,
    last_updated_by TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_banking_transactions_account ON banking_bank_transactions(bank_account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_transactions_statement ON banking_bank_transactions(statement_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_transactions_external ON banking_bank_transactions(external_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_transactions_status ON banking_bank_transactions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_transactions_matched ON banking_bank_transactions(matched_entity_type, matched_entity_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_banking_transactions_date ON banking_bank_transactions(transaction_at DESC) WHERE deleted_at IS NULL;

COMMENT ON TABLE banking_bank_transactions IS 'Normalized bank transactions for reconciliation and matching';
COMMENT ON COLUMN banking_bank_transactions.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN banking_bank_transactions.deleted_at IS 'Soft delete timestamp - NULL means not deleted';
