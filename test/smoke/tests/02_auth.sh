#!/bin/bash

# Smoke Test: Authentication Flow

source "$(dirname "$0")/../helpers.sh"

test_auth_flow() {
    print_section "Authentication Flow"
    
    # Note: Using existing system user instead of registration
    # (Registration requires email verification before tokens are issued)
    
    # 1. Login with credentials
    print_info "Testing POST /auth/login..."
    local login_data=$(cat <<EOF
{
    "email": "system@promenade.com",
    "password": "passw0rd"
}
EOF
)
    
    local login_response=$(http_post "$API_BASE/auth/login" "$login_data" 200)
    if [ $? -ne 0 ]; then
        print_error "User login failed"
        return 1
    fi
    
    assert_json_field "$login_response" ".success" "true" || return 1
    assert_json_field_exists "$login_response" ".data.access_token" || return 1
    
    ACCESS_TOKEN=$(echo "$login_response" | jq -r '.data.access_token')
    USER_ID=$(echo "$login_response" | jq -r '.data.user.id')
    print_success "User logged in successfully"
    
    # 2. Get current user (me)
    print_info "Testing GET /auth/me..."
    local me_response=$(http_get "$API_BASE/auth/me" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get current user failed"
        return 1
    fi
    
    assert_json_field "$me_response" ".success" "true" || return 1
    assert_json_field "$me_response" ".data.email" "system@promenade.com" || return 1
    assert_json_field "$me_response" ".data.name" "System Administrator" || return 1
    
    print_success "Get current user passed"
    
    # Note: Skipping logout to keep token valid for other tests
    # In real scenario, you would test logout separately
    
    print_success "Authentication flow completed successfully"
    return 0
}

# Run test
test_auth_flow
exit $?
