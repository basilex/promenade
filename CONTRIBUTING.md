# Contributing to Promenade CRM

Thank you for your interest in contributing to Promenade! This document provides guidelines and instructions for contributing.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Message Format](#commit-message-format)
- [Pull Request Process](#pull-request-process)

---

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to alexander.vasilenko@gmail.com.

---

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 16 (via Docker)
- Make

### Setup

```bash
# Clone the repository
git clone https://github.com/basilex/promenade.git
cd promenade

# Start PostgreSQL
make docker-up

# Run migrations
make migrate

# Build and run
make build
./bin/promenade

# Or use dev mode (Docker + migrations + API)
make dev
```

### Project Structure

```
promenade/
├── cmd/api/              # Application entry point
├── internal/contexts/    # Bounded Contexts (DDD)
│   ├── shared/          # Reference data (Country, Currency, etc.)
│   ├── identity/        # User, Contact, Profile aggregates
│   └── ...              # Other contexts
├── pkg/                 # Shared packages (bus, logger, uuidv7, etc.)
├── migrations/          # Database migrations (namespace-based)
├── test/                # Three-tier testing (unit, smoke, integration)
└── docs/                # Documentation
```

---

## Development Workflow

### Branch Strategy

- `main` - Production-ready code
- `dev` - Development branch (default)
- `feature/*` - Feature branches
- `fix/*` - Bug fix branches

### Creating a Feature

```bash
# Create feature branch from dev
git checkout dev
git pull origin dev
git checkout -b feature/my-feature

# Make changes, commit, push
git add .
git commit -m "feat: add new feature"
git push origin feature/my-feature

# Create Pull Request to dev
```

---

## Coding Standards

### Go Style Guide

Follow [Effective Go](https://go.dev/doc/effective_go) and [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

### Naming Conventions

**Interfaces**: PascalCase with `I` prefix

```go
type IRepository interface { }
type IUseCase interface { }
```

**Implementations**: lowercase private structs

```go
type contactRepository struct { *BaseRepository }
type UseCase struct { repo IRepository }
```

**Constructors**: `New{Entity}...` returns interface

```go
func NewContactRepository(db *sqlx.DB) IRepository
func NewUseCase(repo IRepository) IUseCase
```

**Errors**: `Err{Entity}{Condition}`

```go
var ErrContactNotFound = errors.New("contact not found")
var ErrEmailAlreadyExists = errors.New("email already exists")
```

### Domain-Driven Design Patterns

1. **Aggregates** - Business entities with invariants
2. **Value Objects** - Immutable domain concepts (Email, Phone, Money)
3. **Repositories** - Data access abstraction
4. **Use Cases** - Business logic orchestration
5. **Domain Events** - Asynchronous communication via Event Bus

### Architecture Rules

- Contexts communicate ONLY via Event Bus (no direct imports)
- Use `uuidv7.New()` for all IDs (never `uuid.New()`)
- Always pass `context.Context` as first parameter
- Soft delete: `WHERE deleted_at IS NULL` in all queries
- Repository methods: `GetByID`, `GetByXxx`, `Create`, `Update`, `Delete`

---

## Testing Guidelines

### Three-Tier Testing Strategy

**1. Unit Tests** (in-place)

- Test individual components in isolation
- Location: Same directory as production code (`*_test.go`)
- Run: `make test-unit` (~5s)

**2. Smoke Tests** (mock-based)

- Test HTTP handlers with mocked dependencies
- Location: `test/smoke/contexts/{context}/{aggregate}/`
- Run: `make test-smoke` (~0.4s)

**3. Integration Tests** (real DB)

- Test repositories with real PostgreSQL
- Location: `test/integration/contexts/{context}/{aggregate}/`
- Run: `make test-integration` (~6s)

### Writing Tests

```go
// Unit test example
func TestContact_NewEmailContact(t *testing.T) {
    userID := uuidv7.New()
    contact, err := NewEmailContact(userID, "test@example.com", "Work")

    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contact.ID)
    assert.Equal(t, userID, contact.UserID)
}

// Table-driven tests
func TestContact_Validate(t *testing.T) {
    tests := []struct {
        name    string
        contact *Contact
        wantErr bool
    }{
        {"valid email", validContact, false},
        {"empty email", emptyContact, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.contact.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Test Coverage

- Maintain 80%+ coverage for all new code
- Use `make test-coverage` to generate HTML report
- Focus on business logic (use cases, entities)

---

## Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (build, dependencies, etc.)

### Examples

```bash
feat(identity): add Profile aggregate with personal info

- Add Profile entity with 24 fields
- Implement ProfileUseCase with 13 update methods
- Add HTTP handler with 15 endpoints
- Add 87 unit tests (83.5% coverage)
- Add 8 smoke tests and 17 integration subtests

Closes #42

---

fix(contact): validate primary contact before deletion

Previously allowed deleting primary contacts, causing
inconsistent state. Now checks IsPrimary flag and returns
ErrPrimaryContactCannotBeDeleted.

Fixes #38

---

docs: update README with Profile aggregate examples

- Add Profile aggregate code example
- Update test statistics (240+ tests)
- Update roadmap (Phase 2 complete)
```

---

## Pull Request Process

### Before Submitting

1. **Run tests**: `make test` (all tests must pass)
2. **Format code**: `make fmt`
3. **Run linter**: `make lint`
4. **Update docs**: Update README.md if adding features
5. **Add tests**: Maintain 80%+ coverage

### PR Checklist

- [ ] Tests pass (`make test`)
- [ ] Code formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts with `dev`

### PR Template

```markdown
## Description

Brief description of changes

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing

- [ ] Unit tests added/updated
- [ ] Smoke tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass

## Checklist

- [ ] Code follows project conventions
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### Review Process

1. Create PR to `dev` branch
2. Automated checks run (tests, linter)
3. Code review by maintainers
4. Address feedback
5. Merge when approved

---

## Architecture Guidelines

### Adding a New Aggregate

1. Create directory: `internal/contexts/{context}/{aggregate}/`
2. Create files:
   - `entity.go` - Aggregate root
   - `entity_test.go` - Entity tests
   - `repository.go` - Repository interface
   - `usecase.go` - Use case implementation
   - `usecase_test.go` - Use case tests
3. Create adapter layer:
   - `adapter/http/handler/{aggregate}_handler.go`
   - `adapter/http/dto/{aggregate}_dto.go`
   - `adapter/repository/postgres/{aggregate}_repository.go`
4. Create migration: `make migrate-new CONTEXT=identity NAME=add_{aggregate}_table`
5. Register routes in context router

### Domain Events

Publish events after successful operations:

```go
// In use case
if err := uc.repo.Create(ctx, contact); err != nil {
    return nil, err
}

// Publish event
event := bus.NewBaseEvent("contact.created", contact.ID)
_ = uc.bus.Publish(ctx, bus.TopicContactCreated, event)
```

Subscribe in handlers:

```go
bus.Subscribe(bus.TopicContactCreated, func(ctx context.Context, e bus.Event) error {
    // Handle event
    return nil
})
```

---

## Resources

### Documentation

- [Main README](README.md) - Project overview
- [Documentation Index](docs/INDEX.md) - All documentation catalog
- [Clean Architecture Summary](docs/CLEAN_ARCHITECTURE_SUMMARY.md) - DDD with Bounded Contexts
- [Testing Patterns](docs/TESTING_PATTERNS.md) - Comprehensive testing guide
- [Testing Guide](test/README.md) - Testing structure
- [Event Bus](pkg/bus/README.md) - Central communication hub
- [AI Instructions](.github/copilot-instructions.md) - AI coding agents guide

### External Resources

- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Event-Driven Architecture](https://martinfowler.com/articles/201701-event-driven.html)
- [Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

---

## Questions?

- Open an issue: https://github.com/basilex/promenade/issues
- Email: alexander.vasilenko@gmail.com

---

**Thank you for contributing to Promenade!**
