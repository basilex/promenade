-- Create profiles_contacts table for storing multiple contact methods per user
CREATE TABLE IF NOT EXISTS profiles_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    
    -- Contact information
    contact_type VARCHAR(50) NOT NULL CHECK (contact_type IN (
        'email', 'phone', 'telegram', 'whatsapp', 'viber', 
        'signal', 'skype', 'discord', 'linkedin', 'other'
    )),
    contact_value VARCHAR(255) NOT NULL,
    label VARCHAR(100), -- e.g., "Work", "Personal", "Emergency"
    
    -- Status and privacy
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE, -- One primary contact per type
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_public BOOLEAN NOT NULL DEFAULT FALSE, -- Visible in public profile
    
    -- Availability schedule (optional)
    available_from TIME, -- Start time (e.g., 09:00)
    available_to TIME,   -- End time (e.g., 18:00)
    available_days VARCHAR(20)[], -- e.g., ['monday', 'tuesday', 'friday']
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Additional metadata
    notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique contact values per user and type
    UNIQUE (user_id, contact_type, contact_value)
);

-- Indexes for performance
CREATE INDEX idx_profiles_contacts_user_id ON profiles_contacts(user_id);
CREATE INDEX idx_profiles_contacts_type ON profiles_contacts(contact_type);
CREATE INDEX idx_profiles_contacts_active ON profiles_contacts(is_active) WHERE is_active = true;
CREATE INDEX idx_profiles_contacts_public ON profiles_contacts(is_public) WHERE is_public = true;
CREATE INDEX idx_profiles_contacts_verified ON profiles_contacts(is_verified) WHERE is_verified = true;

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_profiles_contacts_updated_at
    BEFORE UPDATE ON profiles_contacts
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- Comments for documentation
COMMENT ON TABLE profiles_contacts IS 'Stores multiple contact methods for users with privacy and availability settings';
COMMENT ON COLUMN profiles_contacts.contact_type IS 'Type of contact: email, phone, telegram, whatsapp, viber, signal, skype, discord, linkedin, other';
COMMENT ON COLUMN profiles_contacts.is_primary IS 'Indicates if this is the primary contact of its type';
COMMENT ON COLUMN profiles_contacts.is_public IS 'Whether this contact is visible in public profile/directory';
COMMENT ON COLUMN profiles_contacts.available_from IS 'Start time of availability window (e.g., 09:00)';
COMMENT ON COLUMN profiles_contacts.available_to IS 'End time of availability window (e.g., 18:00)';
COMMENT ON COLUMN profiles_contacts.available_days IS 'Days of week when contact is available';
