package smoke

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostComment_SmokeTest verifies complete comment CRUD flow including threading
func TestPostComment_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()

	// Setup repositories
	userRepo := postgres.NewUserRepository(testDB.DB)
	postRepo := postgres.NewUserPostRepository(testDB.DB)
	commentRepo := postgres.NewPostCommentRepository(testDB.DB)

	// Setup use case
	commentUC := usecase.NewPostCommentUseCase(commentRepo, postRepo)

	// Create test users
	user1 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "author@test.com"
		u.Name = "Author User"
	})
	require.NoError(t, userRepo.Create(ctx, user1))

	user2 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "commenter@test.com"
		u.Name = "Commenter User"
	})
	require.NoError(t, userRepo.Create(ctx, user2))

	// Create test post
	post, err := entity.NewUserPost(user1.ID, "Smoke Test Post", "smoke-test-post", "Post content for testing comments")
	require.NoError(t, err)
	require.NoError(t, postRepo.Create(ctx, post))
	require.NoError(t, postRepo.UpdateStatus(ctx, post.ID, entity.PostStatusPublished))

	var rootCommentID, replyCommentID uuidv7.UUID

	// ========== Comment CRUD ==========

	t.Run("[+] Create_root_comment", func(t *testing.T) {
		comment, err := commentUC.CreateComment(ctx, post.ID, user1.ID, "This is a root comment", nil)
		require.NoError(t, err)
		require.NotNil(t, comment)
		assert.Equal(t, "This is a root comment", comment.Content)
		assert.Nil(t, comment.ParentID, "root comment should have no parent")
		assert.Equal(t, 0, comment.LikeCount)
		assert.Equal(t, 0, comment.ReplyCount)
		rootCommentID = comment.ID
	})

	t.Run("[+] Get_comment_by_ID", func(t *testing.T) {
		comment, err := commentUC.GetComment(ctx, rootCommentID)
		require.NoError(t, err)
		assert.Equal(t, rootCommentID, comment.ID)
		assert.Equal(t, "This is a root comment", comment.Content)
		assert.Equal(t, user1.ID, comment.UserID)
	})

	t.Run("[+] Update_comment_content", func(t *testing.T) {
		updatedComment, err := commentUC.UpdateComment(ctx, rootCommentID, user1.ID, "This is an updated root comment")
		require.NoError(t, err)
		assert.Equal(t, "This is an updated root comment", updatedComment.Content)

		// Verify update persisted
		comment, err := commentUC.GetComment(ctx, rootCommentID)
		require.NoError(t, err)
		assert.Equal(t, "This is an updated root comment", comment.Content)
	})

	// ========== Comment Threading (Replies) ==========

	t.Run("[+] Create_reply", func(t *testing.T) {
		reply, err := commentUC.CreateComment(ctx, post.ID, user2.ID, "This is a reply to root comment", &rootCommentID)
		require.NoError(t, err)
		require.NotNil(t, reply)
		assert.Equal(t, "This is a reply to root comment", reply.Content)
		assert.NotNil(t, reply.ParentID, "reply should have parent")
		assert.Equal(t, rootCommentID, *reply.ParentID)
		replyCommentID = reply.ID

		// Verify reply count incremented on root comment
		rootComment, err := commentUC.GetComment(ctx, rootCommentID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, rootComment.ReplyCount, 1, "root comment should have at least 1 reply")
	})

	t.Run("[+] Create_nested_reply", func(t *testing.T) {
		// Reply to the reply (nested threading)
		nestedReply, err := commentUC.CreateComment(ctx, post.ID, user1.ID, "This is a nested reply", &replyCommentID)
		require.NoError(t, err)
		assert.NotNil(t, nestedReply.ParentID)
		assert.Equal(t, replyCommentID, *nestedReply.ParentID)

		// Verify reply count incremented
		firstReply, err := commentUC.GetComment(ctx, replyCommentID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, firstReply.ReplyCount, 1, "first reply should have nested replies")
	})

	t.Run("[+] Get_comment_replies_with_pagination", func(t *testing.T) {
		replies, metadata, err := commentUC.GetCommentReplies(ctx, rootCommentID, 10, 0)
		require.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.GreaterOrEqual(t, len(replies), 1, "should have at least 1 reply")

		// Verify reply content
		foundReply := false
		for _, reply := range replies {
			if reply.ID == replyCommentID {
				foundReply = true
				assert.Equal(t, rootCommentID, *reply.ParentID)
			}
		}
		assert.True(t, foundReply, "should find the created reply")
	})

	// ========== Post Comments Listing ==========

	t.Run("[+] Get_post_comments_with_pagination", func(t *testing.T) {
		comments, metadata, err := commentUC.GetPostComments(ctx, post.ID, 10, 0)
		require.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.GreaterOrEqual(t, len(comments), 1, "should have at least 1 root-level comment")
		assert.Greater(t, metadata.Total, 0, "total count should be greater than 0")
	})

	t.Run("[+] Get_user_comments", func(t *testing.T) {
		comments, metadata, err := commentUC.GetUserComments(ctx, user2.ID, 10, 0)
		require.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.GreaterOrEqual(t, len(comments), 1, "user2 should have at least 1 comment")

		// Verify all comments belong to user2
		for _, comment := range comments {
			assert.Equal(t, user2.ID, comment.UserID)
		}
	})

	// ========== Soft Delete & Error Cases ==========

	t.Run("[+] Soft_delete_comment", func(t *testing.T) {
		err := commentUC.DeleteComment(ctx, replyCommentID, user2.ID)
		require.NoError(t, err)

		// Soft deleted comments are hidden from GetByID (filtered by deleted_at IS NULL)
		// This is expected behavior - see docs/SOFT_DELETE.md
		_, err = commentRepo.GetByID(ctx, replyCommentID)
		assert.Error(t, err, "soft deleted comment should not be returned by GetByID")
	})

	t.Run("[+] Cannot_reply_to_deleted_comment", func(t *testing.T) {
		_, err := commentUC.CreateComment(ctx, post.ID, user1.ID, "Reply to deleted", &replyCommentID)
		assert.Error(t, err, "should not be able to reply to deleted comment")
	})

	t.Run("[+] Cannot_update_other_users_comment", func(t *testing.T) {
		// User2 tries to update user1's comment
		_, err := commentUC.UpdateComment(ctx, rootCommentID, user2.ID, "Hacked content")
		assert.Error(t, err, "should not be able to update other user's comment")
	})

	t.Run("[+] Cannot_delete_other_users_comment", func(t *testing.T) {
		// User2 tries to delete user1's comment
		err := commentUC.DeleteComment(ctx, rootCommentID, user2.ID)
		assert.Error(t, err, "should not be able to delete other user's comment")
	})

	t.Logf("[SUCCESS] All post comment smoke tests passed!")
}
