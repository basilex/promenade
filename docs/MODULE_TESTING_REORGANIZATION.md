# План реорганизации тестов и моков модулей

## Текущее состояние

### Проблема

Нарушается принцип независимости модулей:

- Моки для модулей posts/profiles находятся в `test/mocks/` (централизовано)
- Use case тесты для модулей находятся в `internal/usecase/` (core)
- Модули зависят от core для тестирования
- Импорты: `github.com/basilex/promenade/internal/domain/entity` (core entity вместо module entity)

### Текущая структура

```
test/mocks/
├── user_post_repository_mock.go        # Модуль posts
├── post_comment_repository_mock.go     # Модуль posts
├── user_profile_repository_mock.go     # Модуль profiles
├── user_contact_repository_mock.go     # Модуль profiles
├── user_repository_mock.go             # Core
├── session_repository_mock.go          # Core
├── role_repository_mock.go             # Core
└── permission_repository_mock.go       # Core

internal/usecase/
├── user_post_usecase_test.go           # Модуль posts (но в core!)
├── post_comment_usecase_test.go        # Модуль posts
├── user_profile_usecase_test.go        # Модуль profiles
├── user_contact_usecase_test.go        # Модуль profiles
├── auth_usecase_test.go                # Core
├── role_usecase_test.go                # Core
└── permission_usecase_test.go          # Core

internal/modules/posts/domain/entity/
├── post_test.go                        # Есть (domain tests)
└── comment_test.go                     # Есть

internal/modules/profiles/entity/
├── user_profile_test.go                # Есть
└── user_contact_test.go                # Есть
```

---

## Целевая структура

### Core (internal/)

```
internal/domain/repository/mocks/       # Моки для core репозиториев
├── user_repository_mock.go
├── session_repository_mock.go
├── role_repository_mock.go
├── permission_repository_mock.go
├── country_repository_mock.go
└── currency_repository_mock.go

internal/usecase/                       # Use case тесты для core
├── auth_usecase_test.go
├── role_usecase_test.go
├── permission_usecase_test.go
├── country_usecase_test.go
└── currency_usecase_test.go
```

### Module: Posts

```
internal/modules/posts/
├── domain/
│   ├── entity/
│   │   ├── post.go
│   │   ├── post_test.go                 Есть
│   │   ├── comment.go
│   │   └── comment_test.go              Есть
│   └── repository/
│       ├── post_repository.go           # Interface
│       ├── comment_repository.go        # Interface
│       └── mocks/                       # 🆕 Создать
│           ├── post_repository_mock.go
│           └── comment_repository_mock.go
│
├── usecase/
│   ├── post_usecase.go
│   ├── post_usecase_test.go            # 🔄 Переместить из internal/usecase
│   ├── comment_usecase.go
│   └── comment_usecase_test.go         # 🔄 Переместить из internal/usecase
│
├── adapter/
│   ├── http/
│   │   ├── handler/
│   │   │   ├── post_handler.go
│   │   │   ├── post_handler_test.go    # 🆕 Создать
│   │   │   ├── comment_handler.go
│   │   │   └── comment_handler_test.go # 🆕 Создать
│   │   └── dto/
│   │       └── dto_test.go             # 🆕 Создать (опционально)
│   └── repository/postgres/
│       └── post_repository_test.go       Есть (integration)
│
└── tests/                               # 🆕 Создать (опционально)
    ├── integration/                     # End-to-end module tests
    └── fixtures/                        # Test data helpers
```

### Module: Profiles

```
internal/modules/profiles/
├── domain/
│   ├── entity/                          # Новая структура (было просто entity/)
│   │   ├── user_profile.go
│   │   ├── user_profile_test.go         Есть
│   │   ├── user_contact.go
│   │   └── user_contact_test.go         Есть
│   └── repository/                      # 🆕 Создать
│       ├── profile_repository.go        # Interface (из repository/)
│       ├── contact_repository.go        # Interface
│       └── mocks/                       # 🆕 Создать
│           ├── profile_repository_mock.go
│           └── contact_repository_mock.go
│
├── usecase/
│   ├── user_profile_usecase.go
│   ├── user_profile_usecase_test.go    # 🔄 Переместить
│   ├── user_contact_usecase.go
│   └── user_contact_usecase_test.go    # 🔄 Переместить
│
├── adapter/
│   ├── http/
│   │   ├── handler/
│   │   │   ├── user_profile_handler.go
│   │   │   ├── user_profile_handler_test.go  # 🆕 Создать
│   │   │   ├── user_contact_handler.go
│   │   │   └── user_contact_handler_test.go  # 🆕 Создать
│   │   └── dto/
│   └── repository/postgres/
│       └── profile_repository_test.go    Есть (integration)
│
└── tests/                               # 🆕 Создать
    └── integration/
```

---

## План миграции

### Фаза 1: Создание структуры моков в модулях

#### Posts module

```bash
mkdir -p internal/modules/posts/domain/repository/mocks

# Переместить и адаптировать моки
mv test/mocks/user_post_repository_mock.go \
   internal/modules/posts/domain/repository/mocks/post_repository_mock.go

mv test/mocks/post_comment_repository_mock.go \
   internal/modules/posts/domain/repository/mocks/comment_repository_mock.go
```

**Изменения в моках:**

```go
// Было
package mocks
import "github.com/basilex/promenade/internal/domain/entity"
import "github.com/basilex/promenade/internal/domain/repository"

// Стало
package mocks
import "github.com/basilex/promenade/internal/modules/posts/domain/entity"
import "github.com/basilex/promenade/internal/modules/posts/domain/repository"
```

#### Profiles module

```bash
mkdir -p internal/modules/profiles/domain/repository/mocks

mv test/mocks/user_profile_repository_mock.go \
   internal/modules/profiles/domain/repository/mocks/profile_repository_mock.go

mv test/mocks/user_contact_repository_mock.go \
   internal/modules/profiles/domain/repository/mocks/contact_repository_mock.go
```

**Изменения в моках:** аналогично posts

### Фаза 2: Перемещение use case тестов

#### Posts module

```bash
# Переместить usecase тесты
mv internal/usecase/user_post_usecase_test.go \
   internal/modules/posts/usecase/post_usecase_test.go

mv internal/usecase/post_comment_usecase_test.go \
   internal/modules/posts/usecase/comment_usecase_test.go
```

**Обновить импорты:**

```go
// Было
import "github.com/basilex/promenade/test/mocks"
import "github.com/basilex/promenade/internal/domain/entity"

// Стало
import "github.com/basilex/promenade/internal/modules/posts/domain/repository/mocks"
import "github.com/basilex/promenade/internal/modules/posts/domain/entity"
```

#### Profiles module

```bash
mv internal/usecase/user_profile_usecase_test.go \
   internal/modules/profiles/usecase/profile_usecase_test.go

mv internal/usecase/user_contact_usecase_test.go \
   internal/modules/profiles/usecase/contact_usecase_test.go
```

### Фаза 3: Реорганизация core моков

```bash
mkdir -p internal/domain/repository/mocks

# Переместить core моки
mv test/mocks/user_repository_mock.go internal/domain/repository/mocks/
mv test/mocks/session_repository_mock.go internal/domain/repository/mocks/
mv test/mocks/role_repository_mock.go internal/domain/repository/mocks/
mv test/mocks/permission_repository_mock.go internal/domain/repository/mocks/
mv test/mocks/country_repository_mock.go internal/domain/repository/mocks/
mv test/mocks/currency_repository_mock.go internal/domain/repository/mocks/

# Удалить старую директорию
rm -rf test/mocks
```

**Обновить импорты в core usecase тестах:**

```go
// Было
import "github.com/basilex/promenade/test/mocks"

// Стало
import "github.com/basilex/promenade/internal/domain/repository/mocks"
```

### Фаза 4: Создание handler тестов

#### Posts handlers

**Создать:** `internal/modules/posts/adapter/http/handler/post_handler_test.go`

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/posts/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/posts/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/usecase/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestPostHandler_CreatePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful creation", func(t *testing.T) {
		mockUC := new(mocks.MockPostUseCase)
		handler := handler.NewUserPostHandler(mockUC)

		userID := uuidv7.New()
		post := &entity.UserPost{
			ID:      uuidv7.New(),
			UserID:  userID,
			Title:   "Test Post",
			Content: "Test content",
		}

		mockUC.On("CreatePost", mock.Anything, userID, "Test Post", "Test content",
			mock.Anything, mock.Anything, mock.Anything).Return(post, nil)

		req := dto.CreatePostRequest{
			Title:   "Test Post",
			Content: "Test content",
		}
		body, _ := json.Marshal(req)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/posts", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", userID)

		handler.CreatePost(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})
}

// TODO: Add more tests
// - TestPostHandler_GetPost
// - TestPostHandler_UpdatePost
// - TestPostHandler_DeletePost
// - TestPostHandler_ListPosts
```

**Создать также:**

- `internal/modules/posts/adapter/http/handler/comment_handler_test.go`
- `internal/modules/profiles/adapter/http/handler/user_profile_handler_test.go`
- `internal/modules/profiles/adapter/http/handler/user_contact_handler_test.go`

### Фаза 5: Создание usecase моков (для handler тестов)

**Создать:** `internal/modules/posts/usecase/mocks/post_usecase_mock.go`

```go
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MockPostUseCase struct {
	mock.Mock
}

func (m *MockPostUseCase) CreatePost(ctx context.Context, userID uuidv7.UUID,
	title, content, excerpt string, tags, categories []string) (*entity.UserPost, error) {
	args := m.Called(ctx, userID, title, content, excerpt, tags, categories)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

func (m *MockPostUseCase) GetPost(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserPost), args.Error(1)
}

// TODO: Implement all UserPostUseCase interface methods
```

---

## Преимущества новой структуры

###  Полная независимость модулей

- Каждый модуль самодостаточен для разработки и тестирования
- Можно скопировать модуль в другой проект вместе с тестами
- Нет зависимости от core для тестирования

###  Соответствие Clean Architecture

- Тесты рядом с тестируемым кодом
- Моки реализуют интерфейсы из domain layer
- Четкое разделение слоев

###  Упрощение разработки

- Разработчик модуля работает только внутри модуля
- Легче понять, что тестируется и как
- Быстрее находить нужные тесты

###  Улучшение CI/CD

```bash
# Тестировать только posts module
go test ./internal/modules/posts/...

# Тестировать только core
go test ./internal/domain/... ./internal/usecase/... ./internal/adapter/...

# Тестировать всё
go test ./...
```

---

## Чеклист выполнения

### Posts Module

- [ ] Создать `internal/modules/posts/domain/repository/mocks/`
- [ ] Переместить и адаптировать моки репозиториев
- [ ] Переместить usecase тесты
- [ ] Обновить все импорты в тестах
- [ ] Создать handler тесты
- [ ] Создать usecase моки
- [ ] Запустить `go test ./internal/modules/posts/...`

### Profiles Module

- [ ] Создать `internal/modules/profiles/domain/` (переструктурировать)
- [ ] Создать `internal/modules/profiles/domain/repository/mocks/`
- [ ] Переместить и адаптировать моки
- [ ] Переместить usecase тесты
- [ ] Обновить импорты
- [ ] Создать handler тесты
- [ ] Создать usecase моки
- [ ] Запустить `go test ./internal/modules/profiles/...`

### Core

- [ ] Создать `internal/domain/repository/mocks/`
- [ ] Переместить core моки
- [ ] Обновить импорты в core usecase тестах
- [ ] Удалить `test/mocks/`
- [ ] Запустить `go test ./internal/...` (core only)

### Финальная проверка

- [ ] `go test ./...` - все тесты проходят
- [ ] Обновить документацию (TESTING_GUIDE.md)
- [ ] Обновить copilot-instructions.md
- [ ] Code review

---

## Дополнительные рекомендации

### 1. Создать helper для тестов модулей

**Создать:** `internal/modules/posts/tests/helpers.go`

```go
package tests

import (
	"testing"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func PostFixture(userID uuidv7.UUID, title string) *entity.UserPost {
	return &entity.UserPost{
		ID:      uuidv7.New(),
		UserID:  userID,
		Title:   title,
		Slug:    "test-slug",
		Content: "Test content",
		Status:  entity.PostStatusDraft,
	}
}

func CommentFixture(postID, userID uuidv7.UUID) *entity.PostComment {
	return &entity.PostComment{
		ID:      uuidv7.New(),
		PostID:  postID,
		UserID:  userID,
		Content: "Test comment",
	}
}
```

### 2. Создать интеграционные тесты модуля

**Создать:** `internal/modules/posts/tests/integration/post_flow_test.go`

```go
// End-to-end тесты для posts module
// Тестируют весь flow: handler -> usecase -> repository -> DB
```

### 3. Обновить Makefile

```makefile
# Test specific module
test-posts:
	go test -v ./internal/modules/posts/...

test-profiles:
	go test -v ./internal/modules/profiles/...

# Test core only
test-core:
	go test -v ./internal/domain/... ./internal/usecase/... ./internal/adapter/...
```

---

## Временная оценка

- **Фаза 1-3** (структура + перемещение): ~2-3 часа
- **Фаза 4** (handler тесты): ~4-6 часов
- **Фаза 5** (usecase моки): ~2-3 часа
- **Финальная проверка**: ~1-2 часа

**Итого:** 9-14 часов работы

---

## Следующие шаги

1. Подтверждение плана
2. Создание ветки `refactor/module-tests-independence`
3. Выполнение по фазам
4. Code review после каждой фазы
5. Merge в dev
