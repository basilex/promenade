-- Create posts_comment_likes table to track which users liked which comments
CREATE TABLE IF NOT EXISTS posts_comment_likes (
    comment_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (comment_id, user_id),
    CONSTRAINT fk_posts_comment_likes_comment FOREIGN KEY (comment_id) 
        REFERENCES posts_comments(id) ON DELETE CASCADE,
    CONSTRAINT fk_posts_comment_likes_user FOREIGN KEY (user_id) 
        REFERENCES core_users(id) ON DELETE CASCADE
);

-- Index for efficient lookups by user (to get all comments a user has liked)
CREATE INDEX idx_posts_comment_likes_user_id ON posts_comment_likes(user_id);

-- Index for efficient lookups by comment (to get all users who liked a comment)
CREATE INDEX idx_posts_comment_likes_comment_id ON posts_comment_likes(comment_id);

-- Index for created_at ordering
CREATE INDEX idx_posts_comment_likes_created_at ON posts_comment_likes(created_at DESC);

-- Add comment explaining the table
COMMENT ON TABLE posts_comment_likes IS 'Stores comment likes to prevent duplicate likes and enable unlike functionality';
COMMENT ON COLUMN posts_comment_likes.comment_id IS 'Reference to the liked comment';
COMMENT ON COLUMN posts_comment_likes.user_id IS 'Reference to the user who liked the comment';
COMMENT ON COLUMN posts_comment_likes.created_at IS 'When the like was created';
