package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestPostForComments creates a test post for comment testing
func createTestPostForComments(t *testing.T, db *sqlx.DB, userID uuidv7.UUID) *entity.UserPost {
	postID := uuidv7.New()
	query := `
		INSERT INTO user_posts (
			id, user_id, title, slug, content, status, 
			is_public, is_comments_enabled, 
			tags, categories, meta_keywords,
			reading_time_minutes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := db.Exec(query,
		postID, userID, "Test Post", "test-post", "Test content", "published",
		true, true,
		`[]`, `[]`, `{}`,
		5, time.Now(), time.Now(),
	)
	require.NoError(t, err, "Failed to create test post")

	return &entity.UserPost{
		ID:     postID,
		UserID: userID,
	}
}

func TestPostCommentRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:         uuidv7.New(),
		PostID:     post.ID,
		UserID:     user.ID,
		Content:    "This is a test comment",
		IsEdited:   false,
		LikeCount:  0,
		ReplyCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Verify comment was created
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, comment.ID, retrieved.ID)
	assert.Equal(t, comment.Content, retrieved.Content)
	assert.Equal(t, comment.PostID, retrieved.PostID)
}

func TestPostCommentRepository_GetByID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Test GetByID
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, comment.ID, retrieved.ID)

	// Test not found
	_, err = repo.GetByID(ctx, uuidv7.New())
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestPostCommentRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Original content",
		IsEdited:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Update comment
	editedAt := time.Now()
	comment.Content = "Updated content"
	comment.IsEdited = true
	comment.EditedAt = &editedAt
	comment.UpdatedAt = time.Now()

	err = repo.Update(ctx, comment)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated content", retrieved.Content)
	assert.True(t, retrieved.IsEdited)
	assert.NotNil(t, retrieved.EditedAt)
}

func TestPostCommentRepository_Delete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Delete comment
	err = repo.Delete(ctx, comment.ID)
	require.NoError(t, err)

	// Verify deletion (hard delete removes completely)
	_, err = repo.GetByID(ctx, comment.ID)
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestPostCommentRepository_SoftDelete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Soft delete comment
	err = repo.SoftDelete(ctx, comment.ID)
	require.NoError(t, err)

	// Comment still exists in DB but has deleted_at
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.DeletedAt)
}

func TestPostCommentRepository_Restore(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Soft delete
	err = repo.SoftDelete(ctx, comment.ID)
	require.NoError(t, err)

	// Restore
	err = repo.Restore(ctx, comment.ID)
	require.NoError(t, err)

	// Verify restoration
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved.DeletedAt)
}

func TestPostCommentRepository_GetPostComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	// Create multiple top-level comments
	for i := 0; i < 3; i++ {
		comment := &entity.PostComment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   "Comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(ctx, comment)
		require.NoError(t, err)
	}

	// Get post comments
	comments, meta, err := repo.GetPostComments(ctx, post.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, comments, 3)
	assert.Equal(t, 3, meta.Total)

	// Test pagination
	comments, meta, err = repo.GetPostComments(ctx, post.ID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, 3, meta.Total)
}

func TestPostCommentRepository_GetCommentReplies(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	// Create parent comment
	parent := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Parent comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, parent)
	require.NoError(t, err)

	// Create replies
	for i := 0; i < 2; i++ {
		reply := &entity.PostComment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			ParentID:  &parent.ID,
			Content:   "Reply",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(ctx, reply)
		require.NoError(t, err)
	}

	// Get replies
	replies, meta, err := repo.GetCommentReplies(ctx, parent.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, replies, 2)
	assert.Equal(t, 2, meta.Total)
}

func TestPostCommentRepository_GetUserComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	// Create multiple comments by user
	for i := 0; i < 3; i++ {
		comment := &entity.PostComment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   "User comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(ctx, comment)
		require.NoError(t, err)
	}

	// Get user comments
	comments, meta, err := repo.GetUserComments(ctx, user.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, comments, 3)
	assert.Equal(t, 3, meta.Total)
}

func TestPostCommentRepository_CountPostComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	// Initially no comments
	count, err := repo.CountPostComments(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create comments
	for i := 0; i < 3; i++ {
		comment := &entity.PostComment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   "Comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(ctx, comment)
		require.NoError(t, err)
	}

	// Count should be 3
	count, err = repo.CountPostComments(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestPostCommentRepository_CountUserComments(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	// Create comments
	for i := 0; i < 2; i++ {
		comment := &entity.PostComment{
			ID:        uuidv7.New(),
			PostID:    post.ID,
			UserID:    user.ID,
			Content:   "Comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(ctx, comment)
		require.NoError(t, err)
	}

	count, err := repo.CountUserComments(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestPostCommentRepository_IncrementLikes(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Test comment",
		LikeCount: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Increment likes
	err = repo.IncrementLikes(ctx, comment.ID)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.LikeCount)
}

func TestPostCommentRepository_DecrementLikes(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:        uuidv7.New(),
		PostID:    post.ID,
		UserID:    user.ID,
		Content:   "Test comment",
		LikeCount: 5,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Decrement likes
	err = repo.DecrementLikes(ctx, comment.ID)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, retrieved.LikeCount)
}

func TestPostCommentRepository_IncrementReplies(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:         uuidv7.New(),
		PostID:     post.ID,
		UserID:     user.ID,
		Content:    "Test comment",
		ReplyCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Increment replies
	err = repo.IncrementReplies(ctx, comment.ID)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, retrieved.ReplyCount)
}

func TestPostCommentRepository_DecrementReplies(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := NewPostCommentRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.CreateTestUser(t, testDB.DB, "test@example.com", "Test User")
	post := createTestPostForComments(t, testDB.DB, user.ID)

	comment := &entity.PostComment{
		ID:         uuidv7.New(),
		PostID:     post.ID,
		UserID:     user.ID,
		Content:    "Test comment",
		ReplyCount: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)

	// Decrement replies
	err = repo.DecrementReplies(ctx, comment.ID)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, retrieved.ReplyCount)
}
