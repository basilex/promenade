-- ============================================================================
-- Accounting Context: Seed Ukrainian Chart of Accounts (План рахунків)
-- ============================================================================
-- Based on П(С)БО (Ukrainian Accounting Standards)
-- Simplified version with essential accounts for SME
-- ============================================================================

-- System user for seeds
DO $$ 
DECLARE
    system_user_id TEXT := '00000000-0000-0000-0000-000000000000';
    system_org_id TEXT := '00000000-0000-0000-0000-000000000001';
BEGIN
    -- Delete existing seed data to allow re-running
    DELETE FROM accounting_chart_of_accounts WHERE organization_id = system_org_id;

    -- 1. ASSETS (Активи) - Class 1-3
    -- Class 1: Основні засоби
    INSERT INTO accounting_chart_of_accounts 
    (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, description, created_by, last_updated_by)
    VALUES
    ('01000000-0000-0000-0000-000000000010', system_org_id, '10', 'Основні засоби', 'asset', NULL, 1, 'UAH', TRUE, 'Матеріальні активи тривалого використання', system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000103', system_org_id, '103', 'Будівлі та споруди', 'asset', '01000000-0000-0000-0000-000000000010', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000104', system_org_id, '104', 'Машини та обладнання', 'asset', '01000000-0000-0000-0000-000000000010', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id);

    -- Class 2: Запаси
    INSERT INTO accounting_chart_of_accounts 
    (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, description, created_by, last_updated_by)
    VALUES
    ('01000000-0000-0000-0000-000000000020', system_org_id, '20', 'Виробничі запаси', 'asset', NULL, 1, 'UAH', TRUE, 'Сировина, матеріали, паливо', system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000201', system_org_id, '201', 'Сировина і матеріали', 'asset', '01000000-0000-0000-0000-000000000020', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000028', system_org_id, '28', 'Товари', 'asset', NULL, 1, 'UAH', TRUE, 'Товари для перепродажу', system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000281', system_org_id, '281', 'Товари на складі', 'asset', '01000000-0000-0000-0000-000000000028', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id);

    -- Class 3: Грошові кошти
    INSERT INTO accounting_chart_of_accounts 
    (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, description, created_by, last_updated_by)
    VALUES
    ('01000000-0000-0000-0000-000000000030', system_org_id, '30', 'Каса', 'asset', NULL, 1, 'UAH', TRUE, 'Готівка в касі', system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000301', system_org_id, '301', 'Каса в національній валюті', 'asset', '01000000-0000-0000-0000-000000000030', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000031', system_org_id, '31', 'Рахунки в банках', 'asset', NULL, 1, 'UAH', TRUE, 'Безготівкові кошти', system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000311', system_org_id, '311', 'Розрахункові рахунки', 'asset', '01000000-0000-0000-0000-000000000031', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000036', system_org_id, '36', 'Розрахунки з покупцями', 'asset', NULL, 1, 'UAH', TRUE, 'Дебіторська заборгованість', system_user_id, system_user_id),
    ('01000000-0000-0000-0000-000000000361', system_org_id, '361', 'Розрахунки з вітчизняними покупцями', 'asset', '01000000-0000-0000-0000-000000000036', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id);

    -- 2. EQUITY (Власний капітал) - Class 4
    INSERT INTO accounting_chart_of_accounts 
    (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, description, created_by, last_updated_by)
    VALUES
    ('04000000-0000-0000-0000-000000000040', system_org_id, '40', 'Статутний капітал', 'equity', NULL, 1, 'UAH', TRUE, 'Зареєстрований капітал', system_user_id, system_user_id),
    ('04000000-0000-0000-0000-000000000401', system_org_id, '401', 'Статутний капітал', 'equity', '04000000-0000-0000-0000-000000000040', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('04000000-0000-0000-0000-000000000044', system_org_id, '44', 'Нерозподілені прибутки', 'equity', NULL, 1, 'UAH', TRUE, 'Фінансовий результат', system_user_id, system_user_id),
    ('04000000-0000-0000-0000-000000000441', system_org_id, '441', 'Прибуток нерозподілений', 'equity', '04000000-0000-0000-0000-000000000044', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id);

    -- 3. LIABILITIES (Зобов'язання) - Class 5-6
    INSERT INTO accounting_chart_of_accounts 
    (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, description, created_by, last_updated_by)
    VALUES
    ('05000000-0000-0000-0000-000000000050', system_org_id, '50', 'Довгострокові позики', 'liability', NULL, 1, 'UAH', TRUE, 'Кредити банків', system_user_id, system_user_id),
    ('05000000-0000-0000-0000-000000000501', system_org_id, '501', 'Довгострокові кредити', 'liability', '05000000-0000-0000-0000-000000000050', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('06000000-0000-0000-0000-000000000063', system_org_id, '63', 'Розрахунки з постачальниками', 'liability', NULL, 1, 'UAH', TRUE, 'Кредиторська заборгованість', system_user_id, system_user_id),
    ('06000000-0000-0000-0000-000000000631', system_org_id, '631', 'Розрахунки з вітчизняними постачальниками', 'liability', '06000000-0000-0000-0000-000000000063', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('06000000-0000-0000-0000-000000000064', system_org_id, '64', 'Розрахунки з податків', 'liability', NULL, 1, 'UAH', TRUE, 'Податки і збори', system_user_id, system_user_id),
    ('06000000-0000-0000-0000-000000000641', system_org_id, '641', 'Податок на прибуток', 'liability', '06000000-0000-0000-0000-000000000064', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id);

    -- 4. INCOME (Доходи) - Class 7
    INSERT INTO accounting_chart_of_accounts 
    (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, description, created_by, last_updated_by)
    VALUES
    ('07000000-0000-0000-0000-000000000070', system_org_id, '70', 'Дохід від реалізації', 'income', NULL, 1, 'UAH', TRUE, 'Виручка від продажу', system_user_id, system_user_id),
    ('07000000-0000-0000-0000-000000000701', system_org_id, '701', 'Дохід від реалізації товарів', 'income', '07000000-0000-0000-0000-000000000070', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('07000000-0000-0000-0000-000000000702', system_org_id, '702', 'Дохід від реалізації послуг', 'income', '07000000-0000-0000-0000-000000000070', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('07000000-0000-0000-0000-000000000071', system_org_id, '71', 'Інший операційний дохід', 'income', NULL, 1, 'UAH', TRUE, 'Додатковий дохід', system_user_id, system_user_id);

    -- 5. EXPENSES (Витрати) - Class 9
    INSERT INTO accounting_chart_of_accounts 
    (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, description, created_by, last_updated_by)
    VALUES
    ('09000000-0000-0000-0000-000000000090', system_org_id, '90', 'Собівартість реалізації', 'expense', NULL, 1, 'UAH', TRUE, 'Прямі витрати', system_user_id, system_user_id),
    ('09000000-0000-0000-0000-000000000901', system_org_id, '901', 'Собівартість реалізованих товарів', 'expense', '09000000-0000-0000-0000-000000000090', 2, 'UAH', TRUE, NULL, system_user_id, system_user_id),
    ('09000000-0000-0000-0000-000000000092', system_org_id, '92', 'Адміністративні витрати', 'expense', NULL, 1, 'UAH', TRUE, 'Загальні витрати', system_user_id, system_user_id),
    ('09000000-0000-0000-0000-000000000093', system_org_id, '93', 'Витрати на збут', 'expense', NULL, 1, 'UAH', TRUE, 'Витрати на продаж', system_user_id, system_user_id),
    ('09000000-0000-0000-0000-000000000094', system_org_id, '94', 'Інші операційні витрати', 'expense', NULL, 1, 'UAH', TRUE, 'Додаткові витрати', system_user_id, system_user_id);

END $$;
