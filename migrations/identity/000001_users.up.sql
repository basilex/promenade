-- ============================================================================
-- Identity Context: Users
-- ============================================================================
-- User aggregate: Core entity representing user identity
-- Part of Identity Bounded Context
-- ============================================================================

-- User status lifecycle
CREATE TYPE user_status AS ENUM (
    'unverified',   -- Registered but email not verified
    'active',       -- Email verified and account active
    'suspended',    -- Temporarily blocked (can be reactivated)
    'banned',       -- Permanently blocked
    'inactive'      -- Deactivated by user (can be reactivated)
);

COMMENT ON TYPE user_status IS 'User account status lifecycle (Identity Context)';

-- Users table (User aggregate root)
CREATE TABLE identity_users (
    id                UUID PRIMARY KEY DEFAULT uuid_v7(),
    email             VARCHAR(255) UNIQUE NOT NULL,
    name              VARCHAR(255) NOT NULL,
    password          VARCHAR(255) NOT NULL,
    status            user_status DEFAULT 'unverified' NOT NULL,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    suspended_reason  TEXT,
    suspended_until   TIMESTAMP WITH TIME ZONE,
    last_login_at     TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_identity_users_email ON identity_users(email);
CREATE INDEX idx_identity_users_status ON identity_users(status);
CREATE INDEX idx_identity_users_created_at ON identity_users(created_at DESC);

-- Trigger for auto-updating updated_at
CREATE TRIGGER trg_identity_users_updated_at
    BEFORE UPDATE ON identity_users
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- Comments for documentation
COMMENT ON TABLE identity_users IS 'User aggregate root (Identity Bounded Context)';
COMMENT ON COLUMN identity_users.id IS 'UUID v7 - time-ordered primary key';
COMMENT ON COLUMN identity_users.email IS 'Unique email address for authentication';
COMMENT ON COLUMN identity_users.password IS 'Bcrypt hashed password (using pgcrypto)';
COMMENT ON COLUMN identity_users.status IS 'Account lifecycle status';
COMMENT ON COLUMN identity_users.email_verified_at IS 'Email verification timestamp (NULL = unverified)';
COMMENT ON COLUMN identity_users.suspended_until IS 'Suspension expiry (NULL = permanent)';

-- Insert default users for development/testing
-- Password: "password" hashed with bcrypt (cost=10)
INSERT INTO identity_users (id, email, name, password, status, email_verified_at, created_at, updated_at)
VALUES
    (uuid_v7(), 'system@promenade.com', 'System Administrator', crypt('password', gen_salt('bf', 10)), 'active', NOW(), NOW(), NOW()),
    (uuid_v7(), 'admin@promenade.com', 'Admin User', crypt('password', gen_salt('bf', 10)), 'active', NOW(), NOW(), NOW()),
    (uuid_v7(), 'moderator@promenade.com', 'Moderator User', crypt('password', gen_salt('bf', 10)), 'active', NOW(), NOW(), NOW()),
    (uuid_v7(), 'developer@promenade.com', 'Developer User', crypt('password', gen_salt('bf', 10)), 'active', NOW(), NOW(), NOW()),
    (uuid_v7(), 'support@promenade.com', 'Support Agent', crypt('password', gen_salt('bf', 10)), 'active', NOW(), NOW(), NOW()),
    (uuid_v7(), 'viewer@promenade.com', 'Report Viewer', crypt('password', gen_salt('bf', 10)), 'active', NOW(), NOW(), NOW()),
    (uuid_v7(), 'alexander.vasilenko@gmail.com', 'Alexander Vasilenko', crypt('password', gen_salt('bf', 10)), 'active', NOW(), NOW(), NOW());
