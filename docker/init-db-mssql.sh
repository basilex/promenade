#!/bin/bash
# MS SQL Server Database Initialization Script
# Creates promenade_dev and promenade_test databases

# Wait for SQL Server to be ready
sleep 30s

echo "Initializing MS SQL Server databases..."

# Create databases using sqlcmd
/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P YourStrong@Passw0rd -Q "
IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'promenade_dev')
BEGIN
    CREATE DATABASE promenade_dev;
    PRINT 'Database promenade_dev created successfully';
END
ELSE
BEGIN
    PRINT 'Database promenade_dev already exists';
END

IF NOT EXISTS (SELECT name FROM sys.databases WHERE name = 'promenade_test')
BEGIN
    CREATE DATABASE promenade_test;
    PRINT 'Database promenade_test created successfully';
END
ELSE
BEGIN
    PRINT 'Database promenade_test already exists';
END
"

echo "MS SQL Server databases initialized"
