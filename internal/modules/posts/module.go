package posts

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/posts/adapter/http/handler"
	postpurge "github.com/basilex/promenade/internal/modules/posts/adapter/purge"
	"github.com/basilex/promenade/internal/modules/posts/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/posts/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/module"
	moduleconfig "github.com/basilex/promenade/pkg/module/config"
	"github.com/basilex/promenade/pkg/purge"
)

// PostsModule implements module.Module for user posts functionality
type PostsModule struct {
	db             *sqlx.DB
	eventBus       bus.Bus
	config         *moduleconfig.Config // Module's own config
	postUseCase    usecase.UserPostUseCase
	postHandler    *handler.UserPostHandler
	commentUseCase usecase.CommentUseCase
	commentHandler *handler.CommentHandler
}

// New creates a new posts module instance
func New() module.Module {
	return &PostsModule{}
}

// Metadata returns module information
func (m *PostsModule) Metadata() module.Metadata {
	return module.Metadata{
		Name:        "posts",
		DisplayName: "User-Generated Content",
		Version:     "1.0.0",
		Description: "Complete UGC solution: posts, comments, likes, and engagement",
		Author:      "Promenade Team",
		License:     "MIT",
		Tags:        []string{"social", "content", "ugc", "comments"},
	}
}

// Dependencies returns list of module dependencies
func (m *PostsModule) Dependencies() []string {
	return []string{} // No dependencies
}

// Initialize sets up the module components
func (m *PostsModule) Initialize(ctx context.Context, core *module.Core) error {
	slog.Info("Initializing posts module")

	m.db = core.DB
	m.eventBus = core.EventBus

	// Load module's own configuration
	cfg, err := moduleconfig.Load("internal/modules/posts/config", "promenade")
	if err != nil {
		slog.Warn("Failed to load posts module config, using defaults", "error", err)
		// Continue with defaults
	} else {
		m.config = cfg
		slog.Info("Posts module config loaded",
			"version", cfg.Module.Version,
			"enabled", cfg.Module.Enabled,
		)
	}

	// Get module-specific settings from config (all values come from config files)
	maxCommentLength := 2000    // Fallback if config missing
	maxCommentDepth := 10       // Fallback if config missing
	postsRetentionDays := 90    // Fallback if config missing
	commentsRetentionDays := 30 // Fallback if config missing

	if m.config != nil {
		// Read comment settings from posts.comments section
		if posts, ok := m.config.Settings["posts"].(map[string]interface{}); ok {
			if comments, ok := posts["comments"].(map[string]interface{}); ok {
				if v, ok := comments["max_content_length"].(int); ok {
					maxCommentLength = v
				}
				if v, ok := comments["max_depth"].(int); ok {
					maxCommentDepth = v
				}
			}
		}

		// Read purge retention from purge section (config helper)
		postsRetentionDays = m.config.GetRetentionDays("posts", 90)
		commentsRetentionDays = m.config.GetRetentionDays("comments", 30)
	}

	slog.Info("Posts module settings",
		"max_comment_length", maxCommentLength,
		"max_comment_depth", maxCommentDepth,
		"posts_retention_days", postsRetentionDays,
		"comments_retention_days", commentsRetentionDays,
	)

	// Initialize repositories
	postRepo := postgres.NewUserPostRepository(m.db)
	commentRepo := postgres.NewCommentRepository(m.db)

	// Initialize use cases
	m.postUseCase = usecase.NewUserPostUseCase(postRepo)
	m.commentUseCase = usecase.NewCommentUseCase(commentRepo, postRepo)

	// Initialize handlers
	m.postHandler = handler.NewUserPostHandler(m.postUseCase)
	m.commentHandler = handler.NewCommentHandler(m.commentUseCase)

	// Register purge handler and retention policy for soft-deleted posts
	postPurgeHandler := postpurge.NewPostPurgeHandler(m.db)
	if err := purge.DefaultRegistry.Register(postPurgeHandler); err != nil {
		slog.Error("Failed to register post purge handler", "error", err)
		return err
	}
	
	postPolicy := purge.RetentionPolicy{
		EntityName:    "user_posts",
		RetentionDays: postsRetentionDays,
		Enabled:       true,
	}
	if err := purge.DefaultPolicyRegistry.RegisterPolicy(postPolicy); err != nil {
		slog.Error("Failed to register post retention policy", "error", err)
		return err
	}
	slog.Info("Registered purge for user_posts", "retention_days", postsRetentionDays)

	// Register purge handler and retention policy for soft-deleted comments
	commentPurgeHandler := postpurge.NewCommentPurgeHandler(m.db)
	if err := purge.DefaultRegistry.Register(commentPurgeHandler); err != nil {
		slog.Error("Failed to register comment purge handler", "error", err)
		return err
	}
	
	commentPolicy := purge.RetentionPolicy{
		EntityName:    "post_comments",
		RetentionDays: commentsRetentionDays,
		Enabled:       true,
	}
	if err := purge.DefaultPolicyRegistry.RegisterPolicy(commentPolicy); err != nil {
		slog.Error("Failed to register comment retention policy", "error", err)
		return err
	}
	slog.Info("Registered purge for post_comments", "retention_days", commentsRetentionDays)

	slog.Info("Posts module initialized successfully")
	return nil
}

// RegisterRoutes registers HTTP routes for this module
func (m *PostsModule) RegisterRoutes(router *gin.RouterGroup) {
	slog.Info("Registering posts module routes")

	// Posts routes
	postsGroup := router.Group("/posts")
	{
		// Public routes
		postsGroup.GET("", m.postHandler.ListPosts)
		postsGroup.GET("/:id", m.postHandler.GetPost)

		// Protected routes
		postsGroup.POST("", m.postHandler.CreatePost)
		postsGroup.PUT("/:id", m.postHandler.UpdatePost)
		postsGroup.DELETE("/:id", m.postHandler.DeletePost)

		// Comment routes under posts
		postsGroup.GET("/:postId/comments", m.commentHandler.GetPostComments)
		postsGroup.POST("/:postId/comments", m.commentHandler.CreateComment)
	}

	// Comments routes (direct access)
	commentsGroup := router.Group("/comments")
	{
		commentsGroup.GET("/:id", m.commentHandler.GetComment)
		commentsGroup.PUT("/:id", m.commentHandler.UpdateComment)
		commentsGroup.DELETE("/:id", m.commentHandler.DeleteComment)
		commentsGroup.GET("/:id/replies", m.commentHandler.GetCommentReplies)
	}

	// User comments route
	router.GET("/users/:userId/comments", m.commentHandler.GetUserComments)

	slog.Info("Posts module routes registered successfully")
}

// RegisterMigrations returns migrations for this module
func (m *PostsModule) RegisterMigrations() []module.Migration {
	return []module.Migration{
		{
			Version:     6,
			Description: "Create user_posts table with soft delete support",
			Up:          "-- Migration managed in migrations/000006_create_user_posts.up.sql",
			Down:        "-- Migration managed in migrations/000006_create_user_posts.down.sql",
		},
		{
			Version:     7,
			Description: "Create post_comments table with soft delete support",
			Up:          "-- Migration managed in migrations/000007_create_post_comments.up.sql",
			Down:        "-- Migration managed in migrations/000007_create_post_comments.down.sql",
		},
	}
}

// RegisterEventHandlers subscribes to events
func (m *PostsModule) RegisterEventHandlers(eventBus bus.Bus) error {
	slog.Info("Registering posts module event handlers")
	return nil
}

// RegisterPermissions returns permissions required by this module
func (m *PostsModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		// Posts permissions
		{Resource: "posts", Action: "create", Description: "Create new posts"},
		{Resource: "posts", Action: "read", Description: "View posts"},
		{Resource: "posts", Action: "update", Description: "Update own posts"},
		{Resource: "posts", Action: "delete", Description: "Delete own posts"},
		{Resource: "posts", Action: "manage", Description: "Manage all posts (admin)"},

		// Comments permissions
		{Resource: "comments", Action: "create", Description: "Create comments on posts"},
		{Resource: "comments", Action: "read", Description: "View comments"},
		{Resource: "comments", Action: "update", Description: "Update own comments"},
		{Resource: "comments", Action: "delete", Description: "Delete own comments"},
		{Resource: "comments", Action: "manage", Description: "Manage all comments (moderator)"},
	}
}

// Start starts the module background tasks
func (m *PostsModule) Start(ctx context.Context) error {
	slog.Info("Starting posts module")
	return nil
}

// Stop gracefully stops the module
func (m *PostsModule) Stop(ctx context.Context) error {
	slog.Info("Stopping posts module")
	return nil
}

// HealthCheck checks module health
func (m *PostsModule) HealthCheck(ctx context.Context) error {
	if m.db == nil {
		return module.ErrModuleNotInitialized
	}

	if err := m.db.PingContext(ctx); err != nil {
		return err
	}

	var count int
	err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_posts WHERE deleted_at IS NULL").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}
