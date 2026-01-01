package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Request DTOs

// CustomerLifecycleRequest for lifecycle endpoint
type CustomerLifecycleRequest struct {
	Period string `form:"period" binding:"omitempty,oneof=month week quarter"`
}

// SalesRepPerformanceRequest for sales rep performance endpoint
type SalesRepPerformanceRequest struct {
	TopN       int    `form:"top_n" binding:"omitempty,min=1,max=100"`
	SalesRepID string `form:"sales_rep_id" binding:"omitempty,uuid"`
}

// ParseSalesRepID converts string UUID to uuidv7.UUID
func (r *SalesRepPerformanceRequest) ParseSalesRepID() (uuidv7.UUID, error) {
	return uuidv7.Parse(r.SalesRepID)
}

// RevenueTimeSeriesRequest for revenue time series endpoint
type RevenueTimeSeriesRequest struct {
	StartDate   string `form:"start_date" binding:"required"`
	EndDate     string `form:"end_date" binding:"required"`
	Granularity string `form:"granularity" binding:"omitempty,oneof=day week month"`
}

// Parse converts string dates to time.Time
func (r *RevenueTimeSeriesRequest) Parse() (time.Time, time.Time, string, error) {
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}

	endDate, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}

	granularity := r.Granularity
	if granularity == "" {
		granularity = "month"
	}

	return startDate, endDate, granularity, nil
}

// Response transformers

// ToCustomerOverviewResponse converts domain model to response map
func ToCustomerOverviewResponse(overview *analytics.CustomerOverview) map[string]interface{} {
	return map[string]interface{}{
		"total_customers":       overview.TotalCustomers,
		"by_state":              overview.ByState,
		"by_tier":               overview.ByTier,
		"new_this_month":        overview.NewThisMonth,
		"churned_this_month":    overview.ChurnedThisMonth,
		"average_lifetime_weeks": overview.AverageLifetimeWeeks,
		"top_sales_reps":        ToSalesRepPerformanceResponse(overview.TopSalesReps),
	}
}

// ToCustomerLifecycleResponse converts domain models to response array
func ToCustomerLifecycleResponse(lifecycles []analytics.CustomerLifecycle) []map[string]interface{} {
	result := make([]map[string]interface{}, len(lifecycles))
	for i, lc := range lifecycles {
		result[i] = map[string]interface{}{
			"period":                lc.Period,
			"lead_count":            lc.LeadCount,
			"prospect_count":        lc.ProspectCount,
			"customer_count":        lc.CustomerCount,
			"churned_count":         lc.ChurnedCount,
			"lead_to_prospect":      lc.LeadToProspect,
			"prospect_to_customer":  lc.ProspectToCustomer,
			"customer_to_churned":   lc.CustomerToChurned,
			"conversion_rate":       lc.ConversionRate,
		}
	}
	return result
}

// ToDealPipelineResponse converts domain model to response map
func ToDealPipelineResponse(pipeline *analytics.DealPipeline) map[string]interface{} {
	return map[string]interface{}{
		"total_deals":          pipeline.TotalDeals,
		"total_value":          pipeline.TotalValue,
		"by_stage":             pipeline.ByStage,
		"average_deal_size":    pipeline.AverageDealSize,
		"win_rate":             pipeline.WinRate,
		"average_days_to_close": pipeline.AverageDaysToClose,
	}
}

// ToSalesRepPerformanceResponse converts domain models to response array
func ToSalesRepPerformanceResponse(performances []analytics.SalesRepPerformance) []map[string]interface{} {
	result := make([]map[string]interface{}, len(performances))
	for i, perf := range performances {
		result[i] = map[string]interface{}{
			"sales_rep_id":         perf.SalesRepID,
			"sales_rep_name":       perf.SalesRepName,
			"active_customers":     perf.ActiveCustomers,
			"deals_won":            perf.DealsWon,
			"deals_lost":           perf.DealsLost,
			"total_revenue":        perf.TotalRevenue,
			"average_deal_size":    perf.AverageDealSize,
			"win_rate":             perf.WinRate,
			"average_days_to_close": perf.AverageDaysToClose,
		}
	}
	return result
}

// ToRevenueTimeSeriesResponse converts domain models to response array
func ToRevenueTimeSeriesResponse(series []analytics.RevenueTimeSeries) []map[string]interface{} {
	result := make([]map[string]interface{}, len(series))
	for i, s := range series {
		result[i] = map[string]interface{}{
			"period":        s.Period,
			"timestamp":     s.Timestamp,
			"new_customers": s.NewCustomers,
			"deals_won":     s.DealsWon,
			"revenue":       s.Revenue,
		}
	}
	return result
}

// ToInteractionInsightsResponse converts domain model to response map
func ToInteractionInsightsResponse(insights *analytics.InteractionInsights) map[string]interface{} {
	return map[string]interface{}{
		"total_interactions":    insights.TotalInteractions,
		"by_type":               insights.ByType,
		"by_outcome":            insights.ByOutcome,
		"average_duration":      insights.AverageDuration,
		"follow_up_pending":     insights.FollowUpPending,
		"interaction_frequency": insights.InteractionFrequency,
	}
}
