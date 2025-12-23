-- Create profiles_profiles table (one-to-one relationship with users)
-- Follows 2025+ best practices: UUID v7, JSONB for flexible data, proper privacy controls

CREATE TABLE IF NOT EXISTS profiles_profiles (
    -- Primary key: UUID v7 for time-ordered performance
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- Foreign key: one-to-one relationship with users table
    user_id UUID NOT NULL UNIQUE REFERENCES core_users(id) ON DELETE CASCADE,
    
    -- Personal Information
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    middle_name VARCHAR(100),
    display_name VARCHAR(100), -- How user wants to be addressed
    nickname VARCHAR(50) UNIQUE, -- @username style, must be unique
    
    -- Biographical Information
    bio TEXT, -- Short bio, up to ~500 chars recommended
    date_of_birth DATE,
    gender VARCHAR(20) CHECK (gender IN ('male', 'female', 'non-binary', 'other', 'prefer-not-to-say')),
    
    -- Location & Localization
    country_id UUID REFERENCES core_countries(id) ON DELETE SET NULL, -- Link to countries table
    city VARCHAR(100),
    timezone VARCHAR(50) DEFAULT 'UTC', -- IANA timezone (e.g., 'Europe/Kiev', 'America/New_York')
    locale VARCHAR(10) DEFAULT 'en-US', -- Language/region (e.g., 'en-US', 'uk-UA', 'ru-RU')
    
    -- Visual Identity
    avatar_url VARCHAR(500), -- URL to profile picture
    cover_url VARCHAR(500), -- URL to cover/banner image
    
    -- Social & Web Presence (JSONB for flexibility)
    -- Example: {"github": "username", "twitter": "handle", "linkedin": "profile-url"}
    social_links JSONB DEFAULT '{}'::jsonb,
    
    -- Website & Professional
    website_url VARCHAR(500),
    company VARCHAR(100),
    job_title VARCHAR(100),
    
    -- Privacy & Verification
    is_public BOOLEAN NOT NULL DEFAULT true, -- Profile visibility
    is_verified BOOLEAN NOT NULL DEFAULT false, -- Verified account badge
    show_email BOOLEAN NOT NULL DEFAULT false, -- Email visibility setting
    show_location BOOLEAN NOT NULL DEFAULT true, -- Location visibility setting
    show_birthday BOOLEAN NOT NULL DEFAULT false, -- Birthday visibility setting
    
    -- User Preferences (JSONB for extensibility)
    -- Example: {"theme": "dark", "notifications": {"email": true}, "dashboard_layout": "grid"}
    preferences JSONB DEFAULT '{}'::jsonb,
    
    -- Metadata & Statistics
    profile_views_count INTEGER NOT NULL DEFAULT 0,
    followers_count INTEGER NOT NULL DEFAULT 0,
    following_count INTEGER NOT NULL DEFAULT 0,
    
    -- Moderation & Safety
    is_banned BOOLEAN NOT NULL DEFAULT false,
    ban_reason TEXT,
    banned_at TIMESTAMPTZ,
    banned_by UUID REFERENCES core_users(id) ON DELETE SET NULL,
    
    -- Timestamps
    last_seen_at TIMESTAMPTZ, -- Track user activity
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT profiles_profiles_nickname_length CHECK (LENGTH(nickname) >= 3),
    CONSTRAINT profiles_profiles_bio_length CHECK (LENGTH(bio) <= 1000),
    CONSTRAINT profiles_profiles_display_name_not_empty CHECK (LENGTH(TRIM(display_name)) > 0 OR display_name IS NULL)
);

-- Indexes for performance
CREATE INDEX idx_profiles_profiles_user_id ON profiles_profiles(user_id); -- Already UNIQUE, but explicit for FK lookups
CREATE INDEX idx_profiles_profiles_nickname ON profiles_profiles(nickname) WHERE nickname IS NOT NULL; -- Username lookups
CREATE INDEX idx_profiles_profiles_country_id ON profiles_profiles(country_id) WHERE country_id IS NOT NULL; -- Location filtering
CREATE INDEX idx_profiles_profiles_is_public ON profiles_profiles(is_public) WHERE is_public = true; -- Public profiles
CREATE INDEX idx_profiles_profiles_is_verified ON profiles_profiles(is_verified) WHERE is_verified = true; -- Verified accounts
CREATE INDEX idx_profiles_profiles_last_seen_at ON profiles_profiles(last_seen_at DESC); -- Activity tracking
CREATE INDEX idx_profiles_profiles_created_at ON profiles_profiles(created_at DESC); -- New users

-- GIN indexes for JSONB columns (efficient querying of JSON data)
CREATE INDEX idx_profiles_profiles_social_links ON profiles_profiles USING GIN (social_links);
CREATE INDEX idx_profiles_profiles_preferences ON profiles_profiles USING GIN (preferences);

-- Trigger: auto-update updated_at timestamp
CREATE TRIGGER tg_profiles_profiles_updated_at
    BEFORE UPDATE ON profiles_profiles
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- Comments for documentation
COMMENT ON TABLE profiles_profiles IS 'User profile information with one-to-one relationship to users table. Contains personal info, preferences, and social links.';
COMMENT ON COLUMN profiles_profiles.user_id IS 'Foreign key to users table (one-to-one relationship). Unique constraint ensures one profile per user.';
COMMENT ON COLUMN profiles_profiles.nickname IS 'Unique username/handle (e.g., @johndoe). Used for profile URLs and mentions.';
COMMENT ON COLUMN profiles_profiles.social_links IS 'JSONB object containing social media links and handles. Example: {"github": "username", "twitter": "handle"}';
COMMENT ON COLUMN profiles_profiles.preferences IS 'JSONB object for user preferences and settings. Flexible schema for feature flags, UI preferences, etc.';
COMMENT ON COLUMN profiles_profiles.timezone IS 'IANA timezone identifier for localization (e.g., Europe/Kiev, America/New_York)';
COMMENT ON COLUMN profiles_profiles.locale IS 'Language and region code for localization (e.g., en-US, uk-UA, ru-RU)';
COMMENT ON COLUMN profiles_profiles.is_public IS 'Controls profile visibility. If false, only user and admins can see full profile.';
COMMENT ON COLUMN profiles_profiles.is_verified IS 'Verified account badge (e.g., for notable users, organizations).';
