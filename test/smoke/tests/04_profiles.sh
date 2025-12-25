#!/bin/bash

# Smoke Test: Profiles

source "$(dirname "$0")/../helpers.sh"

test_profiles() {
    print_section "Profiles"
    
    # Need authentication first
    if [ -z "$ACCESS_TOKEN" ]; then
        print_info "Authenticating first..."
        source "$(dirname "$0")/02_auth.sh"
        if [ $? -ne 0 ]; then
            print_error "Authentication failed, cannot test profiles"
            return 1
        fi
    fi
    
    # 1. Create profile
    print_info "Testing POST /profiles..."
    local profile_data=$(cat <<EOF
{
    "bio": "This is a smoke test bio",
    "location": "Test City",
    "website": "https://example.com",
    "privacy": "public"
}
EOF
)
    
    local create_response=$(http_post "$API_BASE/profiles" "$profile_data" 201 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Create profile failed"
        return 1
    fi
    
    assert_json_field "$create_response" ".status" "success" || return 1
    assert_json_field_exists "$create_response" ".data.id" || return 1
    
    PROFILE_ID=$(echo "$create_response" | jq -r '.data.id')
    print_success "Profile created successfully (ID: $PROFILE_ID)"
    
    # 2. Get profile
    print_info "Testing GET /profiles/me..."
    local get_response=$(http_get "$API_BASE/profiles/me" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get profile failed"
        return 1
    fi
    
    assert_json_field "$get_response" ".status" "success" || return 1
    assert_json_field_exists "$get_response" ".data.id" || return 1
    
    print_success "Get profile passed"
    
    # 3. Update profile
    print_info "Testing PUT /profiles/me..."
    local update_data=$(cat <<EOF
{
    "bio": "Updated smoke test bio",
    "location": "Updated Test City",
    "website": "https://updated.example.com",
    "privacy": "private"
}
EOF
)
    
    local update_response=$(http_put "$API_BASE/profiles/me" "$update_data" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Update profile failed"
        return 1
    fi
    
    assert_json_field "$update_response" ".status" "success" || return 1
    assert_json_field "$update_response" ".data.bio" "Updated smoke test bio" || return 1
    
    print_success "Update profile passed"
    
    print_success "Profiles tests completed successfully"
    return 0
}

# Run test
test_profiles
exit $?
