-- Migration: Drop billing_subscriptions table
-- Description: Rollback user subscriptions table

DROP INDEX IF EXISTS idx_billing_subscriptions_user_status;
DROP INDEX IF EXISTS idx_billing_subscriptions_deleted_at;
DROP INDEX IF EXISTS idx_billing_subscriptions_period_end;
DROP INDEX IF EXISTS idx_billing_subscriptions_status;
DROP INDEX IF EXISTS idx_billing_subscriptions_plan_id;
DROP INDEX IF EXISTS idx_billing_subscriptions_user_id;

DROP TABLE IF EXISTS billing_subscriptions;
