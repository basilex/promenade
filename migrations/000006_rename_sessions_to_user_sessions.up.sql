-- Rename sessions table to user_sessions for naming consistency
-- This aligns with user_contacts, user_profiles pattern

ALTER TABLE sessions RENAME TO user_sessions;

-- Update indexes and constraints
ALTER INDEX sessions_pkey RENAME TO user_sessions_pkey;
ALTER INDEX idx_sessions_user_id RENAME TO idx_user_sessions_user_id;
ALTER INDEX idx_sessions_refresh_token RENAME TO idx_user_sessions_refresh_token;
ALTER INDEX idx_sessions_expires_at RENAME TO idx_user_sessions_expires_at;
ALTER INDEX sessions_refresh_token_key RENAME TO user_sessions_refresh_token_key;

-- Note: No trigger to update as sessions table doesn't have updated_at column

-- Update comments
COMMENT ON TABLE user_sessions IS 'User authentication sessions with JWT refresh tokens';
COMMENT ON COLUMN user_sessions.id IS 'Unique session identifier (UUID v7)';
COMMENT ON COLUMN user_sessions.user_id IS 'Reference to users table';
COMMENT ON COLUMN user_sessions.refresh_token IS 'JWT refresh token (unique)';
COMMENT ON COLUMN user_sessions.user_agent IS 'Client user agent string';
COMMENT ON COLUMN user_sessions.ip_address IS 'Client IP address';
COMMENT ON COLUMN user_sessions.expires_at IS 'Session expiration timestamp';
COMMENT ON COLUMN user_sessions.created_at IS 'Session creation timestamp';
