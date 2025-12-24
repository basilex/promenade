# Authentication System Architecture

Complete reference for authentication system in Promenade. Covers user registration, login, session management, token handling, and security mechanisms.

## Table of Contents

- [Overview](#overview)
- [Database Schema](#database-schema)
- [User States & Lifecycle](#user-states--lifecycle)
- [Authentication Flow](#authentication-flow)
- [Session Management](#session-management)
- [Token System](#token-system)
- [Security Mechanisms](#security-mechanisms)
- [API Endpoints](#api-endpoints)
- [Error Handling](#error-handling)
- [Configuration](#configuration)

---

## Overview

**Authentication System Components:**

| Component       | Purpose                                 | Technology          |
| --------------- | --------------------------------------- | ------------------- |
| User Entity     | Core user account with credentials      | PostgreSQL, bcrypt  |
| Session Entity  | Refresh token storage                   | PostgreSQL, SHA-256 |
| JWT Manager     | Access token generation/validation      | HMAC-SHA256         |
| Auth UseCase    | Business logic for all auth operations  | Go                  |
| Auth Middleware | Request authentication                  | Gin middleware      |
| Event Bus       | Async notifications (email, audit logs) | Memory/Redis        |

**Key Features:**

-  JWT-based authentication (access + refresh tokens)
-  Refresh token rotation (security best practice)
-  Concurrent session management (max 5 per user)
-  User state machine (unverified → active → suspended/banned)
-  Password hashing with bcrypt (cost 10)
-  Refresh token hashing with SHA-256
-  Automatic session cleanup on state changes
-  Async email notifications via event bus

---

## Database Schema

### `core_users` Table

```sql
CREATE TABLE core_users (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    email              VARCHAR(255) NOT NULL UNIQUE,
    name               VARCHAR(255) NOT NULL,
    password           VARCHAR(255) NOT NULL,  -- bcrypt hash
    status             VARCHAR(20) NOT NULL DEFAULT 'unverified',
    email_verified_at  TIMESTAMPTZ,
    suspended_reason   TEXT,
    suspended_until    TIMESTAMPTZ,
    last_login_at      TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON core_users(email);
CREATE INDEX idx_users_status ON core_users(status);
```

### `core_user_sessions` Table

```sql
CREATE TABLE core_user_sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id       UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,  -- SHA-256 hash
    user_agent    TEXT,
    ip_address    VARCHAR(45),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON core_user_sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON core_user_sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON core_user_sessions(expires_at);
```

**Additional Tables (not yet implemented):**

- `core_password_reset_tokens` - Password reset workflow
- `core_email_verification_tokens` - Email verification workflow
- `core_login_attempts` - Brute force protection

---

## User States & Lifecycle

### User Status Enum

```go
type UserStatus string

const (
    UserStatusUnverified UserStatus = "unverified" // Registered but email not verified
    UserStatusActive     UserStatus = "active"     // Email verified and account active
    UserStatusSuspended  UserStatus = "suspended"  // Temporarily blocked (can be reactivated)
    UserStatusBanned     UserStatus = "banned"     // Permanently blocked
    UserStatusInactive   UserStatus = "inactive"   // Deactivated by user (can be reactivated)
)
```

### State Transitions

```
                    Register()
                        │
                        ▼
                 ──────────────
                 │  unverified  │  ──────────────
                 └──────────────                │
                        │                        │
                 VerifyEmail()              Login() allowed
                        │                        │
                        ▼                        ▼
                 ──────────────         User can login
                 │    active    │         (unverified or active)
                 └──────────────
                    │   │   │
        ───────────   │   └───────────
   Suspend()      Ban()           Deactivate()
        │               │                │
        ▼               ▼                ▼
 ───────────   ──────────    ─────────────
 │ suspended │   │  banned  │    │  inactive   │
 └───────────   └──────────    └─────────────
        │                              │
   Reactivate()                   Reactivate()
        │                              │
        └─────────────────────────────
                     │
                     ▼
              ──────────────
              │    active    │
              └──────────────
```

### CanLogin() Logic

```go
func (u *User) CanLogin() bool {
    return u.Status == UserStatusActive || u.Status == UserStatusUnverified
}
```

**Allowed States:**

-  `active` - Full access
-  `unverified` - Can login but features may be restricted

**Blocked States:**

-  `suspended` - Returns `ErrUserSuspended`
-  `banned` - Returns `ErrUserBanned`
-  `inactive` - Returns `ErrUserNotActive`

---

## Authentication Flow

### 1. Registration Flow

```
Client                  API                    UseCase                Database           Event Bus
  │                      │                        │                        │                 │
  │  POST /auth/register │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │                      │  Register(email, name, pwd)                     │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  (not found - OK)      │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Hash(pwd)      │                 │
  │                      │                        │  user.Status = "unverified"              │
  │                      │                        │                        │                 │
  │                      │                        │  Create(user)          │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │                        │                 │
  │                      │                        │  Publish(UserRegisteredEvent)            │
  │                      │                        │─────────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  201 Created         │                        │                        │  EmailWorker    │
  │<─────────────────────│                        │                        │  sends welcome  │
  │  {id, email, name}   │                        │                        │  email (async)  │
```

**Key Points:**

- Password hashed with `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)`
- User created with `status = "unverified"`
- UUID v7 generated for user ID (time-ordered)
- Event published asynchronously - registration doesn't wait for email
- Email worker handles `user.registered` event in background

---

### 2. Login Flow

```
Client                  API                    UseCase                Database           Session
  │                      │                        │                        │                 │
  │  POST /auth/login    │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {email, password}   │  Login(email, pwd, ua, ip)                      │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  user found            │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Compare(pwd, hash)               │
  │                      │                        │  OK                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  YES                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  JWTManager.GenerateAccessToken()        │
  │                      │                        │  crypto/rand 32 bytes for refresh token  │
  │                      │                        │  SHA-256(refresh_token)                  │
  │                      │                        │                        │                 │
  │                      │                        │  CountUserSessions(user_id)              │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  count = 5 (limit!)    │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetOldestSession()    │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  Delete(oldest)        │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │                        │                 │
  │                      │                        │  Create(new session)   │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │                        │  UpdateLastLogin()     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (access_token, refresh_token, user)            │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token,     │                        │                        │                 │
  │   user: {...}}       │                        │                        │                 │
```

**Key Points:**

- **Password validation:** `bcrypt.CompareHashAndPassword(hash, password)`
- **User state check:** Must be `active` or `unverified`
- **Access token:** JWT with 15-minute TTL (default)
- **Refresh token:** 32-byte random + base64, 7-day TTL (default)
- **Refresh token storage:** Hashed with SHA-256 before saving to database
- **Session limit:** Max 5 concurrent sessions per user
- **Oldest session removal:** If limit exceeded, oldest session is deleted automatically
- **Last login:** Updated asynchronously (doesn't block response)

---

### 3. Token Refresh Flow

```
Client                  API                    UseCase                Database           Session
  │                      │                        │                        │                 │
  │  POST /auth/refresh  │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {refresh_token}     │  RefreshToken(token)   │                        │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  SHA-256(token)        │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByRefreshToken(hash)                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │<───────────────────────────────────────│
  │                      │                        │  session found         │                 │
  │                      │                        │                        │                 │
  │                      │                        │  session.IsExpired()?  │                 │
  │                      │                        │  NO                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByID(session.user_id)                │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  user found            │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  YES                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GenerateAccessToken() │                 │
  │                      │                        │  crypto/rand new refresh                 │
  │                      │                        │  SHA-256(new_refresh)  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  Update(session)       │  ← TOKEN ROTATION
  │                      │                        │  - new refresh hash    │                 │
  │                      │                        │  - new expires_at      │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (new_access, new_refresh)                      │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token}     │                        │                        │                 │
```

**Key Points:**

- **Token Rotation:** Old refresh token is invalidated, new one issued
- **Security:** Single-use refresh tokens prevent replay attacks
- **Session reuse:** Updates existing session instead of delete+create (performance)
- **Expiration check:** Expired sessions auto-deleted
- **User state check:** User must still be able to login

---

### 4. Logout Flow

```
Client                  API                    UseCase                Session
  │                      │                        │                        │
  │  POST /auth/logout   │                        │                        │
  │─────────────────────>│                        │                        │
  │  {refresh_token}     │  Logout(token)         │                        │
  │                      │───────────────────────>│                        │
  │                      │                        │  SHA-256(token)        │
  │                      │                        │                        │
  │                      │                        │  GetByRefreshToken(hash)
  │                      │                        │───────────────────────>│
  │                      │                        │<───────────────────────│
  │                      │                        │  session found         │
  │                      │                        │                        │
  │                      │                        │  Delete(session.id)    │
  │                      │                        │───────────────────────>│
  │                      │                        │                        │
  │                      │<───────────────────────│                        │
  │  200 OK              │                        │                        │
  │<─────────────────────│                        │                        │
  │  {message: "success"}│                        │                        │
```

**Key Points:**

- **Session deletion:** Immediately invalidates refresh token
- **Access token:** Still valid until expiration (stateless JWT)
- **Client responsibility:** Client should discard both tokens

---

## Session Management

### Concurrent Session Limit

```go
const MaxConcurrentSessions = 5
```

**Behavior:**

- User can have maximum 5 active sessions across devices
- On 6th login: oldest session is automatically deleted
- Prevents unlimited token generation attacks

### Session Entity

```go
type Session struct {
    ID           UUID      `db:"id"`
    UserID       UUID      `db:"user_id"`
    RefreshToken string    `db:"refresh_token"`  // SHA-256 hashed
    UserAgent    *string   `db:"user_agent"`     // Browser/device info
    IPAddress    *string   `db:"ip_address"`     // Client IP
    ExpiresAt    time.Time `db:"expires_at"`     // Absolute expiration
    CreatedAt    time.Time `db:"created_at"`     // Session start
}
```

### Session Operations

| Operation         | Purpose                      | Trigger                     |
| ----------------- | ---------------------------- | --------------------------- |
| Create            | New session on login         | Login                       |
| Update            | Rotate refresh token         | Token refresh               |
| Delete            | Logout single session        | Logout                      |
| DeleteByUserID    | Invalidate all user sessions | Suspend/Ban/Password change |
| GetOldestSession  | Find oldest for deletion     | Session limit exceeded      |
| CountUserSessions | Check against limit          | Login                       |

---

## Token System

### JWT Access Token

**Properties:**

- **Algorithm:** HS256 (HMAC-SHA256)
- **TTL:** 15 minutes (default, configurable)
- **Storage:** Client-side only (not in database)
- **Validation:** Signature + expiration checked on each request

**Claims Structure:**

```go
type Claims struct {
    UserID uuidv7.UUID `json:"user_id"`
    Email  string      `json:"email"`
    jwt.RegisteredClaims
}

// RegisteredClaims includes:
// - iat (issued at)
// - exp (expires at)
// - nbf (not before)
```

**Example JWT:**

```json
{
  "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
  "email": "user@example.com",
  "iat": 1703347200,
  "exp": 1703348100,
  "nbf": 1703347200
}
```

### Refresh Token

**Properties:**

- **Algorithm:** crypto/rand (32 bytes) + base64
- **TTL:** 7 days (default, configurable)
- **Storage:** Database (SHA-256 hashed)
- **Validation:** Hash comparison + expiration

**Generation:**

```go
func generateRefreshToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

**Hashing (SHA-256):**

```go
func hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return base64.URLEncoding.EncodeToString(hash[:])
}
```

### Token Rotation Pattern

**Security Benefit:** Prevents refresh token reuse attacks

1. Client sends refresh token
2. Server validates and issues new token pair
3. **Old refresh token is invalidated immediately**
4. Client must use new refresh token for next refresh

**Attack Scenario Prevention:**

-  Attacker steals refresh token
-  Attacker tries to use it
-  Token already rotated by legitimate user → **Attack fails**

---

## Security Mechanisms

### Password Security

| Mechanism  | Implementation                    | Purpose                  |
| ---------- | --------------------------------- | ------------------------ |
| Hashing    | `bcrypt` (cost 10)                | One-way encryption       |
| Salt       | Automatic (bcrypt internal)       | Unique hash per password |
| Validation | `bcrypt.CompareHashAndPassword()` | Constant-time comparison |

**Code:**

```go
// Hash on registration
func (u *User) HashPassword(password string) error {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedBytes)
    return nil
}

// Verify on login
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

### Refresh Token Security

| Mechanism         | Implementation           | Purpose                      |
| ----------------- | ------------------------ | ---------------------------- |
| Random generation | `crypto/rand` (32 bytes) | Cryptographically secure     |
| Hashing           | SHA-256                  | Never store plaintext tokens |
| Token rotation    | Single-use tokens        | Invalidate after use         |
| Expiration        | 7-day TTL                | Limit exposure window        |

### JWT Security

| Mechanism  | Implementation  | Purpose                   |
| ---------- | --------------- | ------------------------- |
| Signature  | HMAC-SHA256     | Tamper protection         |
| Secret key | Environment var | Signature verification    |
| Short TTL  | 15 minutes      | Minimize exposure window  |
| Stateless  | No DB lookup    | Performance + scalability |

### Admin Actions Security

**Automatic Session Invalidation:**

When admin suspends/bans user:

1. User status changed in database
2. **All user sessions deleted immediately**
3. User cannot refresh tokens
4. Existing access tokens expire naturally (15 min max)

```go
func (uc *authUseCase) SuspendUser(ctx context.Context, userID UUID, reason string, until *time.Time) error {
    // Update user status
    if err := uc.userRepo.Suspend(ctx, userID, reason, until); err != nil {
        return err
    }

    // CRITICAL: Invalidate all sessions
    if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
        return err
    }

    // Async notification
    uc.eventBus.Publish(ctx, bus.TopicUserSuspended, event)
    return nil
}
```

---

## API Endpoints

### Public Endpoints (No Authentication)

#### POST /api/v1/auth/register

Register new user account.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Response (201 Created):**

```json
{
  "status": "success",
  "data": {
    "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "unverified",
    "email_verified_at": null,
    "last_login_at": null,
    "created_at": "2024-12-23T10:00:00Z",
    "updated_at": "2024-12-23T10:00:00Z"
  }
}
```

---

#### POST /api/v1/auth/login

Authenticate user and receive tokens.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "email": "user@example.com",
      "name": "John Doe",
      "status": "active"
    }
  }
}
```

---

#### POST /api/v1/auth/refresh

Refresh access token using refresh token.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "bmV3cmVmcmVzaHRva2VuaGVyZQ...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

**Note:** Old refresh token is invalidated (token rotation).

---

### Protected Endpoints (Authentication Required)

#### GET /api/v1/auth/me

Get current user profile.

**Request:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": {
    "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "active",
    "email_verified_at": "2024-12-20T14:30:00Z",
    "last_login_at": "2024-12-23T10:00:00Z",
    "created_at": "2024-12-15T09:00:00Z",
    "updated_at": "2024-12-23T10:00:00Z"
  }
}
```

---

#### GET /api/v1/auth/sessions

Get all active sessions for current user.

**Request:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/sessions \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": [
    {
      "id": "01936d6a-9999-7890-a1b2-c3d4e5f67890",
      "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...",
      "ip_address": "192.168.1.100",
      "expires_at": "2024-12-30T10:00:00Z",
      "created_at": "2024-12-23T10:00:00Z"
    },
    {
      "id": "01936d6a-8888-7890-a1b2-c3d4e5f67890",
      "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "user_agent": "Mobile Safari/537.36",
      "ip_address": "192.168.1.101",
      "expires_at": "2024-12-29T15:30:00Z",
      "created_at": "2024-12-22T15:30:00Z"
    }
  ]
}
```

---

#### POST /api/v1/auth/logout

Logout and invalidate refresh token.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "Successfully logged out"
  }
}
```

---

### Admin Endpoints (Require `users:suspend` / `users:ban` permissions)

#### POST /api/v1/auth/users/:id/suspend

Temporarily suspend user account.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/suspend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Violation of community guidelines",
    "suspended_until": "2025-01-23T00:00:00Z"
  }'
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User suspended successfully"
  }
}
```

**Effects:**

- User status changed to `suspended`
- **All active sessions deleted immediately**
- User cannot login until `suspended_until` date or manual reactivation

---

#### POST /api/v1/auth/users/:id/ban

Permanently ban user account.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/ban \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Spam and fraudulent activity"
  }'
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User banned successfully"
  }
}
```

**Effects:**

- User status changed to `banned`
- **All active sessions deleted immediately**
- User cannot login (requires manual reactivation by admin)

---

#### POST /api/v1/auth/users/:id/reactivate

Reactivate suspended/banned/inactive user.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/reactivate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User reactivated successfully"
  }
}
```

---

## Error Handling

### Authentication Errors

| Error Code            | HTTP Status | Message                    | Cause                            |
| --------------------- | ----------- | -------------------------- | -------------------------------- |
| `INVALID_CREDENTIALS` | 401         | Invalid email or password  | Wrong email/password combination |
| `EMAIL_EXISTS`        | 409         | Email already exists       | Registration with existing email |
| `USER_NOT_ACTIVE`     | 403         | User account is not active | Status is `inactive`             |
| `USER_SUSPENDED`      | 403         | User account is suspended  | Status is `suspended`            |
| `USER_BANNED`         | 403         | User account is banned     | Status is `banned`               |
| `INVALID_TOKEN`       | 401         | Invalid or expired token   | Token validation failed          |
| `TOKEN_EXPIRED`       | 401         | Token has expired          | JWT or refresh token expired     |
| `UNAUTHORIZED`        | 401         | Authentication required    | Missing or invalid Authorization |

### Error Response Format

```json
{
  "status": "error",
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password",
    "details": null
  }
}
```

### Validation Errors

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email address"
      },
      {
        "field": "password",
        "message": "must be at least 8 characters"
      }
    ]
  }
}
```

---

## Configuration

### Environment Variables

```bash
# JWT Configuration
JWT_SECRET=your-256-bit-secret-key-here-change-in-production
JWT_ACCESS_TOKEN_TTL=15m    # Access token lifetime (e.g., 15m, 1h)
JWT_REFRESH_TOKEN_TTL=168h  # Refresh token lifetime (e.g., 168h = 7 days)

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=promenade
DB_SSLMODE=disable

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

### Config File (config/app.dev.yaml)

```yaml
jwt:
  secret: ${JWT_SECRET}
  access_token_ttl: 15m
  refresh_token_ttl: 168h

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  name: ${DB_NAME}
  sslmode: ${DB_SSLMODE}
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

server:
  port: ${SERVER_PORT}
  host: ${SERVER_HOST}
  read_timeout: 10s
  write_timeout: 10s
```

### Security Best Practices

| Setting                 | Development   | Production                 |
| ----------------------- | ------------- | -------------------------- |
| `JWT_SECRET`            | Any string    | **256-bit random string**  |
| `JWT_ACCESS_TOKEN_TTL`  | 15m           | 5m - 15m                   |
| `JWT_REFRESH_TOKEN_TTL` | 168h (7 days) | 7-30 days                  |
| `DB_SSLMODE`            | disable       | **require or verify-full** |
| `SERVER_HOST`           | 0.0.0.0       | 0.0.0.0 or specific IP     |

**Critical Production Settings:**

1. Generate secure JWT secret: `openssl rand -base64 32`
2. Use HTTPS only (TLS/SSL certificates)
3. Enable database SSL (`DB_SSLMODE=require`)
4. Set short access token TTL (5-15 minutes)
5. Monitor failed login attempts (rate limiting)

---

## Related Documentation

- [AUTHORIZATION.md](AUTHORIZATION.md) - RBAC permission system (what happens AFTER authentication)
- [CREDENTIALS.md](CREDENTIALS.md) - Default users and roles for development/testing
- [LOGGING.md](LOGGING.md) - Structured logging with authentication context
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) - System architecture and core modules
