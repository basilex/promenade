-- Create identity_contacts table for user contact information (DDD aggregate)
-- This table stores email, phone, and address contacts using value objects pattern
CREATE TABLE IF NOT EXISTS identity_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    
    -- Contact type discriminator
    contact_type VARCHAR(20) NOT NULL CHECK (contact_type IN ('email', 'phone', 'address')),
    
    -- Label for the contact (e.g., "Work", "Personal", "Home")
    label VARCHAR(100) NOT NULL,
    
    -- Email value object (only populated when contact_type = 'email')
    email VARCHAR(255),
    
    -- Phone value object (only populated when contact_type = 'phone')
    phone VARCHAR(50),
    
    -- Address value object fields (only populated when contact_type = 'address')
    address_street VARCHAR(255),
    address_street2 VARCHAR(255),
    address_city VARCHAR(100),
    address_state VARCHAR(100),
    address_postal_code VARCHAR(20),
    address_country CHAR(2), -- ISO 3166-1 alpha-2 code (e.g., 'US', 'UA', 'DE')
    
    -- Business flags
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Check constraints to ensure only one value object is populated
    CONSTRAINT contact_email_check CHECK (
        (contact_type = 'email' AND email IS NOT NULL AND phone IS NULL AND address_street IS NULL)
        OR (contact_type != 'email')
    ),
    CONSTRAINT contact_phone_check CHECK (
        (contact_type = 'phone' AND phone IS NOT NULL AND email IS NULL AND address_street IS NULL)
        OR (contact_type != 'phone')
    ),
    CONSTRAINT contact_address_check CHECK (
        (contact_type = 'address' AND address_street IS NOT NULL AND email IS NULL AND phone IS NULL)
        OR (contact_type != 'address')
    )
);

-- Indexes for performance
CREATE INDEX idx_identity_contacts_user_id ON identity_contacts(user_id);
CREATE INDEX idx_identity_contacts_type ON identity_contacts(contact_type);
CREATE INDEX idx_identity_contacts_user_type ON identity_contacts(user_id, contact_type);

-- Partial unique index: only one primary contact per user per type
-- This ensures business rule: user can have only one primary email, one primary phone, one primary address
CREATE UNIQUE INDEX idx_identity_contacts_user_type_primary 
    ON identity_contacts(user_id, contact_type) 
    WHERE is_primary = true;

-- Comment on table
COMMENT ON TABLE identity_contacts IS 'Contact information for users using DDD value objects pattern (Email, Phone, Address)';
COMMENT ON COLUMN identity_contacts.contact_type IS 'Discriminator for contact type: email, phone, or address';
COMMENT ON COLUMN identity_contacts.is_primary IS 'Only one primary contact per user per type (enforced by partial unique index)';
COMMENT ON COLUMN identity_contacts.is_verified IS 'Whether contact has been verified (e.g., email confirmation, phone OTP)';
COMMENT ON COLUMN identity_contacts.is_public IS 'Whether contact is visible in public user profile';
COMMENT ON COLUMN identity_contacts.address_country IS 'ISO 3166-1 alpha-2 country code (2 letters)';
