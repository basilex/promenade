#!/bin/bash

# Smoke Test Configuration

# API Base URL (override with PROMENADE_API_URL env var)
API_URL="${PROMENADE_API_URL:-http://localhost:8081}"
API_BASE="${API_URL}/api/v1"

# Test Credentials
TEST_EMAIL="smoke_test_$(date +%s)@example.com"
TEST_PASSWORD="SmokeTest123!"
TEST_NAME="Smoke Test User"

# Test Data
TEST_POST_TITLE="Smoke Test Post $(date +%s)"
TEST_POST_CONTENT="This is a smoke test post created at $(date)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# HTTP status codes
HTTP_OK=200
HTTP_CREATED=201
HTTP_NO_CONTENT=204
HTTP_BAD_REQUEST=400
HTTP_UNAUTHORIZED=401
HTTP_NOT_FOUND=404

# Timeouts
CURL_TIMEOUT=10
CURL_MAX_TIME=30

# Global variables for test state
ACCESS_TOKEN=""
USER_ID=""
POST_ID=""
PROFILE_ID=""
