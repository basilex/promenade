package smoke_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserPost_SmokeTest - fast end-to-end smoke test for user posts
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
	user1 := helpers.UserFixture()
	require.NoError(t, userRepo.Create(ctx, user1))

	user2 := helpers.UserFixture(func(u *entity.User) {
		u.Email = "user2@example.com"
	})
	require.NoError(t, userRepo.Create(ctx, user2))

	t.Run("[+] Post_create_and_retrieve", func(t *testing.T) {
		// Create post
		post, err := entity.NewUserPost(user1.ID, "Test Post", "test-post", "Test content")
		require.NoError(t, err)
		err = postRepo.Create(ctx, post)
		require.NoError(t, err)
		assert.NotEmpty(t, post.ID)

		// Get by ID
		retrieved, err := postRepo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, post.ID, retrieved.ID)

		// Update status
		err = postRepo.UpdateStatus(ctx, post.ID, entity.PostStatusPublished)
		require.NoError(t, err)

		// Verify status changed
		updated, err := postRepo.GetByID(ctx, post.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.PostStatusPublished, updated.Status)
	})

	t.Log("🎉 All user post smoke tests passed!")
}
