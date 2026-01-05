# API Documentation (Swagger/OpenAPI)

**Interactive API documentation** with Swagger UI for exploring and testing all 120+ endpoints in Promenade Platform.

---

## Overview

Promenade uses **OpenAPI 3.0** specification with **Swagger UI** to provide interactive, self-documenting REST API. All endpoints across 5 bounded contexts are automatically documented from code annotations.

**Key Features**:
-  120+ documented endpoints
-  Interactive Swagger UI for testing
-  OpenAPI 3.0 JSON/YAML export
-  Request/Response schemas with examples
-  Authentication flows (JWT Bearer tokens)
-  Error code documentation
-  Auto-generated from code annotations

---

## Quick Access

### Swagger UI

Start the server and open Swagger UI in browser:

```bash
make dev
# Open: http://localhost:8081/api/docs/index.html
```

**Swagger UI provides**:
- Interactive API explorer
- Try-it-out functionality
- Request/response examples
- Authentication management
- Schema documentation

### OpenAPI Specification Files

Generated files (auto-updated on build):

```
docs/swagger/
 docs.go         # Go embedded documentation (346KB)
 swagger.json    # JSON OpenAPI 3.0 spec (345KB)
 swagger.yaml    # YAML OpenAPI 3.0 spec (173KB)
```

**Export formats**:
- **JSON**: http://localhost:8081/api/docs/doc.json
- **YAML**: http://localhost:8081/api/docs/swagger.yaml
- **Go code**: `docs/swagger/docs.go` (embedded in binary)

---

## Architecture

### Documentation Generation Flow

```
Go Source Code
 ↓
Swagger Annotations (@title, @description, @Success, @Failure)
 ↓
swaggo/swag tool (swag init)
 ↓
OpenAPI 3.0 Specification (JSON/YAML)
 ↓
Swagger UI (gin-swagger middleware)
 ↓
Interactive Documentation (browser)
```

### Tools & Dependencies

**Generation Tool**:
- **swaggo/swag** v1.16.6 - Swagger documentation generator for Go
- Installation: `go install github.com/swaggo/swag/cmd/swag@latest`

**Runtime Dependencies**:
- **gin-swagger** v1.6.1 - Gin middleware for Swagger UI
- **swaggo/files** v1.0.1 - Static file embedder

**Configuration** (cmd/api/main.go):

```go
// @title Promenade Platform
// @version 0.1.0
// @description Modern backend platform for customer management, orders, and business workflows
// @contact.name API Support
// @contact.email alexander.vasilenko@gmail.com
// @license.name MIT
// @host localhost:8081
// @BasePath /api/v1
```

---

## Writing API Documentation

### Handler Annotation Pattern

**Standard pattern** for all handlers:

```go
// CreateOrder godoc
// @Summary Create new order
// @Description Create a new order for a customer with specified currency
// @Tags orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order creation request"
// @Success 201 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response "Validation error"
// @Failure 404 {object} response.Response "Customer not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /order-mgmt/orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
    // Handler implementation
}
```

### Annotation Reference

**Required Annotations**:

| Annotation       | Purpose                           | Example                                        |
| ---------------- | --------------------------------- | ---------------------------------------------- |
| `@Summary`       | Short description (1 line)        | `@Summary Create new order`                    |
| `@Description`   | Detailed explanation              | `@Description Create order with line items`    |
| `@Tags`          | Group endpoints                   | `@Tags orders`                                 |
| `@Accept`        | Request content type              | `@Accept json`                                 |
| `@Produce`       | Response content type             | `@Produce json`                                |
| `@Success`       | Successful response               | `@Success 200 {object} response.Response`      |
| `@Failure`       | Error response                    | `@Failure 404 {object} response.Response`      |
| `@Router`        | Route path and method             | `@Router /orders [post]`                       |

**Optional Annotations**:

| Annotation    | Purpose                    | Example                                    |
| ------------- | -------------------------- | ------------------------------------------ |
| `@Param`      | Request parameters         | `@Param id path string true "Order ID"`    |
| `@Security`   | Authentication required    | `@Security BearerAuth`                     |
| `@Deprecated` | Mark endpoint as deprecated | `@Deprecated`                             |

### Response Type Patterns

**Success with data**:

```go
// @Success 200 {object} response.Response{data=OrderResponse}
```

**Success with array**:

```go
// @Success 200 {object} response.Response{data=[]OrderResponse}
```

**Success with pagination**:

```go
// @Success 200 {object} response.Response{data=[]OrderResponse,pagination=response.Pagination}
```

**Error response**:

```go
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response "Customer not found"
// @Failure 500 {object} response.Response "Internal server error"
```

** Common Mistakes**:

```go
//  WRONG - ErrorResponse doesn't exist as type
// @Failure 400 {object} response.ErrorResponse

//  CORRECT - Use response.Response
// @Failure 400 {object} response.Response
```

---

## Documented Endpoints

### Endpoint Coverage by Context

| Context                | Aggregates                   | Endpoints | Documentation |
| ---------------------- | ---------------------------- | --------- | ------------- |
| **Identity**           | User, Contact, Profile, Role, Permission | 35        |  Complete    |
| **Customer Management** | Customer, Company, Deal, Interaction, Analytics | 48        |  Complete    |
| **Order Management**   | Order (full lifecycle)       | 14        |  Complete    |
| **Billing**            | Invoice, Payment, Subscription | 22        |  Complete    |
| **Shared**             | Country, Currency, Language, Timezone | 9         |  Complete    |
| **Infrastructure**     | Health Checks                | 4         |  Complete    |

**Total**: **120+ endpoints** across **5 bounded contexts**

### Documentation by Aggregate

**Identity Context** (35 endpoints):
- User: 8 endpoints (register, login, CRUD, password management)
- Contact: 8 endpoints (email, phone, address management)
- Profile: 9 endpoints (personal info, social links, localization)
- Role: 7 endpoints (RBAC role management)
- Permission: 7 endpoints (RBAC permission management)

**Customer Management Context** (48 endpoints):
- Customer: 14 endpoints (lifecycle, segmentation, B2C/B2B)
- Company: 8 endpoints (B2B organizations, hierarchies)
- Deal: 12 endpoints (sales pipeline, stage management)
- Interaction: 14 endpoints (calls, emails, meetings, follow-ups)
- Analytics: 8 endpoints (dashboards, metrics, time series)

**Order Management Context** (14 endpoints):
- Order: 14 endpoints (creation, lifecycle, line items)

**Billing Context** (22 endpoints):
- Invoice: 8 endpoints (generation, line items, status management)
- Payment: 14 endpoints (processing, refunds, provider integration)
- Subscription: 8 endpoints (recurring billing, lifecycle)

**Shared Context** (9 endpoints):
- Country: 3 endpoints (reference data)
- Currency: 2 endpoints (reference data)
- Language: 2 endpoints (reference data)
- Timezone: 2 endpoints (reference data)

**Infrastructure** (4 endpoints):
- Health: 4 endpoints (overall, database, redis, event bus)

---

## Regenerating Documentation

### Automatic Regeneration

Documentation regenerates automatically on every build:

```bash
# Development workflow
make dev        # Starts dev server (auto-regenerates docs)
make build      # Builds binary (auto-regenerates docs)
```

### Manual Regeneration

```bash
# Regenerate only Swagger docs
swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal

# Or use GOPATH/bin/swag if not in PATH
$GOPATH/bin/swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal
```

**Flags explained**:
- `-g` - Entry point file with API metadata
- `-o` - Output directory for generated files
- `--parseDependency` - Parse vendor dependencies
- `--parseInternal` - Parse internal packages

---

## Authentication in Swagger UI

### JWT Bearer Token Setup

1. **Obtain JWT token** via `/api/v1/identity/users/login`:

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": { ... }
  }
}
```

2. **Add to Swagger UI**:
   - Click **"Authorize"** button (lock icon)
   - Enter: `Bearer <access_token>`
   - Click **"Authorize"**
   - All subsequent requests include JWT automatically

### Protected Endpoints

Endpoints requiring authentication are marked with lock icon  in Swagger UI.

**Example**:
-  `GET /api/v1/identity/users` - Requires JWT
-  `POST /api/v1/identity/users/register` - Public
-  `POST /api/v1/identity/users/login` - Public

---

## Response Schema

### Standard Response Format

All API responses follow consistent structure:

**Success Response**:

```json
{
  "status": "success",
  "data": { ... }
}
```

**Error Response**:

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format"
  }
}
```

**Paginated Response**:

```json
{
  "status": "success",
  "data": [ ... ],
  "pagination": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "pages": 5
  }
}
```

### Response Types (pkg/response)

**Core Types**:

```go
type Response struct {
    Status     string      `json:"status"`               // "success" or "error"
    Data       interface{} `json:"data,omitempty"`       // Response payload
    Error      *Error      `json:"error,omitempty"`      // Error details
    Message    string      `json:"message,omitempty"`    // Optional message
    Pagination *Pagination `json:"pagination,omitempty"` // Pagination info
}

type Error struct {
    Code    string `json:"code"`    // Error code (e.g., "VALIDATION_ERROR")
    Message string `json:"message"` // Human-readable message
}

type Pagination struct {
    Total    int64 `json:"total"`     // Total records
    Page     int   `json:"page"`      // Current page (1-based)
    PageSize int   `json:"page_size"` // Records per page
    Pages    int   `json:"pages"`     // Total pages
}
```

---

## Error Codes

### Standard HTTP Status Codes

| Status | Code | Usage                           |
| ------ | ---- | ------------------------------- |
| 200    | OK   | Successful GET/PUT/DELETE       |
| 201    | Created | Successful POST               |
| 204    | No Content | Successful DELETE (Shared) |
| 400    | Bad Request | Validation errors          |
| 401    | Unauthorized | Missing/invalid JWT       |
| 403    | Forbidden | Insufficient permissions     |
| 404    | Not Found | Resource not found           |
| 409    | Conflict | Duplicate resource            |
| 500    | Internal Server Error | Server error    |

### Error Code Patterns

**Identity Context**:
- `USER_NOT_FOUND`, `CONTACT_NOT_FOUND`, `PROFILE_NOT_FOUND`
- `INVALID_CREDENTIALS`, `WEAK_PASSWORD`
- `EMAIL_ALREADY_EXISTS`, `PHONE_ALREADY_EXISTS`

**Customer Management Context**:
- `CUSTOMER_NOT_FOUND`, `COMPANY_NOT_FOUND`, `DEAL_NOT_FOUND`
- `INVALID_STATE_TRANSITION`
- `VALIDATION_ERROR`

**Order Management Context**:
- `ORDER_NOT_FOUND`, `ORDER_ALREADY_CONFIRMED`
- `INVALID_STATUS_TRANSITION`
- `NO_LINE_ITEMS`

**Billing Context**:
- `INVOICE_NOT_FOUND`, `PAYMENT_NOT_FOUND`, `SUBSCRIPTION_NOT_FOUND`
- `INVOICE_ALREADY_PAID`, `PAYMENT_ALREADY_COMPLETED`
- `INVALID_AMOUNT`

**Shared Context** (unique pattern):
- `COUNTRY_NOT_FOUND`, `CURRENCY_NOT_FOUND`, `LANGUAGE_NOT_FOUND`, `TIMEZONE_NOT_FOUND`
- `VALIDATION_ERROR` (not BAD_REQUEST)
- `DELETE_ERROR` (500 status for delete failures)

---

## Testing API with Swagger UI

### Try-It-Out Feature

1. **Navigate to endpoint** in Swagger UI
2. **Click "Try it out"** button
3. **Fill in parameters**:
   - Path parameters (e.g., `id`)
   - Query parameters (e.g., `page`, `page_size`)
   - Request body (JSON editor)
4. **Click "Execute"**
5. **View response**:
   - HTTP status code
   - Response body
   - Response headers
   - cURL command equivalent

### Example Workflow

**1. Register new user**:
- Endpoint: `POST /api/v1/identity/users/register`
- Body:
  ```json
  {
    "email": "test@example.com",
    "name": "Test User",
    "password": "SecurePass123"
  }
  ```

**2. Login to get JWT**:
- Endpoint: `POST /api/v1/identity/users/login`
- Body:
  ```json
  {
    "email": "test@example.com",
    "password": "SecurePass123"
  }
  ```
- Copy `access_token` from response

**3. Authorize Swagger UI**:
- Click "Authorize" button
- Enter: `Bearer <access_token>`

**4. Create customer** (protected endpoint):
- Endpoint: `POST /api/v1/customer-mgmt/customers`
- Body:
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com",
    "status": "lead",
    "tier": "free",
    "source": "website"
  }
  ```

**5. List customers** (protected endpoint):
- Endpoint: `GET /api/v1/customer-mgmt/customers`
- Query params: `page=1`, `page_size=20`

---

## Integration with Tools

### Postman Collection

**Pre-built Postman collection** with 120+ endpoints, authentication flow, and test scripts.

**Quick Import**:

1. Open Postman
2. Click **Import**
3. Select files from `postman/` directory:
   - `Promenade_API.postman_collection.json` (collection)
   - `Development.postman_environment.json` (environment)
4. Select **Development** environment (top-right dropdown)

**Or import via URL**:
```
https://raw.githubusercontent.com/basilex/promenade/dev/postman/Promenade_API.postman_collection.json
```

**Features**:
-  120+ endpoints organized by context
-  Auto-save JWT tokens after login
-  Auto-refresh expired tokens
-  Pre-configured environments (Dev/Staging/Prod)
-  Test scripts for response validation
-  Authentication flow examples

**Complete guide**: [postman/README.md](../../postman/README.md)

**Alternative: Import OpenAPI spec directly**:

1. Open Postman
2. **Import** → **Link**
3. Enter: `http://localhost:8081/api/docs/doc.json`
4. **Import**

**Manual environment setup** (if not using pre-built):

```json
{
  "name": "Promenade Dev",
  "values": [
    {"key": "base_url", "value": "http://localhost:8081"},
    {"key": "api_version", "value": "v1"},
    {"key": "access_token", "value": ""},
    {"key": "refresh_token", "value": ""}
  ]
}
```

### Insomnia

**Import OpenAPI spec**:

1. Open Insomnia
2. **Application** → **Preferences** → **Data** → **Import Data**
3. Select `docs/swagger/swagger.yaml`
4. **Import**

### VS Code REST Client

**Example `.http` file**:

```http
### Variables
@baseUrl = http://localhost:8081/api/v1
@accessToken = your-jwt-token-here

### Register User
POST {{baseUrl}}/identity/users/register
Content-Type: application/json

{
  "email": "test@example.com",
  "name": "Test User",
  "password": "SecurePass123"
}

### Login
POST {{baseUrl}}/identity/users/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "SecurePass123"
}

### List Customers (Protected)
GET {{baseUrl}}/customer-mgmt/customers?page=1&page_size=20
Authorization: Bearer {{accessToken}}
```

---

## Troubleshooting

### Common Issues

**1. Swagger UI shows "Failed to load API definition"**

**Solution**: Regenerate documentation:

```bash
swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal
make build
```

**2. Endpoints not showing up**

**Causes**:
- Missing Swagger annotations in handler
- Handler not registered in router
- Annotation syntax errors

**Check**:

```bash
# Search for handler without annotations
grep -r "func.*Handler.*gin.Context" internal/contexts --include="*.go" -A 5 | grep -v "@Summary"
```

**3. Type definition errors during generation**

**Example error**:
```
cannot find type definition: response.ErrorResponse
```

**Solution**: Use correct types from `pkg/response`:
-  `response.Response`
-  `response.ErrorResponse` (function, not type)
-  `response.SuccessResponse` (function, not type)

**4. Authentication not working in Swagger UI**

**Checklist**:
-  JWT token obtained via `/api/v1/identity/users/login`
-  Token format: `Bearer <access_token>` (with space)
-  Token not expired (15 minutes validity)
-  "Authorize" button clicked after entering token

---

## Best Practices

### Documentation Standards

**DO**:
-  Write clear, concise summaries (1 line)
-  Add detailed descriptions for complex operations
-  Document all parameters (path, query, body)
-  Include example request/response bodies
-  Document all possible error codes
-  Group related endpoints with same `@Tags`
-  Keep annotations up-to-date with code changes

**DON'T**:
-  Skip `@Summary` or `@Description`
-  Use incorrect type references
-  Forget to document error responses
-  Leave endpoints without `@Tags`
-  Hardcode URLs in descriptions (use relative paths)

### Versioning

**Current version**: v1 (`/api/v1/`)

**Future versioning strategy**:
- Breaking changes → new version (`/api/v2/`)
- Non-breaking changes → same version
- Deprecation warnings in Swagger UI
- Migration guide in documentation

---

## Development Workflow

### Adding New Endpoint

1. **Create handler** with Swagger annotations:

```go
// CreateWidget godoc
// @Summary Create new widget
// @Tags widgets
// @Accept json
// @Produce json
// @Param request body CreateWidgetRequest true "Widget data"
// @Success 201 {object} response.Response{data=WidgetResponse}
// @Failure 400 {object} response.Response
// @Router /widgets [post]
func (h *WidgetHandler) Create(c *gin.Context) {
    // Implementation
}
```

2. **Define DTOs** with struct tags:

```go
type CreateWidgetRequest struct {
    Name        string `json:"name" binding:"required" example:"My Widget"`
    Description string `json:"description" example:"Widget description"`
    Price       int    `json:"price" binding:"required,min=0" example:"1999"`
}

type WidgetResponse struct {
    ID          string    `json:"id" example:"01JGABC123..."`
    Name        string    `json:"name" example:"My Widget"`
    Price       int       `json:"price" example:"1999"`
    CreatedAt   time.Time `json:"created_at" example:"2026-01-05T10:00:00Z"`
}
```

3. **Register route** in router:

```go
widgets := api.Group("/widgets")
{
    widgets.POST("", handler.Create)
}
```

4. **Regenerate docs** and test:

```bash
make build
make dev
# Open: http://localhost:8081/api/docs/index.html
```

---

## Configuration

### Swagger Metadata (cmd/api/main.go)

```go
// @title Promenade Platform
// @version 0.1.0
// @description Modern backend platform for customer management, orders, and business workflows with clean DDD architecture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/basilex/promenade
// @contact.email alexander.vasilenko@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1
// @schemes http https
```

### Swagger UI Route (cmd/api/server.go)

```go
import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "github.com/basilex/promenade/docs/swagger" // Import generated docs
)

func (s *Server) SetupRoutes() {
    // Swagger UI documentation
    s.router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    // API routes...
}
```

---

## Related Documentation

- [Main README](../../README.md) - Project overview
- [API Reference](../reference/api-reference.md) - Endpoint specifications
- [Testing Guide](../../test/README.md) - API testing patterns
- [JWT Authentication](../../pkg/jwt/README.md) - Authentication details
- [Rate Limiting](rate-limiting.md) - API protection
- [RBAC Guide](rbac.md) - Authorization patterns

---

## External Resources

- [OpenAPI Specification 3.0](https://swagger.io/specification/)
- [swaggo/swag Documentation](https://github.com/swaggo/swag)
- [gin-swagger Documentation](https://github.com/swaggo/gin-swagger)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)

---

**Last Updated**: January 5, 2026  
**Status**: Production-ready  
**Coverage**: 120+ endpoints documented  
**Maintainer**: Promenade Team
