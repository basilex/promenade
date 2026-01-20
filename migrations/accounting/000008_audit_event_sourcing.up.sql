-- ============================================================================
-- Accounting Context: Audit Log and Event Sourcing
-- ============================================================================
-- Comprehensive audit trail for compliance and debugging
-- Event sourcing for journal entries
-- ============================================================================

-- ============================================================================
-- Audit Log (Enhanced version of existing table)
-- ============================================================================

-- Add additional columns to existing audit log if needed
-- This migration enhances the existing accounting_audit_log table

-- Ensure audit log has all necessary fields
DO $$ BEGIN
    -- Add change_type if not exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounting_audit_log' 
        AND column_name = 'change_type'
    ) THEN
        ALTER TABLE accounting_audit_log 
        ADD COLUMN change_type VARCHAR(20) CHECK (change_type IN ('create', 'update', 'delete', 'post', 'reverse', 'close', 'lock'));
    END IF;

    -- Add before_state if not exists  
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounting_audit_log' 
        AND column_name = 'before_state'
    ) THEN
        ALTER TABLE accounting_audit_log 
        ADD COLUMN before_state TEXT;
    END IF;

    -- Add after_state if not exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounting_audit_log' 
        AND column_name = 'after_state'
    ) THEN
        ALTER TABLE accounting_audit_log 
        ADD COLUMN after_state TEXT;
    END IF;

    -- Add ip_address for security audit
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounting_audit_log' 
        AND column_name = 'ip_address'
    ) THEN
        ALTER TABLE accounting_audit_log 
        ADD COLUMN ip_address VARCHAR(45);
    END IF;

    -- Add user_agent for security audit
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounting_audit_log' 
        AND column_name = 'user_agent'
    ) THEN
        ALTER TABLE accounting_audit_log 
        ADD COLUMN user_agent TEXT;
    END IF;
END $$;

-- Indices for audit log queries
CREATE INDEX IF NOT EXISTS idx_accounting_audit_entity_type 
    ON accounting_audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_accounting_audit_user 
    ON accounting_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_accounting_audit_created_at 
    ON accounting_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_accounting_audit_change_type 
    ON accounting_audit_log(change_type);

COMMENT ON COLUMN accounting_audit_log.change_type IS 'Type of change for filtering and reporting';
COMMENT ON COLUMN accounting_audit_log.before_state IS 'JSON snapshot before change';
COMMENT ON COLUMN accounting_audit_log.after_state IS 'JSON snapshot after change';

-- ============================================================================
-- Journal Entry Event Store (Event Sourcing)
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_journal_entry_events (
    id TEXT PRIMARY KEY,
    
    -- Event metadata
    journal_entry_id TEXT NOT NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'entry_created', 'line_added', 'line_removed', 'line_updated',
        'entry_posted', 'entry_reversed', 'description_updated'
    )),
    event_version INTEGER NOT NULL DEFAULT 1,
    
    -- Event payload
    event_data TEXT NOT NULL, -- JSON
    
    -- Causation and correlation
    causation_id TEXT, -- ID of command that caused this event
    correlation_id TEXT, -- ID for tracing related events
    
    -- Metadata
    organization_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    event_timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Optional context
    metadata TEXT -- JSON for additional context
);

CREATE INDEX idx_accounting_je_events_entry 
    ON accounting_journal_entry_events(journal_entry_id, event_version);
CREATE INDEX idx_accounting_je_events_type 
    ON accounting_journal_entry_events(event_type);
CREATE INDEX idx_accounting_je_events_timestamp 
    ON accounting_journal_entry_events(event_timestamp DESC);
CREATE INDEX idx_accounting_je_events_correlation 
    ON accounting_journal_entry_events(correlation_id) 
    WHERE correlation_id IS NOT NULL;

COMMENT ON TABLE accounting_journal_entry_events IS 
'Event store for journal entries - enables full history reconstruction and compliance';
COMMENT ON COLUMN accounting_journal_entry_events.event_version IS 
'Sequential version number for ordering events';
COMMENT ON COLUMN accounting_journal_entry_events.causation_id IS 
'ID of command that caused this event (for debugging)';
COMMENT ON COLUMN accounting_journal_entry_events.correlation_id IS 
'ID for tracing related events across aggregates';

-- ============================================================================
-- Fiscal Period State Changes (Compliance tracking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_fiscal_period_state_changes (
    id TEXT PRIMARY KEY,
    
    fiscal_period_id TEXT NOT NULL REFERENCES accounting_fiscal_periods(id),
    organization_id TEXT NOT NULL,
    
    -- State transition
    from_status VARCHAR(20) NOT NULL,
    to_status VARCHAR(20) NOT NULL,
    
    -- Who and when
    changed_by TEXT NOT NULL,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Reason and validation
    reason TEXT,
    validation_result TEXT, -- JSON with validation details
    
    -- Financial snapshot at close
    period_summary TEXT -- JSON with revenue, expenses, etc.
);

CREATE INDEX idx_accounting_period_changes_period 
    ON accounting_fiscal_period_state_changes(fiscal_period_id, changed_at DESC);
CREATE INDEX idx_accounting_period_changes_org 
    ON accounting_fiscal_period_state_changes(organization_id);

COMMENT ON TABLE accounting_fiscal_period_state_changes IS 
'Tracks all fiscal period state transitions for compliance and auditing';
COMMENT ON COLUMN accounting_fiscal_period_state_changes.period_summary IS 
'Financial summary at period close for regulatory compliance';

-- ============================================================================
-- Account Balance History (for reconciliation)
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounting_account_balance_snapshots (
    id TEXT PRIMARY KEY,
    
    account_id TEXT NOT NULL REFERENCES accounting_chart_of_accounts(id),
    organization_id TEXT NOT NULL,
    fiscal_period_id TEXT REFERENCES accounting_fiscal_periods(id),
    
    -- Balance at snapshot time
    snapshot_date DATE NOT NULL,
    balance_cents INTEGER NOT NULL,
    currency_code CHAR(3) NOT NULL,
    
    -- Aggregated activity
    period_debit_cents INTEGER NOT NULL DEFAULT 0,
    period_credit_cents INTEGER NOT NULL DEFAULT 0,
    
    -- Metadata
    snapshot_type VARCHAR(20) NOT NULL CHECK (snapshot_type IN ('daily', 'period_end', 'fiscal_year_end')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_accounting_balance_snapshots 
    ON accounting_account_balance_snapshots(account_id, snapshot_date, snapshot_type);
CREATE INDEX idx_accounting_balance_snapshots_account 
    ON accounting_account_balance_snapshots(account_id, snapshot_date DESC);
CREATE INDEX idx_accounting_balance_snapshots_org_date 
    ON accounting_account_balance_snapshots(organization_id, snapshot_date DESC);

COMMENT ON TABLE accounting_account_balance_snapshots IS 
'Historical balance snapshots for trend analysis and reconciliation';

-- ============================================================================
-- Functions for Event Sourcing
-- ============================================================================

-- Reconstruct journal entry from events (for debugging/compliance)
CREATE OR REPLACE FUNCTION accounting_reconstruct_journal_entry_history(p_journal_entry_id TEXT)
RETURNS TABLE (
    event_version INTEGER,
    event_type VARCHAR,
    event_data TEXT,
    user_id TEXT,
    event_timestamp TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        e.event_version,
        e.event_type,
        e.event_data,
        e.user_id,
        e.event_timestamp
    FROM accounting_journal_entry_events e
    WHERE e.journal_entry_id = p_journal_entry_id
    ORDER BY e.event_version ASC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION accounting_reconstruct_journal_entry_history(TEXT) IS 
'Reconstructs complete event history for a journal entry';

-- Get audit trail for entity
CREATE OR REPLACE FUNCTION accounting_get_entity_audit_trail(
    p_entity_type VARCHAR,
    p_entity_id TEXT,
    p_limit INTEGER DEFAULT 100
)
RETURNS TABLE (
    change_type VARCHAR,
    changed_at TIMESTAMP,
    user_id TEXT,
    before_state TEXT,
    after_state TEXT,
    description TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        al.change_type,
        al.created_at AS changed_at,
        al.user_id,
        al.before_state,
        al.after_state,
        al.reason AS description
    FROM accounting_audit_log al
    WHERE al.entity_type = p_entity_type 
      AND al.entity_id = p_entity_id
    ORDER BY al.created_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION accounting_get_entity_audit_trail(VARCHAR, TEXT, INTEGER) IS 
'Retrieves complete audit trail for any entity';
