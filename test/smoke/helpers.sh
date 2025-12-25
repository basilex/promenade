#!/bin/bash

# Smoke Test Helper Functions

# Get script directory (absolute path)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load config if not already loaded
if [ -z "$API_BASE" ]; then
    source "$SCRIPT_DIR/config.sh"
fi

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_section() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# HTTP request wrapper
http_get() {
    local url="$1"
    local expected_status="${2:-200}"
    local headers="${3:-}"
    
    local response
    local http_code
    
    if [ -n "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" -X GET "$url" -H "$headers" --max-time "$CURL_MAX_TIME")
    else
        response=$(curl -s -w "\n%{http_code}" -X GET "$url" --max-time "$CURL_MAX_TIME")
    fi
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" != "$expected_status" ]; then
        print_error "GET $url failed: expected $expected_status, got $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
    
    echo "$body"
    return 0
}

http_post() {
    local url="$1"
    local data="$2"
    local expected_status="${3:-201}"
    local headers="${4:-}"
    
    local response
    local http_code
    
    if [ -n "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
            -H "Content-Type: application/json" \
            -H "$headers" \
            -d "$data" \
            --max-time "$CURL_MAX_TIME")
    else
        response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time "$CURL_MAX_TIME")
    fi
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" != "$expected_status" ]; then
        print_error "POST $url failed: expected $expected_status, got $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
    
    echo "$body"
    return 0
}

http_put() {
    local url="$1"
    local data="$2"
    local expected_status="${3:-200}"
    local headers="${4:-}"
    
    local response
    local http_code
    
    if [ -n "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" -X PUT "$url" \
            -H "Content-Type: application/json" \
            -H "$headers" \
            -d "$data" \
            --max-time "$CURL_MAX_TIME")
    else
        response=$(curl -s -w "\n%{http_code}" -X PUT "$url" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time "$CURL_MAX_TIME")
    fi
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" != "$expected_status" ]; then
        print_error "PUT $url failed: expected $expected_status, got $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
    
    echo "$body"
    return 0
}

http_delete() {
    local url="$1"
    local expected_status="${2:-200}"
    local headers="${3:-}"
    
    local response
    local http_code
    
    if [ -n "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" -X DELETE "$url" \
            -H "$headers" \
            --max-time "$CURL_MAX_TIME")
    else
        response=$(curl -s -w "\n%{http_code}" -X DELETE "$url" \
            --max-time "$CURL_MAX_TIME")
    fi
    
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" != "$expected_status" ]; then
        print_error "DELETE $url failed: expected $expected_status, got $http_code"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
    
    echo "$body"
    return 0
}

# JSON assertions
assert_json_field() {
    local json="$1"
    local field="$2"
    local expected="$3"
    
    local actual=$(echo "$json" | jq -r "$field")
    
    if [ "$actual" != "$expected" ]; then
        print_error "Assertion failed: $field expected '$expected', got '$actual'"
        return 1
    fi
    
    return 0
}

assert_json_field_exists() {
    local json="$1"
    local field="$2"
    
    local value=$(echo "$json" | jq -r "$field")
    
    if [ "$value" == "null" ] || [ -z "$value" ]; then
        print_error "Assertion failed: field $field does not exist or is null"
        return 1
    fi
    
    return 0
}

# Wait for API to be ready
wait_for_api() {
    local max_attempts=30
    local attempt=0
    
    print_info "Waiting for API to be ready at $API_URL..."
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s -f "$API_BASE/health" > /dev/null 2>&1; then
            print_success "API is ready!"
            return 0
        fi
        
        attempt=$((attempt + 1))
        sleep 1
    done
    
    print_error "API did not become ready after $max_attempts seconds"
    return 1
}

# Check if command exists
check_command() {
    local cmd="$1"
    if ! command -v "$cmd" &> /dev/null; then
        print_error "Required command '$cmd' not found. Please install it first."
        return 1
    fi
    return 0
}

# Cleanup function
cleanup() {
    print_info "Cleaning up test data..."
    
    # Delete test post if created
    if [ -n "$POST_ID" ] && [ -n "$ACCESS_TOKEN" ]; then
        http_delete "$API_BASE/posts/$POST_ID" 200 "Authorization: Bearer $ACCESS_TOKEN" > /dev/null 2>&1 || true
    fi
    
    # Add more cleanup as needed
}
