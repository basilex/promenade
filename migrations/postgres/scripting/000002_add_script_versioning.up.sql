-- Migration: add_script_versioning
-- Context: scripting
-- Created: 2026-01-13
-- Purpose: Add version history tracking for scripts

-- Script versions table: immutable audit trail of all script changes
-- Stores historical snapshots of code and metadata for rollback capability
CREATE TABLE scripting_script_versions (
    id TEXT PRIMARY KEY,
    script_id TEXT NOT NULL REFERENCES scripting_scripts(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    
    -- Historical snapshot (stored as TEXT for database-agnostic support)
    code TEXT NOT NULL,
    metadata TEXT NOT NULL,
    
    -- Change tracking
    change_log TEXT,
    
    -- Audit fields
    created_by TEXT,
    created_at TIMESTAMP NOT NULL,
    
    -- Constraints
    UNIQUE(script_id, version),
    CONSTRAINT chk_version_positive CHECK (version > 0)
);

-- Indexes for version history queries
CREATE INDEX idx_script_versions_script ON scripting_script_versions(script_id);
CREATE INDEX idx_script_versions_created_at ON scripting_script_versions(created_at);
CREATE INDEX idx_script_versions_created_by ON scripting_script_versions(created_by) WHERE created_by IS NOT NULL;

-- Comments
COMMENT ON TABLE scripting_script_versions IS 'Immutable version history for script changes (audit trail + rollback support)';
COMMENT ON COLUMN scripting_script_versions.code IS 'Historical snapshot of LUA code at this version';
COMMENT ON COLUMN scripting_script_versions.metadata IS 'Historical snapshot of metadata at this version (stored as TEXT JSON)';
COMMENT ON COLUMN scripting_script_versions.change_log IS 'Optional description of changes in this version';
