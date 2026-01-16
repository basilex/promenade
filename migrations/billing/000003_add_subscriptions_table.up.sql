-- Migration: add_subscriptions_table
-- Context: billing
-- Created: 2026-01-05 07:50:42

-- Create billing_subscriptions table for recurring billing management
CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id TEXT PRIMARY KEY,
    subscription_no VARCHAR(15) NOT NULL UNIQUE, -- SUB-YYYY-NNNNNN format
    customer_id TEXT NOT NULL, -- Reference to customer (no FK for context isolation)
    plan_id VARCHAR(100) NOT NULL, -- Subscription plan identifier
    status VARCHAR(20) NOT NULL CHECK (status IN ('trial', 'active', 'paused', 'cancelled', 'expired')),
    billing_period VARCHAR(20) NOT NULL CHECK (billing_period IN ('monthly', 'quarterly', 'yearly')),
    
    -- Financial details
    currency VARCHAR(3) NOT NULL, -- ISO 4217
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0), -- Subscription price in cents
    
    -- Dates
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NULL, -- Nullable for active subscriptions
    renewal_date TIMESTAMP NOT NULL,
    trial_end_date TIMESTAMP NULL, -- Nullable if no trial
    
    -- Cancellation
    cancelled_at TIMESTAMP NULL,
    cancel_reason TEXT NULL,
    cancellation_effective_date TIMESTAMP NULL,
    
    -- Metadata (TEXT for database-agnostic JSON storage)
    metadata TEXT DEFAULT '{}', -- Additional subscription metadata stored as JSON
    
    -- Standard audit fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Indexes for common queries
CREATE INDEX idx_billing_subscriptions_customer_id ON billing_subscriptions(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_subscriptions_status ON billing_subscriptions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_subscriptions_renewal_date ON billing_subscriptions(renewal_date) WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX idx_billing_subscriptions_subscription_no ON billing_subscriptions(subscription_no) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON TABLE billing_subscriptions IS 'Recurring billing subscriptions with auto-renewal and lifecycle management';
COMMENT ON COLUMN billing_subscriptions.subscription_no IS 'Unique subscription identifier in SUB-YYYY-NNNNNN format';
COMMENT ON COLUMN billing_subscriptions.customer_id IS 'Reference to customer (no FK for bounded context isolation)';
COMMENT ON COLUMN billing_subscriptions.status IS 'Subscription lifecycle status: trial, active, paused, cancelled, expired';
COMMENT ON COLUMN billing_subscriptions.billing_period IS 'Renewal frequency: monthly, quarterly, yearly';
COMMENT ON COLUMN billing_subscriptions.amount_cents IS 'Subscription price in cents for precise currency handling';
