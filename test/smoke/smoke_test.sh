#!/bin/bash

# Promenade Smoke Tests - Main Runner
# Runs smoke tests against running API instance

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Load configuration and helpers
source config.sh
source helpers.sh

# Print header
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}   Promenade API - Smoke Tests${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  ${BLUE}API:${NC}         $API_URL"
echo -e "  ${BLUE}Environment:${NC} ${PROMENADE_ENV:-dev}"
echo -e "  ${BLUE}Date:${NC}        $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# Check required tools
print_info "Checking required tools..."
check_command "curl" || exit 1
check_command "jq" || exit 1
print_success "All required tools found"

# Wait for API to be ready
wait_for_api || exit 1

# Track test results
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

# Trap cleanup on exit
trap cleanup EXIT

# Run tests
run_test() {
    local test_file="$1"
    local test_name=$(basename "$test_file" .sh)
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    print_info "Running test: $test_name..."
    
    if bash "$test_file"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        print_success "Test passed: $test_name"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        FAILED_TESTS+=("$test_name")
        print_error "Test failed: $test_name"
    fi
    
    echo ""
}

# Run all tests in order
print_section "Running Smoke Tests"

run_test "tests/01_health.sh"
run_test "tests/02_auth.sh"
run_test "tests/03_posts.sh"
run_test "tests/04_profiles.sh"
run_test "tests/05_analytics.sh"
run_test "tests/06_notifications.sh"

# Print summary
print_section "Test Summary"

echo -e "  ${BLUE}Total tests:${NC}  $TESTS_RUN"
echo -e "  ${GREEN}Passed:${NC}       $TESTS_PASSED"

if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "  ${RED}Failed:${NC}       $TESTS_FAILED"
    echo ""
    echo -e "${RED}Failed tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗${NC} $test"
    done
    echo ""
    exit 1
else
    echo -e "  ${RED}Failed:${NC}       0"
    echo ""
    print_success "All smoke tests passed!"
    echo ""
    exit 0
fi
