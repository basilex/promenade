-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create UUID v7 generation function (time-ordered UUIDs for better B-tree performance)
CREATE OR REPLACE FUNCTION uuid_v7()
RETURNS UUID
AS $$
DECLARE
    rand_a BYTEA;
    rand_b BYTEA;
    uuid_bytes BYTEA;
    unix_ts_ms BIGINT;
BEGIN
    rand_a := gen_random_bytes(2);
    rand_b := gen_random_bytes(8);

    unix_ts_ms := (EXTRACT(EPOCH FROM clock_timestamp()) * 1000)::BIGINT;
    
    uuid_bytes := 
        substring(int8send(unix_ts_ms) from 3 for 6) ||
        set_byte(rand_a, 0, (b'0111' || get_byte(rand_a, 0)::bit(4))::bit(8)::int) ||
        set_byte(rand_b, 0, (b'10' || substring(get_byte(rand_b, 0)::bit(8) from 3 for 6))::bit(8)::int) ||
        substring(rand_b from 2 for 7);
    
    RETURN substring(encode(uuid_bytes, 'hex') from 1 for 32)::UUID;
END;
$$ LANGUAGE plpgsql VOLATILE;

COMMENT ON FUNCTION uuid_v7() IS 
    'Generates UUID v7 (time-ordered) per RFC 9562 for better database performance';

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION tfn_entity_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

COMMENT ON FUNCTION tfn_entity_updated_at() IS 
'Trigger function to automatically update updated_at timestamp on row modification';
