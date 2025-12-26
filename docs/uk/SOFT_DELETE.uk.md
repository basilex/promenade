[ English](../SOFT_DELETE.md) |  **Українська** | [ Deutsch](SOFT_DELETE.de.md) | [ Português](SOFT_DELETE.pt.md) | [ Español](SOFT_DELETE.es.md)

---

# Посібник з реалізації Soft Delete

## Огляд

Promenade реалізує **soft delete** для таблиць `user_posts` та `post_comments`. Soft delete позначає записи як видалені, встановлюючи timestamp у колонці `deleted_at` замість фізичного видалення їх з бази даних.

## Переваги

- **Відновлення даних**: Видалений контент можна відновити
- **Аудит**: Відстеження часу видалення контенту
- **Цілісність посилань**: Foreign key зв'язки залишаються неушкодженими
- **Аналітика**: Історичні дані залишаються доступними для аналізу
- **Відповідність вимогам**: Виконання вимог щодо зберігання даних

## Реалізація

### Схема бази даних

Обидві таблиці `user_posts` та `post_comments` включають:

```sql
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
```

#### Індекси

Часткові індекси виключають soft-deleted записи для кращої продуктивності:

```sql
-- user_posts
CREATE INDEX idx_user_posts_user_id ON user_posts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_posts_published ON user_posts(published_at DESC)
    WHERE status = 'published' AND is_public = true AND deleted_at IS NULL;

-- post_comments
CREATE INDEX idx_post_comments_post_id ON post_comments(post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_user_id ON post_comments(user_id) WHERE deleted_at IS NULL;
```

Індекс для soft-deleted записів (для адміністративних/відновлювальних операцій):

```sql
CREATE INDEX idx_user_posts_deleted_at ON user_posts(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_post_comments_deleted_at ON post_comments(deleted_at) WHERE deleted_at IS NOT NULL;
```

### Domain Entities

#### UserPost Entity

```go
type UserPost struct {
    // ... other fields
    DeletedAt *time.Time `db:"deleted_at" validate:"omitempty"`
    // ... timestamps
}

func (p *UserPost) SoftDelete() {
    now := time.Now()
    p.DeletedAt = &now
    p.UpdatedAt = now
}

func (p *UserPost) Restore() {
    p.DeletedAt = nil
    p.UpdatedAt = time.Now()
}
```

#### PostComment Entity

```go
type PostComment struct {
    // ... other fields
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at" validate:"omitempty"`
    // ... timestamps
}

func (c *PostComment) SoftDelete() {
    now := time.Now()
    c.DeletedAt = &now
    c.UpdatedAt = now
}

func (c *PostComment) Restore() {
    c.DeletedAt = nil
    c.UpdatedAt = time.Now()
}
```

### Repository Interface

```go
type IUserPostRepository interface {
    // ... CRUD methods

    // Soft delete operations
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}

type IPostCommentRepository interface {
    // ... CRUD methods

    // Soft delete operations
    SoftDelete(ctx context.Context, id uuidv7.UUID) error
    Restore(ctx context.Context, id uuidv7.UUID) error
}
```

### Реалізація Repository

#### Soft Delete

```go
func (r *userPostRepository) SoftDelete(ctx context.Context, id uuidv7.UUID) error {
    query := `UPDATE user_posts SET deleted_at = NOW(), updated_at = NOW()
              WHERE id = $1 AND deleted_at IS NULL`

    executor := r.getExecutor(ctx)
    result, err := executor.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to soft delete post: %w", err)
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return entity.ErrNotFound
    }

    return nil
}
```

#### Restore

```go
func (r *userPostRepository) Restore(ctx context.Context, id uuidv7.UUID) error {
    query := `UPDATE user_posts SET deleted_at = NULL, updated_at = NOW()
              WHERE id = $1 AND deleted_at IS NOT NULL`

    executor := r.getExecutor(ctx)
    result, err := executor.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to restore post: %w", err)
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return entity.ErrNotFound
    }

    return nil
}
```

#### Фільтрація запитів

**КРИТИЧНО**: Всі SELECT запити повинні фільтрувати soft-deleted записи:

```go
func (r *userPostRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `
        SELECT id, user_id, title, slug, content, ...
        FROM user_posts
        WHERE id = $1 AND deleted_at IS NULL  -- CRITICAL: Filter soft-deleted
    `
    return r.scanPost(ctx, query, id)
}

func (r *userPostRepository) List(ctx context.Context, params ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error) {
    conditions := []string{"deleted_at IS NULL"}  // CRITICAL: Base condition
    // ... additional filters
}
```

### Use Case Layer

```go
func (uc *userPostUseCase) SoftDeletePost(ctx context.Context, userID, postID uuidv7.UUID) error {
    // Get post to check ownership
    post, err := uc.postRepo.GetByID(ctx, postID)
    if err != nil {
        return fmt.Errorf("failed to get post: %w", err)
    }

    // Check authorization
    if post.UserID != userID {
        return ErrUnauthorized
    }

    // Perform soft delete
    if err := uc.postRepo.SoftDelete(ctx, postID); err != nil {
        return fmt.Errorf("failed to soft delete post: %w", err)
    }

    return nil
}
```

#### Comment Soft Delete з побічними ефектами

```go
func (uc *postCommentUseCase) DeleteComment(ctx context.Context, id, userID uuidv7.UUID) error {
    comment, err := uc.commentRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, entity.ErrNotFound) {
            return ErrCommentNotFound
        }
        return err
    }

    // Check ownership
    if comment.UserID != userID {
        return ErrUnauthorizedComment
    }

    // Soft delete
    if err := uc.commentRepo.SoftDelete(ctx, id); err != nil {
        return err
    }

    // Decrement reply count on parent if it's a reply
    if comment.ParentID != nil {
        _ = uc.commentRepo.DecrementReplies(ctx, *comment.ParentID)
    }

    // Decrement post comment count
    _ = uc.postRepo.DecrementComments(ctx, comment.PostID)

    return nil
}
```

## Тестування

### Unit тести (Entity)

```go
func TestUserPost_SoftDelete(t *testing.T) {
    post, _ := NewUserPost(userID, "Test", "test", "Content")
    assert.Nil(t, post.DeletedAt)

    post.SoftDelete()

    assert.NotNil(t, post.DeletedAt)
}

func TestUserPost_Restore(t *testing.T) {
    post, _ := NewUserPost(userID, "Test", "test", "Content")
    post.SoftDelete()
    assert.NotNil(t, post.DeletedAt)

    post.Restore()

    assert.Nil(t, post.DeletedAt)
}
```

### Integration тести (Repository)

```go
func TestUserPostRepository_SoftDelete(t *testing.T) {
    // ... setup

    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    // Should not be found after soft delete
    _, err = repo.GetByID(ctx, post.ID)
    assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserPostRepository_Restore(t *testing.T) {
    // ... setup
    err := repo.SoftDelete(ctx, post.ID)
    require.NoError(t, err)

    err = repo.Restore(ctx, post.ID)
    require.NoError(t, err)

    // Verify restored
    restored, err := repo.GetByID(ctx, post.ID)
    require.NoError(t, err)
    assert.Equal(t, post.ID, restored.ID)
    assert.Nil(t, restored.DeletedAt)
}
```

### Smoke тести (End-to-End)

```go
t.Run("[+] Soft_delete_and_restore", func(t *testing.T) {
    // Soft delete
    require.NoError(t, postRepo.SoftDelete(ctx, postID))

    // Verify not accessible
    _, err := postRepo.GetByID(ctx, postID)
    assert.ErrorIs(t, err, entity.ErrNotFound)

    // Restore
    require.NoError(t, postRepo.Restore(ctx, postID))

    // Verify accessible again
    restored, err := postRepo.GetByID(ctx, postID)
    require.NoError(t, err)
    assert.Equal(t, postID, restored.ID)
})
```

## Поширені патерни

### 1. Перевірка перед Soft Delete

Завжди перевіряйте права власності/дозволи перед soft delete:

```go
post, err := uc.postRepo.GetByID(ctx, postID)
if err != nil {
    return err
}

if post.UserID != userID {
    return ErrUnauthorized
}

return uc.postRepo.SoftDelete(ctx, postID)
```

### 2. Ідемпотентні операції

Soft delete повертає `ErrNotFound`, якщо запис вже видалено:

```go
// First delete - success
err := repo.SoftDelete(ctx, postID)  // nil

// Second delete - error
err = repo.SoftDelete(ctx, postID)   // entity.ErrNotFound
```

### 3. Патерн виключення з запитів

Всі list/search запити повинні виключати soft-deleted записи:

```go
conditions := []string{"deleted_at IS NULL"}

// Add additional filters
if params.UserID != nil {
    conditions = append(conditions, fmt.Sprintf("user_id = '%s'", *params.UserID))
}

query := fmt.Sprintf("SELECT ... WHERE %s", strings.Join(conditions, " AND "))
```

### 4. Адміністративні запити (включаючи видалені)

Для адміністративних інтерфейсів створіть окремі методи для запиту видалених записів:

```go
// Future implementation
func (r *userPostRepository) GetByIDIncludingDeleted(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `SELECT ... FROM user_posts WHERE id = $1`  // No deleted_at filter
    return r.scanPost(ctx, query, id)
}
```

## Найкращі практики

### РОБІТЬ [+]

- Завжди фільтруйте `deleted_at IS NULL` у SELECT запитах
- Використовуйте часткові індекси: `WHERE deleted_at IS NULL` на часто запитуваних колонках
- Перевіряйте права власності перед soft delete
- Оновлюйте пов'язані лічильники (як reply_count, comment_count) після soft delete
- Тестуйте як операції soft delete, так і restore
- Документуйте, які entity підтримують soft delete

### НЕ РОБІТЬ [-]

- Не забувайте `deleted_at IS NULL` у запитах (поширена помилка!)
- Не використовуйте hard delete (`DELETE FROM`) для soft-deletable entities
- Не розкривайте soft-deleted записи у публічних API без авторизації
- Не відновлюйте без перевірки дозволів (у продакшені додайте перевірку прав власності)
- Не каскадуйте soft deletes автоматично (рішення дизайну для кожного випадку використання)

## Відомі обмеження

### Поточна реалізація

1. **Немає перевірки прав власності при restore**: `RestorePost()` не перевіряє права власності, оскільки `GetByID()` фільтрує видалені записи. У продакшені реалізуйте `GetByIDIncludingDeleted()` для цієї перевірки.

2. **Немає каскадного soft delete**: Видалення поста не призводить автоматично до soft-delete його коментарів. Це навмисне рішення дизайну для збереження історії коментарів.

3. **Немає timestamp відновлення**: Реалізація не відстежує, коли запис було відновлено (за потреби можна додати `restored_at`).

## Майбутні покращення

- [ ] Додати `GetByIDIncludingDeleted()` для адміністративних операцій
- [ ] Додати `ListDeleted()` для адміністративного UI кошика/відновлення
- [ ] Додати timestamp `restored_at` для аудиту
- [ ] Додати опцію каскадного soft delete для posts → comments
- [ ] Додати масові операції soft delete/restore
- [ ] Додати автоматичне hard delete через X днів (заплановане завдання)

## Дивіться також

- [UUID v7 Migration Guide](UUID_V7_MIGRATION.md) - Стратегія первинних ключів
- [Testing Guide](TESTING_GUIDE.md) - Патерни тестування та помічники
- [Database Migrations](../migrations/) - Визначення схем
- [Repository Base](../internal/adapter/repository/postgres/base_repository.go) - Загальні операції БД
