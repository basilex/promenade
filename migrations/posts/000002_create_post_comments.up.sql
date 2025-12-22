-- Create post_comments table for storing user comments on blog posts
-- Supports nested comments (replies) via parent_id

CREATE TABLE IF NOT EXISTS post_comments (
    -- Primary key (UUID v7 for time-ordered IDs)
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- Foreign keys
    post_id UUID NOT NULL REFERENCES user_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES post_comments(id) ON DELETE CASCADE, -- For nested comments/replies
    
    -- Content
    content TEXT NOT NULL CHECK (char_length(content) >= 1 AND char_length(content) <= 5000),
    
    -- Edit tracking
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    edited_at TIMESTAMP WITH TIME ZONE,
    
    -- Engagement metrics
    like_count INTEGER NOT NULL DEFAULT 0 CHECK (like_count >= 0),
    reply_count INTEGER NOT NULL DEFAULT 0 CHECK (reply_count >= 0),
    
    -- Soft delete support
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_edit_timestamp CHECK (
        (is_edited = FALSE AND edited_at IS NULL) OR
        (is_edited = TRUE AND edited_at IS NOT NULL)
    ),
    CONSTRAINT no_self_parent CHECK (id != parent_id)
);

-- Indexes for performance
CREATE INDEX idx_post_comments_post_id ON post_comments(post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_user_id ON post_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_parent_id ON post_comments(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_created_at ON post_comments(created_at DESC) WHERE deleted_at IS NULL;

-- Composite index for pagination on specific post
CREATE INDEX idx_post_comments_post_created ON post_comments(post_id, created_at DESC) WHERE deleted_at IS NULL AND parent_id IS NULL;

-- Index for nested replies
CREATE INDEX idx_post_comments_parent_created ON post_comments(parent_id, created_at ASC) WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

-- Index for soft delete queries
CREATE INDEX idx_post_comments_deleted_at ON post_comments(deleted_at) WHERE deleted_at IS NOT NULL;

-- Index for user's comments
CREATE INDEX idx_post_comments_user_created ON post_comments(user_id, created_at DESC) WHERE deleted_at IS NULL;

-- Trigger to automatically update updated_at timestamp
CREATE TRIGGER trg_post_comments_updated_at
    BEFORE UPDATE ON post_comments
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- Comments
COMMENT ON TABLE post_comments IS 'User comments on blog posts with support for nested replies';
COMMENT ON COLUMN post_comments.id IS 'Primary key (UUID v7, time-ordered)';
COMMENT ON COLUMN post_comments.post_id IS 'Reference to the post being commented on';
COMMENT ON COLUMN post_comments.user_id IS 'Reference to the user who wrote the comment';
COMMENT ON COLUMN post_comments.parent_id IS 'Reference to parent comment (NULL for top-level comments)';
COMMENT ON COLUMN post_comments.content IS 'Comment text content (1-5000 characters)';
COMMENT ON COLUMN post_comments.is_edited IS 'Whether the comment has been edited';
COMMENT ON COLUMN post_comments.edited_at IS 'Timestamp of last edit (NULL if never edited)';
COMMENT ON COLUMN post_comments.like_count IS 'Number of likes on this comment';
COMMENT ON COLUMN post_comments.reply_count IS 'Number of direct replies to this comment';
COMMENT ON COLUMN post_comments.deleted_at IS 'Soft delete timestamp (NULL if not deleted)';
