# Auth Schema Documentation

## Overview

Migration `000002_create_auth_schema` creates a complete infrastructure for authentication and security.

## Tables

### 1. `users` - Users

**Main account table:**

```sql
- id: UUID v7 (primary key)
- email: unique email for login
- name: user name
- password: bcrypt password hash (cost 10-12)
- email_verified: email verification status
- email_verified_at: verification date
- active: account activity (for blocking/soft delete)
- last_login_at: last login time
```

**Use cases:**

- Registration/login
- Email verification
- Soft delete (via active = false)
- User blocking

### 2. `sessions` - Sessions (Refresh Tokens)

**JWT refresh tokens with metadata:**

```sql
- id: UUID v7
- user_id: FK → users
- refresh_token: hashed refresh token (512 chars)
- user_agent: browser/application
- ip_address: INET
- expires_at: expiration (typically 7-30 days)
```

**Use cases:**

- JWT token rotation
- Logout from specific device
- "Logout from all devices"
- Active session monitoring

**Example:**

```go
// Session created on login
session := &Session{
    UserID: user.ID,
    RefreshToken: hashToken(refreshToken),
    ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
    IPAddress: c.ClientIP(),
    UserAgent: c.Request.UserAgent(),
}
```

### 3. `password_reset_tokens` - Password Reset

**One-time tokens for password reset:**

```sql
- id: UUID v7
- user_id: FK → users
- token: hashed token (sent via email)
- expires_at: expiration (typically 1 hour)
- used: usage flag
- used_at: when used
```

**Flow:**

1. User requests reset → token created
2. Email sent with link
3. User clicks → token verified → password reset
4. Token marked as `used = true`

**Security:**

- Token lives for 1 hour
- One-time use (used flag)
- Old tokens automatically cleaned

### 4. `email_verification_tokens` - Email Verification

**One-time tokens for email confirmation:**

```sql
- id: UUID v7
- user_id: FK → users
- token: hashed token
- expires_at: expiration (typically 24 hours)
- used: usage flag
```

**Flow:**

1. On registration → token created
2. Email sent with confirmation link
3. User clicks → `users.email_verified = true`
4. Token marked as `used = true`

**Best practices:**

- Send on registration
- Resend if expired
- Block critical actions until verified

### 5. `login_attempts` - Login Attempts Log

**Security monitoring and rate limiting:**

```sql
- id: UUID v7
- email: login attempt with this email
- ip_address: INET
- success: successful/failed login
- user_agent: browser
- attempted_at: attempt time
```

**Use cases:**

- Rate limiting (5 failed attempts → 15 min lockout)
- Security alerts (suspicious activity)
- Audit log
- Login statistics

**Example logic:**

```go
// Check rate limit
failedAttempts := GetFailedAttempts(email, ipAddress, last15min)
if failedAttempts >= 5 {
    return ErrTooManyAttempts
}
```

## Functions

### `cleanup_expired_tokens()`

**Automatic cleanup:**

- Expired sessions
- Used tokens (older than 7 days)
- Expired unused tokens
- Old login attempts (older than 30 days)

**Run via cron:**

```sql
-- Daily at 3:00 AM
SELECT cleanup_expired_tokens();
```

## Indexes

**Optimized for frequent queries:**

```sql
-- Users
idx_users_email               -- Login lookup
idx_users_email_verified      -- Filter unverified users
idx_users_active              -- Active users

-- Sessions
idx_sessions_refresh_token    -- Token lookup (unique)
idx_sessions_user_id          -- User sessions
idx_sessions_expires_at       -- Cleanup expired

-- Password Reset
idx_password_reset_token      -- WHERE NOT used (partial)
idx_password_reset_user_id    -- User tokens

-- Login Attempts
idx_login_attempts_email      -- Rate limiting by email
idx_login_attempts_ip         -- Rate limiting by IP
```

## Security Best Practices

### 1. Password Hashing

```go
// Use bcrypt cost 10-12
hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
```

### 2. Token Generation

```go
// Cryptographically secure tokens
token := base64.URLEncoding.EncodeToString(randomBytes(32))
hashedToken := sha256.Sum256([]byte(token))
```

### 3. Rate Limiting

```go
// Check login_attempts before authentication
if FailedAttempts(email, ip) > 5 {
    return ErrTooManyAttempts // 429 HTTP
}
```

### 4. Session Security

```go
// Short JWT (15 min), long refresh (7 days)
accessTokenTTL := 15 * time.Minute
refreshTokenTTL := 7 * 24 * time.Hour
```

### 5. Email Verification

```go
// Block critical actions without verification
if !user.EmailVerified {
    return ErrEmailNotVerified
}
```

## Migration Commands

```bash
# Apply
make migrate-up

# Rollback
make migrate-down

# Check version
make migrate-version
```

## Query Examples

### Check if user exists and active

```sql
SELECT id, email, password, email_verified
FROM users
WHERE email = $1 AND active = true;
```

### Get user's active sessions

```sql
SELECT id, ip_address, user_agent, created_at
FROM sessions
WHERE user_id = $1 AND expires_at > NOW()
ORDER BY created_at DESC;
```

### Rate limit check

```sql
SELECT COUNT(*)
FROM login_attempts
WHERE (email = $1 OR ip_address = $2)
  AND success = false
  AND attempted_at > NOW() - INTERVAL '15 minutes';
```

### Verify email token

```sql
UPDATE email_verification_tokens
SET used = true, used_at = NOW()
WHERE token = $1 AND expires_at > NOW() AND used = false
RETURNING user_id;
```

## Monitoring Queries

### Active users count

```sql
SELECT COUNT(*) FROM users WHERE active = true;
```

### Unverified users

```sql
SELECT COUNT(*) FROM users
WHERE email_verified = false
  AND created_at > NOW() - INTERVAL '7 days';
```

### Active sessions

```sql
SELECT COUNT(*) FROM sessions WHERE expires_at > NOW();
```

### Failed login rate (last hour)

```sql
SELECT
    COUNT(*) FILTER (WHERE success = false) as failed,
    COUNT(*) FILTER (WHERE success = true) as success
FROM login_attempts
WHERE attempted_at > NOW() - INTERVAL '1 hour';
```

## Next Steps

After this migration you can:

1. Create RBAC schema (roles, permissions, user_roles)
2. Add OAuth providers (social login)
3. Add 2FA (two_factor_auth table)
4. User profiles (separate table)

See also: `docs/UUID_V7_GUIDE.md` for UUID v7 information.
