package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPostRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "This is test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsFeatured:         false,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	// Verify post was created
	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, post.ID, retrieved.ID)
	assert.Equal(t, post.Title, retrieved.Title)
	assert.Equal(t, post.Slug, retrieved.Slug)
	assert.Equal(t, post.Status, retrieved.Status)
}

func TestUserPostRepository_GetByID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, post.ID, retrieved.ID)

	// Test not found
	_, err = repo.GetByID(ctx, uuidv7.New())
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserPostRepository_GetBySlug(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "unique-test-slug",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	retrieved, err := repo.GetBySlug(ctx, user.ID, "unique-test-slug")
	require.NoError(t, err)
	assert.Equal(t, post.ID, retrieved.ID)
	assert.Equal(t, "unique-test-slug", retrieved.Slug)
}

func TestUserPostRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Original Title",
		Slug:               "original-slug",
		Content:            "Original content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	// Update post
	post.Title = "Updated Title"
	post.Content = "Updated content"
	err = repo.Update(ctx, post)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, "Updated content", retrieved.Content)
}

func TestUserPostRepository_Delete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.Delete(ctx, post.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, post.ID)
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserPostRepository_SoftDelete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.SoftDelete(ctx, post.ID)
	require.NoError(t, err)

	// Should not be found after soft delete
	_, err = repo.GetByID(ctx, post.ID)
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestUserPostRepository_Restore(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.SoftDelete(ctx, post.ID)
	require.NoError(t, err)

	err = repo.Restore(ctx, post.ID)
	require.NoError(t, err)

	// Should be found after restore
	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, post.ID, retrieved.ID)
}

func TestUserPostRepository_GetUserPosts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create multiple posts
	for i := 0; i < 3; i++ {
		post := &entity.UserPost{
			ID:                 uuidv7.New(),
			UserID:             user.ID,
			Title:              fmt.Sprintf("Test Post %d", i+1),
			Slug:               fmt.Sprintf("test-post-%d", i+1),
			Content:            "Test content",
			Status:             entity.PostStatusDraft,
			IsPublic:           true,
			IsCommentsEnabled:  true,
			ReadingTimeMinutes: 5,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		err := repo.Create(ctx, post)
		require.NoError(t, err)
	}

	posts, err := repo.GetUserPosts(ctx, user.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, posts, 3)
}

func TestUserPostRepository_GetPublishedPosts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create published post
	publishedAt := time.Now()
	post1 := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Published Post",
		Slug:               "published-post",
		Content:            "Published content",
		Status:             entity.PostStatusPublished,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		PublishedAt:        &publishedAt,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	err := repo.Create(ctx, post1)
	require.NoError(t, err)

	// Create draft post (should not be returned)
	post2 := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Draft Post",
		Slug:               "draft-post",
		Content:            "Draft content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	err = repo.Create(ctx, post2)
	require.NoError(t, err)

	posts, err := repo.GetPublishedPosts(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Equal(t, entity.PostStatusPublished, posts[0].Status)
}

func TestUserPostRepository_GetFeaturedPosts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create featured post
	publishedAt := time.Now()
	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Featured Post",
		Slug:               "featured-post",
		Content:            "Featured content",
		Status:             entity.PostStatusPublished,
		IsPublic:           true,
		IsFeatured:         true,
		IsCommentsEnabled:  true,
		PublishedAt:        &publishedAt,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	err := repo.Create(ctx, post)
	require.NoError(t, err)

	posts, err := repo.GetFeaturedPosts(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.True(t, posts[0].IsFeatured)
}

func TestUserPostRepository_Search(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	publishedAt := time.Now()
	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Golang Tutorial",
		Slug:               "golang-tutorial",
		Content:            "Learn Go programming",
		Status:             entity.PostStatusPublished,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		PublishedAt:        &publishedAt,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	err := repo.Create(ctx, post)
	require.NoError(t, err)

	posts, err := repo.Search(ctx, "Golang", 10, 0)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Contains(t, posts[0].Title, "Golang")
}

func TestUserPostRepository_GetByTag(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	publishedAt := time.Now()
	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Golang Post",
		Slug:               "golang-post",
		Content:            "Content about Go",
		Status:             entity.PostStatusPublished,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		Tags:               []string{"golang", "programming"},
		PublishedAt:        &publishedAt,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	err := repo.Create(ctx, post)
	require.NoError(t, err)

	posts, err := repo.GetByTag(ctx, "golang", 10, 0)
	require.NoError(t, err)
	assert.Len(t, posts, 1)
}

func TestUserPostRepository_IncrementViews(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		ViewCount:          0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.IncrementViews(ctx, post.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.ViewCount)
}

func TestUserPostRepository_IncrementLikes(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		LikeCount:          0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.IncrementLikes(ctx, post.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.LikeCount)
}

func TestUserPostRepository_DecrementLikes(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		LikeCount:          5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.DecrementLikes(ctx, post.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, retrieved.LikeCount)
}

func TestUserPostRepository_UpdateStatus(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, post.ID, entity.PostStatusPublished)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.PostStatusPublished, retrieved.Status)
}

func TestUserPostRepository_PublishScheduledPost(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	scheduledTime := time.Now().Add(-1 * time.Hour)
	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Scheduled Post",
		Slug:               "scheduled-post",
		Content:            "Test content",
		Status:             entity.PostStatusScheduled,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ScheduledAt:        &scheduledTime,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.PublishScheduledPost(ctx, post.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.PostStatusPublished, retrieved.Status)
	assert.NotNil(t, retrieved.PublishedAt)
	assert.Nil(t, retrieved.ScheduledAt)
}

func TestUserPostRepository_GetScheduledPosts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create scheduled post in the past (ready to publish)
	pastTime := time.Now().Add(-1 * time.Hour)
	post1 := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Past Scheduled Post",
		Slug:               "past-scheduled",
		Content:            "Test content",
		Status:             entity.PostStatusScheduled,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ScheduledAt:        &pastTime,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	err := repo.Create(ctx, post1)
	require.NoError(t, err)

	// Create scheduled post in the future (not ready)
	futureTime := time.Now().Add(24 * time.Hour)
	post2 := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Future Scheduled Post",
		Slug:               "future-scheduled",
		Content:            "Test content",
		Status:             entity.PostStatusScheduled,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ScheduledAt:        &futureTime,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	err = repo.Create(ctx, post2)
	require.NoError(t, err)

	posts, err := repo.GetScheduledPosts(ctx)
	require.NoError(t, err)
	assert.Len(t, posts, 1) // Only past scheduled post
	assert.Equal(t, post1.ID, posts[0].ID)
}

func TestUserPostRepository_IncrementComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CommentCount:       0,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.IncrementComments(ctx, post.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.CommentCount)
}

func TestUserPostRepository_DecrementComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := &entity.UserPost{
		ID:                 uuidv7.New(),
		UserID:             user.ID,
		Title:              "Test Post",
		Slug:               "test-post",
		Content:            "Test content",
		Status:             entity.PostStatusDraft,
		IsPublic:           true,
		IsCommentsEnabled:  true,
		ReadingTimeMinutes: 5,
		CommentCount:       5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := repo.Create(ctx, post)
	require.NoError(t, err)

	err = repo.DecrementComments(ctx, post.ID)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, retrieved.CommentCount)
}
