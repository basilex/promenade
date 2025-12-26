#!/bin/bash

# Smoke Test: Workflows Module (Commercial - License Required)

source "$(dirname "$0")/../helpers.sh"

test_workflows() {
    print_section "Workflows Module"
    
    # Need authentication first
    if [ -z "$ACCESS_TOKEN" ]; then
        print_info "Authenticating first..."
        source "$(dirname "$0")/02_auth.sh"
        if [ $? -ne 0 ]; then
            print_error "Authentication failed, cannot test workflows"
            return 1
        fi
    fi
    
    # 1. Check license validation (commercial module)
    print_info "Testing license validation..."
    local health_response=$(http_get "$API_BASE/workflows/health" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Workflows health check failed - may need valid license"
        print_info "Set WORKFLOWS_LICENSE_KEY environment variable or license_required: false in config"
        return 1
    fi
    
    assert_json_field "$health_response" ".success" "true" || return 1
    print_success "Workflows module loaded successfully"
    
    # 2. Create workflow definition
    print_info "Testing POST /workflows/definitions..."
    local definition_data=$(cat <<EOF
{
    "name": "smoke_test_workflow",
    "display_name": "Smoke Test Workflow",
    "description": "Workflow created by smoke tests",
    "category": "testing",
    "schema": {
        "initial_state": "pending",
        "states": [
            {"name": "pending", "type": "activity", "is_final": false},
            {"name": "processing", "type": "activity", "is_final": false},
            {"name": "completed", "type": "final", "is_final": true}
        ],
        "transitions": [
            {"from": "pending", "to": "processing", "event": "start"},
            {"from": "processing", "to": "completed", "event": "finish"}
        ]
    }
}
EOF
)
    
    local create_response=$(http_post "$API_BASE/workflows/definitions" "$definition_data" 201 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Create workflow definition failed"
        return 1
    fi
    
    assert_json_field "$create_response" ".success" "true" || return 1
    assert_json_field_exists "$create_response" ".data.id" || return 1
    assert_json_field "$create_response" ".data.name" "smoke_test_workflow" || return 1
    assert_json_field "$create_response" ".data.status" "draft" || return 1
    
    local definition_id=$(echo "$create_response" | jq -r '.data.id')
    print_success "Workflow definition created (ID: $definition_id)"
    
    # 3. Get workflow definition
    print_info "Testing GET /workflows/definitions/$definition_id..."
    local get_response=$(http_get "$API_BASE/workflows/definitions/$definition_id" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get workflow definition failed"
        return 1
    fi
    
    assert_json_field "$get_response" ".success" "true" || return 1
    assert_json_field "$get_response" ".data.id" "$definition_id" || return 1
    
    print_success "Workflow definition retrieved successfully"
    
    # 4. List workflow definitions
    print_info "Testing GET /workflows/definitions..."
    local list_response=$(http_get "$API_BASE/workflows/definitions?page=1&page_size=10" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "List workflow definitions failed"
        return 1
    fi
    
    assert_json_field "$list_response" ".success" "true" || return 1
    assert_json_field_exists "$list_response" ".items" || return 1
    
    print_success "Workflow definitions listed successfully"
    
    # 5. Activate workflow definition
    print_info "Testing PUT /workflows/definitions/$definition_id/activate..."
    local activate_response=$(http_put "$API_BASE/workflows/definitions/$definition_id/activate" "{}" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Activate workflow definition failed"
        return 1
    fi
    
    assert_json_field "$activate_response" ".success" "true" || return 1
    assert_json_field "$activate_response" ".data.status" "active" || return 1
    
    print_success "Workflow definition activated"
    
    # 6. Start workflow instance
    print_info "Testing POST /workflows/instances..."
    local instance_data=$(cat <<EOF
{
    "definition_id": "$definition_id",
    "input": {
        "test_data": "smoke test",
        "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    },
    "priority": "normal",
    "external_reference": "SMOKE-TEST-001"
}
EOF
)
    
    local start_response=$(http_post "$API_BASE/workflows/instances" "$instance_data" 201 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Start workflow instance failed"
        return 1
    fi
    
    assert_json_field "$start_response" ".success" "true" || return 1
    assert_json_field_exists "$start_response" ".data.id" || return 1
    assert_json_field "$start_response" ".data.status" "running" || return 1
    assert_json_field "$start_response" ".data.current_state" "pending" || return 1
    
    local instance_id=$(echo "$start_response" | jq -r '.data.id')
    print_success "Workflow instance started (ID: $instance_id)"
    
    # 7. Get workflow instance
    print_info "Testing GET /workflows/instances/$instance_id..."
    local get_instance_response=$(http_get "$API_BASE/workflows/instances/$instance_id" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Get workflow instance failed"
        return 1
    fi
    
    assert_json_field "$get_instance_response" ".success" "true" || return 1
    assert_json_field "$get_instance_response" ".data.id" "$instance_id" || return 1
    
    print_success "Workflow instance retrieved successfully"
    
    # 8. List workflow instances
    print_info "Testing GET /workflows/instances..."
    local list_instances_response=$(http_get "$API_BASE/workflows/instances?page=1&page_size=10" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "List workflow instances failed"
        return 1
    fi
    
    assert_json_field "$list_instances_response" ".success" "true" || return 1
    assert_json_field_exists "$list_instances_response" ".items" || return 1
    
    print_success "Workflow instances listed successfully"
    
    # 9. Pause workflow instance
    print_info "Testing PUT /workflows/instances/$instance_id/pause..."
    local pause_response=$(http_put "$API_BASE/workflows/instances/$instance_id/pause" "{}" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Pause workflow instance failed"
        return 1
    fi
    
    assert_json_field "$pause_response" ".success" "true" || return 1
    assert_json_field "$pause_response" ".data.status" "paused" || return 1
    
    print_success "Workflow instance paused"
    
    # 10. Resume workflow instance
    print_info "Testing PUT /workflows/instances/$instance_id/resume..."
    local resume_response=$(http_put "$API_BASE/workflows/instances/$instance_id/resume" "{}" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Resume workflow instance failed"
        return 1
    fi
    
    assert_json_field "$resume_response" ".success" "true" || return 1
    assert_json_field "$resume_response" ".data.status" "running" || return 1
    
    print_success "Workflow instance resumed"
    
    # 11. Cancel workflow instance
    print_info "Testing PUT /workflows/instances/$instance_id/cancel..."
    local cancel_data=$(cat <<EOF
{
    "reason": "Smoke test cancellation"
}
EOF
)
    
    local cancel_response=$(http_put "$API_BASE/workflows/instances/$instance_id/cancel" "$cancel_data" 200 "Authorization: Bearer $ACCESS_TOKEN")
    if [ $? -ne 0 ]; then
        print_error "Cancel workflow instance failed"
        return 1
    fi
    
    assert_json_field "$cancel_response" ".success" "true" || return 1
    assert_json_field "$cancel_response" ".data.status" "cancelled" || return 1
    
    print_success "Workflow instance cancelled"
    
    print_section_result "Workflows Module" "PASSED"
    return 0
}

# Run tests if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    test_workflows
fi
