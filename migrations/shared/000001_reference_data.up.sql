-- ============================================================================
-- Shared Kernel: Reference Data (Countries, Currencies, Languages, Timezones)
-- ============================================================================
-- These tables are part of the Shared Kernel - accessible by all Bounded Contexts.
-- Reference data is stable, standardized, and rarely changes.
--
-- Design decisions:
-- - Read-only from application perspective (managed via migrations)
-- - No soft delete (data doesn't "disappear")
-- - is_active flag for deactivation without breaking foreign keys
-- - UUID v7 primary keys for consistency
-- ============================================================================

-- Countries (ISO 3166-1)
CREATE TABLE IF NOT EXISTS shared_countries (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    code VARCHAR(2) NOT NULL UNIQUE,           -- ISO 3166-1 alpha-2 (US, UA, DE)
    code3 VARCHAR(3) NOT NULL UNIQUE,          -- ISO 3166-1 alpha-3 (USA, UKR, DEU)
    numeric_code VARCHAR(3) NOT NULL UNIQUE,   -- ISO 3166-1 numeric (840, 804, 276)
    name VARCHAR(100) NOT NULL,                -- English name
    name_local VARCHAR(100),                   -- Local name (optional)
    phone_code VARCHAR(10) NOT NULL,           -- Phone country code (+1, +380, +49)
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_countries_code ON shared_countries(code) WHERE is_active = true;
CREATE INDEX idx_shared_countries_name ON shared_countries(name) WHERE is_active = true;

COMMENT ON TABLE shared_countries IS 'ISO 3166-1 countries (Shared Kernel)';
COMMENT ON COLUMN shared_countries.code IS 'ISO 3166-1 alpha-2 code (US, UA, DE)';
COMMENT ON COLUMN shared_countries.phone_code IS 'International dialing code (+1, +380, +49)';

-- Currencies (ISO 4217)
CREATE TABLE IF NOT EXISTS shared_currencies (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    code VARCHAR(3) NOT NULL UNIQUE,           -- ISO 4217 alpha code (USD, EUR, UAH)
    numeric_code VARCHAR(3) NOT NULL UNIQUE,   -- ISO 4217 numeric (840, 978, 980)
    name VARCHAR(100) NOT NULL,                -- English name
    symbol VARCHAR(10) NOT NULL,               -- Currency symbol ($, €, ₴)
    decimal_places SMALLINT NOT NULL DEFAULT 2, -- Decimal places (2 for USD, 0 for JPY)
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_currencies_code ON shared_currencies(code) WHERE is_active = true;

COMMENT ON TABLE shared_currencies IS 'ISO 4217 currencies (Shared Kernel)';
COMMENT ON COLUMN shared_currencies.decimal_places IS 'Number of decimal places (2 for USD, 0 for JPY, etc.)';

-- Languages (ISO 639-1)
CREATE TABLE IF NOT EXISTS shared_languages (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    code VARCHAR(2) NOT NULL UNIQUE,           -- ISO 639-1 alpha-2 (en, uk, de)
    code3 VARCHAR(3) NOT NULL UNIQUE,          -- ISO 639-2/T alpha-3 (eng, ukr, deu)
    name VARCHAR(100) NOT NULL,                -- English name
    native_name VARCHAR(100) NOT NULL,         -- Native name (English, Ukrainian, Deutsch)
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_languages_code ON shared_languages(code) WHERE is_active = true;

COMMENT ON TABLE shared_languages IS 'ISO 639-1 languages (Shared Kernel)';

-- Timezones (IANA)
CREATE TABLE IF NOT EXISTS shared_timezones (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) NOT NULL UNIQUE,         -- IANA timezone (America/New_York, Europe/Kyiv)
    abbreviation VARCHAR(10) NOT NULL,         -- Common abbreviation (EST, EET, CET)
    utc_offset INTEGER NOT NULL,               -- UTC offset in seconds
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_timezones_name ON shared_timezones(name) WHERE is_active = true;

COMMENT ON TABLE shared_timezones IS 'IANA timezones (Shared Kernel)';
COMMENT ON COLUMN shared_timezones.utc_offset IS 'UTC offset in seconds';

-- ============================================================================
-- Seed data: Most commonly used reference data
-- ============================================================================

-- Countries (Top 20 most common)
INSERT INTO shared_countries (code, code3, numeric_code, name, name_local, phone_code) VALUES
    ('US', 'USA', '840', 'United States', 'United States', '+1'),
    ('UA', 'UKR', '804', 'Ukraine', 'Україна', '+380'),
    ('DE', 'DEU', '276', 'Germany', 'Deutschland', '+49'),
    ('GB', 'GBR', '826', 'United Kingdom', 'United Kingdom', '+44'),
    ('FR', 'FRA', '250', 'France', 'France', '+33'),
    ('PL', 'POL', '616', 'Poland', 'Polska', '+48'),
    ('IT', 'ITA', '380', 'Italy', 'Italia', '+39'),
    ('ES', 'ESP', '724', 'Spain', 'España', '+34'),
    ('CA', 'CAN', '124', 'Canada', 'Canada', '+1'),
    ('AU', 'AUS', '036', 'Australia', 'Australia', '+61'),
    ('JP', 'JPN', '392', 'Japan', '日本', '+81'),
    ('CN', 'CHN', '156', 'China', '中国', '+86'),
    ('IN', 'IND', '356', 'India', 'भारत', '+91'),
    ('BR', 'BRA', '076', 'Brazil', 'Brasil', '+55'),
    ('MX', 'MEX', '484', 'Mexico', 'México', '+52'),
    ('RU', 'RUS', '643', 'Russia', 'Россия', '+7'),
    ('TR', 'TUR', '792', 'Turkey', 'Türkiye', '+90'),
    ('NL', 'NLD', '528', 'Netherlands', 'Nederland', '+31'),
    ('SE', 'SWE', '752', 'Sweden', 'Sverige', '+46'),
    ('CH', 'CHE', '756', 'Switzerland', 'Schweiz', '+41');

-- Currencies (Top 15 most common)
INSERT INTO shared_currencies (code, numeric_code, name, symbol, decimal_places) VALUES
    ('USD', '840', 'US Dollar', '$', 2),
    ('EUR', '978', 'Euro', '€', 2),
    ('UAH', '980', 'Ukrainian Hryvnia', '₴', 2),
    ('GBP', '826', 'British Pound', '£', 2),
    ('JPY', '392', 'Japanese Yen', '¥', 0),
    ('CHF', '756', 'Swiss Franc', 'CHF', 2),
    ('CAD', '124', 'Canadian Dollar', 'C$', 2),
    ('AUD', '036', 'Australian Dollar', 'A$', 2),
    ('CNY', '156', 'Chinese Yuan', '¥', 2),
    ('INR', '356', 'Indian Rupee', '₹', 2),
    ('BRL', '986', 'Brazilian Real', 'R$', 2),
    ('MXN', '484', 'Mexican Peso', 'MX$', 2),
    ('RUB', '643', 'Russian Ruble', '₽', 2),
    ('PLN', '985', 'Polish Zloty', 'zł', 2),
    ('SEK', '752', 'Swedish Krona', 'kr', 2);

-- Languages (Top 15 most common)
INSERT INTO shared_languages (code, code3, name, native_name) VALUES
    ('en', 'eng', 'English', 'English'),
    ('uk', 'ukr', 'Ukrainian', 'Українська'),
    ('de', 'deu', 'German', 'Deutsch'),
    ('fr', 'fra', 'French', 'Français'),
    ('es', 'spa', 'Spanish', 'Español'),
    ('it', 'ita', 'Italian', 'Italiano'),
    ('pl', 'pol', 'Polish', 'Polski'),
    ('ru', 'rus', 'Russian', 'Русский'),
    ('pt', 'por', 'Portuguese', 'Português'),
    ('ja', 'jpn', 'Japanese', '日本語'),
    ('zh', 'zho', 'Chinese', '中文'),
    ('ar', 'ara', 'Arabic', 'العربية'),
    ('hi', 'hin', 'Hindi', 'हिन्दी'),
    ('tr', 'tur', 'Turkish', 'Türkçe'),
    ('nl', 'nld', 'Dutch', 'Nederlands');

-- Timezones (Top 20 most common)
INSERT INTO shared_timezones (name, abbreviation, utc_offset) VALUES
    ('America/New_York', 'EST', -18000),       -- UTC-5
    ('America/Chicago', 'CST', -21600),        -- UTC-6
    ('America/Denver', 'MST', -25200),         -- UTC-7
    ('America/Los_Angeles', 'PST', -28800),    -- UTC-8
    ('Europe/London', 'GMT', 0),               -- UTC+0
    ('Europe/Paris', 'CET', 3600),             -- UTC+1
    ('Europe/Berlin', 'CET', 3600),            -- UTC+1
    ('Europe/Kyiv', 'EET', 7200),              -- UTC+2
    ('Europe/Moscow', 'MSK', 10800),           -- UTC+3
    ('Asia/Dubai', 'GST', 14400),              -- UTC+4
    ('Asia/Kolkata', 'IST', 19800),            -- UTC+5:30
    ('Asia/Shanghai', 'CST', 28800),           -- UTC+8
    ('Asia/Tokyo', 'JST', 32400),              -- UTC+9
    ('Australia/Sydney', 'AEDT', 39600),       -- UTC+11
    ('Pacific/Auckland', 'NZDT', 46800),       -- UTC+13
    ('America/Toronto', 'EST', -18000),        -- UTC-5
    ('America/Mexico_City', 'CST', -21600),    -- UTC-6
    ('America/Sao_Paulo', 'BRT', -10800),      -- UTC-3
    ('Africa/Cairo', 'EET', 7200),             -- UTC+2
    ('Asia/Singapore', 'SGT', 28800);          -- UTC+8
