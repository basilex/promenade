# Gaps & Technical Debt - Action Plan

**Created:** December 28, 2025  
**Status:** Ready for implementation  
**Priority Order:** Critical → High → Medium

---

## 🔴 **CRITICAL (Day 1 - Tomorrow)**

### 1. RBAC Implementation ⚡ **HIGHEST PRIORITY**

**Current State:**
- JWT includes empty roles array: `roles := []string{} // TODO`
- Migration file missing: `migrations/core/000003_core_rbac_full.up.sql`
- Middleware exists but no data source

**Tasks:**
- [ ] Create RBAC migration (roles, permissions, user_roles, role_permissions tables)
- [ ] Add Role entity in Identity context
- [ ] Add Permission entity in Identity context
- [ ] Update User entity with roles relationship
- [ ] Implement Role repository
- [ ] Implement Permission repository
- [ ] Update Login handler to load user roles
- [ ] Add tests (unit + integration)

**Estimate:** 4-5 hours

**Files to create:**
```
migrations/identity/000006_rbac.up.sql
migrations/identity/000006_rbac.down.sql
internal/contexts/identity/role/
  entity.go
  repository.go
  usecase.go
  adapter/repository/postgres/role_repository.go
internal/contexts/identity/permission/
  entity.go
  repository.go
```

**SQL Schema:**
```sql
-- roles table
CREATE TABLE identity_roles (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- permissions table
CREATE TABLE identity_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- user_roles junction table
CREATE TABLE identity_user_roles (
    user_id UUID NOT NULL REFERENCES identity_users(id),
    role_id UUID NOT NULL REFERENCES identity_roles(id),
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID REFERENCES identity_users(id),
    PRIMARY KEY (user_id, role_id)
);

-- role_permissions junction table
CREATE TABLE identity_role_permissions (
    role_id UUID NOT NULL REFERENCES identity_roles(id),
    permission_id UUID NOT NULL REFERENCES identity_permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

-- Default roles seed data
INSERT INTO identity_roles (id, name, description, is_system) VALUES
    (uuid_v7(), 'admin', 'System administrator', true),
    (uuid_v7(), 'user', 'Regular user', true),
    (uuid_v7(), 'manager', 'Manager role', true);
```

---

### 2. JWT Secret Validation 🔒

**Current State:**
- Default dev secret can be used in production
- No validation on startup

**Tasks:**
- [ ] Add JWT secret validation in `cmd/api/main.go`
- [ ] Check minimum length (32 chars)
- [ ] Check for default dev secret in production
- [ ] Add environment variable example in docs

**Estimate:** 30 minutes

**Code to add:**
```go
// cmd/api/main.go after config load
if cfg.App.Environment == "production" {
    if len(cfg.JWT.Secret) < 32 {
        logger.Fatal("JWT secret must be at least 32 characters in production")
    }
    if cfg.JWT.Secret == "dev-secret-key-change-in-production" {
        logger.Fatal("Default JWT secret cannot be used in production")
    }
}
```

---

### 3. Rate Limiting for Authentication 🚦

**Current State:**
- No rate limiting on Login/Register endpoints
- Vulnerable to brute-force attacks
- Account lockout works but doesn't stop attack

**Tasks:**
- [ ] Add rate limiting middleware (use `golang.org/x/time/rate`)
- [ ] IP-based limiting for Login (5 attempts per minute)
- [ ] IP-based limiting for Register (3 attempts per minute)
- [ ] Add rate limit headers in response
- [ ] Add tests

**Estimate:** 2-3 hours

**Package to use:**
```bash
go get golang.org/x/time/rate
```

**Implementation:**
```go
// pkg/middleware/ratelimit.go
type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
    return &RateLimiter{
        visitors: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        
        rl.mu.Lock()
        limiter, exists := rl.visitors[ip]
        if !exists {
            limiter = rate.NewLimiter(rl.rate, rl.burst)
            rl.visitors[ip] = limiter
        }
        rl.mu.Unlock()
        
        if !limiter.Allow() {
            response.Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests")
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## 🟡 **HIGH PRIORITY (Day 2-3)**

### 4. Token Revocation Mechanism 🔐

**Current State:**
- JWT tokens cannot be revoked before expiry
- No blacklist for compromised tokens

**Tasks:**
- [ ] Add Redis blacklist for revoked tokens
- [ ] Create token revocation endpoint
- [ ] Update JWT middleware to check blacklist
- [ ] Add TTL matching token expiry
- [ ] Add tests

**Estimate:** 3 hours

---

### 5. Health Checks for Dependencies ❤️

**Current State:**
- `/health` endpoint exists but doesn't check dependencies

**Tasks:**
- [ ] Add `/health/db` - PostgreSQL check
- [ ] Add `/health/redis` - Redis check (if used)
- [ ] Add `/health/bus` - Event Bus check
- [ ] Return proper status codes (200/503)
- [ ] Add timeout for checks

**Estimate:** 1 hour

---

### 6. Database Indexes Audit 📊

**Current State:**
- Not all tables have proper indexes
- Potential slow queries

**Tasks:**
- [ ] Audit all queries for missing indexes
- [ ] Add indexes on foreign keys
- [ ] Add composite indexes for common queries
- [ ] Create migration with indexes

**Estimate:** 2 hours

**Queries to analyze:**
```sql
-- Check missing indexes
SELECT * FROM users WHERE email = ?  -- UNIQUE INDEX exists ✅
SELECT * FROM contacts WHERE user_id = ?  -- Need INDEX
SELECT * FROM profiles WHERE user_id = ?  -- Need UNIQUE INDEX
SELECT * FROM customers WHERE email = ?  -- Need INDEX
SELECT * FROM customers WHERE status = ?  -- Need INDEX
```

---

### 7. Error Definitions Consistency 📝

**Current State:**
- User: errors in usecase.go
- Customer: separate errors.go file
- Profile: inline error strings

**Tasks:**
- [ ] Create `errors.go` in each aggregate
- [ ] Move all domain errors to errors.go
- [ ] Standardize error naming
- [ ] Update imports

**Estimate:** 1 hour

---

## 🟢 **MEDIUM PRIORITY (Week 2)**

### 8. Panic Handling Improvements ⚠️

**Current State:**
- `MustGetClaims()` panics if claims not found
- Should return error instead

**Tasks:**
- [ ] Refactor `MustGetClaims()` to return error
- [ ] Refactor `MustGetUserID()` to return error
- [ ] Update all callers
- [ ] Add tests

**Estimate:** 1 hour

---

### 9. Soft Delete Query Audit 🗑️

**Current State:**
- Most queries include `deleted_at IS NULL` ✅
- Need to verify ALL queries

**Tasks:**
- [ ] Audit all SELECT queries
- [ ] Add deleted_at check where missing
- [ ] Consider DB view for active records
- [ ] Add integration tests

**Estimate:** 2 hours

---

### 10. Caching Layer (Redis) 💾

**Current State:**
- No caching implemented
- Reference data queries repeated

**Tasks:**
- [ ] Add Redis cache adapter
- [ ] Cache Countries, Currencies, Languages, Timezones
- [ ] Cache User profiles (with TTL)
- [ ] Add cache invalidation
- [ ] Add tests

**Estimate:** 4 hours

---

### 11. N+1 Query Optimization 🚀

**Current State:**
- Potential N+1 in ListUsers + Profiles
- No eager loading

**Tasks:**
- [ ] Analyze queries with EXPLAIN
- [ ] Add JOIN queries where needed
- [ ] Consider Dataloader pattern
- [ ] Benchmark performance

**Estimate:** 3 hours

---

### 12. CSRF Protection 🛡️

**Current State:**
- Using Bearer tokens (safe) ✅
- If cookies added → need CSRF

**Tasks:**
- [ ] Add CSRF middleware (if needed)
- [ ] Token generation/validation
- [ ] Add to forms
- [ ] Tests

**Status:** Not needed now (Bearer auth), but document for future

---

## 📊 **Summary Statistics**

| Priority | Tasks | Estimated Time | Status |
|----------|-------|----------------|--------|
| 🔴 Critical | 3 | 7-9 hours | **START TOMORROW** |
| 🟡 High | 4 | 9 hours | Week 1-2 |
| 🟢 Medium | 5 | 11 hours | Week 2-3 |
| **TOTAL** | **12** | **27-29 hours** | ~4 working days |

---

## 🎯 **Tomorrow's Action Plan (December 29)**

### Morning (3-4 hours)
1. ⚡ **RBAC Implementation** (PRIORITY #1)
   - Create migration
   - Add Role & Permission entities
   - Update User with roles
   - Load roles in Login handler

### Afternoon (2-3 hours)
2. 🔒 **JWT Secret Validation** (30 min)
3. 🚦 **Rate Limiting** (2 hours)
4. ✅ **Testing all changes** (30 min)

### End of Day
5. 📝 **Commit & Push**
6. ✅ **Update this TODO list**

---

## 📝 **Notes**

- All changes should include tests (unit + integration)
- Update README.md if public API changes
- Keep backwards compatibility where possible
- Document breaking changes in CHANGELOG.md

---

## 🔗 **Related Documents**

- [Clean Architecture Summary](CLEAN_ARCHITECTURE_SUMMARY.md)
- [Testing Patterns](TESTING_PATTERNS.md)
- [AI Instructions](../.github/copilot-instructions.md)

---

**Last Updated:** December 28, 2025  
**Next Review:** After Day 1 completion
