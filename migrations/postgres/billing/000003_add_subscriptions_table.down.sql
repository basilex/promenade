-- Migration: add_subscriptions_table
-- Context: billing
-- Created: 2026-01-05 07:50:42

-- Drop billing_subscriptions table
DROP TABLE IF EXISTS billing_subscriptions CASCADE;

