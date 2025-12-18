package repository

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// PostCommentRepository defines the interface for post comment data operations
type PostCommentRepository interface {
	// Basic CRUD
	Create(ctx context.Context, comment *entity.PostComment) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.PostComment, error)
	Update(ctx context.Context, comment *entity.PostComment) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Soft delete
	SoftDelete(ctx context.Context, id uuidv7.UUID) error
	Restore(ctx context.Context, id uuidv7.UUID) error

	// Listing
	GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error)
	GetCommentReplies(ctx context.Context, parentID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error)
	GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error)
	CountPostComments(ctx context.Context, postID uuidv7.UUID) (int, error)
	CountUserComments(ctx context.Context, userID uuidv7.UUID) (int, error)

	// Engagement
	IncrementLikes(ctx context.Context, id uuidv7.UUID) error
	DecrementLikes(ctx context.Context, id uuidv7.UUID) error
	IncrementReplies(ctx context.Context, id uuidv7.UUID) error
	DecrementReplies(ctx context.Context, id uuidv7.UUID) error

	// Comment likes tracking
	AddLike(ctx context.Context, commentID, userID uuidv7.UUID) error
	RemoveLike(ctx context.Context, commentID, userID uuidv7.UUID) error
	HasUserLiked(ctx context.Context, commentID, userID uuidv7.UUID) (bool, error)
	GetCommentLikers(ctx context.Context, commentID uuidv7.UUID, limit, offset int) ([]uuidv7.UUID, error)
}
