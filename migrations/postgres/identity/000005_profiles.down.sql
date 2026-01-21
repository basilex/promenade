-- Drop identity_profiles table
DROP INDEX IF EXISTS idx_profiles_created_at;
DROP INDEX IF EXISTS idx_profiles_public;
DROP INDEX IF EXISTS idx_profiles_user_id;
DROP TABLE IF EXISTS identity_profiles;
