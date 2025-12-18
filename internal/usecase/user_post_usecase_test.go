package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/mocks"
)

func TestUserPostUseCase_CreatePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful post creation", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		title := "My First Post"
		content := "This is the content of my post with enough words to calculate reading time properly."
		excerpt := "Short excerpt"
		tags := []string{"tech", "golang"}
		categories := []string{"programming"}

		mockRepo.On("GetBySlug", ctx, userID, "my-first-post").Return(nil, entity.ErrNotFound)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		post, err := uc.CreatePost(ctx, userID, title, content, excerpt, tags, categories)
		require.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, title, post.Title)
		assert.Equal(t, "my-first-post", post.Slug)
		assert.Equal(t, userID, post.UserID)
		assert.Equal(t, tags, post.Tags)
		assert.Equal(t, categories, post.Categories)
		assert.Greater(t, post.ReadingTimeMinutes, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("slug already exists", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		existingPost := &entity.UserPost{
			ID:     uuidv7.New(),
			UserID: userID,
			Title:  "Duplicate",
			Slug:   "my-post",
		}

		mockRepo.On("GetBySlug", ctx, userID, "my-post").Return(existingPost, nil)

		_, err := uc.CreatePost(ctx, userID, "My Post", "Content", "", nil, nil)
		assert.ErrorIs(t, err, ErrSlugAlreadyExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty title validation", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		// Mock GetBySlug as it's called before validation
		mockRepo.On("GetBySlug", ctx, userID, "").Return(nil, entity.ErrNotFound)

		_, err := uc.CreatePost(ctx, userID, "", "Content", "", nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})
}

func TestUserPostUseCase_GetPost(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		expectedPost := &entity.UserPost{
			ID:    postID,
			Title: "Test Post",
		}

		mockRepo.On("GetByID", ctx, postID).Return(expectedPost, nil)

		post, err := uc.GetPost(ctx, postID)
		require.NoError(t, err)
		assert.Equal(t, expectedPost, post)
		mockRepo.AssertExpectations(t)
	})

	t.Run("post not found", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		mockRepo.On("GetByID", ctx, postID).Return(nil, entity.ErrNotFound)

		_, err := uc.GetPost(ctx, postID)
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_GetPostBySlug(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	slug := "my-post"

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		expectedPost := &entity.UserPost{
			ID:     uuidv7.New(),
			UserID: userID,
			Slug:   slug,
		}

		mockRepo.On("GetBySlug", ctx, userID, slug).Return(expectedPost, nil)

		post, err := uc.GetPostBySlug(ctx, userID, slug)
		require.NoError(t, err)
		assert.Equal(t, expectedPost, post)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_UpdatePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:      postID,
			UserID:  userID,
			Title:   "Old Title",
			Slug:    "old-title",
			Content: "Old content",
			Status:  entity.PostStatusDraft,
		}

		updates := map[string]interface{}{
			"title":   "New Title",
			"content": "New content with enough words for reading time calculation.",
			"tags":    []string{"newtag"},
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		updated, err := uc.UpdatePost(ctx, userID, postID, updates)
		require.NoError(t, err)
		assert.Equal(t, "New Title", updated.Title)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: uuidv7.New(), // Different user
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)

		_, err := uc.UpdatePost(ctx, userID, postID, map[string]interface{}{})
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_DeletePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: userID,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Delete", ctx, postID).Return(nil)

		err := uc.DeletePost(ctx, userID, postID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: uuidv7.New(),
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)

		err := uc.DeletePost(ctx, userID, postID)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_SoftDeletePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful soft deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: userID,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("SoftDelete", ctx, postID).Return(nil)

		err := uc.SoftDeletePost(ctx, userID, postID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_RestorePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful restoration", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		mockRepo.On("Restore", ctx, postID).Return(nil)

		err := uc.RestorePost(ctx, userID, postID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_PublishPost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful publish", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: userID,
			Status: entity.PostStatusDraft,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		err := uc.PublishPost(ctx, userID, postID)
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusPublished, post.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: uuidv7.New(),
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)

		err := uc.PublishPost(ctx, userID, postID)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_UnpublishPost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful unpublish", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: userID,
			Status: entity.PostStatusPublished,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		err := uc.UnpublishPost(ctx, userID, postID)
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusDraft, post.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_ArchivePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful archive", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: userID,
			Status: entity.PostStatusPublished,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		err := uc.ArchivePost(ctx, userID, postID)
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusArchived, post.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_SchedulePost(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()
	scheduledAt := time.Now().Add(24 * time.Hour)

	t.Run("successful schedule", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: userID,
			Status: entity.PostStatusDraft,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		err := uc.SchedulePost(ctx, userID, postID, scheduledAt)
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusScheduled, post.Status)
		assert.NotNil(t, post.ScheduledAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized user", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:     postID,
			UserID: uuidv7.New(),
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)

		err := uc.SchedulePost(ctx, userID, postID, scheduledAt)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_ToggleFeatured(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful toggle", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:         postID,
			UserID:     userID,
			IsFeatured: false,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		err := uc.ToggleFeatured(ctx, userID, postID)
		require.NoError(t, err)
		assert.True(t, post.IsFeatured)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_ToggleComments(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	postID := uuidv7.New()

	t.Run("successful toggle", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		post := &entity.UserPost{
			ID:                postID,
			UserID:            userID,
			IsCommentsEnabled: true,
		}

		mockRepo.On("GetByID", ctx, postID).Return(post, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserPost")).Return(nil)

		err := uc.ToggleComments(ctx, userID, postID)
		require.NoError(t, err)
		assert.False(t, post.IsCommentsEnabled)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_ViewPost(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()

	t.Run("successful view increment", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		mockRepo.On("IncrementViews", ctx, postID).Return(nil)

		err := uc.ViewPost(ctx, postID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_LikePost(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()

	t.Run("successful like", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		mockRepo.On("IncrementLikes", ctx, postID).Return(nil)

		err := uc.LikePost(ctx, postID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_UnlikePost(t *testing.T) {
	ctx := context.Background()
	postID := uuidv7.New()

	t.Run("successful unlike", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		mockRepo.On("DecrementLikes", ctx, postID).Return(nil)

		err := uc.UnlikePost(ctx, postID)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_GetUserPosts(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		posts := []*entity.UserPost{
			{ID: uuidv7.New(), UserID: userID},
			{ID: uuidv7.New(), UserID: userID},
		}

		mockRepo.On("GetUserPosts", ctx, userID, 20, 0).Return(posts, nil)

		result, err := uc.GetUserPosts(ctx, userID, 20, 0)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_GetPublishedPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		posts := []*entity.UserPost{
			{ID: uuidv7.New(), Status: entity.PostStatusPublished},
		}

		mockRepo.On("GetPublishedPosts", ctx, 20, 0).Return(posts, nil)

		result, err := uc.GetPublishedPosts(ctx, 20, 0)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_GetFeaturedPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		posts := []*entity.UserPost{
			{ID: uuidv7.New(), IsFeatured: true},
		}

		mockRepo.On("GetFeaturedPosts", ctx, 10).Return(posts, nil)

		result, err := uc.GetFeaturedPosts(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_SearchPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("successful search", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		posts := []*entity.UserPost{
			{ID: uuidv7.New(), Title: "Golang Tutorial"},
		}

		mockRepo.On("Search", ctx, "golang", 20, 0).Return(posts, nil)

		result, err := uc.SearchPosts(ctx, "golang", 20, 0)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_GetPostsByTag(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		posts := []*entity.UserPost{
			{ID: uuidv7.New(), Tags: []string{"golang"}},
		}

		mockRepo.On("GetByTag", ctx, "golang", 20, 0).Return(posts, nil)

		result, err := uc.GetPostsByTag(ctx, "golang", 20, 0)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_ListPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list with metadata", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		posts := []*entity.UserPost{
			{ID: uuidv7.New()},
		}
		metadata := &pagination.Metadata{Total: 1}

		params := repository.ListPostsParams{
			Limit:  20,
			Offset: 0,
		}

		mockRepo.On("List", ctx, params).Return(posts, metadata, nil)

		result, meta, err := uc.ListPosts(ctx, params)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 1, meta.Total)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserPostUseCase_ProcessScheduledPosts(t *testing.T) {
	ctx := context.Background()

	t.Run("successful processing", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		pastTime := time.Now().Add(-1 * time.Hour)
		posts := []*entity.UserPost{
			{
				ID:          uuidv7.New(),
				Status:      entity.PostStatusScheduled,
				ScheduledAt: &pastTime,
			},
		}

		mockRepo.On("GetScheduledPosts", ctx).Return(posts, nil)
		mockRepo.On("PublishScheduledPost", ctx, posts[0].ID).Return(nil)

		err := uc.ProcessScheduledPosts(ctx)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no scheduled posts", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		mockRepo.On("GetScheduledPosts", ctx).Return([]*entity.UserPost{}, nil)

		err := uc.ProcessScheduledPosts(ctx)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("continue on individual post error", func(t *testing.T) {
		mockRepo := new(mocks.MockUserPostRepository)
		uc := NewUserPostUseCase(mockRepo)

		pastTime := time.Now().Add(-1 * time.Hour)
		posts := []*entity.UserPost{
			{
				ID:          uuidv7.New(),
				Status:      entity.PostStatusScheduled,
				ScheduledAt: &pastTime,
			},
			{
				ID:          uuidv7.New(),
				Status:      entity.PostStatusScheduled,
				ScheduledAt: &pastTime,
			},
		}

		mockRepo.On("GetScheduledPosts", ctx).Return(posts, nil)
		mockRepo.On("PublishScheduledPost", ctx, posts[0].ID).Return(errors.New("publish error"))
		mockRepo.On("PublishScheduledPost", ctx, posts[1].ID).Return(nil)

		err := uc.ProcessScheduledPosts(ctx)
		require.NoError(t, err) // Should not fail, just log errors
		mockRepo.AssertExpectations(t)
	})
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{"simple title", "My First Post", "my-first-post"},
		{"with special chars", "Hello, World!", "hello-world"},
		{"multiple spaces", "This  Has   Spaces", "this-has-spaces"},
		{"trailing hyphens", "---Title---", "title"},
		{"long title", "This is a very long title that should be truncated to exactly one hundred characters in total length for slug generation", "this-is-a-very-long-title-that-should-be-truncated-to-exactly-one-hundred-characters-in-total-length"},
		{"numbers", "Post 123", "post-123"},
		{"mixed", "Go 1.21 Release: What's New?", "go-121-release-whats-new"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSlug(tt.title)
			assert.Equal(t, tt.expected, result)
			assert.LessOrEqual(t, len(result), 100)
		})
	}
}

func TestCalculateReadingTime(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"empty", "", 1}, // Minimum 1 minute
		{"short", "This is a short post.", 1},
		{"medium", strings.Repeat("word ", 225), 1},     // Exactly 225 words = 1 min
		{"long", strings.Repeat("word ", 450), 2},       // 450 words = 2 min
		{"very long", strings.Repeat("word ", 1125), 5}, // 1125 words = 5 min
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateReadingTime(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}
