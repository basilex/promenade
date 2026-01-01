package analytics_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	companyRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal"
	dealRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	interactionRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

func TestAnalyticsUseCase_GetCustomerOverview(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	createTestCustomersForAnalytics(t, db.DB, ctx)

	overview, err := analyticsUC.GetCustomerOverview(ctx)
	require.NoError(t, err)
	require.NotNil(t, overview)

	assert.Greater(t, overview.TotalCustomers, 0)
	assert.NotNil(t, overview.TopSalesReps)
}

func TestAnalyticsUseCase_GetCustomerLifecycle(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	createTestCustomersForAnalytics(t, db.DB, ctx)

	tests := []struct {
		name   string
		period string
	}{
		{"monthly", "month"},
		{"weekly", "week"},
		{"quarterly", "quarter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle, err := analyticsUC.GetCustomerLifecycle(ctx, tt.period)
			require.NoError(t, err)
			assert.NotNil(t, lifecycle)
		})
	}
}

func TestAnalyticsUseCase_GetCustomerSegmentation(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	createTestCustomersForAnalytics(t, db.DB, ctx)

	segmentation, err := analyticsUC.GetCustomerSegmentation(ctx)
	require.NoError(t, err)
	assert.NotNil(t, segmentation)

	if len(segmentation) > 0 {
		assert.NotEmpty(t, segmentation[0].SegmentName)
		assert.Equal(t, "USD", segmentation[0].TotalRevenue.Currency)
	}
}

func TestAnalyticsUseCase_GetDealPipeline(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	createTestDealsForAnalytics(t, db.DB, ctx)

	pipeline, err := analyticsUC.GetDealPipeline(ctx)
	require.NoError(t, err)
	require.NotNil(t, pipeline)

	assert.Equal(t, "USD", pipeline.TotalValue.Currency)
	assert.NotNil(t, pipeline.ByStage)
}

func TestAnalyticsUseCase_GetSalesRepPerformance(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	createTestDealsForAnalytics(t, db.DB, ctx)

	performance, err := analyticsUC.GetSalesRepPerformance(ctx, 10)
	require.NoError(t, err)
	assert.NotNil(t, performance)

	if len(performance) > 0 {
		assert.Equal(t, "USD", performance[0].TotalRevenue.Currency)
	}
}

func TestAnalyticsUseCase_GetSalesRepPerformanceByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	salesRepID := createTestDealsForAnalytics(t, db.DB, ctx)

	if salesRepID != nil {
		performance, err := analyticsUC.GetSalesRepPerformanceByID(ctx, *salesRepID)
		require.NoError(t, err)
		require.NotNil(t, performance)

		assert.Equal(t, *salesRepID, performance.SalesRepID)
		assert.Equal(t, "USD", performance.TotalRevenue.Currency)
	}
}

func TestAnalyticsUseCase_GetRevenueTimeSeries(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	createTestDealsForAnalytics(t, db.DB, ctx)

	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now().AddDate(0, 1, 0)

	tests := []struct {
		name        string
		granularity string
	}{
		{"daily", "day"},
		{"weekly", "week"},
		{"monthly", "month"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			series, err := analyticsUC.GetRevenueTimeSeries(ctx, startDate, endDate, tt.granularity)
			require.NoError(t, err)
			assert.NotNil(t, series)

			for _, ts := range series {
				assert.Equal(t, "USD", ts.Revenue.Currency)
			}
		})
	}
}

func TestAnalyticsUseCase_GetInteractionInsights(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsUC := analytics.NewUseCase(db.DB)
	createTestInteractionsForAnalytics(t, db.DB, ctx)

	insights, err := analyticsUC.GetInteractionInsights(ctx)
	require.NoError(t, err)
	require.NotNil(t, insights)

	assert.GreaterOrEqual(t, insights.TotalInteractions, 0)
	assert.NotNil(t, insights.ByType)
	assert.NotNil(t, insights.ByOutcome)
}

func createTestCustomersForAnalytics(t *testing.T, db *sqlx.DB, ctx context.Context) {
	t.Helper()

	customerRepository := customerRepo.NewCustomerRepository(db)
	
	// Create sales rep user
	salesRepID := uuidv7.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO identity_users (id, email, password_hash, status, email_verified)
		VALUES ($1, $2, $3, $4, $5)
	`, salesRepID, "salesrep@example.com", "hash", "active", true)
	require.NoError(t, err)

	customers := []struct {
		email  string
		status customer.CustomerStatus
		tier   customer.CustomerTier
	}{
		{"test1@example.com", customer.CustomerStatusCustomer, customer.CustomerTierPro},
		{"test2@example.com", customer.CustomerStatusCustomer, customer.CustomerTierBasic},
		{"test3@example.com", customer.CustomerStatusProspect, customer.CustomerTierFree},
		{"test4@example.com", customer.CustomerStatusLead, customer.CustomerTierFree},
		{"test5@example.com", customer.CustomerStatusChurned, customer.CustomerTierBasic},
	}

	for _, c := range customers {
		cust, err := customer.NewCustomer("Customer", c.email, "website", salesRepID)
		require.NoError(t, err)
		cust.Status = c.status
		cust.Tier = c.tier

		err = customerRepository.Create(ctx, cust)
		require.NoError(t, err)
	}
}

func createTestDealsForAnalytics(t *testing.T, db *sqlx.DB, ctx context.Context) *uuidv7.UUID {
	t.Helper()

	// Create sales rep user
	salesRepID := uuidv7.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO identity_users (id, email, password_hash, status, email_verified)
		VALUES ($1, $2, $3, $4, $5)
	`, salesRepID, "dealsrep@example.com", "hash", "active", true)
	require.NoError(t, err)

	customerRepository := customerRepo.NewCustomerRepository(db)
	cust, err := customer.NewCustomer("Deal Customer", "dealcustomer@example.com", "website", salesRepID)
	require.NoError(t, err)
	err = customerRepository.Create(ctx, cust)
	require.NoError(t, err)

	dealRepository := dealRepo.NewDealRepository(db)

	deals := []struct {
		stage deal.DealStage
		value int64
	}{
		{deal.DealStageQualified, 50000},
		{deal.DealStageProposal, 75000},
		{deal.DealStageNegotiation, 100000},
		{deal.DealStageClosedWon, 150000},
		{deal.DealStageClosedLost, 25000},
	}

	for _, d := range deals {
		value, err := valueobject.NewMoney(d.value, "USD")
		require.NoError(t, err)
		expectedClose := time.Now().Add(30 * 24 * time.Hour)
		newDeal, err := deal.NewDeal(cust.ID, "Test Deal", value, salesRepID, expectedClose)
		require.NoError(t, err)
		newDeal.Stage = d.stage

		if d.stage == deal.DealStageClosedWon {
			now := time.Now()
			newDeal.ActualCloseDate = &now
		}

		err = dealRepository.Create(ctx, newDeal)
		require.NoError(t, err)
	}

	return &salesRepID
}

func createTestInteractionsForAnalytics(t *testing.T, db *sqlx.DB, ctx context.Context) {
	t.Helper()

	// Create sales rep user
	salesRepID := uuidv7.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO identity_users (id, email, password_hash, status, email_verified)
		VALUES ($1, $2, $3, $4, $5)
	`, salesRepID, "interactionrep@example.com", "hash", "active", true)
	require.NoError(t, err)

	customerRepository := customerRepo.NewCustomerRepository(db)
	cust, err := customer.NewCustomer("Interaction Customer", "interaction@example.com", "website", salesRepID)
	require.NoError(t, err)
	err = customerRepository.Create(ctx, cust)
	require.NoError(t, err)

	companyRepository := companyRepo.NewCompanyRepository(db)
	comp, err := company.NewCompany("Company-"+uuidv7.New().String(), "corporation")
	require.NoError(t, err)
	err = companyRepository.Create(ctx, comp)
	require.NoError(t, err)

	interactionRepository := interactionRepo.NewInteractionRepository(db)

	interactions := []struct {
		iType     interaction.InteractionType
		direction interaction.InteractionDirection
		outcome   interaction.InteractionOutcome
	}{
		{interaction.InteractionTypeCall, interaction.InteractionDirectionOutbound, interaction.InteractionOutcomeSuccessful},
		{interaction.InteractionTypeEmail, interaction.InteractionDirectionOutbound, interaction.InteractionOutcomeSuccessful},
		{interaction.InteractionTypeMeeting, interaction.InteractionDirectionInbound, interaction.InteractionOutcomeSuccessful},
		{interaction.InteractionTypeNote, interaction.InteractionDirectionInbound, interaction.InteractionOutcomeNoAnswer},
		{interaction.InteractionTypeCall, interaction.InteractionDirectionOutbound, interaction.InteractionOutcomeSuccessful},
	}

	for i, inter := range interactions {
		startedAt := time.Now().Add(-time.Duration(i) * time.Hour)
		newInter, err := interaction.NewInteraction(
			cust.ID,
			&comp.ID,
			inter.iType,
			inter.direction,
			"Test interaction",
			"Test interaction "+string(rune('0'+i)),
			salesRepID,
			startedAt,
		)
		require.NoError(t, err)

		ended := startedAt.Add(15 * time.Minute)
		newInter.EndedAt = &ended
		duration := 900
		newInter.DurationSec = &duration
		outcome := inter.outcome
		newInter.Outcome = &outcome

		err = interactionRepository.Create(ctx, newInter)
		require.NoError(t, err)
	}
}
