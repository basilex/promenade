# Gaps & Technical Debt - Action Plan

**Created:** December 28, 2025  
**Last Updated:** December 29, 2025  
**Status:** Integration Tests COMPLETED 🎉  
**Priority Order:** Critical → High → Medium

---

## COMPLETED TASKS

### 1. Integration Test Fixes (COMPLETED December 29, 2025) ✅

**Status:** **ALL 24 INTEGRATION TESTS PASSING (100%)**

**Final Fixes:**
1. **Customer Relations test** - Duplicate email constraint violations
   - Problem: Loop creating customers with same email pattern (`b2b_%d@test.com`, `assigned1_%d@test.com`)
   - Solution: Add UUID suffix to each email in loop (`b2b_%s_%d@test.com` with `uuidv7.New().String()[:8]`)
   - Result: 4/4 customer tests passing ✅

2. **Contact WithTransaction test** - Transaction rollback verification
   - Problem: `t.FailNow()` preventing rollback verification code from running
   - Solution: Rewrite with manual transaction + `tx.Rollback()` + separate verification transaction
   - Added imports: `database` package for `SetTxToContext`
   - Result: 3/3 contact tests passing ✅

3. **User Queries test** - ListUsers returning 0 results
   - Problem: Wrong parameter order - `repo.ListUsers(ctx, 10, 0)` means page=10, pageSize=0
   - Solution: Fix to `repo.ListUsers(ctx, 1, 10)` - page=1, pageSize=10
   - Root cause: `LIMIT 0` in SQL query (caused by pageSize=0)
   - Result: 3/3 user tests passing ✅

**Test Results:**
| Context              | Tests | Status |
|---------------------|-------|--------|
| Customer Management | 4     | ✅ 100% |
| Identity Contact    | 3     | ✅ 100% |
| Identity User       | 3     | ✅ 100% |
| Identity Permission | 2     | ✅ 100% |
| Identity Profile    | 2     | ✅ 100% |
| Identity Role       | 2     | ✅ 100% |
| Shared (all)        | 8     | ✅ 100% |
| **Total**           | **24**| **✅ 100%** |

**Duration:** ~2s (all cached after first run)

**Commits:**
1. `9df4b40` - Fix duplicate emails in customer tests
2. `2112529` - Fix transaction isolation (SetTxToContext implementation)
3. `a585aba` - Add user creation for FK constraints + unique identifiers
4. `54dfb5c` - Fix assertions and repository pattern
5. `e4f619d` - Remove valueobject.MustNewEmail usages
6. `afea541` - Fix type assertion issues (int vs int64)
7. `2bb77be` - Fix remaining 3 integration test failures - ALL TESTS PASSING (24/24)

### 2. RBAC Implementation (COMPLETED December 29, 2025)

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

### 3. JWT Secret Validation (COMPLETED December 29, 2025)

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

### 4. Token Revocation Mechanism (COMPLETED December 29, 2025)

**Implementation:**
- Created `pkg/jwt/revocation.go` with Redis-backed token blacklist
- TokenRevoker stores revoked tokens with TTL matching expiration time
- Updated JWT `AuthMiddleware` to check revocation status before validation
- Added `POST /auth/revoke` endpoint for user logout functionality
- Integrated TokenRevoker into Identity router (all protected routes check revocation)

**Components:**
- `pkg/jwt/revocation.go`: Redis blacklist with TTL management
- `pkg/jwt/revocation_test.go`: 9 comprehensive tests (100% passing)
  * Revoke/IsRevoked operations
  * TTL expiration handling
  * Multiple tokens & concurrent revocations
  * Stats monitoring

**Middleware Updates:**
- AuthMiddleware now accepts optional TokenRevoker parameter
- Returns 401 TOKEN_REVOKED error for revoked tokens
- All existing middleware tests updated (pass nil for tests without revocation)

**Configuration:**
- Added `redis` section to AppConfig (addr, password, db)
- Redis connection in main.go with graceful fallback (revocation disabled if Redis unavailable)
- Config updated in `app.dev.yaml`

**Test Results:**
```
TestNewTokenRevoker                         PASS
TestTokenRevoker_Revoke                     PASS
TestTokenRevoker_RevokeExpiredToken         PASS
TestTokenRevoker_IsRevoked_NotRevoked       PASS
TestTokenRevoker_TTL                        PASS
TestTokenRevoker_MultipleTokens             PASS
TestTokenRevoker_Stats                      PASS
TestTokenRevoker_RevokeAllForUser_NotImplemented PASS
TestTokenRevoker_ConcurrentRevocations      PASS
```

**Security:**
- Tokens revoked on logout via `/auth/revoke` endpoint
- Expired tokens not stored (TTL validation prevents unnecessary storage)
- Graceful Redis failure (if Redis unavailable, revocation disabled but app continues)
- Concurrent-safe revocation operations (Redis atomic operations)

### 5. Health Checks for Dependencies (COMPLETED December 29, 2025)

**Implementation:**
- Created `internal/infrastructure/health/health.go` with Checker pattern
- Created `internal/infrastructure/health/handler.go` with 4 HTTP endpoints
- Integrated into `cmd/api/main.go` with dependency injection
- Comprehensive test suite: 21 tests (100% passing)

**Components:**
- **health.Checker**: Core health check logic with timeout-based checks
  * CheckAll(ctx) - All dependencies with 5-second timeout
  * CheckDatabase(ctx) - PostgreSQL ping + query test
  * CheckRedis(ctx) - Redis ping (graceful if not configured)
  * CheckEventBus(ctx) - Event bus health check
- **health.Handler**: HTTP endpoints with proper status codes
  * GET /health - Overall status (200/503)
  * GET /health/db - Database only
  * GET /health/redis - Redis only
  * GET /health/bus - Event Bus only

**Features:**
- **3 Status Levels**: healthy, degraded, unhealthy
- **Timeout Protection**: 5-second timeout for all checks
- **Graceful Degradation**: Optional dependencies (Redis) handled gracefully
- **HTTP Status Codes**: 200 (healthy/degraded), 503 (unhealthy)
- **Version Tracking**: Reports application version in response

**Test Coverage:**
```
TestChecker_CheckDatabase_Healthy                PASS
TestChecker_CheckDatabase_Unhealthy_PingFailed   PASS
TestChecker_CheckDatabase_Degraded_QueryFailed   PASS
TestChecker_CheckRedis_Healthy                   PASS
TestChecker_CheckRedis_Unhealthy                 PASS
TestChecker_CheckRedis_NotConfigured             PASS
TestChecker_CheckEventBus_Healthy                PASS
TestChecker_CheckAll_AllHealthy                  PASS
TestChecker_CheckAll_DatabaseUnhealthy           PASS
TestChecker_CheckAll_WithTimeout                 PASS (5s timeout verified)
TestHandler_CheckAll_Healthy                     PASS
TestHandler_CheckAll_Unhealthy                   PASS
TestHandler_CheckAll_Degraded                    PASS
TestHandler_CheckDatabase_Healthy                PASS
TestHandler_CheckDatabase_Unhealthy              PASS
TestHandler_CheckRedis                           PASS
TestHandler_CheckEventBus                        PASS
TestHandler_RegisterRoutes                       PASS
```

**Dependencies Added:**
- `github.com/DATA-DOG/go-sqlmock` - Database mock for testing
- `github.com/go-redis/redismock/v9` - Redis mock for testing

**Integration:**
- Replaces old simple `/health` endpoint with comprehensive system
- Dependencies: db (PostgreSQL), redisClient (optional), eventBus
- Graceful handling when Redis not available

---

## COMPLETED TASKS (HIGH PRIORITY)

### 1. Database Indexes Audit (COMPLETED December 30, 2025) ✅

**Status:** **ALL QUERIES PROPERLY INDEXED - NO ACTION REQUIRED**

**Audit Results:**
- ✅ Analyzed 13 tables across 4 contexts
- ✅ Reviewed 70+ SELECT/UPDATE/DELETE queries
- ✅ Found 60+ existing indexes
- ✅ All foreign keys have supporting indexes (100% coverage)
- ✅ Strategic composite indexes for multi-column queries
- ✅ Extensive partial indexes (WHERE deleted_at IS NULL)
- ✅ Functional indexes for case-insensitive searches (LOWER(email))
- ✅ GIN indexes for JSONB array operations

**Key Findings:**
- **identity_users**: 4 indexes (email, status, created_at, deleted_at) ✅
- **identity_contacts**: 4 indexes including composite (user_id, contact_type) ✅
- **identity_profiles**: 3 partial indexes with soft delete ✅
- **customer_mgmt_customers**: 9 indexes including GIN for tags ✅
- **identity_permissions/roles**: Complete RBAC indexing ✅
- **identity_user_sessions**: Composite index for (user_id, created_at DESC) ✅
- **identity_login_attempts**: Triple composite for security queries ✅

**Documentation:** [INDEX_AUDIT_REPORT.md](INDEX_AUDIT_REPORT.md) (comprehensive 500+ line report)

**Conclusion:** No missing indexes. Database is professionally optimized.

---

### 2. Error Definitions Consistency (COMPLETED December 30, 2025) ✅

**Status:** **ALL AGGREGATES STANDARDIZED**

**Implementation:**
- Created 9 new `errors.go` files (Identity: 5, Shared: 4)
- Removed old error definitions from entity/usecase files
- Standardized error naming (e.g., `ErrCountryNotFound`, `ErrCurrencyNotFound`)
- Fixed 3 compilation errors (missing imports)
- Updated all unit and integration tests

**Files Created:**
1. `internal/contexts/identity/user/errors.go` - 7 errors
2. `internal/contexts/identity/contact/errors.go` - 6 errors
3. `internal/contexts/identity/profile/errors.go` - 5 errors
4. `internal/contexts/identity/role/errors.go` - 6 errors
5. `internal/contexts/identity/permission/errors.go` - 6 errors
6. `internal/contexts/shared/country/errors.go` - 3 errors
7. `internal/contexts/shared/currency/errors.go` - 3 errors
8. `internal/contexts/shared/language/errors.go` - 2 errors
9. `internal/contexts/shared/timezone/errors.go` - 3 errors

**Files Modified:**
- Removed error definitions from 8 entity/usecase files
- Updated 7 repository files to use new error names
- Fixed 4 unit test files
- Fixed 1 integration test file

**Error Naming Convention:**
- `Err{Aggregate}{Condition}` (e.g., `ErrUserNotFound`, `ErrContactAlreadyExists`)
- Full words, no abbreviations (e.g., `ErrCountryNotFound` not `ErrNotFound`)
- Consistent across all contexts

**Benefits:**
- Single source of truth for domain errors
- Consistent naming conventions
- Better code organization
- Easier to maintain and update

**Test Results:**
- All unit tests passing ✅
- All integration tests passing ✅
- No compilation errors ✅

**Time:** 45 minutes (25% faster than 1h estimate)

---

## HIGH PRIORITY (Week 1-2)

---

### 3. Error Definitions Consistency (MOVED TO COMPLETED)

See section above ✅

---

## COMPLETED TASKS (HIGH PRIORITY)

### 4. Panic Handling Improvements (COMPLETED December 30, 2025) ✅

**Status:** **ALL PANIC FUNCTIONS REMOVED**

**Implementation:**
- Removed `MustGetClaims()` and `MustGetUserID()` functions from pkg/jwt/middleware.go
- Functions were NOT used in production code (only in tests and documentation)
- Violate Go best practice: "don't panic in library code"
- Safe alternatives exist: GetClaims() returns nil, GetUserID() returns ""

**Changes:**
- Removed 2 panic functions (24 lines of code)
- Removed 2 test functions (125 lines)
- Removed unused uuidv7 import
- Updated documentation (JWT README, RBAC guides, pkg README)

**Benefits:**
- No panics in library code
- Follows Go idioms
- Explicit error handling (handlers check for nil/empty)
- Reduced maintenance burden

**Time:** 10 minutes (estimated 1h → 6x faster)

---

## MEDIUM PRIORITY (Week 2-3)

---

### 5. Soft Delete Query Audit

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

| Priority  | Tasks | Estimated Time | Status        |
|-----------|-------|----------------|---------------|
| Completed | 7     | 14 hours       | Done ✅       |
| High      | 0     | 0 hours        | All done!     |
| Medium    | 4     | 9 hours        | Week 2-3      |
| **TOTAL** | **11**| **23 hours**   | ~2 working days |

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

**Last Updated:** December 30, 2025  
**Next Review:** After next major feature
