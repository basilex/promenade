-- Rollback: Drop authorization tables
DROP TABLE IF EXISTS identity_user_roles CASCADE;
DROP TABLE IF EXISTS identity_role_permissions CASCADE;
DROP TABLE IF EXISTS identity_roles CASCADE;
DROP TABLE IF EXISTS identity_permissions CASCADE;
