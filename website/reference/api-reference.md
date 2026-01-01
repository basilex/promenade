# API Reference

Complete REST API documentation for Promenade Platform with authentication, endpoints, request/response examples, and error codes.

---

## Base URL

```
Development: http://localhost:8081/api/v1
Production:  https://api.promenade.com/api/v1
```

---

## Authentication

All protected endpoints require JWT token in `Authorization` header:

```bash
Authorization: Bearer <access_token>
```

### Obtain Token

**Endpoint**: `POST /identity/users/login`

**Request**:
```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-12-29T10:30:00Z",
    "token_type": "Bearer",
    "user": {
      "id": "01JGABC123",
      "email": "user@example.com",
      "name": "John Doe",
      "status": "active"
    }
  }
}
```

### Refresh Token

**Endpoint**: `POST /identity/auth/refresh`

**Request**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "access_token": "new_access_token",
    "refresh_token": "new_refresh_token",
    "expires_at": "2025-12-29T10:45:00Z",
    "token_type": "Bearer"
  }
}
```

---

## Identity Context

### User Management

#### Register User

**POST** `/identity/users/register`

**Public endpoint** (no authentication required)

**Request**:
```json
{
  "email": "newuser@example.com",
  "name": "Jane Smith",
  "password": "SecurePass456"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEF456",
    "email": "newuser@example.com",
    "name": "Jane Smith",
    "status": "active",
    "created_at": "2025-12-29T10:00:00Z"
  }
}
```

#### Get Current User Profile

**GET** `/identity/users/me`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC123",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "active",
    "is_locked": false,
    "failed_login_attempts": 0,
    "last_login_at": "2025-12-29T09:00:00Z",
    "created_at": "2025-01-15T08:00:00Z"
  }
}
```

#### List Users

**GET** `/identity/users?page=1&page_size=20&status=active`

**Authentication**: Required (Admin role)

**Query Parameters**:
- `page` (int): Page number (default: 1)
- `page_size` (int): Items per page (default: 20, max: 100)
- `status` (string): Filter by status (active/suspended/banned)
- `search` (string): Search by email or name

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGABC123",
      "email": "user1@example.com",
      "name": "John Doe",
      "status": "active"
    },
    {
      "id": "01JGDEF456",
      "email": "user2@example.com",
      "name": "Jane Smith",
      "status": "active"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### Update User Status

**PATCH** `/identity/users/:id/status`

**Authentication**: Required (Admin role)

**Request**:
```json
{
  "status": "suspended",
  "reason": "Policy violation"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "User status updated successfully"
  }
}
```

### Contact Management

#### Create Email Contact

**POST** `/identity/contacts`

**Authentication**: Required

**Request**:
```json
{
  "type": "email",
  "email": "john.doe@work.com",
  "label": "Work",
  "is_primary": true
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGHIJ789",
    "user_id": "01JGABC123",
    "type": "email",
    "email": "john.doe@work.com",
    "label": "Work",
    "is_primary": true,
    "is_verified": false,
    "is_public": false,
    "created_at": "2025-12-29T10:15:00Z"
  }
}
```

#### List User Contacts

**GET** `/identity/contacts?type=email`

**Authentication**: Required

**Query Parameters**:
- `type` (string): Filter by type (email/phone/address)
- `is_primary` (bool): Filter primary contacts

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGHIJ789",
      "type": "email",
      "email": "john.doe@work.com",
      "label": "Work",
      "is_primary": true,
      "is_verified": true
    },
    {
      "id": "01JGKLM012",
      "type": "phone",
      "phone": "+380501234567",
      "label": "Mobile",
      "is_primary": true,
      "is_verified": false
    }
  ]
}
```

#### Verify Contact

**POST** `/identity/contacts/:id/verify`

**Authentication**: Required

**Request**:
```json
{
  "verification_code": "123456"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "Contact verified successfully"
  }
}
```

### Profile Management

#### Create Profile

**POST** `/identity/profiles`

**Authentication**: Required

**Request**:
```json
{
  "display_name": "John Doe",
  "first_name": "John",
  "last_name": "Doe",
  "bio": "Software engineer passionate about clean code",
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
    "id": "01JGNOP345",
    "user_id": "01JGABC123",
    "display_name": "John Doe",
    "first_name": "John",
    "last_name": "Doe",
    "bio": "Software engineer passionate about clean code",
    "timezone": "Europe/Kyiv",
    "language": "uk",
    "country": "UA",
    "is_public": false,
    "is_active": true,
    "created_at": "2025-12-29T10:20:00Z"
  }
}
```

#### Get Profile

**GET** `/identity/profiles/:id`

**Authentication**: Required (own profile or public profiles)

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGNOP345",
    "display_name": "John Doe",
    "bio": "Software engineer passionate about clean code",
    "avatar_url": "https://cdn.promenade.com/avatars/01JGNOP345.jpg",
    "timezone": "Europe/Kyiv",
    "language": "uk",
    "country": "UA",
    "website": "https://johndoe.com",
    "linkedin": "https://linkedin.com/in/johndoe",
    "github": "https://github.com/johndoe"
  }
}
```

### Role & Permission Management (RBAC)

#### List Roles

**GET** `/identity/roles`

**Authentication**: Required (Admin role)

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGQRS678",
      "name": "admin",
      "display_name": "Administrator",
      "description": "Full system access",
      "is_system": true,
      "permissions_count": 29
    },
    {
      "id": "01JGTUV901",
      "name": "user",
      "display_name": "Regular User",
      "description": "Standard user access",
      "is_system": true,
      "permissions_count": 8
    }
  ]
}
```

#### Assign Role to User

**POST** `/identity/users/:user_id/roles`

**Authentication**: Required (Admin role)

**Request**:
```json
{
  "role_id": "01JGQRS678"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "Role assigned successfully"
  }
}
```

---

## Shared Context (Reference Data)

### Countries

#### List Countries

**GET** `/shared/countries?is_active=true`

**Public endpoint**

**Query Parameters**:
- `is_active` (bool): Filter by status
- `region` (string): Filter by region (Europe, Asia, etc.)
- `search` (string): Search by name or code

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGWXY234",
      "code": "UA",
      "name": "Ukraine",
      "local_name": "Україна",
      "region": "Europe",
      "calling_code": "+380",
      "currency_code": "UAH",
      "language_code": "uk",
      "is_active": true
    },
    {
      "id": "01JGZAB567",
      "code": "US",
      "name": "United States",
      "region": "North America",
      "calling_code": "+1",
      "currency_code": "USD",
      "language_code": "en",
      "is_active": true
    }
  ]
}
```

### Currencies

#### List Currencies

**GET** `/shared/currencies?is_active=true`

**Public endpoint**

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGCDE890",
      "code": "USD",
      "name": "US Dollar",
      "symbol": "$",
      "decimal_places": 2,
      "is_active": true
    },
    {
      "id": "01JGFGH123",
      "code": "EUR",
      "name": "Euro",
      "symbol": "€",
      "decimal_places": 2,
      "is_active": true
    }
  ]
}
```

---

## Customer Management Context

### Customer Management

#### Create Customer

**POST** `/customer-mgmt/customers`

**Authentication**: Required

**Request**:
```json
{
  "name": "Acme Corporation",
  "type": "company",
  "status": "lead",
  "tier": "enterprise",
  "email": "contact@acme.com",
  "phone": "+1-555-0100",
  "website": "https://acme.com",
  "tags": ["enterprise", "saas", "priority"],
  "assigned_to": "01JGABC123"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGIJK456",
    "name": "Acme Corporation",
    "type": "company",
    "status": "lead",
    "tier": "enterprise",
    "email": "contact@acme.com",
    "phone": "+1-555-0100",
    "website": "https://acme.com",
    "tags": ["enterprise", "saas", "priority"],
    "assigned_to": "01JGABC123",
    "created_at": "2025-12-29T11:00:00Z"
  }
}
```

#### List Customers

**GET** `/customer-mgmt/customers?status=lead&tier=enterprise&page=1&page_size=20`

**Authentication**: Required

**Query Parameters**:
- `status` (string): lead/prospect/customer/churned
- `tier` (string): free/basic/pro/enterprise
- `type` (string): individual/company
- `assigned_to` (uuid): Filter by sales rep
- `search` (string): Search by name, email, phone
- `tags` (string): Filter by tags (comma-separated)
- `page`, `page_size`: Pagination

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGIJK456",
      "name": "Acme Corporation",
      "type": "company",
      "status": "lead",
      "tier": "enterprise",
      "email": "contact@acme.com",
      "created_at": "2025-12-29T11:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 45
  }
}
```

#### Get Customer by ID

**GET** `/customer-mgmt/customers/:id`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGIJK456",
    "name": "Acme Corporation",
    "type": "company",
    "status": "customer",
    "tier": "enterprise",
    "email": "contact@acme.com",
    "phone": "+1-555-0100",
    "website": "https://acme.com",
    "tags": ["enterprise", "saas", "priority"],
    "assigned_to": "01JGABC123",
    "source": "referral",
    "notes": [
      {
        "id": "01JGLMN789",
        "content": "Initial contact made, interested in enterprise plan",
        "created_at": "2025-12-29T11:15:00Z",
        "created_by": "01JGABC123"
      }
    ],
    "created_at": "2025-12-29T11:00:00Z",
    "updated_at": "2025-12-29T11:15:00Z"
  }
}
```

#### Update Customer Status

**PATCH** `/customer-mgmt/customers/:id/status`

**Authentication**: Required

**Request**:
```json
{
  "status": "customer",
  "reason": "Contract signed"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "Customer status updated to customer"
  }
}
```

#### Delete Customer

**DELETE** `/customer-mgmt/customers/:id`

**Authentication**: Required (Admin role)

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "Customer deleted successfully"
  }
}
```

### Deal Management

#### Create Deal

**POST** `/customer-mgmt/deals`

**Authentication**: Required

**Request**:
```json
{
  "title": "Enterprise Plan - Acme Corp",
  "customer_id": "01JGIJK456",
  "value_cents": 5000000,
  "currency": "USD",
  "stage": "lead",
  "expected_close_date": "2026-01-31",
  "assigned_to": "01JGABC123"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGOPQ012",
    "title": "Enterprise Plan - Acme Corp",
    "customer_id": "01JGIJK456",
    "value": {
      "amount": "50000.00",
      "currency": "USD"
    },
    "stage": "lead",
    "probability": 10,
    "expected_close_date": "2026-01-31",
    "assigned_to": "01JGABC123",
    "status": "open",
    "created_at": "2025-12-29T11:30:00Z"
  }
}
```

#### Move Deal to Next Stage

**POST** `/customer-mgmt/deals/:id/move-stage`

**Authentication**: Required

**Request**:
```json
{
  "stage": "qualified",
  "notes": "Budget confirmed, decision maker identified"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGOPQ012",
    "stage": "qualified",
    "probability": 25,
    "updated_at": "2025-12-29T12:00:00Z"
  }
}
```

#### Mark Deal as Won

**POST** `/customer-mgmt/deals/:id/win`

**Authentication**: Required

**Request**:
```json
{
  "actual_value_cents": 4500000,
  "notes": "Contract signed with 10% discount"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGOPQ012",
    "stage": "closed_won",
    "status": "won",
    "actual_value": {
      "amount": "45000.00",
      "currency": "USD"
    },
    "actual_close_date": "2025-12-29",
    "won_at": "2025-12-29T12:30:00Z"
  }
}
```

#### Get Pipeline Statistics

**GET** `/customer-mgmt/deals/stats/pipeline`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "by_stage": [
      {
        "stage": "lead",
        "count": 15,
        "total_value": "750000.00",
        "probability": 10
      },
      {
        "stage": "qualified",
        "count": 8,
        "total_value": "400000.00",
        "probability": 25
      },
      {
        "stage": "proposal",
        "count": 5,
        "total_value": "350000.00",
        "probability": 50
      }
    ],
    "total_deals": 28,
    "total_value": "1500000.00",
    "weighted_value": "525000.00"
  }
}
```

---

## Order Management Context

### Order Management

#### Create Order

**POST** `/order-mgmt/orders`

**Authentication**: Required

**Request**:
```json
{
  "customer_id": "01JGIJK456",
  "currency": "USD",
  "items": [
    {
      "description": "Enterprise Plan - Annual",
      "quantity": 1,
      "unit_price_cents": 1200000
    },
    {
      "description": "Premium Support",
      "quantity": 1,
      "unit_price_cents": 300000
    }
  ]
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGRST345",
    "order_number": "ORD-2025-000042",
    "customer_id": "01JGIJK456",
    "status": "pending",
    "total": {
      "amount": "15000.00",
      "currency": "USD"
    },
    "items": [
      {
        "id": "01JGUVW678",
        "description": "Enterprise Plan - Annual",
        "quantity": 1,
        "unit_price": "12000.00",
        "total": "12000.00"
      },
      {
        "id": "01JGXYZ901",
        "description": "Premium Support",
        "quantity": 1,
        "unit_price": "3000.00",
        "total": "3000.00"
      }
    ],
    "created_at": "2025-12-29T13:00:00Z"
  }
}
```

#### List Orders

**GET** `/order-mgmt/orders?status=pending&customer_id=01JGIJK456`

**Authentication**: Required

**Query Parameters**:
- `status` (string): pending/confirmed/processing/fulfilled/cancelled
- `customer_id` (uuid): Filter by customer
- `date_from`, `date_to` (string): Filter by date range
- `page`, `page_size`: Pagination

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGRST345",
      "order_number": "ORD-2025-000042",
      "customer_id": "01JGIJK456",
      "status": "pending",
      "total": "15000.00",
      "currency": "USD",
      "created_at": "2025-12-29T13:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 1
  }
}
```

#### Get Order by ID

**GET** `/order-mggt/orders/:id`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGRST345",
    "order_number": "ORD-2025-000042",
    "customer_id": "01JGIJK456",
    "status": "confirmed",
    "total": {
      "amount": "15000.00",
      "currency": "USD"
    },
    "items": [
      {
        "id": "01JGUVW678",
        "description": "Enterprise Plan - Annual",
        "quantity": 1,
        "unit_price": "12000.00",
        "total": "12000.00"
      }
    ],
    "confirmed_at": "2025-12-29T13:15:00Z",
    "created_at": "2025-12-29T13:00:00Z",
    "updated_at": "2025-12-29T13:15:00Z"
  }
}
```

#### Confirm Order

**POST** `/order-mgmt/orders/:id/confirm`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGRST345",
    "status": "confirmed",
    "confirmed_at": "2025-12-29T13:15:00Z"
  }
}
```

#### Start Processing Order

**POST** `/order-mgmt/orders/:id/process`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGRST345",
    "status": "processing",
    "processing_started_at": "2025-12-29T13:30:00Z"
  }
}
```

#### Mark Order as Fulfilled

**POST** `/order-mgmt/orders/:id/fulfill`

**Authentication**: Required

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGRST345",
    "status": "fulfilled",
    "fulfilled_at": "2025-12-29T14:00:00Z"
  }
}
```

#### Cancel Order

**POST** `/order-mgmt/orders/:id/cancel`

**Authentication**: Required

**Request**:
```json
{
  "reason": "Customer requested cancellation"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGRST345",
    "status": "cancelled",
    "cancelled_at": "2025-12-29T13:45:00Z"
  }
}
```

---

## Health Checks

### System Health

**GET** `/health`

**Public endpoint**

**Response** (200 OK - Healthy):
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy",
    "event_bus": "healthy"
  },
  "timestamp": "2025-12-29T15:00:00Z"
}
```

**Response** (503 Service Unavailable - Unhealthy):
```json
{
  "status": "unhealthy",
  "version": "1.0.0",
  "dependencies": {
    "database": "unhealthy",
    "redis": "degraded",
    "event_bus": "healthy"
  },
  "timestamp": "2025-12-29T15:00:00Z"
}
```

### Database Health

**GET** `/health/db`

**Response** (200 OK):
```json
{
  "status": "healthy",
  "ping_ms": 5,
  "connections": {
    "active": 12,
    "idle": 8,
    "max": 50
  }
}
```

### Redis Health

**GET** `/health/redis`

**Response** (200 OK):
```json
{
  "status": "healthy",
  "ping_ms": 2,
  "memory_used_mb": 45,
  "connected_clients": 8
}
```

### Event Bus Health

**GET** `/health/bus`

**Response** (200 OK):
```json
{
  "status": "healthy",
  "adapter": "redis",
  "published_total": 15234,
  "failed_total": 3
}
```

---

## Error Responses

### Standard Error Format

All error responses follow this structure:

```json
{
  "status": "error",
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {}
  }
}
```

### HTTP Status Codes

| Code | Meaning            | Description                           |
|------|--------------------|---------------------------------------|
| 200  | OK                 | Request successful                    |
| 201  | Created            | Resource created successfully         |
| 400  | Bad Request        | Invalid request data                  |
| 401  | Unauthorized       | Missing or invalid authentication     |
| 403  | Forbidden          | Insufficient permissions              |
| 404  | Not Found          | Resource not found                    |
| 409  | Conflict           | Resource already exists               |
| 422  | Unprocessable      | Validation failed                     |
| 429  | Too Many Requests  | Rate limit exceeded                   |
| 500  | Internal Error     | Server error                          |
| 503  | Service Unavailable| Service temporarily unavailable       |

### Common Error Codes

**Authentication Errors**:
- `UNAUTHORIZED`: Missing or invalid token
- `TOKEN_EXPIRED`: JWT token expired
- `INVALID_CREDENTIALS`: Invalid email/password
- `ACCOUNT_LOCKED`: Too many failed login attempts

**Validation Errors**:
- `VALIDATION_ERROR`: Request validation failed
- `INVALID_EMAIL`: Invalid email format
- `INVALID_PHONE`: Invalid phone format
- `INVALID_UUID`: Invalid UUID format

**Resource Errors**:
- `NOT_FOUND`: Resource not found
- `ALREADY_EXISTS`: Resource already exists (duplicate)
- `FORBIDDEN`: Insufficient permissions

**Business Logic Errors**:
- `INVALID_STATUS_TRANSITION`: Cannot change status (e.g., pending → fulfilled without processing)
- `INSUFFICIENT_BALANCE`: Not enough funds
- `LIMIT_EXCEEDED`: Operation limit exceeded

**Rate Limiting Errors**:
- `RATE_LIMIT_EXCEEDED`: Too many requests (X-RateLimit-Reset header shows when to retry)

### Error Response Examples

**401 Unauthorized**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing or invalid authentication token"
  }
}
```

**400 Validation Error**:
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "email": "invalid email format",
      "password": "must be at least 8 characters"
    }
  }
}
```

**404 Not Found**:
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "Customer with ID 01JGIJK456 not found"
  }
}
```

**429 Rate Limit**:
```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many login attempts, try again in 60 seconds"
  }
}
```

---

## Rate Limiting

### Rate Limits by Endpoint

| Endpoint                     | Limit      | Window  |
|------------------------------|------------|---------|
| `POST /identity/users/login` | 5 requests | 1 min   |
| `POST /identity/users/register` | 3 requests | 1 min   |
| `POST /identity/auth/refresh` | 10 requests | 1 min   |
| Other authenticated endpoints | 100 requests | 1 min   |
| Public endpoints             | 1000 requests | 1 hour  |

### Rate Limit Headers

Every response includes rate limit information:

```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 3
X-RateLimit-Reset: 1735481400  (Unix timestamp)
```

**See**: [Rate Limiting Guide](rate-limiting.md) for complete documentation

---

## Pagination

### Query Parameters

```
?page=1&page_size=20
```

- `page` (int): Page number (1-indexed, default: 1)
- `page_size` (int): Items per page (default: 20, max: 100)

### Response Format

```json
{
  "status": "success",
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

## Filtering & Sorting

### Filtering

Use query parameters:
```
GET /customer-mgmt/customers?status=lead&tier=enterprise&tags=priority
```

### Sorting

Use `sort` parameter:
```
GET /customer-mgmt/customers?sort=-created_at,name
```

- Prefix with `-` for descending order
- Comma-separated for multiple fields

---

## Versioning

API version is included in URL:
```
/api/v1/...
```

Breaking changes will increment major version:
```
/api/v2/...  (future)
```

---

## CORS

**Allowed Origins** (Production):
- `https://app.promenade.com`
- `https://promenade.com`

**Allowed Methods**:
- GET, POST, PUT, PATCH, DELETE, OPTIONS

**Allowed Headers**:
- Authorization, Content-Type, X-Request-ID

---

## Webhooks (Future)

Subscribe to events:
```json
{
  "url": "https://your-app.com/webhooks/promenade",
  "events": ["customer.created", "order.confirmed", "deal.won"],
  "secret": "webhook_secret_key"
}
```

Event payload:
```json
{
  "event": "customer.created",
  "data": {...},
  "timestamp": "2025-12-29T15:30:00Z"
}
```

**See**: [Webhooks Guide](webhooks.md) (coming Q2 2026)

---

## Related Documentation

- [RBAC Guide](rbac.md) - Role-based access control
- [Rate Limiting](rate-limiting.md) - IP-based protection
- [Health Checks](health-checks.md) - Monitoring endpoints
- [Testing Patterns](testing-patterns.md) - API testing guide
- [Development Workflow](development-workflow.md) - Contributing

---

**Questions?** [Open a discussion](https://github.com/basilex/promenade/discussions)

**Found a bug in API?** [Report an issue](https://github.com/basilex/promenade/issues)
