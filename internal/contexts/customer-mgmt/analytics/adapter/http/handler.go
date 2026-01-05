package http

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics"
	"github.com/basilex/promenade/pkg/response"
)

// AnalyticsHandler handles analytics HTTP requests
type AnalyticsHandler struct {
	analyticsUC analytics.IUseCase
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsUC analytics.IUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsUC: analyticsUC,
	}
}

// GetCustomerOverview godoc
// @Summary Get customer overview statistics
// @Description Returns high-level customer statistics including counts by state/tier, new/churned this month
// @Tags Customer Analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Customer overview statistics"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/customers/overview [get]
func (h *AnalyticsHandler) GetCustomerOverview(c *gin.Context) {
	overview, err := h.analyticsUC.GetCustomerOverview(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, ToCustomerOverviewResponse(overview))
}

// GetCustomerLifecycle godoc
// @Summary Get customer lifecycle transitions
// @Description Returns customer state transitions over time periods (month/week/quarter)
// @Tags Customer Analytics
// @Accept json
// @Produce json
// @Param period query string false "Time period granularity" Enums(month, week, quarter) default(month)
// @Success 200 {array} map[string]interface{} "Customer lifecycle data"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/customers/lifecycle [get]
func (h *AnalyticsHandler) GetCustomerLifecycle(c *gin.Context) {
	var req CustomerLifecycleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	period := req.Period
	if period == "" {
		period = "month"
	}

	lifecycles, err := h.analyticsUC.GetCustomerLifecycle(c.Request.Context(), period)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerLifecycleResponse(lifecycles))
}

// GetCustomerSegmentation godoc
// @Summary Get customer segmentation
// @Description Returns customer distribution by segments (tiers) with revenue data
// @Tags Customer Analytics
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{} "Customer segmentation data"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/customers/segmentation [get]
func (h *AnalyticsHandler) GetCustomerSegmentation(c *gin.Context) {
	segments, err := h.analyticsUC.GetCustomerSegmentation(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, segments)
}

// GetDealPipeline godoc
// @Summary Get deal pipeline statistics
// @Description Returns pipeline statistics including total deals, value, and breakdown by stage
// @Tags Deal Analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Deal pipeline statistics"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/deals/pipeline [get]
func (h *AnalyticsHandler) GetDealPipeline(c *gin.Context) {
	pipeline, err := h.analyticsUC.GetDealPipeline(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, ToDealPipelineResponse(pipeline))
}

// GetDealConversions godoc
// @Summary Get deal conversion rates
// @Description Returns conversion rates between deal stages
// @Tags Deal Analytics
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{} "Deal conversion rates"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/deals/conversions [get]
func (h *AnalyticsHandler) GetDealConversions(c *gin.Context) {
	conversions, err := h.analyticsUC.GetDealConversions(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, conversions)
}

// GetSalesRepPerformance godoc
// @Summary Get sales rep performance
// @Description Returns top N sales reps by performance metrics or specific sales rep data
// @Tags Sales Analytics
// @Accept json
// @Produce json
// @Param top_n query int false "Number of top performers to return" default(10) minimum(1) maximum(100)
// @Param sales_rep_id query string false "Specific sales rep UUID to filter" format(uuid)
// @Success 200 {array} map[string]interface{} "Sales rep performance data"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 404 {object} response.Response "Sales rep not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/sales-reps/performance [get]
func (h *AnalyticsHandler) GetSalesRepPerformance(c *gin.Context) {
	var req SalesRepPerformanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// If specific sales rep requested
	if req.SalesRepID != "" {
		salesRepID, err := req.ParseSalesRepID()
		if err != nil {
			response.BadRequest(c, "Invalid sales rep ID")
			return
		}

		perf, err := h.analyticsUC.GetSalesRepPerformanceByID(c.Request.Context(), salesRepID)
		if err != nil {
			response.NotFound(c, err.Error())
			return
		}

		response.Success(c, ToSalesRepPerformanceResponse([]analytics.SalesRepPerformance{*perf})[0])
		return
	}

	// Get top N performers
	topN := req.TopN
	if topN == 0 {
		topN = 10
	}

	performances, err := h.analyticsUC.GetSalesRepPerformance(c.Request.Context(), topN)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToSalesRepPerformanceResponse(performances))
}

// GetRevenueTimeSeries godoc
// @Summary Get revenue time series
// @Description Returns revenue data over time with specified granularity
// @Tags Revenue Analytics
// @Accept json
// @Produce json
// @Param start_date query string true "Start date" format(date) example(2025-01-01)
// @Param end_date query string true "End date" format(date) example(2025-12-31)
// @Param granularity query string false "Time granularity" Enums(day, week, month) default(month)
// @Success 200 {array} map[string]interface{} "Revenue time series data"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/revenue/time-series [get]
func (h *AnalyticsHandler) GetRevenueTimeSeries(c *gin.Context) {
	var req RevenueTimeSeriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	startDate, endDate, granularity, err := req.Parse()
	if err != nil {
		response.BadRequest(c, "Invalid date format")
		return
	}

	series, err := h.analyticsUC.GetRevenueTimeSeries(c.Request.Context(), startDate, endDate, granularity)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToRevenueTimeSeriesResponse(series))
}

// GetInteractionInsights godoc
// @Summary Get interaction insights
// @Description Returns interaction analytics including counts by type/outcome and average duration
// @Tags Interaction Analytics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Interaction insights"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /customer-mgmt/analytics/interactions/insights [get]
func (h *AnalyticsHandler) GetInteractionInsights(c *gin.Context) {
	insights, err := h.analyticsUC.GetInteractionInsights(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, ToInteractionInsightsResponse(insights))
}
