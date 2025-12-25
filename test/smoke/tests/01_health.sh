#!/bin/bash

# Smoke Test: Health Check

source "$(dirname "$0")/../helpers.sh"

test_health_check() {
    print_section "Health Check"
    
    # Test health endpoint
    print_info "Testing GET /health..."
    local response=$(http_get "$API_BASE/health" 200)
    
    if [ $? -ne 0 ]; then
        print_error "Health check failed"
        return 1
    fi
    
    # Verify response structure
    assert_json_field "$response" ".success" "true" || return 1
    assert_json_field_exists "$response" ".data.status" || return 1
    assert_json_field_exists "$response" ".data.service" || return 1
    assert_json_field_exists "$response" ".timestamp" || return 1
    
    print_success "Health check passed"
    return 0
}

# Run test
test_health_check
exit $?
