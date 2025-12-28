# Response Package

**Purpose:** Standard HTTP response helpers for consistent API responses  
**Status:** Production-ready  
**Tests:** 12 tests, 100% coverage

---

## Overview

The `response` package provides **standardized HTTP response helpers** for Gin framework, ensuring consistent JSON response format across all API endpoints. Every response follows the same structure with proper HTTP status codes.

## Features

- **Consistent Format:** All responses follow standard `Response` struct
- **Error Handling:** Standardized error responses with error codes
- **Type-Safe:** Uses generics for type-safe data responses
- **HTTP Status Codes:** Automatic status code mapping
- **Gin Integration:** Seamless integration with Gin framework

---

## Installation

```go
import "github.com/basilex/promenade/pkg/response"
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

## Response Format

### Success Response

```json
{
  "status": "success",
  "data": {
    "id": "01JGABC...",
    "name": "John Doe"
  }
}
```

### Error Response

```json
{
  "status": "error",
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found"
  }
}
```

---

## Usage Examples

### 1. Success Response (200 OK)

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

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC123",
    "email": "john@example.com",
    "name": "John Doe"
  }
}
```

### 2. Created Response (201 Created)

```go
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "VALIDATION_ERROR", err.Error())
        return
    }
    
    user, err := h.usecase.CreateUser(c.Request.Context(), req.Email, req.Name)
    if err != nil {
        response.InternalError(c, "INTERNAL_ERROR", err.Error())
        return
    }
    
    response.Created(c, user)
}
```

**Response:**
```json
HTTP/1.1 201 Created
{
  "status": "success",
  "data": {
    "id": "01JGABC456",
    "email": "jane@example.com",
    "name": "Jane Doe"
  }
}
```

### 3. Bad Request (400)

```go
func (h *UserHandler) Update(c *gin.Context) {
    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "VALIDATION_ERROR", "Invalid request body")
        return
    }
    
    // Update logic...
}
```

**Response:**
```json
HTTP/1.1 400 Bad Request
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body"
  }
}
```

### 4. Not Found (404)

```go
func (h *UserHandler) GetByID(c *gin.Context) {
    userID, _ := uuidv7.Parse(c.Param("id"))
    
    user, err := h.usecase.GetUser(c.Request.Context(), userID)
    if errors.Is(err, ErrUserNotFound) {
        response.NotFound(c, "USER_NOT_FOUND", "User not found")
        return
    }
    
    response.Success(c, user)
}
```

**Response:**
```json
HTTP/1.1 404 Not Found
{
  "status": "error",
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found"
  }
}
```

### 5. Unauthorized (401)

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            response.Unauthorized(c, "UNAUTHORIZED", "Missing authorization token")
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

**Response:**
```json
HTTP/1.1 401 Unauthorized
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Missing authorization token"
  }
}
```

### 6. Forbidden (403)

```go
func (h *UserHandler) Delete(c *gin.Context) {
    requestingUserID := c.GetString("user_id")
    targetUserID := c.Param("id")
    
    if requestingUserID != targetUserID && !isAdmin(requestingUserID) {
        response.Forbidden(c, "FORBIDDEN", "You cannot delete this user")
        return
    }
    
    // Delete logic...
}
```

**Response:**
```json
HTTP/1.1 403 Forbidden
{
  "status": "error",
  "error": {
    "code": "FORBIDDEN",
    "message": "You cannot delete this user"
  }
}
```

### 7. Internal Server Error (500)

```go
func (h *UserHandler) Create(c *gin.Context) {
    user, err := h.usecase.CreateUser(c.Request.Context(), email, name)
    if err != nil {
        // Log error for debugging
        logger.FromContext(c.Request.Context()).Error("Failed to create user",
            slog.Any("error", err),
        )
        
        response.InternalError(c, "INTERNAL_ERROR", "Failed to create user")
        return
    }
    
    response.Success(c, user)
}
```

**Response:**
```json
HTTP/1.1 500 Internal Server Error
{
  "status": "error",
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to create user"
  }
}
```

---

## API Reference

### Success Responses

```go
// Success sends a successful response with data (200 OK)
func Success[T any](c *gin.Context, data T)

// Created sends a resource creation success response (201 Created)
func Created[T any](c *gin.Context, data T)
```

### Error Responses

```go
// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, code, message string)

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, code, message string)

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, code, message string)

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, code, message string)

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, code, message string)
```

### Response Structs

```go
// Response represents a standard API response
type Response[T any] struct {
    Status  string      `json:"status"`           // "success" or "error"
    Data    T           `json:"data,omitempty"`   // Response data (success only)
    Error   *Error      `json:"error,omitempty"`  // Error details (error only)
    Message string      `json:"message,omitempty"`// Optional message
}

// Error represents an error response
type Error struct {
    Code    string `json:"code"`    // Error code (e.g., "USER_NOT_FOUND")
    Message string `json:"message"` // Human-readable error message
}
```

---

## Best Practices

### DO

- **Use specific error codes** - Makes debugging easier
  ```go
  // Good: Specific error codes
  response.NotFound(c, "USER_NOT_FOUND", "User not found")
  response.BadRequest(c, "INVALID_EMAIL", "Email format is invalid")
  
  // Bad: Generic codes
  response.NotFound(c, "ERROR", "Not found")
  ```

- **Return after error response** - Prevent further execution
  ```go
  if user == nil {
      response.NotFound(c, "USER_NOT_FOUND", "User not found")
      return  // Always return after error response
  }
  ```

- **Log errors before responding** - For debugging
  ```go
  if err != nil {
      logger.FromContext(ctx).Error("Database error", slog.Any("error", err))
      response.InternalError(c, "INTERNAL_ERROR", "Failed to fetch user")
      return
  }
  ```

- **Use proper status codes** - Match HTTP semantics
  ```go
  response.Success(c, user)         // 200 OK - retrieval
  response.Created(c, user)         // 201 Created - resource creation
  response.NotFound(c, ...)         // 404 - resource doesn't exist
  response.BadRequest(c, ...)       // 400 - invalid input
  response.Unauthorized(c, ...)     // 401 - authentication required
  response.Forbidden(c, ...)        // 403 - insufficient permissions
  response.InternalError(c, ...)    // 500 - server error
  ```

### DON'T

- **Don't expose internal errors** - Sanitize error messages
  ```go
  // Bad: Exposes database internals
  response.InternalError(c, "DB_ERROR", err.Error())
  
  // Good: Generic message
  response.InternalError(c, "INTERNAL_ERROR", "Failed to process request")
  ```

- **Don't use success status with error codes**
  ```go
  // Bad: Confusing response
  c.JSON(200, gin.H{"status": "success", "error": "Something failed"})
  
  // Good: Use proper error response
  response.InternalError(c, "INTERNAL_ERROR", "Something failed")
  ```

- **Don't forget to return after errors**
  ```go
  // Bad: Continues execution
  if err != nil {
      response.InternalError(c, "INTERNAL_ERROR", err.Error())
      // Missing return - code continues!
  }
  response.Success(c, user)  // Will execute even if error
  
  // Good: Returns immediately
  if err != nil {
      response.InternalError(c, "INTERNAL_ERROR", err.Error())
      return
  }
  ```

---

## Error Code Conventions

Use consistent error code naming:

### Validation Errors (400)
```go
"VALIDATION_ERROR"      // Generic validation failure
"INVALID_EMAIL"         // Email format invalid
"INVALID_PHONE"         // Phone format invalid
"MISSING_REQUIRED_FIELD" // Required field missing
```

### Authentication Errors (401)
```go
"UNAUTHORIZED"          // No auth token provided
"INVALID_TOKEN"         // Token is invalid/expired
"INVALID_CREDENTIALS"   // Login credentials incorrect
```

### Authorization Errors (403)
```go
"FORBIDDEN"             // Generic permission denied
"INSUFFICIENT_PERMISSIONS" // User lacks required permission
"RESOURCE_ACCESS_DENIED"   // Cannot access this resource
```

### Not Found Errors (404)
```go
"USER_NOT_FOUND"        // User doesn't exist
"CONTACT_NOT_FOUND"     // Contact doesn't exist
"RESOURCE_NOT_FOUND"    // Generic resource not found
```

### Server Errors (500)
```go
"INTERNAL_ERROR"        // Generic server error
"DATABASE_ERROR"        // Database operation failed
"SERVICE_UNAVAILABLE"   // External service unavailable
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
    "github.com/basilex/promenade/pkg/logger"
    "log/slog"
)

type UserHandler struct {
    usecase IUserUseCase
}

func NewUserHandler(uc IUserUseCase) *UserHandler {
    return &UserHandler{usecase: uc}
}

// Create godoc
// @Summary Create new user
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User data"
// @Success 201 {object} response.Response[User]
// @Failure 400 {object} response.Response[any]
// @Failure 500 {object} response.Response[any]
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    
    // Validate request
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "VALIDATION_ERROR", err.Error())
        return
    }
    
    // Create user
    ctx := c.Request.Context()
    user, err := h.usecase.CreateUser(ctx, req.Email, req.Name, req.Password)
    
    // Handle errors
    if err != nil {
        logger.FromContext(ctx).Error("Failed to create user",
            slog.Any("error", err),
            slog.String("email", req.Email),
        )
        response.InternalError(c, "INTERNAL_ERROR", "Failed to create user")
        return
    }
    
    // Success response
    response.Created(c, user)
}

// GetByID godoc
// @Summary Get user by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response[User]
// @Failure 404 {object} response.Response[any]
// @Failure 500 {object} response.Response[any]
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
    // Parse ID
    userID, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "INVALID_ID", "Invalid user ID format")
        return
    }
    
    // Get user
    ctx := c.Request.Context()
    user, err := h.usecase.GetUser(ctx, userID)
    
    // Handle not found
    if errors.Is(err, ErrUserNotFound) {
        response.NotFound(c, "USER_NOT_FOUND", "User not found")
        return
    }
    
    // Handle other errors
    if err != nil {
        logger.FromContext(ctx).Error("Failed to get user",
            slog.Any("error", err),
            slog.String("user_id", userID.String()),
        )
        response.InternalError(c, "INTERNAL_ERROR", "Failed to fetch user")
        return
    }
    
    // Success response
    response.Success(c, user)
}

// Update godoc
// @Summary Update user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "Update data"
// @Success 200 {object} response.Response[User]
// @Failure 400 {object} response.Response[any]
// @Failure 403 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Failure 500 {object} response.Response[any]
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
    // Parse ID
    userID, err := uuidv7.Parse(c.Param("id"))
    if err != nil {
        response.BadRequest(c, "INVALID_ID", "Invalid user ID format")
        return
    }
    
    // Check permissions (assuming middleware set this)
    requestingUserID := c.GetString("user_id")
    if requestingUserID != userID.String() {
        response.Forbidden(c, "FORBIDDEN", "You can only update your own profile")
        return
    }
    
    // Validate request
    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "VALIDATION_ERROR", err.Error())
        return
    }
    
    // Update user
    ctx := c.Request.Context()
    user, err := h.usecase.UpdateUser(ctx, userID, req.Name)
    
    // Handle not found
    if errors.Is(err, ErrUserNotFound) {
        response.NotFound(c, "USER_NOT_FOUND", "User not found")
        return
    }
    
    // Handle other errors
    if err != nil {
        logger.FromContext(ctx).Error("Failed to update user",
            slog.Any("error", err),
            slog.String("user_id", userID.String()),
        )
        response.InternalError(c, "INTERNAL_ERROR", "Failed to update user")
        return
    }
    
    // Success response
    response.Success(c, user)
}
```

---

## Testing

```go
package handler_test

import (
    "testing"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestUserHandler_GetByID_Success(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    // Test success response
    response.Success(c, map[string]string{"id": "123", "name": "John"})
    
    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), `"status":"success"`)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    // Test not found response
    response.NotFound(c, "USER_NOT_FOUND", "User not found")
    
    assert.Equal(t, 404, w.Code)
    assert.Contains(t, w.Body.String(), `"status":"error"`)
    assert.Contains(t, w.Body.String(), `"code":"USER_NOT_FOUND"`)
}
```

---

## Related Packages

- `pkg/logger` - Structured logging (use before error responses)
- `github.com/gin-gonic/gin` - HTTP framework integration
- `pkg/uuidv7` - UUID parsing for ID parameters

---

**Last Updated:** 2025-12-28  
**Status:** Production-ready  
**Maintainer:** Promenade Team
