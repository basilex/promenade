package contact_test

import (
	"context"
	"testing"
"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	contactPostgres "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/identity/user"
	userPostgres "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestUser creates a test user for contact tests
func createTestUser(t *testing.T, ctx context.Context, userRepo user.IRepository) uuidv7.UUID {
	u, err := user.NewUser(fmt.Sprintf("test_%s@example.com", uuidv7.New().String()), "password123")
	require.NoError(t, err)
	err = userRepo.Create(ctx, u)
	require.NoError(t, err)
	return u.ID
}

// TestContactUseCase_CreateEmailContact tests email contact creation
func TestContactUseCase_CreateEmailContact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)
		email := fmt.Sprintf("test_%s@example.com", uuidv7.New().String())
		label := "Work"

		// Create email contact
		c, err := uc.CreateEmailContact(ctx, userID, email, label, true)
		require.NoError(t, err)
		require.NotNil(t, c)

		// Verify properties
		assert.Equal(t, userID, c.UserID)
		assert.Equal(t, contact.ContactTypeEmail, c.Type)
		assert.Equal(t, label, c.Label)
		assert.True(t, c.IsPrimary)
		assert.False(t, c.IsVerified)
		assert.NotNil(t, c.Email)
		assert.Equal(t, email, c.Email.Value())

		// Verify persistence
		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, retrieved.ID)
		assert.Equal(t, email, retrieved.Email.Value())
	})
}

// TestContactUseCase_CreatePhoneContact tests phone contact creation
func TestContactUseCase_CreatePhoneContact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)
		phone := "+380671234567"
		label := "Mobile"

		// Create phone contact
		c, err := uc.CreatePhoneContact(ctx, userID, phone, label, false)
		require.NoError(t, err)
		require.NotNil(t, c)

		// Verify properties
		assert.Equal(t, userID, c.UserID)
		assert.Equal(t, contact.ContactTypePhone, c.Type)
		assert.Equal(t, label, c.Label)
		assert.False(t, c.IsPrimary)
		assert.NotNil(t, c.Phone)
		assert.Equal(t, phone, c.Phone.Value())
	})
}

// TestContactUseCase_CreateAddressContact tests address contact creation
func TestContactUseCase_CreateAddressContact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)
		street := "123 Main St"
		city := "Kyiv"
		country := "UA"
		postalCode := "01001"
		label := "Home"

		// Create address contact
		c, err := uc.CreateAddressContact(ctx, userID, street, city, country, postalCode, label, true)
		require.NoError(t, err)
		require.NotNil(t, c)

		// Verify properties
		assert.Equal(t, userID, c.UserID)
		assert.Equal(t, contact.ContactTypeAddress, c.Type)
		assert.Equal(t, label, c.Label)
		assert.True(t, c.IsPrimary)
		assert.NotNil(t, c.Address)
		assert.Equal(t, street, c.Address.Street)
		assert.Equal(t, city, c.Address.City)
		assert.Equal(t, country, c.Address.Country)
		assert.Equal(t, postalCode, c.Address.PostalCode)
	})
}

// TestContactUseCase_GetContact tests retrieving contact by ID
func TestContactUseCase_GetContact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create contact
		created, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("test_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)

		// Retrieve contact
		retrieved, err := uc.GetContact(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, userID, retrieved.UserID)

		// Test not found
		nonExistentID := uuidv7.New()
		_, err = uc.GetContact(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contact not found")
	})
}

// TestContactUseCase_GetUserContacts tests retrieving all user contacts
func TestContactUseCase_GetUserContacts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create multiple contacts
		_, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("email1_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)
		_, err = uc.CreatePhoneContact(ctx, userID, "+380671234567", "Mobile", false)
		require.NoError(t, err)
		_, err = uc.CreateAddressContact(ctx, userID, "123 Main St", "Kyiv", "UA", "01001", "Home", false)
		require.NoError(t, err)

		// Retrieve all contacts
		contacts, err := uc.GetUserContacts(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, contacts, 3)

		// Verify different types exist
		types := make(map[contact.ContactType]bool)
		for _, c := range contacts {
			types[c.Type] = true
		}
		assert.True(t, types[contact.ContactTypeEmail])
		assert.True(t, types[contact.ContactTypePhone])
		assert.True(t, types[contact.ContactTypeAddress])
	})
}

// TestContactUseCase_GetUserContactsByType tests retrieving contacts by type
func TestContactUseCase_GetUserContactsByType(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create contacts of different types
		_, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("email1_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)
		_, err = uc.CreateEmailContact(ctx, userID, fmt.Sprintf("email2_%s@example.com", uuidv7.New().String()), "Personal", false)
		require.NoError(t, err)
		_, err = uc.CreatePhoneContact(ctx, userID, "+380671234567", "Mobile", false)
		require.NoError(t, err)

		// Retrieve only email contacts
		emailContacts, err := uc.GetUserContactsByType(ctx, userID, contact.ContactTypeEmail)
		require.NoError(t, err)
		assert.Len(t, emailContacts, 2)
		for _, c := range emailContacts {
			assert.Equal(t, contact.ContactTypeEmail, c.Type)
		}

		// Retrieve only phone contacts
		phoneContacts, err := uc.GetUserContactsByType(ctx, userID, contact.ContactTypePhone)
		require.NoError(t, err)
		assert.Len(t, phoneContacts, 1)
		assert.Equal(t, contact.ContactTypePhone, phoneContacts[0].Type)
	})
}

// TestContactUseCase_UpdateContact tests contact updates
func TestContactUseCase_UpdateContact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create contact
		c, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("test_%s@example.com", uuidv7.New().String()), "Work", false)
		require.NoError(t, err)

		// Update label
		c.Label = "Personal"
		err = uc.UpdateContact(ctx, c)
		require.NoError(t, err)

		// Verify update
		retrieved, err := uc.GetContact(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "Personal", retrieved.Label)
	})
}

// TestContactUseCase_DeleteContact tests contact deletion
func TestContactUseCase_DeleteContact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create contact
		c, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("test_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)

		// Delete contact
		err = uc.DeleteContact(ctx, c.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = uc.GetContact(ctx, c.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "contact not found")
	})
}

// TestContactUseCase_SetAsPrimary tests setting contact as primary
func TestContactUseCase_SetAsPrimary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create two email contacts
		c1, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("email1_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)
		c2, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("email2_%s@example.com", uuidv7.New().String()), "Personal", false)
		require.NoError(t, err)

		assert.True(t, c1.IsPrimary)
		assert.False(t, c2.IsPrimary)

		// Set second contact as primary
		err = uc.SetAsPrimary(ctx, c2.ID)
		require.NoError(t, err)

		// Verify first is no longer primary
		c1Retrieved, err := uc.GetContact(ctx, c1.ID)
		require.NoError(t, err)
		assert.False(t, c1Retrieved.IsPrimary)

		// Verify second is now primary
		c2Retrieved, err := uc.GetContact(ctx, c2.ID)
		require.NoError(t, err)
		assert.True(t, c2Retrieved.IsPrimary)
	})
}

// TestContactUseCase_VerifyContact tests contact verification
func TestContactUseCase_VerifyContact(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create contact
		c, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("test_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)
		assert.False(t, c.IsVerified)

		// Verify contact
		err = uc.VerifyContact(ctx, c.ID)
		require.NoError(t, err)

		// Check verification
		retrieved, err := uc.GetContact(ctx, c.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.IsVerified)
	})
}

// TestContactUseCase_UpdateVisibility tests contact visibility updates
func TestContactUseCase_UpdateVisibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// Create contact (default is not public)
		c, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("test_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)
		assert.False(t, c.IsPublic)

		// Make public
		err = uc.UpdateVisibility(ctx, c.ID, true)
		require.NoError(t, err)

		// Verify visibility
		retrieved, err := uc.GetContact(ctx, c.ID)
		require.NoError(t, err)
		assert.True(t, retrieved.IsPublic)

		// Make private again
		err = uc.UpdateVisibility(ctx, c.ID, false)
		require.NoError(t, err)

		// Verify visibility
		retrieved, err = uc.GetContact(ctx, c.ID)
		require.NoError(t, err)
		assert.False(t, retrieved.IsPublic)
	})
}

// TestContactUseCase_CompleteWorkflow tests full contact lifecycle
func TestContactUseCase_CompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		userRepo := userPostgres.NewUserRepository(testDB.DB)
		repo := contactPostgres.NewContactRepository(testDB.DB)
		uc := contact.NewUseCase(repo)

		// Create test user first
		userID := createTestUser(t, ctx, userRepo)

		// 1. Create email contact
		email, err := uc.CreateEmailContact(ctx, userID, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Work", true)
		require.NoError(t, err)

		// 2. Create phone contact
		phone, err := uc.CreatePhoneContact(ctx, userID, "+380671234567", "Mobile", false)
		require.NoError(t, err)

		// 3. Verify email
		err = uc.VerifyContact(ctx, email.ID)
		require.NoError(t, err)

		// 4. Make phone primary
		err = uc.SetAsPrimary(ctx, phone.ID)
		require.NoError(t, err)

		// 5. Make email public
		err = uc.UpdateVisibility(ctx, email.ID, true)
		require.NoError(t, err)

		// 6. Verify final state
		emailFinal, err := uc.GetContact(ctx, email.ID)
		require.NoError(t, err)
		assert.True(t, emailFinal.IsVerified)
		assert.True(t, emailFinal.IsPrimary) // Email remains primary for its type
		assert.True(t, emailFinal.IsPublic)

		phoneFinal, err := uc.GetContact(ctx, phone.ID)
		require.NoError(t, err)
		assert.True(t, phoneFinal.IsPrimary) // Phone is now primary for its type
		assert.False(t, phoneFinal.IsVerified)

		// 7. Get all user contacts
		allContacts, err := uc.GetUserContacts(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, allContacts, 2)

		// 8. Delete phone contact
		err = uc.DeleteContact(ctx, phone.ID)
		require.NoError(t, err)

		// 9. Verify only email remains
		remaining, err := uc.GetUserContacts(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, remaining, 1)
		assert.Equal(t, contact.ContactTypeEmail, remaining[0].Type)
	})
}
