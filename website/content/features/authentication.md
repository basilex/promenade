---
title: "Authentication & RBAC"
description: "JWT authentication with role-based access control"
weight: 3
---

## Authentication & Authorization

Production-ready authentication with **JWT tokens** and **RBAC** (Role-Based Access Control).

### JWT Authentication

```go
// Login returns access + refresh tokens
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "secure123"
}

// Response
{
  "access_token": "eyJhbGc...",   // 15 minutes
  "refresh_token": "eyJhbGc...",  // 7 days
  "user": { ... }
}
```

### RBAC System

**4 Built-in Roles:**

- **Admin** - Full system access (`*` permission)
- **Moderator** - Content moderation
- **User** - Basic operations
- **Guest** - Read-only access

**Permission Format:** `resource:action`

```
posts:create
posts:update
posts:delete
users:manage
*  # Wildcard - full access
```

### Middleware Protection

```go
// Require authentication
router.Use(authMiddleware.RequireAuth())

// Require specific permission
router.POST("/posts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Session Management

- Multiple sessions per user
- Session revocation
- Device tracking
- Activity monitoring

### Default Test Users

```
admin@promenade.com     | passw0rd | Admin
moderator@promenade.com | passw0rd | Moderator
user@promenade.com      | passw0rd | User
```

### Benefits

✅ **Secure** - JWT with HMAC-SHA256 signing  
✅ **Flexible** - Wildcard and granular permissions  
✅ **Scalable** - Stateless tokens, session tracking optional  
✅ **Production-ready** - Rate limiting, session revocation

[Authorization Guide →](/docs/AUTHORIZATION)
