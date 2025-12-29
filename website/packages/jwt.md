# JWT Package

**JSON Web Token (JWT) authentication** for Promenade Platform.

## Overview

Token-based authentication with RBAC support.

## Features

- **Token Generation** - Access + Refresh token pairs
- **Token Validation** - Parse and validate JWT tokens
- **RBAC Support** - Role-based access control via claims
- **Gin Middleware** - Authentication and authorization middleware
- **Token Revocation** - Redis-backed token blacklist

## Quick Start

```go
// Generate token pair
tokenPair, err := jwtManager.GenerateTokenPair(userID, email, roles)

// Validate token
claims, err := jwtManager.ValidateAccessToken(token)

// Protect routes
router.Use(jwt.AuthMiddleware(jwtManager))
router.Use(jwt.RequireRole("admin"))
```

## Configuration

```yaml
jwt:
  secret: "your-secret-key-at-least-32-characters"
  access_token_duration: 15m
  refresh_token_duration: 168h  # 7 days
  issuer: "promenade-platform"
```

## Token Flow

1. User logs in → Generate token pair
2. Client stores access token → Use for API requests
3. Access token expires → Refresh with refresh token
4. User logs out → Revoke tokens (Redis blacklist)

## Test Coverage

18 tests, 87% coverage

## Next Steps

- [Event Bus](/packages/bus) - Central communication hub
- [Identity Context](/contexts/identity) - JWT usage examples
- [GitHub Docs](https://github.com/basilex/promenade/blob/dev/pkg/jwt/README.md) - Complete documentation
