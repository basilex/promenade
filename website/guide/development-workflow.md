# Development Workflow

Step-by-step guide for contributing to Promenade Platform with Git workflow, code standards, and testing requirements.

---

## Quick Start for Contributors

```bash
# 1. Fork and clone
git clone https://github.com/YOUR_USERNAME/promenade.git
cd promenade

# 2. Set up development environment
make dev              # Starts PostgreSQL, runs migrations, starts API

# 3. Create feature branch
git checkout -b feature/my-feature

# 4. Make changes and test
make test            # Run all tests
make lint            # Check code quality

# 5. Commit and push
git add .
git commit -m "feat(customer): add export to CSV feature"
git push origin feature/my-feature

# 6. Open Pull Request on GitHub
```

---

## Git Branching Strategy

### Main Branches

**`main`** - Production-ready code
- Always stable and deployable
- Protected branch (no direct commits)
- Releases tagged with semantic versioning (v0.1.0, v0.2.0)

**`dev`** - Development integration branch
- Latest features and fixes
- Base for all feature branches
- Deployed to staging environment

### Feature Branches

**Naming Convention**:
```bash
feature/<short-description>    # New features
bugfix/<issue-description>     # Bug fixes
hotfix/<critical-fix>          # Production hotfixes
refactor/<component-name>      # Code refactoring
docs/<topic>                   # Documentation only
test/<test-description>        # Test improvements
```

**Examples**:
```bash
git checkout -b feature/customer-export
git checkout -b bugfix/email-validation
git checkout -b docs/api-reference
git checkout -b test/integration-tests
```

### Branch Lifecycle

1. **Create**: Branch from `dev`
   ```bash
   git checkout dev
   git pull origin dev
   git checkout -b feature/my-feature
   ```

2. **Develop**: Make commits
   ```bash
   git add .
   git commit -m "feat: add customer export"
   ```

3. **Sync**: Keep up-to-date with dev
   ```bash
   git fetch origin
   git rebase origin/dev
   ```

4. **Push**: Push to your fork
   ```bash
   git push origin feature/my-feature
   ```

5. **PR**: Open Pull Request to `dev`
6. **Review**: Address feedback
7. **Merge**: Squash and merge into `dev`
8. **Delete**: Remove feature branch after merge

---

## Commit Message Standards

### Conventional Commits Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Commit Types

| Type       | Description                        | Example                                      |
|------------|----------------------------------- |----------------------------------------------|
| `feat`     | New feature                        | `feat(customer): add CSV export`             |
| `fix`      | Bug fix                            | `fix(auth): handle expired JWT tokens`       |
| `docs`     | Documentation changes              | `docs(api): update authentication guide`     |
| `style`    | Code formatting (no logic change)  | `style(user): format with gofmt`             |
| `refactor` | Code refactoring                   | `refactor(repo): use BaseRepository pattern` |
| `test`     | Add or update tests                | `test(contact): add integration tests`       |
| `chore`    | Maintenance tasks                  | `chore(deps): update go modules`             |
| `perf`     | Performance improvement            | `perf(query): optimize customer list query`  |

### Scope Examples

- `customer`, `order`, `deal`, `contact`, `profile`, `user`
- `auth`, `jwt`, `rbac`, `middleware`
- `bus`, `logger`, `cache`, `migration`
- `api`, `handler`, `usecase`, `repository`
- `docs`, `test`, `ci`, `docker`

### Good Commit Messages

```bash
# ✅ GOOD
feat(customer): add export to CSV feature
fix(auth): prevent JWT token reuse after logout
docs(api): add examples for deal endpoints
test(order): add integration tests for order state machine

# ❌ BAD
update files
fix bug
WIP
asdasd
```

### Multi-line Commits

```bash
git commit -m "feat(customer): add advanced search filters

- Add filter by status (active/inactive)
- Add filter by date range
- Add filter by sales rep
- Add full-text search on name and email

Closes #123"
```

---

## Code Standards

### Go Code Style

**Follow Effective Go**: https://go.dev/doc/effective_go

**Key Rules**:
1. **gofmt**: Always format code with `gofmt` (or `make fmt`)
2. **golangci-lint**: Pass all linter checks (`make lint`)
3. **Comments**: Add godoc comments for exported functions
4. **Error handling**: Always check and wrap errors
5. **Context**: First parameter in functions should be `ctx context.Context`
6. **Naming**: Use idiomatic Go names (interfaces start with `I`)

**Example**:
```go
// IRepository defines repository operations for Contact aggregate
type IRepository interface {
    // GetByID retrieves a contact by ID
    GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error)
    
    // Create persists a new contact
    Create(ctx context.Context, contact *Contact) error
}
```

### Project Structure Conventions

**Naming**:
- Interfaces: `IRepository`, `IUseCase` (capital I prefix)
- Implementations: lowercase `repository`, `useCase`
- Constructors: `NewUseCase()`, `NewContactRepository()`
- Handlers: `ContactHandler`, `UserHandler` (PascalCase)
- Files: `entity.go`, `usecase.go`, `repository.go` (lowercase)

**Test Files**:
- Unit tests: Same directory as code (`entity_test.go`)
- Smoke tests: `test/smoke/contexts/` (mirror path)
- Integration tests: `test/integration/contexts/` (mirror path)

**Imports**:
- Standard library first
- Third-party packages second
- Internal packages last
- Blank line between groups

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"

    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/pkg/valueobject"
)
```

---

## Testing Requirements

### Before Submitting PR

**All tests must pass**:
```bash
make test              # Run all tests (360+ tests, ~60s)
make test-unit         # Unit tests only (~5s)
make test-smoke        # Smoke tests (~0.4s)
make test-integration  # Integration tests (~14s)
make lint              # Linter checks
```

### Test Coverage Requirements

- **New features**: 80%+ test coverage
- **Bug fixes**: Add test reproducing the bug
- **Refactoring**: Existing tests must still pass

### Writing Tests

**Unit Tests** (in-place):
```go
// internal/contexts/identity/contact/entity_test.go
func TestContact_NewEmailContact(t *testing.T) {
    userID := uuidv7.New()
    contact, err := NewEmailContact(userID, "test@example.com", "Work")
    
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contact.ID)
    assert.Equal(t, ContactTypeEmail, contact.Type)
}
```

**Integration Tests** (with DB):
```go
// test/integration/contexts/identity/contact/repository_test.go
func TestContactRepository_Create(t *testing.T) {
    db := integration.SetupTestDBWithCleanTables(t)
    repo := postgres.NewContactRepository(db.DB)
    
    contact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
    err := repo.Create(context.Background(), contact)
    
    require.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contact.ID)
}
```

**See**: [Testing Patterns Guide](testing-patterns.md) for comprehensive testing documentation

---

## Pull Request Process

### 1. Prepare Your PR

**Checklist before opening PR**:
- [ ] All tests pass (`make test`)
- [ ] Code formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Meaningful commit messages
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG.md updated (for user-facing changes)

### 2. Open Pull Request

**PR Title Format**:
```
feat(customer): add CSV export feature
fix(auth): prevent token reuse after logout
docs(api): update deal management examples
```

**PR Description Template**:
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change fixing an issue)
- [ ] New feature (non-breaking change adding functionality)
- [ ] Breaking change (fix or feature breaking existing functionality)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Related Issues
Closes #123
```

### 3. Code Review

**Reviewer Responsibilities**:
- Check code quality and adherence to standards
- Verify test coverage
- Test functionality locally
- Provide constructive feedback
- Approve when ready

**Author Responsibilities**:
- Address all review comments
- Update PR based on feedback
- Re-request review after changes
- Keep PR scope focused (one feature/fix per PR)

### 4. Merge Strategy

**Squash and Merge**:
- Multiple commits → Single commit in `dev`
- Keeps history clean
- Commit message = PR title

**After Merge**:
```bash
# Update your local dev branch
git checkout dev
git pull origin dev

# Delete feature branch
git branch -d feature/my-feature
git push origin --delete feature/my-feature
```

---

## Development Environment

### Required Tools

```bash
# macOS
brew install go           # Go 1.24+
brew install postgresql   # PostgreSQL 16
brew install redis        # Redis 7
brew install make         # GNU Make

# VS Code Extensions (recommended)
- Go (golang.go)
- GitLens
- Docker
- PostgreSQL
```

### Make Commands

```bash
# Development
make dev               # Start full dev environment
make build             # Build binary
make run               # Build and run
make clean             # Clean build artifacts

# Code Quality
make fmt               # Format code
make lint              # Run linters

# Testing
make test              # All tests
make test-unit         # Unit tests only
make test-integration  # Integration tests
make test-coverage     # HTML coverage report

# Database
make migrate           # Run all migrations
make migrate-status    # Show migration status
make db-reset          # Drop and recreate database

# Docker
make docker-up         # Start PostgreSQL and Redis
make docker-down       # Stop containers
make docker-logs       # View logs
```

### Environment Variables

**Development** (.env.local):
```bash
export ENVIRONMENT=development
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=promenade_dev
export REDIS_ADDR=localhost:6379
export JWT_SECRET=your-dev-secret-at-least-32-chars
```

**Never commit**:
- `.env.local` (gitignored)
- Secrets or passwords
- API keys

---

## Adding a New Feature

### Step-by-Step Example: Add "Customer Notes" Feature

**1. Create Feature Branch**
```bash
git checkout dev
git pull origin dev
git checkout -b feature/customer-notes
```

**2. Plan the Feature**
- Where does it fit? → Customer Management context
- What's the aggregate? → Customer (add Notes field)
- What's the use case? → AddNote(customerID, note string)
- What's the API? → POST /api/v1/customers/:id/notes

**3. Update Entity**
```go
// internal/contexts/customer-mgmt/customer/entity.go
type Customer struct {
    // ... existing fields
    Notes []Note `json:"notes"`
}

type Note struct {
    ID        uuidv7.UUID `json:"id"`
    Content   string      `json:"content"`
    CreatedAt time.Time   `json:"created_at"`
    CreatedBy uuidv7.UUID `json:"created_by"`
}

func (c *Customer) AddNote(content string, createdBy uuidv7.UUID) error {
    if content == "" {
        return fmt.Errorf("note content cannot be empty")
    }
    
    note := Note{
        ID:        uuidv7.New(),
        Content:   content,
        CreatedAt: time.Now(),
        CreatedBy: createdBy,
    }
    
    c.Notes = append(c.Notes, note)
    c.UpdatedAt = time.Now()
    
    return nil
}
```

**4. Write Tests**
```go
// entity_test.go
func TestCustomer_AddNote(t *testing.T) {
    customer := NewCustomer("Test Corp", CustomerTypeLead)
    userID := uuidv7.New()
    
    err := customer.AddNote("Follow-up call scheduled", userID)
    
    assert.NoError(t, err)
    assert.Len(t, customer.Notes, 1)
    assert.Equal(t, "Follow-up call scheduled", customer.Notes[0].Content)
}
```

**5. Update Repository**
```go
// repository.go (interface)
type IRepository interface {
    // ... existing methods
    AddNote(ctx context.Context, customerID uuidv7.UUID, note Note) error
}

// postgres/customer_repository.go (implementation)
func (r *customerRepository) AddNote(ctx context.Context, customerID uuidv7.UUID, note Note) error {
    query := `UPDATE customer_customers 
              SET notes = notes || $1::jsonb, updated_at = NOW()
              WHERE id = $2 AND deleted_at IS NULL`
    
    noteJSON, _ := json.Marshal(note)
    _, err := r.Exec(ctx, query, noteJSON, customerID)
    return err
}
```

**6. Update Use Case**
```go
// usecase.go
func (uc *useCase) AddNote(ctx context.Context, customerID uuidv7.UUID, content string, createdBy uuidv7.UUID) error {
    customer, err := uc.repo.GetByID(ctx, customerID)
    if err != nil {
        return err
    }
    
    if err := customer.AddNote(content, createdBy); err != nil {
        return err
    }
    
    return uc.repo.Update(ctx, customer)
}
```

**7. Add HTTP Handler**
```go
// handler/customer_handler.go
func (h *CustomerHandler) AddNote(c *gin.Context) {
    customerID, _ := uuidv7.Parse(c.Param("id"))
    userID := jwt.GetUserID(c)
    
    var req struct {
        Content string `json:"content" binding:"required,min=1,max=500"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    
    err := h.usecase.AddNote(c.Request.Context(), customerID, req.Content, userID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "ADD_NOTE_FAILED", err.Error())
        return
    }
    
    response.Success(c, gin.H{"message": "Note added successfully"})
}
```

**8. Register Route**
```go
// router.go
customers.POST("/:id/notes", h.customerHandler.AddNote)
```

**9. Run Tests**
```bash
make test-unit         # Verify entity and use case tests
make test-integration  # Verify repository tests
make test              # Run all tests
```

**10. Test Manually**
```bash
# Start dev server
make dev

# Test endpoint
curl -X POST http://localhost:8081/api/v1/customers/:id/notes \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content":"Follow-up call scheduled for next week"}'
```

**11. Commit and Push**
```bash
git add .
git commit -m "feat(customer): add notes functionality

- Add Note entity with ID, content, timestamp
- Add AddNote method to Customer aggregate
- Implement repository method for note persistence
- Add HTTP endpoint POST /customers/:id/notes
- Add unit and integration tests

Closes #456"

git push origin feature/customer-notes
```

**12. Open Pull Request**
- Title: `feat(customer): add notes functionality`
- Description: Link to issue, describe changes, show test coverage
- Wait for code review

---

## Common Pitfalls

### ❌ Don't Do This

1. **Commit directly to `main` or `dev`**
   ```bash
   # BAD
   git checkout main
   git commit -m "quick fix"
   ```

2. **Create giant PRs**
   - Keep PRs focused (one feature/fix)
   - Break large features into multiple PRs

3. **Skip tests**
   ```bash
   # BAD
   git commit -m "feat: new feature (tests later)"
   ```

4. **Hardcode values**
   ```go
   // BAD
   db := connectToDB("localhost:5432")
   
   // GOOD
   db := connectToDB(cfg.Database.Host)
   ```

5. **Use `uuid.New()` instead of `uuidv7.New()`**
   ```go
   // BAD
   import "github.com/google/uuid"
   id := uuid.New()  // UUID v4 (random, slow inserts)
   
   // GOOD
   import "github.com/basilex/promenade/pkg/uuidv7"
   id := uuidv7.New()  // UUID v7 (time-ordered, 2x faster)
   ```

### ✅ Do This Instead

1. Always create feature branch
2. Write tests first (TDD)
3. Keep commits atomic and focused
4. Use configuration for environment-specific values
5. Follow project conventions (naming, structure)

---

## Getting Help

**Stuck?** Ask for help:
- [GitHub Discussions](https://github.com/basilex/promenade/discussions)
- [Issues](https://github.com/basilex/promenade/issues)
- Email: alexander.vasilenko@gmail.com

**Before asking**:
- Check [README.md](../../README.md)
- Search existing issues
- Review [documentation](../INDEX.md)

---

## Related Documentation

- [Contributing Guide](contributing.md) - How to contribute
- [Testing Patterns](testing-patterns.md) - Comprehensive testing guide
- [Testing Quick Reference](testing-quick-reference.md) - One-page cheat sheet
- [Clean Architecture](../concepts/clean-architecture.md) - DDD principles
- [API Reference](../reference/api-reference.md) - HTTP API documentation

---

**Happy coding!** 🚀
