# API Versioning Strategy

**Status**: Official Policy  
**Effective Date**: January 5, 2026  
**Current API Version**: v1  
**Last Updated**: January 5, 2026

---

## Overview

Promenade Platform uses **URL-based versioning** to ensure API stability and allow non-breaking evolution. This document defines our versioning strategy, deprecation policy, and migration guidelines.

---

## Versioning Approach

### URL-Based Versioning

All API endpoints are versioned via URL path:

```
https://api.promenade.example.com/api/v1/customers
https://api.promenade.example.com/api/v2/customers
```

**Rationale**:
- **Clear**: Version explicit in URL
- **Cacheable**: Different versions can be cached separately
- **Simple**: No custom headers needed
- **Documentation-friendly**: Swagger groups by version

**Alternative approaches NOT used**:
- Header versioning (`Accept: application/vnd.promenade.v1+json`) - harder for developers
- Query parameter (`/api/customers?version=1`) - breaks caching
- Content negotiation - complex for REST API

---

## Version Lifecycle

### 1. Active (Current Version)

**Status**: Fully supported, receives all new features  
**Current**: v1 (launched January 2026)

**Guarantees**:
- Bug fixes applied immediately
- Security patches prioritized
- New features added
- Performance optimizations
- Full documentation maintained

### 2. Deprecated (Previous Version)

**Status**: Supported but discouraged, no new features  
**Duration**: 12 months after successor release  
**Example**: v1 deprecated when v2 released (12 months support)

**Guarantees**:
- Critical bug fixes only
- Security patches applied
- No new features
- Documentation maintained
- Migration guide available

**Indicators**:
- HTTP header: `Deprecation: true`
- HTTP header: `Sunset: 2027-06-01T00:00:00Z` (RFC 8594)
- HTTP header: `Link: </api/v2/customers>; rel="successor-version"`
- Swagger annotation: `@deprecated Use /api/v2/customers instead`

### 3. Sunset (End of Life)

**Status**: No longer supported, returns 410 Gone  
**Timing**: 12 months after deprecation announcement

**Response**:
```http
HTTP/1.1 410 Gone
Content-Type: application/json
Link: </api/v2/customers>; rel="successor-version"

{
  "status": "error",
  "error": {
    "code": "API_VERSION_SUNSET",
    "message": "API v1 was sunset on 2027-06-01. Please migrate to v2.",
    "migration_guide": "https://docs.promenade.example.com/migration/v1-to-v2"
  }
}
```

---

## Breaking Changes Policy

### What Constitutes a Breaking Change

**Major version bump required (v1 → v2)** for:

1. **Removing fields** from response
   ```json
   // v1
   {"id": "123", "name": "John", "email": "john@example.com"}
   
   // v2 (BREAKING - removed email)
   {"id": "123", "name": "John"}
   ```

2. **Renaming fields**
   ```json
   // v1
   {"customer_id": "123"}
   
   // v2 (BREAKING - renamed field)
   {"customerId": "123"}
   ```

3. **Changing field types**
   ```json
   // v1
   {"amount": 100}
   
   // v2 (BREAKING - string to number)
   {"amount": "100.00"}
   ```

4. **Removing endpoints**
   ```
   DELETE /api/v1/users/:id  // Removed in v2
   ```

5. **Changing HTTP methods**
   ```
   PATCH /api/v1/users/:id   // Changed to PUT in v2
   ```

6. **Changing authentication scheme**
   ```
   Bearer token → OAuth 2.0
   ```

7. **Changing error response format**
   ```json
   // v1
   {"error": "Not found"}
   
   // v2 (BREAKING - new structure)
   {"status": "error", "error": {"code": "NOT_FOUND", "message": "Not found"}}
   ```

### Non-Breaking Changes

**Minor version updates (documentation only)** for:

1. **Adding new fields** (with defaults)
   ```json
   // v1
   {"id": "123", "name": "John"}
   
   // v1.1 (NON-BREAKING - added optional field)
   {"id": "123", "name": "John", "phone": "+1234567890"}
   ```

2. **Adding new endpoints**
   ```
   POST /api/v1/customers/:id/notes  // New in v1.1
   ```

3. **Adding new optional query parameters**
   ```
   GET /api/v1/customers?filter=active  // New in v1.1
   ```

4. **Expanding enum values** (if clients ignore unknown values)
   ```json
   // v1: status = "active" | "suspended"
   // v1.1: status = "active" | "suspended" | "pending"  // Added pending
   ```

5. **Improving performance** (no API contract changes)

6. **Fixing bugs** (that align with documented behavior)

---

## Deprecation Process

### Timeline

```
Month 0: New version released (v2)
  ↓
  ├─ Deprecation announcement
  ├─ Migration guide published
  └─ Deprecation headers added to v1

Month 6: Deprecation reminder
  ↓
  └─ Email to all API consumers

Month 9: Final warning
  ↓
  └─ Email + dashboard notification

Month 12: Sunset (v1 disabled)
  ↓
  └─ v1 returns 410 Gone
```

### Deprecation Headers

When an endpoint or version is deprecated:

```http
HTTP/1.1 200 OK
Deprecation: true
Sunset: Sun, 01 Jun 2027 00:00:00 GMT
Link: </api/v2/customers>; rel="successor-version"
X-API-Warn: This endpoint is deprecated and will be removed on 2027-06-01

{
  "data": { ... }
}
```

**Implementation**:
```go
// middleware/deprecation.go
func DeprecateEndpoint(sunsetDate time.Time, successorURL string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Deprecation", "true")
        c.Header("Sunset", sunsetDate.Format(http.TimeFormat))
        c.Header("Link", fmt.Sprintf("<%s>; rel=\"successor-version\"", successorURL))
        c.Header("X-API-Warn", fmt.Sprintf("Deprecated. Sunset: %s", sunsetDate.Format("2006-01-02")))
        c.Next()
    }
}

// Usage in router
deprecated := v1.Group("/customers")
deprecated.Use(middleware.DeprecateEndpoint(
    time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
    "/api/v2/customers",
))
```

### Swagger Annotations

Mark deprecated endpoints in Swagger:

```go
// @Summary Get customer by ID (DEPRECATED)
// @Description Retrieves customer details. DEPRECATED: Use /api/v2/customers/:id instead
// @Tags customers
// @Deprecated
// @Param id path string true "Customer ID"
// @Success 200 {object} CustomerDTO
// @Header 200 {string} Deprecation "true"
// @Header 200 {string} Sunset "Sun, 01 Jun 2027 00:00:00 GMT"
// @Router /api/v1/customers/{id} [get]
func (h *CustomerHandler) GetByID(c *gin.Context) {
    // ...
}
```

---

## Migration Guides

### Creating Migration Guide

When releasing v2, create comprehensive migration guide:

**Template**: `docs/migration/v1-to-v2.md`

**Structure**:
1. **Overview** - Summary of changes
2. **Breaking Changes** - Detailed list with examples
3. **Step-by-Step Migration** - Code examples (before/after)
4. **Testing Checklist** - Validation steps
5. **Rollback Plan** - How to revert if needed
6. **Support** - Contact info for help

**Example** (see section below for full template)

### Parallel Running

**Strategy**: Support both versions simultaneously during transition

```
January 2027: v2 released, v1 active
  ↓
  ├─ /api/v1/* → Active (12 months remaining)
  └─ /api/v2/* → Active (current)

January 2028: v1 sunset
  ↓
  ├─ /api/v1/* → 410 Gone
  └─ /api/v2/* → Active
```

**Benefits**:
- Zero downtime migration
- Clients migrate at their own pace
- A/B testing possible
- Easy rollback if issues found

---

## Version Detection

### Current Version

Clients can discover current API version:

```bash
curl https://api.promenade.example.com/api/v1
```

**Response**:
```json
{
  "message": "Promenade CRM Platform API v1",
  "version": "v1",
  "build": "0.1.0",
  "released": "2026-01-01",
  "deprecated": false,
  "successor": null
}
```

### Deprecated Version

```bash
curl https://api.promenade.example.com/api/v1
```

**Response** (after v2 release):
```json
{
  "message": "Promenade CRM Platform API v1",
  "version": "v1",
  "build": "0.1.0",
  "released": "2026-01-01",
  "deprecated": true,
  "sunset_date": "2028-01-01",
  "successor": "/api/v2",
  "migration_guide": "https://docs.promenade.example.com/migration/v1-to-v2"
}
```

---

## Backward Compatibility

### When to Maintain Backward Compatibility

**Always maintain** for:
1. Security patches in deprecated versions
2. Critical bug fixes
3. Data integrity issues

**Never maintain** for:
1. New features in deprecated versions
2. Performance optimizations (unless critical)
3. Non-critical bug fixes

### Adapter Pattern

For significant changes, use adapter pattern internally:

```go
// v2/customers/adapter.go
type CustomerAdapterV1 struct {
    v2Handler *v2.CustomerHandler
}

func (a *CustomerAdapterV1) GetByID(c *gin.Context) {
    // Call v2 handler
    customer, err := a.v2Handler.GetCustomer(c)
    if err != nil {
        // Handle error
        return
    }
    
    // Convert v2 response to v1 format
    v1Response := convertToV1Format(customer)
    c.JSON(200, v1Response)
}

func convertToV1Format(v2Customer *v2.Customer) *v1.Customer {
    return &v1.Customer{
        CustomerID: v2Customer.ID,  // Renamed field
        FullName:   v2Customer.Name, // Combined field
        // Map all fields
    }
}
```

---

## Versioning in Code

### Directory Structure

```
internal/
  contexts/
    customer-mgmt/
      customer/
        v1/
          entity.go
          usecase.go
          adapter/
            http/
              handler/
                customer_handler.go
              dto/
                customer_dto.go
        v2/
          entity.go  (new fields/structure)
          usecase.go
          adapter/
            http/
              handler/
                customer_handler.go
              dto/
                customer_dto.go
```

**Alternative** (shared core, versioned adapters):
```
internal/
  contexts/
    customer-mgmt/
      customer/
        entity.go      (shared core domain)
        usecase.go     (shared business logic)
        adapter/
          http/
            v1/
              handler/
                customer_handler.go
              dto/
                customer_dto.go
            v2/
              handler/
                customer_handler.go
              dto/
                customer_dto.go
```

### Router Organization

```go
// cmd/api/server.go
func (s *Server) SetupRoutes() {
    api := s.router.Group("/api")
    
    // v1 routes
    v1 := api.Group("/v1")
    {
        customermgmt.RegisterRoutesV1(v1, s.app.DB)
    }
    
    // v2 routes (when ready)
    v2 := api.Group("/v2")
    {
        customermgmt.RegisterRoutesV2(v2, s.app.DB)
    }
}
```

---

## Testing Strategy

### Test Both Versions

When v2 is released, maintain tests for both:

```go
// test/integration/api/v1/customer_test.go
func TestCustomerAPI_V1(t *testing.T) {
    // Test v1 endpoints
}

// test/integration/api/v2/customer_test.go
func TestCustomerAPI_V2(t *testing.T) {
    // Test v2 endpoints
}
```

### Regression Testing

Run full test suite for deprecated version:

```bash
make test-api-v1  # All v1 tests
make test-api-v2  # All v2 tests
```

### Contract Testing

Use contract tests to ensure v2 doesn't break v1:

```go
func TestBackwardCompatibility_V1_to_V2(t *testing.T) {
    // Ensure v1 clients can still work with v2 responses
    // (if using adapter pattern)
}
```

---

## Monitoring & Analytics

### Track Version Usage

Log API version with each request:

```go
// middleware/version_logger.go
func VersionLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        version := extractVersion(c.Request.URL.Path) // Extract from /api/v1/...
        
        // Log metrics
        metrics.APIVersionUsage.WithLabelValues(version).Inc()
        
        c.Next()
    }
}
```

### Deprecation Warnings Dashboard

Track deprecated endpoint usage:

```sql
SELECT 
    endpoint,
    COUNT(*) as calls,
    COUNT(DISTINCT client_id) as unique_clients
FROM api_logs
WHERE version = 'v1' 
  AND deprecated = true
  AND timestamp > NOW() - INTERVAL '30 days'
GROUP BY endpoint
ORDER BY calls DESC;
```

**Alerts**:
- Email clients still using deprecated endpoints (6 months before sunset)
- Dashboard showing migration progress
- Slack notifications for high-volume deprecated endpoint usage

---

## Communication Plan

### Announcement Channels

When releasing new version or deprecating old:

1. **Documentation**
   - Migration guide published
   - Changelog updated
   - API docs updated with deprecation notices

2. **Email**
   - All registered API consumers notified
   - 3 reminders: 0, 6, 9 months before sunset

3. **Dashboard**
   - In-app notification for users
   - Banner on developer portal

4. **Social Media**
   - Blog post announcement
   - Twitter/LinkedIn updates

5. **Developer Community**
   - GitHub discussions
   - Discord/Slack channels

### Email Template

**Subject**: [Action Required] Promenade API v1 Deprecation Notice

```
Hello Promenade Developer,

We're excited to announce the release of Promenade API v2 with improved 
performance and new features!

IMPORTANT: API v1 will be sunset on June 1, 2027 (12 months from now).

What you need to do:
1. Review the migration guide: https://docs.promenade.example.com/migration/v1-to-v2
2. Test your integration with v2 in our staging environment
3. Deploy v2 integration before June 1, 2027

Timeline:
- Now: v2 released, v1 still active
- Sep 1, 2026: Reminder email
- Dec 1, 2026: Final warning
- Jun 1, 2027: v1 sunset (410 Gone responses)

Questions? Reply to this email or visit our developer forum.

Best regards,
Promenade API Team
```

---

## Migration Guide Template

**File**: `docs/migration/v1-to-v2.md`

```markdown
# Migration Guide: API v1 → v2

**Released**: January 1, 2027  
**v1 Sunset**: January 1, 2028  
**Migration Deadline**: December 31, 2027

---

## Overview

API v2 introduces performance improvements and modernized response formats.
This guide helps you migrate from v1 to v2 with minimal disruption.

**Estimated Migration Time**: 2-4 hours for typical integration

---

## Breaking Changes

### 1. Field Renames

**Customer Response**:
```json
// v1
{
  "customer_id": "123",
  "full_name": "John Doe"
}

// v2
{
  "id": "123",
  "name": "John Doe"
}
```

**Action**: Update field names in your code.

### 2. Date Format

**v1**: Unix timestamp (integer)
```json
{"created_at": 1704067200}
```

**v2**: ISO 8601 (string)
```json
{"created_at": "2024-01-01T00:00:00Z"}
```

**Action**: Update date parsing logic.

### 3. Removed Endpoints

**Removed**:
- `DELETE /api/v1/customers/:id/hard-delete`

**Replacement**:
- Use soft delete: `DELETE /api/v2/customers/:id` (reversible)
- Use purge API: `POST /api/v2/admin/purge/customers/:id` (admin only)

---

## Step-by-Step Migration

### Step 1: Update Base URL

```javascript
// Before (v1)
const API_BASE = 'https://api.promenade.example.com/api/v1';

// After (v2)
const API_BASE = 'https://api.promenade.example.com/api/v2';
```

### Step 2: Update Field Names

```javascript
// Before (v1)
const customerId = response.data.customer_id;
const name = response.data.full_name;

// After (v2)
const customerId = response.data.id;
const name = response.data.name;
```

### Step 3: Update Date Parsing

```javascript
// Before (v1)
const date = new Date(response.data.created_at * 1000);

// After (v2)
const date = new Date(response.data.created_at);
```

---

## Testing Checklist

- [ ] All API calls use `/api/v2` base URL
- [ ] Field names updated (customer_id → id, etc.)
- [ ] Date parsing updated (Unix → ISO 8601)
- [ ] Error handling updated (new error codes)
- [ ] Removed endpoint calls replaced
- [ ] Integration tests passing with v2
- [ ] Staging environment tested
- [ ] Rollback plan documented

---

## Rollback Plan

If issues found after migration:

1. **Revert base URL** to `/api/v1`
2. **Re-deploy** previous version
3. **Report issues** to support@promenade.example.com
4. **Schedule re-migration** after fix

v1 will remain available until January 1, 2028.

---

## Support

**Questions?**
- Email: api-support@promenade.example.com
- Discord: https://discord.gg/promenade-dev
- GitHub: https://github.com/basilex/promenade/discussions

**Stuck?**
We offer free migration consulting calls. Book here: https://calendly.com/promenade-api
```

---

## Related Documentation

- [API Reference](../reference/api-reference.md)
- [Swagger Documentation](http://localhost:8081/api/docs/index.html)
- [Changelog](../reference/changelog.md)
- [Developer Portal](../guides/getting-started.md)

---

**Version**: 1.0.0  
**Last Updated**: January 5, 2026  
**Status**: Official Policy  
**Maintainer**: Promenade API Team
