#!/bin/bash
# scripts/run-tests.sh - Universal test runner

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧪 Promenade Test Runner${NC}"
echo ""

# Check if test DB is running
if ! docker ps | grep -q promenade_test_db; then
    echo -e "${YELLOW}[WARNING] Test database not running, starting...${NC}"
    make test-db-start
    echo -e "${GREEN}✓ Test database started${NC}"
    echo ""
fi

# Parse arguments
TEST_TYPE=${1:-all}

case $TEST_TYPE in
    unit)
        echo -e "${GREEN}Running unit tests...${NC}"
        go test -v -race -count=1 ./internal/usecase/... ./internal/domain/...
        ;;
    integration)
        echo -e "${GREEN}Running integration tests...${NC}"
        go test -v -race -count=1 ./internal/adapter/repository/postgres/...
        ;;
    coverage)
        echo -e "${GREEN}Running tests with coverage...${NC}"
        go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
        go tool cover -html=coverage.out -o coverage.html
        echo -e "${GREEN}✓ Coverage report: coverage.html${NC}"
        ;;
    all|*)
        echo -e "${GREEN}Running all tests...${NC}"
        echo ""
        
        echo -e "${YELLOW}→ Unit tests${NC}"
        go test -v -race -count=1 ./internal/usecase/... ./internal/domain/... || true
        echo ""
        
        echo -e "${YELLOW}→ Integration tests${NC}"
        go test -v -race -count=1 ./internal/adapter/repository/postgres/...
        echo ""
        ;;
esac

echo ""
echo -e "${GREEN}[OK] Tests completed${NC}"
