#!/bin/bash

# Promenade Stress Tests Runner
# Uses wrk to perform load testing on API endpoints

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_URL="${PROMENADE_API_URL:-http://localhost:8081}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Print header
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}   Promenade API - Stress Tests${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  ${BLUE}API:${NC}    $API_URL"
echo -e "  ${BLUE}Tool:${NC}   wrk (HTTP benchmarking)"
echo -e "  ${BLUE}Date:${NC}   $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# Check if wrk is installed
if ! command -v wrk &> /dev/null; then
    echo -e "${RED}✗${NC} wrk is not installed"
    echo ""
    echo "Install wrk:"
    echo "  macOS:   brew install wrk"
    echo "  Linux:   sudo apt-get install wrk"
    echo "  Source:  https://github.com/wg/wrk"
    echo ""
    exit 1
fi

echo -e "${GREEN}✓${NC} wrk found: $(wrk --version 2>&1 | head -n 1)"

# Check if API is running
echo -e "${BLUE}ℹ${NC} Checking if API is running..."
if ! curl -sf "$API_URL/api/v1/health" > /dev/null 2>&1; then
    echo -e "${RED}✗${NC} API is not responding at $API_URL"
    echo ""
    echo "Start the API first:"
    echo "  make dev"
    echo ""
    exit 1
fi

echo -e "${GREEN}✓${NC} API is ready"
echo ""

# Run tests
run_test() {
    local name="$1"
    local cmd="$2"
    
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    eval "$cmd"
    
    echo ""
    sleep 2
}

# Test 1: Health Check (Baseline)
run_test "Test 1: Health Check (Baseline)" \
    "wrk -t4 -c100 -d10s $API_URL/api/v1/health"

# Test 2: Authentication Load
run_test "Test 2: Authentication Load" \
    "wrk -t4 -c50 -d10s -s $SCRIPT_DIR/scenarios/auth.lua $API_URL"

# Test 3: Light Load (Reference Data)
run_test "Test 3: Light Load (Countries)" \
    "wrk -t2 -c20 -d10s $API_URL/api/v1/countries"

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${GREEN}✓${NC} Stress tests completed"
echo ""
echo "Next steps:"
echo "  1. Review results above"
echo "  2. Check for errors or slow endpoints"
echo "  3. Run individual scenarios: wrk -t4 -c100 -d30s -s scenarios/posts.lua $API_URL"
echo "  4. Monitor resources: make monitor"
echo ""
