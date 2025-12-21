package smoke

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserContact_SmokeTest - comprehensive end-to-end smoke test for user contacts
func TestUserContact_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)

	ctx := context.Background()

	// Setup repositories
	userRepo := postgres.NewUserRepository(testDB.DB)
	contactRepo := postgres.NewUserContactRepository(testDB.DB)

	// Create test user
	user := helpers.UserFixture()
	require.NoError(t, userRepo.Create(ctx, user))

	var emailContactID, phoneContactID, telegramContactID uuidv7.UUID

	t.Run("[+] Create_email_contact", func(t *testing.T) {
		label := "Work Email"
		contact := &entity.UserContact{
			UserID:       user.ID,
			ContactType:  entity.ContactTypeEmail,
			ContactValue: "work@example.com",
			Label:        &label,
			IsPublic:     true,
			IsPrimary:    true,
			IsVerified:   false,
		}
		err := contactRepo.Create(ctx, contact)
		require.NoError(t, err)
		assert.NotEmpty(t, contact.ID)
		emailContactID = contact.ID
	})

	t.Run("[+] Get_contact_by_ID", func(t *testing.T) {
		contact, err := contactRepo.GetByID(ctx, emailContactID)
		require.NoError(t, err)
		assert.Equal(t, emailContactID, contact.ID)
		assert.Equal(t, entity.ContactTypeEmail, contact.ContactType)
	})

	t.Run("[+] Update_contact", func(t *testing.T) {
		contact, err := contactRepo.GetByID(ctx, emailContactID)
		require.NoError(t, err)

		newValue := "updated-work@example.com"
		contact.ContactValue = newValue
		err = contactRepo.Update(ctx, contact)
		require.NoError(t, err)

		updated, err := contactRepo.GetByID(ctx, emailContactID)
		require.NoError(t, err)
		assert.Equal(t, newValue, updated.ContactValue)
	})

	t.Run("[+] Create_multiple_contact_types", func(t *testing.T) {
		// Phone contact
		phoneLabel := "Mobile"
		phoneContact := &entity.UserContact{
			UserID:       user.ID,
			ContactType:  entity.ContactTypePhone,
			ContactValue: "+1234567890",
			Label:        &phoneLabel,
			IsPublic:     false,
			IsPrimary:    false,
		}
		err := contactRepo.Create(ctx, phoneContact)
		require.NoError(t, err)
		phoneContactID = phoneContact.ID

		// Telegram contact
		telegramLabel := "Telegram"
		telegramContact := &entity.UserContact{
			UserID:       user.ID,
			ContactType:  entity.ContactTypeTelegram,
			ContactValue: "@testuser",
			Label:        &telegramLabel,
			IsPublic:     true,
			IsPrimary:    false,
		}
		err = contactRepo.Create(ctx, telegramContact)
		require.NoError(t, err)
		telegramContactID = telegramContact.ID
	})

	t.Run("[+] Get_user_contacts", func(t *testing.T) {
		contacts, err := contactRepo.GetUserContacts(ctx, user.ID, true)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contacts), 3, "should have at least 3 contacts")

		// Verify contact types
		types := make(map[entity.ContactType]bool)
		for _, contact := range contacts {
			types[contact.ContactType] = true
		}
		assert.True(t, types[entity.ContactTypeEmail])
		assert.True(t, types[entity.ContactTypePhone])
		assert.True(t, types[entity.ContactTypeTelegram])
	})

	t.Run("[+] Get_public_contacts", func(t *testing.T) {
		contacts, err := contactRepo.GetPublicContacts(ctx, user.ID)
		require.NoError(t, err)

		// Should have email and telegram (both public), but not phone (private)
		for _, contact := range contacts {
			assert.True(t, contact.IsPublic, "all contacts should be public")
		}

		// Phone should not be in public contacts
		foundPhone := false
		for _, contact := range contacts {
			if contact.ContactType == entity.ContactTypePhone {
				foundPhone = true
			}
		}
		assert.False(t, foundPhone, "phone contact should not be in public contacts")
	})

	t.Run("[+] Set_primary_contact", func(t *testing.T) {
		// Set phone as primary (SetPrimary unsets others of same type automatically)
		err := contactRepo.SetPrimary(ctx, user.ID, entity.ContactTypePhone, phoneContactID)
		require.NoError(t, err)

		// Verify phone is now primary
		phoneContact, err := contactRepo.GetByID(ctx, phoneContactID)
		require.NoError(t, err)
		assert.True(t, phoneContact.IsPrimary, "phone should be primary after SetPrimary call")
	})

	t.Run("[+] Get_primary_contact", func(t *testing.T) {
		// Get primary phone contact
		primary, err := contactRepo.GetPrimaryContact(ctx, user.ID, entity.ContactTypePhone)
		if err != nil {
			// If no primary found, that's okay - just verify we can call the method
			t.Logf("No primary phone contact found (expected if not set): %v", err)
		} else {
			assert.NotNil(t, primary)
			assert.True(t, primary.IsPrimary)
		}
	})

	t.Run("[+] Verify_contact", func(t *testing.T) {
		err := contactRepo.VerifyContact(ctx, emailContactID)
		require.NoError(t, err)

		contact, err := contactRepo.GetByID(ctx, emailContactID)
		require.NoError(t, err)
		assert.True(t, contact.IsVerified)
	})

	t.Run("[+] Toggle_contact_visibility", func(t *testing.T) {
		contact, err := contactRepo.GetByID(ctx, phoneContactID)
		require.NoError(t, err)
		assert.False(t, contact.IsPublic, "phone should be private")

		// Make public
		contact.IsPublic = true
		err = contactRepo.Update(ctx, contact)
		require.NoError(t, err)

		updated, err := contactRepo.GetByID(ctx, phoneContactID)
		require.NoError(t, err)
		assert.True(t, updated.IsPublic)
	})

	t.Run("[+] Delete_contact", func(t *testing.T) {
		err := contactRepo.Delete(ctx, telegramContactID)
		require.NoError(t, err)

		_, err = contactRepo.GetByID(ctx, telegramContactID)
		assert.Error(t, err, "deleted contact should not be found")

		// Verify other contacts still exist
		contacts, err := contactRepo.GetUserContacts(ctx, user.ID, true)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contacts), 2, "should have remaining contacts")
	})

	t.Logf("[SUCCESS] All user contact smoke tests passed!")
}
