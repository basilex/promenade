-- Drop identity_contacts table and related objects
DROP INDEX IF EXISTS idx_identity_contacts_user_type_primary;
DROP INDEX IF EXISTS idx_identity_contacts_user_type;
DROP INDEX IF EXISTS idx_identity_contacts_type;
DROP INDEX IF EXISTS idx_identity_contacts_user_id;
DROP TABLE IF EXISTS identity_contacts;
