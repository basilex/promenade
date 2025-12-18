package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/mocks"
)

func TestPostCommentUseCase_CreateComment(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()
	userID := uuidv7.New()
	content := "Great post!"

	t.Run("successful comment creation", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		post := &entity.UserPost{
			ID:                postID,
			UserID:            userID,
			Title:             "Test Post",
			IsCommentsEnabled: true,
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("Create", ctx, mock.AnythingOfType("*entity.PostComment")).Return(nil)
		mockPostRepo.On("IncrementComments", ctx, postID).Return(nil)

		comment, err := uc.CreateComment(ctx, postID, userID, content, nil)
		require.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, postID, comment.PostID)
		assert.Equal(t, userID, comment.UserID)
		assert.Equal(t, content, comment.Content)
		mockPostRepo.AssertExpectations(t)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("post not found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockPostRepo.On("GetByID", ctx, postID).Return(nil, entity.ErrNotFound)

		_, err := uc.CreateComment(ctx, postID, userID, content, nil)
		assert.ErrorIs(t, err, ErrInvalidCommentData)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("comments disabled on post", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		post := &entity.UserPost{
			ID:                postID,
			IsCommentsEnabled: false,
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)

		_, err := uc.CreateComment(ctx, postID, userID, content, nil)
		assert.ErrorIs(t, err, ErrInvalidCommentData)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("successful reply to comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		parentID := uuidv7.New()
		parentComment := &entity.PostComment{
			ID:     parentID,
			PostID: postID,
			UserID: uuidv7.New(),
		}

		post := &entity.UserPost{
			ID:                postID,
			IsCommentsEnabled: true,
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("GetByID", ctx, parentID).Return(parentComment, nil)
		mockCommentRepo.On("Create", ctx, mock.AnythingOfType("*entity.PostComment")).Return(nil)
		mockCommentRepo.On("IncrementReplies", ctx, parentID).Return(nil)
		mockPostRepo.On("IncrementComments", ctx, postID).Return(nil)

		comment, err := uc.CreateComment(ctx, postID, userID, content, &parentID)
		require.NoError(t, err)
		assert.NotNil(t, comment)
		assert.NotNil(t, comment.ParentID)
		assert.Equal(t, parentID, *comment.ParentID)
		mockPostRepo.AssertExpectations(t)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("cannot reply to deleted comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		parentID := uuidv7.New()
		deletedAt := time.Now()
		parentComment := &entity.PostComment{
			ID:        parentID,
			PostID:    postID,
			DeletedAt: &deletedAt,
		}

		post := &entity.UserPost{
			ID:                postID,
			IsCommentsEnabled: true,
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("GetByID", ctx, parentID).Return(parentComment, nil)

		_, err := uc.CreateComment(ctx, postID, userID, content, &parentID)
		assert.ErrorIs(t, err, ErrCannotReplyToDeleted)
		mockPostRepo.AssertExpectations(t)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("parent comment from different post", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		parentID := uuidv7.New()
		parentComment := &entity.PostComment{
			ID:     parentID,
			PostID: uuidv7.New(), // Different post ID
		}

		post := &entity.UserPost{
			ID:                postID,
			IsCommentsEnabled: true,
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("GetByID", ctx, parentID).Return(parentComment, nil)

		_, err := uc.CreateComment(ctx, postID, userID, content, &parentID)
		assert.ErrorIs(t, err, ErrInvalidCommentData)
		mockPostRepo.AssertExpectations(t)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_GetComment(t *testing.T) {
	ctx := context.Background()
	commentID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		expectedComment := &entity.PostComment{
			ID:      commentID,
			Content: "Test comment",
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(expectedComment, nil)

		comment, err := uc.GetComment(ctx, commentID)
		require.NoError(t, err)
		assert.Equal(t, expectedComment, comment)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("comment not found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, entity.ErrNotFound)

		_, err := uc.GetComment(ctx, commentID)
		assert.ErrorIs(t, err, ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_UpdateComment(t *testing.T) {
	ctx := context.Background()
	commentID := uuidv7.New()
	userID := uuidv7.New()
	newContent := "Updated content"

	t.Run("successful update", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID:      commentID,
			UserID:  userID,
			Content: "Old content",
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)
		mockCommentRepo.On("Update", ctx, mock.AnythingOfType("*entity.PostComment")).Return(nil)

		updated, err := uc.UpdateComment(ctx, commentID, userID, newContent)
		require.NoError(t, err)
		assert.Equal(t, newContent, updated.Content)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("comment not found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, entity.ErrNotFound)

		_, err := uc.UpdateComment(ctx, commentID, userID, newContent)
		assert.ErrorIs(t, err, ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID:     commentID,
			UserID: uuidv7.New(), // Different user
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)

		_, err := uc.UpdateComment(ctx, commentID, userID, newContent)
		assert.ErrorIs(t, err, ErrUnauthorizedComment)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("cannot update deleted comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		deletedAt := time.Now()
		comment := &entity.PostComment{
			ID:        commentID,
			UserID:    userID,
			DeletedAt: &deletedAt,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)

		_, err := uc.UpdateComment(ctx, commentID, userID, newContent)
		assert.ErrorIs(t, err, ErrCommentDeleted)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_DeleteComment(t *testing.T) {
	ctx := context.Background()
	commentID := uuidv7.New()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID:     commentID,
			UserID: userID,
			PostID: postID,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)
		mockCommentRepo.On("SoftDelete", ctx, commentID).Return(nil)
		mockPostRepo.On("DecrementComments", ctx, postID).Return(nil)

		err := uc.DeleteComment(ctx, commentID, userID)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
		mockPostRepo.AssertExpectations(t)
	})

	t.Run("comment not found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, entity.ErrNotFound)

		err := uc.DeleteComment(ctx, commentID, userID)
		assert.ErrorIs(t, err, ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID:     commentID,
			UserID: uuidv7.New(), // Different user
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)

		err := uc.DeleteComment(ctx, commentID, userID)
		assert.ErrorIs(t, err, ErrUnauthorizedComment)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("deletion with parent comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		parentID := uuidv7.New()
		comment := &entity.PostComment{
			ID:       commentID,
			UserID:   userID,
			PostID:   postID,
			ParentID: &parentID,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)
		mockCommentRepo.On("SoftDelete", ctx, commentID).Return(nil)
		mockCommentRepo.On("DecrementReplies", ctx, parentID).Return(nil)
		mockPostRepo.On("DecrementComments", ctx, postID).Return(nil)

		err := uc.DeleteComment(ctx, commentID, userID)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
		mockPostRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_GetPostComments(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()

	t.Run("successful list", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comments := []*entity.PostComment{
			{ID: uuidv7.New(), Content: "Comment 1"},
			{ID: uuidv7.New(), Content: "Comment 2"},
		}
		metadata := &pagination.Metadata{Total: 2}

		mockCommentRepo.On("GetPostComments", ctx, postID, 20, 0).Return(comments, metadata, nil)

		result, meta, err := uc.GetPostComments(ctx, postID, 20, 0)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, meta.Total)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("default pagination values", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetPostComments", ctx, postID, 20, 0).Return([]*entity.PostComment{}, &pagination.Metadata{}, nil)

		_, _, err := uc.GetPostComments(ctx, postID, 0, 0)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("max limit enforcement", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetPostComments", ctx, postID, 100, 0).Return([]*entity.PostComment{}, &pagination.Metadata{}, nil)

		_, _, err := uc.GetPostComments(ctx, postID, 150, 0)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_GetCommentReplies(t *testing.T) {
	ctx := context.Background()
	commentID := uuidv7.New()

	t.Run("successful list", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		parentComment := &entity.PostComment{
			ID: commentID,
		}

		replies := []*entity.PostComment{
			{ID: uuidv7.New(), ParentID: &commentID},
		}
		metadata := &pagination.Metadata{Total: 1}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(parentComment, nil)
		mockCommentRepo.On("GetCommentReplies", ctx, commentID, 20, 0).Return(replies, metadata, nil)

		result, meta, err := uc.GetCommentReplies(ctx, commentID, 20, 0)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 1, meta.Total)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("parent comment not found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, entity.ErrNotFound)

		_, _, err := uc.GetCommentReplies(ctx, commentID, 20, 0)
		assert.ErrorIs(t, err, ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_GetUserComments(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful list", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comments := []*entity.PostComment{
			{ID: uuidv7.New(), UserID: userID},
			{ID: uuidv7.New(), UserID: userID},
		}
		metadata := &pagination.Metadata{Total: 2}

		mockCommentRepo.On("GetUserComments", ctx, userID, 20, 0).Return(comments, metadata, nil)

		result, meta, err := uc.GetUserComments(ctx, userID, 20, 0)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 2, meta.Total)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_LikeComment(t *testing.T) {
	ctx := context.Background()
	commentID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("successful like", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID: commentID,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)
		mockCommentRepo.On("HasUserLiked", ctx, commentID, userID).Return(false, nil)
		mockCommentRepo.On("AddLike", ctx, commentID, userID).Return(nil)

		err := uc.LikeComment(ctx, commentID, userID)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("already liked - idempotent", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID: commentID,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)
		mockCommentRepo.On("HasUserLiked", ctx, commentID, userID).Return(true, nil)

		err := uc.LikeComment(ctx, commentID, userID)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("comment not found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, entity.ErrNotFound)

		err := uc.LikeComment(ctx, commentID, userID)
		assert.ErrorIs(t, err, ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("cannot like deleted comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		deletedAt := time.Now()
		comment := &entity.PostComment{
			ID:        commentID,
			DeletedAt: &deletedAt,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)

		err := uc.LikeComment(ctx, commentID, userID)
		assert.ErrorIs(t, err, ErrCommentDeleted)
		mockCommentRepo.AssertExpectations(t)
	})
}

func TestPostCommentUseCase_UnlikeComment(t *testing.T) {
	ctx := context.Background()
	commentID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("successful unlike", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID: commentID,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)
		mockCommentRepo.On("HasUserLiked", ctx, commentID, userID).Return(true, nil)
		mockCommentRepo.On("RemoveLike", ctx, commentID, userID).Return(nil)

		err := uc.UnlikeComment(ctx, commentID, userID)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("not liked - idempotent", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		comment := &entity.PostComment{
			ID: commentID,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)
		mockCommentRepo.On("HasUserLiked", ctx, commentID, userID).Return(false, nil)

		err := uc.UnlikeComment(ctx, commentID, userID)
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("comment not found", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, entity.ErrNotFound)

		err := uc.UnlikeComment(ctx, commentID, userID)
		assert.ErrorIs(t, err, ErrCommentNotFound)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("cannot unlike deleted comment", func(t *testing.T) {
		mockCommentRepo := new(mocks.MockPostCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewPostCommentUseCase(mockCommentRepo, mockPostRepo)

		deletedAt := time.Now()
		comment := &entity.PostComment{
			ID:        commentID,
			DeletedAt: &deletedAt,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(comment, nil)

		err := uc.UnlikeComment(ctx, commentID, userID)
		assert.ErrorIs(t, err, ErrCommentDeleted)
		mockCommentRepo.AssertExpectations(t)
	})
}
