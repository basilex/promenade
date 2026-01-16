package http

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ReceiptHandler handles HTTP requests for receipt operations
type ReceiptHandler struct {
	usecase receipt.IUseCase
}

// NewReceiptHandler creates a new receipt handler
func NewReceiptHandler(usecase receipt.IUseCase) *ReceiptHandler {
	return &ReceiptHandler{usecase: usecase}
}

// Create creates a new receipt
// @Summary Create fiscal receipt
// @Description Create a new fiscal receipt
// @Tags Receipt
// @Accept json
// @Produce json
// @Param request body CreateReceiptRequest true "Receipt creation request"
// @Success 201 {object} response.Response{data=ReceiptResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/receipts [post]
func (h *ReceiptHandler) Create(c *gin.Context) {
	var req CreateReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cashRegisterID, err := uuidv7.Parse(req.CashRegisterID)
	if err != nil {
		response.BadRequest(c, "Invalid cash_register_id format")
		return
	}

	orderID, err := uuidv7.Parse(req.OrderID)
	if err != nil {
		response.BadRequest(c, "Invalid order_id format")
		return
	}

	createdBy, err := uuidv7.Parse(req.CreatedBy)
	if err != nil {
		response.BadRequest(c, "Invalid created_by format")
		return
	}

	lines := make([]receipt.ReceiptLine, len(req.Lines))
	for i, line := range req.Lines {
		lines[i] = receipt.ReceiptLine{
			Name:       line.Name,
			Quantity:   line.Quantity,
			PriceCents: line.PriceCents,
			TaxRate:    line.TaxRate,
		}
	}

	rec, err := h.usecase.CreateReceipt(
		c.Request.Context(),
		cashRegisterID,
		orderID,
		receipt.PaymentType(req.PaymentType),
		receipt.ReceiptType(req.ReceiptType),
		req.Currency,
		lines,
		createdBy,
	)
	if err != nil {
		switch {
		case errors.Is(err, receipt.ErrCashRegisterIDRequired):
			response.BadRequest(c, "Cash register ID is required")
		case errors.Is(err, receipt.ErrOrderIDRequired):
			response.BadRequest(c, "Order ID is required")
		case errors.Is(err, receipt.ErrPaymentTypeRequired):
			response.BadRequest(c, "Payment type is required")
		case errors.Is(err, receipt.ErrReceiptTypeRequired):
			response.BadRequest(c, "Receipt type is required")
		case errors.Is(err, receipt.ErrCurrencyRequired):
			response.BadRequest(c, "Currency is required")
		case errors.Is(err, receipt.ErrCreatedByRequired):
			response.BadRequest(c, "Created by is required")
		case errors.Is(err, receipt.ErrReceiptAlreadyExists):
			response.BadRequest(c, "Receipt already exists for order")
		case errors.Is(err, receipt.ErrReceiptLineNameRequired):
			response.BadRequest(c, "Receipt line name is required")
		case errors.Is(err, receipt.ErrReceiptLineQuantityInvalid):
			response.BadRequest(c, "Receipt line quantity must be greater than zero")
		case errors.Is(err, receipt.ErrReceiptLinePriceInvalid):
			response.BadRequest(c, "Receipt line price must be zero or greater")
		case errors.Is(err, receipt.ErrReceiptLineTaxRateInvalid):
			response.BadRequest(c, "Receipt line tax rate must be between 0 and 100")
		default:
			response.InternalError(c, "Failed to create receipt")
		}
		return
	}

	response.Created(c, ToReceiptResponse(rec))
}

// GetByID retrieves receipt by ID
// @Summary Get receipt by ID
// @Description Get receipt by ID
// @Tags Receipt
// @Produce json
// @Param id path string true "Receipt ID"
// @Success 200 {object} response.Response{data=ReceiptResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/receipts/{id} [get]
func (h *ReceiptHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid receipt ID format")
		return
	}

	rec, err := h.usecase.GetReceipt(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, receipt.ErrReceiptNotFound) {
			response.NotFound(c, "Receipt not found")
			return
		}
		response.InternalError(c, "Failed to retrieve receipt")
		return
	}

	response.Success(c, ToReceiptResponse(rec))
}

// List retrieves receipts
// @Summary List receipts
// @Description Get list of receipts
// @Tags Receipt
// @Produce json
// @Param cash_register_id query string false "Cash register ID"
// @Param order_id query string false "Order ID"
// @Param status query string false "Receipt status"
// @Success 200 {object} response.Response{data=[]ReceiptResponse}
// @Failure 500 {object} response.Response
// @Router /fiscal/receipts [get]
func (h *ReceiptHandler) List(c *gin.Context) {
	filters := &receipt.ListFilters{}

	if cashRegisterIDStr := c.Query("cash_register_id"); cashRegisterIDStr != "" {
		cashRegisterID, err := uuidv7.Parse(cashRegisterIDStr)
		if err != nil {
			response.BadRequest(c, "Invalid cash_register_id format")
			return
		}
		filters.CashRegisterID = &cashRegisterID
	}

	if orderIDStr := c.Query("order_id"); orderIDStr != "" {
		orderID, err := uuidv7.Parse(orderIDStr)
		if err != nil {
			response.BadRequest(c, "Invalid order_id format")
			return
		}
		filters.OrderID = &orderID
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := receipt.ReceiptStatus(statusStr)
		filters.Status = &status
	}

	receipts, err := h.usecase.ListReceipts(c.Request.Context(), filters)
	if err != nil {
		response.InternalError(c, "Failed to list receipts")
		return
	}

	response.Success(c, ToReceiptListResponse(receipts))
}

// MarkPrinted sets fiscal data and marks receipt as printed
// @Summary Mark receipt as printed
// @Description Set fiscal data and mark receipt as printed
// @Tags Receipt
// @Accept json
// @Produce json
// @Param id path string true "Receipt ID"
// @Param request body MarkPrintedRequest true "Fiscal data"
// @Success 200 {object} response.Response{data=ReceiptResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/receipts/{id}/print [post]
func (h *ReceiptHandler) MarkPrinted(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid receipt ID format")
		return
	}

	var req MarkPrintedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	printedBy, err := uuidv7.Parse(req.PrintedBy)
	if err != nil {
		response.BadRequest(c, "Invalid printed_by format")
		return
	}

	rec, err := h.usecase.MarkPrinted(c.Request.Context(), id, req.FiscalNumber, req.FiscalURL, req.QRCode, printedBy)
	if err != nil {
		switch {
		case errors.Is(err, receipt.ErrReceiptNotFound):
			response.NotFound(c, "Receipt not found")
		case errors.Is(err, receipt.ErrFiscalNumberRequired):
			response.BadRequest(c, "Fiscal number is required")
		case errors.Is(err, receipt.ErrReceiptAlreadyPrinted):
			response.BadRequest(c, "Receipt is already printed")
		case errors.Is(err, receipt.ErrReceiptAlreadyCancelled):
			response.BadRequest(c, "Receipt is cancelled")
		default:
			response.InternalError(c, "Failed to mark receipt as printed")
		}
		return
	}

	response.Success(c, ToReceiptResponse(rec))
}

// Cancel cancels receipt
// @Summary Cancel receipt
// @Description Cancel receipt
// @Tags Receipt
// @Accept json
// @Produce json
// @Param id path string true "Receipt ID"
// @Param request body CancelReceiptRequest true "Cancel request"
// @Success 200 {object} response.Response{data=ReceiptResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/receipts/{id}/cancel [post]
func (h *ReceiptHandler) Cancel(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid receipt ID format")
		return
	}

	var req CancelReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cancelledBy, err := uuidv7.Parse(req.CancelledBy)
	if err != nil {
		response.BadRequest(c, "Invalid cancelled_by format")
		return
	}

	rec, err := h.usecase.CancelReceipt(c.Request.Context(), id, req.Reason, cancelledBy)
	if err != nil {
		switch {
		case errors.Is(err, receipt.ErrReceiptNotFound):
			response.NotFound(c, "Receipt not found")
		case errors.Is(err, receipt.ErrReceiptCancelReasonRequired):
			response.BadRequest(c, "Cancellation reason is required")
		case errors.Is(err, receipt.ErrReceiptAlreadyCancelled):
			response.BadRequest(c, "Receipt is already cancelled")
		default:
			response.InternalError(c, "Failed to cancel receipt")
		}
		return
	}

	response.Success(c, ToReceiptResponse(rec))
}

// Delete deletes receipt
// @Summary Delete receipt
// @Description Delete receipt (soft delete)
// @Tags Receipt
// @Produce json
// @Param id path string true "Receipt ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /fiscal/receipts/{id} [delete]
func (h *ReceiptHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid receipt ID format")
		return
	}

	if err := h.usecase.DeleteReceipt(c.Request.Context(), id); err != nil {
		if errors.Is(err, receipt.ErrReceiptNotFound) {
			response.NotFound(c, "Receipt not found")
			return
		}
		response.InternalError(c, "Failed to delete receipt")
		return
	}

	response.Success(c, gin.H{"message": "Receipt deleted successfully"})
}
