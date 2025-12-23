-- Drop profiles_profiles table and all related objects

-- Drop indexes (will be dropped automatically with table, but explicit for clarity)
DROP INDEX IF EXISTS idx_profiles_profiles_preferences;
DROP INDEX IF EXISTS idx_profiles_profiles_social_links;
DROP INDEX IF EXISTS idx_profiles_profiles_created_at;
DROP INDEX IF EXISTS idx_profiles_profiles_last_seen_at;
DROP INDEX IF EXISTS idx_profiles_profiles_is_verified;
DROP INDEX IF EXISTS idx_profiles_profiles_is_public;
DROP INDEX IF EXISTS idx_profiles_profiles_country_id;
DROP INDEX IF EXISTS idx_profiles_profiles_nickname;
DROP INDEX IF EXISTS idx_profiles_profiles_user_id;

-- Drop trigger
DROP TRIGGER IF EXISTS tg_profiles_profiles_updated_at ON profiles_profiles;

-- Drop table (CASCADE will drop foreign key constraints)
DROP TABLE IF EXISTS profiles_profiles CASCADE;
