package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/billing/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/billing/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type InvoiceHandler struct {
	invoiceUC usecase.IInvoiceUseCase
}

func NewInvoiceHandler(invoiceUC usecase.IInvoiceUseCase) *InvoiceHandler {
	return &InvoiceHandler{invoiceUC: invoiceUC}
}

func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req dto.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	// Parse subscription ID (optional)
	var subscriptionID uuidv7.UUID
	if req.SubscriptionID != nil {
		subscriptionID, _ = uuidv7.Parse(*req.SubscriptionID)
	}

	// Default due days is 30 if not specified
	dueDays := 30

	invoice, err := h.invoiceUC.CreateInvoice(c.Request.Context(), subscriptionID, req.Subtotal, req.Tax, req.Currency, dueDays)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToInvoiceResponse(invoice))
}

func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	idStr := c.Param("id")
	invoiceID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INVOICE_ID", err)
		return
	}

	invoice, err := h.invoiceUC.GetInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "INVOICE_NOT_FOUND", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToInvoiceResponse(invoice))
}

func (h *InvoiceHandler) GetMyInvoices(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", nil)
		return
	}
	_ = userID

	invoices, err := h.invoiceUC.ListInvoices(c.Request.Context(), nil, 100, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToInvoiceListResponse(invoices))
}

func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	invoices, err := h.invoiceUC.ListInvoices(c.Request.Context(), nil, 100, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToInvoiceListResponse(invoices))
}

func (h *InvoiceHandler) FinalizeInvoice(c *gin.Context) {
	idStr := c.Param("id")
	invoiceID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INVOICE_ID", err)
		return
	}

	if err := h.invoiceUC.FinalizeInvoice(c.Request.Context(), invoiceID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *InvoiceHandler) VoidInvoice(c *gin.Context) {
	idStr := c.Param("id")
	invoiceID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_INVOICE_ID", err)
		return
	}

	if err := h.invoiceUC.VoidInvoice(c.Request.Context(), invoiceID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}
