-- Migration: ui_metadata
-- Context: ui
-- Created: 2026-01-16
-- Purpose: UI form definitions (database-agnostic JSON storage)

CREATE TABLE ui_form_definitions (
    id TEXT PRIMARY KEY,
    form_id TEXT UNIQUE NOT NULL,
    entity_type TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,

    -- JSON metadata (stored as TEXT for database-agnostic support)
    layout TEXT NOT NULL,
    fields TEXT NOT NULL,
    validation TEXT,
    events TEXT,
    permissions TEXT,
    i18n TEXT,

    version INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    tenant_id TEXT,

    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE ui_form_versions (
    id TEXT PRIMARY KEY,
    form_id TEXT NOT NULL REFERENCES ui_form_definitions(id),
    version INTEGER NOT NULL,

    -- Historical snapshot
    metadata TEXT NOT NULL,
    change_log TEXT,

    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(form_id, version)
);

CREATE INDEX idx_ui_forms_form_id ON ui_form_definitions(form_id);
CREATE INDEX idx_ui_forms_entity ON ui_form_definitions(entity_type);
CREATE INDEX idx_ui_forms_tenant ON ui_form_definitions(tenant_id);
CREATE INDEX idx_ui_forms_active ON ui_form_definitions(is_active);
CREATE INDEX idx_ui_forms_created_at ON ui_form_definitions(created_at);

CREATE INDEX idx_ui_form_versions_form ON ui_form_versions(form_id);
CREATE INDEX idx_ui_form_versions_created_at ON ui_form_versions(created_at);
