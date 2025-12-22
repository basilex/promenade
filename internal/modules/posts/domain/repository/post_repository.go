package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserPostRepository defines the interface for user post persistence
type UserPostRepository interface {
	// Create creates a new post
	Create(ctx context.Context, post *entity.UserPost) error

	// GetByID retrieves a post by its ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserPost, error)

	// GetBySlug retrieves a post by user ID and slug
	GetBySlug(ctx context.Context, userID uuidv7.UUID, slug string) (*entity.UserPost, error)

	// Update updates an existing post
	Update(ctx context.Context, post *entity.UserPost) error

	// Delete performs hard delete of a post
	Delete(ctx context.Context, id uuidv7.UUID) error

	// SoftDelete performs soft delete of a post
	SoftDelete(ctx context.Context, id uuidv7.UUID) error

	// Restore restores a soft-deleted post
	Restore(ctx context.Context, id uuidv7.UUID) error

	// List lists posts with pagination and filters
	List(ctx context.Context, params ListPostsParams) ([]*entity.UserPost, *pagination.Metadata, error)

	// GetUserPosts gets all posts by user ID with pagination
	GetUserPosts(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.UserPost, error)

	// GetPublishedPosts gets all published public posts with pagination
	GetPublishedPosts(ctx context.Context, limit, offset int) ([]*entity.UserPost, error)

	// GetFeaturedPosts gets featured posts
	GetFeaturedPosts(ctx context.Context, limit int) ([]*entity.UserPost, error)

	// Search searches posts by title, content, or tags
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserPost, error)

	// GetByTag gets posts by tag
	GetByTag(ctx context.Context, tag string, limit, offset int) ([]*entity.UserPost, error)

	// GetScheduledPosts gets posts scheduled for publication
	GetScheduledPosts(ctx context.Context) ([]*entity.UserPost, error)

	// IncrementViews increments post view counter
	IncrementViews(ctx context.Context, id uuidv7.UUID) error

	// IncrementLikes increments post like counter
	IncrementLikes(ctx context.Context, id uuidv7.UUID) error

	// DecrementLikes decrements post like counter
	DecrementLikes(ctx context.Context, id uuidv7.UUID) error

	// IncrementComments increments post comment counter
	IncrementComments(ctx context.Context, id uuidv7.UUID) error

	// DecrementComments decrements post comment counter
	DecrementComments(ctx context.Context, id uuidv7.UUID) error

	// IncrementShares increments post share counter
	IncrementShares(ctx context.Context, id uuidv7.UUID) error

	// UpdateStatus updates post status
	UpdateStatus(ctx context.Context, id uuidv7.UUID, status entity.PostStatus) error

	// PublishScheduledPost publishes a scheduled post
	PublishScheduledPost(ctx context.Context, id uuidv7.UUID) error
}

// ListPostsParams contains parameters for listing posts
type ListPostsParams struct {
	UserID      *uuidv7.UUID       // Filter by user
	Status      *entity.PostStatus // Filter by status
	IsPublic    *bool              // Filter by visibility
	IsFeatured  *bool              // Filter by featured status
	Tag         *string            // Filter by tag
	SearchQuery *string            // Search in title/content
	Limit       int                // Pagination limit
	Offset      int                // Pagination offset
	SortBy      string             // Sort field (created_at, published_at, view_count, etc.)
	SortOrder   string             // Sort order (asc, desc)
}
