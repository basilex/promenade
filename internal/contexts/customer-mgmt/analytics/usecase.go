package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IUseCase defines analytics business logic interface
type IUseCase interface {
	// Customer Analytics
	GetCustomerOverview(ctx context.Context) (*CustomerOverview, error)
	GetCustomerLifecycle(ctx context.Context, period string) ([]CustomerLifecycle, error)
	GetCustomerSegmentation(ctx context.Context) ([]CustomerSegmentation, error)

	// Deal Analytics
	GetDealPipeline(ctx context.Context) (*DealPipeline, error)
	GetDealConversions(ctx context.Context) ([]DealConversion, error)

	// Sales Rep Analytics
	GetSalesRepPerformance(ctx context.Context, topN int) ([]SalesRepPerformance, error)
	GetSalesRepPerformanceByID(ctx context.Context, salesRepID uuidv7.UUID) (*SalesRepPerformance, error)

	// Revenue Analytics
	GetRevenueTimeSeries(ctx context.Context, startDate, endDate time.Time, granularity string) ([]RevenueTimeSeries, error)

	// Interaction Analytics
	GetInteractionInsights(ctx context.Context) (*InteractionInsights, error)
}

type useCase struct {
	db *sqlx.DB
}

// NewUseCase creates a new analytics use case
func NewUseCase(db *sqlx.DB) IUseCase {
	return &useCase{
		db: db,
	}
}

// GetCustomerOverview returns high-level customer statistics
func (uc *useCase) GetCustomerOverview(ctx context.Context) (*CustomerOverview, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'lead' THEN 1 END) as leads,
			COUNT(CASE WHEN status = 'prospect' THEN 1 END) as prospects,
			COUNT(CASE WHEN status = 'customer' THEN 1 END) as customers,
			COUNT(CASE WHEN status = 'churned' THEN 1 END) as churned,
			COUNT(CASE WHEN tier = 'free' THEN 1 END) as free_tier,
			COUNT(CASE WHEN tier = 'basic' THEN 1 END) as basic_tier,
			COUNT(CASE WHEN tier = 'pro' THEN 1 END) as pro_tier,
			COUNT(CASE WHEN tier = 'enterprise' THEN 1 END) as enterprise_tier,
			COUNT(CASE WHEN created_at >= DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as new_this_month,
			COUNT(CASE WHEN status = 'churned' AND updated_at >= DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as churned_this_month,
			AVG(EXTRACT(EPOCH FROM (COALESCE(updated_at, CURRENT_TIMESTAMP) - created_at)) / 604800) as avg_lifetime_weeks
		FROM customer_mgmt_customers
		WHERE deleted_at IS NULL
	`

	var row struct {
		Total            int     `db:"total"`
		Leads            int     `db:"leads"`
		Prospects        int     `db:"prospects"`
		Customers        int     `db:"customers"`
		Churned          int     `db:"churned"`
		FreeTier         int     `db:"free_tier"`
		BasicTier        int     `db:"basic_tier"`
		ProTier          int     `db:"pro_tier"`
		EnterpriseTier   int     `db:"enterprise_tier"`
		NewThisMonth     int     `db:"new_this_month"`
		ChurnedThisMonth int     `db:"churned_this_month"`
		AvgLifetimeWeeks float64 `db:"avg_lifetime_weeks"`
	}

	if err := uc.db.GetContext(ctx, &row, query); err != nil {
		logger.FromContext(ctx).Error("Failed to get customer overview", "error", err)
		return nil, fmt.Errorf("failed to get customer overview: %w", err)
	}

	overview := &CustomerOverview{
		TotalCustomers:      row.Total,
		NewThisMonth:        row.NewThisMonth,
		ChurnedThisMonth:    row.ChurnedThisMonth,
		AverageLifetimeWeeks: row.AvgLifetimeWeeks,
		ByState: map[string]int{
			"lead":     row.Leads,
			"prospect": row.Prospects,
			"customer": row.Customers,
			"churned":  row.Churned,
		},
		ByTier: map[string]int{
			"free":       row.FreeTier,
			"basic":      row.BasicTier,
			"pro":        row.ProTier,
			"enterprise": row.EnterpriseTier,
		},
	}

	// Get top 5 sales reps
	topReps, err := uc.GetSalesRepPerformance(ctx, 5)
	if err == nil {
		overview.TopSalesReps = topReps
	}

	return overview, nil
}

// GetCustomerLifecycle returns customer state transitions over time
func (uc *useCase) GetCustomerLifecycle(ctx context.Context, period string) ([]CustomerLifecycle, error) {
	// Validate period
	var dateFormat string
	switch period {
	case "month":
		dateFormat = "YYYY-MM"
	case "week":
		dateFormat = "IYYY-IW"
	case "quarter":
		dateFormat = "YYYY-Q"
	default:
		return nil, fmt.Errorf("invalid period: %s (allowed: month, week, quarter)", period)
	}

	query := fmt.Sprintf(`
		SELECT 
			TO_CHAR(created_at, '%s') as period,
			COUNT(CASE WHEN status = 'lead' THEN 1 END) as lead_count,
			COUNT(CASE WHEN status = 'prospect' THEN 1 END) as prospect_count,
			COUNT(CASE WHEN status = 'customer' THEN 1 END) as customer_count,
			COUNT(CASE WHEN status = 'churned' THEN 1 END) as churned_count
		FROM customer_mgmt_customers
		WHERE deleted_at IS NULL 
		  AND created_at >= CURRENT_DATE - INTERVAL '12 months'
		GROUP BY period
		ORDER BY period DESC
		LIMIT 12
	`, dateFormat)

	var rows []struct {
		Period        string `db:"period"`
		LeadCount     int    `db:"lead_count"`
		ProspectCount int    `db:"prospect_count"`
		CustomerCount int    `db:"customer_count"`
		ChurnedCount  int    `db:"churned_count"`
	}

	if err := uc.db.SelectContext(ctx, &rows, query); err != nil {
		logger.FromContext(ctx).Error("Failed to get customer lifecycle", "error", err)
		return nil, fmt.Errorf("failed to get customer lifecycle: %w", err)
	}

	result := make([]CustomerLifecycle, len(rows))
	for i, row := range rows {
		result[i] = CustomerLifecycle{
			Period:        row.Period,
			LeadCount:     row.LeadCount,
			ProspectCount: row.ProspectCount,
			CustomerCount: row.CustomerCount,
			ChurnedCount:  row.ChurnedCount,
		}
	}

	return result, nil
}

// GetCustomerSegmentation returns customer distribution by tier with revenue
func (uc *useCase) GetCustomerSegmentation(ctx context.Context) ([]CustomerSegmentation, error) {
	query := `
		SELECT 
			c.tier as segment_name,
			COUNT(c.id) as count,
			AVG(EXTRACT(EPOCH FROM (COALESCE(c.updated_at, CURRENT_TIMESTAMP) - c.created_at)) / 604800) as avg_lifetime_weeks,
			COALESCE(SUM(d.value_cents), 0) as total_revenue_cents
		FROM customer_mgmt_customers c
		LEFT JOIN customer_deals d ON c.id = d.customer_id AND d.stage = 'closed_won' AND d.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
		GROUP BY c.tier
		ORDER BY count DESC
	`

	var rows []struct {
		SegmentName      string  `db:"segment_name"`
		Count            int     `db:"count"`
		AvgLifetimeWeeks float64 `db:"avg_lifetime_weeks"`
		TotalRevenueCents int64   `db:"total_revenue_cents"`
	}

	if err := uc.db.SelectContext(ctx, &rows, query); err != nil {
		logger.FromContext(ctx).Error("Failed to get customer segmentation", "error", err)
		return nil, fmt.Errorf("failed to get customer segmentation: %w", err)
	}

	total := 0
	for _, row := range rows {
		total += row.Count
	}

	result := make([]CustomerSegmentation, len(rows))
	for i, row := range rows {
		percentage := 0.0
		if total > 0 {
			percentage = float64(row.Count) / float64(total) * 100
		}

		totalRevenue, err := valueobject.NewMoney(row.TotalRevenueCents, "USD")
		if err != nil {
			return nil, fmt.Errorf("failed to create money: %w", err)
		}

		result[i] = CustomerSegmentation{
			SegmentName:         row.SegmentName,
			Count:               row.Count,
			Percentage:          percentage,
			AverageLifetimeWeeks: row.AvgLifetimeWeeks,
			TotalRevenue:        totalRevenue,
		}
	}

	return result, nil
}

// GetDealPipeline returns pipeline statistics by stage
func (uc *useCase) GetDealPipeline(ctx context.Context) (*DealPipeline, error) {
	query := `
		SELECT 
			COUNT(*) as total_deals,
			COALESCE(SUM(value_cents), 0) as total_value_cents,
			ROUND(COALESCE(AVG(value_cents), 0)) as avg_deal_cents,
			COUNT(CASE WHEN stage = 'closed_won' THEN 1 END) as won_count,
			COUNT(CASE WHEN stage = 'closed_lost' THEN 1 END) as lost_count,
			COALESCE(AVG(CASE WHEN actual_close_date IS NOT NULL THEN EXTRACT(EPOCH FROM (actual_close_date - created_at)) / 86400 END), 0) as avg_days_to_close
		FROM customer_deals
		WHERE deleted_at IS NULL
		  AND stage NOT IN ('closed_won', 'closed_lost')
	`

	var row struct {
		TotalDeals       int     `db:"total_deals"`
		TotalValueCents  int64   `db:"total_value_cents"`
		AvgDealCents     int64   `db:"avg_deal_cents"`
		WonCount         int     `db:"won_count"`
		LostCount        int     `db:"lost_count"`
		AvgDaysToClose   float64 `db:"avg_days_to_close"`
	}

	if err := uc.db.GetContext(ctx, &row, query); err != nil {
		logger.FromContext(ctx).Error("Failed to get deal pipeline", "error", err)
		return nil, fmt.Errorf("failed to get deal pipeline: %w", err)
	}

	winRate := 0.0
	if row.WonCount+row.LostCount > 0 {
		winRate = float64(row.WonCount) / float64(row.WonCount+row.LostCount) * 100
	}

	totalValue, err := valueobject.NewMoney(row.TotalValueCents, "USD")
	if err != nil {
		return nil, fmt.Errorf("failed to create money: %w", err)
	}

	averageDealSize, err := valueobject.NewMoney(row.AvgDealCents, "USD")
	if err != nil {
		return nil, fmt.Errorf("failed to create money: %w", err)
	}

	pipeline := &DealPipeline{
		TotalDeals:         row.TotalDeals,
		TotalValue:         totalValue,
		AverageDealSize:    averageDealSize,
		WinRate:            winRate,
		AverageDaysToClose: row.AvgDaysToClose,
		ByStage:            []DealStageStats{},
	}

	return pipeline, nil
}

// GetDealConversions returns conversion rates between stages
func (uc *useCase) GetDealConversions(ctx context.Context) ([]DealConversion, error) {
	// TODO: Implement stage conversion tracking
	// This requires tracking deal stage history
	return []DealConversion{}, nil
}

// GetSalesRepPerformance returns top N sales reps by performance
func (uc *useCase) GetSalesRepPerformance(ctx context.Context, topN int) ([]SalesRepPerformance, error) {
	query := `
		SELECT 
			c.assigned_to,
			COUNT(DISTINCT c.id) as active_customers,
			COUNT(CASE WHEN d.stage = 'closed_won' THEN 1 END) as deals_won,
			COUNT(CASE WHEN d.stage = 'closed_lost' THEN 1 END) as deals_lost,
			COALESCE(SUM(CASE WHEN d.stage = 'closed_won' THEN d.value_cents ELSE 0 END), 0) as total_revenue_cents,
			ROUND(COALESCE(AVG(CASE WHEN d.stage = 'closed_won' THEN d.value_cents END), 0)) as avg_deal_cents,
			COALESCE(AVG(CASE WHEN d.actual_close_date IS NOT NULL THEN EXTRACT(EPOCH FROM (d.actual_close_date - d.created_at)) / 86400 END), 0) as avg_days_to_close
		FROM customer_mgmt_customers c
		LEFT JOIN customer_deals d ON c.id = d.customer_id AND d.deleted_at IS NULL
		WHERE c.deleted_at IS NULL AND c.assigned_to IS NOT NULL
		GROUP BY c.assigned_to
		ORDER BY total_revenue_cents DESC
		LIMIT $1
	`

	var rows []struct {
		SalesRepID       uuidv7.UUID `db:"assigned_to"`
		ActiveCustomers  int         `db:"active_customers"`
		DealsWon         int         `db:"deals_won"`
		DealsLost        int         `db:"deals_lost"`
		TotalRevenueCents int64       `db:"total_revenue_cents"`
		AvgDealCents     int64       `db:"avg_deal_cents"`
		AvgDaysToClose   float64     `db:"avg_days_to_close"`
	}

	if err := uc.db.SelectContext(ctx, &rows, query, topN); err != nil {
		logger.FromContext(ctx).Error("Failed to get sales rep performance", "error", err)
		return nil, fmt.Errorf("failed to get sales rep performance: %w", err)
	}

	result := make([]SalesRepPerformance, len(rows))
	for i, row := range rows {
		winRate := 0.0
		if row.DealsWon+row.DealsLost > 0 {
			winRate = float64(row.DealsWon) / float64(row.DealsWon+row.DealsLost) * 100
		}

		totalRevenue, err := valueobject.NewMoney(row.TotalRevenueCents, "USD")
		if err != nil {
			return nil, fmt.Errorf("failed to create money: %w", err)
		}

		averageDealSize, err := valueobject.NewMoney(row.AvgDealCents, "USD")
		if err != nil {
			return nil, fmt.Errorf("failed to create money: %w", err)
		}

		result[i] = SalesRepPerformance{
			SalesRepID:         row.SalesRepID,
			SalesRepName:       "",
			ActiveCustomers:    row.ActiveCustomers,
			DealsWon:           row.DealsWon,
			DealsLost:          row.DealsLost,
			TotalRevenue:       totalRevenue,
			AverageDealSize:    averageDealSize,
			WinRate:            winRate,
			AverageDaysToClose: row.AvgDaysToClose,
		}
	}

	return result, nil
}

// GetSalesRepPerformanceByID returns performance for specific sales rep
func (uc *useCase) GetSalesRepPerformanceByID(ctx context.Context, salesRepID uuidv7.UUID) (*SalesRepPerformance, error) {
	query := `
		SELECT 
			c.assigned_to,
			COUNT(DISTINCT c.id) as active_customers,
			COUNT(CASE WHEN d.stage = 'closed_won' THEN 1 END) as deals_won,
			COUNT(CASE WHEN d.stage = 'closed_lost' THEN 1 END) as deals_lost,
			COALESCE(SUM(CASE WHEN d.stage = 'closed_won' THEN d.value_cents ELSE 0 END), 0) as total_revenue_cents,
			ROUND(COALESCE(AVG(CASE WHEN d.stage = 'closed_won' THEN d.value_cents END), 0)) as avg_deal_cents,
			COALESCE(AVG(CASE WHEN d.actual_close_date IS NOT NULL THEN EXTRACT(EPOCH FROM (d.actual_close_date - d.created_at)) / 86400 END), 0) as avg_days_to_close
		FROM customer_mgmt_customers c
		LEFT JOIN customer_deals d ON c.id = d.customer_id AND d.deleted_at IS NULL
		WHERE c.deleted_at IS NULL AND c.assigned_to = $1
		GROUP BY c.assigned_to
	`

	var row struct {
		SalesRepID       uuidv7.UUID `db:"assigned_to"`
		ActiveCustomers  int         `db:"active_customers"`
		DealsWon         int         `db:"deals_won"`
		DealsLost        int         `db:"deals_lost"`
		TotalRevenueCents int64       `db:"total_revenue_cents"`
		AvgDealCents     int64       `db:"avg_deal_cents"`
		AvgDaysToClose   float64     `db:"avg_days_to_close"`
	}

	if err := uc.db.GetContext(ctx, &row, query, salesRepID); err != nil {
		logger.FromContext(ctx).Error("Failed to get sales rep performance", "error", err, "sales_rep_id", salesRepID)
		return nil, fmt.Errorf("failed to get sales rep performance: %w", err)
	}

	winRate := 0.0
	if row.DealsWon+row.DealsLost > 0 {
		winRate = float64(row.DealsWon) / float64(row.DealsWon+row.DealsLost) * 100
	}

	totalRevenue, err := valueobject.NewMoney(row.TotalRevenueCents, "USD")
	if err != nil {
		return nil, fmt.Errorf("failed to create money: %w", err)
	}

	averageDealSize, err := valueobject.NewMoney(row.AvgDealCents, "USD")
	if err != nil {
		return nil, fmt.Errorf("failed to create money: %w", err)
	}

	return &SalesRepPerformance{
		SalesRepID:         row.SalesRepID,
		SalesRepName:       "",
		ActiveCustomers:    row.ActiveCustomers,
		DealsWon:           row.DealsWon,
		DealsLost:          row.DealsLost,
		TotalRevenue:       totalRevenue,
		AverageDealSize:    averageDealSize,
		WinRate:            winRate,
		AverageDaysToClose: row.AvgDaysToClose,
	}, nil
}

// GetRevenueTimeSeries returns revenue data over time
func (uc *useCase) GetRevenueTimeSeries(ctx context.Context, startDate, endDate time.Time, granularity string) ([]RevenueTimeSeries, error) {
	// Validate granularity
	var dateFormat string
	var truncFunc string
	switch granularity {
	case "day":
		dateFormat = "YYYY-MM-DD"
		truncFunc = "day"
	case "week":
		dateFormat = "IYYY-IW"
		truncFunc = "week"
	case "month":
		dateFormat = "YYYY-MM"
		truncFunc = "month"
	default:
		return nil, fmt.Errorf("invalid granularity: %s (allowed: day, week, month)", granularity)
	}

	query := fmt.Sprintf(`
		SELECT 
			TO_CHAR(DATE_TRUNC('%s', actual_close_date), '%s') as period,
			DATE_TRUNC('%s', actual_close_date) as timestamp,
			COUNT(*) as deals_won,
			COALESCE(SUM(value_cents), 0) as revenue_cents
		FROM customer_deals
		WHERE deleted_at IS NULL 
		  AND stage = 'closed_won'
		  AND actual_close_date BETWEEN $1 AND $2
		GROUP BY period, timestamp
		ORDER BY timestamp DESC
	`, truncFunc, dateFormat, truncFunc)

	var rows []struct {
		Period      string    `db:"period"`
		Timestamp   time.Time `db:"timestamp"`
		DealsWon    int       `db:"deals_won"`
		RevenueCents int64     `db:"revenue_cents"`
	}

	if err := uc.db.SelectContext(ctx, &rows, query, startDate, endDate); err != nil {
		logger.FromContext(ctx).Error("Failed to get revenue time series", "error", err)
		return nil, fmt.Errorf("failed to get revenue time series: %w", err)
	}

	result := make([]RevenueTimeSeries, len(rows))
	for i, row := range rows {
		revenue, err := valueobject.NewMoney(row.RevenueCents, "USD")
		if err != nil {
			return nil, fmt.Errorf("failed to create money: %w", err)
		}

		result[i] = RevenueTimeSeries{
			Period:    row.Period,
			Timestamp: row.Timestamp,
			DealsWon:  row.DealsWon,
			Revenue:   revenue,
		}
	}

	return result, nil
}

// GetInteractionInsights returns interaction analytics
func (uc *useCase) GetInteractionInsights(ctx context.Context) (*InteractionInsights, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN type = 'call' THEN 1 END) as call_count,
			COUNT(CASE WHEN type = 'email' THEN 1 END) as email_count,
			COUNT(CASE WHEN type = 'meeting' THEN 1 END) as meeting_count,
			COUNT(CASE WHEN type = 'note' THEN 1 END) as note_count,
			COUNT(CASE WHEN outcome = 'successful' THEN 1 END) as successful_count,
			COUNT(CASE WHEN outcome = 'not_interested' THEN 1 END) as failed_count,
			COUNT(CASE WHEN outcome = 'no_answer' THEN 1 END) as no_answer_count,
			COUNT(CASE WHEN outcome = 'scheduled' THEN 1 END) as scheduled_count,
			COUNT(CASE WHEN outcome = 'voicemail' THEN 1 END) as cancelled_count,
			COALESCE(AVG(duration_sec) / 60.0, 0) as avg_duration_minutes,
			COUNT(CASE WHEN follow_up_required = true THEN 1 END) as follow_up_pending
		FROM customer_interactions
		WHERE deleted_at IS NULL
	`

	var row struct {
		Total              int     `db:"total"`
		CallCount          int     `db:"call_count"`
		EmailCount         int     `db:"email_count"`
		MeetingCount       int     `db:"meeting_count"`
		NoteCount          int     `db:"note_count"`
		SuccessfulCount    int     `db:"successful_count"`
		FailedCount        int     `db:"failed_count"`
		NoAnswerCount      int     `db:"no_answer_count"`
		ScheduledCount     int     `db:"scheduled_count"`
		CancelledCount     int     `db:"cancelled_count"`
		AvgDurationMinutes float64 `db:"avg_duration_minutes"`
		FollowUpPending    int     `db:"follow_up_pending"`
	}

	if err := uc.db.GetContext(ctx, &row, query); err != nil {
		logger.FromContext(ctx).Error("Failed to get interaction insights", "error", err)
		return nil, fmt.Errorf("failed to get interaction insights: %w", err)
	}

	return &InteractionInsights{
		TotalInteractions: row.Total,
		ByType: map[string]int{
			"call":    row.CallCount,
			"email":   row.EmailCount,
			"meeting": row.MeetingCount,
			"note":    row.NoteCount,
		},
		ByOutcome: map[string]int{
			"successful": row.SuccessfulCount,
			"failed":     row.FailedCount,
			"no_answer":  row.NoAnswerCount,
			"scheduled":  row.ScheduledCount,
			"cancelled":  row.CancelledCount,
		},
		AverageDuration:      row.AvgDurationMinutes,
		FollowUpPending:      row.FollowUpPending,
		InteractionFrequency: 0, // TODO: Calculate per customer per month
	}, nil
}
