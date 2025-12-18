#!/usr/bin/env bash

API_URL="http://localhost:8081"
EMAIL="test@example.com"
PASSWORD="SecurePass123"

echo "Testing Promenade API"
echo "========================"
echo ""

# 0. API Info endpoints
echo "Testing API Root (GET /api)..."
curl -s $API_URL/api | jq .
echo ""

echo "Testing API v1 Info (GET /api/v1)..."
curl -s $API_URL/api/v1 | jq .
echo ""

echo "Testing API v2 Info (GET /api/v2)..."
curl -s $API_URL/api/v2 | jq .
echo ""

# 1. Health Check
echo "Testing Health Check..."
curl -s $API_URL/api/v1/health | jq .
echo ""

# 2. Register
echo "Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST $API_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"name\":  \"Test User\",
    \"password\": \"$PASSWORD\"
  }")

echo "$REGISTER_RESPONSE" | jq . 

# Extract token
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.access_token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Registration failed"
  exit 1
fi

echo "Token obtained: ${TOKEN: 0:20}..."
echo ""

# 3. Get Me
echo "Getting current user info..."
curl -s $API_URL/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq . 
echo ""

# 4. Create Product
echo "Creating product..."
PRODUCT_RESPONSE=$(curl -s -X POST $API_URL/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MacBook Pro M4"
  }')

echo "$PRODUCT_RESPONSE" | jq . 

PRODUCT_ID=$(echo "$PRODUCT_RESPONSE" | jq -r '.data.id')
echo "Product ID: $PRODUCT_ID"
echo ""

# 5. List Products
echo "Listing products..."
curl -s $API_URL/api/v1/products \
  -H "Authorization: Bearer $TOKEN" | jq . 
echo ""

# 6. Get Product by ID
echo "Getting product by ID..."
curl -s "$API_URL/api/v1/products/$PRODUCT_ID" \
  -H "Authorization:  Bearer $TOKEN" | jq . 
echo ""

# 7. Update Product
echo "Updating product..."
curl -s -X PUT "$API_URL/api/v1/products/$PRODUCT_ID" \
  -H "Authorization:  Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MacBook Pro M4 Max",
    "active": true
  }' | jq .
echo ""

# 8. Logout
echo "Logging out..."
curl -s -X POST $API_URL/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

echo "All tests completed!"
