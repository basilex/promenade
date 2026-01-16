-- ============================================================================
-- Identity Context: Users
-- ============================================================================
-- User aggregate: Core entity representing user identity
-- Part of Identity Bounded Context
-- ============================================================================

-- User status lifecycle
CREATE TYPE user_status AS ENUM (
    'active',       -- Account active
    'inactive',     -- Deactivated by user (can be reactivated)
    'suspended',    -- Temporarily blocked (can be reactivated)
    'banned'        -- Permanently blocked
);

COMMENT ON TYPE user_status IS 'User account status lifecycle (Identity Context)';

-- Users table (User aggregate root)
CREATE TABLE identity_users (
    id                   TEXT PRIMARY KEY,
    email                VARCHAR(255) UNIQUE NOT NULL,
    password_hash        VARCHAR(255) NOT NULL,
    status               user_status DEFAULT 'active' NOT NULL,
    
    -- Email verification
    email_verified       BOOLEAN DEFAULT FALSE NOT NULL,
    email_verified_at    TIMESTAMP WITH TIME ZONE,
    
    -- Account locking (security)
    failed_login_count   INTEGER DEFAULT 0 NOT NULL,
    locked_until         TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    last_login_at        TIMESTAMP WITH TIME ZONE,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at           TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX idx_identity_users_email ON identity_users(LOWER(email));
CREATE INDEX idx_identity_users_status ON identity_users(status);
CREATE INDEX idx_identity_users_created_at ON identity_users(created_at DESC);
CREATE INDEX idx_identity_users_deleted_at ON identity_users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Comments for documentation
COMMENT ON TABLE identity_users IS 'User aggregate root (Identity Bounded Context)';
COMMENT ON COLUMN identity_users.id IS 'UUID v7 stored as TEXT (time-ordered primary key)';
COMMENT ON COLUMN identity_users.email IS 'Unique email address for authentication';
COMMENT ON COLUMN identity_users.password_hash IS 'Bcrypt hashed password';
COMMENT ON COLUMN identity_users.status IS 'Account lifecycle status';
COMMENT ON COLUMN identity_users.email_verified IS 'Whether email has been verified';
COMMENT ON COLUMN identity_users.email_verified_at IS 'Email verification timestamp';
COMMENT ON COLUMN identity_users.failed_login_count IS 'Number of consecutive failed login attempts';
COMMENT ON COLUMN identity_users.locked_until IS 'Account lock expiry timestamp';
COMMENT ON COLUMN identity_users.deleted_at IS 'Soft delete timestamp';

-- Note: IDs will be generated in Go via uuidv7.New()
-- These seed users are for initial setup only
