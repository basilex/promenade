package analytics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewUseCase(t *testing.T) {
	uc := NewUseCase(nil)
	assert.NotNil(t, uc)
}

func TestCustomerOverview_Structure(t *testing.T) {
	overview := CustomerOverview{
		TotalCustomers:       100,
		ByState:              map[string]int{"customer": 80},
		ByTier:               map[string]int{"pro": 50},
		NewThisMonth:         10,
		ChurnedThisMonth:     2,
		AverageLifetimeWeeks: 26,
		TopSalesReps:         []SalesRepPerformance{},
	}

	assert.Equal(t, 100, overview.TotalCustomers)
	assert.Equal(t, 80, overview.ByState["customer"])
	assert.Equal(t, 50, overview.ByTier["pro"])
	assert.Equal(t, 10, overview.NewThisMonth)
	assert.Equal(t, 2, overview.ChurnedThisMonth)
	assert.Equal(t, 26.0, overview.AverageLifetimeWeeks)
	assert.NotNil(t, overview.TopSalesReps)
}

func TestCustomerLifecycle_Structure(t *testing.T) {
	lifecycle := CustomerLifecycle{
		Period:             "2025-12",
		LeadCount:          10,
		ProspectCount:      15,
		CustomerCount:      50,
		ChurnedCount:       2,
		LeadToProspect:     5,
		ProspectToCustomer: 8,
		CustomerToChurned:  2,
		ConversionRate:     50.0,
	}

	assert.Equal(t, "2025-12", lifecycle.Period)
	assert.Equal(t, 10, lifecycle.LeadCount)
	assert.Equal(t, 15, lifecycle.ProspectCount)
	assert.Equal(t, 50, lifecycle.CustomerCount)
	assert.Equal(t, 2, lifecycle.ChurnedCount)
	assert.Equal(t, 5, lifecycle.LeadToProspect)
	assert.Equal(t, 8, lifecycle.ProspectToCustomer)
	assert.Equal(t, 2, lifecycle.CustomerToChurned)
	assert.Equal(t, 50.0, lifecycle.ConversionRate)
}

func TestDealPipeline_Structure(t *testing.T) {
	totalValue, err := valueobject.NewMoney(500000, "USD")
	assert.NoError(t, err)

	avgDealSize, err := valueobject.NewMoney(50000, "USD")
	assert.NoError(t, err)

	pipeline := DealPipeline{
		TotalDeals:         10,
		TotalValue:         totalValue,
		ByStage:            []DealStageStats{},
		AverageDealSize:    avgDealSize,
		WinRate:            60.0,
		AverageDaysToClose: 30,
	}

	assert.Equal(t, 10, pipeline.TotalDeals)
	assert.Equal(t, int64(500000), pipeline.TotalValue.Amount)
	assert.Equal(t, int64(50000), pipeline.AverageDealSize.Amount)
	assert.Equal(t, 60.0, pipeline.WinRate)
	assert.Equal(t, 30.0, pipeline.AverageDaysToClose)
	assert.NotNil(t, pipeline.ByStage)
}

func TestSalesRepPerformance_Structure(t *testing.T) {
	salesRepID := uuidv7.New()

	totalRevenue, err := valueobject.NewMoney(1000000, "USD")
	assert.NoError(t, err)

	avgDealSize, err := valueobject.NewMoney(100000, "USD")
	assert.NoError(t, err)

	perf := SalesRepPerformance{
		SalesRepID:         salesRepID,
		SalesRepName:       "John Doe",
		ActiveCustomers:    25,
		DealsWon:           10,
		DealsLost:          2,
		TotalRevenue:       totalRevenue,
		AverageDealSize:    avgDealSize,
		WinRate:            83.33,
		AverageDaysToClose: 28,
	}

	assert.Equal(t, salesRepID, perf.SalesRepID)
	assert.Equal(t, "John Doe", perf.SalesRepName)
	assert.Equal(t, 25, perf.ActiveCustomers)
	assert.Equal(t, 10, perf.DealsWon)
	assert.Equal(t, 2, perf.DealsLost)
	assert.Equal(t, int64(1000000), perf.TotalRevenue.Amount)
	assert.Equal(t, int64(100000), perf.AverageDealSize.Amount)
	assert.Equal(t, 83.33, perf.WinRate)
	assert.Equal(t, 28.0, perf.AverageDaysToClose)
}

func TestRevenueTimeSeries_Structure(t *testing.T) {
	revenue, err := valueobject.NewMoney(250000, "USD")
	assert.NoError(t, err)

	ts := RevenueTimeSeries{
		Period:       "2025-12",
		Timestamp:    time.Now(),
		NewCustomers: 5,
		DealsWon:     8,
		Revenue:      revenue,
	}

	assert.Equal(t, "2025-12", ts.Period)
	assert.Equal(t, 5, ts.NewCustomers)
	assert.Equal(t, 8, ts.DealsWon)
	assert.Equal(t, "USD", ts.Revenue.Currency)
	assert.Equal(t, int64(250000), ts.Revenue.Amount)
	assert.False(t, ts.Timestamp.IsZero())
}

func TestInteractionInsights_Structure(t *testing.T) {
	insights := InteractionInsights{
		TotalInteractions:    500,
		ByType:               map[string]int{"call": 200},
		ByOutcome:            map[string]int{"successful": 400},
		AverageDuration:      30.0,
		FollowUpPending:      25,
		InteractionFrequency: 5.0,
	}

	assert.Equal(t, 500, insights.TotalInteractions)
	assert.Equal(t, 200, insights.ByType["call"])
	assert.Equal(t, 400, insights.ByOutcome["successful"])
	assert.Equal(t, 30.0, insights.AverageDuration)
	assert.Equal(t, 25, insights.FollowUpPending)
	assert.Equal(t, 5.0, insights.InteractionFrequency)
}
