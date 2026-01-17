package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/aggregate"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

func TestToCustomerResponse(t *testing.T) {
	assignedTo := uuidv7.New()

	t.Run("B2C customer with all fields", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		c.ID = uuidv7.New()
		phone, _ := valueobject.NewPhone("+380501234567")
		c.Phone = &phone
		c.Tags = []string{"vip", "premium"}
		_ = c.QualifyAsProspect()

		resp := ToCustomerResponse(c)

		assert.Equal(t, c.ID.String(), resp.ID)
		assert.Equal(t, "John Doe", resp.Name)
		assert.Equal(t, "john@example.com", resp.Email)
		assert.Equal(t, "prospect", resp.Status)
		assert.Equal(t, "free", resp.Tier)
		assert.Equal(t, "website", resp.Source)
		assert.Equal(t, assignedTo.String(), resp.AssignedTo)
		require.NotNil(t, resp.Phone)
		assert.Equal(t, "+380501234567", *resp.Phone)
		assert.Len(t, resp.Tags, 2)
		assert.False(t, resp.IsLead)
		assert.True(t, resp.IsProspect)
		assert.False(t, resp.IsActiveCustomer)
		assert.False(t, resp.IsChurned)
		assert.False(t, resp.IsB2B)
		assert.False(t, resp.HasAccount)
	})

	t.Run("B2B customer with company", func(t *testing.T) {
		companyID := uuidv7.New()
		c, _ := aggregate.NewB2BCustomer("Jane Smith", "jane@company.com", "referral", companyID, assignedTo)
		c.ID = uuidv7.New()

		resp := ToCustomerResponse(c)

		assert.Equal(t, c.ID.String(), resp.ID)
		assert.Equal(t, "Jane Smith", resp.Name)
		require.NotNil(t, resp.CompanyID)
		assert.Equal(t, companyID.String(), *resp.CompanyID)
		assert.True(t, resp.IsB2B)
		assert.Nil(t, resp.UserID)
	})

	t.Run("customer with user account", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Bob Test", "bob@example.com", "api", assignedTo)
		c.ID = uuidv7.New()
		userID := uuidv7.New()
		_ = c.LinkToUser(userID)

		resp := ToCustomerResponse(c)

		require.NotNil(t, resp.UserID)
		assert.Equal(t, userID.String(), *resp.UserID)
		assert.True(t, resp.HasAccount)
	})

	t.Run("converted customer", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Alice Buyer", "alice@example.com", "website", assignedTo)
		c.ID = uuidv7.New()
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()

		resp := ToCustomerResponse(c)

		assert.Equal(t, "customer", resp.Status)
		assert.False(t, resp.IsLead)
		assert.False(t, resp.IsProspect)
		assert.True(t, resp.IsActiveCustomer)
		assert.False(t, resp.IsChurned)
		assert.NotNil(t, resp.ConvertedAt)
	})

	t.Run("churned customer", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Lost Customer", "lost@example.com", "website", assignedTo)
		c.ID = uuidv7.New()
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		_ = c.Churn("poor service")

		resp := ToCustomerResponse(c)

		assert.Equal(t, "churned", resp.Status)
		assert.False(t, resp.IsActiveCustomer)
		assert.True(t, resp.IsChurned)
		assert.NotNil(t, resp.ChurnedAt)
		assert.Equal(t, "poor service", resp.ChurnReason)
	})

	t.Run("customer without phone", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("No Phone", "nophone@example.com", "website", assignedTo)
		c.ID = uuidv7.New()

		resp := ToCustomerResponse(c)

		assert.Nil(t, resp.Phone)
	})

	t.Run("customer with upgraded tier", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Premium User", "premium@example.com", "website", assignedTo)
		c.ID = uuidv7.New()
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		_ = c.UpgradeTier(aggregate.CustomerTierPro)

		resp := ToCustomerResponse(c)

		assert.Equal(t, "pro", resp.Tier)
	})
}

func TestToCustomerListResponse(t *testing.T) {
	assignedTo := uuidv7.New()

	t.Run("empty list", func(t *testing.T) {
		customers := []*aggregate.Customer{}
		resp := ToCustomerListResponse(customers)

		assert.Empty(t, resp)
		assert.NotNil(t, resp)
	})

	t.Run("multiple customers", func(t *testing.T) {
		c1, _ := aggregate.NewCustomer("Customer 1", "c1@example.com", "website", assignedTo)
		c1.ID = uuidv7.New()

		c2, _ := aggregate.NewCustomer("Customer 2", "c2@example.com", "api", assignedTo)
		c2.ID = uuidv7.New()
		_ = c2.QualifyAsProspect()

		c3, _ := aggregate.NewCustomer("Customer 3", "c3@example.com", "referral", assignedTo)
		c3.ID = uuidv7.New()
		_ = c3.QualifyAsProspect()
		_ = c3.ConvertToCustomer()

		customers := []*aggregate.Customer{c1, c2, c3}
		resp := ToCustomerListResponse(customers)

		require.Len(t, resp, 3)
		assert.Equal(t, c1.ID.String(), resp[0].ID)
		assert.Equal(t, c2.ID.String(), resp[1].ID)
		assert.Equal(t, c3.ID.String(), resp[2].ID)
		assert.Equal(t, "lead", resp[0].Status)
		assert.Equal(t, "prospect", resp[1].Status)
		assert.Equal(t, "customer", resp[2].Status)
	})

	t.Run("mixed B2C and B2B", func(t *testing.T) {
		c1, _ := aggregate.NewCustomer("B2C Customer", "b2c@example.com", "website", assignedTo)
		c1.ID = uuidv7.New()

		companyID := uuidv7.New()
		c2, _ := aggregate.NewB2BCustomer("B2B Customer", "b2b@company.com", "referral", companyID, assignedTo)
		c2.ID = uuidv7.New()

		customers := []*aggregate.Customer{c1, c2}
		resp := ToCustomerListResponse(customers)

		require.Len(t, resp, 2)
		assert.False(t, resp[0].IsB2B)
		assert.Nil(t, resp[0].CompanyID)
		assert.True(t, resp[1].IsB2B)
		assert.NotNil(t, resp[1].CompanyID)
	})
}

func TestToCustomerStatsResponse(t *testing.T) {
	t.Run("complete stats", func(t *testing.T) {
		stats := &usecase.CustomerStats{
			TotalCustomers:   100,
			ActiveCustomers:  80,
			ChurnedCustomers: 20,
			ByStatus: map[aggregate.CustomerStatus]int{
				aggregate.CustomerStatusLead:     30,
				aggregate.CustomerStatusProspect: 25,
				aggregate.CustomerStatusCustomer: 25,
				aggregate.CustomerStatusChurned:  20,
			},
			ByTier: map[aggregate.CustomerTier]int{
				aggregate.CustomerTierFree:       60,
				aggregate.CustomerTierBasic:      20,
				aggregate.CustomerTierPro:        15,
				aggregate.CustomerTierEnterprise: 5,
			},
		}

		resp := ToCustomerStatsResponse(stats)

		assert.Equal(t, 100, resp.TotalCustomers)
		assert.Equal(t, 80, resp.ActiveCustomers)
		assert.Equal(t, 20, resp.ChurnedCustomers)
		require.Len(t, resp.ByStatus, 4)
		assert.Equal(t, 30, resp.ByStatus["lead"])
		assert.Equal(t, 25, resp.ByStatus["prospect"])
		assert.Equal(t, 25, resp.ByStatus["customer"])
		assert.Equal(t, 20, resp.ByStatus["churned"])
		require.Len(t, resp.ByTier, 4)
		assert.Equal(t, 60, resp.ByTier["free"])
		assert.Equal(t, 20, resp.ByTier["basic"])
		assert.Equal(t, 15, resp.ByTier["pro"])
		assert.Equal(t, 5, resp.ByTier["enterprise"])
	})

	t.Run("empty stats", func(t *testing.T) {
		stats := &usecase.CustomerStats{
			TotalCustomers:   0,
			ActiveCustomers:  0,
			ChurnedCustomers: 0,
			ByStatus:         make(map[aggregate.CustomerStatus]int),
			ByTier:           make(map[aggregate.CustomerTier]int),
		}

		resp := ToCustomerStatsResponse(stats)

		assert.Equal(t, 0, resp.TotalCustomers)
		assert.Equal(t, 0, resp.ActiveCustomers)
		assert.Equal(t, 0, resp.ChurnedCustomers)
		assert.Empty(t, resp.ByStatus)
		assert.Empty(t, resp.ByTier)
	})

	t.Run("partial stats", func(t *testing.T) {
		stats := &usecase.CustomerStats{
			TotalCustomers:   50,
			ActiveCustomers:  40,
			ChurnedCustomers: 10,
			ByStatus: map[aggregate.CustomerStatus]int{
				aggregate.CustomerStatusLead:     20,
				aggregate.CustomerStatusProspect: 20,
			},
			ByTier: map[aggregate.CustomerTier]int{
				aggregate.CustomerTierFree:  45,
				aggregate.CustomerTierBasic: 5,
			},
		}

		resp := ToCustomerStatsResponse(stats)

		assert.Equal(t, 50, resp.TotalCustomers)
		require.Len(t, resp.ByStatus, 2)
		assert.Equal(t, 20, resp.ByStatus["lead"])
		assert.Equal(t, 20, resp.ByStatus["prospect"])
		require.Len(t, resp.ByTier, 2)
		assert.Equal(t, 45, resp.ByTier["free"])
		assert.Equal(t, 5, resp.ByTier["basic"])
	})
}

func TestCustomerResponse_JSONSerialization(t *testing.T) {
	assignedTo := uuidv7.New()

	t.Run("serializes to JSON correctly", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = uuidv7.New()

		resp := ToCustomerResponse(c)

		jsonData, err := json.Marshal(resp)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		assert.Contains(t, string(jsonData), `"id"`)
		assert.Contains(t, string(jsonData), `"name"`)
		assert.Contains(t, string(jsonData), `"email"`)
		assert.Contains(t, string(jsonData), `"status"`)
		assert.Contains(t, string(jsonData), `"tier"`)
	})

	t.Run("omits nil optional fields", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		c.ID = uuidv7.New()

		resp := ToCustomerResponse(c)
		jsonData, _ := json.Marshal(resp)

		assert.NotContains(t, string(jsonData), `"user_id"`)
		assert.NotContains(t, string(jsonData), `"company_id"`)
		assert.NotContains(t, string(jsonData), `"phone"`)
	})

	t.Run("includes optional fields when present", func(t *testing.T) {
		companyID := uuidv7.New()
		c, _ := aggregate.NewB2BCustomer("Test User", "test@example.com", "website", companyID, assignedTo)
		c.ID = uuidv7.New()
		phone, _ := valueobject.NewPhone("+380501234567")
		c.Phone = &phone

		resp := ToCustomerResponse(c)
		jsonData, _ := json.Marshal(resp)

		assert.Contains(t, string(jsonData), `"company_id"`)
		assert.Contains(t, string(jsonData), `"phone"`)
	})
}

func TestCustomerStatsResponse_JSONSerialization(t *testing.T) {
	t.Run("serializes stats to JSON", func(t *testing.T) {
		stats := &usecase.CustomerStats{
			TotalCustomers:   100,
			ActiveCustomers:  80,
			ChurnedCustomers: 20,
			ByStatus: map[aggregate.CustomerStatus]int{
				aggregate.CustomerStatusLead: 50,
			},
			ByTier: map[aggregate.CustomerTier]int{
				aggregate.CustomerTierFree: 80,
			},
		}

		resp := ToCustomerStatsResponse(stats)
		jsonData, err := json.Marshal(resp)

		require.NoError(t, err)
		assert.Contains(t, string(jsonData), `"total_customers"`)
		assert.Contains(t, string(jsonData), `"by_status"`)
		assert.Contains(t, string(jsonData), `"by_tier"`)
	})
}

func TestCustomerResponse_HelperFlags(t *testing.T) {
	assignedTo := uuidv7.New()

	t.Run("lead status flags", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Lead", "lead@example.com", "website", assignedTo)
		resp := ToCustomerResponse(c)

		assert.True(t, resp.IsLead)
		assert.False(t, resp.IsProspect)
		assert.False(t, resp.IsActiveCustomer)
		assert.False(t, resp.IsChurned)
	})

	t.Run("prospect status flags", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Prospect", "prospect@example.com", "website", assignedTo)
		_ = c.QualifyAsProspect()
		resp := ToCustomerResponse(c)

		assert.False(t, resp.IsLead)
		assert.True(t, resp.IsProspect)
		assert.False(t, resp.IsActiveCustomer)
		assert.False(t, resp.IsChurned)
	})

	t.Run("active customer flags", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Customer", "customer@example.com", "website", assignedTo)
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		resp := ToCustomerResponse(c)

		assert.False(t, resp.IsLead)
		assert.False(t, resp.IsProspect)
		assert.True(t, resp.IsActiveCustomer)
		assert.False(t, resp.IsChurned)
	})

	t.Run("churned customer flags", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Churned", "churned@example.com", "website", assignedTo)
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		_ = c.Churn("left")
		resp := ToCustomerResponse(c)

		assert.False(t, resp.IsLead)
		assert.False(t, resp.IsProspect)
		assert.False(t, resp.IsActiveCustomer)
		assert.True(t, resp.IsChurned)
	})
}

func TestCustomerResponse_Timestamps(t *testing.T) {
	assignedTo := uuidv7.New()

	t.Run("includes created_at and updated_at", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Test", "test@example.com", "website", assignedTo)
		resp := ToCustomerResponse(c)

		assert.False(t, resp.CreatedAt.IsZero())
		assert.False(t, resp.UpdatedAt.IsZero())
	})

	t.Run("includes converted_at when converted", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Test", "test@example.com", "website", assignedTo)
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		resp := ToCustomerResponse(c)

		assert.NotNil(t, resp.ConvertedAt)
		assert.False(t, resp.ConvertedAt.IsZero())
	})

	t.Run("includes churned_at when churned", func(t *testing.T) {
		c, _ := aggregate.NewCustomer("Test", "test@example.com", "website", assignedTo)
		_ = c.QualifyAsProspect()
		_ = c.ConvertToCustomer()
		_ = c.Churn("reason")
		resp := ToCustomerResponse(c)

		assert.NotNil(t, resp.ChurnedAt)
		assert.False(t, resp.ChurnedAt.IsZero())
		assert.NotEmpty(t, resp.ChurnReason)
	})
}
