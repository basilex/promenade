-- Drop indexes
DROP INDEX IF EXISTS idx_audit_events_entity_created;
DROP INDEX IF EXISTS idx_audit_events_user_created;
DROP INDEX IF EXISTS idx_audit_events_request_id;
DROP INDEX IF EXISTS idx_audit_events_created_at;
DROP INDEX IF EXISTS idx_audit_events_action;
DROP INDEX IF EXISTS idx_audit_events_entity;
DROP INDEX IF EXISTS idx_audit_events_user_id;

-- Drop table
DROP TABLE IF EXISTS audit_events;
