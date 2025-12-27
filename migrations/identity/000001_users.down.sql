-- Rollback: Drop identity_users table and user_status type
DROP TRIGGER IF EXISTS trg_identity_users_updated_at ON identity_users;
DROP TABLE IF EXISTS identity_users CASCADE;
DROP TYPE IF EXISTS user_status CASCADE;
