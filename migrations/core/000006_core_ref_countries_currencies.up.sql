-- Create region enum for geographical grouping
CREATE TYPE country_region AS ENUM (
    'north_america',
    'south_america',
    'western_europe',
    'eastern_europe',
    'asia',
    'middle_east',
    'africa',
    'oceania'
);

COMMENT ON TYPE country_region IS 'Geographical regions for country grouping';

-- Create countries table
CREATE TABLE IF NOT EXISTS core_countries (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(10) NOT NULL,
    iso2 CHAR(2) NOT NULL UNIQUE,
    iso3 CHAR(3) NOT NULL UNIQUE,
    region country_region NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Create currencies table
CREATE TABLE IF NOT EXISTS core_currencies (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(10) NOT NULL UNIQUE,
    symbol VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Create many-to-many junction table
CREATE TABLE IF NOT EXISTS core_country_currencies (
    country_id UUID NOT NULL REFERENCES core_countries(id) ON DELETE CASCADE,
    currency_id UUID NOT NULL REFERENCES core_currencies(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (country_id, currency_id)
);

-- Indexes for better query performance
CREATE INDEX idx_core_countries_code ON core_countries(code);
CREATE INDEX idx_core_countries_iso2 ON core_countries(iso2);
CREATE INDEX idx_core_countries_iso3 ON core_countries(iso3);
CREATE INDEX idx_core_countries_region ON core_countries(region);
CREATE INDEX idx_core_currencies_code ON core_currencies(code);
CREATE INDEX idx_country_core_currencies_country_id ON core_country_currencies(country_id);
CREATE INDEX idx_country_core_currencies_currency_id ON core_country_currencies(currency_id);

-- Insert real currency data
INSERT INTO core_currencies (name, code, symbol) VALUES
    ('US Dollar', 'USD', '$'),
    ('Euro', 'EUR', '€'),
    ('British Pound', 'GBP', '£'),
    ('Japanese Yen', 'JPY', '¥'),
    ('Swiss Franc', 'CHF', 'Fr'),
    ('Canadian Dollar', 'CAD', 'C$'),
    ('Australian Dollar', 'AUD', 'A$'),
    ('Chinese Yuan', 'CNY', '¥'),
    ('Russian Ruble', 'RUB', '₽'),
    ('Indian Rupee', 'INR', '₹'),
    ('Brazilian Real', 'BRL', 'R$'),
    ('South African Rand', 'ZAR', 'R'),
    ('Mexican Peso', 'MXN', '$'),
    ('Singapore Dollar', 'SGD', 'S$'),
    ('Hong Kong Dollar', 'HKD', 'HK$'),
    ('Swedish Krona', 'SEK', 'kr'),
    ('Norwegian Krone', 'NOK', 'kr'),
    ('Danish Krone', 'DKK', 'kr'),
    ('Polish Zloty', 'PLN', 'zł'),
    ('Turkish Lira', 'TRY', '₺'),
    ('South Korean Won', 'KRW', '₩'),
    ('New Zealand Dollar', 'NZD', 'NZ$'),
    ('Thai Baht', 'THB', '฿'),
    ('Malaysian Ringgit', 'MYR', 'RM'),
    ('Indonesian Rupiah', 'IDR', 'Rp'),
    ('Philippine Peso', 'PHP', '₱'),
    ('Czech Koruna', 'CZK', 'Kč'),
    ('Israeli Shekel', 'ILS', '₪'),
    ('Chilean Peso', 'CLP', '$'),
    ('Colombian Peso', 'COP', '$'),
    ('Argentine Peso', 'ARS', '$'),
    ('Peruvian Sol', 'PEN', 'S/'),
    ('Uruguayan Peso', 'UYU', '$U'),
    ('Venezuelan Bolívar', 'VES', 'Bs.'),
    ('Egyptian Pound', 'EGP', '£'),
    ('Nigerian Naira', 'NGN', '₦'),
    ('Kenyan Shilling', 'KES', 'KSh'),
    ('Moroccan Dirham', 'MAD', 'DH'),
    ('Ghanaian Cedi', 'GHS', '₵'),
    ('Tunisian Dinar', 'TND', 'د.ت'),
    ('Saudi Riyal', 'SAR', '﷼'),
    ('UAE Dirham', 'AED', 'د.إ'),
    ('Kuwaiti Dinar', 'KWD', 'د.ك'),
    ('Qatari Riyal', 'QAR', '﷼'),
    ('Bahraini Dinar', 'BHD', 'ب.د'),
    ('Omani Rial', 'OMR', '﷼'),
    ('Jordanian Dinar', 'JOD', 'د.ا'),
    ('Lebanese Pound', 'LBP', 'ل.ل'),
    ('Pakistani Rupee', 'PKR', '₨'),
    ('Bangladeshi Taka', 'BDT', '৳'),
    ('Sri Lankan Rupee', 'LKR', '₨'),
    ('Vietnamese Dong', 'VND', '₫'),
    ('Myanmar Kyat', 'MMK', 'K'),
    ('Cambodian Riel', 'KHR', '៛'),
    ('Lao Kip', 'LAK', '₭'),
    ('Nepalese Rupee', 'NPR', '₨'),
    ('Afghan Afghani', 'AFN', '؋'),
    ('Kazakhstani Tenge', 'KZT', '₸'),
    ('Uzbekistani Som', 'UZS', 'so''m'),
    ('Azerbaijani Manat', 'AZN', '₼'),
    ('Georgian Lari', 'GEL', '₾'),
    ('Armenian Dram', 'AMD', '֏'),
    ('Ukrainian Hryvnia', 'UAH', '₴'),
    ('Belarusian Ruble', 'BYN', 'Br'),
    ('Moldovan Leu', 'MDL', 'L'),
    ('Romanian Leu', 'RON', 'lei'),
    ('Bulgarian Lev', 'BGN', 'лв'),
    ('Serbian Dinar', 'RSD', 'дин'),
    ('Croatian Kuna', 'HRK', 'kn'),
    ('Hungarian Forint', 'HUF', 'Ft'),
    ('Icelandic Króna', 'ISK', 'kr'),
    ('Albanian Lek', 'ALL', 'L'),
    ('Macedonian Denar', 'MKD', 'ден'),
    ('Bosnia-Herzegovina Mark', 'BAM', 'KM'),
    ('Ethiopian Birr', 'ETB', 'Br'),
    ('Tanzanian Shilling', 'TZS', 'TSh'),
    ('Ugandan Shilling', 'UGX', 'USh'),
    ('Zambian Kwacha', 'ZMW', 'ZK'),
    ('Botswana Pula', 'BWP', 'P'),
    ('Namibian Dollar', 'NAD', '$'),
    ('Mauritian Rupee', 'MUR', '₨'),
    ('Seychellois Rupee', 'SCR', '₨');

-- Insert real country data (code = phone country code)
INSERT INTO core_countries (name, code, iso2, iso3, region) VALUES
    -- North America
    ('United States', '+1', 'US', 'USA', 'north_america'),
    ('Canada', '+1', 'CA', 'CAN', 'north_america'),
    ('Mexico', '+52', 'MX', 'MEX', 'north_america'),
    -- South America
    ('Brazil', '+55', 'BR', 'BRA', 'south_america'),
    ('Argentina', '+54', 'AR', 'ARG', 'south_america'),
    ('Chile', '+56', 'CL', 'CHL', 'south_america'),
    ('Colombia', '+57', 'CO', 'COL', 'south_america'),
    ('Peru', '+51', 'PE', 'PER', 'south_america'),
    ('Venezuela', '+58', 'VE', 'VEN', 'south_america'),
    ('Uruguay', '+598', 'UY', 'URY', 'south_america'),
    -- Western Europe
    ('United Kingdom', '+44', 'GB', 'GBR', 'western_europe'),
    ('Germany', '+49', 'DE', 'DEU', 'western_europe'),
    ('France', '+33', 'FR', 'FRA', 'western_europe'),
    ('Italy', '+39', 'IT', 'ITA', 'western_europe'),
    ('Spain', '+34', 'ES', 'ESP', 'western_europe'),
    ('Netherlands', '+31', 'NL', 'NLD', 'western_europe'),
    ('Belgium', '+32', 'BE', 'BEL', 'western_europe'),
    ('Austria', '+43', 'AT', 'AUT', 'western_europe'),
    ('Portugal', '+351', 'PT', 'PRT', 'western_europe'),
    ('Greece', '+30', 'GR', 'GRC', 'western_europe'),
    ('Finland', '+358', 'FI', 'FIN', 'western_europe'),
    ('Ireland', '+353', 'IE', 'IRL', 'western_europe'),
    ('Luxembourg', '+352', 'LU', 'LUX', 'western_europe'),
    ('Switzerland', '+41', 'CH', 'CHE', 'western_europe'),
    ('Sweden', '+46', 'SE', 'SWE', 'western_europe'),
    ('Norway', '+47', 'NO', 'NOR', 'western_europe'),
    ('Denmark', '+45', 'DK', 'DNK', 'western_europe'),
    ('Iceland', '+354', 'IS', 'ISL', 'western_europe'),
    -- Eastern Europe
    ('Poland', '+48', 'PL', 'POL', 'eastern_europe'),
    ('Czech Republic', '+420', 'CZ', 'CZE', 'eastern_europe'),
    ('Hungary', '+36', 'HU', 'HUN', 'eastern_europe'),
    ('Romania', '+40', 'RO', 'ROU', 'eastern_europe'),
    ('Bulgaria', '+359', 'BG', 'BGR', 'eastern_europe'),
    ('Croatia', '+385', 'HR', 'HRV', 'eastern_europe'),
    ('Serbia', '+381', 'RS', 'SRB', 'eastern_europe'),
    ('Ukraine', '+380', 'UA', 'UKR', 'eastern_europe'),
    ('Russia', '+7', 'RU', 'RUS', 'eastern_europe'),
    ('Belarus', '+375', 'BY', 'BLR', 'eastern_europe'),
    ('Moldova', '+373', 'MD', 'MDA', 'eastern_europe'),
    ('Albania', '+355', 'AL', 'ALB', 'eastern_europe'),
    ('North Macedonia', '+389', 'MK', 'MKD', 'eastern_europe'),
    ('Bosnia and Herzegovina', '+387', 'BA', 'BIH', 'eastern_europe'),
    ('Turkey', '+90', 'TR', 'TUR', 'eastern_europe'),
    -- Asia
    ('China', '+86', 'CN', 'CHN', 'asia'),
    ('Japan', '+81', 'JP', 'JPN', 'asia'),
    ('South Korea', '+82', 'KR', 'KOR', 'asia'),
    ('India', '+91', 'IN', 'IND', 'asia'),
    ('Pakistan', '+92', 'PK', 'PAK', 'asia'),
    ('Bangladesh', '+880', 'BD', 'BGD', 'asia'),
    ('Sri Lanka', '+94', 'LK', 'LKA', 'asia'),
    ('Nepal', '+977', 'NP', 'NPL', 'asia'),
    ('Afghanistan', '+93', 'AF', 'AFG', 'asia'),
    ('Thailand', '+66', 'TH', 'THA', 'asia'),
    ('Vietnam', '+84', 'VN', 'VNM', 'asia'),
    ('Malaysia', '+60', 'MY', 'MYS', 'asia'),
    ('Singapore', '+65', 'SG', 'SGP', 'asia'),
    ('Indonesia', '+62', 'ID', 'IDN', 'asia'),
    ('Philippines', '+63', 'PH', 'PHL', 'asia'),
    ('Myanmar', '+95', 'MM', 'MMR', 'asia'),
    ('Cambodia', '+855', 'KH', 'KHM', 'asia'),
    ('Laos', '+856', 'LA', 'LAO', 'asia'),
    ('Hong Kong', '+852', 'HK', 'HKG', 'asia'),
    -- Oceania
    ('Australia', '+61', 'AU', 'AUS', 'oceania'),
    ('New Zealand', '+64', 'NZ', 'NZL', 'oceania'),
    -- Africa
    ('South Africa', '+27', 'ZA', 'ZAF', 'africa'),
    ('Egypt', '+20', 'EG', 'EGY', 'africa'),
    ('Nigeria', '+234', 'NG', 'NGA', 'africa'),
    ('Kenya', '+254', 'KE', 'KEN', 'africa'),
    ('Morocco', '+212', 'MA', 'MAR', 'africa'),
    ('Ghana', '+233', 'GH', 'GHA', 'africa'),
    ('Tunisia', '+216', 'TN', 'TUN', 'africa'),
    ('Ethiopia', '+251', 'ET', 'ETH', 'africa'),
    ('Tanzania', '+255', 'TZ', 'TZA', 'africa'),
    ('Uganda', '+256', 'UG', 'UGA', 'africa'),
    ('Zambia', '+260', 'ZM', 'ZMB', 'africa'),
    ('Botswana', '+267', 'BW', 'BWA', 'africa'),
    ('Namibia', '+264', 'NA', 'NAM', 'africa'),
    ('Mauritius', '+230', 'MU', 'MUS', 'africa'),
    ('Seychelles', '+248', 'SC', 'SYC', 'africa'),
    -- Middle East
    ('Saudi Arabia', '+966', 'SA', 'SAU', 'middle_east'),
    ('United Arab Emirates', '+971', 'AE', 'ARE', 'middle_east'),
    ('Kuwait', '+965', 'KW', 'KWT', 'middle_east'),
    ('Qatar', '+974', 'QA', 'QAT', 'middle_east'),
    ('Bahrain', '+973', 'BH', 'BHR', 'middle_east'),
    ('Oman', '+968', 'OM', 'OMN', 'middle_east'),
    ('Jordan', '+962', 'JO', 'JOR', 'middle_east'),
    ('Lebanon', '+961', 'LB', 'LBN', 'middle_east'),
    ('Israel', '+972', 'IL', 'ISR', 'middle_east'),
    ('Kazakhstan', '+7', 'KZ', 'KAZ', 'middle_east'),
    ('Uzbekistan', '+998', 'UZ', 'UZB', 'middle_east'),
    ('Azerbaijan', '+994', 'AZ', 'AZE', 'middle_east'),
    ('Georgia', '+995', 'GE', 'GEO', 'middle_east'),
    ('Armenia', '+374', 'AM', 'ARM', 'middle_east');

-- Link countries with their currencies
INSERT INTO core_country_currencies (country_id, currency_id, is_primary)
SELECT c.id, cur.id, true
FROM core_countries c
JOIN core_currencies cur ON (
    (c.code = 'US' AND cur.code = 'USD') OR
    (c.code = 'CA' AND cur.code = 'CAD') OR
    (c.code = 'GB' AND cur.code = 'GBP') OR
    (c.code = 'JP' AND cur.code = 'JPY') OR
    (c.code = 'CH' AND cur.code = 'CHF') OR
    (c.code = 'AU' AND cur.code = 'AUD') OR
    (c.code = 'CN' AND cur.code = 'CNY') OR
    (c.code = 'RU' AND cur.code = 'RUB') OR
    (c.code = 'IN' AND cur.code = 'INR') OR
    (c.code = 'BR' AND cur.code = 'BRL') OR
    (c.code = 'ZA' AND cur.code = 'ZAR') OR
    (c.code = 'MX' AND cur.code = 'MXN') OR
    (c.code = 'SG' AND cur.code = 'SGD') OR
    (c.code = 'HK' AND cur.code = 'HKD') OR
    (c.code = 'SE' AND cur.code = 'SEK') OR
    (c.code = 'NO' AND cur.code = 'NOK') OR
    (c.code = 'DK' AND cur.code = 'DKK') OR
    (c.code = 'PL' AND cur.code = 'PLN') OR
    (c.code = 'TR' AND cur.code = 'TRY') OR
    (c.code = 'KR' AND cur.code = 'KRW') OR
    (c.code = 'NZ' AND cur.code = 'NZD') OR
    (c.code = 'TH' AND cur.code = 'THB') OR
    (c.code = 'MY' AND cur.code = 'MYR') OR
    (c.code = 'ID' AND cur.code = 'IDR') OR
    (c.code = 'PH' AND cur.code = 'PHP') OR
    (c.code = 'CZ' AND cur.code = 'CZK') OR
    (c.code = 'IL' AND cur.code = 'ILS') OR
    (c.code = 'CL' AND cur.code = 'CLP') OR
    (c.code = 'CO' AND cur.code = 'COP') OR
    (c.code = 'AR' AND cur.code = 'ARS') OR
    (c.code = 'PE' AND cur.code = 'PEN') OR
    (c.code = 'VE' AND cur.code = 'VES') OR
    (c.code = 'UY' AND cur.code = 'UYU') OR
    -- Western Europe (non-Euro)
    (c.code = 'GB' AND cur.code = 'GBP') OR
    (c.code = 'CH' AND cur.code = 'CHF') OR
    (c.code = 'SE' AND cur.code = 'SEK') OR
    (c.code = 'NO' AND cur.code = 'NOK') OR
    (c.code = 'DK' AND cur.code = 'DKK') OR
    (c.code = 'IS' AND cur.code = 'ISK') OR
    -- Eastern Europe
    (c.code = 'PL' AND cur.code = 'PLN') OR
    (c.code = 'CZ' AND cur.code = 'CZK') OR
    (c.code = 'HU' AND cur.code = 'HUF') OR
    (c.code = 'RO' AND cur.code = 'RON') OR
    (c.code = 'BG' AND cur.code = 'BGN') OR
    (c.code = 'HR' AND cur.code = 'HRK') OR
    (c.code = 'RS' AND cur.code = 'RSD') OR
    (c.code = 'UA' AND cur.code = 'UAH') OR
    (c.code = 'RU' AND cur.code = 'RUB') OR
    (c.code = 'BY' AND cur.code = 'BYN') OR
    (c.code = 'MD' AND cur.code = 'MDL') OR
    (c.code = 'AL' AND cur.code = 'ALL') OR
    (c.code = 'MK' AND cur.code = 'MKD') OR
    (c.code = 'BA' AND cur.code = 'BAM') OR
    (c.code = 'TR' AND cur.code = 'TRY') OR
    -- Asia
    (c.code = 'CN' AND cur.code = 'CNY') OR
    (c.code = 'JP' AND cur.code = 'JPY') OR
    (c.code = 'KR' AND cur.code = 'KRW') OR
    (c.code = 'IN' AND cur.code = 'INR') OR
    (c.code = 'PK' AND cur.code = 'PKR') OR
    (c.code = 'BD' AND cur.code = 'BDT') OR
    (c.code = 'LK' AND cur.code = 'LKR') OR
    (c.code = 'NP' AND cur.code = 'NPR') OR
    (c.code = 'AF' AND cur.code = 'AFN') OR
    (c.code = 'TH' AND cur.code = 'THB') OR
    (c.code = 'VN' AND cur.code = 'VND') OR
    (c.code = 'MY' AND cur.code = 'MYR') OR
    (c.code = 'SG' AND cur.code = 'SGD') OR
    (c.code = 'ID' AND cur.code = 'IDR') OR
    (c.code = 'PH' AND cur.code = 'PHP') OR
    (c.code = 'MM' AND cur.code = 'MMK') OR
    (c.code = 'KH' AND cur.code = 'KHR') OR
    (c.code = 'LA' AND cur.code = 'LAK') OR
    (c.code = 'HK' AND cur.code = 'HKD') OR
    -- Oceania
    (c.code = 'AU' AND cur.code = 'AUD') OR
    (c.code = 'NZ' AND cur.code = 'NZD') OR
    -- Africa
    (c.code = 'ZA' AND cur.code = 'ZAR') OR
    (c.code = 'EG' AND cur.code = 'EGP') OR
    (c.code = 'NG' AND cur.code = 'NGN') OR
    (c.code = 'KE' AND cur.code = 'KES') OR
    (c.code = 'MA' AND cur.code = 'MAD') OR
    (c.code = 'GH' AND cur.code = 'GHS') OR
    (c.code = 'TN' AND cur.code = 'TND') OR
    (c.code = 'ET' AND cur.code = 'ETB') OR
    (c.code = 'TZ' AND cur.code = 'TZS') OR
    (c.code = 'UG' AND cur.code = 'UGX') OR
    (c.code = 'ZM' AND cur.code = 'ZMW') OR
    (c.code = 'BW' AND cur.code = 'BWP') OR
    (c.code = 'NA' AND cur.code = 'NAD') OR
    (c.code = 'MU' AND cur.code = 'MUR') OR
    (c.code = 'SC' AND cur.code = 'SCR') OR
    -- Middle East
    (c.code = 'SA' AND cur.code = 'SAR') OR
    (c.code = 'AE' AND cur.code = 'AED') OR
    (c.code = 'KW' AND cur.code = 'KWD') OR
    (c.code = 'QA' AND cur.code = 'QAR') OR
    (c.code = 'BH' AND cur.code = 'BHD') OR
    (c.code = 'OM' AND cur.code = 'OMR') OR
    (c.code = 'JO' AND cur.code = 'JOD') OR
    (c.code = 'LB' AND cur.code = 'LBP') OR
    (c.code = 'IL' AND cur.code = 'ILS') OR
    -- Central Asia
    (c.code = 'KZ' AND cur.code = 'KZT') OR
    (c.code = 'UZ' AND cur.code = 'UZS') OR
    (c.code = 'AZ' AND cur.code = 'AZN') OR
    (c.code = 'GE' AND cur.code = 'GEL') OR
    (c.code = 'AM' AND cur.code = 'AMD')
);

-- Eurozone countries with EUR
INSERT INTO core_country_currencies (country_id, currency_id, is_primary)
SELECT c.id, cur.id, true
FROM core_countries c
CROSS JOIN core_currencies cur
WHERE cur.code = 'EUR' 
AND c.code IN ('DE', 'FR', 'IT', 'ES', 'NL', 'BE', 'AT', 'PT', 'GR', 'FI', 'IE', 'LU');

-- Comment on tables
COMMENT ON TABLE core_countries IS 'Countries reference data';
COMMENT ON COLUMN core_countries.code IS 'Country calling code (phone prefix, e.g., +1, +44, +7)';
COMMENT ON COLUMN core_countries.iso2 IS 'ISO 3166-1 alpha-2 code (2 letters)';
COMMENT ON COLUMN core_countries.iso3 IS 'ISO 3166-1 alpha-3 code (3 letters)';
COMMENT ON TABLE core_currencies IS 'Currencies reference data';
COMMENT ON TABLE core_country_currencies IS 'Many-to-many relationship between countries and their currencies';
