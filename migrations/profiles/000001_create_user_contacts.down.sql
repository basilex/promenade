-- Drop trigger
DROP TRIGGER IF EXISTS update_user_contacts_updated_at ON user_contacts;

-- Drop indexes
DROP INDEX IF EXISTS idx_user_contacts_verified;
DROP INDEX IF EXISTS idx_user_contacts_public;
DROP INDEX IF EXISTS idx_user_contacts_active;
DROP INDEX IF EXISTS idx_user_contacts_type;
DROP INDEX IF EXISTS idx_user_contacts_user_id;

-- Drop table
DROP TABLE IF EXISTS user_contacts;
