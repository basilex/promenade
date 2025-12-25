#!/bin/bash

# Smoke Test: Analytics

source "$(dirname "$0")/../helpers.sh"

test_analytics() {
    print_section "Analytics"
    
    # Need authentication first
    if [ -z "$ACCESS_TOKEN" ]; then
        print_info "Authenticating first..."
        source "$(dirname "$0")/02_auth.sh"
        if [ $? -ne 0 ]; then
            print_error "Authentication failed, cannot test analytics"
            return 1
        fi
    fi
    
    # 1. Create metric
    print_info "Testing POST /analytics/metrics..."
    local metric_data=$(cat <<EOF
{
    "name": "smoke_test_metric",
    "value": 42.5,
    "tags": {
        "test": "smoke",
        "environment": "test"
    }
}
EOF
)
    
    local create_response=$(http_post "$API_BASE/analytics/metrics" "$metric_data" 201 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Create metric failed"
        return 1
    fi
    
    assert_json_field "$create_response" ".success" "true" || return 1
    assert_json_field_exists "$create_response" ".data.id" || return 1
    
    print_success "Metric created successfully"
    
    # 2. List metrics
    print_info "Testing GET /analytics/metrics..."
    local list_response=$(http_get "$API_BASE/analytics/metrics?page=1&page_size=10" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "List metrics failed"
        return 1
    fi
    
    assert_json_field "$list_response" ".success" "true" || return 1
    assert_json_field_exists "$list_response" ".data" || return 1
    
    print_success "List metrics passed"
    
    print_success "Analytics tests completed successfully"
    return 0
}

# Run test
test_analytics
exit $?
