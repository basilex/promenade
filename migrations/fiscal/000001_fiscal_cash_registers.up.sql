-- ============================================================================
-- Fiscal Context: Cash Registers (ПРРО - Програмний РРО)
-- ============================================================================
-- Cash register aggregate for Ukrainian fiscal compliance
-- Part of Fiscal Bounded Context
-- ============================================================================

-- Cash Registers table
CREATE TABLE IF NOT EXISTS fiscal_cash_registers (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,
    
    organization_id TEXT NOT NULL,
    fiscal_number VARCHAR(50) UNIQUE NOT NULL,
    model VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive' CHECK (status IN ('inactive', 'active', 'maintenance', 'suspended')),
    
    license_key VARCHAR(255),
    last_sync_at TIMESTAMP,
    provider_cash_register_id TEXT,
    active_shift_id TEXT,
    shift_opened_at TIMESTAMP,
    shift_closed_at TIMESTAMP,
    last_z_report_id TEXT,
    last_z_report_at TIMESTAMP,
    last_updated_by TEXT NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for cash registers
CREATE INDEX idx_fiscal_cash_registers_organization ON fiscal_cash_registers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_fiscal_cash_registers_fiscal_number ON fiscal_cash_registers(fiscal_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_fiscal_cash_registers_status ON fiscal_cash_registers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_fiscal_cash_registers_active ON fiscal_cash_registers(status) WHERE status = 'active' AND deleted_at IS NULL;
CREATE INDEX idx_fiscal_cash_registers_provider_id ON fiscal_cash_registers(provider_cash_register_id) WHERE deleted_at IS NULL;

-- Comments for documentation
COMMENT ON TABLE fiscal_cash_registers IS 'Fiscal cash registers (ПРРО) for Ukrainian compliance';
COMMENT ON COLUMN fiscal_cash_registers.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN fiscal_cash_registers.deleted_at IS 'Soft delete timestamp - NULL means not deleted';
