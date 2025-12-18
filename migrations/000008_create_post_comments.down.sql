-- Rollback: Drop post_comments table and related objects

-- Drop trigger
DROP TRIGGER IF EXISTS trg_post_comments_updated_at ON post_comments;

-- Drop indexes (will be dropped automatically with table, but explicit for clarity)
DROP INDEX IF EXISTS idx_post_comments_user_created;
DROP INDEX IF EXISTS idx_post_comments_deleted_at;
DROP INDEX IF EXISTS idx_post_comments_parent_created;
DROP INDEX IF EXISTS idx_post_comments_post_created;
DROP INDEX IF EXISTS idx_post_comments_created_at;
DROP INDEX IF EXISTS idx_post_comments_parent_id;
DROP INDEX IF EXISTS idx_post_comments_user_id;
DROP INDEX IF EXISTS idx_post_comments_post_id;

-- Drop table (CASCADE will drop all dependent objects)
DROP TABLE IF EXISTS post_comments CASCADE;
