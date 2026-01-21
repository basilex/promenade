-- ============================================================================
-- Accounting Context: Seed Periods, Tax, Groups (Rollback)
-- ============================================================================

DELETE FROM accounting_cost_centers WHERE organization_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM accounting_account_groups WHERE organization_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM accounting_tax_codes WHERE organization_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM accounting_fiscal_periods WHERE organization_id = '00000000-0000-0000-0000-000000000001';
