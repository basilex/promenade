# API Versioning - Practical Examples

This document provides **practical code examples** for implementing API versioning in Promenade Platform.

See [API Versioning Strategy](api-versioning.md) for complete policy and guidelines.

---

## Table of Contents

1. [Basic Setup](#basic-setup)
2. [Deprecating an Endpoint](#deprecating-an-endpoint)
3. [Sunsetting a Version](#sunsetting-a-version)
4. [Version Detection API](#version-detection-api)
5. [Swagger Annotations](#swagger-annotations)
6. [Testing Deprecated Endpoints](#testing-deprecated-endpoints)

---

## Basic Setup

### Current Structure (v1 only)

```go
// cmd/api/server.go
func (s *Server) SetupRoutes() {
    api := s.router.Group("/api")
    {
        v1 := api.Group("/v1")
        {
            // All contexts registered under v1
            sharedRouter.RegisterRoutes(v1)
            identityRouter.RegisterRoutes(v1)
            customerMgmtRouter.RegisterRoutes(v1)
            orderMgmtRouter.RegisterRoutes(v1)
            billingRouter.RegisterRoutes(v1)
        }
    }
}
```

### Adding v2 (Future)

```go
// cmd/api/server.go
func (s *Server) SetupRoutes() {
    api := s.router.Group("/api")
    {
        // v1 - Deprecated but still active
        v1 := api.Group("/v1")
        v1.Use(middleware.DeprecateEndpoint(
            time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
            "/api/v2",
        ))
        {
            // Legacy routes
            sharedRouter.RegisterRoutes(v1)
            identityRouter.RegisterRoutesV1(v1)  // Old version
            customerMgmtRouter.RegisterRoutesV1(v1)
            orderMgmtRouter.RegisterRoutesV1(v1)
            billingRouter.RegisterRoutesV1(v1)
        }

        // v2 - Current active version
        v2 := api.Group("/v2")
        {
            // New routes with breaking changes
            sharedRouter.RegisterRoutes(v2)
            identityRouter.RegisterRoutes(v2)  // New version
            customerMgmtRouter.RegisterRoutes(v2)
            orderMgmtRouter.RegisterRoutes(v2)
            billingRouter.RegisterRoutes(v2)
        }
    }
}
```

---

## Deprecating an Endpoint

### Scenario: Deprecating GET /api/v1/customers

**Timeline**: Announced 2026-06-01, Sunset 2027-06-01 (12 months)

### Step 1: Add Deprecation Middleware

```go
// internal/contexts/customer-mgmt/router.go
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
    customers := api.Group("/customers")
    {
        // NEW: v2 endpoint with fixed field name
        customers.GET("", r.customerHandler.List)
        customers.GET("/:id", r.customerHandler.GetByID)
        
        // ... other routes
    }
}

func (r *Router) RegisterRoutesV1(api *gin.RouterGroup) {
    customers := api.Group("/customers")
    
    // Apply deprecation to entire v1 customers group
    customers.Use(middleware.DeprecateEndpoint(
        time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
        "/api/v2/customers",
    ))
    {
        // OLD: v1 endpoints (deprecated)
        customers.GET("", r.customerHandler.ListV1)  // Uses old response format
        customers.GET("/:id", r.customerHandler.GetByIDV1)
        
        // ... other routes
    }
}
```

### Step 2: Client Receives Deprecation Headers

```bash
curl -i https://api.promenade.example.com/api/v1/customers
```

**Response**:
```http
HTTP/1.1 200 OK
Content-Type: application/json
Deprecation: true
Sunset: Sun, 01 Jun 2027 00:00:00 GMT
Link: </api/v2/customers>; rel="successor-version"
X-API-Warn: This API version is deprecated and will be removed on 2027-06-01. Please migrate to /api/v2/customers

{
  "status": "success",
  "data": [...]
}
```

### Step 3: Swagger Documentation

```go
// internal/contexts/customer-mgmt/adapter/http/handler/customer_handler.go

// ListV1 lists all customers (DEPRECATED - use /api/v2/customers)
//
// @Summary      List customers (v1 - DEPRECATED)
// @Description   DEPRECATED: This endpoint will be removed on 2027-06-01. Use /api/v2/customers instead.
// @Description  **Migration Guide**: https://docs.promenade.example.com/migration/v1-to-v2#customers
// @Description  **Breaking Changes**: Response field `fullName` renamed to `full_name` (snake_case)
// @Tags         Customers (Deprecated)
// @Accept       json
// @Produce      json
// @Param        page     query    int    false  "Page number (default: 1)"
// @Param        page_size query   int    false  "Page size (default: 20)"
// @Success      200      {object} response.PaginatedResponse{data=[]dto.CustomerResponseV1}
// @Failure      401      {object} response.ErrorResponse
// @Failure      500      {object} response.ErrorResponse
// @Deprecated
// @Router       /api/v1/customers [get]
// @Security     BearerAuth
func (h *CustomerHandler) ListV1(c *gin.Context) {
    // Implementation with old response format
    // ...
}
```

---

## Sunsetting a Version

### Scenario: Completely removing v1 after 12 months

### Step 1: Add Sunset Middleware

```go
// cmd/api/server.go
func (s *Server) SetupRoutes() {
    api := s.router.Group("/api")
    {
        // v1 - SUNSET (returns 410 Gone)
        v1 := api.Group("/v1")
        v1.Use(middleware.SunsetVersion(
            time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
            "/api/v2",
            "https://docs.promenade.example.com/migration/v1-to-v2",
        ))
        {
            // Routes won't be reached - middleware returns 410
            // Keep for reference
        }

        // v2 - Active
        v2 := api.Group("/v2")
        {
            // Current routes
        }
    }
}
```

### Step 2: Client Receives 410 Gone

```bash
curl -i https://api.promenade.example.com/api/v1/customers
```

**Response**:
```http
HTTP/1.1 410 Gone
Content-Type: application/json
Link: </api/v2>; rel="successor-version"

{
  "status": "error",
  "error": {
    "code": "API_VERSION_SUNSET",
    "message": "API version was sunset on 2027-06-01. Please migrate to the successor version.",
    "details": {
      "sunset_date": "2027-06-01",
      "successor_version": "/api/v2",
      "migration_guide": "https://docs.promenade.example.com/migration/v1-to-v2"
    }
  }
}
```

---

## Version Detection API

### Handler Implementation

```go
// cmd/api/server.go
func (s *Server) SetupRoutes() {
    api := s.router.Group("/api")
    
    // Version info endpoint (no version prefix)
    api.GET("", s.versionInfoHandler)
    
    {
        v1 := api.Group("/v1")
        v1.GET("", s.v1InfoHandler)
        // ... routes
        
        v2 := api.Group("/v2")
        v2.GET("", s.v2InfoHandler)
        // ... routes
    }
}

// versionInfoHandler returns all available API versions
func (s *Server) versionInfoHandler(c *gin.Context) {
    response.Success(c, gin.H{
        "versions": []gin.H{
            {
                "version":    "v1",
                "status":     "deprecated",
                "deprecated": true,
                "sunset_date": "2027-06-01",
                "url":        "/api/v1",
                "docs":       "https://docs.promenade.example.com/api/v1",
            },
            {
                "version":    "v2",
                "status":     "active",
                "deprecated": false,
                "url":        "/api/v2",
                "docs":       "https://docs.promenade.example.com/api/v2",
            },
        },
        "current_version": "v2",
    })
}

// v1InfoHandler returns v1 status
func (s *Server) v1InfoHandler(c *gin.Context) {
    response.Success(c, gin.H{
        "version":     "v1",
        "status":      "deprecated",
        "deprecated":  true,
        "sunset_date": "2027-06-01",
        "successor":   "/api/v2",
        "migration_guide": "https://docs.promenade.example.com/migration/v1-to-v2",
    })
}

// v2InfoHandler returns v2 status
func (s *Server) v2InfoHandler(c *gin.Context) {
    response.Success(c, gin.H{
        "version":    "v2",
        "status":     "active",
        "deprecated": false,
    })
}
```

### Client Usage

```bash
# Get all versions
curl https://api.promenade.example.com/api

# Get v1 status
curl https://api.promenade.example.com/api/v1

# Get v2 status
curl https://api.promenade.example.com/api/v2
```

**Response** (GET /api):
```json
{
  "status": "success",
  "data": {
    "versions": [
      {
        "version": "v1",
        "status": "deprecated",
        "deprecated": true,
        "sunset_date": "2027-06-01",
        "url": "/api/v1",
        "docs": "https://docs.promenade.example.com/api/v1"
      },
      {
        "version": "v2",
        "status": "active",
        "deprecated": false,
        "url": "/api/v2",
        "docs": "https://docs.promenade.example.com/api/v2"
      }
    ],
    "current_version": "v2"
  }
}
```

---

## Swagger Annotations

### Deprecated Endpoint Example

```go
// internal/contexts/identity/adapter/http/handler/user_handler.go

// GetUser retrieves user by ID (v1 - DEPRECATED)
//
// @Summary      Get user by ID (DEPRECATED)
// @Description   DEPRECATED: This endpoint will be removed on 2027-06-01.
// @Description  **Successor**: GET /api/v2/users/:id
// @Description  **Migration Guide**: https://docs.promenade.example.com/migration/v1-to-v2#users
// @Description  **Breaking Changes**:
// @Description  - Response field `fullName` → `full_name`
// @Description  - Response field `createdDate` → `created_at`
// @Description  - Response field `isActive` → `is_active`
// @Tags         Users (Deprecated)
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "User ID"
// @Success      200  {object} response.SuccessResponse{data=dto.UserResponseV1}
// @Failure      404  {object} response.ErrorResponse
// @Failure      500  {object} response.ErrorResponse
// @Deprecated
// @Router       /api/v1/users/{id} [get]
// @Security     BearerAuth
func (h *UserHandler) GetUserV1(c *gin.Context) {
    // Implementation
}
```

### Active Endpoint Example (v2)

```go
// internal/contexts/identity/adapter/http/handler/user_handler.go

// GetUser retrieves user by ID
//
// @Summary      Get user by ID
// @Description  Returns user details by ID. Requires authentication.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "User ID"
// @Success      200  {object} response.SuccessResponse{data=dto.UserResponse}
// @Failure      404  {object} response.ErrorResponse
// @Failure      500  {object} response.ErrorResponse
// @Router       /api/v2/users/{id} [get]
// @Security     BearerAuth
func (h *UserHandler) GetUser(c *gin.Context) {
    // Implementation
}
```

### Swagger UI Display

The `@Deprecated` tag will show:
-  Warning icon next to endpoint
- "Deprecated" badge in Swagger UI
- Deprecation notice in description
- Migration guide link
- Successor endpoint link

---

## Testing Deprecated Endpoints

### Unit Test Example

```go
// internal/contexts/customer-mgmt/adapter/http/handler/customer_handler_test.go

func TestDeprecatedCustomerList_ReturnsDeprecationHeaders(t *testing.T) {
    router := setupTestRouter()
    
    // Apply deprecation middleware
    v1 := router.Group("/api/v1")
    v1.Use(middleware.DeprecateEndpoint(
        time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC),
        "/api/v2/customers",
    ))
    v1.GET("/customers", handler.ListV1)

    // Make request
    req := httptest.NewRequest("GET", "/api/v1/customers", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assert deprecation headers
    assert.Equal(t, "true", w.Header().Get("Deprecation"))
    assert.Equal(t, "Sun, 01 Jun 2027 00:00:00 GMT", w.Header().Get("Sunset"))
    assert.Equal(t, "</api/v2/customers>; rel=\"successor-version\"", w.Header().Get("Link"))
    
    // Response still works
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Integration Test Example

```go
// test/integration/contexts/customer-mgmt/customer/versioning_test.go

func TestCustomerAPI_V1Deprecated_V2Active(t *testing.T) {
    db := integration.SetupTestDB(t)
    defer db.Close()

    server := setupTestServer(db)

    t.Run("v1 returns deprecation headers", func(t *testing.T) {
        resp, err := http.Get(server.URL + "/api/v1/customers")
        require.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusOK, resp.StatusCode)
        assert.Equal(t, "true", resp.Header.Get("Deprecation"))
        assert.NotEmpty(t, resp.Header.Get("Sunset"))
    })

    t.Run("v2 active without deprecation", func(t *testing.T) {
        resp, err := http.Get(server.URL + "/api/v2/customers")
        require.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusOK, resp.StatusCode)
        assert.Empty(t, resp.Header.Get("Deprecation"))
        assert.Empty(t, resp.Header.Get("Sunset"))
    })
}
```

---

## Monitoring Version Usage

### Log Version in Middleware

```go
// cmd/api/server.go
func (s *Server) SetupRoutes() {
    // Add version logger to all API routes
    api := s.router.Group("/api")
    api.Use(middleware.VersionLogger())
    
    // ... version groups
}
```

### Extract Version in Handlers (Optional)

```go
func (h *CustomerHandler) List(c *gin.Context) {
    // Get version from context
    version, exists := c.Get("api_version")
    if exists {
        logger.FromContext(c.Request.Context()).Info("API request",
            slog.String("version", version.(string)),
            slog.String("endpoint", c.Request.URL.Path),
        )
    }
    
    // Handler logic
}
```

### Prometheus Metrics (Future)

```go
// pkg/middleware/metrics.go
func MetricsMiddleware() gin.HandlerFunc {
    // Track API version usage
    apiVersionRequests := promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_version_requests_total",
            Help: "Total API requests by version",
        },
        []string{"version", "endpoint", "method"},
    )
    
    return func(c *gin.Context) {
        version := extractVersionFromPath(c.Request.URL.Path)
        c.Next()
        
        apiVersionRequests.WithLabelValues(
            version,
            c.Request.URL.Path,
            c.Request.Method,
        ).Inc()
    }
}
```

---

## Related Documentation

- [API Versioning Strategy](api-versioning.md) - Complete policy and guidelines
- [Migration Guide Template](../migration/README.md) - Template for v1→v2 migration
- [Swagger Documentation](../reference/api-reference.md) - Full API reference
- [Testing Guide](testing-patterns.md) - Testing strategy

---

**Last Updated**: January 5, 2026  
**Status**: Production-ready  
**Maintainer**: Promenade Team

