# Contributing Guide

**Development workflow and contribution guidelines for Promenade Platform**

---

## Getting Started

### Prerequisites

- **Go 1.23+** installed
- **Docker & Docker Compose** (for PostgreSQL)
- **Make** (comes with macOS/Linux)
- **Git** configured
- **Code Editor**: VS Code, GoLand, or Vim/Neovim

### Fork and Clone

```bash
# Fork repository on GitHub (click "Fork" button)

# Clone your fork
git clone https://github.com/YOUR_USERNAME/promenade.git
cd promenade

# Add upstream remote
git remote add upstream https://github.com/basilex/promenade.git

# Verify remotes
git remote -v
```

### Setup Development Environment

```bash
# Start PostgreSQL
make docker-up

# Run migrations
make migrate

# Run tests to verify setup
make test

# Start development server
make dev
```

✅ **Success**: Server running on http://localhost:8081

---

## Development Workflow

### Branch Strategy

- `dev` - Development branch (default, **target for PRs**)
- `main` - Production-ready code (protected)
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates

### Creating a Feature

```bash
# 1. Sync with upstream
git checkout dev
git pull upstream dev

# 2. Create feature branch
git checkout -b feature/my-awesome-feature

# 3. Make changes and commit
git add .
git commit -m "feat: add awesome feature"

# 4. Push to your fork
git push origin feature/my-awesome-feature

# 5. Open Pull Request on GitHub (to basilex/promenade:dev)
```

---

## Project Structure

```
promenade/
├── cmd/
│   ├── api/              # HTTP server entry point
│   └── migrate/          # Migration CLI
├── internal/
│   ├── contexts/         # Bounded Contexts (DDD)
│   │   ├── shared/      # Reference data
│   │   ├── identity/    # User management
│   │   └── customer-mgmt/ # CRM
│   └── infrastructure/   # Config, DB, etc.
├── pkg/                  # Shared packages
│   ├── bus/             # Event Bus
│   ├── jwt/             # JWT auth
│   ├── logger/          # Logging
│   └── uuidv7/          # UUID v7
├── migrations/           # Database migrations
├── test/                 # Three-tier tests
│   ├── smoke/           # Mock-based tests
│   └── integration/     # DB tests
└── docs/                 # Documentation
```

---

## Coding Standards

### Go Code Style

1. **Follow `gofmt`**: Code must be formatted with `go fmt`

```bash
# Format all code
make fmt

# Check formatting
go fmt ./...
```

2. **Follow `golangci-lint`**: Pass all linter checks

```bash
# Run linters
make lint

# Auto-fix some issues
golangci-lint run --fix
```

3. **Naming Conventions**:

```go
// ✅ Interfaces: I prefix + PascalCase
type IUserRepository interface {}
type ICustomerUseCase interface {}

// ✅ Implementations: lowercase private
type userRepository struct {}
type useCase struct {}

// ✅ Constructors: New{Entity} or NewUseCase
func NewUserRepository(db *sqlx.DB) IUserRepository
func NewUseCase(repo IRepository) IUseCase

// ✅ Methods: PascalCase
func (r *userRepository) Create(ctx context.Context, user *User) error

// ✅ Constants: SCREAMING_SNAKE_CASE or PascalCase
const DefaultPageSize = 20
const UserStatusActive = "active"

// ✅ Variables: camelCase
var userCount int
var isActive bool
```

4. **Error Handling**:

```go
// ✅ Always handle errors
if err := repo.Create(ctx, user); err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ✅ Use error wrapping
return fmt.Errorf("validation failed: %w", err)

// ✅ Define domain errors
var ErrUserNotFound = errors.New("user not found")
var ErrEmailAlreadyExists = errors.New("email already exists")

// ✅ Check specific errors
if errors.Is(err, ErrUserNotFound) {
    response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
    return
}
```

5. **Context Propagation**:

```go
// ✅ Always pass context as first parameter
func (uc *useCase) CreateUser(ctx context.Context, email, password string) (*User, error)

// ✅ Use context for logging
logger.FromContext(ctx).Info("user created", slog.String("user_id", userID.String()))

// ✅ Pass context to all repository calls
err := uc.repo.Create(ctx, user)
```

---

## Testing Requirements

### Three-Tier Testing Strategy

All code must have appropriate test coverage:

| Test Type       | Location                  | Required Coverage |
| --------------- | ------------------------- | ----------------- |
| **Unit**        | Same directory as code    | 85%+              |
| **Smoke**       | `test/smoke/contexts/`    | All handlers      |
| **Integration** | `test/integration/contexts/` | All repositories |

### Writing Tests

**Unit Test** (entity logic):
```go
// internal/contexts/identity/user/entity_test.go
func TestUser_NewUser_ValidEmail(t *testing.T) {
    user, err := NewUser("test@example.com", "password123")
    
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.NotEqual(t, uuidv7.Nil, user.ID)
}
```

**Smoke Test** (handler with mock):
```go
// test/smoke/contexts/identity/user/handler_test.go
func TestUserHandler_Register_Returns201(t *testing.T) {
    gin.SetMode(gin.TestMode)
    mockUC := new(MockUserUseCase)
    handler := NewUserHandler(mockUC)
    
    mockUC.On("Register", mock.Anything, "test@example.com", "Test", "pass123").
        Return(&user.User{ID: uuidv7.New()}, nil).Once()
    
    // Test HTTP request...
    assert.Equal(t, http.StatusCreated, w.Code)
    mockUC.AssertExpectations(t)
}
```

**Integration Test** (repository with real DB):
```go
// test/integration/contexts/identity/user/repository_test.go
func TestUserRepository_Create(t *testing.T) {
    db := integration.SetupTestDBWithCleanTables(t)
    repo := postgres.NewUserRepository(db.DB)
    ctx := context.Background()
    
    u, _ := user.NewUser("test@example.com", "password123")
    err := repo.Create(ctx, u)
    
    require.NoError(t, err)
    assert.NotEqual(t, uuidv7.Nil, u.ID)
}
```

### Running Tests

```bash
# All tests
make test

# By type
make test-unit
make test-smoke
make test-integration

# Specific context
go test ./internal/contexts/identity/... -v

# With coverage
make test-coverage
```

**Before submitting PR**: All tests must pass!

---

## Commit Message Format

We follow **Conventional Commits** specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring (no functional changes)
- `test`: Adding or updating tests
- `chore`: Build process, dependencies, tooling
- `perf`: Performance improvements

### Examples

**Feature**:
```
feat(identity): add email verification for users

- Implement email verification endpoint
- Add verification token to User entity
- Send verification email via event bus

Closes #123
```

**Bug Fix**:
```
fix(customer): resolve duplicate email validation

- Add unique constraint on email per company
- Update validation logic in Customer entity

Fixes #456
```

**Documentation**:
```
docs(rbac): add comprehensive RBAC guide

- Document roles and permissions
- Add API endpoint examples
- Include JWT integration guide
```

**Refactoring**:
```
refactor(contact): extract value object creation logic

- Move Email, Phone, Address creation to factory methods
- Simplify Contact entity constructor
```

---

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:
```bash
git checkout dev
git pull upstream dev
git checkout your-branch
git rebase dev
```

2. **Run all checks**:
```bash
make fmt      # Format code
make lint     # Pass linters
make test     # All tests pass
make build    # Builds successfully
```

3. **Update documentation** (if applicable):
   - Update README.md if API changes
   - Add/update context README if new aggregate
   - Update docs/ if architecture changes

### Submitting PR

1. **Push to your fork**:
```bash
git push origin your-branch
```

2. **Open Pull Request** on GitHub:
   - **Base**: `basilex/promenade:dev`
   - **Head**: `your-fork:your-branch`

3. **Fill PR template**:

```markdown
## Description
Brief description of changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Smoke tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass locally

## Checklist
- [ ] Code follows project style guide
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings/errors
```

### Code Review Process

1. **Automated Checks**: CI/CD runs tests, linters, build
2. **Maintainer Review**: Code review by project maintainers
3. **Address Feedback**: Make requested changes
4. **Approval**: At least 1 approval required
5. **Merge**: Maintainer merges to `dev`

---

## Adding a New Aggregate

**Step-by-step guide** to add new aggregate in existing context:

### 1. Create Directory Structure

```bash
mkdir -p internal/contexts/identity/newaggregate/adapter/http/handler/dto
mkdir -p internal/contexts/identity/newaggregate/adapter/repository/postgres
```

### 2. Create Entity

```go
// internal/contexts/identity/newaggregate/entity.go
package newaggregate

import (
    "github.com/basilex/promenade/pkg/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type NewAggregate struct {
    aggregate.BaseAggregate
    
    ID   uuidv7.UUID
    Name string
    // ... other fields
}

// Factory method
func NewNewAggregate(name string) (*NewAggregate, error) {
    if name == "" {
        return nil, fmt.Errorf("name is required")
    }
    
    return &NewAggregate{
        BaseAggregate: aggregate.NewBase(),
        ID:            uuidv7.New(),
        Name:          name,
    }, nil
}
```

### 3. Create Repository Interface

```go
// internal/contexts/identity/newaggregate/repository.go
package newaggregate

type IRepository interface {
    Create(ctx context.Context, agg *NewAggregate) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*NewAggregate, error)
    Update(ctx context.Context, agg *NewAggregate) error
    Delete(ctx context.Context, id uuidv7.UUID) error
}
```

### 4. Create UseCase

```go
// internal/contexts/identity/newaggregate/usecase.go
package newaggregate

type IUseCase interface {
    CreateNewAggregate(ctx context.Context, name string) (*NewAggregate, error)
    GetNewAggregate(ctx context.Context, id uuidv7.UUID) (*NewAggregate, error)
}

type useCase struct {
    repo IRepository
}

func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}

func (uc *useCase) CreateNewAggregate(ctx context.Context, name string) (*NewAggregate, error) {
    agg, err := NewNewAggregate(name)
    if err != nil {
        return nil, err
    }
    
    if err := uc.repo.Create(ctx, agg); err != nil {
        return nil, fmt.Errorf("failed to create aggregate: %w", err)
    }
    
    return agg, nil
}
```

### 5. Create Migration

```bash
make migrate-new CONTEXT=identity NAME=add_newaggregate_table
```

Edit generated files in `migrations/identity/`

### 6. Write Tests

- Unit tests: `entity_test.go`, `usecase_test.go`
- Smoke tests: `test/smoke/contexts/identity/newaggregate/handler_test.go`
- Integration tests: `test/integration/contexts/identity/newaggregate/repository_test.go`

### 7. Register in Router

```go
// internal/contexts/identity/router.go
func NewRouter(db *sqlx.DB) *Router {
    // ... existing aggregates
    
    // New aggregate
    newAggRepo := newaggregateRepo.NewNewAggregateRepository(db)
    newAggUC := newaggregate.NewUseCase(newAggRepo)
    newAggHandler := newaggregateHTTP.NewNewAggregateHandler(newAggUC)
    
    return &Router{
        // ...
        newAggregateHandler: newAggHandler,
    }
}

func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
    // ...
    
    newaggregates := identity.Group("/newaggregates")
    {
        newaggregates.POST("", r.newAggregateHandler.Create)
        newaggregates.GET("/:id", r.newAggregateHandler.GetByID)
    }
}
```

---

## Documentation

### Updating Documentation

**When to update docs**:
- New API endpoint → Update API Reference
- New aggregate → Update context README
- Architecture change → Update docs/
- New package → Add package README

**Documentation locations**:
- `README.md` - Project overview
- `docs/` - Architecture guides
- `internal/contexts/*/README.md` - Context documentation
- `pkg/*/README.md` - Package documentation
- `website/` - VitePress documentation site

### Building Documentation Site

```bash
# Install dependencies
cd website
npm install

# Development server
npm run docs:dev

# Build static site
npm run docs:build
```

---

## Common Issues

### Tests Failing

**Problem**: Integration tests fail with duplicate key errors

**Solution**: Use `SetupTestDBWithCleanTables(t)` instead of `SetupTestDB(t)`

```go
// ❌ Wrong
db := integration.SetupTestDB(t)

// ✅ Correct
db := integration.SetupTestDBWithCleanTables(t)
```

### Linter Errors

**Problem**: `golangci-lint` reports `errcheck` errors

**Solution**: Handle all error returns

```go
// ❌ Wrong
db.Close()

// ✅ Correct
defer func() {
    if err := db.Close(); err != nil {
        logger.Error("failed to close db", slog.Any("error", err))
    }
}()
```

### Migration Issues

**Problem**: Migration fails with "relation already exists"

**Solution**: Reset database

```bash
make db-reset
make migrate
```

---

## Getting Help

### Resources

- [Documentation](https://basilex.github.io/promenade/) - Complete guides
- [Architecture Guide](/guide/architecture) - DDD concepts
- [Testing Guide](/guide/testing) - Testing patterns
- [API Reference](/guide/api-reference) - Endpoint documentation

### Communication

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: Questions, ideas
- **Pull Requests**: Code contributions
- **Email**: alexander.vasilenko@gmail.com

### Code of Conduct

Be respectful, professional, and constructive in all interactions.

---

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to Promenade!** 🚀
