#!/bin/bash

# Smoke Test: Posts CRUD

source "$(dirname "$0")/../helpers.sh"

test_posts_crud() {
    print_section "Posts CRUD"
    
    # Need authentication first
    if [ -z "$ACCESS_TOKEN" ]; then
        print_info "Authenticating first..."
        source "$(dirname "$0")/02_auth.sh"
        if [ $? -ne 0 ]; then
            print_error "Authentication failed, cannot test posts"
            return 1
        fi
    fi
    
    # 1. Create post
    print_info "Testing POST /posts..."
    local post_data=$(cat <<EOF
{
    "title": "$TEST_POST_TITLE",
    "content": "$TEST_POST_CONTENT",
    "status": "draft"
}
EOF
)
    
    local create_response=$(http_post "$API_BASE/posts" "$post_data" 201 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Create post failed"
        return 1
    fi
    
    assert_json_field "$create_response" ".success" "true" || return 1
    assert_json_field_exists "$create_response" ".data.id" || return 1
    assert_json_field "$create_response" ".data.title" "$TEST_POST_TITLE" || return 1
    
    POST_ID=$(echo "$create_response" | jq -r '.data.id')
    print_success "Post created successfully (ID: $POST_ID)"
    
    # 2. Get post
    print_info "Testing GET /posts/$POST_ID..."
    local get_response=$(http_get "$API_BASE/posts/$POST_ID" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get post failed"
        return 1
    fi
    
    assert_json_field "$get_response" ".success" "true" || return 1
    assert_json_field "$get_response" ".data.id" "$POST_ID" || return 1
    assert_json_field "$get_response" ".data.title" "$TEST_POST_TITLE" || return 1
    
    print_success "Get post passed"
    
    # 3. List posts
    print_info "Testing GET /posts..."
    local list_response=$(http_get "$API_BASE/posts?page=1&page_size=10" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "List posts failed"
        return 1
    fi
    
    assert_json_field "$list_response" ".success" "true" || return 1
    assert_json_field_exists "$list_response" ".data" || return 1
    
    print_success "List posts passed"
    
    # 4. Update post
    print_info "Testing PUT /posts/$POST_ID..."
    local update_data=$(cat <<EOF
{
    "title": "$TEST_POST_TITLE (Updated)",
    "content": "$TEST_POST_CONTENT (Updated)",
    "status": "published"
}
EOF
)
    
    local update_response=$(http_put "$API_BASE/posts/$POST_ID" "$update_data" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Update post failed"
        return 1
    fi
    
    assert_json_field "$update_response" ".success" "true" || return 1
    assert_json_field "$update_response" ".data.title" "$TEST_POST_TITLE (Updated)" || return 1
    
    print_success "Update post passed"
    
    # 5. Delete post
    print_info "Testing DELETE /posts/$POST_ID..."
    local delete_response=$(http_delete "$API_BASE/posts/$POST_ID" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Delete post failed"
        return 1
    fi
    
    assert_json_field "$delete_response" ".success" "true" || return 1
    print_success "Delete post passed"
    
    # Clear POST_ID so cleanup doesn't try to delete again
    POST_ID=""
    
    print_success "Posts CRUD completed successfully"
    return 0
}

# Run test
test_posts_crud
exit $?
