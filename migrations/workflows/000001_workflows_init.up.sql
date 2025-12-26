-- ============================================================================
-- Workflows Module - Initial Schema
-- ============================================================================
-- Creates core tables for workflow management system:
-- - workflows_definitions: Workflow templates
-- - workflows_instances: Running workflows
-- - workflows_steps: Execution audit trail
-- - workflows_variables: Context storage
-- - workflows_events: Event triggers
-- ============================================================================

-- ============================================================================
-- 1. Workflow Definitions (Templates/Blueprints)
-- ============================================================================
CREATE TABLE workflows_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(255) NOT NULL,  -- Unique identifier (e.g., "order_fulfillment")
    display_name VARCHAR(255) NOT NULL,  -- Human-readable name
    description TEXT,
    version INTEGER NOT NULL DEFAULT 1,  -- Version number for schema evolution
    status VARCHAR(50) NOT NULL DEFAULT 'draft',  -- draft, active, deprecated, archived
    category VARCHAR(100),  -- Category for organization (sales, hr, etc.)
    tags TEXT[] DEFAULT '{}',  -- Tags for filtering
    definition JSONB NOT NULL,  -- Workflow schema (states, transitions, activities)
    input_schema JSONB,  -- JSON Schema for input validation
    created_by UUID NOT NULL REFERENCES core_users(id),  -- Creator
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,  -- Soft delete
    
    -- Constraints
    CONSTRAINT workflows_definitions_version_check CHECK (version > 0),
    CONSTRAINT workflows_definitions_status_check CHECK (status IN ('draft', 'active', 'deprecated', 'archived')),
    UNIQUE (name, version)  -- Unique name per version
);

-- Indexes for performance
CREATE INDEX idx_workflows_definitions_name ON workflows_definitions(name);
CREATE INDEX idx_workflows_definitions_status ON workflows_definitions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflows_definitions_category ON workflows_definitions(category) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflows_definitions_created_by ON workflows_definitions(created_by);
CREATE INDEX idx_workflows_definitions_deleted_at ON workflows_definitions(deleted_at);

-- GIN index for tags array search
CREATE INDEX idx_workflows_definitions_tags ON workflows_definitions USING GIN(tags);

-- GIN index for JSON search in definition
CREATE INDEX idx_workflows_definitions_definition ON workflows_definitions USING GIN(definition);

-- Comment
COMMENT ON TABLE workflows_definitions IS 'Workflow templates/blueprints that define the structure and flow';
COMMENT ON COLUMN workflows_definitions.definition IS 'JSONB schema: {states: [], transitions: [], initial_state: "", final_states: []}';
COMMENT ON COLUMN workflows_definitions.input_schema IS 'JSON Schema for validating workflow input data';


-- ============================================================================
-- 2. Workflow Instances (Running Workflows)
-- ============================================================================
CREATE TABLE workflows_instances (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    definition_id UUID NOT NULL REFERENCES workflows_definitions(id),  -- Template
    definition_version INTEGER NOT NULL,  -- Version used
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, running, waiting, paused, completed, failed, cancelled, timed_out
    current_state VARCHAR(255) NOT NULL,  -- Current state name
    previous_state VARCHAR(255),  -- Previous state name
    context JSONB NOT NULL DEFAULT '{}',  -- Workflow variables (JSON)
    input JSONB,  -- Initial input data
    output JSONB,  -- Final output data
    error_message TEXT,  -- Error if failed
    error_details JSONB,  -- Detailed error info
    started_by UUID NOT NULL REFERENCES core_users(id),  -- User who started
    assigned_to UUID REFERENCES core_users(id),  -- Current assignee (for manual tasks)
    priority INTEGER NOT NULL DEFAULT 5,  -- Priority (1-10, higher = more urgent)
    due_date TIMESTAMP,  -- Expected completion date
    parent_instance_id UUID REFERENCES workflows_instances(id),  -- Parent workflow (sub-workflows)
    external_reference VARCHAR(255),  -- External ID (order_id, ticket_id, etc.)
    tags TEXT[] DEFAULT '{}',  -- Tags for filtering
    retry_count INTEGER NOT NULL DEFAULT 0,  -- Number of retries
    started_at TIMESTAMP,  -- When execution started
    completed_at TIMESTAMP,  -- When execution completed
    state_entered_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- When entered current state
    state_timeout_at TIMESTAMP,  -- When current state times out
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,  -- Soft delete
    
    -- Constraints
    CONSTRAINT workflows_instances_status_check CHECK (status IN ('pending', 'running', 'waiting', 'paused', 'completed', 'failed', 'cancelled', 'timed_out')),
    CONSTRAINT workflows_instances_priority_check CHECK (priority >= 1 AND priority <= 10),
    CONSTRAINT workflows_instances_retry_count_check CHECK (retry_count >= 0)
);

-- Indexes for performance
CREATE INDEX idx_workflows_instances_definition_id ON workflows_instances(definition_id);
CREATE INDEX idx_workflows_instances_status ON workflows_instances(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflows_instances_current_state ON workflows_instances(current_state) WHERE deleted_at IS NULL;
CREATE INDEX idx_workflows_instances_started_by ON workflows_instances(started_by);
CREATE INDEX idx_workflows_instances_assigned_to ON workflows_instances(assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX idx_workflows_instances_priority ON workflows_instances(priority) WHERE status IN ('pending', 'running', 'waiting');
CREATE INDEX idx_workflows_instances_due_date ON workflows_instances(due_date) WHERE due_date IS NOT NULL AND status NOT IN ('completed', 'failed', 'cancelled', 'timed_out');
CREATE INDEX idx_workflows_instances_parent_instance_id ON workflows_instances(parent_instance_id) WHERE parent_instance_id IS NOT NULL;
CREATE INDEX idx_workflows_instances_external_reference ON workflows_instances(external_reference) WHERE external_reference IS NOT NULL;
CREATE INDEX idx_workflows_instances_state_timeout_at ON workflows_instances(state_timeout_at) WHERE state_timeout_at IS NOT NULL;
CREATE INDEX idx_workflows_instances_created_at ON workflows_instances(created_at);
CREATE INDEX idx_workflows_instances_deleted_at ON workflows_instances(deleted_at);

-- GIN index for tags
CREATE INDEX idx_workflows_instances_tags ON workflows_instances USING GIN(tags);

-- GIN index for context search
CREATE INDEX idx_workflows_instances_context ON workflows_instances USING GIN(context);

-- Composite indexes for common queries
CREATE INDEX idx_workflows_instances_active ON workflows_instances(definition_id, status) WHERE status IN ('pending', 'running', 'waiting', 'paused');
-- Note: Removed idx_workflows_instances_overdue - CURRENT_TIMESTAMP not allowed in index predicate
-- Use query-time filtering for overdue instances instead

-- Comment
COMMENT ON TABLE workflows_instances IS 'Running instances of workflows (executions)';
COMMENT ON COLUMN workflows_instances.context IS 'Workflow variables stored as JSON (global, state, local scopes)';
COMMENT ON COLUMN workflows_instances.priority IS 'Priority 1-10: 1=lowest, 10=highest, 5=default';
COMMENT ON COLUMN workflows_instances.external_reference IS 'External system reference (order ID, ticket ID, etc.)';


-- ============================================================================
-- 3. Workflow Steps (Execution Audit Trail)
-- ============================================================================
CREATE TABLE workflows_steps (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    instance_id UUID NOT NULL REFERENCES workflows_instances(id) ON DELETE CASCADE,  -- Parent instance
    definition_id UUID NOT NULL REFERENCES workflows_definitions(id),  -- Definition reference
    step_number INTEGER NOT NULL,  -- Sequential number (1, 2, 3...)
    type VARCHAR(50) NOT NULL,  -- activity, transition, gateway, event, timer
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, running, completed, failed, skipped, retrying
    state_name VARCHAR(255) NOT NULL,  -- State being executed
    activity_name VARCHAR(255),  -- Activity name if type=activity
    event VARCHAR(255),  -- Event that triggered transition
    input JSONB,  -- Step input
    output JSONB,  -- Step output
    error_message TEXT,  -- Error if failed
    error_details JSONB,  -- Detailed error
    retry_count INTEGER NOT NULL DEFAULT 0,  -- Retry attempts
    duration BIGINT,  -- Execution time in milliseconds
    executed_by UUID REFERENCES core_users(id),  -- User if manual step
    started_at TIMESTAMP,  -- Execution start
    completed_at TIMESTAMP,  -- Execution end
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT workflows_steps_type_check CHECK (type IN ('activity', 'transition', 'gateway', 'event', 'timer')),
    CONSTRAINT workflows_steps_status_check CHECK (status IN ('pending', 'running', 'completed', 'failed', 'skipped', 'retrying')),
    CONSTRAINT workflows_steps_step_number_check CHECK (step_number > 0),
    CONSTRAINT workflows_steps_duration_check CHECK (duration >= 0)
);

-- Indexes for performance
CREATE INDEX idx_workflows_steps_instance_id ON workflows_steps(instance_id);
CREATE INDEX idx_workflows_steps_definition_id ON workflows_steps(definition_id);
CREATE INDEX idx_workflows_steps_step_number ON workflows_steps(instance_id, step_number);
CREATE INDEX idx_workflows_steps_status ON workflows_steps(status);
CREATE INDEX idx_workflows_steps_state_name ON workflows_steps(state_name);
CREATE INDEX idx_workflows_steps_type ON workflows_steps(type);
CREATE INDEX idx_workflows_steps_executed_by ON workflows_steps(executed_by) WHERE executed_by IS NOT NULL;
CREATE INDEX idx_workflows_steps_created_at ON workflows_steps(created_at);

-- Comment
COMMENT ON TABLE workflows_steps IS 'Audit trail of workflow execution steps';
COMMENT ON COLUMN workflows_steps.duration IS 'Execution time in milliseconds';


-- ============================================================================
-- 4. Workflow Variables (Context Storage)
-- ============================================================================
CREATE TABLE workflows_variables (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    instance_id UUID NOT NULL REFERENCES workflows_instances(id) ON DELETE CASCADE,  -- Parent instance
    name VARCHAR(255) NOT NULL,  -- Variable name
    value JSONB NOT NULL,  -- Variable value (JSON)
    type VARCHAR(50) NOT NULL,  -- Type hint: string, number, boolean, object, array
    scope VARCHAR(50) NOT NULL DEFAULT 'global',  -- global, state, local
    state_name VARCHAR(255),  -- State if scope=state
    set_by UUID REFERENCES core_users(id),  -- User who set the value
    set_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- When value was set
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT workflows_variables_type_check CHECK (type IN ('string', 'number', 'boolean', 'object', 'array')),
    CONSTRAINT workflows_variables_scope_check CHECK (scope IN ('global', 'state', 'local')),
    CONSTRAINT workflows_variables_state_scope CHECK (scope != 'state' OR state_name IS NOT NULL),
    UNIQUE (instance_id, name, scope, state_name)  -- Unique variable per instance/scope/state
);

-- Indexes for performance
CREATE INDEX idx_workflows_variables_instance_id ON workflows_variables(instance_id);
CREATE INDEX idx_workflows_variables_name ON workflows_variables(name);
CREATE INDEX idx_workflows_variables_scope ON workflows_variables(scope);
CREATE INDEX idx_workflows_variables_state_name ON workflows_variables(state_name) WHERE state_name IS NOT NULL;

-- GIN index for value search
CREATE INDEX idx_workflows_variables_value ON workflows_variables USING GIN(value);

-- Comment
COMMENT ON TABLE workflows_variables IS 'Variables stored during workflow execution';
COMMENT ON COLUMN workflows_variables.scope IS 'Variable scope: global (everywhere), state (specific state), local (current activity)';


-- ============================================================================
-- 5. Workflow Events (External Triggers)
-- ============================================================================
CREATE TABLE workflows_events (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    instance_id UUID NOT NULL REFERENCES workflows_instances(id) ON DELETE CASCADE,  -- Parent instance
    type VARCHAR(50) NOT NULL,  -- signal, message, timer, error, manual, webhook, internal
    name VARCHAR(255) NOT NULL,  -- Event name (e.g., "order_paid", "approval_received")
    payload JSONB,  -- Event data
    triggered_by UUID REFERENCES core_users(id),  -- User who triggered
    source VARCHAR(500),  -- Event source (webhook URL, user action, etc.)
    processed BOOLEAN NOT NULL DEFAULT FALSE,  -- Whether event was processed
    processed_at TIMESTAMP,  -- When event was processed
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT workflows_events_type_check CHECK (type IN ('signal', 'message', 'timer', 'error', 'manual', 'webhook', 'internal'))
);

-- Indexes for performance
CREATE INDEX idx_workflows_events_instance_id ON workflows_events(instance_id);
CREATE INDEX idx_workflows_events_type ON workflows_events(type);
CREATE INDEX idx_workflows_events_name ON workflows_events(name);
CREATE INDEX idx_workflows_events_processed ON workflows_events(processed) WHERE processed = FALSE;
CREATE INDEX idx_workflows_events_triggered_by ON workflows_events(triggered_by) WHERE triggered_by IS NOT NULL;
CREATE INDEX idx_workflows_events_created_at ON workflows_events(created_at);

-- Comment
COMMENT ON TABLE workflows_events IS 'Events that can trigger workflow state transitions';
COMMENT ON COLUMN workflows_events.type IS 'Event type: signal (external), message (from system), timer (scheduled), etc.';


-- ============================================================================
-- Triggers for Updated_at
-- ============================================================================

-- Workflow Definitions
CREATE TRIGGER workflows_definitions_updated_at
    BEFORE UPDATE ON workflows_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Workflow Instances
CREATE TRIGGER workflows_instances_updated_at
    BEFORE UPDATE ON workflows_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Workflow Steps
CREATE TRIGGER workflows_steps_updated_at
    BEFORE UPDATE ON workflows_steps
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Workflow Variables
CREATE TRIGGER workflows_variables_updated_at
    BEFORE UPDATE ON workflows_variables
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- Statistics Views (Optional - for monitoring)
-- ============================================================================

-- Active workflows summary
CREATE OR REPLACE VIEW workflows_active_summary AS
SELECT 
    wd.name AS workflow_name,
    wd.version,
    wi.status,
    COUNT(*) AS count,
    AVG(EXTRACT(EPOCH FROM (COALESCE(wi.completed_at, CURRENT_TIMESTAMP) - wi.started_at))) AS avg_duration_seconds
FROM workflows_instances wi
JOIN workflows_definitions wd ON wi.definition_id = wd.id
WHERE wi.deleted_at IS NULL
GROUP BY wd.name, wd.version, wi.status;

COMMENT ON VIEW workflows_active_summary IS 'Summary of active workflows by status with average duration';

-- ============================================================================
-- Success!
-- ============================================================================
