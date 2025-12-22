-- Create timezones table for timezone reference data
CREATE TABLE IF NOT EXISTS timezones (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) NOT NULL UNIQUE,          -- IANA timezone name (e.g., "Europe/Moscow")
    abbreviation VARCHAR(10) NOT NULL,          -- Timezone abbreviation (e.g., "MSK")
    utc_offset VARCHAR(10) NOT NULL,            -- UTC offset (e.g., "+03:00")
    utc_dst_offset VARCHAR(10),                 -- DST offset (e.g., "+04:00")
    description VARCHAR(255),                   -- Human-readable description
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on name for fast lookups
CREATE INDEX idx_timezones_name ON timezones(name);
CREATE INDEX idx_timezones_is_active ON timezones(is_active);

-- Insert common timezones
INSERT INTO timezones (name, abbreviation, utc_offset, utc_dst_offset, description) VALUES
('UTC', 'UTC', '+00:00', NULL, 'Coordinated Universal Time'),
('Europe/London', 'GMT', '+00:00', '+01:00', 'Greenwich Mean Time / British Summer Time'),
('Europe/Paris', 'CET', '+01:00', '+02:00', 'Central European Time'),
('Europe/Moscow', 'MSK', '+03:00', NULL, 'Moscow Standard Time'),
('America/New_York', 'EST', '-05:00', '-04:00', 'Eastern Standard Time'),
('America/Chicago', 'CST', '-06:00', '-05:00', 'Central Standard Time'),
('America/Denver', 'MST', '-07:00', '-06:00', 'Mountain Standard Time'),
('America/Los_Angeles', 'PST', '-08:00', '-07:00', 'Pacific Standard Time'),
('Asia/Dubai', 'GST', '+04:00', NULL, 'Gulf Standard Time'),
('Asia/Kolkata', 'IST', '+05:30', NULL, 'India Standard Time'),
('Asia/Shanghai', 'CST', '+08:00', NULL, 'China Standard Time'),
('Asia/Tokyo', 'JST', '+09:00', NULL, 'Japan Standard Time'),
('Australia/Sydney', 'AEDT', '+10:00', '+11:00', 'Australian Eastern Daylight Time'),
('Pacific/Auckland', 'NZDT', '+12:00', '+13:00', 'New Zealand Daylight Time');

-- Add comment
COMMENT ON TABLE timezones IS 'Reference data for timezones (IANA timezone database)';
