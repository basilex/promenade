# Soft Delete Implementation Guide

## Overview

Promenade implements **soft delete** for `user_posts` and `post_comments` tables. Soft delete marks records as deleted by setting a timestamp in the `deleted_at` column instead of physically removing them from the database.

## Benefits

- **Data Recovery**: Deleted content can be restored
- **Audit Trail**: Track when content was deleted
- **Referential Integrity**: Foreign key relationships remain intact
- **Analytics**: Historical data remains available for analysis
- **Compliance**: Meet data retention requirements

## Implementation

### Database Schema

Both `user_posts` and `post_comments` tables include:

```sql
deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
```

#### Indexes

Partial indexes exclude soft-deleted records for better performance:

```sql
-- user_posts
CREATE INDEX idx_user_posts_user_id ON user_posts(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_posts_published ON user_posts(published_at DESC)
    WHERE status = 'published' AND is_public = true AND deleted_at IS NULL;

-- post_comments
CREATE INDEX idx_post_comments_post_id ON post_comments(post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_post_comments_user_id ON post_comments(user_id) WHERE deleted_at IS NULL;
```

Index for soft-deleted records (for admin/restore operations):

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

### Repository Implementation

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

#### Query Filtering

**CRITICAL**: All SELECT queries must filter out soft-deleted records:

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

#### Comment Soft Delete with Side Effects

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

## Testing

### Unit Tests (Entity)

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

### Integration Tests (Repository)

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

### Smoke Tests (End-to-End)

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

## Common Patterns

### 1. Check Before Soft Delete

Always verify ownership/permissions before soft delete:

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

### 2. Idempotent Operations

Soft delete returns `ErrNotFound` if record already deleted:

```go
// First delete - success
err := repo.SoftDelete(ctx, postID)  // nil

// Second delete - error
err = repo.SoftDelete(ctx, postID)   // entity.ErrNotFound
```

### 3. Query Exclusion Pattern

All list/search queries must exclude soft-deleted records:

```go
conditions := []string{"deleted_at IS NULL"}

// Add additional filters
if params.UserID != nil {
    conditions = append(conditions, fmt.Sprintf("user_id = '%s'", *params.UserID))
}

query := fmt.Sprintf("SELECT ... WHERE %s", strings.Join(conditions, " AND "))
```

### 4. Admin Queries (Including Deleted)

For admin interfaces, create separate methods to query deleted records:

```go
// Future implementation
func (r *userPostRepository) GetByIDIncludingDeleted(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error) {
    query := `SELECT ... FROM user_posts WHERE id = $1`  // No deleted_at filter
    return r.scanPost(ctx, query, id)
}
```

## Best Practices

### DO [+]

- Always filter `deleted_at IS NULL` in SELECT queries
- Use partial indexes: `WHERE deleted_at IS NULL` on frequently queried columns
- Check ownership before soft delete
- Update related counters (like reply_count, comment_count) after soft delete
- Test both soft delete and restore operations
- Document which entities support soft delete

### DON'T [-]

- Don't forget `deleted_at IS NULL` in queries (common bug!)
- Don't use hard delete (`DELETE FROM`) for soft-deletable entities
- Don't expose soft-deleted records in public APIs without authorization
- Don't restore without verifying permissions (in production, add ownership check)
- Don't cascade soft deletes automatically (design decision per use case)

## Known Limitations

### Current Implementation

1. **No ownership check on restore**: `RestorePost()` doesn't verify ownership because `GetByID()` filters out deleted records. In production, implement `GetByIDIncludingDeleted()` for this check.

2. **No cascade soft delete**: Deleting a post doesn't automatically soft-delete its comments. This is intentional - design decision to preserve comment history.

3. **No restore timestamp**: The implementation doesn't track when a record was restored (could add `restored_at` if needed).

## Future Enhancements

- [ ] Add `GetByIDIncludingDeleted()` for admin operations
- [ ] Add `ListDeleted()` for admin trash/recovery UI
- [ ] Add `restored_at` timestamp for audit trail
- [ ] Add cascade soft delete option for posts → comments
- [ ] Add bulk soft delete/restore operations
- [ ] Add automatic hard delete after X days (scheduled job)

## See Also

- [UUID v7 Migration Guide](UUID_V7_MIGRATION.md) - Primary key strategy
- [Testing Guide](TESTING_GUIDE.md) - Testing patterns and helpers
- [Database Migrations](../migrations/) - Schema definitions
- [Repository Base](../internal/adapter/repository/postgres/base_repository.go) - Common DB operations
