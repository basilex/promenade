-- Create audit_events table for immutable audit logging
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    changes JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    metadata JSONB,
    signature VARCHAR(128) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common query patterns
CREATE INDEX idx_audit_events_user_id ON audit_events(user_id);
CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_action ON audit_events(action);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at DESC);
CREATE INDEX idx_audit_events_request_id ON audit_events(request_id) WHERE request_id IS NOT NULL;

-- Composite indexes for common filter combinations
CREATE INDEX idx_audit_events_user_created ON audit_events(user_id, created_at DESC);
CREATE INDEX idx_audit_events_entity_created ON audit_events(entity_type, entity_id, created_at DESC);

-- Add comment
COMMENT ON TABLE audit_events IS 'Immutable audit log for all system actions';
COMMENT ON COLUMN audit_events.signature IS 'HMAC-SHA256 signature for tamper detection';
COMMENT ON COLUMN audit_events.changes IS 'JSON object with before/after values';
COMMENT ON COLUMN audit_events.metadata IS 'Additional context-specific metadata';
