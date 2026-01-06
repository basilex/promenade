package interaction_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	companyRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	interactionRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/adapter/repository/postgres"
	roleRepo "github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/identity/user"
	userRepo "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// setupTestData creates sales rep and customer for interaction tests
func setupTestData(t *testing.T, ctx context.Context, customerUC customer.ICustomerUseCase, userUC user.IUseCase) (*customer.Customer, *user.User) {
	// Create test sales rep user first (required for customer.assignedTo)
	salesRep, err := userUC.Register(ctx, fmt.Sprintf("salesrep_%s@example.com", uuidv7.New().String()), "Sales Rep", "password123")
	require.NoError(t, err)

	// Create test customer (with valid assignedTo)
	testCustomer, err := customerUC.CreateCustomer(ctx, "Test Customer", fmt.Sprintf("customer_%s@example.com", uuidv7.New().String()), "website", salesRep.ID)
	require.NoError(t, err)

	return testCustomer, salesRep
}

// TestInteractionUseCase_CreateInteraction tests interaction creation
func TestInteractionUseCase_CreateInteraction(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	// Setup
	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Use helper to create sales rep and customer
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)

	// Create test user (interaction creator)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil, // No company
		"call",
		"outbound",
		"Follow-up call",
		"Discussed product features",
		testUser.ID,
		startedAt,
	)

	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, i.ID)
	assert.Equal(t, interaction.InteractionTypeCall, i.Type)
	assert.Equal(t, interaction.InteractionDirectionOutbound, i.Direction)
	assert.Equal(t, "Follow-up call", i.Subject)
	assert.Equal(t, "Discussed product features", i.Description)
	assert.Equal(t, testCustomer.ID, i.CustomerID)
	assert.Nil(t, i.CompanyID)
	assert.Equal(t, testUser.ID, i.CreatedBy)
	assert.Equal(t, startedAt.Unix(), i.StartedAt.Unix())
}

// TestInteractionUseCase_CreateInteractionWithCompany tests interaction creation with company
func TestInteractionUseCase_CreateInteractionWithCompany(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	companyRepository := companyRepo.NewCompanyRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	companyUC := company.NewUseCase(companyRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer with valid assignedTo
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)

	// Create test company with unique name
	uuidSuffix := uuidv7.New().String()
	testCompany, err := companyUC.CreateCompany(ctx, fmt.Sprintf("Test Company %s", uuidSuffix), nil, string(company.CompanyTypeLLC), nil, nil, nil, nil, nil, nil, nil, string(company.CompanySizeSmall), 0, 0, "USD", nil, nil)
	require.NoError(t, err)

	// Create test user
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction with company
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		&testCompany.ID,
		"meeting",
		"inbound",
		"Product demo",
		"Showcased new features",
		testUser.ID,
		startedAt,
	)

	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, i.ID)
	assert.NotNil(t, i.CompanyID)
	assert.Equal(t, testCompany.ID, *i.CompanyID)
}

// TestInteractionUseCase_GetInteraction tests retrieving an interaction by ID
func TestInteractionUseCase_GetInteraction(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	created, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil,
		"email",
		"outbound",
		"Test email",
		"Test description",
		testUser.ID,
		startedAt,
	)
	require.NoError(t, err)

	// Retrieve interaction
	retrieved, err := interactionUC.GetInteraction(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "Test email", retrieved.Subject)
}

// TestInteractionUseCase_GetInteraction_NotFound tests error when interaction not found
func TestInteractionUseCase_GetInteraction_NotFound(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	interactionUC := interaction.NewUseCase(interactionRepository)

	ctx := context.Background()

	// Try to retrieve non-existent interaction
	nonExistentID := uuidv7.New()
	_, err := interactionUC.GetInteraction(ctx, nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestInteractionUseCase_UpdateContent tests updating interaction content
func TestInteractionUseCase_UpdateContent(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil,
		"note",
		"outbound",
		"Original subject",
		"Original description",
		testUser.ID,
		startedAt,
	)
	require.NoError(t, err)

	// Update content
	updated, err := interactionUC.UpdateContent(ctx, i.ID, "Updated subject", "Updated description")
	require.NoError(t, err)
	assert.Equal(t, "Updated subject", updated.Subject)
	assert.Equal(t, "Updated description", updated.Description)
}

// TestInteractionUseCase_SetOutcome tests setting interaction outcome
func TestInteractionUseCase_SetOutcome(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil,
		"call",
		"outbound",
		"Sales call",
		"Pitch product",
		testUser.ID,
		startedAt,
	)
	require.NoError(t, err)

	// Set outcome
	updated, err := interactionUC.SetOutcome(ctx, i.ID, "successful")
	require.NoError(t, err)
	assert.NotNil(t, updated.Outcome)
	if updated.Outcome != nil {
		assert.Equal(t, interaction.InteractionOutcomeSuccessful, *updated.Outcome)
	}
}

// TestInteractionUseCase_EndInteraction tests ending an interaction
func TestInteractionUseCase_EndInteraction(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil,
		"call",
		"outbound",
		"Sales call",
		"Call description",
		testUser.ID,
		startedAt,
	)
	require.NoError(t, err)

	// End interaction
	endedAt := time.Now().Add(10 * time.Minute)
	ended, err := interactionUC.EndInteraction(ctx, i.ID, endedAt)
	require.NoError(t, err)
	assert.NotNil(t, ended.EndedAt)
	assert.Equal(t, endedAt.Unix(), ended.EndedAt.Unix())
	require.NotNil(t, ended.DurationSec)
	if ended.DurationSec != nil {
		assert.Greater(t, *ended.DurationSec, 0) // DurationSec is *int, not *int64
	}
}

// TestInteractionUseCase_SetFollowUp tests setting follow-up for an interaction
func TestInteractionUseCase_SetFollowUp(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil,
		"call",
		"outbound",
		"Follow-up call",
		"Initial call",
		testUser.ID,
		startedAt,
	)
	require.NoError(t, err)

	// Set follow-up
	followUpDate := time.Now().Add(7 * 24 * time.Hour)
	updated, err := interactionUC.SetFollowUp(ctx, i.ID, true, &followUpDate, "Follow up next week")
	require.NoError(t, err)
	assert.True(t, updated.FollowUpRequired)
	assert.NotNil(t, updated.FollowUpDate)
	assert.Equal(t, followUpDate.Unix(), updated.FollowUpDate.Unix())
	assert.Equal(t, "Follow up next week", updated.FollowUpNotes)
}

// TestInteractionUseCase_AddAttendee tests adding an attendee to an interaction
func TestInteractionUseCase_AddAttendee(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create another user (attendee)
	attendeeUser, err := userUC.Register(ctx, fmt.Sprintf("attendee_%s@example.com", uuidv7.New().String()), "Attendee User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil,
		"meeting",
		"inbound",
		"Team meeting",
		"Discuss strategy",
		testUser.ID,
		startedAt,
	)
	require.NoError(t, err)

	// Add attendee
	updated, err := interactionUC.AddAttendee(ctx, i.ID, attendeeUser.ID)
	require.NoError(t, err)
	assert.Len(t, updated.Attendees, 1)
	assert.Equal(t, attendeeUser.ID, updated.Attendees[0])
}

// TestInteractionUseCase_RemoveAttendee tests removing an attendee from an interaction
func TestInteractionUseCase_RemoveAttendee(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and users
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)
	attendeeUser, err := userUC.Register(ctx, fmt.Sprintf("attendee_%s@example.com", uuidv7.New().String()), "Attendee User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(
		ctx,
		testCustomer.ID,
		nil,
		"meeting",
		"inbound",
		"Team meeting",
		"Discuss strategy",
		testUser.ID,
		startedAt,
	)
	require.NoError(t, err)

	// Add attendee
	withAttendee, err := interactionUC.AddAttendee(ctx, i.ID, attendeeUser.ID)
	require.NoError(t, err)
	assert.Len(t, withAttendee.Attendees, 1)

	// Remove attendee
	withoutAttendee, err := interactionUC.RemoveAttendee(ctx, i.ID, attendeeUser.ID)
	require.NoError(t, err)
	assert.Len(t, withoutAttendee.Attendees, 0)
}

// TestInteractionUseCase_ListByCustomer tests listing interactions by customer
func TestInteractionUseCase_ListByCustomer(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create multiple interactions
	startedAt := time.Now()
	for i := 0; i < 3; i++ {
		_, err := interactionUC.CreateInteraction(
			ctx,
			testCustomer.ID,
			nil,
			"call",
			"outbound",
			fmt.Sprintf("Call %d", i+1),
			"Description",
			testUser.ID,
			startedAt.Add(time.Duration(i)*time.Hour),
		)
		require.NoError(t, err)
	}

	// List interactions
	interactions, total, err := interactionUC.ListByCustomer(ctx, testCustomer.ID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, interactions, 3)
}

// TestInteractionUseCase_ListByType tests listing interactions by type
func TestInteractionUseCase_ListByType(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interactions of different types
	startedAt := time.Now()
	_, err = interactionUC.CreateInteraction(ctx, testCustomer.ID, nil, "call", "outbound", "Call 1", "Desc", testUser.ID, startedAt)
	require.NoError(t, err)
	_, err = interactionUC.CreateInteraction(ctx, testCustomer.ID, nil, "call", "outbound", "Call 2", "Desc", testUser.ID, startedAt)
	require.NoError(t, err)
	_, err = interactionUC.CreateInteraction(ctx, testCustomer.ID, nil, "email", "outbound", "Email 1", "Desc", testUser.ID, startedAt)
	require.NoError(t, err)

	// List call interactions
	interactions, total, err := interactionUC.ListByType(ctx, "call", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, interactions, 2)
	for _, i := range interactions {
		assert.Equal(t, interaction.InteractionTypeCall, i.Type)
	}
}

// TestInteractionUseCase_ListPendingFollowUps tests listing pending follow-ups
func TestInteractionUseCase_ListPendingFollowUps(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction with follow-up
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(ctx, testCustomer.ID, nil, string(interaction.InteractionTypeCall), string(interaction.InteractionDirectionOutbound), "Call 1", "Desc", testUser.ID, startedAt)
	require.NoError(t, err)

	followUpDate := time.Now().Add(-1 * time.Hour) // Past date for pending follow-ups
	_, err = interactionUC.SetFollowUp(ctx, i.ID, true, &followUpDate, "Follow up needed")
	require.NoError(t, err)

	// List pending follow-ups
	interactions, total, err := interactionUC.ListPendingFollowUps(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, interactions, 1)
	if len(interactions) > 0 {
		assert.True(t, interactions[0].FollowUpRequired)
	}
}

// TestInteractionUseCase_DeleteInteraction tests soft deleting an interaction
func TestInteractionUseCase_DeleteInteraction(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	defer db.Cleanup()

	interactionRepository := interactionRepo.NewInteractionRepository(db.DB)
	customerRepository := customerRepo.NewCustomerRepository(db.DB)
	userRepository := userRepo.NewUserRepository(db.DB)
	roleRepository := roleRepo.NewRoleRepository(db.DB)

	interactionUC := interaction.NewUseCase(interactionRepository)
	customerUC := customer.NewUseCase(customerRepository)
	userUC := user.NewUseCase(userRepository, roleRepository)

	ctx := context.Background()

	// Create test customer and user
	testCustomer, _ := setupTestData(t, ctx, customerUC, userUC)
	testUser, err := userUC.Register(ctx, fmt.Sprintf("user_%s@example.com", uuidv7.New().String()), "Test User", "password123")
	require.NoError(t, err)

	// Create interaction
	startedAt := time.Now()
	i, err := interactionUC.CreateInteraction(ctx, testCustomer.ID, nil, "call", "outbound", "Call 1", "Desc", testUser.ID, startedAt)
	require.NoError(t, err)

	// Delete interaction
	err = interactionUC.DeleteInteraction(ctx, i.ID)
	require.NoError(t, err)

	// Verify deletion (should not be found)
	_, err = interactionUC.GetInteraction(ctx, i.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
