-- Drop user_profiles table and all related objects

-- Drop indexes (will be dropped automatically with table, but explicit for clarity)
DROP INDEX IF EXISTS idx_user_profiles_preferences;
DROP INDEX IF EXISTS idx_user_profiles_social_links;
DROP INDEX IF EXISTS idx_user_profiles_created_at;
DROP INDEX IF EXISTS idx_user_profiles_last_seen_at;
DROP INDEX IF EXISTS idx_user_profiles_is_verified;
DROP INDEX IF EXISTS idx_user_profiles_is_public;
DROP INDEX IF EXISTS idx_user_profiles_country_id;
DROP INDEX IF EXISTS idx_user_profiles_nickname;
DROP INDEX IF EXISTS idx_user_profiles_user_id;

-- Drop trigger
DROP TRIGGER IF EXISTS tg_user_profiles_updated_at ON user_profiles;

-- Drop table (CASCADE will drop foreign key constraints)
DROP TABLE IF EXISTS user_profiles CASCADE;
