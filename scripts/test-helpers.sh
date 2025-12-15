#!/usr/bin/env bash

# Load helpers
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/helpers.sh"

echo "========================================="
echo "Testing Helper Functions"
echo "========================================="
echo ""

# Test cases
test_cases=(
    "Product"
    "ProductCategory"
    "UserProfile"
    "OrderItem"
    "HTTPServer"
    "APIKey"
)

echo "TEST 1: to_snake_case"
echo "---------------------"
for test in "${test_cases[@]}"; do
    result=$(to_snake_case "$test")
    echo "$test -> $result"
done

echo ""
echo "TEST 2: to_lower"
echo "---------------------"
for test in "${test_cases[@]}"; do
    result=$(to_lower "$test")
    echo "$test -> $result"
done

echo ""
echo "TEST 3: to_plural"
echo "---------------------"
words=("product" "category" "box" "knife" "leaf" "baby" "hero" "photo")
for word in "${words[@]}"; do
    result=$(to_plural "$word")
    echo "$word -> $result"
done

echo ""
echo "TEST 4: get_module_name"
echo "---------------------"
module=$(get_module_name)
echo "Module: $module"

echo ""
echo "========================================="
echo "All Tests Complete"
echo "========================================="
