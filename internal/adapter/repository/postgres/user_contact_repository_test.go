package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
)

func TestUserContactRepository_Create(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		// Create test user
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err)

		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "contact@example.com",
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = repo.Create(ctx, contact)
		require.NoError(t, err)

		// Verify contact was created
		retrieved, err := repo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, contact.ID, retrieved.ID)
		assert.Equal(t, contact.ContactValue, retrieved.ContactValue)
	})

	t.Run("duplicate contact value same type", func(t *testing.T) {
		// Create test user
		user := helpers.UserFixture()
		err := userRepo.Create(ctx, user)
		require.NoError(t, err, "failed to create user")

		contact1 := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypePhone,
			ContactValue: "+1234567890",
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = repo.Create(ctx, contact1)
		require.NoError(t, err)

		// Try to create duplicate
		contact2 := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypePhone,
			ContactValue: "+1234567890",
			IsPublic:     false,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = repo.Create(ctx, contact2)
		assert.Error(t, err)
	})
}

func TestUserContactRepository_GetByID(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("existing contact", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeTelegram,
			ContactValue: "@testuser",
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, contact)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, contact.ID, retrieved.ID)
		assert.Equal(t, contact.ContactValue, retrieved.ContactValue)
		assert.Equal(t, contact.ContactType, retrieved.ContactType)
	})

	t.Run("non-existent contact", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuidv7.New())
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestUserContactRepository_GetUserContacts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple contacts
	contacts := []*entity.UserContact{
		{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "email1@example.com",
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypePhone,
			ContactValue: "+1234567890",
			IsPublic:     false,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeTelegram,
			ContactValue: "@testuser",
			IsPublic:     true,
			IsActive:     false,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, c := range contacts {
		err := repo.Create(ctx, c)
		require.NoError(t, err)
	}

	t.Run("get all active contacts", func(t *testing.T) {
		result, err := repo.GetUserContacts(ctx, user.ID, false)
		require.NoError(t, err)
		assert.Len(t, result, 2) // Only active contacts
	})

	t.Run("get all contacts including inactive", func(t *testing.T) {
		result, err := repo.GetUserContacts(ctx, user.ID, true)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})
}

func TestUserContactRepository_GetUserContactsByType(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create contacts of different types
	emailContact1 := &entity.UserContact{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		ContactType:  entity.ContactTypeEmail,
		ContactValue: "email1@example.com",
		IsPublic:     true,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	emailContact2 := &entity.UserContact{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		ContactType:  entity.ContactTypeEmail,
		ContactValue: "email2@example.com",
		IsPublic:     false,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	phoneContact := &entity.UserContact{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		ContactType:  entity.ContactTypePhone,
		ContactValue: "+1234567890",
		IsPublic:     true,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, repo.Create(ctx, emailContact1))
	require.NoError(t, repo.Create(ctx, emailContact2))
	require.NoError(t, repo.Create(ctx, phoneContact))

	t.Run("get email contacts only", func(t *testing.T) {
		result, err := repo.GetUserContactsByType(ctx, user.ID, entity.ContactTypeEmail)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		for _, c := range result {
			assert.Equal(t, entity.ContactTypeEmail, c.ContactType)
		}
	})

	t.Run("get phone contacts only", func(t *testing.T) {
		result, err := repo.GetUserContactsByType(ctx, user.ID, entity.ContactTypePhone)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, entity.ContactTypePhone, result[0].ContactType)
	})
}

func TestUserContactRepository_GetPrimaryContact(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("get primary contact", func(t *testing.T) {
		primaryContact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "primary@example.com",
			IsPrimary:    true,
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		regularContact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "regular@example.com",
			IsPrimary:    false,
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		require.NoError(t, repo.Create(ctx, primaryContact))
		require.NoError(t, repo.Create(ctx, regularContact))

		result, err := repo.GetPrimaryContact(ctx, user.ID, entity.ContactTypeEmail)
		require.NoError(t, err)
		assert.Equal(t, primaryContact.ID, result.ID)
		assert.True(t, result.IsPrimary)
	})

	t.Run("no primary contact", func(t *testing.T) {
		// Create a different user with unique email
		user2 := helpers.UserFixture()
		err := userRepo.Create(ctx, user2)
		require.NoError(t, err, "failed to create second user")

		_, err = repo.GetPrimaryContact(ctx, user2.ID, entity.ContactTypePhone)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestUserContactRepository_GetPublicContacts(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create public and private contacts
	publicContact := &entity.UserContact{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		ContactType:  entity.ContactTypeEmail,
		ContactValue: "public@example.com",
		IsPublic:     true,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	privateContact := &entity.UserContact{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		ContactType:  entity.ContactTypePhone,
		ContactValue: "+1234567890",
		IsPublic:     false,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	require.NoError(t, repo.Create(ctx, publicContact))
	require.NoError(t, repo.Create(ctx, privateContact))

	t.Run("get only public contacts", func(t *testing.T) {
		result, err := repo.GetPublicContacts(ctx, user.ID)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.True(t, result[0].IsPublic)
		assert.Equal(t, publicContact.ID, result[0].ID)
	})
}

func TestUserContactRepository_Update(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "old@example.com",
			IsPublic:     false,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, contact)
		require.NoError(t, err)

		// Update contact
		contact.ContactValue = "new@example.com"
		contact.IsPublic = true
		contact.UpdatedAt = time.Now()

		err = repo.Update(ctx, contact)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.Equal(t, "new@example.com", retrieved.ContactValue)
		assert.True(t, retrieved.IsPublic)
	})
}

func TestUserContactRepository_Delete(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeWhatsApp,
			ContactValue: "+1234567890",
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, contact)
		require.NoError(t, err)

		// Delete contact
		err = repo.Delete(ctx, contact.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, contact.ID)
		assert.ErrorIs(t, err, entity.ErrNotFound)
	})
}

func TestUserContactRepository_SetPrimary(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("set primary contact", func(t *testing.T) {
		contact1 := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "contact1@example.com",
			IsPrimary:    true,
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		contact2 := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "contact2@example.com",
			IsPrimary:    false,
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		require.NoError(t, repo.Create(ctx, contact1))
		require.NoError(t, repo.Create(ctx, contact2))

		// Set contact2 as primary
		err := repo.SetPrimary(ctx, user.ID, entity.ContactTypeEmail, contact2.ID)
		require.NoError(t, err)

		// Verify contact2 is now primary
		retrieved2, err := repo.GetByID(ctx, contact2.ID)
		require.NoError(t, err)
		assert.True(t, retrieved2.IsPrimary)

		// Verify contact1 is no longer primary
		retrieved1, err := repo.GetByID(ctx, contact1.ID)
		require.NoError(t, err)
		assert.False(t, retrieved1.IsPrimary)
	})
}

func TestUserContactRepository_VerifyContact(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("verify unverified contact", func(t *testing.T) {
		contact := &entity.UserContact{
			ID:           uuidv7.New(),
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "verify@example.com",
			IsVerified:   false,
			IsPublic:     true,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err := repo.Create(ctx, contact)
		require.NoError(t, err)

		// Verify contact
		err = repo.VerifyContact(ctx, contact.ID)
		require.NoError(t, err)

		// Check verification status
		retrieved, err := repo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.IsVerified)
	})
}

func TestUserContactRepository_WithAvailability(t *testing.T) {
	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	repo := postgres.NewUserContactRepository(testDB.DB)
	userRepo := postgres.NewUserRepository(testDB.DB)
	ctx := context.Background()

	user := helpers.UserFixture()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("create contact with availability", func(t *testing.T) {
		from := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
		to := time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC)

		contact := &entity.UserContact{
			ID:            uuidv7.New(),
			UserID:        user.ID,
			ContactType:   entity.ContactTypePhone,
			ContactValue:  "+1234567890",
			IsPublic:      true,
			IsActive:      true,
			AvailableFrom: &from,
			AvailableTo:   &to,
			AvailableDays: pq.StringArray{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.Create(ctx, contact)
		require.NoError(t, err)

		// Verify availability fields
		retrieved, err := repo.GetByID(ctx, contact.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.AvailableFrom)
		assert.NotNil(t, retrieved.AvailableTo)
		assert.NotEmpty(t, retrieved.AvailableDays)
		assert.Len(t, retrieved.AvailableDays, 5)
	})
}
