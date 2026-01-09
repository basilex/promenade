package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/billing/invoice"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// InvoiceHandler handles HTTP requests for invoice operations
type InvoiceHandler struct {
	useCase invoice.IUseCase
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(useCase invoice.IUseCase) *InvoiceHandler {
	return &InvoiceHandler{
		useCase: useCase,
	}
}

// Create godoc
// @Summary Create invoice
// @Description Create a new draft invoice with auto-generated invoice number
// @Tags invoices
// @Accept json
// @Produce json
// @Param request body CreateInvoiceRequest true "Invoice creation request"
// @Success 201 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /invoices [post]
func (h *InvoiceHandler) Create(c *gin.Context) {
	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	customerID, err := ParseCustomerID(req.CustomerID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", err.Error())
		return
	}

	orderID, err := ParseOrderID(req.OrderID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ORDER_ID", err.Error())
		return
	}

	dueDate, err := ParseDueDate(req.DueDate)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_DUE_DATE", "due_date must be RFC3339 format")
		return
	}

	inv, err := h.useCase.CreateInvoice(c.Request.Context(), customerID, orderID, dueDate, req.Currency)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	response.Created(c, ToInvoiceResponse(inv))
}

// GetByID godoc
// @Summary Get invoice by ID
// @Description Retrieve invoice details by ID
// @Tags invoices
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id} [get]
func (h *InvoiceHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	inv, err := h.useCase.GetInvoice(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to retrieve invoice")
		return
	}

	response.Success(c, ToInvoiceResponse(inv))
}

// GetByNumber godoc
// @Summary Get invoice by number
// @Description Retrieve invoice details by invoice number
// @Tags invoices
// @Produce json
// @Param number path string true "Invoice Number (e.g., INV-2026-000001)"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 404 {object} response.Response
// @Router /invoices/number/{number} [get]
func (h *InvoiceHandler) GetByNumber(c *gin.Context) {
	invoiceNo := c.Param("number")

	inv, err := h.useCase.GetInvoiceByNumber(c.Request.Context(), invoiceNo)
	if err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to retrieve invoice")
		return
	}

	response.Success(c, ToInvoiceResponse(inv))
}

// Delete godoc
// @Summary Delete invoice
// @Description Soft delete a draft invoice (only draft status allowed)
// @Tags invoices
// @Param id path string true "Invoice ID (UUID)"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id} [delete]
func (h *InvoiceHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	if err := h.useCase.DeleteInvoice(c.Request.Context(), id); err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "DELETE_FAILED", "Failed to delete invoice")
		return
	}

	c.Status(http.StatusNoContent)
}

// AddLineItem godoc
// @Summary Add line item
// @Description Add a line item to a draft invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Param request body AddLineItemRequest true "Line item request"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/lines [post]
func (h *InvoiceHandler) AddLineItem(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	var req AddLineItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Get invoice to determine currency
	inv, err := h.useCase.GetInvoice(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to retrieve invoice")
		return
	}

	unitPrice, err := valueobject.NewMoney(req.UnitPrice, inv.Currency)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_UNIT_PRICE", "Invalid unit price format")
		return
	}

	updatedInv, err := h.useCase.AddLineItem(c.Request.Context(), id, req.Description, req.Quantity, unitPrice)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ADD_LINE_FAILED", "Failed to add line item")
		return
	}

	response.Success(c, ToInvoiceResponse(updatedInv))
}

// RemoveLineItem godoc
// @Summary Remove line item
// @Description Remove a line item from a draft invoice
// @Tags invoices
// @Param id path string true "Invoice ID (UUID)"
// @Param lineId path string true "Line Item ID (UUID)"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/lines/{lineId} [delete]
func (h *InvoiceHandler) RemoveLineItem(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	lineID, err := uuidv7.Parse(c.Param("lineId"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_LINE_ID", err.Error())
		return
	}

	updatedInv, err := h.useCase.RemoveLineItem(c.Request.Context(), id, lineID)
	if err != nil {
		if errors.Is(err, invoice.ErrInvoiceLineNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "LINE_NOT_FOUND", "Line item not found")
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "REMOVE_LINE_FAILED", "Failed to remove line item")
		return
	}

	response.Success(c, ToInvoiceResponse(updatedInv))
}

// UpdateLineItem godoc
// @Summary Update line item
// @Description Update the quantity of a line item in a draft invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Param lineId path string true "Line Item ID (UUID)"
// @Param request body UpdateLineItemRequest true "Update request"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/lines/{lineId} [put]
func (h *InvoiceHandler) UpdateLineItem(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	lineID, err := uuidv7.Parse(c.Param("lineId"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_LINE_ID", err.Error())
		return
	}

	var req UpdateLineItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	updatedInv, err := h.useCase.UpdateLineItem(c.Request.Context(), id, lineID, req.Quantity)
	if err != nil {
		if errors.Is(err, invoice.ErrInvoiceLineNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "LINE_NOT_FOUND", "Line item not found")
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "UPDATE_LINE_FAILED", "Failed to update line item")
		return
	}

	response.Success(c, ToInvoiceResponse(updatedInv))
}

// Send godoc
// @Summary Send invoice
// @Description Mark a draft invoice as sent (requires at least one line item)
// @Tags invoices
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/send [post]
func (h *InvoiceHandler) Send(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	if err := h.useCase.SendInvoice(c.Request.Context(), id); err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "SEND_FAILED", "Failed to send invoice")
		return
	}

	// Return updated invoice
	inv, _ := h.useCase.GetInvoice(c.Request.Context(), id)
	response.Success(c, ToInvoiceResponse(inv))
}

// MarkAsPaid godoc
// @Summary Mark as paid
// @Description Mark a sent invoice as paid
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Param request body MarkAsPaidRequest true "Payment date"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/pay [post]
func (h *InvoiceHandler) MarkAsPaid(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	var req MarkAsPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	paidDate, err := ParsePaidDate(req.PaidDate)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_PAID_DATE", "paid_date must be RFC3339 format")
		return
	}

	if err := h.useCase.MarkAsPaid(c.Request.Context(), id, paidDate); err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "MARK_PAID_FAILED", "Failed to mark invoice as paid")
		return
	}

	inv, _ := h.useCase.GetInvoice(c.Request.Context(), id)
	response.Success(c, ToInvoiceResponse(inv))
}

// Cancel godoc
// @Summary Cancel invoice
// @Description Cancel an invoice (cannot cancel paid/void/cancelled invoices)
// @Tags invoices
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/cancel [post]
func (h *InvoiceHandler) Cancel(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	if err := h.useCase.CancelInvoice(c.Request.Context(), id); err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "CANCEL_FAILED", "Failed to cancel invoice")
		return
	}

	inv, _ := h.useCase.GetInvoice(c.Request.Context(), id)
	response.Success(c, ToInvoiceResponse(inv))
}

// Void godoc
// @Summary Void invoice
// @Description Void an invoice for accounting purposes (cannot void paid/void/cancelled invoices)
// @Tags invoices
// @Param id path string true "Invoice ID (UUID)"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/void [post]
func (h *InvoiceHandler) Void(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	if err := h.useCase.VoidInvoice(c.Request.Context(), id); err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusBadRequest, "VOID_FAILED", "Failed to void invoice")
		return
	}

	inv, _ := h.useCase.GetInvoice(c.Request.Context(), id)
	response.Success(c, ToInvoiceResponse(inv))
}

// UpdateTax godoc
// @Summary Update tax amount
// @Description Update the tax amount on a draft invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID (UUID)"
// @Param request body UpdateTaxRequest true "Tax amount"
// @Success 200 {object} response.Response{data=InvoiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /invoices/{id}/tax [put]
func (h *InvoiceHandler) UpdateTax(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", err.Error())
		return
	}

	var req UpdateTaxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Get invoice to determine currency
	inv, err := h.useCase.GetInvoice(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, invoice.ErrInvoiceNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to retrieve invoice")
		return
	}

	taxAmount, err := valueobject.NewMoney(req.TaxAmount, inv.Currency)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_TAX_AMOUNT", "Invalid tax amount format")
		return
	}

	if err := h.useCase.UpdateTaxAmount(c.Request.Context(), id, taxAmount); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "UPDATE_TAX_FAILED", "Failed to update tax amount")
		return
	}

	updatedInv, _ := h.useCase.GetInvoice(c.Request.Context(), id)
	response.Success(c, ToInvoiceResponse(updatedInv))
}

// List godoc
// @Summary List invoices
// @Description Retrieve invoices with pagination and optional filters
// @Tags invoices
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Param customer_id query string false "Filter by customer ID"
// @Param order_id query string false "Filter by order ID"
// @Param status query string false "Filter by status (draft/sent/paid/overdue/cancelled/void)"
// @Success 200 {object} response.Response{data=[]InvoiceResponse}
// @Failure 400 {object} response.Response
// @Router /invoices [get]
func (h *InvoiceHandler) List(c *gin.Context) {
	var req ListInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	var invoices []*invoice.Invoice
	var total int64
	var err error

	// Apply filters
	if req.CustomerID != "" {
		customerID, parseErr := ParseCustomerID(req.CustomerID)
		if parseErr != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", parseErr.Error())
			return
		}
		var tmpTotal int
		invoices, tmpTotal, err = h.useCase.ListByCustomer(c.Request.Context(), customerID, req.Page, req.PageSize)
		total = int64(tmpTotal)
	} else if req.OrderID != "" {
		orderID, parseErr := ParseCustomerID(req.OrderID)
		if parseErr != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ORDER_ID", parseErr.Error())
			return
		}
		invoices, err = h.useCase.ListByOrder(c.Request.Context(), orderID)
		total = int64(len(invoices))
	} else if req.Status != "" {
		status := invoice.InvoiceStatus(req.Status)
		var tmpTotal int
		invoices, tmpTotal, err = h.useCase.ListByStatus(c.Request.Context(), status, req.Page, req.PageSize)
		total = int64(tmpTotal)
	} else {
		var tmpTotal int
		invoices, tmpTotal, err = h.useCase.ListInvoices(c.Request.Context(), req.Page, req.PageSize)
		total = int64(tmpTotal)
	}

	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list invoices")
		return
	}

	response.SuccessWithPagination(c, ToInvoiceListResponse(invoices), total, req.Page, req.PageSize)
}

// ListOverdue godoc
// @Summary List overdue invoices
// @Description Retrieve overdue invoices with pagination
// @Tags invoices
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} response.Response{data=[]InvoiceResponse}
// @Failure 400 {object} response.Response
// @Router /invoices/overdue [get]
func (h *InvoiceHandler) ListOverdue(c *gin.Context) {
	var req ListInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	var tmpTotal int
	invoices, tmpTotal, err := h.useCase.ListOverdue(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list invoices")
		return
	}

	response.SuccessWithPagination(c, ToInvoiceListResponse(invoices), int64(tmpTotal), req.Page, req.PageSize)
}
