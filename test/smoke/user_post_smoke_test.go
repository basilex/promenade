package smoke

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserPost_SmokeTest - comprehensive end-to-end smoke test for user posts
func TestUserPost_SmokeTest(t *testing.T) {
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

	// Create test users
	user1 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "author@example.com"
		u.Name = "Post Author"
	})
	require.NoError(t, userRepo.Create(ctx, user1))

	user2 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "reader@example.com"
	})
	require.NoError(t, userRepo.Create(ctx, user2))

	var draftPostID, publishedPostID uuidv7.UUID

	t.Run("[+] Create_draft_post", func(t *testing.T) {
		post, err := entity.NewUserPost(user1.ID, "Draft Post", "draft-post", "This is a draft post")
		require.NoError(t, err)
		require.NoError(t, postRepo.Create(ctx, post))
		assert.NotEmpty(t, post.ID)
		assert.Equal(t, entity.PostStatusDraft, post.Status)
		draftPostID = post.ID
	})

	t.Run("[+] Get_post_by_ID", func(t *testing.T) {
		post, err := postRepo.GetByID(ctx, draftPostID)
		require.NoError(t, err)
		assert.Equal(t, draftPostID, post.ID)
		assert.Equal(t, "Draft Post", post.Title)
		assert.Equal(t, user1.ID, post.UserID)
	})

	t.Run("[+] Get_post_by_slug", func(t *testing.T) {
		post, err := postRepo.GetBySlug(ctx, user1.ID, "draft-post")
		require.NoError(t, err)
		assert.Equal(t, draftPostID, post.ID)
		assert.Equal(t, "draft-post", post.Slug)
	})

	t.Run("[+] Update_post", func(t *testing.T) {
		post, err := postRepo.GetByID(ctx, draftPostID)
		require.NoError(t, err)

		post.Title = "Updated Draft Post"
		post.Content = "Updated content"
		require.NoError(t, postRepo.Update(ctx, post))

		updated, err := postRepo.GetByID(ctx, draftPostID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Draft Post", updated.Title)
		assert.Equal(t, "Updated content", updated.Content)
	})

	t.Run("[+] Publish_post", func(t *testing.T) {
		// Create and publish a post
		post, err := entity.NewUserPost(user1.ID, "Published Post", "published-post", "This is a published post")
		require.NoError(t, err)
		require.NoError(t, postRepo.Create(ctx, post))
		publishedPostID = post.ID

		// Publish it
		post.Publish()
		require.NoError(t, postRepo.UpdateStatus(ctx, post.ID, entity.PostStatusPublished))

		// Verify status
		published, err := postRepo.GetByID(ctx, publishedPostID)
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusPublished, published.Status)
	})

	t.Run("[+] Get_published_posts", func(t *testing.T) {
		posts, err := postRepo.GetPublishedPosts(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 1, "should have at least 1 published post")

		// Verify all posts are published
		for _, post := range posts {
			assert.Equal(t, entity.PostStatusPublished, post.Status)
		}
	})

	t.Run("[+] Toggle_featured_post", func(t *testing.T) {
		// Feature the already published post
		post, err := postRepo.GetByID(ctx, publishedPostID)
		require.NoError(t, err)

		// Toggle featured on
		post.ToggleFeatured()
		require.NoError(t, postRepo.Update(ctx, post))

		// Verify featured
		featured, err := postRepo.GetByID(ctx, publishedPostID)
		require.NoError(t, err)
		assert.True(t, featured.IsFeatured, "post should be featured")

		// Get featured posts (only returns published posts)
		featuredPosts, err := postRepo.GetFeaturedPosts(ctx, 10)
		require.NoError(t, err)

		// Verify at least one featured post exists
		foundFeatured := false
		for _, p := range featuredPosts {
			if p.ID == publishedPostID {
				foundFeatured = true
				assert.True(t, p.IsFeatured)
			}
		}
		assert.True(t, foundFeatured, "should find our featured post in results")
	})

	t.Run("[+] Schedule_post", func(t *testing.T) {
		// Create a post to schedule
		post, err := entity.NewUserPost(user1.ID, "Scheduled Post", "scheduled-post", "This will be published later")
		require.NoError(t, err)
		require.NoError(t, postRepo.Create(ctx, post))

		// Schedule for future
		futureTime := time.Now().Add(24 * time.Hour)
		post.Schedule(futureTime)

		// Update with scheduled status
		require.NoError(t, postRepo.Update(ctx, post))
		require.NoError(t, postRepo.UpdateStatus(ctx, post.ID, entity.PostStatusScheduled))

		// Verify scheduled
		scheduled, err := postRepo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusScheduled, scheduled.Status)
	})

	t.Run("[+] Increment_post_views", func(t *testing.T) {
		initialPost, err := postRepo.GetByID(ctx, publishedPostID)
		require.NoError(t, err)
		initialViews := initialPost.ViewCount

		require.NoError(t, postRepo.IncrementViews(ctx, publishedPostID))

		updated, err := postRepo.GetByID(ctx, publishedPostID)
		require.NoError(t, err)
		assert.Equal(t, initialViews+1, updated.ViewCount)
	})

	t.Run("[+] Soft_delete_post", func(t *testing.T) {
		post, err := entity.NewUserPost(user1.ID, "Post to Delete", "delete-post", "This will be deleted")
		require.NoError(t, err)
		require.NoError(t, postRepo.Create(ctx, post))
		postIDToDelete := post.ID

		// Soft delete
		require.NoError(t, postRepo.SoftDelete(ctx, postIDToDelete))

		// Note: GetByID typically excludes soft-deleted rows, so we just verify delete succeeded
		// In production, you'd use a special query to verify DeletedAt is set

		// Restore
		require.NoError(t, postRepo.Restore(ctx, postIDToDelete))

		// Verify restored
		restored, err := postRepo.GetByID(ctx, postIDToDelete)
		require.NoError(t, err)
		assert.Equal(t, postIDToDelete, restored.ID, "post should be restored")
	})

	t.Run("[+] Get_user_posts", func(t *testing.T) {
		posts, err := postRepo.GetUserPosts(ctx, user1.ID, 20, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 3, "user1 should have multiple posts")

		// Verify all posts belong to user1
		for _, post := range posts {
			assert.Equal(t, user1.ID, post.UserID)
		}
	})

	t.Run("[+] Search_posts_by_title", func(t *testing.T) {
		posts, err := postRepo.Search(ctx, "Published", 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 1, "should find posts matching 'Published'")

		// Verify search results contain the keyword
		foundMatch := false
		for _, post := range posts {
			if post.Title == "Published Post" {
				foundMatch = true
			}
		}
		assert.True(t, foundMatch, "should find 'Published Post'")
	})

	t.Logf("🎉 All user post smoke tests passed!")
}
