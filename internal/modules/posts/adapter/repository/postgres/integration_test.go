//go:build integration

package postgres_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/posts/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/posts/domain/entity"
	postrepo "github.com/basilex/promenade/internal/modules/posts/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestPostRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	repo := postgres.NewUserPostRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user first
	user := fixtures.CreateUser(t, "post-author@example.com", "password123")

	t.Run("Create and GetByID", func(t *testing.T) {
		post := &entity.UserPost{
			ID:        uuidv7.New(),
			UserID:    user.ID,
			Title:     "Test Post",
			Slug:      "test-post",
			Content:   "This is test content",
			Status:    entity.PostStatusPublished,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.Create(ctx, post)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, post.Title, retrieved.Title)
		assert.Equal(t, post.Content, retrieved.Content)
		assert.Equal(t, post.Status, retrieved.Status)
	})

	t.Run("Update", func(t *testing.T) {
		post := &entity.UserPost{
			ID:        uuidv7.New(),
			UserID:    user.ID,
			Title:     "Original Title",
			Slug:      "original-title",
			Content:   "Original content",
			Status:    entity.PostStatusDraft,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, post))

		post.Title = "Updated Title"
		post.Content = "Updated content"
		post.Status = entity.PostStatusPublished
		err := repo.Update(ctx, post)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", retrieved.Title)
		assert.Equal(t, "Updated content", retrieved.Content)
		assert.Equal(t, entity.PostStatusPublished, retrieved.Status)
	})

	t.Run("Delete (soft delete)", func(t *testing.T) {
		post := &entity.UserPost{
			ID:        uuidv7.New(),
			UserID:    user.ID,
			Title:     "To Delete",
			Slug:      "to-delete",
			Content:   "Will be deleted",
			Status:    entity.PostStatusPublished,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, post))

		err := repo.Delete(ctx, post.ID)
		require.NoError(t, err)

		// Should not be found after soft delete
		_, err = repo.GetByID(ctx, post.ID)
		assert.Error(t, err)
	})

	t.Run("ListByUserID", func(t *testing.T) {
		// Create multiple posts for user
		for i := 0; i < 3; i++ {
			post := &entity.UserPost{
				ID:        uuidv7.New(),
				UserID:    user.ID,
				Title:     fmt.Sprintf("User Post %d %d", i, time.Now().UnixNano()),
				Slug:      fmt.Sprintf("user-post-%d-%d", i, time.Now().UnixNano()),
				Content:   "Content",
				Status:    entity.PostStatusPublished,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			require.NoError(t, repo.Create(ctx, post))
		}

		posts, err := repo.GetUserPosts(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 3)
	})

	t.Run("List by status", func(t *testing.T) {
		// Create posts with different statuses
		draft := &entity.UserPost{
			ID:        uuidv7.New(),
			UserID:    user.ID,
			Title:     "Draft Post",
			Slug:      "draft-post",
			Content:   "Draft content",
			Status:    entity.PostStatusDraft,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, repo.Create(ctx, draft))

		draftStatus := entity.PostStatusDraft
		params := postrepo.ListPostsParams{
			Status: &draftStatus,
			Limit:  10,
			Offset: 0,
		}
		posts, metadata, err := repo.List(ctx, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 1)
		assert.GreaterOrEqual(t, metadata.Total, 1)

		// All posts should have draft status
		for _, p := range posts {
			assert.Equal(t, entity.PostStatusDraft, p.Status)
		}
	})

	t.Run("Count user posts", func(t *testing.T) {
		posts, err := repo.GetUserPosts(ctx, user.ID, 100, 0)
		require.NoError(t, err)
		assert.Greater(t, len(posts), 0)
	})
}

func TestCommentRepository_Integration(t *testing.T) {
	testDB := integration.SetupTestDBWithCleanTables(t)
	fixtures := integration.NewFixtures(testDB.DB)

	postRepo := postgres.NewUserPostRepository(testDB.DB)
	commentRepo := postgres.NewCommentRepository(testDB.DB)
	ctx := testDB.GetContext()

	// Create test user and post
	user := fixtures.CreateUser(t, "comment-author@example.com", "password123")
	post := &entity.UserPost{
		ID:        uuidv7.New(),
		UserID:    user.ID,
		Title:     "Post for Comments",
		Content:   "Content",
		Status:    entity.PostStatusPublished,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, postRepo.Create(ctx, post))

	t.Run("Create and GetByID", func(t *testing.T) {
		comment := &entity.Comment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   "This is a test comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := commentRepo.Create(ctx, comment)
		require.NoError(t, err)

		retrieved, err := commentRepo.GetByID(ctx, comment.ID)
		require.NoError(t, err)
		assert.Equal(t, comment.Content, retrieved.Content)
		assert.Equal(t, comment.PostID, retrieved.PostID)
	})

	t.Run("Create reply comment", func(t *testing.T) {
		parent := &entity.Comment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   "Parent comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, commentRepo.Create(ctx, parent))

		reply := &entity.Comment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			ParentID:  &parent.ID,
			Content:   "Reply to parent",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := commentRepo.Create(ctx, reply)
		require.NoError(t, err)

		retrieved, err := commentRepo.GetByID(ctx, reply.ID)
		require.NoError(t, err)
		assert.Equal(t, parent.ID, *retrieved.ParentID)
	})

	t.Run("ListByPostID", func(t *testing.T) {
		// Create multiple comments
		for i := 0; i < 3; i++ {
			comment := &entity.Comment{
				ID:        uuidv7.New(),
				PostID:    post.ID,
				UserID:    user.ID,
				Content:   "Comment " + uuidv7.New().String(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			require.NoError(t, commentRepo.Create(ctx, comment))
		}

		comments, metadata, err := commentRepo.GetPostComments(ctx, post.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(comments), 3)
		assert.GreaterOrEqual(t, metadata.Total, 3)
	})

	t.Run("GetUserComments", func(t *testing.T) {
		comments, metadata, err := commentRepo.GetUserComments(ctx, user.ID, 10, 0)
		require.NoError(t, err)
		assert.Greater(t, len(comments), 0)
		assert.Greater(t, metadata.Total, 0)
	})

	t.Run("CountPostComments", func(t *testing.T) {
		count, err := commentRepo.CountPostComments(ctx, post.ID)
		require.NoError(t, err)
		assert.Greater(t, count, 0)
	})

	t.Run("Delete (soft delete)", func(t *testing.T) {
		comment := &entity.Comment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   "To be deleted",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		require.NoError(t, commentRepo.Create(ctx, comment))

		err := commentRepo.Delete(ctx, comment.ID)
		require.NoError(t, err)

		// Should not be found after soft delete
		_, err = commentRepo.GetByID(ctx, comment.ID)
		assert.Error(t, err)
	})
}
