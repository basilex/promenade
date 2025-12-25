#!/bin/bash

# Smoke Test: Notifications Module

source "$(dirname "$0")/../helpers.sh"

test_notifications() {
    print_section "Notifications Module"
    
    # Need authentication first
    if [ -z "$ACCESS_TOKEN" ]; then
        print_info "Authenticating first..."
        source "$(dirname "$0")/02_auth.sh"
        if [ $? -ne 0 ]; then
            print_error "Authentication failed, cannot test notifications"
            return 1
        fi
    fi
    
    # 1. Get user preferences (should auto-create if not exist)
    print_info "Testing GET /notifications/preferences..."
    local prefs_response=$(http_get "$API_BASE/notifications/preferences" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get preferences failed"
        return 1
    fi
    
    assert_json_field "$prefs_response" ".success" "true" || return 1
    assert_json_field_exists "$prefs_response" ".data.id" || return 1
    assert_json_field_exists "$prefs_response" ".data.user_id" || return 1
    assert_json_field "$prefs_response" ".data.email_enabled" "true" || return 1
    
    print_success "User preferences retrieved successfully"
    
    # 2. Update preferences
    print_info "Testing PUT /notifications/preferences..."
    local update_data=$(cat <<EOF
{
    "marketing_enabled": true,
    "sms_enabled": false,
    "timezone": "Europe/Kiev"
}
EOF
)
    
    local update_response=$(http_put "$API_BASE/notifications/preferences" "$update_data" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Update preferences failed"
        return 1
    fi
    
    assert_json_field "$update_response" ".success" "true" || return 1
    assert_json_field "$update_response" ".data.marketing_enabled" "true" || return 1
    assert_json_field "$update_response" ".data.timezone" "Europe/Kiev" || return 1
    
    print_success "Preferences updated successfully"
    
    # 3. Send notification
    print_info "Testing POST /notifications..."
    local notif_data=$(cat <<EOF
{
    "type": "product",
    "channel": "in_app",
    "subject": "Smoke Test Notification",
    "content": "This is a test notification from smoke tests",
    "data": {
        "test": true,
        "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    }
}
EOF
)
    
    local create_response=$(http_post "$API_BASE/notifications" "$notif_data" 201 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Create notification failed"
        return 1
    fi
    
    assert_json_field "$create_response" ".success" "true" || return 1
    assert_json_field_exists "$create_response" ".data.id" || return 1
    
    local notification_id=$(echo "$create_response" | jq -r '.data.id')
    print_success "Notification created successfully (ID: $notification_id)"
    
    # 4. List notifications
    print_info "Testing GET /notifications..."
    local list_response=$(http_get "$API_BASE/notifications?page=1&page_size=10" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "List notifications failed"
        return 1
    fi
    
    assert_json_field "$list_response" ".success" "true" || return 1
    assert_json_field_exists "$list_response" ".items" || return 1
    
    print_success "Notifications listed successfully"
    
    # 5. Get specific notification
    print_info "Testing GET /notifications/$notification_id..."
    local get_response=$(http_get "$API_BASE/notifications/$notification_id" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get notification failed"
        return 1
    fi
    
    assert_json_field "$get_response" ".success" "true" || return 1
    assert_json_field "$get_response" ".data.id" "$notification_id" || return 1
    
    print_success "Notification retrieved successfully"
    
    # 6. Get unread count
    print_info "Testing GET /notifications/unread-count..."
    local count_response=$(http_get "$API_BASE/notifications/unread-count" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get unread count failed"
        return 1
    fi
    
    assert_json_field "$count_response" ".success" "true" || return 1
    assert_json_field_exists "$count_response" ".data.count" || return 1
    
    print_success "Unread count retrieved successfully"
    
    # 7. Mark notification as opened
    print_info "Testing POST /notifications/$notification_id/opened..."
    local opened_response=$(http_post "$API_BASE/notifications/$notification_id/opened" "" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Mark as opened failed"
        return 1
    fi
    
    assert_json_field "$opened_response" ".success" "true" || return 1
    
    print_success "Notification marked as opened"
    
    print_success "All notifications tests passed"
    return 0
}

# Run test
test_notifications
exit $?
