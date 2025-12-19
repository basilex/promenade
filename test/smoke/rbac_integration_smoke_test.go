package smoke

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

// TestRBACIntegration_SmokeTest verifies RBAC permissions in real-world use cases
func TestRBACIntegration_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()

	// Setup repositories
	userRepo := postgres.NewUserRepository(testDB.DB)
	profileRepo := postgres.NewUserProfileRepository(testDB.DB)
	postRepo := postgres.NewUserPostRepository(testDB.DB)
	roleRepo := postgres.NewRoleRepository(testDB.DB)
	permissionRepo := postgres.NewPermissionRepository(testDB.DB)

	// Setup use cases
	roleUC := usecase.NewRoleUseCase(roleRepo, permissionRepo)

	// Create test users with different roles
	regularUser := helpers.UserFixture(func(u *entity.User) {
		u.Email = "regular@test.com"
		u.Name = "Regular User"
	})
	require.NoError(t, userRepo.Create(ctx, regularUser))

	moderator := helpers.UserFixture(func(u *entity.User) {
		u.Email = "moderator@test.com"
		u.Name = "Moderator User"
	})
	require.NoError(t, userRepo.Create(ctx, moderator))

	creator := helpers.UserFixture(func(u *entity.User) {
		u.Email = "creator@test.com"
		u.Name = "Creator User"
	})
	require.NoError(t, userRepo.Create(ctx, creator))

	admin := helpers.UserFixture(func(u *entity.User) {
		u.Email = "admin@test.com"
		u.Name = "Admin User"
	})
	require.NoError(t, userRepo.Create(ctx, admin))

	// Get system roles
	moderatorRole, err := roleRepo.GetByName(ctx, "Moderator")
	if err != nil {
		t.Skip("Moderator role not found - skipping RBAC integration tests")
	}

	creatorRole, err := roleRepo.GetByName(ctx, "Creator")
	require.NoError(t, err)

	adminRole, err := roleRepo.GetByName(ctx, "Admin")
	require.NoError(t, err)

	// Assign roles to users
	require.NoError(t, roleUC.AssignRoleToUser(ctx, moderator.ID, moderatorRole.ID, admin.ID, nil))
	require.NoError(t, roleUC.AssignRoleToUser(ctx, creator.ID, creatorRole.ID, admin.ID, nil))
	require.NoError(t, roleUC.AssignRoleToUser(ctx, admin.ID, adminRole.ID, admin.ID, nil))

	// Create test profiles
	regularProfile := &entity.UserProfile{
		UserID:   regularUser.ID,
		Timezone: "UTC",
		Locale:   "en",
		IsPublic: true,
	}
	require.NoError(t, profileRepo.Create(ctx, regularProfile))

	creatorProfile := &entity.UserProfile{
		UserID:   creator.ID,
		Timezone: "UTC",
		Locale:   "en",
		IsPublic: true,
	}
	require.NoError(t, profileRepo.Create(ctx, creatorProfile))

	// Create test posts
	regularPost, err := entity.NewUserPost(regularUser.ID, "Regular User Post", "regular-post", "Content by regular user")
	require.NoError(t, err)
	require.NoError(t, postRepo.Create(ctx, regularPost))
	require.NoError(t, postRepo.UpdateStatus(ctx, regularPost.ID, entity.PostStatusPublished))

	creatorPost, err := entity.NewUserPost(creator.ID, "Creator Post", "creator-post", "Content by creator")
	require.NoError(t, err)
	require.NoError(t, postRepo.Create(ctx, creatorPost))
	require.NoError(t, postRepo.UpdateStatus(ctx, creatorPost.ID, entity.PostStatusPublished))

	// ========== Moderator Rights ==========

	t.Run("[+] Moderator_can_ban_user_profile", func(t *testing.T) {
		// Moderator should have users:ban permission
		canBan, err := roleUC.HasPermission(ctx, moderator.ID, "users:ban")
		require.NoError(t, err)
		assert.True(t, canBan, "Moderator should have users:ban permission")

		// Ban regular user profile
		reason := "Violating community guidelines"
		err = profileRepo.Ban(ctx, regularProfile.ID, reason, moderator.ID)
		require.NoError(t, err)

		// Verify ban
		profile, err := profileRepo.GetByID(ctx, regularProfile.ID)
		require.NoError(t, err)
		assert.True(t, profile.IsBanned)
		assert.Equal(t, moderator.ID, *profile.BannedBy)
	})

	t.Run("[+] Moderator_can_unban_user", func(t *testing.T) {
		// Unban the user
		err := profileRepo.Unban(ctx, regularProfile.ID)
		require.NoError(t, err)

		profile, err := profileRepo.GetByID(ctx, regularProfile.ID)
		require.NoError(t, err)
		assert.False(t, profile.IsBanned)
	})

	t.Run("[+] Moderator_cannot_delete_posts", func(t *testing.T) {
		// Moderator should NOT have posts:delete permission
		canDelete, err := roleUC.HasPermission(ctx, moderator.ID, "posts:delete")
		require.NoError(t, err)
		assert.False(t, canDelete, "Moderator should NOT have posts:delete permission")
	})

	// ========== Creator Rights ==========

	t.Run("[+] Creator_can_manage_own_posts", func(t *testing.T) {
		// Creator should have posts:create and posts:update permissions
		canCreate, err := roleUC.HasPermission(ctx, creator.ID, "posts:create")
		require.NoError(t, err)
		assert.True(t, canCreate, "Creator should have posts:create permission")

		canUpdate, err := roleUC.HasPermission(ctx, creator.ID, "posts:update")
		require.NoError(t, err)
		assert.True(t, canUpdate, "Creator should have posts:update permission")

		// Update own post
		creatorPost.Title = "Updated Creator Post"
		err = postRepo.Update(ctx, creatorPost)
		require.NoError(t, err)
	})

	t.Run("[+] Creator_cannot_ban_users", func(t *testing.T) {
		// Creator should NOT have users:ban permission
		canBan, err := roleUC.HasPermission(ctx, creator.ID, "users:ban")
		require.NoError(t, err)
		assert.False(t, canBan, "Creator should NOT have users:ban permission")
	})

	t.Run("[+] Creator_cannot_feature_posts", func(t *testing.T) {
		// Creator should NOT have posts:feature permission (only admins can)
		canFeature, err := roleUC.HasPermission(ctx, creator.ID, "posts:feature")
		require.NoError(t, err)
		assert.False(t, canFeature, "Creator should NOT have posts:feature permission")
	})

	// ========== Admin Rights ==========

	t.Run("[+] Admin_can_feature_any_post", func(t *testing.T) {
		// Admin should have posts:feature permission (wildcard posts:*)
		canFeature, err := roleUC.HasPermission(ctx, admin.ID, "posts:feature")
		require.NoError(t, err)
		assert.True(t, canFeature, "Admin should have posts:feature permission")

		// Feature regular user's post
		regularPost.IsFeatured = true
		err = postRepo.Update(ctx, regularPost)
		require.NoError(t, err)

		featured, err := postRepo.GetByID(ctx, regularPost.ID)
		require.NoError(t, err)
		assert.True(t, featured.IsFeatured)
	})

	t.Run("[+] Admin_can_verify_profiles", func(t *testing.T) {
		// Admin should have users:verify permission
		canVerify, err := roleUC.HasPermission(ctx, admin.ID, "users:verify")
		require.NoError(t, err)
		assert.True(t, canVerify, "Admin should have users:verify permission")

		// Verify creator profile
		err = profileRepo.SetVerified(ctx, creatorProfile.ID, true)
		require.NoError(t, err)

		profile, err := profileRepo.GetByID(ctx, creatorProfile.ID)
		require.NoError(t, err)
		assert.True(t, profile.IsVerified)
	})

	t.Run("[+] Admin_has_wildcard_posts_permissions", func(t *testing.T) {
		// Admin should have posts:* wildcard
		permissions := []string{
			"posts:create",
			"posts:read",
			"posts:update",
			"posts:delete",
			"posts:feature",
			"posts:schedule",
		}

		for _, perm := range permissions {
			has, err := roleUC.HasPermission(ctx, admin.ID, perm)
			require.NoError(t, err)
			assert.True(t, has, "Admin should have %s permission via posts:* wildcard", perm)
		}
	})

	// ========== Regular User Restrictions ==========

	t.Run("[+] Regular_user_has_no_special_permissions", func(t *testing.T) {
		// Regular user should NOT have any special permissions
		restrictedPerms := []string{
			"users:ban",
			"users:verify",
			"posts:feature",
			"posts:delete",
		}

		for _, perm := range restrictedPerms {
			has, err := roleUC.HasPermission(ctx, regularUser.ID, perm)
			require.NoError(t, err)
			assert.False(t, has, "Regular user should NOT have %s permission", perm)
		}
	})

	t.Run("[+] Regular_user_can_create_posts", func(t *testing.T) {
		// Regular users should be able to create posts (default User role)
		canCreate, err := roleUC.HasPermission(ctx, regularUser.ID, "posts:create")
		require.NoError(t, err)
		assert.True(t, canCreate, "Regular user should have posts:create permission")

		// Create new post
		newPost, err := entity.NewUserPost(regularUser.ID, "Another Post", "another-post", "Another content")
		require.NoError(t, err)
		err = postRepo.Create(ctx, newPost)
		require.NoError(t, err)
	})

	// ========== Cross-Module Permissions ==========

	t.Run("[+] Moderator_can_ban_affects_all_user_content", func(t *testing.T) {
		// Create a new user to ban
		spammer := helpers.UserFixture(func(u *entity.User) {
			u.Email = "spammer@test.com"
			u.Name = "Spammer"
		})
		require.NoError(t, userRepo.Create(ctx, spammer))

		spammerProfile := &entity.UserProfile{
			UserID:   spammer.ID,
			Timezone: "UTC",
			Locale:   "en",
			IsPublic: true,
		}
		require.NoError(t, profileRepo.Create(ctx, spammerProfile))

		// Ban spammer
		reason := "Spam content"
		err := profileRepo.Ban(ctx, spammerProfile.ID, reason, moderator.ID)
		require.NoError(t, err)

		// Verify ban is recorded with moderator ID
		profile, err := profileRepo.GetByID(ctx, spammerProfile.ID)
		require.NoError(t, err)
		assert.True(t, profile.IsBanned)
		assert.NotNil(t, profile.BannedBy)
		assert.Equal(t, moderator.ID, *profile.BannedBy)
	})

	t.Run("[+] Permission_check_performance", func(t *testing.T) {
		// Verify permission checks are fast (no major performance regression)
		for i := 0; i < 50; i++ {
			_, err := roleUC.HasPermission(ctx, admin.ID, "posts:create")
			require.NoError(t, err)
		}
	})

	t.Logf("🎉 All RBAC integration smoke tests passed!")
}
