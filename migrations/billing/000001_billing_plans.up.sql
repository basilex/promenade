-- Migration: Create billing_plans table
-- Description: Store subscription plan definitions (Free, Basic, Pro, Enterprise)

CREATE TABLE IF NOT EXISTS billing_plans (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- Plan identification
    slug VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Plan status
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'archived')),
    
    -- Pricing
    interval VARCHAR(50) NOT NULL CHECK (interval IN ('month', 'year')),
    amount BIGINT NOT NULL CHECK (amount >= 0), -- Amount in cents/kopecks
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Trial period
    trial_days INTEGER NOT NULL DEFAULT 0 CHECK (trial_days >= 0),
    
    -- Features and limits
    features JSONB NOT NULL DEFAULT '[]'::jsonb, -- Array of feature names
    max_users INTEGER CHECK (max_users > 0),
    max_projects INTEGER CHECK (max_projects > 0),
    max_storage BIGINT CHECK (max_storage > 0), -- Storage in bytes
    
    -- Unlimited flags
    is_unlimited_users BOOLEAN NOT NULL DEFAULT false,
    is_unlimited_projects BOOLEAN NOT NULL DEFAULT false,
    is_unlimited_storage BOOLEAN NOT NULL DEFAULT false,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP -- Soft delete
);

-- Indexes for performance
CREATE INDEX idx_billing_plans_slug ON billing_plans(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_plans_status ON billing_plans(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_plans_interval ON billing_plans(interval) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_plans_deleted_at ON billing_plans(deleted_at);

-- Comments for documentation
COMMENT ON TABLE billing_plans IS 'Subscription plan definitions with features and pricing';
COMMENT ON COLUMN billing_plans.slug IS 'URL-friendly unique identifier for the plan';
COMMENT ON COLUMN billing_plans.amount IS 'Price in smallest currency unit (cents for USD, kopecks for UAH)';
COMMENT ON COLUMN billing_plans.features IS 'JSON array of feature names included in this plan';
COMMENT ON COLUMN billing_plans.max_storage IS 'Maximum storage in bytes (NULL = no limit)';
