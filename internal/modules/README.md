# Promenade Modules

**Modules** are independent vertical slices that implement domain-specific business logic. Each module is a complete, self-contained feature that can be enabled/disabled without affecting other modules.

---

## 🧩 IModule Philosophy

### Core Principles

1. **Vertical Slicing**

   - One module = one domain area (e.g., "posts" includes posts + comments + likes)
   - Complete feature from database to HTTP handler in one place

2. **Independence**

   - Modules **never** import other modules
   - Modules **only** import `pkg/*` (shared utilities)
   - Modules communicate via events (event bus), not direct calls

3. **Self-Contained**

   - Own domain entities & business logic
   - Own database migrations (namespace-based)
   - Own HTTP handlers & routes
   - Own configuration
   - Own lifecycle (init, start, stop)

4. **Registry-Based**
   - Modules register themselves via `init()` functions
   - Core discovers modules automatically (no manual wiring)
   - Enable/disable via configuration only

---

## Available Modules

### **Posts** (`internal/modules/posts/`)

**Domain**: User-generated content management

**Entities**:

- `Post` - User posts with title, content, status (draft/published)
- `Comment` - Threaded comments on posts
- `Like` - Likes on comments

**Features**:

- Create/edit/delete posts (soft delete)
- Comment threads with depth limit
- Like system for comments
- Automated purge for deleted content

**Migrations**: 3 migrations (namespace: `posts`)

- `000001_create_user_posts` - Posts table with soft delete
- `000002_create_post_comments` - Comments with parent_id for threading
- `000003_create_comment_likes_table` - Likes with unique constraints

**Configuration** (`config/modules.yaml`):

```yaml
posts:
  settings:
    max_post_length: 10000
    max_comment_length: 2000
    max_comment_depth: 10
    allow_media: true
```

**Purge Policy**: 90 days retention for soft-deleted posts/comments

---

### **Profiles** (`internal/modules/profiles/`)

**Domain**: User profile and contact management

**Entities**:

- `UserProfile` - Extended user information (bio, avatar, location)
- `UserContact` - Contact methods (phone, social media, website)

**Features**:

- Profile management (create/update/delete)
- Multiple contact methods per user
- Location with country/timezone references
- Social media links

**Migrations**: 2 migrations (namespace: `profiles`)

- `000001_create_user_contacts` - Contact information table
- `000002_create_user_profiles` - User profiles with references to core tables

**Configuration** (`config/modules.yaml`):

```yaml
profiles:
  settings:
    max_bio_length: 500
    allow_custom_avatar: true
    require_location: false
```

---

---

### **Analytics** (`internal/modules/analytics/`) - _ Commercial (Active)_

**Domain**: Business analytics, metrics, reports, and dashboards

**Status**: Enabled - Requires license key (BASIC/PRO/ENTERPRISE tiers)

**Entities**:

- `Metric` - Performance metrics and KPIs
- `Report` - Custom reports and analytics
- `Dashboard` - Visual dashboards

**Features**:

- Metrics collection and aggregation
- Custom report generation
- Interactive dashboards
- Data retention policies (tier-based)

**Use Case**: Business intelligence, performance monitoring, data-driven insights

**License Tiers**:

- BASIC: 30-day retention, 5 reports, 2 dashboards
- PRO: 90-day retention, 10 reports, 5 dashboards
- ENTERPRISE: 365-day retention, unlimited

See [analytics/README.md](analytics/README.md) for full details.

---

### **Warehouse** (`internal/modules/warehouse/`) - Future Module

**Domain**: Inventory and product management

**Status**: Planned - Not yet implemented (commercial module)

**Planned Entities**:

- `Product` - Product catalog
- `Inventory` - Stock levels
- `Supplier` - Supplier management

**Use Case**: E-commerce, inventory systems, retail

_This module is in design phase. Structure exists as placeholder._

---

## Module Structure

Every module follows this structure:

```
internal/modules/{module}/
├── module.go                    # IModule registration & lifecycle
│
├── domain/                      # Domain layer (pure business logic)
│   └── entity/                  # Domain entities
│       ├── {entity}.go
│       └── {entity}_test.go
│
├── repository/                  # Data access layer
│   ├── {entity}_repository.go   # Interface
│   ├── postgres/                # PostgreSQL implementation
│   │   └── {entity}_repository.go
│   └── mock/                    # Mock implementation (for tests)
│       └── {entity}_repository.go
│
├── usecase/                     # Business logic layer
│   ├── {entity}_usecase.go
│   └── {entity}_usecase_test.go
│
├── adapter/                     # Adapters layer
│   └── handler/                 # HTTP handlers
│       ├── {entity}_handler.go  # HTTP handler
│       ├── dto/                 # Request/Response DTOs
│       │   └── {entity}_dto.go
│       └── router/              # Route registration
│           └── {module}_router.go
│
└── README.md                    # IModule-specific documentation
```

### Example: Posts Module

```
internal/modules/posts/
├── module.go                    # RegisterModule(), Initialize(), Start(), Stop()
│
├── domain/
│   └── entity/
│       ├── post.go              # Post entity
│       ├── comment.go           # Comment entity
│       └── like.go              # Like entity
│
├── repository/
│   ├── post_repository.go       # Interface: GetByID(), Create(), Update(), etc.
│   └── postgres/
│       ├── post_repository.go   # Implements using sqlx
│       ├── comment_repository.go
│       └── like_repository.go
│
├── usecase/
│   ├── post_usecase.go          # CreatePost(), PublishPost(), DeletePost()
│   ├── comment_usecase.go       # AddComment(), UpdateComment()
│   └── like_usecase.go          # ToggleLike(), GetLikeCount()
│
└── adapter/
    └── handler/
        ├── post_handler.go      # HandleCreatePost(), HandleGetPost()
        ├── comment_handler.go
        ├── like_handler.go
        ├── dto/
        │   ├── post_dto.go      # CreatePostRequest, PostResponse
        │   ├── comment_dto.go
        │   └── like_dto.go
        └── router/
            └── posts_router.go  # Register routes: POST /posts, GET /posts/:id
```

---

## Module Lifecycle

### 1. Registration (Auto-Discovery)

Modules register themselves in `init()`:

```go
// module.go
package posts

import "github.com/basilex/promenade/pkg/module"

type IModule struct {
    // dependencies initialized in Initialize()
}

func init() {
    module.DefaultRegistry.Register(&IModule{})
}
```

When the package is imported (`import _ "github.com/basilex/promenade/internal/modules/posts"`), `init()` runs automatically.

### 2. Initialization

Core calls `Initialize()` with services:

```go
func (m *IModule) Initialize(ctx context.Context, core module.Core) error {
    // Get services from core
    db := core.DB()
    logger := core.Logger()
    eventBus := core.EventBus()
    router := core.Router()

    // Initialize repositories
    postRepo := postgres.NewPostRepository(db)

    // Initialize use cases
    postUseCase := usecase.NewPostUseCase(postRepo, eventBus, logger)

    // Initialize handlers
    postHandler := handler.NewPostHandler(postUseCase, logger)

    // Register routes
    postsRouter := router.Group("/posts")
    postsRouter.POST("", postHandler.Create)
    postsRouter.GET("/:id", postHandler.GetByID)

    // Register purge policies
    purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
        EntityName:    "user_posts",
        RetentionDays: 90,
        Enabled:       true,
    })

    return nil
}
```

### 3. Startup

Core calls `Start()` for background workers:

```go
func (m *IModule) Start(ctx context.Context) error {
    // Start background workers (if any)
    // e.g., notification processor, scheduled tasks
    return nil
}
```

### 4. Shutdown

Core calls `Stop()` for graceful cleanup:

```go
func (m *IModule) Stop(ctx context.Context) error {
    // Stop background workers
    // Close connections
    // Flush caches
    return nil
}
```

---

## Creating a New IModule

### Step 1: Create Directory Structure

```bash
mkdir -p internal/modules/mymodule/{domain/entity,repository/postgres,usecase,adapter/handler/dto}
```

### Step 2: Create IModule Interface Implementation

```go
// internal/modules/mymodule/module.go
package mymodule

import (
    "context"
    "github.com/basilex/promenade/pkg/module"
)

type IModule struct {
    // Store dependencies if needed for Stop()
}

func (m *IModule) Name() string {
    return "mymodule"
}

func (m *IModule) Initialize(ctx context.Context, core module.Core) error {
    // Initialize repositories, use cases, handlers
    // Register routes
    // Register purge policies (if needed)
    return nil
}

func (m *IModule) Start(ctx context.Context) error {
    // Start background workers (optional)
    return nil
}

func (m *IModule) Stop(ctx context.Context) error {
    // Graceful shutdown (optional)
    return nil
}

func init() {
    module.DefaultRegistry.Register(&IModule{})
}
```

### Step 3: Create Domain Entities

```go
// internal/modules/mymodule/domain/entity/item.go
package entity

import "time"

type Item struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (i *Item) Validate() error {
    if i.Name == "" {
        return errors.New("name is required")
    }
    return nil
}
```

### Step 4: Create Migrations

```bash
make migrate-create MODULE=mymodule NAME=create_items_table
```

This creates:

- `migrations/mymodule/000001_create_items_table.up.sql`
- `migrations/mymodule/000001_create_items_table.down.sql`

Write SQL:

```sql
-- migrations/mymodule/000001_create_items_table.up.sql
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_items_updated_at
    BEFORE UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();
```

```sql
-- migrations/mymodule/000001_create_items_table.down.sql
DROP TRIGGER IF EXISTS trg_items_updated_at ON items;
DROP TABLE IF EXISTS items;
```

### Step 5: Enable IModule

Add to `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts
    - profiles
    - mymodule # ← Add here

  config:
    mymodule:
      version: "1.0.0"
      settings:
        max_items: 1000
```

### Step 6: Import IModule

Add to `cmd/api/main.go`:

```go
import (
    // ... existing imports
    _ "github.com/basilex/promenade/internal/modules/mymodule"
)
```

### Step 7: Run Migrations & Start

```bash
# Run migrations
make migrate-module MODULE=mymodule

# Start application
make dev
```

---

## IModule Conventions

### Naming

- **Package**: Lowercase singular or plural (e.g., `posts`, `profiles`, `warehouse`)
- **Entities**: Singular PascalCase (e.g., `Post`, `Comment`, `UserProfile`)
- **Files**: Snake_case (e.g., `post_repository.go`, `create_post_dto.go`)

### Imports

```go
//  ALLOWED
import "github.com/basilex/promenade/pkg/logger"
import "github.com/basilex/promenade/pkg/bus"
import "github.com/basilex/promenade/pkg/module"

//  FORBIDDEN
import "github.com/basilex/promenade/internal/modules/posts"    // Other module
import "github.com/basilex/promenade/internal/domain"           // Core domain
```

**Rule**: Modules can only import `pkg/*` packages.

### Database Tables

- Prefix with entity name: `user_posts`, `post_comments`, `comment_likes`
- Use UUID v7 for primary keys: `id UUID PRIMARY KEY DEFAULT uuid_v7()`
- Include timestamps: `created_at`, `updated_at`
- Use `deleted_at` for soft delete (user-generated content)

### Migrations

- Namespace = module name (lowercase)
- Sequential numbering: `000001`, `000002`, `000003`
- Descriptive names: `create_user_posts`, `add_post_status_index`
- Always provide DOWN migrations

---

## IModule Independence Rules

### Modules CAN

- Import packages from `pkg/*`
- Use services provided by Core (`module.Core` interface)
- Publish domain events to event bus
- Register themselves with registries (module, purge)
- Define their own entities, use cases, handlers
- Have their own migrations (namespace-based)
- Have their own configuration in `config/modules.yaml`

### Modules CANNOT

- Import other modules (`internal/modules/*`)
- Import core domain entities (`internal/domain/entity`)
- Call other modules' use cases directly
- Share database connections (use Core's DB)
- Depend on other modules being enabled
- Modify core configuration
- Access core's private implementation details

---

## 🔗 Inter-IModule Communication

Modules communicate **asynchronously** via the event bus:

```go
// IModule A: Publish event
eventBus.Publish(ctx, "user.updated", &UserUpdatedEvent{
    UserID: user.ID,
})

// IModule B: Subscribe to event
eventBus.Subscribe(ctx, "user.updated", func(ctx context.Context, e bus.Event) error {
    evt := e.(*UserUpdatedEvent)
    // Update local cache or related data
    return nil
})
```

**Never** call another module directly:

```go
//  WRONG
import "github.com/basilex/promenade/internal/modules/posts"
postsModule.CreatePost(...)

//  CORRECT
eventBus.Publish(ctx, "post.create.requested", &PostCreateEvent{...})
```

---

## 📖 Documentation Requirements

Each module should have a `README.md` with:

1. **Overview**: What problem does the module solve?
2. **Entities**: List of domain entities
3. **Features**: Key features and use cases
4. **Migrations**: List of migrations with descriptions
5. **Configuration**: Configuration options in `modules.yaml`
6. **API**: List of HTTP endpoints
7. **Events**: Events published/subscribed
8. **Dependencies**: External dependencies (if any)

See [posts/README.md](posts/README.md) for an example.

---

## Testing Modules

### Unit Tests

Test business logic in isolation:

```go
func TestCreatePost(t *testing.T) {
    mockRepo := mock.NewPostRepository()
    useCase := usecase.NewPostUseCase(mockRepo, nil, nil)

    post, err := useCase.CreatePost(ctx, "Title", "Content")
    assert.NoError(t, err)
    assert.Equal(t, "Title", post.Title)
}
```

### Integration Tests

Test with real database:

```go
func TestPostRepository_Create(t *testing.T) {
    db := test.SetupTestDB(t)
    repo := postgres.NewPostRepository(db)

    post := &entity.Post{
        ID:      uuidv7.New(),
        Title:   "Test",
        Content: "Content",
    }

    err := repo.Create(context.Background(), post)
    assert.NoError(t, err)
}
```

See [test/README.md](../../test/README.md) for testing infrastructure.

---

## 📚 Related Documentation

- **[README.md](../../README.md)** - Main project README
- **[internal/CORE.md](../CORE.md)** - Core responsibilities
- **[docs/MODULE_DEVELOPMENT.md](../../docs/MODULE_DEVELOPMENT.md)** - Detailed development guide
- **[docs/MODULE_INDEPENDENCE.md](../../docs/MODULE_INDEPENDENCE.md)** - Independence rules
- **[docs/MODULE_CONFIG_ARCHITECTURE.md](../../docs/MODULE_CONFIG_ARCHITECTURE.md)** - Configuration system
- **[migrations/README.md](../../migrations/README.md)** - Migration system

---

**Modules = Self-Contained Vertical Slices. Independence = Key.**
