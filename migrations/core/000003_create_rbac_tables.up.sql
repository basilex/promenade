-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create role_permissions junction table (many-to-many)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Create user_roles junction table (many-to-many with metadata)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_user_roles_expires ON user_roles(expires_at) WHERE expires_at IS NOT NULL;

-- Add comments for documentation
COMMENT ON TABLE permissions IS 'Defines granular permissions in resource:action format';
COMMENT ON TABLE roles IS 'Groups permissions into roles for assignment to users';
COMMENT ON TABLE role_permissions IS 'Maps permissions to roles (many-to-many)';
COMMENT ON TABLE user_roles IS 'Assigns roles to users with optional expiration';

COMMENT ON COLUMN permissions.resource IS 'Resource name (e.g., posts, users, comments)';
COMMENT ON COLUMN permissions.action IS 'Action name (e.g., create, read, update, delete, or *)';
COMMENT ON COLUMN roles.is_system IS 'System roles cannot be deleted';
COMMENT ON COLUMN user_roles.expires_at IS 'Optional expiration date for temporary role assignments';
COMMENT ON COLUMN user_roles.assigned_by IS 'User who assigned this role';

-- Insert wildcard permission for superadmin
INSERT INTO permissions (id, resource, action, description, created_at)
VALUES (
    uuid_v7(),
    '*',
    '*',
    'Full access to all resources and actions',
    NOW()
);

-- Insert common permissions
INSERT INTO permissions (id, resource, action, description, created_at)
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

-- Create system roles
INSERT INTO roles (id, name, display_name, description, is_system, created_at, updated_at)
VALUES
    (uuid_v7(), 'superadmin', 'Super Administrator', 'Full system access with all permissions', TRUE, NOW(), NOW()),
    (uuid_v7(), 'admin', 'Administrator', 'Administrative access to manage users and content', TRUE, NOW(), NOW()),
    (uuid_v7(), 'moderator', 'Moderator', 'Can moderate user content and comments', TRUE, NOW(), NOW()),
    (uuid_v7(), 'user', 'User', 'Regular user with basic permissions', TRUE, NOW(), NOW()),
    (uuid_v7(), 'guest', 'Guest', 'Limited read-only access', TRUE, NOW(), NOW());

-- Assign permissions to superadmin (all permissions via wildcard)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'superadmin' AND p.resource = '*' AND p.action = '*';

-- Assign permissions to admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin' AND (
    (p.resource = 'users' AND p.action = '*') OR
    (p.resource = 'posts' AND p.action = '*') OR
    (p.resource = 'comments' AND p.action = '*') OR
    (p.resource = 'profiles' AND p.action = '*') OR
    (p.resource = 'roles' AND p.action IN ('read', 'list', 'assign'))
);

-- Assign permissions to moderator
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'moderator' AND (
    (p.resource = 'users' AND p.action IN ('read', 'list')) OR
    (p.resource = 'posts' AND p.action IN ('read', 'update', 'delete', 'list')) OR
    (p.resource = 'comments' AND p.action = '*') OR
    (p.resource = 'profiles' AND p.action IN ('read', 'list'))
);

-- Assign permissions to user
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user' AND (
    (p.resource = 'posts' AND p.action IN ('create', 'read')) OR
    (p.resource = 'comments' AND p.action IN ('create', 'read')) OR
    (p.resource = 'profiles' AND p.action IN ('create', 'read'))
);

-- Assign permissions to guest
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'guest' AND (
    (p.resource = 'posts' AND p.action = 'read') OR
    (p.resource = 'comments' AND p.action = 'read') OR
    (p.resource = 'profiles' AND p.action = 'read')
);

-- ============================================================================
-- Assign default roles to existing users
-- ============================================================================
-- Maps users created in migration 2 to their appropriate roles.
-- This enables RBAC testing and demonstration with realistic permission sets.
-- ============================================================================

-- 1. Assign superadmin role to system@promenade.com (bootstrap user)
INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    u.id  -- Self-assigned during initial setup
FROM users u
CROSS JOIN roles r
WHERE u.email = 'system@promenade.com' AND r.name = 'superadmin'
ON CONFLICT DO NOTHING;

-- 2. Assign superadmin role to superadmin@promenade.com
INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM users WHERE email = 'system@promenade.com')
FROM users u
CROSS JOIN roles r
WHERE u.email = 'superadmin@promenade.com' AND r.name = 'superadmin'
ON CONFLICT DO NOTHING;

-- 3. Assign admin role to admin@promenade.com
INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM users WHERE email = 'system@promenade.com')
FROM users u
CROSS JOIN roles r
WHERE u.email = 'admin@promenade.com' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- 4. Assign moderator role to moderator@promenade.com
INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM users WHERE email = 'system@promenade.com')
FROM users u
CROSS JOIN roles r
WHERE u.email = 'moderator@promenade.com' AND r.name = 'moderator'
ON CONFLICT DO NOTHING;

-- 5. Assign user role to alexander.vasilenko@gmail.com (regular user)
INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM users WHERE email = 'system@promenade.com')
FROM users u
CROSS JOIN roles r
WHERE u.email = 'alexander.vasilenko@gmail.com' AND r.name = 'user'
ON CONFLICT DO NOTHING;

-- 6. Assign user role to any other existing users (default fallback)
INSERT INTO user_roles (user_id, role_id, assigned_at, assigned_by)
SELECT 
    u.id,
    r.id,
    NOW(),
    (SELECT id FROM users WHERE email = 'system@promenade.com')
FROM users u
CROSS JOIN roles r
WHERE u.email NOT IN (
    'system@promenade.com',
    'superadmin@promenade.com', 
    'admin@promenade.com',
    'moderator@promenade.com',
    'alexander.vasilenko@gmail.com'
) AND r.name = 'user'
ON CONFLICT DO NOTHING;
