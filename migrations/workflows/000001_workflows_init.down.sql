-- ============================================================================
-- Workflows Module - Rollback Initial Schema
-- ============================================================================

-- Drop views
DROP VIEW IF EXISTS workflows_active_summary;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS workflows_events CASCADE;
DROP TABLE IF EXISTS workflows_variables CASCADE;
DROP TABLE IF EXISTS workflows_steps CASCADE;
DROP TABLE IF EXISTS workflows_instances CASCADE;
DROP TABLE IF EXISTS workflows_definitions CASCADE;

-- Note: We don't drop the update_updated_at_column function as it's shared
