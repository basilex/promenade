-- Drop trigger
DROP TRIGGER IF EXISTS update_profiles_contacts_updated_at ON profiles_contacts;

-- Drop indexes
DROP INDEX IF EXISTS idx_profiles_contacts_verified;
DROP INDEX IF EXISTS idx_profiles_contacts_public;
DROP INDEX IF EXISTS idx_profiles_contacts_active;
DROP INDEX IF EXISTS idx_profiles_contacts_type;
DROP INDEX IF EXISTS idx_profiles_contacts_user_id;

-- Drop table
DROP TABLE IF EXISTS profiles_contacts;
