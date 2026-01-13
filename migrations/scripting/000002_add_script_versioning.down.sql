-- Rollback: add_script_versioning
-- Context: scripting
-- Created: 2026-01-13

-- Drop indexes first
DROP INDEX IF EXISTS idx_script_versions_created_by;
DROP INDEX IF EXISTS idx_script_versions_created_at;
DROP INDEX IF EXISTS idx_script_versions_script;

-- Drop table
DROP TABLE IF EXISTS scripting_script_versions;
