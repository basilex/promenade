-- Migration: create_scripts_table
-- Context: scripting
-- Created: 2026-01-08 08:01:36

-- Scripts table: stores LUA scripts with versioning
-- Note: ID, timestamps managed by BaseAggregate in Go code
CREATE TABLE scripting_scripts (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    code TEXT NOT NULL,
    version INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    
    created_by UUID,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_status CHECK (status IN ('draft', 'active', 'inactive', 'archived'))
);

-- Script executions table: audit trail for script runs
-- Note: ID, executed_at managed by Go code
CREATE TABLE scripting_script_executions (
    id UUID PRIMARY KEY,
    script_id UUID NOT NULL REFERENCES scripting_scripts(id) ON DELETE CASCADE,
    script_name VARCHAR(255) NOT NULL,
    
    input_params JSONB,
    output_result JSONB,
    error TEXT,
    duration_ms INTEGER,
    
    executed_by UUID NOT NULL,
    executed_at TIMESTAMP NOT NULL,
    
    CONSTRAINT chk_duration CHECK (duration_ms >= 0)
);

-- Indexes for performance
CREATE INDEX idx_scripts_name ON scripting_scripts(name);
CREATE INDEX idx_scripts_status ON scripting_scripts(status);
CREATE INDEX idx_scripts_created_at ON scripting_scripts(created_at);

CREATE INDEX idx_executions_script_id ON scripting_script_executions(script_id);
CREATE INDEX idx_executions_executed_at ON scripting_script_executions(executed_at);
CREATE INDEX idx_executions_error ON scripting_script_executions(error) WHERE error IS NOT NULL;
