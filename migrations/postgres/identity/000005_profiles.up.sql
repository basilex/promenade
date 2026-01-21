-- Create identity_profiles table
CREATE TABLE IF NOT EXISTS identity_profiles (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    
    -- Display Information
    display_name VARCHAR(100) NOT NULL,
    bio VARCHAR(500),
    avatar_url VARCHAR(500),
    
    -- Personal Information
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    middle_name VARCHAR(50),
    gender VARCHAR(20) NOT NULL DEFAULT 'not_specified',
    date_of_birth DATE,
    
    -- Localization
    timezone VARCHAR(100),
    language VARCHAR(2), -- ISO 639-1 code
    country VARCHAR(2),  -- ISO 3166-1 alpha-2 code
    
    -- Social Links
    website VARCHAR(500),
    linkedin VARCHAR(500),
    twitter VARCHAR(500),
    github VARCHAR(500),
    facebook VARCHAR(500),
    instagram VARCHAR(500),
    
    -- Privacy & Status
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Lifecycle
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT fk_profile_user FOREIGN KEY (user_id) 
        REFERENCES identity_users(id) ON DELETE CASCADE,
    CONSTRAINT chk_gender CHECK (gender IN ('male', 'female', 'other', 'not_specified')),
    CONSTRAINT chk_language_length CHECK (language IS NULL OR length(language) = 2),
    CONSTRAINT chk_country_length CHECK (country IS NULL OR length(country) = 2)
);

-- Indexes
CREATE INDEX idx_profiles_user_id ON identity_profiles(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_profiles_public ON identity_profiles(is_public, is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_profiles_created_at ON identity_profiles(created_at DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE identity_profiles IS 'User profiles with display information, personal details, and social links';
COMMENT ON COLUMN identity_profiles.display_name IS 'User display name (public)';
COMMENT ON COLUMN identity_profiles.bio IS 'Short biography (max 500 characters)';
COMMENT ON COLUMN identity_profiles.gender IS 'User gender: male, female, other, not_specified';
COMMENT ON COLUMN identity_profiles.is_public IS 'Profile visibility flag';
COMMENT ON COLUMN identity_profiles.is_active IS 'Profile active status';
