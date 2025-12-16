-- ============================================================================
-- Authentication Schema Migration
-- ============================================================================
-- This migration creates the complete authentication infrastructure:
-- 1. Users table with email verification
-- 2. Sessions table for refresh tokens (JWT)
-- 3. Password reset tokens
-- 4. Email verification tokens
-- 5. Login attempts tracking (security)
-- ============================================================================

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
CREATE TABLE users (
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
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- Trigger for users
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

COMMENT ON TABLE users IS 'User accounts with status-based lifecycle management';
COMMENT ON COLUMN users.id IS 'UUID v7 primary key (time-ordered)';
COMMENT ON COLUMN users.email IS 'Unique email address for login';
COMMENT ON COLUMN users.password IS 'Bcrypt hashed password (cost 10-12)';
COMMENT ON COLUMN users.status IS 'Current account status (see user_status enum)';
COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified (NULL = not verified)';
COMMENT ON COLUMN users.suspended_reason IS 'Reason for suspension/ban (if applicable)';
COMMENT ON COLUMN users.suspended_until IS 'Auto-reactivation date for temporary suspensions';

-- ----------------------------------------------------------------------------
-- 3. SESSIONS TABLE (for refresh tokens)
-- ----------------------------------------------------------------------------
CREATE TABLE sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(512) UNIQUE NOT NULL,
    user_agent    TEXT,
    ip_address    INET,
    expires_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for sessions
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

COMMENT ON TABLE sessions IS 'Active user sessions with refresh tokens';
COMMENT ON COLUMN sessions.refresh_token IS 'Hashed refresh token for JWT rotation';
COMMENT ON COLUMN sessions.expires_at IS 'Session expiration (typically 7-30 days)';

-- ----------------------------------------------------------------------------
-- 4. PASSWORD RESET TOKENS
-- ----------------------------------------------------------------------------
CREATE TABLE password_reset_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    used       BOOLEAN DEFAULT false NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for password reset
CREATE INDEX idx_password_reset_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_token ON password_reset_tokens(token) WHERE NOT used;
CREATE INDEX idx_password_reset_expires ON password_reset_tokens(expires_at) WHERE NOT used;

COMMENT ON TABLE password_reset_tokens IS 'One-time password reset tokens';
COMMENT ON COLUMN password_reset_tokens.token IS 'Hashed reset token sent via email';
COMMENT ON COLUMN password_reset_tokens.expires_at IS 'Token expiration (typically 1 hour)';
COMMENT ON COLUMN password_reset_tokens.used IS 'Prevents token reuse';

-- ----------------------------------------------------------------------------
-- 5. EMAIL VERIFICATION TOKENS
-- ----------------------------------------------------------------------------
CREATE TABLE email_verification_tokens (
    id         UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    used       BOOLEAN DEFAULT false NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for email verification
CREATE INDEX idx_email_verification_user_id ON email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_token ON email_verification_tokens(token) WHERE NOT used;
CREATE INDEX idx_email_verification_expires ON email_verification_tokens(expires_at) WHERE NOT used;

COMMENT ON TABLE email_verification_tokens IS 'One-time email verification tokens';
COMMENT ON COLUMN email_verification_tokens.token IS 'Hashed verification token sent via email';
COMMENT ON COLUMN email_verification_tokens.expires_at IS 'Token expiration (typically 24 hours)';

-- ----------------------------------------------------------------------------
-- 6. LOGIN ATTEMPTS (brute force protection)
-- ----------------------------------------------------------------------------
CREATE TABLE login_attempts (
    id          UUID PRIMARY KEY DEFAULT uuid_v7(),
    email       VARCHAR(255) NOT NULL,
    ip_address  INET NOT NULL,
    success     BOOLEAN NOT NULL,
    user_agent  TEXT,
    attempted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for login attempts
CREATE INDEX idx_login_attempts_email ON login_attempts(email, attempted_at DESC);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address, attempted_at DESC);
CREATE INDEX idx_login_attempts_attempted_at ON login_attempts(attempted_at DESC);

COMMENT ON TABLE login_attempts IS 'Login attempts log for security monitoring and rate limiting';
COMMENT ON COLUMN login_attempts.success IS 'True if login succeeded, false if failed';

-- ----------------------------------------------------------------------------
-- DEFAULT USERS
-- ----------------------------------------------------------------------------

-- Insert default system administrator
INSERT INTO users (email, name, password, status, email_verified_at)
VALUES (
    'system@promenade.com',
    'System Administrator',
    crypt('passw0rd', gen_salt('bf', 10)),
    'active',
    NOW()
);

-- Insert default user
INSERT INTO users (email, name, password, status, email_verified_at)
VALUES (
    'alexander.vasilenko@gmail.com',
    'Alexander Vasilenko',
    crypt('03041965', gen_salt('bf', 10)),
    'active',
    NOW()
);

-- ----------------------------------------------------------------------------
-- CLEANUP FUNCTIONS
-- ----------------------------------------------------------------------------

-- Function to cleanup expired tokens (call via cron/scheduler)
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
RETURNS void AS $$
BEGIN
    -- Delete expired sessions
    DELETE FROM sessions WHERE expires_at < NOW();
    
    -- Delete old used password reset tokens (older than 7 days)
    DELETE FROM password_reset_tokens 
    WHERE used = true AND used_at < NOW() - INTERVAL '7 days';
    
    -- Delete expired unused password reset tokens
    DELETE FROM password_reset_tokens 
    WHERE used = false AND expires_at < NOW();
    
    -- Delete old used email verification tokens (older than 7 days)
    DELETE FROM email_verification_tokens 
    WHERE used = true AND used_at < NOW() - INTERVAL '7 days';
    
    -- Delete expired unused email verification tokens
    DELETE FROM email_verification_tokens 
    WHERE used = false AND expires_at < NOW();
    
    -- Delete old login attempts (older than 30 days)
    DELETE FROM login_attempts WHERE attempted_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_tokens() IS 
    'Cleanup expired/used tokens and old login attempts. Run daily via cron.';
