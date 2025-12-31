-- ============================================================================
-- Migration: 000003_deals
-- Description: Create customer_deals table for sales pipeline management
-- Context: Customer Management
-- Created: 2025-12-30
-- ============================================================================

-- ============================================================================
-- Deal Stage Enum
-- ============================================================================
-- Represents sales pipeline stages
CREATE TYPE deal_stage AS ENUM (
    'lead',         -- Initial contact
    'qualified',    -- Qualified opportunity
    'proposal',     -- Proposal sent
    'negotiation',  -- In negotiation
    'closed_won',   -- Deal won
    'closed_lost'   -- Deal lost
);

COMMENT ON TYPE deal_stage IS 'Sales pipeline stages for deal progression';

-- ============================================================================
-- Deal Source Enum
-- ============================================================================
-- Represents how the deal originated
CREATE TYPE deal_source AS ENUM (
    'inbound',     -- Website, content, SEO
    'outbound',    -- Cold outreach
    'referral',    -- Customer referral
    'partner',     -- Partner channel
    'event',       -- Conference, webinar
    'advertising'  -- Paid ads
);

COMMENT ON TYPE deal_source IS 'Deal origin channels for lead attribution';

-- ============================================================================
-- Customer Deals Table
-- ============================================================================
-- Stores sales opportunities and pipeline management
CREATE TABLE customer_deals (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT uuid_v7(),

    -- Relationships
    customer_id UUID NOT NULL,                  -- Link to customer (required)
    company_id UUID,                            -- Link to company (optional, for B2B)

    -- Basic Information
    name VARCHAR(255) NOT NULL,                 -- Deal name/title
    description TEXT,                           -- Deal description

    -- Financial Information
    value_cents BIGINT NOT NULL DEFAULT 0,      -- Deal value in cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD', -- ISO 4217 currency code

    -- Pipeline Information
    stage deal_stage NOT NULL DEFAULT 'lead',   -- Current pipeline stage
    probability INTEGER NOT NULL DEFAULT 10,    -- Win probability (0-100%)
    source deal_source NOT NULL DEFAULT 'inbound', -- Deal origin

    -- Dates
    expected_close_date DATE NOT NULL,          -- Expected close date
    actual_close_date DATE,                     -- Actual close date (when closed)

    -- Ownership
    assigned_to UUID NOT NULL,                  -- Sales rep (Identity.User)

    -- Closure Information
    close_reason TEXT,                          -- Win/loss reason

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,                     -- Soft delete

    -- Constraints
    CONSTRAINT deals_name_not_empty CHECK (name <> ''),
    CONSTRAINT deals_value_positive CHECK (value_cents >= 0),
    CONSTRAINT deals_currency_valid CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT deals_probability_range CHECK (probability >= 0 AND probability <= 100),
    CONSTRAINT deals_expected_close_date_valid CHECK (expected_close_date >= CURRENT_DATE - INTERVAL '1 year'),

    -- Foreign Keys
    CONSTRAINT fk_deals_customer
        FOREIGN KEY (customer_id)
        REFERENCES customer_customers(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_deals_company
        FOREIGN KEY (company_id)
        REFERENCES customer_companies(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_deals_assigned_to
        FOREIGN KEY (assigned_to)
        REFERENCES identity_users(id)
        ON DELETE RESTRICT
);

-- ============================================================================
-- Table Comments
-- ============================================================================
COMMENT ON TABLE customer_deals IS 'Sales opportunities and pipeline management';
COMMENT ON COLUMN customer_deals.id IS 'Unique deal identifier (UUID v7)';
COMMENT ON COLUMN customer_deals.customer_id IS 'Customer this deal belongs to (required)';
COMMENT ON COLUMN customer_deals.company_id IS 'Company associated with deal (optional, B2B)';
COMMENT ON COLUMN customer_deals.name IS 'Deal name/title for identification';
COMMENT ON COLUMN customer_deals.description IS 'Detailed description of opportunity';
COMMENT ON COLUMN customer_deals.value_cents IS 'Deal value in cents (for precision)';
COMMENT ON COLUMN customer_deals.currency IS 'Currency code (ISO 4217: USD, EUR, etc.)';
COMMENT ON COLUMN customer_deals.stage IS 'Current pipeline stage';
COMMENT ON COLUMN customer_deals.probability IS 'Win probability percentage (0-100)';
COMMENT ON COLUMN customer_deals.source IS 'How the deal originated';
COMMENT ON COLUMN customer_deals.expected_close_date IS 'When deal is expected to close';
COMMENT ON COLUMN customer_deals.actual_close_date IS 'When deal actually closed (won/lost)';
COMMENT ON COLUMN customer_deals.assigned_to IS 'Sales rep responsible for deal';
COMMENT ON COLUMN customer_deals.close_reason IS 'Reason for win or loss';
COMMENT ON COLUMN customer_deals.created_at IS 'Timestamp when deal was created';
COMMENT ON COLUMN customer_deals.updated_at IS 'Timestamp of last update';
COMMENT ON COLUMN customer_deals.deleted_at IS 'Soft delete timestamp (NULL = active)';

-- ============================================================================
-- Indexes for Performance
-- ============================================================================

-- Primary lookup indexes
CREATE INDEX idx_deals_customer_id ON customer_deals(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_company_id ON customer_deals(company_id) WHERE deleted_at IS NULL AND company_id IS NOT NULL;
CREATE INDEX idx_deals_assigned_to ON customer_deals(assigned_to) WHERE deleted_at IS NULL;

-- Pipeline queries
CREATE INDEX idx_deals_stage ON customer_deals(stage) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_stage_assigned_to ON customer_deals(stage, assigned_to) WHERE deleted_at IS NULL;

-- Date-based queries
CREATE INDEX idx_deals_expected_close_date ON customer_deals(expected_close_date) WHERE deleted_at IS NULL AND stage NOT IN ('closed_won', 'closed_lost');
CREATE INDEX idx_deals_actual_close_date ON customer_deals(actual_close_date) WHERE deleted_at IS NULL AND actual_close_date IS NOT NULL;

-- Analytics queries
CREATE INDEX idx_deals_stage_created_at ON customer_deals(stage, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_source ON customer_deals(source) WHERE deleted_at IS NULL;

-- Composite indexes for common queries
CREATE INDEX idx_deals_customer_stage ON customer_deals(customer_id, stage) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_assigned_stage_value ON customer_deals(assigned_to, stage, value_cents DESC) WHERE deleted_at IS NULL;

-- Soft delete index
CREATE INDEX idx_deals_deleted_at ON customer_deals(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================================
-- Index Comments
-- ============================================================================
COMMENT ON INDEX idx_deals_customer_id IS 'Fast lookup of deals by customer';
COMMENT ON INDEX idx_deals_company_id IS 'Fast lookup of deals by company (B2B)';
COMMENT ON INDEX idx_deals_assigned_to IS 'Fast lookup of deals by sales rep';
COMMENT ON INDEX idx_deals_stage IS 'Pipeline stage filtering';
COMMENT ON INDEX idx_deals_stage_assigned_to IS 'Sales rep pipeline view';
COMMENT ON INDEX idx_deals_expected_close_date IS 'Deals closing soon (active only)';
COMMENT ON INDEX idx_deals_actual_close_date IS 'Historical close date analysis';
COMMENT ON INDEX idx_deals_stage_created_at IS 'Pipeline chronological order';
COMMENT ON INDEX idx_deals_source IS 'Lead attribution analysis';
COMMENT ON INDEX idx_deals_customer_stage IS 'Customer deal pipeline';
COMMENT ON INDEX idx_deals_assigned_stage_value IS 'Sales rep weighted pipeline';
COMMENT ON INDEX idx_deals_deleted_at IS 'Soft delete queries';

-- ============================================================================
-- Triggers
-- ============================================================================

-- Update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_deals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_deals_updated_at
    BEFORE UPDATE ON customer_deals
    FOR EACH ROW
    EXECUTE FUNCTION update_deals_updated_at();

COMMENT ON FUNCTION update_deals_updated_at() IS 'Automatically updates updated_at timestamp on row modification';
COMMENT ON TRIGGER trigger_deals_updated_at ON customer_deals IS 'Ensures updated_at is always current';

-- ============================================================================
-- Grants (Optional - adjust based on your roles)
-- ============================================================================
-- GRANT SELECT, INSERT, UPDATE, DELETE ON customer_deals TO app_user;
-- GRANT USAGE ON TYPE deal_stage TO app_user;
-- GRANT USAGE ON TYPE deal_source TO app_user;

-- ============================================================================
-- Sample Data (Optional - for development)
-- ============================================================================
-- INSERT INTO customer_deals (customer_id, name, value_cents, currency, stage, probability, assigned_to, expected_close_date)
-- VALUES (
--     '01H8...',  -- Replace with actual customer_id
--     'Enterprise License Deal',
--     50000000,   -- $500,000.00
--     'USD',
--     'proposal',
--     50,
--     '01H8...',  -- Replace with actual user_id
--     CURRENT_DATE + INTERVAL '30 days'
-- );
