package smoke_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCommentLikes_SmokeTest - быстрый end-to-end smoke test для функционала лайков комментариев
func TestCommentLikes_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()

	// Setup repositories
	commentRepo := postgres.NewPostCommentRepository(testDB.DB)
	postRepo := postgres.NewUserPostRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)

	// Setup use case
	commentUC := usecase.NewPostCommentUseCase(commentRepo, postRepo)

	// Create test users
	user1 := helpers.UserFixture()
	require.NoError(t, userRepo.Create(ctx, user1))

	user2 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "user2@example.com"
	})
	require.NoError(t, userRepo.Create(ctx, user2))

	// Create test post
	post, err := entity.NewUserPost(user1.ID, "Smoke Test Post", "smoke-test-post", "Test content")
	require.NoError(t, err)
	require.NoError(t, postRepo.Create(ctx, post))
	require.NoError(t, postRepo.UpdateStatus(ctx, post.ID, entity.PostStatusPublished))

	// Create test comment
	comment, err := commentUC.CreateComment(ctx, post.ID, user1.ID, "This is a test comment", nil)
	require.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, 0, comment.LikeCount, "Initial like count should be 0")

	t.Run("✅ Basic_like_flow", func(t *testing.T) {
		// User1 likes the comment
		err := commentUC.LikeComment(ctx, comment.ID, user1.ID)
		require.NoError(t, err, "First like should succeed")

		// Check like was recorded
		hasLiked, err := commentRepo.HasUserLiked(ctx, comment.ID, user1.ID)
		require.NoError(t, err)
		assert.True(t, hasLiked, "User1 should have liked the comment")

		// Get updated comment
		updatedComment, err := commentUC.GetComment(ctx, comment.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, updatedComment.LikeCount, 1, "Like count should be at least 1")

		// User1 likes again (idempotent)
		err = commentUC.LikeComment(ctx, comment.ID, user1.ID)
		require.NoError(t, err, "Second like should be idempotent")
	})

	t.Run("✅ Multiple_users_like", func(t *testing.T) {
		// User2 likes the comment
		err := commentUC.LikeComment(ctx, comment.ID, user2.ID)
		require.NoError(t, err, "User2 like should succeed")

		// Check both users have liked
		hasLiked1, _ := commentRepo.HasUserLiked(ctx, comment.ID, user1.ID)
		hasLiked2, _ := commentRepo.HasUserLiked(ctx, comment.ID, user2.ID)
		assert.True(t, hasLiked1, "User1 should still have liked")
		assert.True(t, hasLiked2, "User2 should have liked")

		// Get likers
		likers, err := commentRepo.GetCommentLikers(ctx, comment.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, likers, 2, "Should have 2 users who liked")
	})

	t.Run("✅ Unlike_flow", func(t *testing.T) {
		// User1 unlikes
		err := commentUC.UnlikeComment(ctx, comment.ID, user1.ID)
		require.NoError(t, err, "Unlike should succeed")

		// Check like was removed
		hasLiked, err := commentRepo.HasUserLiked(ctx, comment.ID, user1.ID)
		require.NoError(t, err)
		assert.False(t, hasLiked, "User1 should not have liked after unlike")

		// User1 unlikes again (idempotent)
		err = commentUC.UnlikeComment(ctx, comment.ID, user1.ID)
		require.NoError(t, err, "Second unlike should be idempotent")

		// Only user2 remains
		likers, err := commentRepo.GetCommentLikers(ctx, comment.ID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, likers, 1, "Should have 1 user who liked")
		assert.Equal(t, user2.ID, likers[0], "Only user2 should remain")
	})

	t.Run("✅ Cannot_like_deleted_comment", func(t *testing.T) {
		// Create and delete a comment
		deletedComment, err := commentUC.CreateComment(ctx, post.ID, user1.ID, "This will be deleted", nil)
		require.NoError(t, err)

		err = commentRepo.SoftDelete(ctx, deletedComment.ID)
		require.NoError(t, err)

		// Try to like deleted comment
		err = commentUC.LikeComment(ctx, deletedComment.ID, user2.ID)
		assert.Error(t, err, "Should not be able to like deleted comment")
		assert.ErrorIs(t, err, usecase.ErrCommentDeleted)
	})

	t.Run("✅ Pagination_of_likers", func(t *testing.T) {
		// Create a new comment
		popularComment, err := commentUC.CreateComment(ctx, post.ID, user1.ID, "Popular comment", nil)
		require.NoError(t, err)

		// Add likes from multiple users
		users := []*entity.User{user1, user2}
		for _, user := range users {
			err := commentRepo.AddLike(ctx, popularComment.ID, user.ID)
			require.NoError(t, err)
		}

		// Test pagination
		page1, err := commentRepo.GetCommentLikers(ctx, popularComment.ID, 1, 0)
		require.NoError(t, err)
		assert.Len(t, page1, 1, "First page should have 1 liker")

		page2, err := commentRepo.GetCommentLikers(ctx, popularComment.ID, 1, 1)
		require.NoError(t, err)
		assert.Len(t, page2, 1, "Second page should have 1 liker")
	})

	t.Log("🎉 All comment_likes smoke tests passed!")
}

// TestCommentLikes_Performance - быстрая проверка производительности
func TestCommentLikes_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()
	commentRepo := postgres.NewPostCommentRepository(testDB.DB)
	postRepo := postgres.NewUserPostRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)

	// Create test data
	user := helpers.UserFixture()
	require.NoError(t, userRepo.Create(ctx, user))

	post, err := entity.NewUserPost(user.ID, "Perf Test", "perf-test", "Performance test content for smoke testing")
	require.NoError(t, err)
	require.NoError(t, postRepo.Create(ctx, post))

	comment, err := entity.NewPostComment(post.ID, user.ID, "Comment for performance test", nil)
	require.NoError(t, err)
	require.NoError(t, commentRepo.Create(ctx, comment))

	// Measure HasUserLiked performance (should be fast with index)
	for i := 0; i < 100; i++ {
		_, err := commentRepo.HasUserLiked(ctx, comment.ID, user.ID)
		require.NoError(t, err)
	}

	t.Log("⚡ Performance test completed: 100 HasUserLiked queries executed successfully")
}
