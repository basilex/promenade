-- ============================================================================
-- Identity Context: Authorization (RBAC)
-- ============================================================================
-- Role-Based Access Control: Permissions, Roles, User-Role assignments
-- Part of Identity Bounded Context
-- ============================================================================

-- Permissions Table
CREATE TABLE identity_permissions (
    id          UUID PRIMARY KEY DEFAULT uuid_v7(),
    resource    VARCHAR(50) NOT NULL,
    action      VARCHAR(50) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    UNIQUE(resource, action)
);

CREATE INDEX idx_identity_permissions_resource ON identity_permissions(resource);
CREATE INDEX idx_identity_permissions_action ON identity_permissions(action);

COMMENT ON TABLE identity_permissions IS 'Granular permissions in resource:action format';
COMMENT ON COLUMN identity_permissions.resource IS 'Resource name (e.g., posts, users, *)';
COMMENT ON COLUMN identity_permissions.action IS 'Action name (e.g., create, read, update, delete, *)';

-- Roles Table
CREATE TABLE identity_roles (
    id           UUID PRIMARY KEY DEFAULT uuid_v7(),
    name         VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description  TEXT,
    is_system    BOOLEAN DEFAULT false NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_identity_roles_name ON identity_roles(name);
CREATE INDEX idx_identity_roles_system ON identity_roles(is_system);

CREATE TRIGGER trg_identity_roles_updated_at
    BEFORE UPDATE ON identity_roles
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

COMMENT ON TABLE identity_roles IS 'Groups of permissions for assignment to users';
COMMENT ON COLUMN identity_roles.is_system IS 'System roles cannot be deleted';

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

-- Common permissions
INSERT INTO identity_permissions (id, resource, action, description, created_at)
VALUES 
    -- Users
    (uuid_v7(), 'users', 'create', 'Create new users', NOW()),
    (uuid_v7(), 'users', 'read', 'Read user information', NOW()),
    (uuid_v7(), 'users', 'update', 'Update user information', NOW()),
    (uuid_v7(), 'users', 'delete', 'Delete users', NOW()),
    (uuid_v7(), 'users', 'list', 'List all users', NOW()),
    (uuid_v7(), 'users', 'ban', 'Ban/suspend users', NOW()),
    (uuid_v7(), 'users', '*', 'All user operations', NOW()),
    
    -- Posts
    (uuid_v7(), 'posts', 'create', 'Create new posts', NOW()),
    (uuid_v7(), 'posts', 'read', 'Read posts', NOW()),
    (uuid_v7(), 'posts', 'update', 'Update posts', NOW()),
    (uuid_v7(), 'posts', 'delete', 'Delete posts', NOW()),
    (uuid_v7(), 'posts', 'list', 'List posts', NOW()),
    (uuid_v7(), 'posts', '*', 'All post operations', NOW()),
    
    -- Comments
    (uuid_v7(), 'comments', 'create', 'Create comments', NOW()),
    (uuid_v7(), 'comments', 'read', 'Read comments', NOW()),
    (uuid_v7(), 'comments', 'update', 'Update comments', NOW()),
    (uuid_v7(), 'comments', 'delete', 'Delete comments', NOW()),
    (uuid_v7(), 'comments', 'list', 'List comments', NOW()),
    (uuid_v7(), 'comments', '*', 'All comment operations', NOW()),
    
    -- Profiles
    (uuid_v7(), 'profiles', 'create', 'Create profiles', NOW()),
    (uuid_v7(), 'profiles', 'read', 'Read profiles', NOW()),
    (uuid_v7(), 'profiles', 'update', 'Update profiles', NOW()),
    (uuid_v7(), 'profiles', 'delete', 'Delete profiles', NOW()),
    (uuid_v7(), 'profiles', 'list', 'List profiles', NOW()),
    (uuid_v7(), 'profiles', '*', 'All profile operations', NOW()),
    
    -- Roles
    (uuid_v7(), 'roles', 'create', 'Create roles', NOW()),
    (uuid_v7(), 'roles', 'read', 'Read roles', NOW()),
    (uuid_v7(), 'roles', 'update', 'Update roles', NOW()),
    (uuid_v7(), 'roles', 'delete', 'Delete roles', NOW()),
    (uuid_v7(), 'roles', 'list', 'List roles', NOW()),
    (uuid_v7(), 'roles', 'assign', 'Assign roles to users', NOW()),
    (uuid_v7(), 'roles', '*', 'All role operations', NOW());

-- ============================================================================
-- Seed Data: System Roles
-- ============================================================================

INSERT INTO identity_roles (id, name, display_name, description, is_system, created_at, updated_at)
VALUES
    (uuid_v7(), 'system', 'System Administrator', 'Full system access with all permissions (superadmin)', TRUE, NOW(), NOW()),
    (uuid_v7(), 'admin', 'Administrator', 'Full administrative access', TRUE, NOW(), NOW()),
    (uuid_v7(), 'moderator', 'Moderator', 'Can moderate user content and comments', TRUE, NOW(), NOW()),
    (uuid_v7(), 'developer', 'Developer', 'Development and debugging access', TRUE, NOW(), NOW()),
    (uuid_v7(), 'support', 'Support Agent', 'Customer support access', TRUE, NOW(), NOW()),
    (uuid_v7(), 'viewer', 'Viewer', 'Read-only access for reporting', TRUE, NOW(), NOW()),
    (uuid_v7(), 'user', 'User', 'Regular user with basic permissions', TRUE, NOW(), NOW()),
    (uuid_v7(), 'guest', 'Guest', 'Limited read-only access', TRUE, NOW(), NOW());

-- ============================================================================
-- Seed Data: Role-Permission Assignments
-- ============================================================================

-- System role: wildcard permission
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'system' AND p.resource = '*' AND p.action = '*';

-- Admin role: wildcard permission
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'admin' AND p.resource = '*' AND p.action = '*';

-- Moderator: content moderation
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'moderator' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action IN ('read', 'update', 'delete', 'list')) OR
    (p.resource = 'comments' AND p.action = '*') OR
    (p.resource = 'profiles' AND p.action IN ('read', 'list'))
);

-- Developer: full read + debug access
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'developer' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action IN ('read', 'list')) OR
    (p.resource = 'comments' AND p.action IN ('read', 'list')) OR
    (p.resource = 'profiles' AND p.action IN ('read', 'list'))
);

-- Support: customer support access
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'support' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action = 'read') OR
    (p.resource = 'comments' AND p.action = 'read') OR
    (p.resource = 'profiles' AND p.action = 'read')
);

-- Viewer: read-only analytics
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'viewer' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action IN ('read', 'list')) OR
    (p.resource = 'comments' AND p.action IN ('read', 'list')) OR
    (p.resource = 'profiles' AND p.action IN ('read', 'list'))
);

-- User: basic permissions
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'user' AND (
    (p.resource = 'posts' AND p.action IN ('create', 'read')) OR
    (p.resource = 'comments' AND p.action IN ('create', 'read')) OR
    (p.resource = 'profiles' AND p.action IN ('create', 'read'))
);

-- Guest: minimal read access
INSERT INTO identity_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM identity_roles r, identity_permissions p
WHERE r.name = 'guest' AND (
    (p.resource = 'posts' AND p.action = 'read') OR
    (p.resource = 'comments' AND p.action = 'read') OR
    (p.resource = 'profiles' AND p.action = 'read')
);

-- ============================================================================
-- Seed Data: Assign Roles to Default Users
-- ============================================================================

-- 1. System Administrator (superadmin)
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), u.id  -- Self-assigned during bootstrap
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email = 'system@promenade.com' AND r.name = 'system'
ON CONFLICT DO NOTHING;

-- 2. Admin User
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), (SELECT id FROM identity_users WHERE email = 'system@promenade.com')
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email = 'admin@promenade.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- 3. Moderator
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), (SELECT id FROM identity_users WHERE email = 'system@promenade.com')
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email = 'moderator@promenade.com' AND r.name = 'moderator'
ON CONFLICT DO NOTHING;

-- 4. Developer
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), (SELECT id FROM identity_users WHERE email = 'system@promenade.com')
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email = 'developer@promenade.com' AND r.name = 'developer'
ON CONFLICT DO NOTHING;

-- 5. Support Agent
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), (SELECT id FROM identity_users WHERE email = 'system@promenade.com')
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email = 'support@promenade.com' AND r.name = 'support'
ON CONFLICT DO NOTHING;

-- 6. Report Viewer
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), (SELECT id FROM identity_users WHERE email = 'system@promenade.com')
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email = 'viewer@promenade.com' AND r.name = 'viewer'
ON CONFLICT DO NOTHING;

-- 7. Project Owner (alexander.vasilenko@gmail.com)
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), (SELECT id FROM identity_users WHERE email = 'system@promenade.com')
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email = 'alexander.vasilenko@gmail.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

-- 8. Fallback: Assign 'user' role to any other existing users
INSERT INTO identity_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT u.id, r.id, NOW(), (SELECT id FROM identity_users WHERE email = 'system@promenade.com')
FROM identity_users u
CROSS JOIN identity_roles r
WHERE u.email NOT IN (
    'system@promenade.com',
    'admin@promenade.com',
    'moderator@promenade.com',
    'developer@promenade.com',
    'support@promenade.com',
    'viewer@promenade.com',
    'alexander.vasilenko@gmail.com'
) AND r.name = 'user'
ON CONFLICT DO NOTHING;
