# Quick Start Guide

**Get up and running with Promenade API in 5 minutes** - practical guide with copy-paste curl examples.

---

## Prerequisites

- **Go 1.24+** installed
- **Docker** & **Docker Compose** (for PostgreSQL)
- **curl** (command-line HTTP client)
- **jq** (optional, for pretty JSON formatting)

```bash
# Check prerequisites
go version        # Should be 1.24+
docker --version  # Docker installed
curl --version    # curl available
```

---

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/basilex/promenade.git
cd promenade
```

### 2. Configure Workspace

```bash
# PostgreSQL + Development (default)
make switch-postgres-dev

# Or SQLite + Development (no Docker needed)
make switch-sqlite-dev
```

### 3. Start Application

```bash
# Start PostgreSQL (if using postgres)
make docker-up

# Run migrations
make migrate

# Start API server
make dev
```

**Server starts on**: `http://localhost:8081`

---

## Your First API Request

### Health Check

```bash
curl http://localhost:8081/health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2026-01-05T14:30:00Z",
  "version": "0.1.0",
  "checks": {
    "database": "healthy",
    "redis": "healthy",
    "event_bus": "healthy"
  }
}
```

 **Server is running!** Let's create your first user.

---

## Authentication Flow

Promenade uses **JWT tokens** for authentication. Follow these 3 steps:

### Step 1: Register a User

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC123DEF456GHI789JKL0",
    "email": "john@example.com",
    "name": "John Doe",
    "status": "active",
    "created_at": "2026-01-05T14:31:00Z"
  }
}
```

### Step 2: Login to Get JWT Token

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-01-05T14:46:00Z",
    "token_type": "Bearer",
    "user": {
      "id": "01JGABC123DEF456GHI789JKL0",
      "email": "john@example.com",
      "name": "John Doe"
    }
  }
}
```

### Step 3: Save Your Token

```bash
# Save access token to environment variable
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Or save to file
echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." > token.txt
export TOKEN=$(cat token.txt)
```

**Token Duration**:
- Access Token: **15 minutes** (for API requests)
- Refresh Token: **7 days** (to generate new access tokens)

---

## Your First Customer

### Create Customer (B2C)

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "phone": "+380501234567",
    "type": "individual",
    "status": "lead",
    "tier": "free"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEF789GHI012JKL345MNO6",
    "name": "Alice Smith",
    "email": "alice@example.com",
    "phone": "+380501234567",
    "type": "individual",
    "status": "lead",
    "tier": "free",
    "created_at": "2026-01-05T14:35:00Z"
  }
}
```

### Get Customer Details

```bash
# Save customer ID
export CUSTOMER_ID="01JGDEF789GHI012JKL345MNO6"

# Get customer
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### List All Customers

```bash
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/customers?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Your First Order

### Create Order

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "currency": "USD"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPQR123STU456VWX789YZA0",
    "order_no": "ORD-2026-000001",
    "customer_id": "01JGDEF789GHI012JKL345MNO6",
    "status": "pending",
    "currency": "USD",
    "total_amount": 0,
    "created_at": "2026-01-05T14:40:00Z"
  }
}
```

### Add Line Items

```bash
# Save order ID
export ORDER_ID="01JGPQR123STU456VWX789YZA0"

# Add product to order
curl -X POST "http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/lines" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "01JGPROD123456789ABCDEFGHI",
    "quantity": 2,
    "unit_price": 4999
  }'
```

**Note**: Prices are in **cents** (4999 = $49.99)

### Confirm Order

```bash
curl -X POST "http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/confirm" \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPQR123STU456VWX789YZA0",
    "order_no": "ORD-2026-000001",
    "status": "confirmed",
    "total_amount": 9998,
    "confirmed_at": "2026-01-05T14:42:00Z"
  }
}
```

---

## Common Operations

### Filter & Search

#### Filter Customers by Status

```bash
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/customers?status=lead" \
  -H "Authorization: Bearer $TOKEN"
```

#### Filter Orders by Status

```bash
curl -X GET "http://localhost:8081/api/v1/order-mgmt/orders?status=confirmed" \
  -H "Authorization: Bearer $TOKEN"
```

#### Search Customers by Name

```bash
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/customers?search=Alice" \
  -H "Authorization: Bearer $TOKEN"
```

### Pagination

```bash
# Page 1, 20 items per page
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/customers?page=1&page_size=20" \
  -H "Authorization: Bearer $TOKEN"

# Page 2
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/customers?page=2&page_size=20" \
  -H "Authorization: Bearer $TOKEN"
```

### Update Resources

```bash
# Update customer tier
curl -X PUT "http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/tier" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tier": "pro"
  }'
```

### Delete Resources

```bash
# Soft delete customer
curl -X DELETE "http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Using Swagger UI

**Interactive API Documentation**: http://localhost:8081/api/docs/index.html

1. Open Swagger UI in browser
2. Click **"Authorize"** button (top-right)
3. Enter token: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
4. Click **"Authorize"** and **"Close"**
5. Try endpoints directly from browser

**Benefits**:
- See all 120+ endpoints
- Test requests interactively
- View request/response schemas
- Copy curl commands

---

## Using Postman Collection

**Pre-built Collection**: `postman/Promenade_API.postman_collection.json`

### Import to Postman

1. Open Postman
2. Click **Import** → Select `Promenade_API.postman_collection.json`
3. Import environment: `postman/Development.postman_environment.json`
4. Select **"Development"** environment (top-right)

### Auto-Save Tokens

**Login Request** automatically saves tokens to environment variables:
- `access_token` - Used in all requests
- `refresh_token` - For token refresh

No need to copy-paste tokens manually!

---

## Common Use Cases

### Customer Lifecycle

```bash
# 1. Create lead
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob Johnson","email":"bob@example.com","status":"lead","tier":"free"}'

# 2. Qualify as prospect
curl -X POST "http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/qualify-as-prospect" \
  -H "Authorization: Bearer $TOKEN"

# 3. Convert to customer
curl -X POST "http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/convert-to-customer" \
  -H "Authorization: Bearer $TOKEN"

# 4. Upgrade tier
curl -X PUT "http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/tier" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tier":"pro"}'
```

### Deal Pipeline

```bash
# 1. Create deal
curl -X POST http://localhost:8081/api/v1/customer-mgmt/deals \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "title": "Enterprise License",
    "value": 10000000,
    "currency": "USD",
    "stage": "lead"
  }'

# Save deal ID
export DEAL_ID="01JGDEAL123456789ABCDEFGHI"

# 2. Move to qualified
curl -X POST "http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/move-to-stage" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"stage":"qualified"}'

# 3. Mark as won
curl -X POST "http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/mark-won" \
  -H "Authorization: Bearer $TOKEN"
```

### Invoice & Payment

```bash
# 1. Create invoice
curl -X POST http://localhost:8081/api/v1/billing/invoices \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "amount": 9999,
    "currency": "USD",
    "due_date": "2026-02-05"
  }'

# Save invoice ID
export INVOICE_ID="01JGINV123456789ABCDEFGHI"

# 2. Create payment
curl -X POST http://localhost:8081/api/v1/billing/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "invoice_id": "'"$INVOICE_ID"'",
    "amount": 9999,
    "currency": "USD",
    "method": "credit_card"
  }'

# Save payment ID
export PAYMENT_ID="01JGPAY123456789ABCDEFGHI"

# 3. Process payment
curl -X POST "http://localhost:8081/api/v1/billing/payments/$PAYMENT_ID/process" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Token Refresh

### When Access Token Expires

```bash
# Use refresh token to get new access token
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'"$REFRESH_TOKEN"'"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-01-05T15:01:00Z",
    "token_type": "Bearer"
  }
}
```

### Auto-Refresh Script (Bash)

```bash
#!/bin/bash
# auto-refresh.sh - Automatically refresh token when expired

TOKEN_FILE="token.txt"
REFRESH_TOKEN_FILE="refresh_token.txt"

# Function to refresh token
refresh_token() {
  REFRESH_TOKEN=$(cat $REFRESH_TOKEN_FILE)
  
  RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/identity/auth/refresh \
    -H "Content-Type: application/json" \
    -d '{"refresh_token":"'"$REFRESH_TOKEN"'"}')
  
  NEW_TOKEN=$(echo $RESPONSE | jq -r '.data.access_token')
  NEW_REFRESH=$(echo $RESPONSE | jq -r '.data.refresh_token')
  
  echo $NEW_TOKEN > $TOKEN_FILE
  echo $NEW_REFRESH > $REFRESH_TOKEN_FILE
  
  export TOKEN=$NEW_TOKEN
  echo "Token refreshed!"
}

# Check if token is expired (simple check)
if [ ! -f $TOKEN_FILE ]; then
  echo "No token found. Please login first."
  exit 1
fi

# Try request, refresh if fails
RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null \
  -H "Authorization: Bearer $(cat $TOKEN_FILE)" \
  http://localhost:8081/api/v1/identity/users/me)

if [ "$RESPONSE" = "401" ]; then
  echo "Token expired. Refreshing..."
  refresh_token
else
  echo "Token is valid"
fi
```

---

## Error Handling

### Common Error Responses

#### 400 Bad Request - Validation Error

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email: must be a valid email address"
  }
}
```

#### 401 Unauthorized - Missing/Invalid Token

```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "missing or invalid token"
  }
}
```

#### 403 Forbidden - Insufficient Permissions

```json
{
  "status": "error",
  "error": {
    "code": "FORBIDDEN",
    "message": "admin role required"
  }
}
```

#### 404 Not Found

```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "customer not found"
  }
}
```

#### 429 Too Many Requests - Rate Limit

```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "too many requests, retry after 60 seconds"
  }
}
```

**Rate Limits**:
- Login: **5 requests per minute**
- Register: **3 requests per minute**

---

## Troubleshooting

### Server Won't Start

**Problem**: `make dev` fails

**Solutions**:
```bash
# Check workspace configuration
make workspace

# Check if PostgreSQL is running
make docker-ps

# Check logs
make docker-logs

# Fresh restart
make docker-down
make docker-up
make migrate
make dev
```

### Token Errors

**Problem**: `401 Unauthorized` on every request

**Solutions**:
```bash
# Check token format (must include "Bearer ")
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/v1/identity/users/me

# Refresh token if expired
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"'"$REFRESH_TOKEN"'"}'

# Login again
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"SecurePass123"}'
```

### Database Errors

**Problem**: `database connection failed`

**Solutions**:
```bash
# Check database status
make docker-ps

# Restart database
make docker-down
make docker-up

# Reset database
make db-reset
make migrate
```

### Migration Errors

**Problem**: `migration failed`

**Solutions**:
```bash
# Check migration status
make migrate-status

# Fresh migrations
make db-fresh
```

---

## Next Steps

###  Explore Documentation

- **[API Reference](api-reference.md)** - Complete endpoint documentation
- **[Authentication Guide](../pkg/jwt/README.md)** - JWT & RBAC details
- **[API Versioning](api-versioning.md)** - Deprecation policy
- **[Testing Guide](testing-patterns.md)** - Write tests
- **[Architecture Guide](../concepts/clean-architecture.md)** - DDD principles

###  Advanced Topics

- **Rate Limiting**: [Rate Limiting Guide](rate-limiting.md)
- **CSRF Protection**: [CSRF Guide](csrf-protection.md)
- **Caching**: [Caching Guide](caching.md)
- **Health Checks**: [Health Checks Guide](health-checks.md)

###  Build Your Application

- Create custom aggregates
- Implement business logic
- Add new endpoints
- Write tests

###  Get Help

- **GitHub Issues**: https://github.com/basilex/promenade/issues
- **Documentation**: https://basilex.github.io/promenade/
- **Email**: alexander.vasilenko@gmail.com

---

**Need help?** Open an issue on GitHub or check our comprehensive documentation!

**Version**: 0.1.0  
**Last Updated**: January 5, 2026  
**Status**: Production-ready
