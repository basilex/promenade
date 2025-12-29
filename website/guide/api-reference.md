# API Reference

**Complete REST API documentation for Promenade Platform**

---

## Base URL

```
http://localhost:8081/api/v1
```

**Production**: `https://api.promenade.example.com/api/v1`

---

## Authentication

Most endpoints require JWT authentication via Bearer token:

```http
Authorization: Bearer <access_token>
```

**Token Types**:
- **Access Token**: 15 minutes TTL (for API requests)
- **Refresh Token**: 7 days TTL (to get new access token)

**Authentication Flow**:
1. Register or Login → Get tokens
2. Use Access Token for API calls
3. When expired → Use Refresh Token to get new Access Token

---

## Response Format

### Success Response

```json
{
  "status": "success",
  "data": { /* response data */ }
}
```

### Error Response

```json
{
  "status": "error",
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

### Paginated Response

```json
{
  "status": "success",
  "data": [ /* items */ ],
  "pagination": {
    "total": 150,
    "page": 1,
    "page_size": 20,
    "total_pages": 8
  }
}
```

---

## Identity Context

### User Management

#### Register User

```http
POST /identity/users/register
Content-Type: application/json

{
  "email": "john@example.com",
  "name": "John Doe",
  "password": "SecurePass123"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "expires_at": "2025-12-29T12:15:00Z",
    "token_type": "Bearer",
    "user": {
      "id": "01JGABC...",
      "email": "john@example.com",
      "name": "John Doe",
      "status": "active"
    }
  }
}
```

#### Login

```http
POST /identity/users/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePass123"
}
```

**Response** (200 OK): Same as Register

#### Refresh Token

```http
POST /identity/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGci..."
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "eyJhbGci...",
    "expires_at": "2025-12-29T12:30:00Z"
  }
}
```

#### Revoke Token (Logout)

```http
POST /identity/auth/revoke
Authorization: Bearer <token>
Content-Type: application/json

{
  "token": "<access_token_to_revoke>"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "Token revoked successfully"
  }
}
```

#### Get My Profile

```http
GET /identity/users/me
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC...",
    "email": "john@example.com",
    "name": "John Doe",
    "status": "active",
    "is_email_verified": false,
    "created_at": "2025-12-29T12:00:00Z"
  }
}
```

#### List Users (Admin)

```http
GET /identity/users?page=1&page_size=20
Authorization: Bearer <admin_token>
```

**Response** (200 OK): Paginated list of users

---

### Profile Management

#### Create Profile

```http
POST /identity/profiles
Authorization: Bearer <token>
Content-Type: application/json

{
  "display_name": "John D.",
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Software developer",
  "timezone": "Europe/Kyiv",
  "language": "uk",
  "country": "UA"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEF...",
    "user_id": "01JGABC...",
    "display_name": "John D.",
    "first_name": "John",
    "last_name": "Doe",
    "bio": "Software developer",
    "timezone": "Europe/Kyiv",
    "language": "uk",
    "country": "UA",
    "is_public": false
  }
}
```

#### Get My Profile

```http
GET /identity/profiles/me
Authorization: Bearer <token>
```

#### Update Profile

```http
PUT /identity/profiles/me
Authorization: Bearer <token>
Content-Type: application/json

{
  "bio": "Senior Software Developer",
  "is_public": true
}
```

---

### Contact Management

#### Create Email Contact

```http
POST /identity/contacts
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "email",
  "email": "john.work@company.com",
  "label": "Work",
  "is_primary": false
}
```

#### Create Phone Contact

```http
POST /identity/contacts
Content-Type: application/json

{
  "type": "phone",
  "phone": "+441234567890",
  "label": "Mobile",
  "is_primary": true
}
```

#### List My Contacts

```http
GET /identity/contacts?type=email
Authorization: Bearer <token>
```

#### Verify Contact

```http
POST /identity/contacts/:id/verify
Authorization: Bearer <token>
```

---

### Role & Permission Management (Admin)

#### Create Role

```http
POST /identity/roles
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "moderator",
  "display_name": "Moderator",
  "description": "Content moderation role"
}
```

#### Assign Permissions to Role

```http
POST /identity/roles/:id/permissions
Content-Type: application/json

{
  "permission_ids": [
    "01JGPERM1...",
    "01JGPERM2..."
  ]
}
```

#### Assign Role to User

```http
POST /identity/users/:id/roles
Content-Type: application/json

{
  "role_id": "01JGROLE..."
}
```

---

## Shared Context

### Reference Data (Public)

#### Get Countries

```http
GET /shared/countries?is_active=true
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "code": "UA",
      "name": "Ukraine",
      "name_uk": "Україна",
      "phone_code": "+380",
      "currency_code": "UAH",
      "is_active": true
    }
  ]
}
```

#### Get Currencies

```http
GET /shared/currencies
```

#### Get Languages

```http
GET /shared/languages
```

#### Get Timezones

```http
GET /shared/timezones?region=Europe
```

---

## Customer Management Context

### Customer Management

#### Create Customer

```http
POST /customer-mgmt/customers
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+441234567890",
  "status": "lead",
  "tier": "free",
  "source": "website",
  "tags": ["marketing"]
}
```

#### List Customers

```http
GET /customer-mgmt/customers?status=lead&tier=free&page=1
Authorization: Bearer <token>
```

**Query Parameters**:
- `status` - Filter by status (lead, prospect, customer, churned)
- `tier` - Filter by tier (free, basic, pro, enterprise)
- `assigned_to` - Filter by sales rep ID
- `page`, `page_size` - Pagination

#### Get Customer

```http
GET /customer-mgmt/customers/:id
Authorization: Bearer <token>
```

#### Update Customer

```http
PUT /customer-mgmt/customers/:id
Content-Type: application/json

{
  "name": "John Doe Updated",
  "tags": ["vip"]
}
```

#### Qualify as Prospect

```http
POST /customer-mgmt/customers/:id/qualify
Authorization: Bearer <token>
```

#### Convert to Customer

```http
POST /customer-mgmt/customers/:id/convert
Authorization: Bearer <token>
```

#### Churn Customer

```http
POST /customer-mgmt/customers/:id/churn
Content-Type: application/json

{
  "reason": "Price too high"
}
```

#### Assign Sales Rep

```http
POST /customer-mgmt/customers/:id/assign
Content-Type: application/json

{
  "assigned_to": "01JGREP..."
}
```

---

## Error Codes

| Code                    | HTTP Status | Description                        |
| ----------------------- | ----------- | ---------------------------------- |
| `VALIDATION_ERROR`      | 400         | Request validation failed          |
| `UNAUTHORIZED`          | 401         | Missing or invalid token           |
| `TOKEN_EXPIRED`         | 401         | Access token expired               |
| `TOKEN_REVOKED`         | 401         | Token has been revoked             |
| `FORBIDDEN`             | 403         | Insufficient permissions           |
| `NOT_FOUND`             | 404         | Resource not found                 |
| `USER_NOT_FOUND`        | 404         | User does not exist                |
| `CUSTOMER_NOT_FOUND`    | 404         | Customer does not exist            |
| `CONFLICT`              | 409         | Resource already exists            |
| `EMAIL_ALREADY_EXISTS`  | 409         | Email already registered           |
| `RATE_LIMIT_EXCEEDED`   | 429         | Too many requests                  |
| `INTERNAL_ERROR`        | 500         | Server error                       |

---

## Rate Limiting

**IP-based rate limiting** on authentication endpoints:

- **Login**: 5 requests per minute
- **Register**: 3 requests per minute

**Response Headers**:
```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 3
X-RateLimit-Reset: 1735476000
```

**Error Response** (429 Too Many Requests):
```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 45 seconds."
  }
}
```

---

## Pagination

**Query Parameters**:
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)

**Example**:
```http
GET /identity/users?page=2&page_size=50
```

**Response**:
```json
{
  "status": "success",
  "data": [ /* items */ ],
  "pagination": {
    "total": 150,
    "page": 2,
    "page_size": 50,
    "total_pages": 3
  }
}
```

---

## Filtering

**Query Parameters** (endpoint-specific):

**Users**:
- `status` - active, suspended, banned
- `is_email_verified` - true, false

**Customers**:
- `status` - lead, prospect, customer, churned
- `tier` - free, basic, pro, enterprise
- `assigned_to` - Sales rep UUID
- `tags` - Comma-separated tags

**Example**:
```http
GET /customer-mgmt/customers?status=customer&tier=enterprise&tags=vip
```

---

## Testing with cURL

### Complete User Setup

```bash
# 1. Register
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test","password":"Pass123!"}' | \
  jq -r '.data.access_token')

# 2. Create Profile
curl -X POST http://localhost:8081/api/v1/identity/profiles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Test User","timezone":"UTC","language":"en"}'

# 3. Add Contact
curl -X POST http://localhost:8081/api/v1/identity/contacts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"email","email":"work@example.com","label":"Work"}'

# 4. Get My Profile
curl http://localhost:8081/api/v1/identity/users/me \
  -H "Authorization: Bearer $TOKEN"
```

---

## SDK & Client Libraries

**Coming Soon**:
- Go SDK
- JavaScript/TypeScript SDK
- Python SDK

**Current**: Use standard HTTP clients (fetch, axios, requests, etc.)

---

## Next Steps

- [Quick Start Guide](/guide/quick-start) - Get started in 5 minutes
- [Authentication Guide](/guide/rbac) - JWT and RBAC implementation
- [Testing Guide](/guide/testing) - API testing patterns
- [GitHub Repository](https://github.com/basilex/promenade) - Source code
