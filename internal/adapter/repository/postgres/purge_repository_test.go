package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// insertPostWithDeletedAt inserts a post directly with deleted_at (bypassing Create which doesn't support it)
func insertPostWithDeletedAt(t *testing.T, db *sqlx.DB, post *entity.UserPost) {
	query := `
		INSERT INTO user_posts (
			id, user_id, title, slug, content, status,
			tags, categories, meta_keywords,
			is_public, is_comments_enabled, reading_time_minutes,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := db.Exec(query,
		post.ID, post.UserID, post.Title, post.Slug, post.Content, post.Status,
		pq.Array(post.Tags), pq.Array(post.Categories), pq.Array(post.MetaKeywords),
		post.IsPublic, post.IsCommentsEnabled, post.ReadingTimeMinutes,
		post.CreatedAt, post.UpdatedAt, post.DeletedAt,
	)
	require.NoError(t, err, "Failed to insert post with deleted_at")
}

// insertCommentWithDeletedAt inserts a comment directly with deleted_at
func insertCommentWithDeletedAt(t *testing.T, db *sqlx.DB, comment *entity.PostComment) {
	query := `
		INSERT INTO post_comments (
			id, post_id, user_id, content,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.Exec(query,
		comment.ID, comment.PostID, comment.UserID, comment.Content,
		comment.CreatedAt, comment.UpdatedAt, comment.DeletedAt,
	)
	require.NoError(t, err, "Failed to insert comment with deleted_at")
}

func TestPurgeRepository_PurgeUserPosts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPurgeRepository(testDB.DB)
	postRepo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create posts with different deleted_at timestamps using direct SQL
	// because Create() doesn't support setting deleted_at
	oldPost := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
		p.Slug = "old-post"
		p.Title = "Old Post"
	})
	oldPost.DeletedAt = timePtr(time.Now().AddDate(0, 0, -100)) // 100 days ago
	insertPostWithDeletedAt(t, testDB.DB, oldPost)

	recentPost := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
		p.Slug = "recent-post"
		p.Title = "Recent Post"
	})
	recentPost.DeletedAt = timePtr(time.Now().AddDate(0, 0, -10)) // 10 days ago
	insertPostWithDeletedAt(t, testDB.DB, recentPost)

	activePost := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
		p.Slug = "active-post"
		p.Title = "Active Post"
	})
	// No deleted_at - this should not be purged
	require.NoError(t, postRepo.Create(ctx, activePost))

	// Purge posts older than 90 days
	cutoffDate := time.Now().AddDate(0, 0, -90)
	purged, err := repo.PurgeUserPosts(ctx, cutoffDate, 1000, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), purged, "Should purge 1 post (100 days old)")

	// Verify old post is gone (permanently deleted)
	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM user_posts WHERE id = $1", oldPost.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Old post should be permanently deleted")

	// Verify recent post still exists (soft-deleted but not purged)
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM user_posts WHERE id = $1", recentPost.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Recent post should still exist")

	// Verify active post still exists
	retrieved, err := postRepo.GetByID(ctx, activePost.ID)
	require.NoError(t, err)
	assert.Equal(t, activePost.ID, retrieved.ID)
}

func TestPurgeRepository_PurgeUserPosts_DryRun(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPurgeRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create old deleted post
	oldPost := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
		p.Slug = "old-post"
		p.Title = "Old Post"
	})
	oldPost.DeletedAt = timePtr(time.Now().AddDate(0, 0, -100))
	insertPostWithDeletedAt(t, testDB.DB, oldPost)

	// Dry run should count but not delete
	cutoffDate := time.Now().AddDate(0, 0, -90)
	purged, err := repo.PurgeUserPosts(ctx, cutoffDate, 1000, true)
	require.NoError(t, err)
	assert.Equal(t, int64(1), purged, "Should count 1 post")

	// Verify post still exists after dry run (check in DB directly)
	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM user_posts WHERE id = $1", oldPost.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Post should still exist after dry run")
}

func TestPurgeRepository_PurgePostComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPurgeRepository(testDB.DB)
	commentRepo := NewPostCommentRepository(testDB.DB)
	postRepo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create post
	post := helpers.UserPostFixture(user.ID)
	require.NoError(t, postRepo.Create(ctx, post))

	// Create comments with different deleted_at timestamps
	oldComment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Old comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: timePtr(time.Now().AddDate(0, 0, -50)), // 50 days ago
	}
	insertCommentWithDeletedAt(t, testDB.DB, oldComment)

	recentComment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Recent comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: timePtr(time.Now().AddDate(0, 0, -5)), // 5 days ago
	}
	insertCommentWithDeletedAt(t, testDB.DB, recentComment)

	activeComment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Active comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		// No deleted_at
	}
	require.NoError(t, commentRepo.Create(ctx, activeComment))

	// Purge comments older than 30 days
	cutoffDate := time.Now().AddDate(0, 0, -30)
	purged, err := repo.PurgePostComments(ctx, cutoffDate, 1000, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), purged, "Should purge 1 comment (50 days old)")

	// Verify old comment is gone (permanently deleted)
	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM post_comments WHERE id = $1", oldComment.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Old comment should be permanently deleted")

	// Verify recent comment still exists (soft-deleted but not purged)
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM post_comments WHERE id = $1", recentComment.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Recent comment should still exist")

	// Verify active comment still exists
	retrieved, err := commentRepo.GetByID(ctx, activeComment.ID)
	require.NoError(t, err)
	assert.Equal(t, activeComment.ID, retrieved.ID)
}

func TestPurgeRepository_CountDeletableUserPosts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPurgeRepository(testDB.DB)
	postRepo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create 3 old deleted posts
	for i := 0; i < 3; i++ {
		post := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
			p.Slug = fmt.Sprintf("old-post-%d", i)
			p.Title = fmt.Sprintf("Old Post %d", i)
		})
		post.DeletedAt = timePtr(time.Now().AddDate(0, 0, -100))
		insertPostWithDeletedAt(t, testDB.DB, post)
	}

	// Create 2 recent deleted posts
	for i := 0; i < 2; i++ {
		post := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
			p.Slug = fmt.Sprintf("recent-post-%d", i)
			p.Title = fmt.Sprintf("Recent Post %d", i)
		})
		post.DeletedAt = timePtr(time.Now().AddDate(0, 0, -10))
		insertPostWithDeletedAt(t, testDB.DB, post)
	}

	// Create 1 active post
	post := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
		p.Slug = "active-post"
		p.Title = "Active Post"
	})
	require.NoError(t, postRepo.Create(ctx, post))

	// Count posts older than 90 days
	cutoffDate := time.Now().AddDate(0, 0, -90)
	count, err := repo.CountDeletableUserPosts(ctx, cutoffDate)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "Should count 3 old deleted posts")

	// Count posts older than 5 days (should include all deleted posts)
	cutoffDate = time.Now().AddDate(0, 0, -5)
	count, err = repo.CountDeletableUserPosts(ctx, cutoffDate)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count, "Should count all 5 deleted posts")
}

func TestPurgeRepository_CountDeletablePostComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPurgeRepository(testDB.DB)
	postRepo := NewUserPostRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	post := helpers.UserPostFixture(user.ID)
	require.NoError(t, postRepo.Create(ctx, post))

	// Create 2 old deleted comments
	for i := 0; i < 2; i++ {
		comment := &entity.PostComment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   fmt.Sprintf("Old comment %d", i),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: timePtr(time.Now().AddDate(0, 0, -50)),
		}
		insertCommentWithDeletedAt(t, testDB.DB, comment)
	}

	// Create 1 recent deleted comment
	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Recent comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: timePtr(time.Now().AddDate(0, 0, -5)),
	}
	insertCommentWithDeletedAt(t, testDB.DB, comment)

	// Count comments older than 30 days
	cutoffDate := time.Now().AddDate(0, 0, -30)
	count, err := repo.CountDeletablePostComments(ctx, cutoffDate)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "Should count 2 old deleted comments")
}

func TestPurgeRepository_BatchProcessing(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPurgeRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")

	// Create 25 old deleted posts to test batch processing
	for i := 0; i < 25; i++ {
		post := helpers.UserPostFixture(user.ID, func(p *entity.UserPost) {
			p.Slug = fmt.Sprintf("batch-post-%d", i)
			p.Title = fmt.Sprintf("Batch Post %d", i)
		})
		post.DeletedAt = timePtr(time.Now().AddDate(0, 0, -100))
		insertPostWithDeletedAt(t, testDB.DB, post)
	}

	// Purge with small batch size
	cutoffDate := time.Now().AddDate(0, 0, -90)
	purged, err := repo.PurgeUserPosts(ctx, cutoffDate, 10, false)
	require.NoError(t, err)
	assert.Equal(t, int64(25), purged, "Should purge all 25 posts in batches of 10")

	// Verify all posts are gone
	count, err := repo.CountDeletableUserPosts(ctx, cutoffDate)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "All posts should be purged")
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
