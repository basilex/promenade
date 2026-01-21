-- Drop junction tables first (foreign keys)
IF OBJECT_ID('dbo.shared_country_timezones', 'U') IS NOT NULL
    DROP TABLE dbo.shared_country_timezones;
GO

IF OBJECT_ID('dbo.shared_country_languages', 'U') IS NOT NULL
    DROP TABLE dbo.shared_country_languages;
GO

IF OBJECT_ID('dbo.shared_country_currencies', 'U') IS NOT NULL
    DROP TABLE dbo.shared_country_currencies;
GO

-- Drop reference tables (Shared Kernel)
IF OBJECT_ID('dbo.shared_timezones', 'U') IS NOT NULL
    DROP TABLE dbo.shared_timezones;
GO

IF OBJECT_ID('dbo.shared_languages', 'U') IS NOT NULL
    DROP TABLE dbo.shared_languages;
GO

IF OBJECT_ID('dbo.shared_currencies', 'U') IS NOT NULL
    DROP TABLE dbo.shared_currencies;
GO

IF OBJECT_ID('dbo.shared_countries', 'U') IS NOT NULL
    DROP TABLE dbo.shared_countries;
GO
