package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	"github.com/basilex/promenade/internal/modules/posts/domain/repository/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TestCommentUseCase_CreateComment tests comment creation
//  Module-independent: uses only module types and mocks
func TestCommentUseCase_CreateComment(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("successful top-level comment", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		post := &entity.UserPost{
			ID:      postID,
			UserID:  uuidv7.New(),
			Title:   "Test Post",
			Content: "Content",
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("Create", ctx, mock.AnythingOfType("*entity.Comment")).Return(nil)

		// Act
		comment, err := uc.CreateComment(ctx, postID, userID, "Great post!", nil)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, "Great post!", comment.Content)
		assert.Equal(t, postID, comment.PostID)
		assert.Equal(t, userID, comment.UserID)
		assert.Nil(t, comment.ParentID)
		assert.Equal(t, 0, comment.Depth)

		mockPostRepo.AssertExpectations(t)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("successful reply comment", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		parentCommentID := uuidv7.New()
		parentComment := &entity.Comment{
			ID:      parentCommentID,
			PostID:  postID,
			UserID:  uuidv7.New(),
			Content: "Parent comment",
			Depth:   0,
			Path:    "/" + postID.String(),
		}

		post := &entity.UserPost{
			ID:      postID,
			UserID:  uuidv7.New(),
			Title:   "Test Post",
			Content: "Content",
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("GetByID", ctx, parentCommentID).Return(parentComment, nil)
		mockCommentRepo.On("Create", ctx, mock.AnythingOfType("*entity.Comment")).Return(nil)
		mockCommentRepo.On("IncrementRepliesCount", ctx, parentCommentID).Return(nil)

		// Act
		comment, err := uc.CreateComment(ctx, postID, userID, "Good point!", &parentCommentID)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, "Good point!", comment.Content)
		assert.Equal(t, &parentCommentID, comment.ParentID)
		assert.Equal(t, 1, comment.Depth)

		mockPostRepo.AssertExpectations(t)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("post not found", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		mockPostRepo.On("GetByID", ctx, postID).Return(nil, entity.ErrNotFound)

		// Act
		comment, err := uc.CreateComment(ctx, postID, userID, "Comment", nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.True(t, errors.Is(err, entity.ErrNotFound))

		mockPostRepo.AssertExpectations(t)
		mockCommentRepo.AssertNotCalled(t, "Create")
	})

	t.Run("parent comment not found", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		parentCommentID := uuidv7.New()
		post := &entity.UserPost{
			ID:      postID,
			UserID:  uuidv7.New(),
			Title:   "Test Post",
			Content: "Content",
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("GetByID", ctx, parentCommentID).Return(nil, entity.ErrNotFound)

		// Act
		comment, err := uc.CreateComment(ctx, postID, userID, "Reply", &parentCommentID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.True(t, errors.Is(err, entity.ErrNotFound))

		mockCommentRepo.AssertNotCalled(t, "Create")
	})

	t.Run("max depth exceeded", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		parentCommentID := uuidv7.New()
		parentComment := &entity.Comment{
			ID:      parentCommentID,
			PostID:  postID,
			UserID:  uuidv7.New(),
			Content: "Deep comment",
			Depth:   10, // Already at max depth
		}

		post := &entity.UserPost{
			ID:      postID,
			UserID:  uuidv7.New(),
			Title:   "Test Post",
			Content: "Content",
		}

		mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockCommentRepo.On("GetByID", ctx, parentCommentID).Return(parentComment, nil)

		// Act
		comment, err := uc.CreateComment(ctx, postID, userID, "Too deep", &parentCommentID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Contains(t, err.Error(), "maximum comment depth exceeded")

		mockCommentRepo.AssertNotCalled(t, "Create")
	})
}

// TestCommentUseCase_GetComment tests comment retrieval
func TestCommentUseCase_GetComment(t *testing.T) {
	ctx := context.Background()
	commentID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		expectedComment := &entity.Comment{
			ID:      commentID,
			PostID:  uuidv7.New(),
			UserID:  uuidv7.New(),
			Content: "Test comment",
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(expectedComment, nil)

		// Act
		comment, err := uc.GetComment(ctx, commentID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedComment, comment)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("comment not found", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		mockCommentRepo.On("GetByID", ctx, commentID).Return(nil, entity.ErrNotFound)

		// Act
		comment, err := uc.GetComment(ctx, commentID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, comment)
		mockCommentRepo.AssertExpectations(t)
	})
}

// TestCommentUseCase_UpdateComment tests comment updates
func TestCommentUseCase_UpdateComment(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	commentID := uuidv7.New()

	t.Run("successful update", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		existingComment := &entity.Comment{
			ID:      commentID,
			PostID:  uuidv7.New(),
			UserID:  userID,
			Content: "Old content",
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(existingComment, nil)
		mockCommentRepo.On("Update", ctx, mock.AnythingOfType("*entity.Comment")).Return(nil)

		// Act
		comment, err := uc.UpdateComment(ctx, userID, commentID, "New content")

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, comment)
		assert.Equal(t, "New content", comment.Content)

		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		differentUserID := uuidv7.New()
		existingComment := &entity.Comment{
			ID:      commentID,
			PostID:  uuidv7.New(),
			UserID:  differentUserID,
			Content: "Comment",
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(existingComment, nil)

		// Act
		comment, err := uc.UpdateComment(ctx, userID, commentID, "New content")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, comment)
		assert.Contains(t, err.Error(), "unauthorized")

		mockCommentRepo.AssertNotCalled(t, "Update")
	})
}

// TestCommentUseCase_DeleteComment tests comment deletion
func TestCommentUseCase_DeleteComment(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	commentID := uuidv7.New()

	t.Run("successful deletion - top-level comment", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		existingComment := &entity.Comment{
			ID:       commentID,
			PostID:   uuidv7.New(),
			UserID:   userID,
			Content:  "Comment to delete",
			ParentID: nil,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(existingComment, nil)
		mockCommentRepo.On("Delete", ctx, commentID).Return(nil)

		// Act
		err := uc.DeleteComment(ctx, userID, commentID)

		// Assert
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("successful deletion - reply comment", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		parentID := uuidv7.New()
		existingComment := &entity.Comment{
			ID:       commentID,
			PostID:   uuidv7.New(),
			UserID:   userID,
			Content:  "Reply to delete",
			ParentID: &parentID,
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(existingComment, nil)
		mockCommentRepo.On("Delete", ctx, commentID).Return(nil)
		mockCommentRepo.On("DecrementRepliesCount", ctx, parentID).Return(nil)

		// Act
		err := uc.DeleteComment(ctx, userID, commentID)

		// Assert
		require.NoError(t, err)
		mockCommentRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		// Arrange
		mockCommentRepo := new(mocks.MockCommentRepository)
		mockPostRepo := new(mocks.MockUserPostRepository)
		uc := NewCommentUseCase(mockCommentRepo, mockPostRepo)

		differentUserID := uuidv7.New()
		existingComment := &entity.Comment{
			ID:      commentID,
			PostID:  uuidv7.New(),
			UserID:  differentUserID,
			Content: "Comment",
		}

		mockCommentRepo.On("GetByID", ctx, commentID).Return(existingComment, nil)

		// Act
		err := uc.DeleteComment(ctx, userID, commentID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")

		mockCommentRepo.AssertNotCalled(t, "Delete")
	})
}
