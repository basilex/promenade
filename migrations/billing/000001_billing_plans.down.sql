-- Migration: Drop billing_plans table
-- Description: Rollback plan definitions table

DROP INDEX IF EXISTS idx_billing_plans_deleted_at;
DROP INDEX IF EXISTS idx_billing_plans_interval;
DROP INDEX IF EXISTS idx_billing_plans_status;
DROP INDEX IF EXISTS idx_billing_plans_slug;

DROP TABLE IF EXISTS billing_plans;
