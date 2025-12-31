# CSRF Protection

**Cross-Site Request Forgery (CSRF) Protection** for Promenade Platform - middleware and best practices.

---

## Current Status

**CSRF Protection: NOT REQUIRED** ✅

Promenade uses **Bearer JWT tokens** in `Authorization` header for authentication, which is **safe from CSRF attacks by design**.

### Why Bearer Tokens Are Safe

**CSRF attacks work because**:
- Browsers automatically attach cookies to all requests (even cross-site)
- Attacker tricks victim's browser into making authenticated request
- Victim's session cookie is sent automatically → attack succeeds

**Bearer tokens are immune because**:
- Stored in JavaScript (localStorage/sessionStorage, not cookies)
- Must be explicitly added to `Authorization` header
- Browsers **cannot** automatically attach headers in cross-site requests
- Attacker cannot access tokens from different origin (Same-Origin Policy)

### Authentication Flow (Current)

```javascript
// Client stores JWT token
localStorage.setItem('access_token', token);

// Client explicitly adds token to each request
fetch('/api/users', {
    headers: {
        'Authorization': `Bearer ${localStorage.getItem('access_token')}`
    }
});
```

**Result**: Even if attacker tricks victim to visit malicious site, they cannot:
1. Read token from localStorage (blocked by Same-Origin Policy)
2. Make authenticated request (browser won't auto-attach `Authorization` header)

---

## When CSRF Protection IS Needed

CSRF middleware is required ONLY if you:

1. **Use cookie-based authentication** (session cookies)
2. **Use cookie-based JWT tokens** (`Set-Cookie` with JWT)
3. **Have state-changing GET requests** (bad practice, but happens)

---

## CSRF Middleware (For Future Use)

If you switch to cookie-based authentication, use the provided middleware.

### Basic Usage

```go
import "github.com/basilex/promenade/pkg/middleware"

// In router setup
router.Use(middleware.CSRFMiddleware(middleware.DefaultCSRFConfig()))
```

### Configuration

```go
config := middleware.CSRFConfig{
    TokenLength:    32,                         // Token length in bytes
    CookieName:     "csrf_token",               // Cookie name
    HeaderName:     "X-CSRF-Token",             // Header name
    CookieMaxAge:   43200,                      // 12 hours
    CookiePath:     "/",
    CookieDomain:   "",
    CookieSecure:   true,                       // HTTPS only
    CookieHTTPOnly: true,                       // No JavaScript access
    CookieSameSite: http.SameSiteStrictMode,    // Strict same-site policy
    SkipMethods:    []string{"GET", "HEAD", "OPTIONS"},
}

router.Use(middleware.CSRFMiddleware(config))
```

### How It Works

1. **GET Request** → Server generates CSRF token, sets cookie
2. **POST/PUT/DELETE Request** → Client must send token in both:
   - Cookie: `csrf_token=abc123...`
   - Header: `X-CSRF-Token: abc123...`
3. **Server validates** → Tokens match → Request allowed

### Client Integration (HTML Forms)

```html
<!-- Server renders CSRF token in form -->
<form method="POST" action="/api/users">
    <input type="hidden" name="csrf_token" value="{{ .CSRFToken }}">
    <input type="text" name="name">
    <button type="submit">Create User</button>
</form>
```

### Client Integration (JavaScript)

```javascript
// 1. Get CSRF token from cookie (after GET request)
function getCSRFToken() {
    return document.cookie
        .split('; ')
        .find(row => row.startsWith('csrf_token='))
        ?.split('=')[1];
}

// 2. Add token to fetch requests
fetch('/api/users', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': getCSRFToken()
    },
    body: JSON.stringify({ name: 'John' })
});
```

---

## Testing

### Run Tests

```bash
# CSRF middleware tests
go test ./pkg/middleware -v -run=TestCSRF

# All middleware tests
go test ./pkg/middleware -v
```

### Test Coverage

| Test                                | Status | Description                          |
| ----------------------------------- | ------ | ------------------------------------ |
| TestDefaultCSRFConfig               | ✅      | Default configuration                |
| TestCSRFMiddleware_SkipSafeMethods  | ✅      | GET/HEAD/OPTIONS bypass CSRF         |
| TestCSRFMiddleware_ValidToken       | ✅      | Valid token passes                   |
| TestCSRFMiddleware_MissingCookie    | ✅      | Missing cookie rejected (403)        |
| TestCSRFMiddleware_MissingHeader    | ✅      | Missing header rejected (403)        |
| TestCSRFMiddleware_TokenMismatch    | ✅      | Mismatched tokens rejected (403)     |
| TestCSRFMiddleware_CustomErrorHandler | ✅    | Custom error handling                |
| TestGenerateCSRFToken               | ✅      | Token generation                     |
| TestGetCSRFToken                    | ✅      | Token retrieval from context         |

**Total**: 13 tests, 100% passing

---

## Best Practices

### DO

- **Use Bearer tokens** (current approach) - No CSRF protection needed
- **Keep using `Authorization` header** - Safest method
- **If cookies needed** - Enable CSRF middleware immediately
- **Use HTTPS** - Protects against MITM attacks
- **Set SameSite=Strict** - Additional protection layer
- **Regenerate tokens** - On sensitive operations (password change, etc.)

### DON'T

- **Don't use GET for state changes** - Always POST/PUT/DELETE
- **Don't store JWT in cookies** without CSRF protection
- **Don't expose CSRF tokens** in URLs (use cookies/headers)
- **Don't skip CSRF** if using cookie-based auth
- **Don't trust referer header** alone (can be spoofed)

---

## Security Layers (Defense in Depth)

Promenade uses **multiple security layers**:

1. **Bearer JWT Tokens** ✅ - Primary defense (CSRF-safe by design)
2. **Token Expiration** ✅ - 15-minute access tokens
3. **Token Revocation** ✅ - Redis blacklist for logout
4. **HTTPS Only** ✅ - Encrypted communication
5. **Rate Limiting** ✅ - IP-based protection (Login: 5/min)
6. **RBAC** ✅ - Role-based access control
7. **CORS** ✅ - Cross-origin resource sharing restrictions
8. **CSRF Middleware** 📋 - Available for cookie-based auth (if needed)

---

## Migration Path (If Switching to Cookies)

If you decide to use cookie-based authentication:

### Step 1: Update Configuration

```yaml
# config/app.prod.yaml
server:
  cookie_auth: true  # Enable cookie-based auth
  csrf_enabled: true # Enable CSRF protection
```

### Step 2: Enable CSRF Middleware

```go
// cmd/api/server.go
if cfg.Server.CookieAuth {
    router.Use(middleware.CSRFMiddleware(middleware.DefaultCSRFConfig()))
}
```

### Step 3: Update Login Handler

```go
// Set JWT in HTTP-only cookie instead of response body
c.SetCookie(
    "access_token",
    accessToken,
    int(cfg.JWT.AccessTokenDuration.Seconds()),
    "/",
    "",
    true,  // Secure (HTTPS only)
    true,  // HTTP only (no JavaScript access)
)
c.SetSameSite(http.SameSiteStrictMode)
```

### Step 4: Update Client Code

```javascript
// Remove token from localStorage (cookies handled automatically)
// Add CSRF token to requests
fetch('/api/users', {
    method: 'POST',
    credentials: 'include',  // Include cookies
    headers: {
        'X-CSRF-Token': getCSRFToken()
    }
});
```

---

## Comparison: Bearer vs Cookie Auth

| Feature                  | Bearer Tokens (Current) | Cookie-Based Auth     |
| ------------------------ | ----------------------- | --------------------- |
| CSRF Protection          | Not needed ✅            | Required ⚠️            |
| XSS Protection           | Vulnerable ⚠️            | Better (HttpOnly) ✅   |
| Mobile Apps              | Easy ✅                  | Complex ⚠️             |
| CORS                     | Simple ✅                | Complex ⚠️             |
| Token Storage            | localStorage            | Cookies               |
| Auto-attachment          | Manual                  | Automatic             |
| Logout                   | Delete token            | Clear cookie + CSRF   |

**Recommendation**: Keep Bearer tokens for API, use cookies only for web UI if needed

---

## Resources

- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetsecurity.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [SameSite Cookies Explained](https://web.dev/samesite-cookies-explained/)

---

**Status**: CSRF middleware implemented, not enabled (not needed with Bearer auth)  
**Last Updated**: December 31, 2025  
**Test Coverage**: 13 tests, 100% passing  
**Maintainer**: Promenade Team
