package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/dto"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/usecase"
	"github.com/basilex/promenade/pkg/response"
)

// SalesReportHandler handles sales report HTTP requests
type SalesReportHandler struct {
	usecase usecase.ISalesReportUseCase
}

// NewSalesReportHandler creates a new sales report handler
func NewSalesReportHandler(uc usecase.ISalesReportUseCase) *SalesReportHandler {
	return &SalesReportHandler{usecase: uc}
}

// RegisterRoutes registers analytics routes
func (h *SalesReportHandler) RegisterRoutes(router *gin.RouterGroup) {
	analytics := router.Group("/analytics")
	{
		analytics.GET("/sales-report", h.GetSalesReport)
	}
}

// GetSalesReport handles GET /api/v1/analytics/sales-report
// @Summary Get sales report
// @Description Retrieve sales report with filters (date range, manager, customer, status)
// @Tags Analytics
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (RFC3339)"
// @Param end_date query string false "End date (RFC3339)"
// @Param manager_id query string false "Manager UUID"
// @Param customer_id query string false "Customer UUID"
// @Param status query string false "Order status" Enums(confirmed, paid, fulfilled, cancelled)
// @Param limit query int false "Limit (default 100, max 1000)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.SalesReportResponse
// @Failure 400 {object} response.Error
// @Failure 500 {object} response.Error
// @Router /api/v1/analytics/sales-report [get]
func (h *SalesReportHandler) GetSalesReport(c *gin.Context) {
	var req dto.SalesReportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	report, err := h.usecase.GetSalesReport(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to retrieve sales report")
		return
	}

	c.JSON(http.StatusOK, report)
}
