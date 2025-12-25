-- Migration: Create billing_subscriptions table
-- Description: Store user subscriptions to plans with lifecycle management

CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- References
    user_id UUID NOT NULL,
    plan_id UUID NOT NULL,
    
    -- Subscription status
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('trialing', 'active', 'past_due', 'canceled', 'expired')),
    
    -- Billing period
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    
    -- Trial period
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    
    -- Cancellation
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    canceled_at TIMESTAMP,
    
    -- Flexible metadata storage
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP, -- Soft delete
    
    -- Foreign keys
    CONSTRAINT fk_billing_subscriptions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_billing_subscriptions_plan FOREIGN KEY (plan_id) REFERENCES billing_plans(id) ON DELETE RESTRICT,
    
    -- Business constraints
    CONSTRAINT chk_billing_subscriptions_period CHECK (current_period_end > current_period_start),
    CONSTRAINT chk_billing_subscriptions_trial CHECK (
        (trial_start IS NULL AND trial_end IS NULL) OR 
        (trial_start IS NOT NULL AND trial_end IS NOT NULL AND trial_end > trial_start)
    )
);

-- Indexes for performance
CREATE INDEX idx_billing_subscriptions_user_id ON billing_subscriptions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_subscriptions_plan_id ON billing_subscriptions(plan_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_subscriptions_status ON billing_subscriptions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_subscriptions_period_end ON billing_subscriptions(current_period_end) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_subscriptions_deleted_at ON billing_subscriptions(deleted_at);

-- Composite index for finding active user subscriptions
CREATE INDEX idx_billing_subscriptions_user_status ON billing_subscriptions(user_id, status) WHERE deleted_at IS NULL;

-- Comments for documentation
COMMENT ON TABLE billing_subscriptions IS 'User subscriptions to billing plans with lifecycle tracking';
COMMENT ON COLUMN billing_subscriptions.cancel_at_period_end IS 'If true, subscription will be canceled at the end of current period';
COMMENT ON COLUMN billing_subscriptions.metadata IS 'Flexible JSON storage for additional subscription data';
