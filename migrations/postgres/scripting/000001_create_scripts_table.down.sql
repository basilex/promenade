-- Migration: create_scripts_table
-- Context: scripting
-- Created: 2026-01-08 08:01:36

-- Drop indexes
DROP INDEX IF EXISTS idx_executions_error;
DROP INDEX IF EXISTS idx_executions_executed_at;
DROP INDEX IF EXISTS idx_executions_script_id;

DROP INDEX IF EXISTS idx_scripts_created_at;
DROP INDEX IF EXISTS idx_scripts_status;
DROP INDEX IF EXISTS idx_scripts_name;

-- Drop tables (order matters: executions references scripts)
DROP TABLE IF EXISTS scripting_script_executions;
DROP TABLE IF EXISTS scripting_scripts;
