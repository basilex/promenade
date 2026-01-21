-- Migration: create_scripts_table
-- Context: scripting
-- Created: 2026-01-08 08:01:36
-- Updated: 2026-01-13 (Database-agnostic refactor)

-- Scripts table: stores LUA scripts with versioning and classification
-- Note: ID, timestamps managed by BaseAggregate in Go code
CREATE TABLE scripting_scripts (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    code TEXT NOT NULL,
    version INTEGER NOT NULL,
    status TEXT NOT NULL,
    
    -- Script classification
    script_type TEXT,  -- 'validation', 'workflow', 'report', 'pricing', 'notification', 'automation', 'custom'
    entity_type TEXT,  -- 'customer', 'order', 'deal', 'invoice', 'contract', 'payment', 'product'
    
    -- Metadata (stored as TEXT JSON for database-agnostic support)
    metadata TEXT DEFAULT '{}',
    
    created_by TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_status CHECK (status IN ('draft', 'active', 'inactive', 'archived')),
    CONSTRAINT chk_script_type CHECK (
        script_type IS NULL OR
        script_type IN ('validation', 'workflow', 'report', 'pricing', 'notification', 'automation', 'custom')
    ),
    CONSTRAINT chk_entity_type CHECK (
        entity_type IS NULL OR
        entity_type IN ('customer', 'order', 'deal', 'invoice', 'contract', 'payment', 'product', 'inventory', 'user')
    )
);

-- Script executions table: audit trail for script runs
-- Note: ID, executed_at managed by Go code
CREATE TABLE scripting_script_executions (
    id TEXT PRIMARY KEY,
    script_id TEXT NOT NULL REFERENCES scripting_scripts(id) ON DELETE CASCADE,
    script_name TEXT NOT NULL,
    
    -- JSON stored as TEXT for database-agnostic support
    input_params TEXT,
    output_result TEXT,
    error TEXT,
    duration_ms INTEGER,
    
    executed_by TEXT NOT NULL,
    executed_at TIMESTAMP NOT NULL,
    
    CONSTRAINT chk_duration CHECK (duration_ms >= 0)
);

-- Indexes for performance
CREATE INDEX idx_scripts_name ON scripting_scripts(name);
CREATE INDEX idx_scripts_status ON scripting_scripts(status);
CREATE INDEX idx_scripts_created_at ON scripting_scripts(created_at);
CREATE INDEX idx_scripts_type ON scripting_scripts(script_type) WHERE script_type IS NOT NULL;
CREATE INDEX idx_scripts_entity ON scripting_scripts(entity_type) WHERE entity_type IS NOT NULL;

CREATE INDEX idx_executions_script_id ON scripting_script_executions(script_id);
CREATE INDEX idx_executions_executed_at ON scripting_script_executions(executed_at);
CREATE INDEX idx_executions_error ON scripting_script_executions(error) WHERE error IS NOT NULL;
