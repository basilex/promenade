package usecase

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// PostRepository is an alias to avoid import cycle
type PostRepository = repository.IUserPostRepository

// ICommentUseCase defines business logic for comment operations
type ICommentUseCase interface {
	// CreateComment creates a new comment on a post
	CreateComment(ctx context.Context, postID, userID uuidv7.UUID, content string, parentID *uuidv7.UUID) (*entity.Comment, error)

	// GetComment retrieves a comment by ID
	GetComment(ctx context.Context, id uuidv7.UUID) (*entity.Comment, error)

	// UpdateComment updates a comment
	UpdateComment(ctx context.Context, userID, commentID uuidv7.UUID, content string) (*entity.Comment, error)

	// DeleteComment soft-deletes a comment
	DeleteComment(ctx context.Context, userID, commentID uuidv7.UUID) error

	// GetPostComments retrieves comments for a post
	GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error)

	// GetCommentReplies retrieves replies to a comment
	GetCommentReplies(ctx context.Context, commentID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error)

	// GetUserComments retrieves comments by a user
	GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error)
}

type commentUseCase struct {
	commentRepo repository.ICommentRepository
	postRepo    PostRepository
}

// NewCommentUseCase creates a new comment use case
func NewCommentUseCase(commentRepo repository.ICommentRepository, postRepo PostRepository) ICommentUseCase {
	return &commentUseCase{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

func (uc *commentUseCase) CreateComment(ctx context.Context, postID, userID uuidv7.UUID, content string, parentID *uuidv7.UUID) (*entity.Comment, error) {
	// Verify post exists
	_, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}

	// If this is a reply, verify parent comment exists
	depth := 0
	path := ""
	if parentID != nil {
		parentComment, err := uc.commentRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found: %w", err)
		}

		// Check max depth (e.g., 10 levels)
		if parentComment.Depth >= 10 {
			return nil, fmt.Errorf("maximum comment depth exceeded")
		}

		depth = parentComment.Depth + 1
		path = fmt.Sprintf("%s/%s", parentComment.Path, parentComment.ID.String())
	} else {
		path = "/" + postID.String()
	}

	// Create comment
	comment, err := entity.NewComment(postID, userID, content, parentID)
	if err != nil {
		return nil, err
	}

	comment.Depth = depth
	comment.Path = path

	if err := uc.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	// Increment parent comment's replies count
	if parentID != nil {
		if err := uc.commentRepo.IncrementRepliesCount(ctx, *parentID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add logger
		}
	}

	return comment, nil
}

func (uc *commentUseCase) GetComment(ctx context.Context, id uuidv7.UUID) (*entity.Comment, error) {
	return uc.commentRepo.GetByID(ctx, id)
}

func (uc *commentUseCase) UpdateComment(ctx context.Context, userID, commentID uuidv7.UUID, content string) (*entity.Comment, error) {
	// Get existing comment
	comment, err := uc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if comment.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you can only update your own comments")
	}

	// Update content
	comment.Content = content

	// Validate
	if err := comment.Validate(); err != nil {
		return nil, err
	}

	// Save
	if err := uc.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

func (uc *commentUseCase) DeleteComment(ctx context.Context, userID, commentID uuidv7.UUID) error {
	// Get existing comment
	comment, err := uc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Verify ownership
	if comment.UserID != userID {
		return fmt.Errorf("unauthorized: you can only delete your own comments")
	}

	// Soft delete
	if err := uc.commentRepo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Decrement parent comment's replies count
	if comment.ParentID != nil {
		if err := uc.commentRepo.DecrementRepliesCount(ctx, *comment.ParentID); err != nil {
			// Log error but don't fail the operation
			// TODO: Add logger
		}
	}

	return nil
}

func (uc *commentUseCase) GetPostComments(ctx context.Context, postID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	// Verify post exists
	if _, err := uc.postRepo.GetByID(ctx, postID); err != nil {
		return nil, nil, fmt.Errorf("post not found: %w", err)
	}

	return uc.commentRepo.GetPostComments(ctx, postID, limit, offset)
}

func (uc *commentUseCase) GetCommentReplies(ctx context.Context, commentID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	// Verify comment exists
	if _, err := uc.commentRepo.GetByID(ctx, commentID); err != nil {
		return nil, nil, fmt.Errorf("comment not found: %w", err)
	}

	return uc.commentRepo.GetReplies(ctx, commentID, limit, offset)
}

func (uc *commentUseCase) GetUserComments(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Comment, *pagination.Metadata, error) {
	return uc.commentRepo.GetUserComments(ctx, userID, limit, offset)
}
