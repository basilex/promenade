# Posts IModule

User-generated content management system with posts, threaded comments, and likes.

---

## Overview

The **posts** module provides a complete content management system for user-generated posts with threaded comments and a like system. It demonstrates vertical slicing - one module containing multiple related entities (posts, comments, likes) that work together.

**Key Features**:

- Create, edit, publish, and delete posts
- Threaded comments with configurable depth limit
- Like system for comments
- Soft delete with automated purge
- Draft/Published status workflow
- Rich text content support

**Status**: Enabled by default

**Namespace**: `posts` (for migrations)

---

## Entities

### Post

User-generated content with lifecycle management.

**Fields**:

- `id` (UUID v7) - Primary key
- `user_id` (UUID) - Author reference (FK to auth.users)
- `title` (VARCHAR 255) - Post title
- `content` (TEXT) - Post content (rich text)
- `status` (VARCHAR 20) - `draft` or `published`
- `created_at` (TIMESTAMP) - Creation timestamp
- `updated_at` (TIMESTAMP) - Last update timestamp
- `deleted_at` (TIMESTAMP) - Soft delete timestamp (nullable)

**Validations**:

- Title: Required, 1-255 characters
- Content: Required, max length configurable (default: 10,000 chars)
- Status: Must be `draft` or `published`

**Business Rules**:

- Draft posts visible only to author
- Published posts visible to all users
- Deleted posts hidden from views (soft delete)
- Purge after 90 days of soft deletion

**File**: [domain/entity/post.go](domain/entity/post.go)

---

### Comment

Threaded comments on posts with parent/child relationships.

**Fields**:

- `id` (UUID v7) - Primary key
- `post_id` (UUID) - Post reference (FK to user_posts)
- `user_id` (UUID) - Commenter reference (FK to auth.users)
- `parent_id` (UUID) - Parent comment for threading (nullable)
- `content` (TEXT) - Comment content
- `depth` (INTEGER) - Thread depth (0 = top-level)
- `created_at` (TIMESTAMP) - Creation timestamp
- `updated_at` (TIMESTAMP) - Last update timestamp
- `deleted_at` (TIMESTAMP) - Soft delete timestamp (nullable)

**Validations**:

- Content: Required, max length configurable (default: 2,000 chars)
- Depth: Must be ≤ max_comment_depth (default: 10)

**Business Rules**:

- Top-level comments have `parent_id = NULL` and `depth = 0`
- Replies increment depth: `parent_depth + 1`
- Max depth enforced to prevent infinite threading
- Deleted comments shown as "[deleted]" to preserve thread structure
- Purge after 90 days of soft deletion

**File**: [domain/entity/comment.go](domain/entity/comment.go)

---

### Like

Like system for comments (not posts directly).

**Fields**:

- `id` (UUID v7) - Primary key
- `comment_id` (UUID) - Comment reference (FK to post_comments)
- `user_id` (UUID) - User who liked (FK to auth.users)
- `created_at` (TIMESTAMP) - Creation timestamp

**Constraints**:

- Unique per (comment_id, user_id) - one like per user per comment

**Business Rules**:

- Toggle behavior: Like if not exists, unlike if exists
- No soft delete (hard delete on unlike)
- Like count aggregated via COUNT query

**File**: [domain/entity/like.go](domain/entity/like.go)

---

## Database Schema

### Tables

**`user_posts`** (Namespace: `posts`, Migration: 000001)

```sql
CREATE TABLE user_posts (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_user_posts_user_id ON user_posts(user_id);
CREATE INDEX idx_user_posts_status ON user_posts(status);
CREATE INDEX idx_user_posts_deleted_at ON user_posts(deleted_at);
```

**`post_comments`** (Namespace: `posts`, Migration: 000002)

```sql
CREATE TABLE post_comments (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    post_id UUID NOT NULL REFERENCES user_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES post_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    depth INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_post_comments_post_id ON post_comments(post_id);
CREATE INDEX idx_post_comments_parent_id ON post_comments(parent_id);
CREATE INDEX idx_post_comments_deleted_at ON post_comments(deleted_at);
```

**`comment_likes`** (Namespace: `posts`, Migration: 000003)

```sql
CREATE TABLE comment_likes (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    comment_id UUID NOT NULL REFERENCES post_comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(comment_id, user_id)
);

CREATE INDEX idx_comment_likes_comment_id ON comment_likes(comment_id);
CREATE INDEX idx_comment_likes_user_id ON comment_likes(user_id);
```

---

## Migrations

**Namespace**: `posts`

| Version | Name                         | Description                                                   |
| ------- | ---------------------------- | ------------------------------------------------------------- |
| 000001  | `create_user_posts`          | Creates user_posts table with soft delete support             |
| 000002  | `create_post_comments`       | Creates post_comments table with threading (parent_id, depth) |
| 000003  | `create_comment_likes_table` | Creates comment_likes table with unique constraint            |

**Run migrations**:

```bash
make migrate-module MODULE=posts
```

**Rollback**:

```bash
make migrate-rollback MODULE=posts STEPS=1
```

---

## Configuration

Configuration in `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts

  config:
    posts:
      version: "1.0.0"
      settings:
        max_post_length: 10000 # Max characters in post content
        max_comment_length: 2000 # Max characters in comment content
        max_comment_depth: 10 # Max threading depth
        allow_media: true # Allow media attachments (future)
```

**Environment Variables**:

- None (uses main app config for DB, logger, event bus)

---

## API Endpoints

### Posts

#### Create Post

```http
POST /api/v1/posts
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "My First Post",
  "content": "This is the content of my post.",
  "status": "draft"
}
```

**Response**:

```json
{
  "status": "success",
  "data": {
    "id": "01931234-5678-7abc-def0-123456789abc",
    "user_id": "01931234-0000-7000-8000-000000000001",
    "title": "My First Post",
    "content": "This is the content of my post.",
    "status": "draft",
    "created_at": "2025-12-22T10:00:00Z",
    "updated_at": "2025-12-22T10:00:00Z"
  }
}
```

#### Get Post

```http
GET /api/v1/posts/{id}
```

#### Update Post

```http
PUT /api/v1/posts/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "title": "Updated Title",
  "content": "Updated content",
  "status": "published"
}
```

#### Delete Post (Soft Delete)

```http
DELETE /api/v1/posts/{id}
Authorization: Bearer {token}
```

#### List Posts

```http
GET /api/v1/posts?page=1&page_size=20&status=published
```

---

### Comments

#### Add Comment

```http
POST /api/v1/posts/{post_id}/comments
Authorization: Bearer {token}
Content-Type: application/json

{
  "content": "Great post!",
  "parent_id": null  // or UUID for replies
}
```

#### Get Comments for Post

```http
GET /api/v1/posts/{post_id}/comments
```

**Response** (with threading):

```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "content": "Great post!",
      "depth": 0,
      "replies": [
        {
          "id": "...",
          "content": "I agree!",
          "depth": 1,
          "replies": []
        }
      ]
    }
  ]
}
```

#### Update Comment

```http
PUT /api/v1/comments/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "content": "Updated comment"
}
```

#### Delete Comment

```http
DELETE /api/v1/comments/{id}
Authorization: Bearer {token}
```

---

### Likes

#### Toggle Like

```http
POST /api/v1/comments/{comment_id}/like
Authorization: Bearer {token}
```

**Response**:

```json
{
  "status": "success",
  "data": {
    "liked": true,
    "like_count": 42
  }
}
```

#### Get Like Count

```http
GET /api/v1/comments/{comment_id}/likes
```

---

## Events

### Published Events

#### `post.created`

```go
type PostCreatedEvent struct {
    PostID   string
    UserID   string
    Title    string
    Status   string
}
```

#### `post.published`

```go
type PostPublishedEvent struct {
    PostID   string
    UserID   string
    Title    string
}
```

#### `post.deleted`

```go
type PostDeletedEvent struct {
    PostID   string
    UserID   string
}
```

#### `comment.created`

```go
type CommentCreatedEvent struct {
    CommentID  string
    PostID     string
    UserID     string
    ParentID   *string  // nil for top-level
}
```

#### `like.toggled`

```go
type LikeToggledEvent struct {
    CommentID  string
    UserID     string
    Liked      bool  // true = liked, false = unliked
}
```

### Subscribed Events

- `user.deleted` - Cascades via DB foreign keys (ON DELETE CASCADE)

---

## Purge Policy

**Retention**: 90 days for soft-deleted records

**Tables**:

- `user_posts` (WHERE deleted_at IS NOT NULL)
- `post_comments` (WHERE deleted_at IS NOT NULL)

**Schedule**: Daily at 02:00 AM (configured in `internal/infrastructure/scheduler`)

**Manual Purge**:

```bash
# Via API (requires admin role)
POST /api/v1/admin/purge/execute
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "entities": ["user_posts", "post_comments"]
}
```

**Registration** (in `module.go`):

```go
purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
})
```

---

## Testing

### Unit Tests

```bash
# Run module unit tests
go test ./internal/modules/posts/usecase/... -v
```

### Integration Tests

```bash
# Run module integration tests
make test-integration
```

**Test Files**:

- `usecase/post_usecase_test.go` - Business logic tests
- `repository/postgres/post_repository_test.go` - Database tests
- `test/smoke/posts_smoke_test.go` - End-to-end smoke tests

**Test Coverage**: 120+ tests across posts/comments/likes (part of global test suite)

---

## File Structure

```
internal/modules/posts/
├── module.go                          # IModule registration & lifecycle
│
├── domain/
│   └── entity/
│       ├── post.go                    # Post entity with validation
│       ├── comment.go                 # Comment entity with threading
│       └── like.go                    # Like entity
│
├── repository/
│   ├── post_repository.go             # Repository interface
│   ├── comment_repository.go
│   ├── like_repository.go
│   └── postgres/
│       ├── post_repository.go         # PostgreSQL implementation
│       ├── comment_repository.go
│       └── like_repository.go
│
├── usecase/
│   ├── post_usecase.go                # Post business logic
│   ├── post_usecase_test.go
│   ├── comment_usecase.go             # Comment business logic
│   └── like_usecase.go                # Like business logic
│
└── adapter/
    └── handler/
        ├── post_handler.go            # HTTP handlers
        ├── comment_handler.go
        ├── like_handler.go
        ├── dto/
        │   ├── post_dto.go            # Request/Response DTOs
        │   ├── comment_dto.go
        │   └── like_dto.go
        └── router/
            └── posts_router.go        # Route registration
```

---

## Usage Examples

### Creating a Post Programmatically

```go
import (
    "github.com/basilex/promenade/internal/modules/posts/usecase"
    "github.com/basilex/promenade/internal/modules/posts/domain/entity"
)

// In Initialize()
postUseCase := usecase.NewPostUseCase(postRepo, eventBus, logger)

// Create draft post
post, err := postUseCase.CreatePost(ctx, entity.Post{
    UserID:  userID,
    Title:   "My Post",
    Content: "Content here",
    Status:  "draft",
})

// Publish post
err = postUseCase.PublishPost(ctx, post.ID)
```

### Adding Threaded Comments

```go
// Add top-level comment
comment1, err := commentUseCase.AddComment(ctx, entity.Comment{
    PostID:   postID,
    UserID:   userID,
    Content:  "Great post!",
    ParentID: nil,
    Depth:    0,
})

// Add reply
comment2, err := commentUseCase.AddComment(ctx, entity.Comment{
    PostID:   postID,
    UserID:   userID2,
    Content:  "I agree!",
    ParentID: &comment1.ID,
    Depth:    1,
})
```

### Toggle Like

```go
liked, likeCount, err := likeUseCase.ToggleLike(ctx, commentID, userID)
if err != nil {
    return err
}

if liked {
    logger.Info("Comment liked", "comment_id", commentID, "likes", likeCount)
} else {
    logger.Info("Comment unliked", "comment_id", commentID, "likes", likeCount)
}
```

---

## Permissions

**RBAC Permissions** (if enabled):

- `posts:create` - Create posts
- `posts:read` - Read posts (auto-granted to all authenticated users)
- `posts:update` - Update own posts
- `posts:delete` - Delete own posts
- `posts:update:any` - Update any user's posts (admin)
- `posts:delete:any` - Delete any user's posts (admin)

**Default Roles**:

- **User**: `posts:create`, `posts:read`, `posts:update`, `posts:delete`
- **Admin**: All permissions + `posts:update:any`, `posts:delete:any`

---

## Future Enhancements

Potential features (not yet implemented):

- [ ] Media attachments (images, videos)
- [ ] Post reactions (beyond likes on comments)
- [ ] Post tags/categories
- [ ] Search and filtering
- [ ] Mentions (@username)
- [ ] Post analytics (views, engagement)
- [ ] Scheduled publishing
- [ ] Content moderation flags

---

## 📚 Related Documentation

- **[../README.md](../README.md)** - IModule system overview
- **[../../README.md](../../README.md)** - Main project README
- **[../../docs/SOFT_DELETE.md](../../docs/SOFT_DELETE.md)** - Soft delete pattern
- **[../../docs/MODULE_DEVELOPMENT.md](../../docs/MODULE_DEVELOPMENT.md)** - IModule development guide
- **[../../migrations/README.md](../../migrations/README.md)** - Migration system

---

**IModule Status**: Production-ready | 120+ tests | 3 migrations | RBAC-enabled
