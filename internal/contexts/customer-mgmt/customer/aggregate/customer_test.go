package aggregate

import (
	"errors"
	"testing"
	"time"

	customererrors "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// Factory Methods Tests
// ============================================================================

func TestNewCustomer_Success(t *testing.T) {
	assignedTo := uuidv7.New()
	customer, err := NewCustomer("John Doe", "john@example.com", "website", assignedTo)

	assert.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, customer.ID)
	assert.Equal(t, "John Doe", customer.Name)
	assert.Equal(t, "john@example.com", customer.Email.Value())
	assert.Equal(t, CustomerStatusLead, customer.Status)
	assert.Equal(t, CustomerTierFree, customer.Tier)
	assert.Equal(t, "website", customer.Source)
	assert.Equal(t, assignedTo, customer.AssignedTo)
	assert.Nil(t, customer.Phone)
	assert.Nil(t, customer.UserID)
	assert.Nil(t, customer.CompanyID)
	assert.Empty(t, customer.Tags)
	assert.False(t, customer.CreatedAt.IsZero())
	assert.False(t, customer.UpdatedAt.IsZero())
	assert.Nil(t, customer.LastContactedAt)
	assert.Nil(t, customer.ConvertedAt)
	assert.Nil(t, customer.ChurnedAt)
}

func TestNewCustomer_InvalidEmail(t *testing.T) {
	assignedTo := uuidv7.New()

	tests := []struct {
		name  string
		email string
	}{
		{"empty email", ""},
		{"invalid format", "not-an-email"},
		{"missing domain", "user@"},
		{"missing @", "user.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := NewCustomer("John Doe", tt.email, "website", assignedTo)

			assert.Error(t, err)
			assert.Nil(t, customer)
		})
	}
}

func TestNewCustomer_EmptyFields(t *testing.T) {
	tests := []struct {
		name       string
		custName   string
		email      string
		source     string
		assignedTo uuidv7.UUID
		wantErr    bool
	}{
		{"empty name", "", "john@example.com", "website", uuidv7.New(), true},
		{"empty source", "John Doe", "john@example.com", "", uuidv7.New(), true},
		{"nil assigned to", "John Doe", "john@example.com", "website", uuidv7.Nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := NewCustomer(tt.custName, tt.email, tt.source, tt.assignedTo)

			assert.Error(t, err)
			assert.Nil(t, customer)
		})
	}
}

func TestNewB2BCustomer_Success(t *testing.T) {
	companyID := uuidv7.New()
	assignedTo := uuidv7.New()

	customer, err := NewB2BCustomer("Acme Corp", "contact@acme.com", "referral", companyID, assignedTo)

	assert.NoError(t, err)
	assert.NotNil(t, customer)
	assert.Equal(t, "Acme Corp", customer.Name)
	assert.Equal(t, companyID, *customer.CompanyID)
	assert.True(t, customer.IsB2B())
}

func TestNewB2BCustomer_NilCompanyID(t *testing.T) {
	assignedTo := uuidv7.New()

	customer, err := NewB2BCustomer("Acme Corp", "contact@acme.com", "referral", uuidv7.Nil, assignedTo)

	assert.Error(t, err)
	assert.Nil(t, customer)
	assert.True(t, errors.Is(err, customererrors.ErrCustomerCompanyIDEmpty))
}

// ============================================================================
// Status Transition Tests
// ============================================================================

func TestCustomer_QualifyAsProspect_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	err := customer.QualifyAsProspect()

	assert.NoError(t, err)
	assert.Equal(t, CustomerStatusProspect, customer.Status)
}

func TestCustomer_QualifyAsProspect_InvalidTransition(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus CustomerStatus
	}{
		{"from prospect", CustomerStatusProspect},
		{"from customer", CustomerStatusCustomer},
		{"from churned", CustomerStatusChurned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Status = tt.initialStatus

			err := customer.QualifyAsProspect()

			assert.Error(t, err)
			assert.Equal(t, tt.initialStatus, customer.Status)
		})
	}
}

func TestCustomer_ConvertToCustomer_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	customer.Status = CustomerStatusProspect

	err := customer.ConvertToCustomer()

	assert.NoError(t, err)
	assert.Equal(t, CustomerStatusCustomer, customer.Status)
	assert.NotNil(t, customer.ConvertedAt)
	assert.False(t, customer.ConvertedAt.IsZero())
}

func TestCustomer_ConvertToCustomer_InvalidTransition(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus CustomerStatus
	}{
		{"from lead", CustomerStatusLead},
		{"from customer", CustomerStatusCustomer},
		{"from churned", CustomerStatusChurned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Status = tt.initialStatus

			err := customer.ConvertToCustomer()

			assert.Error(t, err)
			assert.Equal(t, tt.initialStatus, customer.Status)
			assert.Nil(t, customer.ConvertedAt)
		})
	}
}

func TestCustomer_Churn_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	customer.Status = CustomerStatusCustomer
	customer.ConvertedAt = timePtr(time.Now().Add(-30 * 24 * time.Hour))

	err := customer.Churn("moved to competitor")

	assert.NoError(t, err)
	assert.Equal(t, CustomerStatusChurned, customer.Status)
	assert.NotNil(t, customer.ChurnedAt)
	assert.Equal(t, "moved to competitor", customer.ChurnReason)
}

func TestCustomer_Churn_InvalidTransition(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus CustomerStatus
	}{
		{"from lead", CustomerStatusLead},
		{"from prospect", CustomerStatusProspect},
		{"from churned", CustomerStatusChurned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Status = tt.initialStatus

			err := customer.Churn("some reason")

			assert.Error(t, err)
			assert.Equal(t, tt.initialStatus, customer.Status)
			assert.Nil(t, customer.ChurnedAt)
		})
	}
}

func TestCustomer_Churn_EmptyReason(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	customer.Status = CustomerStatusCustomer

	err := customer.Churn("")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, customererrors.ErrCustomerChurnReasonEmpty))
	assert.Equal(t, CustomerStatusCustomer, customer.Status)
}

func TestCustomer_Reactivate_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	customer.Status = CustomerStatusChurned
	customer.ChurnedAt = timePtr(time.Now().Add(-10 * 24 * time.Hour))
	customer.ChurnReason = "moved to competitor"

	err := customer.Reactivate()

	assert.NoError(t, err)
	assert.Equal(t, CustomerStatusCustomer, customer.Status)
	assert.Nil(t, customer.ChurnedAt)
	assert.Empty(t, customer.ChurnReason)
}

func TestCustomer_Reactivate_InvalidTransition(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus CustomerStatus
	}{
		{"from lead", CustomerStatusLead},
		{"from prospect", CustomerStatusProspect},
		{"from customer", CustomerStatusCustomer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Status = tt.initialStatus

			err := customer.Reactivate()

			assert.Error(t, err)
			assert.Equal(t, tt.initialStatus, customer.Status)
		})
	}
}

// ============================================================================
// Tier Management Tests
// ============================================================================

func TestCustomer_UpgradeTier_Success(t *testing.T) {
	tests := []struct {
		name        string
		currentTier CustomerTier
		newTier     CustomerTier
	}{
		{"free to basic", CustomerTierFree, CustomerTierBasic},
		{"free to pro", CustomerTierFree, CustomerTierPro},
		{"free to enterprise", CustomerTierFree, CustomerTierEnterprise},
		{"basic to pro", CustomerTierBasic, CustomerTierPro},
		{"basic to enterprise", CustomerTierBasic, CustomerTierEnterprise},
		{"pro to enterprise", CustomerTierPro, CustomerTierEnterprise},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Tier = tt.currentTier

			err := customer.UpgradeTier(tt.newTier)

			assert.NoError(t, err)
			assert.Equal(t, tt.newTier, customer.Tier)
		})
	}
}

func TestCustomer_UpgradeTier_InvalidUpgrade(t *testing.T) {
	tests := []struct {
		name        string
		currentTier CustomerTier
		newTier     CustomerTier
	}{
		{"basic to free", CustomerTierBasic, CustomerTierFree},
		{"pro to free", CustomerTierPro, CustomerTierFree},
		{"pro to basic", CustomerTierPro, CustomerTierBasic},
		{"enterprise to pro", CustomerTierEnterprise, CustomerTierPro},
		{"same tier", CustomerTierBasic, CustomerTierBasic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Tier = tt.currentTier

			err := customer.UpgradeTier(tt.newTier)

			assert.Error(t, err)
			assert.Equal(t, tt.currentTier, customer.Tier)
		})
	}
}

func TestCustomer_DowngradeTier_Success(t *testing.T) {
	tests := []struct {
		name        string
		currentTier CustomerTier
		newTier     CustomerTier
	}{
		{"basic to free", CustomerTierBasic, CustomerTierFree},
		{"pro to basic", CustomerTierPro, CustomerTierBasic},
		{"pro to free", CustomerTierPro, CustomerTierFree},
		{"enterprise to pro", CustomerTierEnterprise, CustomerTierPro},
		{"enterprise to basic", CustomerTierEnterprise, CustomerTierBasic},
		{"enterprise to free", CustomerTierEnterprise, CustomerTierFree},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Tier = tt.currentTier

			err := customer.DowngradeTier(tt.newTier)

			assert.NoError(t, err)
			assert.Equal(t, tt.newTier, customer.Tier)
		})
	}
}

func TestCustomer_DowngradeTier_InvalidDowngrade(t *testing.T) {
	tests := []struct {
		name        string
		currentTier CustomerTier
		newTier     CustomerTier
	}{
		{"free to basic", CustomerTierFree, CustomerTierBasic},
		{"basic to pro", CustomerTierBasic, CustomerTierPro},
		{"same tier", CustomerTierBasic, CustomerTierBasic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			customer.Tier = tt.currentTier

			err := customer.DowngradeTier(tt.newTier)

			assert.Error(t, err)
			assert.Equal(t, tt.currentTier, customer.Tier)
		})
	}
}

// ============================================================================
// Tag Management Tests
// ============================================================================

func TestCustomer_AddTag_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	err := customer.AddTag("vip")

	assert.NoError(t, err)
	assert.Contains(t, customer.Tags, "vip")
	assert.Len(t, customer.Tags, 1)
}

func TestCustomer_AddTag_Multiple(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	_ = customer.AddTag("vip")
	_ = customer.AddTag("enterprise")
	_ = customer.AddTag("priority")

	assert.Len(t, customer.Tags, 3)
	assert.Contains(t, customer.Tags, "vip")
	assert.Contains(t, customer.Tags, "enterprise")
	assert.Contains(t, customer.Tags, "priority")
}

func TestCustomer_AddTag_Duplicate(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	_ = customer.AddTag("vip")
	err := customer.AddTag("vip")

	assert.Error(t, err)
	assert.Len(t, customer.Tags, 1)
	assert.Equal(t, customererrors.ErrCustomerTagAlreadyExists, err)
}

func TestCustomer_AddTag_Empty(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	err := customer.AddTag("")

	assert.Error(t, err)
	assert.Empty(t, customer.Tags)
}

func TestCustomer_RemoveTag_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	_ = customer.AddTag("vip")
	_ = customer.AddTag("priority")

	err := customer.RemoveTag("vip")

	assert.NoError(t, err)
	assert.NotContains(t, customer.Tags, "vip")
	assert.Contains(t, customer.Tags, "priority")
	assert.Len(t, customer.Tags, 1)
}

func TestCustomer_RemoveTag_NonExistent(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	_ = customer.AddTag("vip")

	err := customer.RemoveTag("enterprise")

	assert.Error(t, err)
	assert.Equal(t, customererrors.ErrCustomerTagNotFound, err)
	assert.Len(t, customer.Tags, 1)
}

// ============================================================================
// Business Methods Tests
// ============================================================================

func TestCustomer_SetPhone_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	err := customer.SetPhone("+380501234567")

	assert.NoError(t, err)
	assert.NotNil(t, customer.Phone)
	assert.Equal(t, "+380501234567", customer.Phone.Value())
}

func TestCustomer_SetPhone_Invalid(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	err := customer.SetPhone("invalid-phone")

	assert.Error(t, err)
	assert.Nil(t, customer.Phone)
}

func TestCustomer_SetPhone_Clear(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	phone, _ := valueobject.NewPhone("+380501234567")
	customer.Phone = &phone

	err := customer.SetPhone("")

	assert.NoError(t, err)
	assert.Nil(t, customer.Phone)
}

func TestCustomer_LinkToUser_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	userID := uuidv7.New()

	err := customer.LinkToUser(userID)

	assert.NoError(t, err)
	assert.NotNil(t, customer.UserID)
	assert.Equal(t, userID, *customer.UserID)
	assert.True(t, customer.HasAccount())
}

func TestCustomer_LinkToUser_NilUserID(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	err := customer.LinkToUser(uuidv7.Nil)

	assert.Error(t, err)
	assert.Nil(t, customer.UserID)
}

func TestCustomer_LinkToUser_AlreadyLinked(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
	userID1 := uuidv7.New()
	userID2 := uuidv7.New()
	customer.UserID = &userID1

	err := customer.LinkToUser(userID2)

	assert.Error(t, err)
	assert.Equal(t, userID1, *customer.UserID)
	assert.True(t, errors.Is(err, customererrors.ErrCustomerUserAlreadyLinked))
}

func TestCustomer_Reassign_Success(t *testing.T) {
	oldRep := uuidv7.New()
	newRep := uuidv7.New()
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", oldRep)

	err := customer.Reassign(newRep)

	assert.NoError(t, err)
	assert.Equal(t, newRep, customer.AssignedTo)
}

func TestCustomer_Reassign_NilRepID(t *testing.T) {
	oldRep := uuidv7.New()
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", oldRep)

	err := customer.Reassign(uuidv7.Nil)

	assert.Error(t, err)
	assert.Equal(t, oldRep, customer.AssignedTo)
}

// ============================================================================
// Query Methods Tests
// ============================================================================

func TestCustomer_StatusChecks(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	// Lead status
	assert.True(t, customer.IsLead())
	assert.False(t, customer.IsProspect())
	assert.False(t, customer.IsActiveCustomer())
	assert.False(t, customer.IsChurned())

	// Prospect status
	customer.Status = CustomerStatusProspect
	assert.False(t, customer.IsLead())
	assert.True(t, customer.IsProspect())
	assert.False(t, customer.IsActiveCustomer())
	assert.False(t, customer.IsChurned())

	// Customer status
	customer.Status = CustomerStatusCustomer
	assert.False(t, customer.IsLead())
	assert.False(t, customer.IsProspect())
	assert.True(t, customer.IsActiveCustomer())
	assert.False(t, customer.IsChurned())

	// Churned status
	customer.Status = CustomerStatusChurned
	assert.False(t, customer.IsLead())
	assert.False(t, customer.IsProspect())
	assert.False(t, customer.IsActiveCustomer())
	assert.True(t, customer.IsChurned())
}

func TestCustomer_IsB2B(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	// B2C customer (no company)
	assert.False(t, customer.IsB2B())

	// B2B customer (has company)
	companyID := uuidv7.New()
	customer.CompanyID = &companyID
	assert.True(t, customer.IsB2B())
}

func TestCustomer_HasAccount(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	// No account linked
	assert.False(t, customer.HasAccount())

	// Account linked
	userID := uuidv7.New()
	customer.UserID = &userID
	assert.True(t, customer.HasAccount())
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestCustomer_Validate_Success(t *testing.T) {
	customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())

	err := customer.Validate()

	assert.NoError(t, err)
}

func TestCustomer_Validate_InvalidFields(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Customer)
		errorMsg string
	}{
		{
			name: "empty name",
			setup: func(c *Customer) {
				c.Name = ""
			},
			errorMsg: "customer name cannot be empty",
		},
		{
			name: "invalid status",
			setup: func(c *Customer) {
				c.Status = "invalid"
			},
			errorMsg: "invalid status",
		},
		{
			name: "invalid tier",
			setup: func(c *Customer) {
				c.Tier = "invalid"
			},
			errorMsg: "invalid tier",
		},
		{
			name: "empty source",
			setup: func(c *Customer) {
				c.Source = ""
			},
			errorMsg: "source is required",
		},
		{
			name: "nil assigned to",
			setup: func(c *Customer) {
				c.AssignedTo = uuidv7.Nil
			},
			errorMsg: "assigned sales rep is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, _ := NewCustomer("John Doe", "john@example.com", "website", uuidv7.New())
			tt.setup(customer)

			err := customer.Validate()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

// ============================================================================
// Helper Functions Tests
// ============================================================================

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status CustomerStatus
		valid  bool
	}{
		{CustomerStatusLead, true},
		{CustomerStatusProspect, true},
		{CustomerStatusCustomer, true},
		{CustomerStatusChurned, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := isValidStatus(tt.status)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidTier(t *testing.T) {
	tests := []struct {
		tier  CustomerTier
		valid bool
	}{
		{CustomerTierFree, true},
		{CustomerTierBasic, true},
		{CustomerTierPro, true},
		{CustomerTierEnterprise, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			result := isValidTier(tt.tier)
			assert.Equal(t, tt.valid, result)
		})
	}
}

// ============================================================================
// Test Helpers
// ============================================================================

func timePtr(t time.Time) *time.Time {
	return &t
}
