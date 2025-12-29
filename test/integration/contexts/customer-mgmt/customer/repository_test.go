package customer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

func TestCustomerRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	t.Run("create B2C customer", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, err := customer.NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		require.NoError(t, err)

		err = repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, retrieved.ID)
		assert.Equal(t, c.Name, retrieved.Name)
		assert.Equal(t, c.Email.Value(), retrieved.Email.Value())
		assert.Equal(t, customer.CustomerStatusLead, retrieved.Status)
		assert.Equal(t, customer.CustomerTierFree, retrieved.Tier)
		assert.Equal(t, "website", retrieved.Source)
		assert.Equal(t, assignedTo, retrieved.AssignedTo)
		assert.Nil(t, retrieved.UserID)
		assert.Nil(t, retrieved.CompanyID)
	})

	t.Run("create B2B customer", func(t *testing.T) {
		assignedTo := uuidv7.New()
		companyID := uuidv7.New()
		c, err := customer.NewB2BCustomer("Jane Smith", "jane@company.com", "referral", companyID, assignedTo)
		require.NoError(t, err)

		err = repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, retrieved.ID)
		assert.True(t, retrieved.IsB2B())
		require.NotNil(t, retrieved.CompanyID)
		assert.Equal(t, companyID, *retrieved.CompanyID)
	})

	t.Run("create customer with phone", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, err := customer.NewCustomer("Bob Test", "bob@example.com", "api", assignedTo)
		require.NoError(t, err)

		phone, err := valueobject.NewPhone("+380501234567")
		require.NoError(t, err)
		c.Phone = &phone

		err = repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved.Phone)
		assert.Equal(t, "+380501234567", retrieved.Phone.Value())
	})

	t.Run("create customer with tags", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, err := customer.NewCustomer("Tagged User", "tagged@example.com", "website", assignedTo)
		require.NoError(t, err)

		_ = c.AddTag("vip")
		_ = c.AddTag("premium")

		err = repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.Tags, 2)
		assert.Contains(t, retrieved.Tags, "vip")
		assert.Contains(t, retrieved.Tags, "premium")
	})
}

func TestCustomerRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	t.Run("get existing customer", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Test User", "test@example.com", "website", assignedTo)
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, retrieved.ID)
		assert.Equal(t, c.Name, retrieved.Name)
	})

	t.Run("get non-existent customer", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		retrieved, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestCustomerRepository_GetByEmail(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	t.Run("get by existing email", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Email Test", "email@example.com", "website", assignedTo)
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByEmail(ctx, "email@example.com")
		require.NoError(t, err)
		assert.Equal(t, c.ID, retrieved.ID)
		assert.Equal(t, "email@example.com", retrieved.Email.Value())
	})

	t.Run("get by non-existent email", func(t *testing.T) {
		retrieved, err := repo.GetByEmail(ctx, "notfound@example.com")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestCustomerRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	t.Run("update customer name", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Original Name", "update@example.com", "website", assignedTo)
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		c.Name = "Updated Name"
		err = repo.Update(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
	})

	t.Run("update customer status", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Status Test", "status@example.com", "website", assignedTo)
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		err = c.QualifyAsProspect()
		require.NoError(t, err)
		err = repo.Update(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerStatusProspect, retrieved.Status)
	})

	t.Run("update customer tier", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Tier Test", "tier@example.com", "website", assignedTo)
		require.NoError(t, c.QualifyAsProspect())
		require.NoError(t, c.ConvertToCustomer())
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		err = c.UpgradeTier(customer.CustomerTierPro)
		require.NoError(t, err)
		err = repo.Update(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, customer.CustomerTierPro, retrieved.Tier)
	})

	t.Run("update customer tags", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Tag Test", "tags@example.com", "website", assignedTo)
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		_ = c.AddTag("important")
		err = repo.Update(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Contains(t, retrieved.Tags, "important")
	})
}

func TestCustomerRepository_Delete(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	t.Run("soft delete customer", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Delete Test", "delete@example.com", "website", assignedTo)
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		err = repo.Delete(ctx, c.ID)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, c.ID)
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestCustomerRepository_ExistsByEmail(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	t.Run("exists by email", func(t *testing.T) {
		assignedTo := uuidv7.New()
		c, _ := customer.NewCustomer("Exists Test", "exists@example.com", "website", assignedTo)
		err := repo.Create(ctx, c)
		require.NoError(t, err)

		exists, err := repo.ExistsByEmail(ctx, "exists@example.com")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("does not exist by email", func(t *testing.T) {
		exists, err := repo.ExistsByEmail(ctx, "notfound@example.com")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestCustomerRepository_ListByStatus(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()
	assignedTo := uuidv7.New()

	// Create test data BEFORE subtests to ensure it persists
	lead1, _ := customer.NewCustomer("Lead 1", "lead1@example.com", "website", assignedTo)
	require.NoError(t, repo.Create(ctx, lead1))

	lead2, _ := customer.NewCustomer("Lead 2", "lead2@example.com", "website", assignedTo)
	require.NoError(t, repo.Create(ctx, lead2))

	prospect, _ := customer.NewCustomer("Prospect", "prospect@example.com", "website", assignedTo)
	require.NoError(t, prospect.QualifyAsProspect())
	require.NoError(t, repo.Create(ctx, prospect))

	t.Run("list leads", func(t *testing.T) {
		customers, total, err := repo.ListByStatus(ctx, customer.CustomerStatusLead, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 2)
		assert.GreaterOrEqual(t, total, 2)
		for _, c := range customers {
			assert.Equal(t, customer.CustomerStatusLead, c.Status)
		}
	})

	t.Run("list prospects", func(t *testing.T) {
		customers, total, err := repo.ListByStatus(ctx, customer.CustomerStatusProspect, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 1)
		assert.GreaterOrEqual(t, total, 1)
		for _, c := range customers {
			assert.Equal(t, customer.CustomerStatusProspect, c.Status)
		}
	})
}

func TestCustomerRepository_ListByTier(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()
	assignedTo := uuidv7.New()

	// Create test data BEFORE subtests to ensure it persists
	free1, _ := customer.NewCustomer("Free 1", "free1@example.com", "website", assignedTo)
	require.NoError(t, repo.Create(ctx, free1))

	free2, _ := customer.NewCustomer("Free 2", "free2@example.com", "website", assignedTo)
	require.NoError(t, repo.Create(ctx, free2))

	pro, _ := customer.NewCustomer("Pro", "pro@example.com", "website", assignedTo)
	require.NoError(t, pro.QualifyAsProspect())
	require.NoError(t, pro.ConvertToCustomer())
	require.NoError(t, pro.UpgradeTier(customer.CustomerTierPro))
	require.NoError(t, repo.Create(ctx, pro))

	t.Run("list free tier", func(t *testing.T) {
		customers, total, err := repo.ListByTier(ctx, customer.CustomerTierFree, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 2)
		assert.GreaterOrEqual(t, total, 2)
		for _, c := range customers {
			assert.Equal(t, customer.CustomerTierFree, c.Tier)
		}
	})

	t.Run("list pro tier", func(t *testing.T) {
		customers, total, err := repo.ListByTier(ctx, customer.CustomerTierPro, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 1)
		assert.GreaterOrEqual(t, total, 1)
		for _, c := range customers {
			assert.Equal(t, customer.CustomerTierPro, c.Tier)
		}
	})
}

func TestCustomerRepository_List(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()
	assignedTo := uuidv7.New()

	// Create test data BEFORE subtests to ensure it persists
	for i := 1; i <= 5; i++ {
		email := fmt.Sprintf("%s@example.com", uuidv7.New())
		c, err := customer.NewCustomer("Customer", email, "website", assignedTo)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, c))
	}

	t.Run("list all customers", func(t *testing.T) {
		customers, total, err := repo.List(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 5)
		assert.GreaterOrEqual(t, total, 5)
	})

	t.Run("list with pagination", func(t *testing.T) {
		customers, total, err := repo.List(ctx, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(customers), 2)
		assert.GreaterOrEqual(t, total, 5)
	})

	t.Run("list with offset", func(t *testing.T) {
		customers, total, err := repo.List(ctx, 2, 2)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(customers), 2)
		assert.GreaterOrEqual(t, total, 5)
	})
}

func TestCustomerRepository_CountByStatus(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()
	assignedTo := uuidv7.New()

	for i := 1; i <= 3; i++ {
		email := fmt.Sprintf("%s@example.com", uuidv7.New())
		c, _ := customer.NewCustomer("Lead", email, "website", assignedTo)
		_ = repo.Create(ctx, c)
	}

	for i := 1; i <= 2; i++ {
		email := fmt.Sprintf("%s@example.com", uuidv7.New())
		c, _ := customer.NewCustomer("Prospect", email, "website", assignedTo)
		_ = c.QualifyAsProspect()
		_ = repo.Create(ctx, c)
	}

	t.Run("count leads", func(t *testing.T) {
		count, err := repo.CountByStatus(ctx, customer.CustomerStatusLead)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 3)
	})

	t.Run("count prospects", func(t *testing.T) {
		count, err := repo.CountByStatus(ctx, customer.CustomerStatusProspect)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
	})
}

func TestCustomerRepository_CountByTier(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()
	assignedTo := uuidv7.New()

	for i := 1; i <= 4; i++ {
		email := fmt.Sprintf("%s@example.com", uuidv7.New())
		c, _ := customer.NewCustomer("Free", email, "website", assignedTo)
		_ = repo.Create(ctx, c)
	}

	c, _ := customer.NewCustomer("Basic User", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", assignedTo)
	_ = c.QualifyAsProspect()
	_ = c.ConvertToCustomer()
	_ = c.UpgradeTier(customer.CustomerTierBasic)
	_ = repo.Create(ctx, c)

	t.Run("count free tier", func(t *testing.T) {
		count, err := repo.CountByTier(ctx, customer.CustomerTierFree)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 4)
	})

	t.Run("count basic tier", func(t *testing.T) {
		count, err := repo.CountByTier(ctx, customer.CustomerTierBasic)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, 1)
	})
}

func TestCustomerRepository_GetByUserID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	t.Run("get customer by user ID", func(t *testing.T) {
		// Create a user first to satisfy foreign key
		userID := uuidv7.New()
		userEmail := fmt.Sprintf("%s@example.com", uuidv7.New())
		_, err := db.DB.Exec(`
			INSERT INTO identity_users (id, email, password_hash, status, created_at, updated_at)
			VALUES ($1, $2, 'hash', 'active', NOW(), NOW())
		`, userID, userEmail)
		require.NoError(t, err)

		assignedTo := uuidv7.New()
		customerEmail := fmt.Sprintf("%s@example.com", uuidv7.New())
		c, _ := customer.NewCustomer("User Linked", customerEmail, "website", assignedTo)
		_ = c.LinkToUser(userID)
		err = repo.Create(ctx, c)
		require.NoError(t, err)

		retrieved, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, c.ID, retrieved.ID)
		require.NotNil(t, retrieved.UserID)
		assert.Equal(t, userID, *retrieved.UserID)
	})

	t.Run("get by non-existent user ID", func(t *testing.T) {
		nonExistentUserID := uuidv7.New()
		retrieved, err := repo.GetByUserID(ctx, nonExistentUserID)
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestCustomerRepository_ListByAssignedTo(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()

	rep1 := uuidv7.New()
	rep2 := uuidv7.New()

	for i := 1; i <= 3; i++ {
		email := fmt.Sprintf("%s@example.com", uuidv7.New())
		c, _ := customer.NewCustomer("Rep1 Customer", email, "website", rep1)
		_ = repo.Create(ctx, c)
	}

	c, _ := customer.NewCustomer("Rep2 Customer", fmt.Sprintf("%s@example.com", uuidv7.New()), "website", rep2)
	_ = repo.Create(ctx, c)

	t.Run("list by assigned rep", func(t *testing.T) {
		customers, total, err := repo.ListByAssignedTo(ctx, rep1, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 3)
		assert.GreaterOrEqual(t, total, 3)
		for _, customer := range customers {
			assert.Equal(t, rep1, customer.AssignedTo)
		}
	})

	t.Run("list by different rep", func(t *testing.T) {
		customers, total, err := repo.ListByAssignedTo(ctx, rep2, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 1)
		assert.GreaterOrEqual(t, total, 1)
		for _, customer := range customers {
			assert.Equal(t, rep2, customer.AssignedTo)
		}
	})
}

func TestCustomerRepository_ListByCompanyID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewCustomerRepository(db.DB)
	ctx := context.Background()
	assignedTo := uuidv7.New()

	company1 := uuidv7.New()
	company2 := uuidv7.New()

	for i := 1; i <= 2; i++ {
		email := fmt.Sprintf("%s@example.com", uuidv7.New())
		c, _ := customer.NewB2BCustomer("Company1 Contact", email, "referral", company1, assignedTo)
		_ = repo.Create(ctx, c)
	}

	c, _ := customer.NewB2BCustomer("Company2 Contact", fmt.Sprintf("%s@example.com", uuidv7.New()), "referral", company2, assignedTo)
	_ = repo.Create(ctx, c)

	t.Run("list by company ID", func(t *testing.T) {
		customers, err := repo.ListByCompanyID(ctx, company1)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 2)
		for _, customer := range customers {
			require.NotNil(t, customer.CompanyID)
			assert.Equal(t, company1, *customer.CompanyID)
		}
	})

	t.Run("list by different company", func(t *testing.T) {
		customers, err := repo.ListByCompanyID(ctx, company2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(customers), 1)
		for _, customer := range customers {
			require.NotNil(t, customer.CompanyID)
			assert.Equal(t, company2, *customer.CompanyID)
		}
	})
}
