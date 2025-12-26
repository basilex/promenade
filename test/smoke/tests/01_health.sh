#!/bin/bash

# Smoke Test: Health Check

source "$(dirname "$0")/../helpers.sh"

test_health_check() {
    print_section "Health Check"
    
    # Test core health endpoint
    print_info "Testing GET /health (Core)..."
    local response=$(http_get "$API_BASE/health" 200)
    
    if [ $? -ne 0 ]; then
        print_error "Core health check failed"
        return 1
    fi
    
    # Verify response structure
    assert_json_field "$response" ".success" "true" || return 1
    assert_json_field_exists "$response" ".data.status" || return 1
    assert_json_field_exists "$response" ".data.service" || return 1
    assert_json_field_exists "$response" ".timestamp" || return 1
    
    print_success "Core health check passed"
    
    # Test module health endpoints
    local modules=("posts" "profiles" "analytics" "notifications" "billing")
    
    for module in "${modules[@]}"; do
        print_info "Testing GET /${module}/health..."
        local module_response=$(http_get "$API_BASE/${module}/health" 200)
        
        if [ $? -ne 0 ]; then
            print_error "${module} health check failed"
            return 1
        fi
        
        # Verify module health response
        local status=$(echo "$module_response" | jq -r '.status // .data.status')
        if [ "$status" != "healthy" ] && [ "$status" != "ok" ]; then
            print_error "${module} health status is not healthy: $status"
            return 1
        fi
        
        print_success "${module} health check passed"
    done
    
    print_success "All health checks passed"
    return 0
}

# Run test
test_health_check
exit $?
