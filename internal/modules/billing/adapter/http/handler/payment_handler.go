package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/billing/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type PaymentHandler struct {
	paymentUC usecase.IPaymentUseCase
}

func NewPaymentHandler(paymentUC usecase.IPaymentUseCase) *PaymentHandler {
	return &PaymentHandler{paymentUC: paymentUC}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", errors.New("user not authenticated"))
		return
	}

	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	var invoiceID uuidv7.UUID
	if req.InvoiceID != nil {
		id, err := uuidv7.Parse(*req.InvoiceID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_INVOICE_ID", errors.New("invalid invoice ID"))
			return
		}
		invoiceID = id
	}

	payment, err := h.paymentUC.CreatePayment(c.Request.Context(), userID.(uuidv7.UUID), invoiceID, req.Amount, req.Currency, entity.PaymentMethod(req.PaymentMethod), "")
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToPaymentResponse(payment))
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	idStr := c.Param("id")
	paymentID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PAYMENT_ID", err)
		return
	}

	payment, err := h.paymentUC.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPaymentResponse(payment))
}

func (h *PaymentHandler) GetMyPayments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", errors.New("user not authenticated"))
		return
	}

	payments, err := h.paymentUC.GetUserPayments(c.Request.Context(), userID.(uuidv7.UUID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPaymentListResponse(payments))
}

func (h *PaymentHandler) ListPayments(c *gin.Context) {
	payments, err := h.paymentUC.ListPayments(c.Request.Context(), nil, 100, 0)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPaymentListResponse(payments))
}

func (h *PaymentHandler) CompletePayment(c *gin.Context) {
	idStr := c.Param("id")
	paymentID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PAYMENT_ID", errors.New("invalid payment ID"))
		return
	}

	transactionID := c.PostForm("transaction_id")
	if transactionID == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_TRANSACTION_ID", errors.New("transaction ID required"))
		return
	}

	if err := h.paymentUC.CompletePayment(c.Request.Context(), paymentID, transactionID); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	idStr := c.Param("id")
	paymentID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_PAYMENT_ID", errors.New("invalid payment ID"))
		return
	}

	var req dto.RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err)
		return
	}

	if err := h.paymentUC.RefundPayment(c.Request.Context(), paymentID, req.Reason); err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err)
		return
	}

	c.Status(http.StatusNoContent)
}
