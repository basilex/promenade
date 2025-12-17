-- Rollback: Rename user_sessions back to sessions

ALTER TABLE user_sessions RENAME TO sessions;

-- Rollback indexes and constraints
ALTER INDEX user_sessions_pkey RENAME TO sessions_pkey;
ALTER INDEX idx_user_sessions_user_id RENAME TO idx_sessions_user_id;
ALTER INDEX idx_user_sessions_refresh_token RENAME TO idx_sessions_refresh_token;
ALTER INDEX idx_user_sessions_expires_at RENAME TO idx_sessions_expires_at;
ALTER INDEX user_sessions_refresh_token_key RENAME TO sessions_refresh_token_key;

-- Note: No trigger to restore as original sessions table doesn't have updated_at column

-- Rollback comments
COMMENT ON TABLE sessions IS 'User authentication sessions with JWT refresh tokens';
COMMENT ON COLUMN sessions.id IS 'Unique session identifier (UUID v7)';
COMMENT ON COLUMN sessions.user_id IS 'Reference to users table';
COMMENT ON COLUMN sessions.refresh_token IS 'JWT refresh token (unique)';
COMMENT ON COLUMN sessions.user_agent IS 'Client user agent string';
COMMENT ON COLUMN sessions.ip_address IS 'Client IP address';
COMMENT ON COLUMN sessions.expires_at IS 'Session expiration timestamp';
COMMENT ON COLUMN sessions.created_at IS 'Session creation timestamp';
