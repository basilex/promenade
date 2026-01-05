# Authentication Flow Guide

**Complete guide to JWT authentication** in Promenade Platform - from registration to RBAC.

---

## Overview

Promenade uses **JWT (JSON Web Tokens)** for stateless authentication with **Role-Based Access Control (RBAC)**.

### Key Features

- **Token-based**: Stateless authentication (no server-side sessions)
- **Dual Tokens**: Access token (15 min) + Refresh token (7 days)
- **RBAC**: Role-based authorization (5 system roles, 29+ permissions)
- **Token Revocation**: Redis-based blacklist for logout
- **Rate Limiting**: Protection against brute-force attacks
- **Secure**: bcrypt password hashing, HTTPS recommended

---

## Authentication Flow Diagram

```

   Client    

       
        1. POST /register (email, password, name)
       
                                                       
                                                       
               
  Server: User Handler                               
  - Validate input                                   
  - Hash password (bcrypt)                           
  - Create user in database                          
  - Return user details (no token)                   
               
                                                       
        2. 201 Created                                 
        {"id":"...", "email":"...", "status":"active"} 
       
       
        3. POST /login (email, password)
       
                                                       
                                                       
               
  Server: Auth Handler                               
  - Validate credentials                              
  - Load user roles from database                     
  - Generate access token (15 min)                   
  - Generate refresh token (7 days)                  
  - Return token pair                                 
               
                                                       
        4. 200 OK                                      
        {"access_token":"...", "refresh_token":"..."}  
       
       
        (Save tokens in memory/secure storage)
       
        5. GET /api/v1/customers
           Authorization: Bearer <access_token>
       
                                                       
                                                       
               
  Server: Auth Middleware                            
  - Extract token from header                         
  - Validate token signature                          
  - Check expiration                                  
  - Check revocation (Redis blacklist)                
  - Extract claims (user_id, roles)                   
  - Store in context                                  
               
                                                       
        6. Pass to Protected Handler                   
                                                       
               
  Protected Handler                                  
  - Access user_id from context                       
  - Process business logic                            
  - Return response                                   
               
                                                       
        7. 200 OK                                      
        {"status":"success", "data":[...]}             
       
```

---

## Step-by-Step Authentication

### Step 1: User Registration

**Endpoint**: `POST /api/v1/identity/users/register`

**Request**:
```bash
curl -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Response** (201 Created):
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

**What Happens**:
1. Server validates input (email format, password strength)
2. Password hashed with bcrypt (cost factor: 10)
3. User created in `identity_users` table
4. Default role `user` assigned
5. User details returned (no token yet)

**Password Requirements**:
- Minimum 8 characters
- At least 1 letter
- At least 1 digit
- No maximum length

**Rate Limit**: 3 requests per minute per IP

---

### Step 2: Login to Get JWT Token

**Endpoint**: `POST /api/v1/identity/users/login`

**Request**:
```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123"
  }'
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFKR0FCQzEyM0RFRjQ1NkdISTc4OUpLTDAiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sImlzcyI6InByb21lbmFkZS1wbGF0Zm9ybSIsImV4cCI6MTczNjA4NzE2MCwiaWF0IjoxNzM2MDg2MjYwfQ.Gz5o8Kx3yHqJR4dL2fN9vP1wQ6sT7uX8zA0bC1dE2fG",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMDFKR0FCQzEyM0RFRjQ1NkdISTc4OUpLTDAiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sImlzcyI6InByb21lbmFkZS1wbGF0Zm9ybSIsImV4cCI6MTczNjY5MTA2MCwiaWF0IjoxNzM2MDg2MjYwfQ.H1a2b3C4d5E6f7G8h9I0j1K2l3M4n5O6p7Q8r9S0t1U",
    "expires_at": "2026-01-05T14:46:00Z",
    "token_type": "Bearer",
    "user": {
      "id": "01JGABC123DEF456GHI789JKL0",
      "email": "john@example.com",
      "name": "John Doe",
      "roles": ["user"]
    }
  }
}
```

**Token Structure**:

**Access Token** (decoded):
```json
{
  "user_id": "01JGABC123DEF456GHI789JKL0",
  "email": "john@example.com",
  "roles": ["user"],
  "iss": "promenade-platform",
  "exp": 1736087160,
  "iat": 1736086260
}
```

**What Happens**:
1. Server validates email and password
2. Checks user status (active/suspended/banned)
3. Loads user roles from database
4. Generates access token (15 min expiration)
5. Generates refresh token (7 days expiration)
6. Returns token pair

**Rate Limit**: 5 requests per minute per IP

---

### Step 3: Use Access Token

**All Protected Endpoints** require `Authorization` header:

```bash
curl -X GET http://localhost:8081/api/v1/customer-mgmt/customers \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Header Format**: `Authorization: Bearer <access_token>`

**What Happens**:
1. Middleware extracts token from header
2. Validates token signature (HMAC-SHA256)
3. Checks expiration time
4. Checks revocation list (Redis)
5. Extracts claims (user_id, email, roles)
6. Stores in request context
7. Passes to handler

---

### Step 4: Token Refresh

**When Access Token Expires** (after 15 minutes):

**Endpoint**: `POST /api/v1/identity/auth/refresh`

**Request**:
```bash
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Response** (200 OK):
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

**What Happens**:
1. Server validates refresh token
2. Checks expiration (7 days)
3. Loads user from database
4. Generates new access token
5. Generates new refresh token (rotation)
6. Returns new token pair

**Token Rotation**: Each refresh generates new refresh token (prevents replay attacks)

---

### Step 5: Logout

**Endpoint**: `POST /api/v1/identity/auth/logout`

**Request**:
```bash
curl -X POST http://localhost:8081/api/v1/identity/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "message": "logged out successfully"
  }
}
```

**What Happens**:
1. Server extracts access token
2. Calculates remaining TTL
3. Adds token to Redis blacklist
4. Sets expiration = token TTL
5. Future requests with this token fail

**Note**: Refresh token also invalidated (add to blacklist separately)

---

## Role-Based Access Control (RBAC)

### System Roles

Promenade includes **5 pre-configured roles**:

| Role          | Description                           | Permissions Count |
| ------------- | ------------------------------------- | ----------------- |
| **superadmin**| Full system access                    | 29+ (all)         |
| **admin**     | Administrative access                 | 25                |
| **manager**   | Management operations                 | 18                |
| **user**      | Standard user access                  | 12                |
| **guest**     | Read-only access                      | 5                 |

### Permission Categories

**29+ permissions** across 6 resource categories:

1. **Users**: `users:create`, `users:read`, `users:update`, `users:delete`, `users:list`
2. **Customers**: `customers:create`, `customers:read`, `customers:update`, `customers:delete`, `customers:list`
3. **Orders**: `orders:create`, `orders:read`, `orders:update`, `orders:delete`, `orders:list`
4. **Invoices**: `invoices:create`, `invoices:read`, `invoices:update`, `invoices:delete`, `invoices:list`
5. **Payments**: `payments:create`, `payments:read`, `payments:update`, `payments:delete`, `payments:list`
6. **System**: `roles:manage`, `permissions:manage`, `settings:manage`, `logs:read`

### Checking User Roles

**Get Current User**:
```bash
curl -X GET http://localhost:8081/api/v1/identity/users/me \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC123DEF456GHI789JKL0",
    "email": "john@example.com",
    "name": "John Doe",
    "roles": ["user", "manager"],
    "permissions": [
      "customers:read",
      "customers:create",
      "customers:update",
      "orders:read",
      "orders:create"
    ]
  }
}
```

### Role Assignment

**Admin assigns roles** to users:

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/$USER_ID/roles \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "manager"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "user_id": "01JGABC123DEF456GHI789JKL0",
    "role": "manager",
    "assigned_at": "2026-01-05T15:10:00Z"
  }
}
```

### Protected Routes with RBAC

#### Admin-Only Endpoint

```bash
# Create new user (admin only)
curl -X POST http://localhost:8081/api/v1/identity/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "name": "New User",
    "password": "TempPass123",
    "role": "user"
  }'
```

**Error if not admin** (403 Forbidden):
```json
{
  "status": "error",
  "error": {
    "code": "FORBIDDEN",
    "message": "admin role required"
  }
}
```

#### Manager-Level Endpoint

```bash
# Suspend user (manager or admin)
curl -X POST http://localhost:8081/api/v1/identity/users/$USER_ID/suspend \
  -H "Authorization: Bearer $MANAGER_TOKEN"
```

---

## Security Best Practices

### Token Storage

**Client-Side Storage Options**:

| Storage           | Security Level | Use Case                 | Risks                          |
| ----------------- | -------------- | ------------------------ | ------------------------------ |
| **Memory**        | High          | Single-page apps         | Lost on page refresh           |
| **sessionStorage**| Medium        | Tab-scoped sessions      | Lost on tab close              |
| **localStorage**  | Low           | Avoid for tokens         | XSS vulnerable                 |
| **HTTP-only Cookie** | High       | Production recommended   | CSRF protection required       |

**Recommended**: Store access token in **memory**, refresh token in **HTTP-only secure cookie**.

### Password Security

**Server-Side**:
- Bcrypt hashing (cost factor: 10)
- Salted automatically (bcrypt includes salt)
- Never store plaintext passwords
- Never log passwords

**Client-Side**:
- HTTPS only (no plaintext transmission)
- Don't cache passwords
- Clear password fields after login
- Use password managers

### Token Security

**Access Token** (15 minutes):
- Short-lived (reduces impact of theft)
- Stateless (no server-side storage)
- Includes minimal claims

**Refresh Token** (7 days):
- Long-lived (better UX)
- Stored securely (HTTP-only cookie)
- Rotated on each refresh
- Can be revoked

### HTTPS in Production

**Always use HTTPS** in production:

```yaml
# config/app.postgres-prod.yaml
server:
  host: "0.0.0.0"
  port: 443
  ssl:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
```

---

## Common Authentication Errors

### 401 Unauthorized - Missing Token

**Error**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "missing authorization header"
  }
}
```

**Solution**: Include `Authorization: Bearer <token>` header

---

### 401 Unauthorized - Invalid Token

**Error**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid token signature"
  }
}
```

**Solutions**:
- Check token format (must be valid JWT)
- Token might be corrupted
- Login again to get new token

---

### 401 Unauthorized - Expired Token

**Error**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "token is expired"
  }
}
```

**Solution**: Use refresh token to get new access token

```bash
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"'"$REFRESH_TOKEN"'"}'
```

---

### 401 Unauthorized - Revoked Token

**Error**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "token has been revoked"
  }
}
```

**Solution**: Login again (token was invalidated by logout)

---

### 403 Forbidden - Insufficient Permissions

**Error**:
```json
{
  "status": "error",
  "error": {
    "code": "FORBIDDEN",
    "message": "admin role required"
  }
}
```

**Solutions**:
- Check user roles: `GET /api/v1/identity/users/me`
- Request admin access from system administrator
- Use account with appropriate role

---

### 429 Too Many Requests - Rate Limit

**Error**:
```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "too many login attempts, retry after 60 seconds"
  }
}
```

**Rate Limits**:
- **Login**: 5 requests/minute per IP
- **Register**: 3 requests/minute per IP

**Solution**: Wait for rate limit window to expire (60 seconds)

---

## Advanced Topics

### Multi-Factor Authentication (MFA)

**Status**: Planned for Q2 2026

**Planned Features**:
- TOTP (Time-based One-Time Password)
- SMS verification
- Email verification codes
- Backup codes

---

### OAuth2 Integration

**Status**: Planned for Q3 2026

**Planned Providers**:
- Google
- GitHub
- Microsoft
- Facebook

---

### API Keys for Machine-to-Machine

**Status**: Planned for Q2 2026

**Use Cases**:
- Server-to-server communication
- CI/CD pipelines
- Webhooks
- Third-party integrations

---

## Testing Authentication

### Unit Tests

**Test Login Handler**:
```go
func TestUserHandler_Login_Success(t *testing.T) {
    // Mock UseCase
    mockUC := &MockUserUseCase{
        AuthenticateFunc: func(ctx context.Context, email, password string) (*user.User, error) {
            return &user.User{
                ID:    uuidv7.New(),
                Email: email,
                Name:  "John Doe",
            }, nil
        },
    }
    
    // Create handler
    handler := NewUserHandler(mockUC, jwtManager)
    
    // Test request
    router := gin.New()
    router.POST("/login", handler.Login)
    
    body := `{"email":"john@example.com","password":"SecurePass123"}`
    req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, 200, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "success", response["status"])
    assert.NotEmpty(t, response["data"].(map[string]interface{})["access_token"])
}
```

### Integration Tests

**Test Full Authentication Flow**:
```go
func TestAuthenticationFlow(t *testing.T) {
    db := integration.SetupTestDB(t)
    defer db.Close()
    
    // 1. Register
    registerReq := `{"email":"test@example.com","name":"Test","password":"Pass123"}`
    registerResp := testRequest(t, "POST", "/register", registerReq, "")
    assert.Equal(t, 201, registerResp.Code)
    
    // 2. Login
    loginReq := `{"email":"test@example.com","password":"Pass123"}`
    loginResp := testRequest(t, "POST", "/login", loginReq, "")
    assert.Equal(t, 200, loginResp.Code)
    
    var loginData map[string]interface{}
    json.Unmarshal(loginResp.Body.Bytes(), &loginData)
    token := loginData["data"].(map[string]interface{})["access_token"].(string)
    
    // 3. Access protected endpoint
    protectedResp := testRequest(t, "GET", "/users/me", "", token)
    assert.Equal(t, 200, protectedResp.Code)
    
    // 4. Logout
    logoutResp := testRequest(t, "POST", "/logout", "", token)
    assert.Equal(t, 200, logoutResp.Code)
    
    // 5. Try to use revoked token
    revokedResp := testRequest(t, "GET", "/users/me", "", token)
    assert.Equal(t, 401, revokedResp.Code)
}
```

---

## Related Documentation

- **[JWT Package README](../../pkg/jwt/README.md)** - Technical JWT implementation details
- **[RBAC Guide](rbac.md)** - Complete RBAC documentation
- **[Rate Limiting Guide](rate-limiting.md)** - Rate limiting configuration
- **[Quick Start Guide](quick-start.md)** - Getting started with curl examples
- **[API Reference](api-reference.md)** - Complete API documentation

---

## Summary

### Authentication Flow
1. **Register** → Create user account
2. **Login** → Get access + refresh tokens
3. **Use Token** → Include in `Authorization: Bearer <token>` header
4. **Refresh** → Get new tokens when access token expires
5. **Logout** → Revoke tokens

### Key Points
- Access tokens expire in **15 minutes**
- Refresh tokens last **7 days**
- Tokens rotated on each refresh
- RBAC with 5 roles and 29+ permissions
- Rate limiting on auth endpoints
- Secure password hashing (bcrypt)

### Best Practices
- Store tokens securely
- Always use HTTPS in production
- Refresh tokens before expiration
- Logout when done
- Check user roles for authorization

---

**Version**: 0.1.0  
**Last Updated**: January 5, 2026  
**Status**: Production-ready
