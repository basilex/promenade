-- ============================================================================
-- User Posts Schema Migration
-- ============================================================================
-- Creates table for user-generated content (blog posts, articles)
-- Supports draft/published workflow, SEO optimization, and engagement metrics
-- ============================================================================

-- ----------------------------------------------------------------------------
-- POST STATUS ENUM
-- ----------------------------------------------------------------------------
CREATE TYPE post_status AS ENUM (
    'draft',      -- Not yet published
    'published',  -- Live and visible
    'archived',   -- No longer active but preserved
    'scheduled'   -- Scheduled for future publication
);

COMMENT ON TYPE post_status IS 'Post publication status lifecycle';

-- ----------------------------------------------------------------------------
-- USER_POSTS TABLE
-- ----------------------------------------------------------------------------
CREATE TABLE user_posts (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Content
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL, -- URL-friendly title (auto-generated from title)
    excerpt TEXT, -- Short description/preview (max 500 chars)
    content TEXT NOT NULL,
    
    -- Media
    featured_image JSONB, -- {url, alt, width, height, thumbnail_url}
    
    -- Status & Visibility
    status post_status DEFAULT 'draft' NOT NULL,
    is_public BOOLEAN DEFAULT true NOT NULL,
    is_featured BOOLEAN DEFAULT false NOT NULL, -- Featured/pinned posts
    is_comments_enabled BOOLEAN DEFAULT true NOT NULL,
    
    -- Publishing
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE, -- For scheduled posts
    
    -- SEO & Organization
    tags JSONB DEFAULT '[]'::jsonb, -- Array of tag strings
    categories JSONB DEFAULT '[]'::jsonb, -- Array of category IDs
    meta_title VARCHAR(70), -- SEO title (default: title)
    meta_description VARCHAR(160), -- SEO description (default: excerpt)
    meta_keywords TEXT[], -- SEO keywords array
    
    -- Engagement Metrics
    view_count INTEGER DEFAULT 0 NOT NULL,
    like_count INTEGER DEFAULT 0 NOT NULL,
    comment_count INTEGER DEFAULT 0 NOT NULL,
    share_count INTEGER DEFAULT 0 NOT NULL,
    
    -- Reading time (auto-calculated from content)
    reading_time_minutes INTEGER DEFAULT 0 NOT NULL,
    
    -- Soft delete support
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Unique slug per user (same slug allowed for different users)
    UNIQUE (user_id, slug)
);

-- ----------------------------------------------------------------------------
-- INDEXES
-- ----------------------------------------------------------------------------
-- User's posts lookup
CREATE INDEX idx_user_posts_user_id ON user_posts(user_id) WHERE deleted_at IS NULL;

-- Published posts discovery
CREATE INDEX idx_user_posts_published ON user_posts(published_at DESC) 
    WHERE status = 'published' AND is_public = true AND deleted_at IS NULL;

-- Status filtering
CREATE INDEX idx_user_posts_status ON user_posts(status) WHERE deleted_at IS NULL;

-- Featured posts
CREATE INDEX idx_user_posts_featured ON user_posts(is_featured, published_at DESC) 
    WHERE is_featured = true AND status = 'published' AND deleted_at IS NULL;

-- Slug lookups (for SEO-friendly URLs)
CREATE INDEX idx_user_posts_slug ON user_posts(slug) WHERE deleted_at IS NULL;

-- Full-text search on title and content
CREATE INDEX idx_user_posts_search ON user_posts USING gin(to_tsvector('english', title || ' ' || COALESCE(excerpt, '') || ' ' || content))
    WHERE deleted_at IS NULL;

-- Tag search (JSONB)
CREATE INDEX idx_user_posts_tags ON user_posts USING gin(tags) WHERE deleted_at IS NULL;

-- Scheduled posts
CREATE INDEX idx_user_posts_scheduled ON user_posts(scheduled_at) 
    WHERE status = 'scheduled' AND scheduled_at IS NOT NULL AND deleted_at IS NULL;

-- Soft delete support
CREATE INDEX idx_user_posts_deleted_at ON user_posts(deleted_at) WHERE deleted_at IS NOT NULL;

-- ----------------------------------------------------------------------------
-- TRIGGER
-- ----------------------------------------------------------------------------
CREATE TRIGGER trg_user_posts_updated_at
    BEFORE UPDATE ON user_posts
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- ----------------------------------------------------------------------------
-- COMMENTS
-- ----------------------------------------------------------------------------
COMMENT ON TABLE user_posts IS 'User-generated blog posts and articles with SEO optimization';
COMMENT ON COLUMN user_posts.id IS 'UUID v7 primary key (time-ordered)';
COMMENT ON COLUMN user_posts.user_id IS 'Author reference';
COMMENT ON COLUMN user_posts.title IS 'Post title (max 255 chars)';
COMMENT ON COLUMN user_posts.slug IS 'URL-friendly identifier (unique per user)';
COMMENT ON COLUMN user_posts.excerpt IS 'Short preview/description (max 500 chars recommended)';
COMMENT ON COLUMN user_posts.content IS 'Full post content (Markdown or HTML)';
COMMENT ON COLUMN user_posts.featured_image IS 'JSONB: {url, alt, width, height, thumbnail_url}';
COMMENT ON COLUMN user_posts.status IS 'Publication status (draft, published, archived, scheduled)';
COMMENT ON COLUMN user_posts.is_public IS 'Public visibility (false = private/unlisted)';
COMMENT ON COLUMN user_posts.is_featured IS 'Featured/highlighted post';
COMMENT ON COLUMN user_posts.published_at IS 'Publication timestamp (NULL for drafts)';
COMMENT ON COLUMN user_posts.scheduled_at IS 'Future publication timestamp';
COMMENT ON COLUMN user_posts.tags IS 'JSONB array of tag strings ["golang", "api"]';
COMMENT ON COLUMN user_posts.categories IS 'JSONB array of category IDs';
COMMENT ON COLUMN user_posts.meta_title IS 'SEO title (default: title)';
COMMENT ON COLUMN user_posts.meta_description IS 'SEO description (default: excerpt)';
COMMENT ON COLUMN user_posts.view_count IS 'Total views counter';
COMMENT ON COLUMN user_posts.like_count IS 'Total likes counter';
COMMENT ON COLUMN user_posts.reading_time_minutes IS 'Estimated reading time';
COMMENT ON COLUMN user_posts.deleted_at IS 'Soft delete timestamp';
