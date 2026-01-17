package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/dto"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ISalesReportUseCase defines sales report query operations
type ISalesReportUseCase interface {
	GetSalesReport(ctx context.Context, filters dto.SalesReportRequest) (*dto.SalesReportResponse, error)
}

type salesReportUseCase struct {
	repo *postgres.SalesReportRepository
}

// NewSalesReportUseCase creates a new sales report use case
func NewSalesReportUseCase(repo *postgres.SalesReportRepository) ISalesReportUseCase {
	return &salesReportUseCase{repo: repo}
}

// GetSalesReport retrieves sales report with filters
func (uc *salesReportUseCase) GetSalesReport(ctx context.Context, filters dto.SalesReportRequest) (*dto.SalesReportResponse, error) {
	// Set defaults
	if filters.Limit == 0 {
		filters.Limit = 100
	}

	// Build query filters
	queryFilters := postgres.SalesReportFilters{
		StartDate: filters.StartDate,
		EndDate:   filters.EndDate,
		Status:    filters.Status,
		Limit:     filters.Limit,
		Offset:    filters.Offset,
	}

	if filters.ManagerID != nil {
		managerUUID, err := uuidv7.Parse(*filters.ManagerID)
		if err != nil {
			return nil, err
		}
		queryFilters.ManagerID = &managerUUID
	}

	if filters.CustomerID != nil {
		customerUUID, err := uuidv7.Parse(*filters.CustomerID)
		if err != nil {
			return nil, err
		}
		queryFilters.CustomerID = &customerUUID
	}

	// Query repository
	orders, totalCount, totalValue, err := uc.repo.GetSalesReport(ctx, queryFilters)
	if err != nil {
		return nil, err
	}

	// Map to DTOs
	orderSummaries := make([]dto.SalesOrderSummary, 0, len(orders))
	for _, order := range orders {
		summary := dto.SalesOrderSummary{
			OrderID:     order.OrderID.String(),
			CustomerID:  order.CustomerID.String(),
			TotalCents:  order.TotalCents,
			Currency:    order.Currency,
			Status:      order.Status,
			ItemCount:   order.ItemCount,
			ConfirmedAt: order.ConfirmedAt,
			UpdatedAt:   order.UpdatedAt,
		}

		if order.ManagerID != nil {
			managerIDStr := order.ManagerID.String()
			summary.ManagerID = &managerIDStr
		}

		orderSummaries = append(orderSummaries, summary)
	}

	// Determine currency (assume all orders same currency or default to first)
	currency := "UAH"
	if len(orderSummaries) > 0 {
		currency = orderSummaries[0].Currency
	}

	return &dto.SalesReportResponse{
		Orders:     orderSummaries,
		TotalCount: totalCount,
		TotalValue: totalValue,
		Currency:   currency,
	}, nil
}
