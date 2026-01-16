-- ============================================================================
-- Fiscal Context: Receipts
-- ============================================================================
-- Fiscal receipts for Ukrainian compliance
-- ============================================================================

CREATE TABLE IF NOT EXISTS fiscal_receipts (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,

    cash_register_id TEXT NOT NULL,
    order_id TEXT NOT NULL,

    payment_type VARCHAR(20) NOT NULL CHECK (payment_type IN ('cash', 'card', 'cashless')),
    receipt_type VARCHAR(20) NOT NULL CHECK (receipt_type IN ('sale', 'return', 'service_in', 'service_out')),
    currency VARCHAR(3) NOT NULL,

    total_amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,

    fiscal_number VARCHAR(50),
    fiscal_url TEXT,
    qr_code TEXT,

    printed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,

    lines TEXT NOT NULL DEFAULT '[]',

    created_by TEXT NOT NULL,
    last_updated_by TEXT NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'printed', 'cancelled')),

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_fiscal_receipts_cash_register ON fiscal_receipts(cash_register_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_fiscal_receipts_order ON fiscal_receipts(order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_fiscal_receipts_status ON fiscal_receipts(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_fiscal_receipts_created_at ON fiscal_receipts(created_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_fiscal_receipts_order_unique ON fiscal_receipts(order_id) WHERE deleted_at IS NULL;

COMMENT ON TABLE fiscal_receipts IS 'Fiscal receipts for Ukrainian compliance';
COMMENT ON COLUMN fiscal_receipts.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN fiscal_receipts.lines IS 'Receipt lines stored as JSON text (database-agnostic)';