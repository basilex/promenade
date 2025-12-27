#!/bin/bash

# Test script for Contact API endpoints
# Prerequisites: 
# - Docker running with PostgreSQL
# - API server started: make dev

BASE_URL="http://localhost:8081/api/v1"
USER_ID="01234567-89ab-cdef-0123-456789abcdef"  # UUID v7 format

echo "=========================================="
echo "Contact API Test Suite"
echo "=========================================="
echo ""

# 1. Health Check
echo "1. Health Check"
curl -s "${BASE_URL%/api/v1}/health" | jq .
echo -e "\n"

# 2. API Info
echo "2. API Info"
curl -s "$BASE_URL" | jq .
echo -e "\n"

# 3. Create Email Contact
echo "3. Create Email Contact"
curl -s -X POST "$BASE_URL/identity/contacts?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "label": "Work",
    "email": "john.doe@company.com"
  }' | jq .
echo -e "\n"

# 4. Create Phone Contact
echo "4. Create Phone Contact"
curl -s -X POST "$BASE_URL/identity/contacts?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "phone",
    "label": "Mobile",
    "phone": "+1234567890"
  }' | jq .
echo -e "\n"

# 5. Create Address Contact
echo "5. Create Address Contact"
curl -s -X POST "$BASE_URL/identity/contacts?user_id=$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "address",
    "label": "Home",
    "address": {
      "street": "123 Main St",
      "city": "New York",
      "country": "USA",
      "postal_code": "10001"
    }
  }' | jq .
echo -e "\n"

# 6. List All Contacts for User
echo "6. List All Contacts for User"
curl -s "$BASE_URL/identity/contacts?user_id=$USER_ID" | jq .
echo -e "\n"

# 7. Get Contact by ID (replace with actual ID from create response)
# CONTACT_ID="<paste-id-here>"
# echo "7. Get Contact by ID"
# curl -s "$BASE_URL/identity/contacts/$CONTACT_ID" | jq .
# echo -e "\n"

# 8. Update Contact (replace with actual ID)
# echo "8. Update Contact"
# curl -s -X PUT "$BASE_URL/identity/contacts/$CONTACT_ID" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "label": "Personal",
#     "is_public": true
#   }' | jq .
# echo -e "\n"

# 9. Verify Contact (replace with actual ID)
# echo "9. Verify Contact"
# curl -s -X PUT "$BASE_URL/identity/contacts/$CONTACT_ID/verify" | jq .
# echo -e "\n"

# 10. Set as Primary (replace with actual ID)
# echo "10. Set as Primary"
# curl -s -X PUT "$BASE_URL/identity/contacts/$CONTACT_ID/primary" | jq .
# echo -e "\n"

# 11. Delete Contact (replace with actual ID)
# echo "11. Delete Contact"
# curl -s -X DELETE "$BASE_URL/identity/contacts/$CONTACT_ID" | jq .
# echo -e "\n"

echo "=========================================="
echo "Test Complete!"
echo "=========================================="
echo ""
echo "NOTE: Uncomment sections 7-11 and replace CONTACT_ID"
echo "      with actual ID from create response to test"
echo "      Get/Update/Verify/Primary/Delete operations."
