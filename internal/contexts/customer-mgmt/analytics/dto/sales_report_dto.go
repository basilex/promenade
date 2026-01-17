package dto

import "time"

// SalesReportRequest defines filters for sales report query
type SalesReportRequest struct {
	StartDate  *time.Time `form:"start_date" binding:"omitempty"`
	EndDate    *time.Time `form:"end_date" binding:"omitempty"`
	ManagerID  *string    `form:"manager_id" binding:"omitempty,uuid"`
	CustomerID *string    `form:"customer_id" binding:"omitempty,uuid"`
	Status     *string    `form:"status" binding:"omitempty,oneof=confirmed paid fulfilled cancelled"`
	Limit      int        `form:"limit" binding:"omitempty,min=1,max=1000"`
	Offset     int        `form:"offset" binding:"omitempty,min=0"`
}

// SalesReportResponse represents aggregated sales data
type SalesReportResponse struct {
	Orders     []SalesOrderSummary `json:"orders"`
	TotalCount int                 `json:"total_count"`
	TotalValue int64               `json:"total_value_cents"`
	Currency   string              `json:"currency"`
}

// SalesOrderSummary represents a single order in the report
type SalesOrderSummary struct {
	OrderID      string    `json:"order_id"`
	CustomerID   string    `json:"customer_id"`
	ManagerID    *string   `json:"manager_id,omitempty"`
	TotalCents   int64     `json:"total_cents"`
	Currency     string    `json:"currency"`
	Status       string    `json:"status"`
	ItemCount    int       `json:"item_count"`
	ConfirmedAt  time.Time `json:"confirmed_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}