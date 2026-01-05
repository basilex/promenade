# Response Package

**Standard HTTP response helpers for consistent API responses**

---

## Overview

The `response` package provides **standardized HTTP response helpers** for Gin framework, ensuring consistent JSON response format across all API endpoints. Every response follows the same structure with proper HTTP status codes.

**Status**: Production-ready  
**Tests**: 12 tests, 100% coverage  
**Location**: `pkg/response/`

---

## Features

- **Consistent Format** - All responses follow standard `Response` struct
- **Error Handling** - Standardized error responses with error codes
- **Type-Safe** - Uses generics for type-safe data responses
- **HTTP Status Codes** - Automatic status code mapping
- **Gin Integration** - Seamless integration with Gin framework
- **Pagination Support** - Built-in pagination metadata

---

## Response Format

### Success Response

```json
{
  "status": "success",
  "data": {
    "id": "01JGABC...",
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

### Error Response

```json
{
  "status": "error",
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID 01JGABC... not found"
  }
}
```

### Paginated Response

```json
{
  "status": "success",
  "data": [
    {"id": "01JGABC...", "name": "User 1"},
    {"id": "01JGDEF...", "name": "User 2"}
  ],
  "pagination": {
    "total": 150,
    "page": 1,
    "page_size": 20,
    "total_pages": 8
  }
}
```

---

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/pkg/response"
)

func main() {
    r := gin.Default()
    
    r.GET("/users/:id", func(c *gin.Context) {
        user := GetUser(c.Param("id"))
        if user == nil {
            response.NotFound(c, "USER_NOT_FOUND", "User not found")
            return
        }
        response.Success(c, user)
    })
    
    r.Run()
}
```

---

## Usage Examples

### Success Response (200 OK)

```go
func (h *UserHandler) GetByID(c *gin.Context) {
    userID, _ := uuidv7.Parse(c.Param("id"))
    
    user, err := h.usecase.GetUser(c.Request.Context(), userID)
    if err != nil {
        response.InternalError(c, "INTERNAL_ERROR", err.Error())
        return
    }
    
    response.Success(c, user)
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC...",
    "email": "john@example.com",
    "name": "John Doe"
  }
}
```

### Created Response (201 Created)

```go
func (h *UserHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "VALIDATION_ERROR", err.Error())
        return
    }
    
    user, err := h.usecase.Register(ctx, req.Email, req.Name, req.Password)
    if err != nil {
        response.InternalError(c, "REGISTRATION_FAILED", err.Error())
        return
    }
    
    response.Created(c, user)
}
```

### Error Responses

#### Validation Error (400 Bad Request)

```go
if req.Email == "" {
    response.BadRequest(c, "VALIDATION_ERROR", "Email is required")
    return
}
```

#### Not Found (404 Not Found)

```go
if errors.Is(err, ErrUserNotFound) {
    response.NotFound(c, "USER_NOT_FOUND", err.Error())
    return
}
```

#### Conflict (409 Conflict)

```go
if errors.Is(err, ErrEmailAlreadyExists) {
    response.Conflict(c, "EMAIL_EXISTS", "Email already registered")
    return
}
```

#### Unauthorized (401 Unauthorized)

```go
if !validCredentials {
    response.Unauthorized(c, "INVALID_CREDENTIALS", "Invalid email or password")
    return
}
```

#### Forbidden (403 Forbidden)

```go
if !userHasPermission {
    response.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "Admin access required")
    return
}
```

#### Internal Error (500 Internal Server Error)

```go
if err != nil {
    response.InternalError(c, "INTERNAL_ERROR", err.Error())
    return
}
```

---

## API Reference

### Success Responses

#### Success (200 OK)

```go
func Success[T any](c *gin.Context, data T)
```

Returns successful response with data.

#### Created (201 Created)

```go
func Created[T any](c *gin.Context, data T)
```

Returns created response (for POST requests).

#### NoContent (204 No Content)

```go
func NoContent(c *gin.Context)
```

Returns empty successful response (for DELETE).

### Error Responses

All error responses follow the same signature:

```go
func ErrorName(c *gin.Context, code, message string)
```

- **BadRequest (400)** - Validation errors
- **Unauthorized (401)** - Authentication failures
- **Forbidden (403)** - Authorization failures
- **NotFound (404)** - Resource not found
- **Conflict (409)** - Resource conflicts
- **InternalError (500)** - Server errors

### Pagination

```go
func Paginated[T any](c *gin.Context, data []T, pagination Pagination)
```

Returns paginated response with metadata.

**Example**:
```go
response.Paginated(c, users, response.Pagination{
    Total:      150,
    Page:       1,
    PageSize:   20,
    TotalPages: 8,
})
```

---

## Complete Handler Example

```go
package handler

import (
    "errors"
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type UserHandler struct {
    usecase user.IUseCase
}

// List users with pagination
func (h *UserHandler) List(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    pageSize := c.DefaultQuery("page_size", "20")
    
    users, total, err := h.usecase.ListUsers(c.Request.Context(), page, pageSize)
    if err != nil {
        response.InternalError(c, "INTERNAL_ERROR", err.Error())
        return
    }
    
    response.Paginated(c, users, response.Pagination{
        Total:      total,
        Page:       pageInt,
        PageSize:   pageSizeInt,
        TotalPages: (total + pageSizeInt - 1) / pageSizeInt,
    })
}

// Get user by ID
func (h *UserHandler) GetByID(c *gin.Context) {
    userID, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "INVALID_ID", "Invalid user ID format")
        return
    }
    
    user, err := h.usecase.GetUser(c.Request.Context(), userID)
    if errors.Is(err, ErrUserNotFound) {
        response.NotFound(c, "USER_NOT_FOUND", "User not found")
        return
    }
    if err != nil {
        response.InternalError(c, "INTERNAL_ERROR", err.Error())
        return
    }
    
    response.Success(c, user)
}

// Create user
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "VALIDATION_ERROR", err.Error())
        return
    }
    
    user, err := h.usecase.CreateUser(c.Request.Context(), req.Email, req.Name)
    if errors.Is(err, ErrEmailAlreadyExists) {
        response.Conflict(c, "EMAIL_EXISTS", "Email already registered")
        return
    }
    if err != nil {
        response.InternalError(c, "CREATE_FAILED", err.Error())
        return
    }
    
    response.Created(c, user)
}

// Delete user
func (h *UserHandler) Delete(c *gin.Context) {
    userID, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "INVALID_ID", "Invalid user ID format")
        return
    }
    
    err = h.usecase.DeleteUser(c.Request.Context(), userID)
    if errors.Is(err, ErrUserNotFound) {
        response.NotFound(c, "USER_NOT_FOUND", "User not found")
        return
    }
    if err != nil {
        response.InternalError(c, "DELETE_FAILED", err.Error())
        return
    }
    
    response.NoContent(c)
}
```

---

## Error Code Conventions

**Format**: `ENTITY_ACTION` or `ERROR_TYPE`

**Examples**:
- `USER_NOT_FOUND` - User does not exist
- `EMAIL_EXISTS` - Email already registered
- `VALIDATION_ERROR` - Request validation failed
- `INVALID_CREDENTIALS` - Login failed
- `TOKEN_EXPIRED` - JWT token expired
- `INSUFFICIENT_PERMISSIONS` - RBAC check failed
- `INTERNAL_ERROR` - Generic server error

---

## Best Practices

### DO

 **Use specific error codes**:
```go
response.NotFound(c, "USER_NOT_FOUND", "User not found")
```

 **Check domain errors**:
```go
if errors.Is(err, ErrUserNotFound) {
    response.NotFound(c, "USER_NOT_FOUND", err.Error())
    return
}
```

 **Provide clear error messages**:
```go
response.BadRequest(c, "VALIDATION_ERROR", "Email must be a valid email address")
```

 **Use pagination for lists**:
```go
response.Paginated(c, users, pagination)
```

### DON'T

 **Don't expose internal errors**:
```go
// Bad
response.InternalError(c, "DB_ERROR", "pq: duplicate key value violates unique constraint")

// Good
response.Conflict(c, "EMAIL_EXISTS", "Email already registered")
```

 **Don't use generic error codes**:
```go
// Bad
response.NotFound(c, "NOT_FOUND", "Not found")

// Good
response.NotFound(c, "USER_NOT_FOUND", "User with ID 01JGABC... not found")
```

---

## Testing

```go
func TestResponse_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    data := map[string]string{"message": "hello"}
    response.Success(c, data)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), `"status":"success"`)
    assert.Contains(t, w.Body.String(), `"message":"hello"`)
}
```

---

## Related Documentation

- [API Reference](/guide/api-reference) - Complete endpoint documentation
- [Testing Guide](/guide/testing) - Handler testing patterns
- [Contributing Guide](/guide/contributing) - Development workflow
- [GitHub Repository](https://github.com/basilex/promenade/tree/dev/pkg/response) - Source code
