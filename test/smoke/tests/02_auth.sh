#!/bin/bash

# Smoke Test: Authentication Flow

source "$(dirname "$0")/../helpers.sh"

test_auth_flow() {
    print_section "Authentication Flow"
    
    # 1. Register new user
    print_info "Testing POST /auth/register..."
    local register_data=$(cat <<EOF
{
    "email": "$TEST_EMAIL",
    "password": "$TEST_PASSWORD",
    "name": "$TEST_NAME"
}
EOF
)
    
    local register_response=$(http_post "$API_BASE/auth/register" "$register_data" 201)
    if [ $? -ne 0 ]; then
        print_error "User registration failed"
        return 1
    fi
    
    assert_json_field "$register_response" ".status" "success" || return 1
    assert_json_field_exists "$register_response" ".data.user.id" || return 1
    assert_json_field_exists "$register_response" ".data.access_token" || return 1
    assert_json_field_exists "$register_response" ".data.refresh_token" || return 1
    
    USER_ID=$(echo "$register_response" | jq -r '.data.user.id')
    ACCESS_TOKEN=$(echo "$register_response" | jq -r '.data.access_token')
    
    print_success "User registered successfully (ID: $USER_ID)"
    
    # 2. Login with credentials
    print_info "Testing POST /auth/login..."
    local login_data=$(cat <<EOF
{
    "email": "$TEST_EMAIL",
    "password": "$TEST_PASSWORD"
}
EOF
)
    
    local login_response=$(http_post "$API_BASE/auth/login" "$login_data" 200)
    if [ $? -ne 0 ]; then
        print_error "User login failed"
        return 1
    fi
    
    assert_json_field "$login_response" ".status" "success" || return 1
    assert_json_field_exists "$login_response" ".data.access_token" || return 1
    
    ACCESS_TOKEN=$(echo "$login_response" | jq -r '.data.access_token')
    print_success "User logged in successfully"
    
    # 3. Get current user (me)
    print_info "Testing GET /auth/me..."
    local me_response=$(http_get "$API_BASE/auth/me" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get current user failed"
        return 1
    fi
    
    assert_json_field "$me_response" ".status" "success" || return 1
    assert_json_field "$me_response" ".data.email" "$TEST_EMAIL" || return 1
    assert_json_field "$me_response" ".data.name" "$TEST_NAME" || return 1
    
    print_success "Get current user passed"
    
    # 4. Logout
    print_info "Testing POST /auth/logout..."
    local logout_response=$(http_post "$API_BASE/auth/logout" "{}" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "User logout failed"
        return 1
    fi
    
    assert_json_field "$logout_response" ".status" "success" || return 1
    print_success "User logged out successfully"
    
    print_success "Authentication flow completed successfully"
    return 0
}

# Run test
test_auth_flow
exit $?
