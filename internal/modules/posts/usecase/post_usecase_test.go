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

// TestUserPostUseCase_CreatePost tests the CreatePost use case
//  Uses ONLY module-internal types and mocks - fully independent
func TestUserPostUseCase_CreatePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful post creation", func(t *testing.T) {
		// Arrange - create mock with module's own mock
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		title := "My Test Post"
		content := "This is test content for the post"
		excerpt := "Test excerpt"
		tags := []string{"golang", "testing"}
		categories := []string{"tech"}

		// Mock expects no existing slug
		mockRepo.On("GetBySlug", ctx, userID, mock.AnythingOfType("string")).Return(nil, entity.ErrNotFound)

		// Mock expects Create to be called
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		// Act
		post, err := uc.CreatePost(ctx, userID, title, content, excerpt, tags, categories)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, title, post.Title)
		assert.Equal(t, content, post.Content)
		assert.Equal(t, userID, post.UserID)
		assert.NotEmpty(t, post.Slug)
		assert.Equal(t, entity.PostStatusDraft, post.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("slug already exists", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		existingPost := &entity.UserPost{
			ID:     uuidv7.New(),
			UserID: userID,
			Title:  "Existing Post",
			Slug:   "my-test-post",
		}

		// Mock returns existing post with same slug
		mockRepo.On("GetBySlug", ctx, userID, mock.AnythingOfType("string")).Return(existingPost, nil)

		// Act
		post, err := uc.CreatePost(ctx, userID, "My Test Post", "Content", "", nil, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.Equal(t, ErrSlugAlreadyExists, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty title", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		// Mock для GetBySlug (slug будет пустой строкой из пустого title)
		mockRepo.On("GetBySlug", ctx, userID, mock.AnythingOfType("string")).Return(nil, entity.ErrNotFound)

		// Act
		post, err := uc.CreatePost(ctx, userID, "", "Content", "", nil, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, post)
		// Create не должен быть вызван, так как валидация entity провалится
		mockRepo.AssertNotCalled(t, "Create")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_GetPost tests the GetPost use case
func TestUserPostUseCase_GetPost(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		expectedPost := &entity.UserPost{
			ID:      postID,
			UserID:  uuidv7.New(),
			Title:   "Test Post",
			Content: "Test content",
			Status:  entity.PostStatusPublished,
		}

		mockRepo.On("GetByID", ctx, postID).Return(expectedPost, nil)

		// Act
		post, err := uc.GetPost(ctx, postID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedPost, post)
		mockRepo.AssertExpectations(t)
	})

	t.Run("post not found", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		mockRepo.On("GetByID", ctx, postID).Return(nil, entity.ErrNotFound)

		// Act
		post, err := uc.GetPost(ctx, postID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, post)
		// Проверяем что ошибка обернута, но содержит ErrNotFound
		assert.True(t, errors.Is(err, entity.ErrNotFound))
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_UpdatePost tests the UpdatePost use case
func TestUserPostUseCase_UpdatePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful update", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		existingPost := &entity.UserPost{
			ID:      postID,
			UserID:  userID,
			Title:   "Old Title",
			Content: "Old content",
			Status:  entity.PostStatusDraft,
		}

		updates := map[string]any{
			"title":   "New Title",
			"content": "New content with more details",
		}

		mockRepo.On("GetByID", ctx, postID).Return(existingPost, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		// Act
		post, err := uc.UpdatePost(ctx, userID, postID, updates)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, "New Title", post.Title)
		assert.Equal(t, "New content with more details", post.Content)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		differentUserID := uuidv7.New()
		existingPost := &entity.UserPost{
			ID:     postID,
			UserID: differentUserID, // Different user owns this post
			Title:  "Old Title",
		}

		updates := map[string]any{
			"title": "New Title",
		}

		mockRepo.On("GetByID", ctx, postID).Return(existingPost, nil)

		// Act
		post, err := uc.UpdatePost(ctx, userID, postID, updates)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.Equal(t, ErrUnauthorized, err)
		mockRepo.AssertNotCalled(t, "Update")
		mockRepo.AssertExpectations(t)
	})

	t.Run("post not found", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		updates := map[string]any{
			"title": "New Title",
		}

		mockRepo.On("GetByID", ctx, postID).Return(nil, entity.ErrNotFound)

		// Act
		post, err := uc.UpdatePost(ctx, userID, postID, updates)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, post)
		assert.True(t, errors.Is(err, entity.ErrNotFound))
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_DeletePost tests the DeletePost use case
func TestUserPostUseCase_DeletePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful deletion", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		existingPost := &entity.UserPost{
			ID:     postID,
			UserID: userID,
			Title:  "Post to Delete",
		}

		mockRepo.On("GetByID", ctx, postID).Return(existingPost, nil)
		mockRepo.On("Delete", ctx, postID).Return(nil)

		// Act
		err := uc.DeletePost(ctx, userID, postID)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		differentUserID := uuidv7.New()
		existingPost := &entity.UserPost{
			ID:     postID,
			UserID: differentUserID,
			Title:  "Post to Delete",
		}

		mockRepo.On("GetByID", ctx, postID).Return(existingPost, nil)

		// Act
		err := uc.DeletePost(ctx, userID, postID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		mockRepo.AssertNotCalled(t, "Delete")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_PublishPost tests the PublishPost use case
func TestUserPostUseCase_PublishPost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful publish", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		draftPost := &entity.UserPost{
			ID:      postID,
			UserID:  userID,
			Title:   "Draft Post",
			Content: "Content to publish",
			Status:  entity.PostStatusDraft,
		}

		mockRepo.On("GetByID", ctx, postID).Return(draftPost, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *entity.UserPost) bool {
			return p.Status == entity.PostStatusPublished
		})).Return(nil)

		// Act
		err := uc.PublishPost(ctx, userID, postID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusPublished, draftPost.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		differentUserID := uuidv7.New()
		draftPost := &entity.UserPost{
			ID:     postID,
			UserID: differentUserID,
			Title:  "Draft Post",
			Status: entity.PostStatusDraft,
		}

		mockRepo.On("GetByID", ctx, postID).Return(draftPost, nil)

		// Act
		err := uc.PublishPost(ctx, userID, postID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		mockRepo.AssertNotCalled(t, "Update")
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_UnpublishPost tests the UnpublishPost use case
func TestUserPostUseCase_UnpublishPost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful unpublish", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		publishedPost := &entity.UserPost{
			ID:      postID,
			UserID:  userID,
			Title:   "Published Post",
			Content: "Published content",
			Status:  entity.PostStatusPublished,
		}

		mockRepo.On("GetByID", ctx, postID).Return(publishedPost, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *entity.UserPost) bool {
			return p.Status == entity.PostStatusDraft
		})).Return(nil)

		// Act
		err := uc.UnpublishPost(ctx, userID, postID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusDraft, publishedPost.Status)
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_ArchivePost tests the ArchivePost use case
func TestUserPostUseCase_ArchivePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful archive", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		publishedPost := &entity.UserPost{
			ID:      postID,
			UserID:  userID,
			Title:   "Post to Archive",
			Content: "Content",
			Status:  entity.PostStatusPublished,
		}

		mockRepo.On("GetByID", ctx, postID).Return(publishedPost, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *entity.UserPost) bool {
			return p.Status == entity.PostStatusArchived
		})).Return(nil)

		// Act
		err := uc.ArchivePost(ctx, userID, postID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusArchived, publishedPost.Status)
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_ToggleFeatured tests the ToggleFeatured use case
func TestUserPostUseCase_ToggleFeatured(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("toggle from false to true", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:         postID,
			UserID:     userID,
			Title:      "Regular Post",
			IsFeatured: false,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *entity.UserPost) bool {
			return p.IsFeatured == true
		})).Return(nil)

		// Act
		err := uc.ToggleFeatured(ctx, userID, postID)

		// Assert
		require.NoError(t, err)
		assert.True(t, post.IsFeatured)
		mockRepo.AssertExpectations(t)
	})

	t.Run("toggle from true to false", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:         postID,
			UserID:     userID,
			Title:      "Featured Post",
			IsFeatured: true,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.MatchedBy(func(p *entity.UserPost) bool {
			return p.IsFeatured == false
		})).Return(nil)

		// Act
		err := uc.ToggleFeatured(ctx, userID, postID)

		// Assert
		require.NoError(t, err)
		assert.False(t, post.IsFeatured)
		mockRepo.AssertExpectations(t)
	})
}

// TestUserPostUseCase_SoftDeletePost tests the SoftDeletePost use case
func TestUserPostUseCase_SoftDeletePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful soft delete", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		existingPost := &entity.UserPost{
			ID:     postID,
			UserID: userID,
			Title:  "Post to Soft Delete",
		}

		mockRepo.On("GetByID", ctx, postID).Return(existingPost, nil)
		mockRepo.On("SoftDelete", ctx, postID).Return(nil)

		// Act
		err := uc.SoftDeletePost(ctx, userID, postID)

		// Assert
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		differentUserID := uuidv7.New()
		existingPost := &entity.UserPost{
			ID:     postID,
			UserID: differentUserID,
			Title:  "Post to Soft Delete",
		}

		mockRepo.On("GetByID", ctx, postID).Return(existingPost, nil)

		// Act
		err := uc.SoftDeletePost(ctx, userID, postID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
		mockRepo.AssertNotCalled(t, "SoftDelete")
		mockRepo.AssertExpectations(t)
	})
}
