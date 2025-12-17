-- Drop user_posts table and related objects

DROP TABLE IF EXISTS user_posts CASCADE;
DROP TYPE IF EXISTS post_status CASCADE;

-- Note: Indexes and triggers are automatically dropped with the table
