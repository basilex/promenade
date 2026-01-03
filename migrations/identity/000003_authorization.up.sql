-- ============================================================================
-- Identity Context: Authorization (RBAC)
-- ============================================================================
-- Role-Based Access Control: Permissions, Roles, User-Role assignments
-- Part of Identity Bounded Context
-- ============================================================================

-- Permissions Table
CREATE TABLE identity_permissions (
    id          UUID PRIMARY KEY,
    name        VARCHAR(101) GENERATED ALWAYS AS (resource || ':' || action) STORED NOT NULL,
    resource    VARCHAR(50) NOT NULL,
    action      VARCHAR(50) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at  TIMESTAMP WITH TIME ZONE,
    UNIQUE(resource, action)
);

CREATE INDEX idx_identity_permissions_resource ON identity_permissions(resource);
CREATE INDEX idx_identity_permissions_action ON identity_permissions(action);
CREATE INDEX idx_identity_permissions_deleted_at ON identity_permissions(deleted_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_identity_permissions_name ON identity_permissions(name) WHERE deleted_at IS NULL;

COMMENT ON TABLE identity_permissions IS 'Granular permissions in resource:action format';
COMMENT ON COLUMN identity_permissions.name IS 'Permission name in resource:action format (e.g., users:read)';
COMMENT ON COLUMN identity_permissions.resource IS 'Resource name (e.g., users, roles, customers)';
COMMENT ON COLUMN identity_permissions.action IS 'Action name (e.g., read, write, delete)';
COMMENT ON COLUMN identity_permissions.deleted_at IS 'Soft delete timestamp (NULL = active)';

-- Roles Table
CREATE TABLE identity_roles (
    id           UUID PRIMARY KEY,
    name         VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description  TEXT,
    is_system    BOOLEAN DEFAULT FALSE NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at   TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_identity_roles_name ON identity_roles(name);
CREATE INDEX idx_identity_roles_system ON identity_roles(is_system);
CREATE INDEX idx_identity_roles_deleted_at ON identity_roles(deleted_at) WHERE deleted_at IS NULL;

COMMENT ON TABLE identity_roles IS 'Groups of permissions for assignment to users';
COMMENT ON COLUMN identity_roles.is_system IS 'System roles cannot be deleted';
COMMENT ON COLUMN identity_roles.deleted_at IS 'Soft delete timestamp (NULL = active)';

-- Role-Permissions Junction (Many-to-Many)
CREATE TABLE identity_role_permissions (
    role_id       UUID NOT NULL REFERENCES identity_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES identity_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_identity_role_permissions_role ON identity_role_permissions(role_id);
CREATE INDEX idx_identity_role_permissions_perm ON identity_role_permissions(permission_id);

COMMENT ON TABLE identity_role_permissions IS 'Maps permissions to roles (many-to-many)';

-- User-Roles Junction (Many-to-Many with metadata)
CREATE TABLE identity_user_roles (
    user_id     UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    role_id     UUID NOT NULL REFERENCES identity_roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    assigned_by UUID REFERENCES identity_users(id) ON DELETE SET NULL,
    expires_at  TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_identity_user_roles_user ON identity_user_roles(user_id);
CREATE INDEX idx_identity_user_roles_role ON identity_user_roles(role_id);
CREATE INDEX idx_identity_user_roles_expires ON identity_user_roles(expires_at) WHERE expires_at IS NOT NULL;

COMMENT ON TABLE identity_user_roles IS 'Assigns roles to users with optional expiration';
COMMENT ON COLUMN identity_user_roles.assigned_by IS 'User who assigned this role (audit trail)';
COMMENT ON COLUMN identity_user_roles.expires_at IS 'Optional expiration for temporary assignments';

-- ============================================================================
-- Seed Data: Permissions, Roles, and Assignments
-- ============================================================================
-- Note: Seed data removed for DB-agnostic approach.
-- Use application layer to populate RBAC data with proper UUID v7 generation.

