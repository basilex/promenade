-- ============================================================================
-- Identity Context: Authorization (RBAC)
-- ============================================================================
-- Role-Based Access Control: Permissions, Roles, User-Role assignments
-- Part of Identity Bounded Context
-- ============================================================================

-- Permissions Table
CREATE TABLE identity_permissions (
    id          UUID PRIMARY KEY DEFAULT uuid_v7(),
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

CREATE TRIGGER trg_identity_permissions_updated_at
    BEFORE UPDATE ON identity_permissions
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

COMMENT ON TABLE identity_permissions IS 'Granular permissions in resource:action format';
COMMENT ON COLUMN identity_permissions.name IS 'Permission name in resource:action format (e.g., users:read)';
COMMENT ON COLUMN identity_permissions.resource IS 'Resource name (e.g., users, roles, customers)';
COMMENT ON COLUMN identity_permissions.action IS 'Action name (e.g., read, write, delete)';
COMMENT ON COLUMN identity_permissions.deleted_at IS 'Soft delete timestamp (NULL = active)';

-- Roles Table
CREATE TABLE identity_roles (
    id           UUID PRIMARY KEY DEFAULT uuid_v7(),
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

CREATE TRIGGER trg_identity_roles_updated_at
    BEFORE UPDATE ON identity_roles
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

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
-- Seed Data: Permissions
-- ============================================================================

-- Wildcard permission for superadmin
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES (uuid_v7(), '*', '*', 'Full access to all resources and actions (superadmin)', NOW());

-- User management permissions
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES
    (uuid_v7(), 'users', 'read', 'View user information', NOW()),
    (uuid_v7(), 'users', 'write', 'Create and update users', NOW()),
    (uuid_v7(), 'users', 'delete', 'Delete users', NOW()),
    (uuid_v7(), 'users', 'suspend', 'Suspend user accounts', NOW()),
    (uuid_v7(), 'users', 'ban', 'Ban user accounts', NOW()),
    (uuid_v7(), 'users', 'activate', 'Activate user accounts', NOW());

-- Role management permissions
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES
    (uuid_v7(), 'roles', 'read', 'View roles', NOW()),
    (uuid_v7(), 'roles', 'write', 'Create and update roles', NOW()),
    (uuid_v7(), 'roles', 'delete', 'Delete roles', NOW()),
    (uuid_v7(), 'roles', 'assign', 'Assign roles to users', NOW());

-- Permission management permissions
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES
    (uuid_v7(), 'permissions', 'read', 'View permissions', NOW()),
    (uuid_v7(), 'permissions', 'write', 'Create and update permissions', NOW()),
    (uuid_v7(), 'permissions', 'delete', 'Delete permissions', NOW()),
    (uuid_v7(), 'permissions', 'assign', 'Assign permissions to roles', NOW());

-- Contact management permissions
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES
    (uuid_v7(), 'contacts', 'read', 'View contacts', NOW()),
    (uuid_v7(), 'contacts', 'write', 'Create and update contacts', NOW()),
    (uuid_v7(), 'contacts', 'delete', 'Delete contacts', NOW()),
    (uuid_v7(), 'contacts', 'verify', 'Verify contacts', NOW());

-- Profile management permissions
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES
    (uuid_v7(), 'profiles', 'read', 'View profiles', NOW()),
    (uuid_v7(), 'profiles', 'write', 'Create and update profiles', NOW()),
    (uuid_v7(), 'profiles', 'delete', 'Delete profiles', NOW());

-- Customer management permissions
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES
    (uuid_v7(), 'customers', 'read', 'View customers', NOW()),
    (uuid_v7(), 'customers', 'write', 'Create and update customers', NOW()),
    (uuid_v7(), 'customers', 'delete', 'Delete customers', NOW()),
    (uuid_v7(), 'customers', 'manage', 'Full customer management', NOW());

-- ============================================================================
-- Seed Data: Roles
-- ============================================================================

-- System roles (cannot be deleted)
INSERT INTO identity_roles (id, name, display_name, description, is_system, created_at)
VALUES
    (uuid_v7(), 'superadmin', 'Super Administrator', 'Full system access with all permissions', TRUE, NOW()),
    (uuid_v7(), 'admin', 'Administrator', 'System administrator with most permissions', TRUE, NOW()),
    (uuid_v7(), 'manager', 'Manager', 'Team manager with user and customer management', TRUE, NOW()),
    (uuid_v7(), 'user', 'User', 'Regular user with basic permissions', TRUE, NOW()),
    (uuid_v7(), 'guest', 'Guest', 'Guest user with read-only access', TRUE, NOW());

-- ============================================================================
-- Seed Data: Role-Permission Assignments
-- ============================================================================

-- Superadmin: all permissions (wildcard)
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r
CROSS JOIN identity_permissions p
WHERE r.name = 'superadmin'
  AND p.resource = '*'
  AND p.action = '*';

-- Admin: all permissions except wildcard
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r
CROSS JOIN identity_permissions p
WHERE r.name = 'admin'
  AND NOT (p.resource = '*' AND p.action = '*');

-- Manager: user, customer, and profile management
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r
CROSS JOIN identity_permissions p
WHERE r.name = 'manager'
  AND p.resource IN ('users', 'customers', 'profiles', 'contacts')
  AND p.action IN ('read', 'write', 'manage');

-- User: own profile and contact management, read customers
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r
CROSS JOIN identity_permissions p
WHERE r.name = 'user'
  AND (
      (p.resource IN ('profiles', 'contacts') AND p.action IN ('read', 'write'))
      OR (p.resource = 'customers' AND p.action = 'read')
  );

-- Guest: read-only access
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r
CROSS JOIN identity_permissions p
WHERE r.name = 'guest'
  AND p.action = 'read'
  AND p.resource IN ('profiles', 'customers');
