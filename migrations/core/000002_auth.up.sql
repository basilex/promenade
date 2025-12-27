-- ============================================================================
-- Authentication Schema Migration
-- ============================================================================
-- This migration creates the complete authentication infrastructure:
-- 1. Users table with email verification
-- 2. User sessions table for refresh tokens (JWT)
-- 3. Password reset tokens
-- 4. Email verification tokens
-- 5. Login attempts tracking (security)
-- 6. Default users with bcrypt passwords (using pgcrypto)
-- ============================================================================

-- Enable pgcrypto extension for bcrypt password hashing
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ----------------------------------------------------------------------------
-- 1. USER STATUS ENUM
-- ----------------------------------------------------------------------------
CREATE TYPE user_status AS ENUM (
    'unverified',   -- Registered but email not verified
    'active',       -- Email verified and account active
    'suspended',    -- Temporarily blocked (can be reactivated)
    'banned',       -- Permanently blocked
    'inactive'      -- Deactivated by user (can be reactivated)
);

COMMENT ON TYPE user_status IS 'User account status lifecycle';

-- ----------------------------------------------------------------------------
-- 2. USERS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE core_users (
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

-- Indexes for users
CREATE INDEX idx_core_users_email ON core_users(email);
CREATE INDEX idx_core_users_status ON core_users(status);
CREATE INDEX idx_core_users_created_at ON core_users(created_at DESC);

-- Trigger for users
CREATE TRIGGER trg_core_users_updated_at
    BEFORE UPDATE ON core_users
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

COMMENT ON TABLE core_users IS 'User accounts with status-based lifecycle management';
COMMENT ON COLUMN core_users.id IS 'UUID v7 primary key (time-ordered)';
COMMENT ON COLUMN core_users.email IS 'Unique email address for login';
COMMENT ON COLUMN core_users.password IS 'Bcrypt hashed password (cost 10-12)';
COMMENT ON COLUMN core_users.status IS 'Current account status (see user_status enum)';
COMMENT ON COLUMN core_users.email_verified_at IS 'Timestamp when email was verified (NULL = not verified)';
COMMENT ON COLUMN core_users.suspended_reason IS 'Reason for suspension/ban (if applicable)';
COMMENT ON COLUMN core_users.suspended_until IS 'Auto-reactivation date for temporary suspensions';

-- ----------------------------------------------------------------------------
-- 3. USER SESSIONS TABLE (for refresh tokens)
-- ----------------------------------------------------------------------------
CREATE TABLE core_user_sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id       UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(512) UNIQUE NOT NULL,
    user_agent    TEXT,
    ip_address    INET,
    expires_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for user_sessions
CREATE INDEX idx_core_user_sessions_user_id ON core_user_sessions(user_id);
CREATE INDEX idx_core_user_sessions_refresh_token ON core_user_sessions(refresh_token);
CREATE INDEX idx_core_user_sessions_expires_at ON core_user_sessions(expires_at);
CREATE INDEX idx_core_user_sessions_user_created ON core_user_sessions(user_id, created_at DESC);

COMMENT ON TABLE core_user_sessions IS 'Active user sessions with refresh tokens (JWT rotation pattern)';
COMMENT ON COLUMN core_user_sessions.id IS 'Session ID (UUID v7, time-ordered)';
COMMENT ON COLUMN core_user_sessions.user_id IS 'User who owns this session';
COMMENT ON COLUMN core_user_sessions.refresh_token IS 'Hashed refresh token (SHA256) for JWT rotation';
COMMENT ON COLUMN core_user_sessions.user_agent IS 'Client user agent for device tracking';
COMMENT ON COLUMN core_user_sessions.ip_address IS 'Client IP address for security audit';
COMMENT ON COLUMN core_user_sessions.expires_at IS 'Session expiration (typically 7-30 days)';
COMMENT ON COLUMN core_user_sessions.created_at IS 'Session creation timestamp';

-- ----------------------------------------------------------------------------
-- 4. PASSWORD RESET TOKENS
-- ----------------------------------------------------------------------------
CREATE TABLE core_password_reset_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id    UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    used       BOOLEAN DEFAULT false NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for password reset
CREATE INDEX idx_core_password_reset_user_id ON core_password_reset_tokens(user_id);
CREATE INDEX idx_core_password_reset_token ON core_password_reset_tokens(token) WHERE NOT used;
CREATE INDEX idx_core_password_reset_expires ON core_password_reset_tokens(expires_at) WHERE NOT used;

COMMENT ON TABLE core_password_reset_tokens IS 'One-time password reset tokens';
COMMENT ON COLUMN core_password_reset_tokens.token IS 'Hashed reset token sent via email';
COMMENT ON COLUMN core_password_reset_tokens.expires_at IS 'Token expiration (typically 1 hour)';
COMMENT ON COLUMN core_password_reset_tokens.used IS 'Prevents token reuse';

-- ----------------------------------------------------------------------------
-- 5. EMAIL VERIFICATION TOKENS
-- ----------------------------------------------------------------------------
CREATE TABLE core_email_verification_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id    UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    used       BOOLEAN DEFAULT false NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for email verification
CREATE INDEX idx_core_email_verification_user_id ON core_email_verification_tokens(user_id);
CREATE INDEX idx_core_email_verification_token ON core_email_verification_tokens(token) WHERE NOT used;
CREATE INDEX idx_core_email_verification_expires ON core_email_verification_tokens(expires_at) WHERE NOT used;

COMMENT ON TABLE core_email_verification_tokens IS 'One-time email verification tokens';
COMMENT ON COLUMN core_email_verification_tokens.token IS 'Hashed verification token sent via email';
COMMENT ON COLUMN core_email_verification_tokens.expires_at IS 'Token expiration (typically 24 hours)';

-- ----------------------------------------------------------------------------
-- 6. LOGIN ATTEMPTS (brute force protection)
-- ----------------------------------------------------------------------------
CREATE TABLE core_login_attempts (
    id          UUID PRIMARY KEY DEFAULT uuid_v7(),
    email       VARCHAR(255) NOT NULL,
    ip_address  INET NOT NULL,
    success     BOOLEAN NOT NULL,
    user_agent  TEXT,
    attempted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for login attempts
CREATE INDEX idx_core_login_attempts_email ON core_login_attempts(email, attempted_at DESC);
CREATE INDEX idx_core_login_attempts_ip ON core_login_attempts(ip_address, attempted_at DESC);
CREATE INDEX idx_core_login_attempts_attempted_at ON core_login_attempts(attempted_at DESC);

COMMENT ON TABLE core_login_attempts IS 'Login attempts log for security monitoring and rate limiting';
COMMENT ON COLUMN core_login_attempts.success IS 'True if login succeeded, false if failed';

-- ============================================================================
-- 7. DEFAULT USERS (for development and RBAC setup)
-- ============================================================================
-- These users are created with predefined UUID v7 IDs for consistency.
-- Roles will be assigned in migration 000003_rbac.up.sql
-- 
-- USER HIERARCHY (by access level):
-- 1. system       → SuperAdmin (unrestricted system access)
-- 2. admin        → Admin (full application management)
-- 3. moderator    → Moderator (content moderation)
-- 4. developer    → Developer (API access, debugging)
-- 5. support      → Support (customer help, read-only user data)
-- 6. viewer       → Viewer (read-only analytics/reports)
-- 7. owner        → User (project owner - Alexander Vasilenko)
--
-- SECURITY:
-- - Passwords hashed using pgcrypto crypt() with bcrypt (cost 10)
-- - Default password: "Promenade2025!"
-- - ⚠️  CHANGE ALL PASSWORDS IN PRODUCTION!
-- ============================================================================

INSERT INTO core_users (id, email, name, password, status, email_verified_at, created_at, updated_at)
VALUES
    -- System Administrator (super admin, all permissions)
    (
        '019b5eb7-b592-768e-b687-5519410c6852'::uuid,
        'system@promenade.com',
        'System Administrator',
        crypt('Promenade2025!', gen_salt('bf', 10)),
        'active',
        NOW(),
        NOW(),
        NOW()
    ),
    
    -- Main Administrator (full admin access)
    (
        '019b5eb7-b5bf-749a-8eec-a9d707486a07'::uuid,
        'admin@promenade.com',
        'Administrator',
        crypt('Promenade2025!', gen_salt('bf', 10)),
        'active',
        NOW(),
        NOW(),
        NOW()
    ),
    
    -- Content Moderator (content management)
    (
        '019b5eb7-b5eb-7768-ae5c-08a88ee29757'::uuid,
        'moderator@promenade.com',
        'Content Moderator',
        crypt('Promenade2025!', gen_salt('bf', 10)),
        'active',
        NOW(),
        NOW(),
        NOW()
    ),
    
    -- Developer (API access, debugging)
    (
        '019b5eb7-b610-709f-b0db-ca7f3feef2e6'::uuid,
        'developer@promenade.com',
        'API Developer',
        crypt('Promenade2025!', gen_salt('bf', 10)),
        'active',
        NOW(),
        NOW(),
        NOW()
    ),
    
    -- Support Agent (customer help)
    (
        '019b5eb7-b632-78f6-bb46-3f1d2870faf3'::uuid,
        'support@promenade.com',
        'Customer Support',
        crypt('Promenade2025!', gen_salt('bf', 10)),
        'active',
        NOW(),
        NOW(),
        NOW()
    ),
    
    -- Analytics Viewer (read-only reports)
    (
        '019b5eb7-b658-7e67-8a45-e7af4bf0b3b3'::uuid,
        'viewer@promenade.com',
        'Analytics Viewer',
        crypt('Promenade2025!', gen_salt('bf', 10)),
        'active',
        NOW(),
        NOW(),
        NOW()
    ),
    
    -- Project Owner (full access to own data)
    (
        '019b5eb7-b614-7d30-bb16-aeceaadd1374'::uuid,
        'alexander.vasilenko@gmail.com',
        'Alexander Vasilenko',
        crypt('Promenade2025!', gen_salt('bf', 10)),
        'active',
        NOW(),
        NOW(),
        NOW()
    );

COMMENT ON COLUMN core_users.password IS 
    'Bcrypt hashed password (cost 10). Default: "Promenade2025!" - CHANGE IN PRODUCTION!';

-- ============================================================================
-- 8. CLEANUP FUNCTIONS
-- ============================================================================

-- Function to cleanup expired tokens (call via cron/scheduler)
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS void AS $$
BEGIN
    -- Delete expired sessions
    DELETE FROM core_user_sessions WHERE expires_at < NOW();
    
    -- Delete old used password reset tokens (older than 7 days)
    DELETE FROM core_password_reset_tokens 
    WHERE used = true AND used_at < NOW() - INTERVAL '7 days';
    
    -- Delete expired unused password reset tokens
    DELETE FROM core_password_reset_tokens 
    WHERE used = false AND expires_at < NOW();
    
    -- Delete old used email verification tokens (older than 7 days)
    DELETE FROM core_email_verification_tokens 
    WHERE used = true AND used_at < NOW() - INTERVAL '7 days';
    
    -- Delete expired unused email verification tokens
    DELETE FROM core_email_verification_tokens 
    WHERE used = false AND expires_at < NOW();
    
    -- Delete old login attempts (older than 30 days)
    DELETE FROM core_login_attempts WHERE attempted_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_tokens() IS 
    'Cleanup expired/used tokens and old login attempts. Run daily via cron.';

