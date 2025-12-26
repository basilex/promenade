---
title: "Clean Architecture"
description: "Strict layered architecture with clear separation of concerns"
weight: 1
---

## Clean Architecture

Promenade follows **Uncle Bob's Clean Architecture** principles with strict dependency rules and clear layer boundaries.

### Four Layers

```
Domain (entities, interfaces)
    ↓
Use Case (business logic)
    ↓
Adapter (repos, handlers)
    ↓
Infrastructure (DB, HTTP, events)
```

### Dependency Rule

**Inner layers never depend on outer layers**. This ensures:

- Business logic is framework-agnostic
- Easy testing with mock implementations
- Flexible infrastructure changes
- Clear boundaries between components

### Example Structure

```go
// Domain layer - pure business entities
type User struct {
    ID        uuid.UUID
    Email     string
    Status    UserStatus
}

// Use case - business logic only
type UserUseCase interface {
    RegisterUser(ctx, email, password) (*User, error)
}

// Adapter - framework integration
type UserHandler struct {
    usecase UserUseCase
}

func (h *UserHandler) Register(c *gin.Context) {
    // HTTP-specific code here
}
```

### Benefits

✅ **Testable** - Mock any layer independently  
✅ **Maintainable** - Changes isolated to single layer  
✅ **Scalable** - Add features without breaking existing code  
✅ **Framework-independent** - Business logic doesn't know about HTTP or DB

[Learn more →](/promenade/docs/ARCHITECTURE_OVERVIEW)
