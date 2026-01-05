# Testing Quick Reference

**One-page cheat sheet for Promenade testing patterns**

---

##  Three Test Types

| Type            | Location                     | Database           | Speed   | Purpose        |
| --------------- | ---------------------------- | ------------------ | ------- | -------------- |
| **Unit**        | In-place (`*_test.go`)       |  Mocks           |  5s   | Business logic |
| **Smoke**       | `test/smoke/contexts/`       |  Mock UseCase    |  0.4s | HTTP handlers  |
| **Integration** | `test/integration/contexts/` |  Real PostgreSQL |  6s   | Repositories   |

---

##  Integration Test Pattern

```go
package user_test

func TestUserRepository_Create(t *testing.T) {
    // 1. SetupTestDBWithCleanTables - TRUNCATE all tables
    db := integration.SetupTestDBWithCleanTables(t)
    repo := postgres.NewUserRepository(db.DB)
    ctx := context.Background()

    t.Run("create user successfully", func(t *testing.T) {
        // 2. Static email (DB is clean)
        u, _ := user.NewUser("test@example.com", "password")

        err := repo.Create(ctx, u)

        require.NoError(t, err)
        assert.NotEqual(t, uuidv7.Nil, u.ID)
    })
}
```

** DO**: SetupTestDBWithCleanTables, static emails, ctx first  
** DON'T**: SetupTestDB, UUID everywhere, WithTransaction

---

##  Smoke Test Pattern

```go
package user_test

// 1. Mock ALL IUseCase methods
type MockUserUseCase struct { mock.Mock }

func (m *MockUserUseCase) Register(ctx context.Context, email, name, password string) (*user.User, error) {
    args := m.Called(ctx, email, name, password)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*user.User), args.Error(1)
}

// 2. Router setup
func setupUserRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    return gin.New()
}

// 3. ONE test function
func TestUserHandler_Smoke(t *testing.T) {
    mockUC := new(MockUserUseCase)
    handler := userHTTP.NewUserHandler(mockUC)
    router := setupUserRouter()

    router.POST("/users/register", handler.Register)

    // 4. Subtest per endpoint
    t.Run("Register returns 201", func(t *testing.T) {
        u := &user.User{
            ID:           uuidv7.New(),
            PasswordHash: "hashedPassword",
            Status:       user.UserStatusActive,
        }

        mockUC.On("Register", mock.Anything, "test@example.com", "Name", "password").Return(u, nil).Once()

        reqBody := userHTTP.RegisterRequest{Email: "test@example.com", Name: "Name", Password: "password"}
        body, _ := json.Marshal(reqBody)
        req := httptest.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, http.StatusCreated, w.Code)
        mockUC.AssertExpectations(t)
    })
}
```

** DO**: Mock ALL usecase calls, create entities directly, AssertExpectations  
** DON'T**: Use real DB, call factory methods, forget secondary mocks

---

##  Unit Test Pattern

```go
package user

func TestUser_NewUser(t *testing.T) {
    tests := []struct {
        name     string
        email    string
        password string
        wantErr  bool
    }{
        {"valid", "test@example.com", "ValidPass123!", false},
        {"invalid email", "not-email", "ValidPass123!", true},
        {"weak password", "test@example.com", "123", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, err := NewUser(tt.email, tt.password)

            if tt.wantErr {
                require.Error(t, err)
                assert.Nil(t, user)
            } else {
                require.NoError(t, err)
                assert.NotNil(t, user)
            }
        })
    }
}
```

** DO**: Table-driven tests, require for critical, mock repositories  
** DON'T**: Test private methods, skip AssertExpectations, use random data

---

##  Running Tests

```bash
# All tests
make test                  # ~40s

# By type
make test-unit             # ~5s
make test-smoke            # ~0.4s
make test-integration      # ~6s

# By context
go test ./internal/contexts/identity/... -v
go test ./test/smoke/contexts/identity/... -v
go test ./test/integration/contexts/identity/... -v

# Coverage
make test-coverage
```

---

##  Common Mistakes

### Integration

```go
//  SetupTestDB - shared DB
db := integration.SetupTestDB(t)

//  SetupTestDBWithCleanTables - TRUNCATE
db := integration.SetupTestDBWithCleanTables(t)
```

### Smoke

```go
//  Missing secondary mock
mockUC.On("VerifyEmail", ...).Return(nil).Once()
// Handler ALSO calls GetUser!

//  Mock ALL calls
mockUC.On("VerifyEmail", ...).Return(nil).Once()
mockUC.On("GetUser", ...).Return(u, nil).Once()
```

### Unit

```go
//  No verification
mockRepo.On("Create", ...).Return(nil)

//  Verify called
mockRepo.On("Create", ...).Return(nil).Once()
mockRepo.AssertExpectations(t)
```

---

##  Full Documentation

- [Testing Patterns Guide](TESTING_PATTERNS.md) - Comprehensive 800+ line guide
- [Testing Structure](../test/TESTING_STRUCTURE.md) - Directory organization
- [Testing README](../test/README.md) - Overview

---

**Last Updated**: 2025-12-28  
**Quick Reference** - See [TESTING_PATTERNS.md](TESTING_PATTERNS.md) for complete details
