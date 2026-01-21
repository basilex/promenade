-- ============================================================================
-- Accounting Context: Audit Log and Event Sourcing - Rollback
-- ============================================================================

-- Drop functions
DROP FUNCTION IF EXISTS accounting_get_entity_audit_trail(VARCHAR, TEXT, INTEGER);
DROP FUNCTION IF EXISTS accounting_reconstruct_journal_entry_history(TEXT);

-- Drop tables
DROP TABLE IF EXISTS accounting_account_balance_snapshots;
DROP TABLE IF EXISTS accounting_fiscal_period_state_changes;
DROP TABLE IF EXISTS accounting_journal_entry_events;

-- Revert audit log changes (only if columns were added by this migration)
-- Note: We don't drop existing audit_log columns to preserve backward compatibility
