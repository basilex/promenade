package usecase

import (
	"context"
	"errors"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var (
	ErrCommentNotFound      = errors.New("comment not found")
	ErrUnauthorizedComment  = errors.New("unauthorized to modify this comment")
	ErrInvalidCommentData   = errors.New("invalid comment data")
	ErrCommentDeleted       = errors.New("comment is deleted")
	ErrCannotReplyToDeleted = errors.New("cannot reply to deleted comment")
)

// PostCommentUseCase defines the interface for post comment business logic
type PostCommentUseCase interface {
	// Comment operations
	CreateComment(ctx context.Context, postID, userID uuidv7.UUID, content string, parentID *uuidv7.UUID) (*entity.PostComment, error)
	GetComment(ctx context.Context, id uuidv7.UUID) (*entity.PostComment, error)
	UpdateComment(ctx context.Context, id, userID uuidv7.UUID, content string) (*entity.PostComment, error)
	DeleteComment(ctx context.Context, id, userID uuidv7.UUID) error

	// Listing
	GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error)
	GetCommentReplies(ctx context.Context, commentID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error)
	GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error)

	// Engagement
	LikeComment(ctx context.Context, commentID, userID uuidv7.UUID) error
	UnlikeComment(ctx context.Context, commentID, userID uuidv7.UUID) error
}

type postCommentUseCase struct {
	commentRepo repository.PostCommentRepository
	postRepo    repository.UserPostRepository
}

// NewPostCommentUseCase creates a new post comment use case
func NewPostCommentUseCase(commentRepo repository.PostCommentRepository, postRepo repository.UserPostRepository) PostCommentUseCase {
	return &postCommentUseCase{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

func (uc *postCommentUseCase) CreateComment(ctx context.Context, postID, userID uuidv7.UUID, content string, parentID *uuidv7.UUID) (*entity.PostComment, error) {
	// Validate post exists
	post, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrInvalidCommentData
		}
		return nil, err
	}

	// Check if comments are enabled
	if !post.IsCommentsEnabled {
		return nil, ErrInvalidCommentData
	}

	// If replying to a comment, validate parent exists and is not deleted
	if parentID != nil {
		parent, err := uc.commentRepo.GetByID(ctx, *parentID)
		if err != nil {
			if errors.Is(err, entity.ErrNotFound) {
				return nil, ErrInvalidCommentData
			}
			return nil, err
		}

		if parent.IsDeleted() {
			return nil, ErrCannotReplyToDeleted
		}

		// Ensure parent belongs to the same post
		if parent.PostID != postID {
			return nil, ErrInvalidCommentData
		}
	}

	// Create comment
	comment, err := entity.NewPostComment(postID, userID, content, parentID)
	if err != nil {
		return nil, ErrInvalidCommentData
	}

	if err := uc.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Increment reply count on parent if it's a reply
	if parentID != nil {
		_ = uc.commentRepo.IncrementReplies(ctx, *parentID)
	}

	// Increment post comment count
	_ = uc.postRepo.IncrementComments(ctx, postID)

	return comment, nil
}

func (uc *postCommentUseCase) GetComment(ctx context.Context, id uuidv7.UUID) (*entity.PostComment, error) {
	comment, err := uc.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	return comment, nil
}

func (uc *postCommentUseCase) UpdateComment(ctx context.Context, id, userID uuidv7.UUID, content string) (*entity.PostComment, error) {
	// Get comment
	comment, err := uc.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	// Check ownership
	if comment.UserID != userID {
		return nil, ErrUnauthorizedComment
	}

	// Check if deleted
	if comment.IsDeleted() {
		return nil, ErrCommentDeleted
	}

	// Update content
	if err := comment.UpdateContent(content); err != nil {
		return nil, ErrInvalidCommentData
	}

	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (uc *postCommentUseCase) DeleteComment(ctx context.Context, id, userID uuidv7.UUID) error {
	// Get comment
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

func (uc *postCommentUseCase) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return uc.commentRepo.GetPostComments(ctx, postID, limit, offset)
}

func (uc *postCommentUseCase) GetCommentReplies(ctx context.Context, commentID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	// Verify comment exists
	_, err := uc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, ErrCommentNotFound
		}
		return nil, nil, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return uc.commentRepo.GetCommentReplies(ctx, commentID, limit, offset)
}

func (uc *postCommentUseCase) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.PostComment, *pagination.Metadata, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return uc.commentRepo.GetUserComments(ctx, userID, limit, offset)
}

func (uc *postCommentUseCase) LikeComment(ctx context.Context, commentID, userID uuidv7.UUID) error {
	// Verify comment exists and is not deleted
	comment, err := uc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.IsDeleted() {
		return ErrCommentDeleted
	}

	// TODO: Check if user already liked (requires comment_likes table)
	return uc.commentRepo.IncrementLikes(ctx, commentID)
}

func (uc *postCommentUseCase) UnlikeComment(ctx context.Context, commentID, userID uuidv7.UUID) error {
	// Verify comment exists and not deleted
	comment, err := uc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.IsDeleted() {
		return ErrCommentDeleted
	}

	// TODO: Check if user actually liked (requires comment_likes table)
	return uc.commentRepo.DecrementLikes(ctx, commentID)
}
