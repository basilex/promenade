-- ============================================================================
-- Identity Context: Authentication
-- ============================================================================
-- Session management, password reset, email verification, login attempts
-- Part of Identity Bounded Context
-- ============================================================================

-- User Sessions (Refresh Token Rotation Pattern for JWT)
CREATE TABLE identity_user_sessions (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(512) UNIQUE NOT NULL,
    user_agent    TEXT,
    ip_address    INET,
    expires_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_identity_sessions_user_id ON identity_user_sessions(user_id);
CREATE INDEX idx_identity_sessions_token ON identity_user_sessions(refresh_token);
CREATE INDEX idx_identity_sessions_expires ON identity_user_sessions(expires_at);
CREATE INDEX idx_identity_sessions_user_created ON identity_user_sessions(user_id, created_at DESC);

COMMENT ON TABLE identity_user_sessions IS 'User sessions with refresh tokens (JWT rotation)';
COMMENT ON COLUMN identity_user_sessions.refresh_token IS 'Hashed refresh token (SHA256)';
COMMENT ON COLUMN identity_user_sessions.expires_at IS 'Session expiration (7-30 days)';

-- Password Reset Tokens
CREATE TABLE identity_password_reset_tokens (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    token      VARCHAR(255) UNIQUE NOT NULL,
    used       BOOLEAN DEFAULT false NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_identity_password_reset_user ON identity_password_reset_tokens(user_id);
CREATE INDEX idx_identity_password_reset_token ON identity_password_reset_tokens(token) WHERE NOT used;
CREATE INDEX idx_identity_password_reset_expires ON identity_password_reset_tokens(expires_at) WHERE NOT used;

COMMENT ON TABLE identity_password_reset_tokens IS 'One-time password reset tokens';
COMMENT ON COLUMN identity_password_reset_tokens.token IS 'Random token (typically UUID or secure random string)';
COMMENT ON COLUMN identity_password_reset_tokens.used IS 'Token can only be used once';

-- Email Verification Tokens
CREATE TABLE identity_email_verification_tokens (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    email      VARCHAR(255) NOT NULL,
    token      VARCHAR(255) UNIQUE NOT NULL,
    used       BOOLEAN DEFAULT false NOT NULL,
    used_at    TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_identity_email_verification_user ON identity_email_verification_tokens(user_id);
CREATE INDEX idx_identity_email_verification_token ON identity_email_verification_tokens(token) WHERE NOT used;
CREATE INDEX idx_identity_email_verification_expires ON identity_email_verification_tokens(expires_at) WHERE NOT used;

COMMENT ON TABLE identity_email_verification_tokens IS 'Email verification tokens (registration + email change)';
COMMENT ON COLUMN identity_email_verification_tokens.email IS 'Email being verified (can differ from current user email)';
COMMENT ON COLUMN identity_email_verification_tokens.token IS 'Verification token (sent via email)';

-- Login Attempts (Security: Rate Limiting & Breach Detection)
CREATE TABLE identity_login_attempts (
    id         TEXT PRIMARY KEY,
    email      VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    success    BOOLEAN DEFAULT false NOT NULL,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE INDEX idx_identity_login_attempts_email ON identity_login_attempts(email, created_at DESC);
CREATE INDEX idx_identity_login_attempts_ip ON identity_login_attempts(ip_address, created_at DESC);
CREATE INDEX idx_identity_login_attempts_email_ip ON identity_login_attempts(email, ip_address, created_at DESC);

COMMENT ON TABLE identity_login_attempts IS 'Login attempt tracking (security: rate limiting, breach detection)';
COMMENT ON COLUMN identity_login_attempts.success IS 'true = successful login, false = failed attempt';
COMMENT ON COLUMN identity_login_attempts.created_at IS 'Timestamp for rate limiting calculations';
