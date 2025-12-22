package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CommentRepository defines the interface for comment data operations
type CommentRepository interface {
	// Create creates a new comment
	Create(ctx context.Context, comment *entity.Comment) error

	// GetByID retrieves a comment by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Comment, error)

	// Update updates a comment
	Update(ctx context.Context, comment *entity.Comment) error

	// Delete soft-deletes a comment
	Delete(ctx context.Context, id uuidv7.UUID) error

	// GetPostComments retrieves comments for a post with pagination
	GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error)

	// GetReplies retrieves replies to a comment
	GetReplies(ctx context.Context, parentID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error)

	// GetUserComments retrieves comments by a user
	GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error)

	// CountPostComments counts comments for a post
	CountPostComments(ctx context.Context, postID uuidv7.UUID) (int, error)

	// IncrementRepliesCount increments the replies counter for a parent comment
	IncrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error

	// DecrementRepliesCount decrements the replies counter for a parent comment
	DecrementRepliesCount(ctx context.Context, parentID uuidv7.UUID) error
}