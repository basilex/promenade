-- ============================================================================
-- Accounting Context: Seed Chart of Accounts (Rollback)
-- ============================================================================
-- Remove all seeded accounts in reverse order
-- ============================================================================

-- Delete all seeded accounts for default organization
DELETE FROM accounting_chart_of_accounts 
WHERE organization_id = '00000000-0000-0000-0000-000000000001'
AND code IN (
    -- Class 1-3: Assets
    '10', '103', '104',
    '20', '201', '28', '281',
    '30', '301', '31', '311', '36', '361',
    -- Class 4: Equity
    '40', '401', '44', '441',
    -- Class 5-6: Liabilities
    '50', '501', '60', '601', '63', '631', '64', '641', '66', '661',
    -- Class 7: Income
    '70', '701', '702', '703', '71', '712',
    -- Class 8-9: Expenses
    '90', '901', '902', '92', '93', '94', '95', '951'
);
