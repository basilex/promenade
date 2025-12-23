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
    ('Seychellois Rupee', 'SCR', '₨'),
    -- Additional currencies for completeness
    ('Algerian Dinar', 'DZD', 'د.ج'),
    ('Angolan Kwanza', 'AOA', 'Kz'),
    ('Brunei Dollar', 'BND', 'B$'),
    ('Central African CFA Franc', 'XAF', 'FCFA'),
    ('West African CFA Franc', 'XOF', 'CFA'),
    ('Costa Rican Colón', 'CRC', '₡'),
    ('Cuban Peso', 'CUP', '₱'),
    ('Dominican Peso', 'DOP', 'RD$'),
    ('Eritrean Nakfa', 'ERN', 'Nfk'),
    ('Guatemalan Quetzal', 'GTQ', 'Q'),
    ('Honduran Lempira', 'HNL', 'L'),
    ('Haitian Gourde', 'HTG', 'G'),
    ('Iranian Rial', 'IRR', '﷼'),
    ('Iraqi Dinar', 'IQD', 'ع.د'),
    ('Jamaican Dollar', 'JMD', 'J$'),
    ('Libyan Dinar', 'LYD', 'ل.د'),
    ('Maldivian Rufiyaa', 'MVR', 'Rf'),
    ('Mongolian Tögrög', 'MNT', '₮'),
    ('Mozambican Metical', 'MZN', 'MT'),
    ('Nicaraguan Córdoba', 'NIO', 'C$'),
    ('Panamanian Balboa', 'PAB', 'B/.'),
    ('Paraguayan Guaraní', 'PYG', '₲'),
    ('Rwandan Franc', 'RWF', 'FRw'),
    ('Sudanese Pound', 'SDG', 'ج.س'),
    ('Syrian Pound', 'SYP', '£S'),
    ('Trinidad and Tobago Dollar', 'TTD', 'TT$'),
    ('New Taiwan Dollar', 'TWD', 'NT$'),
    ('Yemeni Rial', 'YER', '﷼'),
    ('Zimbabwean Dollar', 'ZWL', 'Z$'),
    ('Lithuanian Litas', 'LTL', 'Lt'),
    ('Latvian Lats', 'LVL', 'Ls'),
    ('Estonian Kroon', 'EEK', 'kr'),
    ('Slovenian Tolar', 'SIT', 'SIT'),
    ('Maltese Lira', 'MTL', '₤'),
    ('Cypriot Pound', 'CYP', '£'),
    ('Slovak Koruna', 'SKK', 'Sk'),
    ('Bolivian Boliviano', 'BOB', 'Bs.'),
    ('Congolese Franc', 'CDF', 'FC'),
    ('Malagasy Ariary', 'MGA', 'Ar'),
    ('Turkmenistan Manat', 'TMT', 'm'),
    ('Tajikistani Somoni', 'TJS', 'ЅМ'),
    ('Kyrgyzstani Som', 'KGS', 'с');

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
    ('Armenia', '+374', 'AM', 'ARM', 'middle_east'),
    -- Additional countries for completeness
    ('Taiwan', '+886', 'TW', 'TWN', 'asia'),
    ('Iran', '+98', 'IR', 'IRN', 'middle_east'),
    ('Iraq', '+964', 'IQ', 'IRQ', 'middle_east'),
    ('Syria', '+963', 'SY', 'SYR', 'middle_east'),
    ('Yemen', '+967', 'YE', 'YEM', 'middle_east'),
    ('Slovakia', '+421', 'SK', 'SVK', 'eastern_europe'),
    ('Slovenia', '+386', 'SI', 'SVN', 'eastern_europe'),
    ('Estonia', '+372', 'EE', 'EST', 'eastern_europe'),
    ('Latvia', '+371', 'LV', 'LVA', 'eastern_europe'),
    ('Lithuania', '+370', 'LT', 'LTU', 'eastern_europe'),
    ('Malta', '+356', 'MT', 'MLT', 'western_europe'),
    ('Cyprus', '+357', 'CY', 'CYP', 'western_europe'),
    ('Costa Rica', '+506', 'CR', 'CRI', 'north_america'),
    ('Panama', '+507', 'PA', 'PAN', 'north_america'),
    ('Guatemala', '+502', 'GT', 'GTM', 'north_america'),
    ('Honduras', '+504', 'HN', 'HND', 'north_america'),
    ('Nicaragua', '+505', 'NI', 'NIC', 'north_america'),
    ('El Salvador', '+503', 'SV', 'SLV', 'north_america'),
    ('Cuba', '+53', 'CU', 'CUB', 'north_america'),
    ('Jamaica', '+1876', 'JM', 'JAM', 'north_america'),
    ('Dominican Republic', '+1809', 'DO', 'DOM', 'north_america'),
    ('Haiti', '+509', 'HT', 'HTI', 'north_america'),
    ('Trinidad and Tobago', '+1868', 'TT', 'TTO', 'north_america'),
    ('Paraguay', '+595', 'PY', 'PRY', 'south_america'),
    ('Ecuador', '+593', 'EC', 'ECU', 'south_america'),
    ('Bolivia', '+591', 'BO', 'BOL', 'south_america'),
    ('Algeria', '+213', 'DZ', 'DZA', 'africa'),
    ('Angola', '+244', 'AO', 'AGO', 'africa'),
    ('Rwanda', '+250', 'RW', 'RWA', 'africa'),
    ('Mozambique', '+258', 'MZ', 'MOZ', 'africa'),
    ('Zimbabwe', '+263', 'ZW', 'ZWE', 'africa'),
    ('Libya', '+218', 'LY', 'LBY', 'africa'),
    ('Sudan', '+249', 'SD', 'SDN', 'africa'),
    ('Eritrea', '+291', 'ER', 'ERI', 'africa'),
    ('Benin', '+229', 'BJ', 'BEN', 'africa'),
    ('Cameroon', '+237', 'CM', 'CMR', 'africa'),
    ('Ivory Coast', '+225', 'CI', 'CIV', 'africa'),
    ('Senegal', '+221', 'SN', 'SEN', 'africa'),
    ('Mali', '+223', 'ML', 'MLI', 'africa'),
    ('Burkina Faso', '+226', 'BF', 'BFA', 'africa'),
    ('Niger', '+227', 'NE', 'NER', 'africa'),
    ('Chad', '+235', 'TD', 'TCD', 'africa'),
    ('Gabon', '+241', 'GA', 'GAB', 'africa'),
    ('Congo', '+242', 'CG', 'COG', 'africa'),
    ('Democratic Republic of Congo', '+243', 'CD', 'COD', 'africa'),
    ('Madagascar', '+261', 'MG', 'MDG', 'africa'),
    ('Maldives', '+960', 'MV', 'MDV', 'asia'),
    ('Brunei', '+673', 'BN', 'BRN', 'asia'),
    ('Mongolia', '+976', 'MN', 'MNG', 'asia'),
    ('Turkmenistan', '+993', 'TM', 'TKM', 'middle_east'),
    ('Tajikistan', '+992', 'TJ', 'TJK', 'middle_east'),
    ('Kyrgyzstan', '+996', 'KG', 'KGZ', 'middle_east');

-- Link countries with their currencies
INSERT INTO core_country_currencies (country_id, currency_id, is_primary)
SELECT c.id, cur.id, true
FROM core_countries c
JOIN core_currencies cur ON (
    -- North America
    (c.iso2 = 'US' AND cur.code = 'USD') OR
    (c.iso2 = 'CA' AND cur.code = 'CAD') OR
    (c.iso2 = 'MX' AND cur.code = 'MXN') OR
    -- South America
    (c.iso2 = 'BR' AND cur.code = 'BRL') OR
    (c.iso2 = 'AR' AND cur.code = 'ARS') OR
    (c.iso2 = 'CL' AND cur.code = 'CLP') OR
    (c.iso2 = 'CO' AND cur.code = 'COP') OR
    (c.iso2 = 'PE' AND cur.code = 'PEN') OR
    (c.iso2 = 'VE' AND cur.code = 'VES') OR
    (c.iso2 = 'UY' AND cur.code = 'UYU') OR
    (c.iso2 = 'PY' AND cur.code = 'PYG') OR
    (c.iso2 = 'BO' AND cur.code = 'BOB') OR
    -- Western Europe (non-Euro)
    (c.iso2 = 'GB' AND cur.code = 'GBP') OR
    (c.iso2 = 'CH' AND cur.code = 'CHF') OR
    (c.iso2 = 'SE' AND cur.code = 'SEK') OR
    (c.iso2 = 'NO' AND cur.code = 'NOK') OR
    (c.iso2 = 'DK' AND cur.code = 'DKK') OR
    (c.iso2 = 'IS' AND cur.code = 'ISK') OR
    -- Eastern Europe
    (c.iso2 = 'PL' AND cur.code = 'PLN') OR
    (c.iso2 = 'CZ' AND cur.code = 'CZK') OR
    (c.iso2 = 'HU' AND cur.code = 'HUF') OR
    (c.iso2 = 'RO' AND cur.code = 'RON') OR
    (c.iso2 = 'BG' AND cur.code = 'BGN') OR
    (c.iso2 = 'HR' AND cur.code = 'HRK') OR
    (c.iso2 = 'RS' AND cur.code = 'RSD') OR
    (c.iso2 = 'UA' AND cur.code = 'UAH') OR
    (c.iso2 = 'RU' AND cur.code = 'RUB') OR
    (c.iso2 = 'BY' AND cur.code = 'BYN') OR
    (c.iso2 = 'MD' AND cur.code = 'MDL') OR
    (c.iso2 = 'AL' AND cur.code = 'ALL') OR
    (c.iso2 = 'MK' AND cur.code = 'MKD') OR
    (c.iso2 = 'BA' AND cur.code = 'BAM') OR
    (c.iso2 = 'TR' AND cur.code = 'TRY') OR
    -- Asia
    (c.iso2 = 'CN' AND cur.code = 'CNY') OR
    (c.iso2 = 'JP' AND cur.code = 'JPY') OR
    (c.iso2 = 'KR' AND cur.code = 'KRW') OR
    (c.iso2 = 'IN' AND cur.code = 'INR') OR
    (c.iso2 = 'PK' AND cur.code = 'PKR') OR
    (c.iso2 = 'BD' AND cur.code = 'BDT') OR
    (c.iso2 = 'LK' AND cur.code = 'LKR') OR
    (c.iso2 = 'NP' AND cur.code = 'NPR') OR
    (c.iso2 = 'AF' AND cur.code = 'AFN') OR
    (c.iso2 = 'TH' AND cur.code = 'THB') OR
    (c.iso2 = 'VN' AND cur.code = 'VND') OR
    (c.iso2 = 'MY' AND cur.code = 'MYR') OR
    (c.iso2 = 'SG' AND cur.code = 'SGD') OR
    (c.iso2 = 'ID' AND cur.code = 'IDR') OR
    (c.iso2 = 'PH' AND cur.code = 'PHP') OR
    (c.iso2 = 'MM' AND cur.code = 'MMK') OR
    (c.iso2 = 'KH' AND cur.code = 'KHR') OR
    (c.iso2 = 'LA' AND cur.code = 'LAK') OR
    (c.iso2 = 'HK' AND cur.code = 'HKD') OR
    (c.iso2 = 'TW' AND cur.code = 'TWD') OR
    (c.iso2 = 'MV' AND cur.code = 'MVR') OR
    (c.iso2 = 'BN' AND cur.code = 'BND') OR
    (c.iso2 = 'MN' AND cur.code = 'MNT') OR
    -- Oceania
    (c.iso2 = 'AU' AND cur.code = 'AUD') OR
    (c.iso2 = 'NZ' AND cur.code = 'NZD') OR
    -- Africa
    (c.iso2 = 'ZA' AND cur.code = 'ZAR') OR
    (c.iso2 = 'EG' AND cur.code = 'EGP') OR
    (c.iso2 = 'NG' AND cur.code = 'NGN') OR
    (c.iso2 = 'KE' AND cur.code = 'KES') OR
    (c.iso2 = 'MA' AND cur.code = 'MAD') OR
    (c.iso2 = 'GH' AND cur.code = 'GHS') OR
    (c.iso2 = 'TN' AND cur.code = 'TND') OR
    (c.iso2 = 'ET' AND cur.code = 'ETB') OR
    (c.iso2 = 'TZ' AND cur.code = 'TZS') OR
    (c.iso2 = 'UG' AND cur.code = 'UGX') OR
    (c.iso2 = 'ZM' AND cur.code = 'ZMW') OR
    (c.iso2 = 'BW' AND cur.code = 'BWP') OR
    (c.iso2 = 'NA' AND cur.code = 'NAD') OR
    (c.iso2 = 'MU' AND cur.code = 'MUR') OR
    (c.iso2 = 'SC' AND cur.code = 'SCR') OR
    (c.iso2 = 'DZ' AND cur.code = 'DZD') OR
    (c.iso2 = 'AO' AND cur.code = 'AOA') OR
    (c.iso2 = 'RW' AND cur.code = 'RWF') OR
    (c.iso2 = 'MZ' AND cur.code = 'MZN') OR
    (c.iso2 = 'ZW' AND cur.code = 'ZWL') OR
    (c.iso2 = 'LY' AND cur.code = 'LYD') OR
    (c.iso2 = 'SD' AND cur.code = 'SDG') OR
    (c.iso2 = 'ER' AND cur.code = 'ERN') OR
    (c.iso2 = 'MG' AND cur.code = 'MGA') OR
    -- West/Central Africa (CFA Franc zones)
    (c.iso2 = 'BJ' AND cur.code = 'XOF') OR
    (c.iso2 = 'CI' AND cur.code = 'XOF') OR
    (c.iso2 = 'SN' AND cur.code = 'XOF') OR
    (c.iso2 = 'ML' AND cur.code = 'XOF') OR
    (c.iso2 = 'BF' AND cur.code = 'XOF') OR
    (c.iso2 = 'NE' AND cur.code = 'XOF') OR
    (c.iso2 = 'CM' AND cur.code = 'XAF') OR
    (c.iso2 = 'TD' AND cur.code = 'XAF') OR
    (c.iso2 = 'GA' AND cur.code = 'XAF') OR
    (c.iso2 = 'CG' AND cur.code = 'XAF') OR
    (c.iso2 = 'CD' AND cur.code = 'CDF') OR
    -- Middle East
    (c.iso2 = 'SA' AND cur.code = 'SAR') OR
    (c.iso2 = 'AE' AND cur.code = 'AED') OR
    (c.iso2 = 'KW' AND cur.code = 'KWD') OR
    (c.iso2 = 'QA' AND cur.code = 'QAR') OR
    (c.iso2 = 'BH' AND cur.code = 'BHD') OR
    (c.iso2 = 'OM' AND cur.code = 'OMR') OR
    (c.iso2 = 'JO' AND cur.code = 'JOD') OR
    (c.iso2 = 'LB' AND cur.code = 'LBP') OR
    (c.iso2 = 'IL' AND cur.code = 'ILS') OR
    (c.iso2 = 'IR' AND cur.code = 'IRR') OR
    (c.iso2 = 'IQ' AND cur.code = 'IQD') OR
    (c.iso2 = 'SY' AND cur.code = 'SYP') OR
    (c.iso2 = 'YE' AND cur.code = 'YER') OR
    -- Central Asia
    (c.iso2 = 'KZ' AND cur.code = 'KZT') OR
    (c.iso2 = 'UZ' AND cur.code = 'UZS') OR
    (c.iso2 = 'AZ' AND cur.code = 'AZN') OR
    (c.iso2 = 'GE' AND cur.code = 'GEL') OR
    (c.iso2 = 'AM' AND cur.code = 'AMD') OR
    (c.iso2 = 'TM' AND cur.code = 'TMT') OR
    (c.iso2 = 'TJ' AND cur.code = 'TJS') OR
    (c.iso2 = 'KG' AND cur.code = 'KGS') OR
    -- Central America & Caribbean
    (c.iso2 = 'CR' AND cur.code = 'CRC') OR
    (c.iso2 = 'PA' AND cur.code = 'PAB') OR
    (c.iso2 = 'GT' AND cur.code = 'GTQ') OR
    (c.iso2 = 'HN' AND cur.code = 'HNL') OR
    (c.iso2 = 'NI' AND cur.code = 'NIO') OR
    (c.iso2 = 'CU' AND cur.code = 'CUP') OR
    (c.iso2 = 'JM' AND cur.code = 'JMD') OR
    (c.iso2 = 'DO' AND cur.code = 'DOP') OR
    (c.iso2 = 'HT' AND cur.code = 'HTG') OR
    (c.iso2 = 'TT' AND cur.code = 'TTD')
);

-- Eurozone countries with EUR (including newer members)
INSERT INTO core_country_currencies (country_id, currency_id, is_primary)
SELECT c.id, cur.id, true
FROM core_countries c
CROSS JOIN core_currencies cur
WHERE cur.code = 'EUR' 
AND c.iso2 IN ('DE', 'FR', 'IT', 'ES', 'NL', 'BE', 'AT', 'PT', 'GR', 'FI', 'IE', 'LU', 
               'SK', 'SI', 'EE', 'LV', 'LT', 'MT', 'CY');

-- Countries using USD as primary currency (El Salvador, Ecuador) or alongside their own (Panama with PAB)
INSERT INTO core_country_currencies (country_id, currency_id, is_primary)
SELECT c.id, cur.id, CASE WHEN c.iso2 IN ('SV', 'EC') THEN true ELSE false END
FROM core_countries c
CROSS JOIN core_currencies cur
WHERE cur.code = 'USD' 
AND c.iso2 IN ('SV', 'EC', 'PA');

-- Comment on tables
COMMENT ON TABLE core_countries IS 'Countries reference data';
COMMENT ON COLUMN core_countries.code IS 'Country calling code (phone prefix, e.g., +1, +44, +7)';
COMMENT ON COLUMN core_countries.iso2 IS 'ISO 3166-1 alpha-2 code (2 letters)';
COMMENT ON COLUMN core_countries.iso3 IS 'ISO 3166-1 alpha-3 code (3 letters)';
COMMENT ON TABLE core_currencies IS 'Currencies reference data';
COMMENT ON TABLE core_country_currencies IS 'Many-to-many relationship between countries and their currencies';
