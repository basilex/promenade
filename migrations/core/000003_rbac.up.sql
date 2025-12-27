-- Create permissions table
CREATE TABLE IF NOT EXISTS core_permissions (
    id UUID PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- Create roles table
CREATE TABLE IF NOT EXISTS core_roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create role_permissions junction table (many-to-many)
CREATE TABLE IF NOT EXISTS core_role_permissions (
    role_id UUID NOT NULL REFERENCES core_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES core_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Create user_roles junction table (many-to-many with metadata)
CREATE TABLE IF NOT EXISTS core_user_roles (
    user_id UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES core_roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES core_users(id) ON DELETE SET NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_core_permissions_resource ON core_permissions(resource);
CREATE INDEX idx_core_permissions_action ON core_permissions(action);
CREATE INDEX idx_role_core_permissions_role ON core_role_permissions(role_id);
CREATE INDEX idx_role_core_permissions_permission ON core_role_permissions(permission_id);
CREATE INDEX idx_core_user_roles_user ON core_user_roles(user_id);
CREATE INDEX idx_core_user_roles_role ON core_user_roles(role_id);
CREATE INDEX idx_core_user_roles_expires ON core_user_roles(expires_at) WHERE expires_at IS NOT NULL;

-- Add comments for documentation
COMMENT ON TABLE core_permissions IS 'Defines granular permissions in resource:action format';
COMMENT ON TABLE core_roles IS 'Groups permissions into roles for assignment to users';
COMMENT ON TABLE core_role_permissions IS 'Maps permissions to roles (many-to-many)';
COMMENT ON TABLE core_user_roles IS 'Assigns roles to users with optional expiration';

COMMENT ON COLUMN core_permissions.resource IS 'Resource name (e.g., posts, users, comments)';
COMMENT ON COLUMN core_permissions.action IS 'Action name (e.g., create, read, update, delete, or *)';
COMMENT ON COLUMN core_roles.is_system IS 'System roles cannot be deleted';
COMMENT ON COLUMN core_user_roles.expires_at IS 'Optional expiration date for temporary role assignments';
COMMENT ON COLUMN core_user_roles.assigned_by IS 'User who assigned this role';

-- Insert wildcard permission for admin role
INSERT INTO core_permissions (id, resource, action, description, created_at)
VALUES (
    uuid_v7(),
    '*',
    '*',
    'Full access to all resources and actions (superadmin)',
    NOW()
);

-- Insert common permissions
INSERT INTO core_permissions (id, resource, action, description, created_at)
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

-- Create system roles (no superadmin - system user has admin role with full permissions)
-- Insert system roles (immutable)
INSERT INTO core_roles (id, name, display_name, description, is_system, created_at, updated_at)
VALUES
    (uuid_v7(), 'system', 'System Administrator', 'Full system access with all permissions (superadmin)', TRUE, NOW(), NOW()),
    (uuid_v7(), 'admin', 'Administrator', 'Full administrative access', TRUE, NOW(), NOW()),
    (uuid_v7(), 'moderator', 'Moderator', 'Can moderate user content and comments', TRUE, NOW(), NOW()),
    (uuid_v7(), 'developer', 'Developer', 'Development and debugging access', TRUE, NOW(), NOW()),
    (uuid_v7(), 'support', 'Support Agent', 'Customer support access', TRUE, NOW(), NOW()),
    (uuid_v7(), 'viewer', 'Viewer', 'Read-only access for reporting', TRUE, NOW(), NOW()),
    (uuid_v7(), 'user', 'User', 'Regular user with basic permissions', TRUE, NOW(), NOW()),
    (uuid_v7(), 'guest', 'Guest', 'Limited read-only access', TRUE, NOW(), NOW());

-- Assign permissions to system role (superadmin with wildcard)
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'system' AND p.resource = '*' AND p.action = '*';

-- Assign permissions to admin (all permissions via wildcard)
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'admin' AND p.resource = '*' AND p.action = '*';

-- Assign permissions to moderator (content moderation)
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'moderator' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action IN ('read', 'update', 'delete', 'list')) OR
    (p.resource = 'comments' AND p.action = '*') OR
    (p.resource = 'profiles' AND p.action IN ('read', 'list'))
);

-- Assign permissions to developer (full read + debug access)
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'developer' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action IN ('read', 'list')) OR
    (p.resource = 'comments' AND p.action IN ('read', 'list')) OR
    (p.resource = 'profiles' AND p.action IN ('read', 'list'))
);

-- Assign permissions to support (customer support access)
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'support' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action = 'read') OR
    (p.resource = 'comments' AND p.action = 'read') OR
    (p.resource = 'profiles' AND p.action = 'read')
);

-- Assign permissions to viewer (read-only analytics)
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'viewer' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action IN ('read', 'list')) OR
    (p.resource = 'comments' AND p.action IN ('read', 'list')) OR
    (p.resource = 'profiles' AND p.action IN ('read', 'list'))
);

-- Assign permissions to user
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'user' AND (
    (p.resource = 'posts' AND p.action IN ('create', 'read')) OR
    (p.resource = 'comments' AND p.action IN ('create', 'read')) OR
    (p.resource = 'profiles' AND p.action IN ('create', 'read'))
);

-- Assign permissions to guest
INSERT INTO core_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM core_roles r, core_permissions p
WHERE r.name = 'guest' AND (
    (p.resource = 'posts' AND p.action = 'read') OR
    (p.resource = 'comments' AND p.action = 'read') OR
    (p.resource = 'profiles' AND p.action = 'read')
);

-- ============================================================================
-- Assign default roles to users created in migration 000002
-- ============================================================================
-- Maps users to their appropriate roles for RBAC testing and demonstration.
-- All assignments are performed by system@promenade.com (bootstrap admin).
-- ============================================================================

-- 1. System Administrator (superadmin with wildcard permissions)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    u.id  -- Self-assigned during bootstrap
FROM core_users u
CROSS JOIN core_roles r
WHERE u.email = 'system@promenade.com' AND r.name = 'system'
ON CONFLICT DO NOTHING;

-- 2. Main Administrator (full admin access)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM core_users WHERE email = 'system@promenade.com')
FROM core_users u
CROSS JOIN core_roles r
WHERE u.email = 'admin@promenade.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- 3. Content Moderator (moderation permissions)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM core_users WHERE email = 'system@promenade.com')
FROM core_users u
CROSS JOIN core_roles r
WHERE u.email = 'moderator@promenade.com' AND r.name = 'moderator'
ON CONFLICT DO NOTHING;

-- 4. Developer (development and debugging access)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM core_users WHERE email = 'system@promenade.com')
FROM core_users u
CROSS JOIN core_roles r
WHERE u.email = 'developer@promenade.com' AND r.name = 'developer'
ON CONFLICT DO NOTHING;

-- 5. Support Agent (customer support access)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM core_users WHERE email = 'system@promenade.com')
FROM core_users u
CROSS JOIN core_roles r
WHERE u.email = 'support@promenade.com' AND r.name = 'support'
ON CONFLICT DO NOTHING;

-- 6. Report Viewer (read-only analytics)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM core_users WHERE email = 'system@promenade.com')
FROM core_users u
CROSS JOIN core_roles r
WHERE u.email = 'viewer@promenade.com' AND r.name = 'viewer'
ON CONFLICT DO NOTHING;

-- 7. Project Owner (regular user with full access to own data)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM core_users WHERE email = 'system@promenade.com')
FROM core_users u
CROSS JOIN core_roles r
WHERE u.email = 'alexander.vasilenko@gmail.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

-- 8. Assign default user role to any other existing users (fallback)
INSERT INTO core_user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM core_users WHERE email = 'system@promenade.com')
FROM core_users u
CROSS JOIN core_roles r
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
