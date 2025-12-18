#!/bin/bash
set -e

# Create databases if they don't exist
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE promenade_dev'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'promenade_dev')\gexec

    SELECT 'CREATE DATABASE promenade_test'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'promenade_test')\gexec
EOSQL

echo "Databases created successfully"
