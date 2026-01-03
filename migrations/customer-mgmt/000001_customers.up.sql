-- Customer Management Context - Customers Table
-- Aggregate: Customer (Lead → Prospect → Customer → Churned lifecycle)

-- Customer Status enum
CREATE TYPE customer_status AS ENUM ('lead', 'prospect', 'customer', 'churned');

-- Customer Tier enum (subscription level)
CREATE TYPE customer_tier AS ENUM ('free', 'basic', 'pro', 'enterprise');

-- Customers table
CREATE TABLE customer_customers (
    -- Identity
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES identity_users(id) ON DELETE SET NULL, -- Linked account (optional)
    company_id UUID, -- B2B company reference (future: REFERENCES customer_companies(id))
    
    -- Basic Info
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50), -- Optional, stored as string (valueobject.Phone in domain)
    
    -- Lifecycle
    status customer_status NOT NULL DEFAULT 'lead',
    tier customer_tier NOT NULL DEFAULT 'free',
    source VARCHAR(100) NOT NULL, -- "website", "referral", "cold_call", etc.
    assigned_to UUID NOT NULL, -- Sales rep (future: REFERENCES identity_users(id))
    
    -- Metadata
    tags TEXT, -- JSON array stored as TEXT for cross-DB compatibility (e.g., ["vip", "high-value"])
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_contacted_at TIMESTAMP WITH TIME ZONE, -- Last interaction timestamp
    converted_at TIMESTAMP WITH TIME ZONE, -- When lead/prospect became customer
    churned_at TIMESTAMP WITH TIME ZONE, -- When customer churned
    churn_reason TEXT, -- Why customer churned
    deleted_at TIMESTAMP WITH TIME ZONE, -- Soft delete
    
    -- Constraints
    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT check_converted_at_after_created CHECK (converted_at IS NULL OR converted_at >= created_at),
    CONSTRAINT check_churned_at_after_converted CHECK (churned_at IS NULL OR converted_at IS NULL OR churned_at >= converted_at)
);

-- Partial unique index for email (only for non-deleted rows)
CREATE UNIQUE INDEX idx_customers_email_unique ON customer_customers(email) WHERE deleted_at IS NULL;

-- Indexes for common queries
CREATE INDEX idx_customers_user_id ON customer_customers(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_company_id ON customer_customers(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_email ON customer_customers(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_status ON customer_customers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_tier ON customer_customers(tier) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_assigned_to ON customer_customers(assigned_to) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_created_at ON customer_customers(created_at DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE customer_customers IS 'Customer aggregate with Lead → Prospect → Customer → Churned lifecycle';
COMMENT ON COLUMN customer_customers.status IS 'Customer lifecycle stage (one-way transitions: lead → prospect → customer → churned)';
COMMENT ON COLUMN customer_customers.tier IS 'Subscription tier (free → basic → pro → enterprise, bidirectional)';
COMMENT ON COLUMN customer_customers.source IS 'Customer acquisition channel (website, referral, cold_call, etc.)';
COMMENT ON COLUMN customer_customers.tags IS 'Flexible JSON array stored as TEXT for customer segmentation (e.g., ["vip", "high-value"])';
COMMENT ON COLUMN customer_customers.converted_at IS 'Timestamp when lead/prospect became paying customer';
COMMENT ON COLUMN customer_customers.churned_at IS 'Timestamp when customer left';
COMMENT ON COLUMN customer_customers.churn_reason IS 'Explanation why customer churned (for analytics)';
