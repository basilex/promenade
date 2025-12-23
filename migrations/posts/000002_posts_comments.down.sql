-- Rollback: Drop posts_comments table and related objects

-- Drop trigger
DROP TRIGGER IF EXISTS trg_posts_comments_updated_at ON posts_comments;

-- Drop indexes (will be dropped automatically with table, but explicit for clarity)
DROP INDEX IF EXISTS idx_posts_comments_user_created;
DROP INDEX IF EXISTS idx_posts_comments_deleted_at;
DROP INDEX IF EXISTS idx_posts_comments_parent_created;
DROP INDEX IF EXISTS idx_posts_comments_post_created;
DROP INDEX IF EXISTS idx_posts_comments_created_at;
DROP INDEX IF EXISTS idx_posts_comments_parent_id;
DROP INDEX IF EXISTS idx_posts_comments_user_id;
DROP INDEX IF EXISTS idx_posts_comments_post_id;

-- Drop table (CASCADE will drop all dependent objects)
DROP TABLE IF EXISTS posts_comments CASCADE;
