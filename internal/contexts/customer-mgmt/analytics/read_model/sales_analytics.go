package readmodel

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// Read models are optimized for querying and reporting, separate from write models.

// CustomerOverview provides high-level statistics about customers
type CustomerOverview struct {
	TotalCustomers      int                      `json:"total_customers"`
	ByState             map[string]int           `json:"by_state"`             // Lead, Prospect, Customer, Churned
	ByTier              map[string]int           `json:"by_tier"`              // free, basic, pro, enterprise
	NewThisMonth        int                      `json:"new_this_month"`
	ChurnedThisMonth    int                      `json:"churned_this_month"`
	AverageLifetimeWeeks float64                 `json:"average_lifetime_weeks"`
	TopSalesReps        []SalesRepPerformance    `json:"top_sales_reps"`
}

// CustomerLifecycle tracks customer state transitions over time
type CustomerLifecycle struct {
	Period              string                   `json:"period"`               // "2025-12", "2025-W52", "2025-Q4"
	LeadCount           int                      `json:"lead_count"`
	ProspectCount       int                      `json:"prospect_count"`
	CustomerCount       int                      `json:"customer_count"`
	ChurnedCount        int                      `json:"churned_count"`
	LeadToProspect      int                      `json:"lead_to_prospect"`     // Conversions
	ProspectToCustomer  int                      `json:"prospect_to_customer"`
	CustomerToChurned   int                      `json:"customer_to_churned"`
	ConversionRate      float64                  `json:"conversion_rate"`      // Lead → Customer %
}

// DealPipeline provides pipeline statistics by stage
type DealPipeline struct {
	TotalDeals          int                      `json:"total_deals"`
	TotalValue          valueobject.Money        `json:"total_value"`
	ByStage             []DealStageStats         `json:"by_stage"`
	AverageDealSize     valueobject.Money        `json:"average_deal_size"`
	WinRate             float64                  `json:"win_rate"`              // Won / (Won + Lost)
	AverageDaysToClose  float64                  `json:"average_days_to_close"`
}

// DealStageStats provides statistics for a specific deal stage
type DealStageStats struct {
	Stage               string                   `json:"stage"`
	Count               int                      `json:"count"`
	TotalValue          valueobject.Money        `json:"total_value"`
	AverageProbability  float64                  `json:"average_probability"`
	AverageDaysInStage  float64                  `json:"average_days_in_stage"`
}

// DealConversion tracks conversion rates between stages
type DealConversion struct {
	FromStage           string                   `json:"from_stage"`
	ToStage             string                   `json:"to_stage"`
	Count               int                      `json:"count"`
	ConversionRate      float64                  `json:"conversion_rate"`       // Percentage
	AverageDays         float64                  `json:"average_days"`          // Days to convert
}

// SalesRepPerformance tracks individual sales rep metrics
type SalesRepPerformance struct {
	SalesRepID          uuidv7.UUID              `json:"sales_rep_id"`
	SalesRepName        string                   `json:"sales_rep_name"`        // From users table
	ActiveCustomers     int                      `json:"active_customers"`
	DealsWon            int                      `json:"deals_won"`
	DealsLost           int                      `json:"deals_lost"`
	TotalRevenue        valueobject.Money        `json:"total_revenue"`
	AverageDealSize     valueobject.Money        `json:"average_deal_size"`
	WinRate             float64                  `json:"win_rate"`
	AverageDaysToClose  float64                  `json:"average_days_to_close"`
}

// RevenueTimeSeries tracks revenue over time periods
type RevenueTimeSeries struct {
	Period              string                   `json:"period"`                // "2025-12", "2025-W52"
	NewCustomers        int                      `json:"new_customers"`
	DealsWon            int                      `json:"deals_won"`
	Revenue             valueobject.Money        `json:"revenue"`
	Timestamp           time.Time                `json:"timestamp"`
}

// CustomerSegmentation provides customer distribution by segments
type CustomerSegmentation struct {
	SegmentName         string                   `json:"segment_name"`
	Count               int                      `json:"count"`
	Percentage          float64                  `json:"percentage"`
	AverageLifetimeWeeks float64                 `json:"average_lifetime_weeks"`
	TotalRevenue        valueobject.Money        `json:"total_revenue"`
}

// InteractionInsights provides interaction analytics
type InteractionInsights struct {
	TotalInteractions   int                      `json:"total_interactions"`
	ByType              map[string]int           `json:"by_type"`               // call, email, meeting, note
	ByOutcome           map[string]int           `json:"by_outcome"`            // successful, failed, etc.
	AverageDuration     float64                  `json:"average_duration_minutes"`
	FollowUpPending     int                      `json:"follow_up_pending"`
	InteractionFrequency float64                 `json:"interaction_frequency"` // Per customer per month
}
