# Gaps & Technical Debt - Action Plan

**Created:** December 28, 2025  
**Last Updated:** December 29, 2025  
**Status:** Updated after Rate Limiting completion  
**Priority Order:** Critical → High → Medium

---

## COMPLETED TASKS

### 1. RBAC Implementation (COMPLETED December 29, 2025)

**Implementation:**
- Created migration `migrations/identity/000003_authorization.up.sql` with RBAC tables
- Added Role entity in Identity context with full CRUD
- Added Permission entity in Identity context with full CRUD
- Created PostgreSQL repositories for Role and Permission
- Implemented use cases with business logic
- Created HTTP handlers with 14 JWT-protected endpoints (7 Role + 7 Permission)
- Added comprehensive tests (unit + smoke + integration)
- Integrated into Identity router with JWT middleware

**Migration Details:**
- `identity_permissions` table with GENERATED ALWAYS column: `name AS (resource || ':' || action)`
- `identity_roles` table with soft delete support
- Junction tables: `identity_user_roles`, `identity_role_permissions`
- Seed data: 29+ permissions (wildcard, users, roles, permissions, contacts, profiles, customers)
- Seed data: 5 system roles (superadmin, admin, manager, user, guest)

**API Endpoints:**
- Role Management: Create, List, GetByID, GetByName, Update, Delete, GetUserRoles
- Permission Management: Create, List, GetByID, GetByName, Update, Delete, GetRolePermissions

### 2. JWT Secret Validation (COMPLETED December 29, 2025)

**Implementation:**
- Added `Validate()` method in `internal/infrastructure/config/yaml_config.go`
- Added `validateJWTSecret()` with environment-aware checks
- Production validation: minimum 32 characters + blocks default dev secret
- Development: allows any secret (including default for convenience)
- Called from `cmd/api/main.go` after config load (before any initialization)
- Comprehensive tests: 12+ test cases covering all scenarios

**Validation Rules:**
- Empty secret: blocked in all environments
- Production: minimum 32 chars required
- Production: default "dev-secret-key-change-in-production" blocked
- Development: no restrictions (allows short secrets for local testing)
- Clear error messages with guidance

**Test Coverage:**
```
TestValidate_ValidConfig             PASS
TestValidate_JWTSecretValidation     PASS (7 subtests)
TestValidate_DatabaseValidation      PASS (3 subtests)
TestValidate_ServerValidation        PASS
```

### 3. Rate Limiting for Authentication (COMPLETED December 29, 2025)

**Implementation:**
- Created `pkg/middleware/ratelimit.go` with RateLimiter using token bucket algorithm
- IP-based tracking with per-IP rate limiters (thread-safe with RWMutex)
- Configurable rate and burst size using `golang.org/x/time/rate`
- Standard X-RateLimit-* headers (Limit, Remaining, Reset)
- Memory cleanup method (`CleanupVisitors()`) to prevent leaks
- Comprehensive test suite: 10 tests covering all scenarios (100% coverage)

**Integration:**
- Applied to Identity context authentication endpoints:
  * Login: 5 attempts per minute (burst 1)
  * Register: 3 attempts per minute (burst 1)
- Returns 429 Too Many Requests with clear error message
- Separate limits per IP address (one abusive client doesn't affect others)

**Test Coverage:**
```
TestNewRateLimiter                          PASS
TestRateLimiter_AllowsRequests              PASS
TestRateLimiter_BlocksExcessRequests        PASS
TestRateLimiter_SetsRateLimitHeaders        PASS
TestRateLimiter_SetsResetHeaderWhenLimitExceeded PASS
TestRateLimiter_SeparatesIPAddresses        PASS
TestRateLimiter_AllowsBurstRequests         PASS
TestRateLimiter_CleanupVisitors             PASS
TestRateLimiter_ConcurrentAccess            PASS
TestRateLimiter_GetVisitorCount             PASS
```

**Documentation:**
- Created `docs/RATE_LIMITING.md` (680+ lines) - Complete implementation guide
- Updated `README.md` with Rate Limiting overview section

---

## CRITICAL (High Priority)

### 1. Token Revocation Mechanism

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

## HIGH PRIORITY (Week 1-2)

### 2. Health Checks for Dependencies

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

### 3. Database Indexes Audit

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
SELECT * FROM users WHERE email = ?  -- UNIQUE INDEX exists
SELECT * FROM contacts WHERE user_id = ?  -- Need INDEX
SELECT * FROM profiles WHERE user_id = ?  -- Need UNIQUE INDEX
SELECT * FROM customers WHERE email = ?  -- Need INDEX
SELECT * FROM customers WHERE status = ?  -- Need INDEX
```

---

### 4. Error Definitions Consistency

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

## MEDIUM PRIORITY (Week 2-3)

### 5. Panic Handling Improvements

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

### 6. Soft Delete Query Audit

**Current State:**
- Most queries include `deleted_at IS NULL`
- Need to verify ALL queries

**Tasks:**
- [ ] Audit all SELECT queries
- [ ] Add deleted_at check where missing
- [ ] Consider DB view for active records
- [ ] Add integration tests

**Estimate:** 2 hours

---

### 7. Caching Layer (Redis)

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

### 8. N+1 Query Optimization

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

### 9. CSRF Protection

**Current State:**
- Using Bearer tokens (safe)
- If cookies added, need CSRF

**Tasks:**
- [ ] Add CSRF middleware (if needed)
- [ ] Token generation/validation
- [ ] Add to forms
- [ ] Tests

**Status:** Not needed now (Bearer auth), but document for future

---

## Summary Statistics

| Priority | Tasks | Estimated Time | Status |
|----------|-------|----------------|--------|
| Critical | 1 | 3 hours | Ready to start |
| High | 3 | 4 hours | Week 1-2 |
| Medium | 5 | 10 hours | Week 2-3 |
| **TOTAL** | **9** | **17 hours** | ~2 working days |

---

## Notes

- All changes should include tests (unit + integration)
- Update README.md if public API changes
- Keep backwards compatibility where possible
- Document breaking changes in CHANGELOG.md

---

## Related Documents

- [Clean Architecture Summary](CLEAN_ARCHITECTURE_SUMMARY.md)
- [Testing Patterns](TESTING_PATTERNS.md)
- [AI Instructions](../.github/copilot-instructions.md)

---

**Last Updated:** December 29, 2025  
**Next Review:** After next major feature
