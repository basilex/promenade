-- Drop cleanup function
DROP FUNCTION IF EXISTS cleanup_expired_tokens();

-- Drop tables in reverse order (respect foreign keys)
DROP TABLE IF EXISTS core_login_attempts CASCADE;
DROP TABLE IF EXISTS core_email_verification_tokens CASCADE;
DROP TABLE IF EXISTS core_password_reset_tokens CASCADE;
DROP TABLE IF EXISTS core_user_sessions CASCADE;
DROP TABLE IF EXISTS core_users CASCADE;

-- Drop enum type
DROP TYPE IF EXISTS user_status;
