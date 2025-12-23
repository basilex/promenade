-- +goose Up
-- +goose StatementBegin

-- Regions/States/Provinces/Oblasts table
CREATE TABLE core_regions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    country_id UUID NOT NULL REFERENCES core_countries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NOT NULL, -- ISO 3166-2 subdivision code or local code
    region_type VARCHAR(30) NOT NULL, -- state, province, oblast, region, territory, district
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_core_region_code UNIQUE (country_id, code)
);

COMMENT ON TABLE core_regions IS 'Administrative regions: states, provinces, oblasts, etc.';
COMMENT ON COLUMN core_regions.code IS 'ISO 3166-2 subdivision code or local standard code';
COMMENT ON COLUMN core_regions.region_type IS 'Type: state, province, oblast, region, territory, district, etc.';
COMMENT ON COLUMN core_regions.latitude IS 'Region center latitude for mapping/distance calculations';
COMMENT ON COLUMN core_regions.longitude IS 'Region center longitude for mapping/distance calculations';

-- Indexes for regions
CREATE INDEX idx_core_regions_country ON core_regions(country_id);
CREATE INDEX idx_core_regions_code ON core_regions(code);
CREATE INDEX idx_core_regions_name ON core_regions(name);
CREATE INDEX idx_core_regions_active ON core_regions(is_active);

-- Trigger for updated_at
CREATE TRIGGER trg_core_regions_updated_at
    BEFORE UPDATE ON core_regions
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- Cities table
CREATE TABLE core_cities (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    region_id UUID REFERENCES core_regions(id) ON DELETE SET NULL,
    country_id UUID NOT NULL REFERENCES core_countries(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    name_local VARCHAR(100), -- Name in local language
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    population INTEGER,
    is_capital BOOLEAN NOT NULL DEFAULT FALSE, -- Capital of country
    is_regional_capital BOOLEAN NOT NULL DEFAULT FALSE, -- Capital of region/state
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE core_cities IS 'Cities and major settlements';
COMMENT ON COLUMN core_cities.name_local IS 'City name in local language/script';
COMMENT ON COLUMN core_cities.is_capital IS 'TRUE if capital of the country';
COMMENT ON COLUMN core_cities.is_regional_capital IS 'TRUE if capital of the region/state/province';
COMMENT ON COLUMN core_cities.population IS 'Approximate population for sorting/filtering';

-- Indexes for cities
CREATE INDEX idx_core_cities_country ON core_cities(country_id);
CREATE INDEX idx_core_cities_region ON core_cities(region_id);
CREATE INDEX idx_core_cities_name ON core_cities(name);
CREATE INDEX idx_core_cities_active ON core_cities(is_active);
CREATE INDEX idx_core_cities_capital ON core_cities(is_capital) WHERE is_capital = TRUE;
CREATE INDEX idx_core_cities_coordinates ON core_cities(latitude, longitude) WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

-- Trigger for updated_at
CREATE TRIGGER trg_core_cities_updated_at
    BEFORE UPDATE ON core_cities
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();

-- Insert sample data for popular regions (examples for major countries)

-- USA States (top 10 by population)
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'US'), 'California', 'CA', 'state', 36.7783, -119.4179, 1),
((SELECT id FROM core_countries WHERE iso2 = 'US'), 'Texas', 'TX', 'state', 31.9686, -99.9018, 2),
((SELECT id FROM core_countries WHERE iso2 = 'US'), 'Florida', 'FL', 'state', 27.6648, -81.5158, 3),
((SELECT id FROM core_countries WHERE iso2 = 'US'), 'New York', 'NY', 'state', 43.2994, -74.2179, 4),
((SELECT id FROM core_countries WHERE iso2 = 'US'), 'Pennsylvania', 'PA', 'state', 41.2033, -77.1945, 5);

-- Russia Federal Subjects (top regions)
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'RU'), 'Moscow', 'MOW', 'federal_city', 55.7558, 37.6173, 1),
((SELECT id FROM core_countries WHERE iso2 = 'RU'), 'Saint Petersburg', 'SPE', 'federal_city', 59.9311, 30.3609, 2),
((SELECT id FROM core_countries WHERE iso2 = 'RU'), 'Moscow Oblast', 'MOS', 'oblast', 55.5815, 37.0708, 3),
((SELECT id FROM core_countries WHERE iso2 = 'RU'), 'Krasnodar Krai', 'KDA', 'krai', 45.0355, 38.9753, 4);

-- Ukraine Oblasts (top regions)
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'UA'), 'Kyiv City', 'KY', 'city', 50.4501, 30.5234, 1),
((SELECT id FROM core_countries WHERE iso2 = 'UA'), 'Kyiv Oblast', 'KV', 'oblast', 50.0521, 30.7617, 2),
((SELECT id FROM core_countries WHERE iso2 = 'UA'), 'Lviv Oblast', 'LV', 'oblast', 49.8397, 24.0297, 3),
((SELECT id FROM core_countries WHERE iso2 = 'UA'), 'Odesa Oblast', 'OD', 'oblast', 46.4825, 30.7233, 4);

-- UK Countries/Regions
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'GB'), 'England', 'ENG', 'country', 52.3555, -1.1743, 1),
((SELECT id FROM core_countries WHERE iso2 = 'GB'), 'Scotland', 'SCT', 'country', 56.4907, -4.2026, 2),
((SELECT id FROM core_countries WHERE iso2 = 'GB'), 'Wales', 'WLS', 'country', 52.1307, -3.7837, 3),
((SELECT id FROM core_countries WHERE iso2 = 'GB'), 'Northern Ireland', 'NIR', 'country', 54.7877, -6.4923, 4);

-- Germany States (Länder)
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'DE'), 'Bavaria', 'BY', 'state', 48.7904, 11.4979, 1),
((SELECT id FROM core_countries WHERE iso2 = 'DE'), 'Baden-Württemberg', 'BW', 'state', 48.6616, 9.3501, 2),
((SELECT id FROM core_countries WHERE iso2 = 'DE'), 'North Rhine-Westphalia', 'NW', 'state', 51.4332, 7.6616, 3),
((SELECT id FROM core_countries WHERE iso2 = 'DE'), 'Berlin', 'BE', 'city_state', 52.5200, 13.4050, 4);

-- France Regions
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'FR'), 'Île-de-France', 'IDF', 'region', 48.8499, 2.6370, 1),
((SELECT id FROM core_countries WHERE iso2 = 'FR'), 'Provence-Alpes-Côte d''Azur', 'PAC', 'region', 43.9351, 6.0679, 2),
((SELECT id FROM core_countries WHERE iso2 = 'FR'), 'Auvergne-Rhône-Alpes', 'ARA', 'region', 45.4472, 4.3853, 3);

-- Canada Provinces
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'CA'), 'Ontario', 'ON', 'province', 51.2538, -85.3232, 1),
((SELECT id FROM core_countries WHERE iso2 = 'CA'), 'Quebec', 'QC', 'province', 52.9399, -73.5491, 2),
((SELECT id FROM core_countries WHERE iso2 = 'CA'), 'British Columbia', 'BC', 'province', 53.7267, -127.6476, 3);

-- Australia States
INSERT INTO core_regions (country_id, name, code, region_type, latitude, longitude, sort_order) VALUES
((SELECT id FROM core_countries WHERE iso2 = 'AU'), 'New South Wales', 'NSW', 'state', -31.2532, 146.9211, 1),
((SELECT id FROM core_countries WHERE iso2 = 'AU'), 'Victoria', 'VIC', 'state', -37.4713, 144.7852, 2),
((SELECT id FROM core_countries WHERE iso2 = 'AU'), 'Queensland', 'QLD', 'state', -20.9176, 142.7028, 3);

-- Insert sample cities (world capitals and major cities)

-- USA
INSERT INTO core_cities (region_id, country_id, name, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'CA' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'US')), 
 (SELECT id FROM core_countries WHERE iso2 = 'US'), 'Los Angeles', 34.0522, -118.2437, 3900000, FALSE, FALSE, 1),
((SELECT id FROM core_regions WHERE code = 'CA' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'US')), 
 (SELECT id FROM core_countries WHERE iso2 = 'US'), 'San Francisco', 37.7749, -122.4194, 870000, FALSE, FALSE, 2),
((SELECT id FROM core_regions WHERE code = 'NY' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'US')), 
 (SELECT id FROM core_countries WHERE iso2 = 'US'), 'New York City', 40.7128, -74.0060, 8400000, FALSE, FALSE, 3);

-- Russia
INSERT INTO core_cities (region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'MOW' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'RU')), 
 (SELECT id FROM core_countries WHERE iso2 = 'RU'), 'Moscow', 'Москва', 55.7558, 37.6173, 12600000, TRUE, TRUE, 1),
((SELECT id FROM core_regions WHERE code = 'SPE' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'RU')), 
 (SELECT id FROM core_countries WHERE iso2 = 'RU'), 'Saint Petersburg', 'Санкт-Петербург', 59.9311, 30.3609, 5400000, FALSE, TRUE, 2);

-- Ukraine
INSERT INTO core_cities (region_id, country_id, name, name_local, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'KY' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'UA')), 
 (SELECT id FROM core_countries WHERE iso2 = 'UA'), 'Kyiv', 'Київ', 50.4501, 30.5234, 2900000, TRUE, TRUE, 1),
((SELECT id FROM core_regions WHERE code = 'LV' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'UA')), 
 (SELECT id FROM core_countries WHERE iso2 = 'UA'), 'Lviv', 'Львів', 49.8397, 24.0297, 720000, FALSE, TRUE, 2);

-- UK
INSERT INTO core_cities (region_id, country_id, name, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'ENG' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'GB')), 
 (SELECT id FROM core_countries WHERE iso2 = 'GB'), 'London', 51.5074, -0.1278, 9000000, TRUE, TRUE, 1),
((SELECT id FROM core_regions WHERE code = 'ENG' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'GB')), 
 (SELECT id FROM core_countries WHERE iso2 = 'GB'), 'Manchester', 53.4808, -2.2426, 550000, FALSE, FALSE, 2);

-- Germany
INSERT INTO core_cities (region_id, country_id, name, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'BE' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'DE')), 
 (SELECT id FROM core_countries WHERE iso2 = 'DE'), 'Berlin', 52.5200, 13.4050, 3600000, TRUE, TRUE, 1),
((SELECT id FROM core_regions WHERE code = 'BY' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'DE')), 
 (SELECT id FROM core_countries WHERE iso2 = 'DE'), 'Munich', 48.1351, 11.5820, 1500000, FALSE, TRUE, 2);

-- France
INSERT INTO core_cities (region_id, country_id, name, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'IDF' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'FR')), 
 (SELECT id FROM core_countries WHERE iso2 = 'FR'), 'Paris', 48.8566, 2.3522, 2200000, TRUE, TRUE, 1),
((SELECT id FROM core_regions WHERE code = 'PAC' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'FR')), 
 (SELECT id FROM core_countries WHERE iso2 = 'FR'), 'Marseille', 43.2965, 5.3698, 870000, FALSE, FALSE, 2);

-- Canada
INSERT INTO core_cities (region_id, country_id, name, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'ON' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'CA')), 
 (SELECT id FROM core_countries WHERE iso2 = 'CA'), 'Toronto', 43.6532, -79.3832, 2900000, FALSE, TRUE, 1),
((SELECT id FROM core_regions WHERE code = 'QC' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'CA')), 
 (SELECT id FROM core_countries WHERE iso2 = 'CA'), 'Montreal', 45.5017, -73.5673, 1700000, FALSE, FALSE, 2);

-- Australia
INSERT INTO core_cities (region_id, country_id, name, latitude, longitude, population, is_capital, is_regional_capital, sort_order) VALUES
((SELECT id FROM core_regions WHERE code = 'NSW' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'AU')), 
 (SELECT id FROM core_countries WHERE iso2 = 'AU'), 'Sydney', -33.8688, 151.2093, 5300000, FALSE, TRUE, 1),
((SELECT id FROM core_regions WHERE code = 'VIC' AND country_id = (SELECT id FROM core_countries WHERE iso2 = 'AU')), 
 (SELECT id FROM core_countries WHERE iso2 = 'AU'), 'Melbourne', -37.8136, 144.9631, 5100000, FALSE, TRUE, 2);

-- +goose StatementEnd
