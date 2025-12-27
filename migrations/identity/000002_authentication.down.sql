-- Rollback: Drop authentication tables
DROP TABLE IF EXISTS identity_login_attempts CASCADE;
DROP TABLE IF EXISTS identity_email_verification_tokens CASCADE;
DROP TABLE IF EXISTS identity_password_reset_tokens CASCADE;
DROP TABLE IF EXISTS identity_user_sessions CASCADE;
