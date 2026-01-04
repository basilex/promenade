package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/billing/payment"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/valueobject"
)

// PaymentHandler handles HTTP requests for payments
type PaymentHandler struct {
	usecase payment.IUseCase
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(usecase payment.IUseCase) *PaymentHandler {
	return &PaymentHandler{
		usecase: usecase,
	}
}

// Create creates a new payment
// @Summary Create payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param payment body CreatePaymentRequest true "Payment data"
// @Success 201 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments [post]
func (h *PaymentHandler) Create(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	customerID, err := ParseUUID(req.CustomerID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	// Create Money value object
	amount, err := valueobject.NewMoney(req.Amount, req.Currency)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_AMOUNT", err.Error())
		return
	}

	// Create payment with basic fields
	p, err := h.usecase.CreatePayment(c.Request.Context(), customerID, amount, payment.PaymentMethod(req.Method))
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	// Link to invoice if provided
	if req.InvoiceID != nil {
		invoiceID, err := ParseUUID(*req.InvoiceID)
		if err != nil {
			response.ErrorResponse(c, http.StatusBadRequest, "INVALID_INVOICE_ID", "Invalid invoice ID format")
			return
		}
		if err := h.usecase.LinkToInvoice(c.Request.Context(), p.GetID(), invoiceID); err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "LINK_INVOICE_FAILED", err.Error())
			return
		}
	}

	// Add notes if provided
	if req.Notes != "" {
		if err := h.usecase.AddNote(c.Request.Context(), p.GetID(), req.Notes); err != nil {
			response.ErrorResponse(c, http.StatusInternalServerError, "ADD_NOTE_FAILED", err.Error())
			return
		}
	}

	// Reload payment to get updated state
	p, err = h.usecase.GetPayment(c.Request.Context(), p.GetID())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusCreated, ToPaymentResponse(p))
}

// GetByID retrieves a payment by ID
// @Summary Get payment by ID
// @Tags Payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} PaymentResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id} [get]
func (h *PaymentHandler) GetByID(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if errors.Is(err, payment.ErrPaymentNotFound) {
		response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// GetByNumber retrieves a payment by payment number
// @Summary Get payment by number
// @Tags Payments
// @Produce json
// @Param number path string true "Payment number"
// @Success 200 {object} PaymentResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/number/{number} [get]
func (h *PaymentHandler) GetByNumber(c *gin.Context) {
	paymentNo := c.Param("number")

	p, err := h.usecase.GetPaymentByNumber(c.Request.Context(), paymentNo)
	if errors.Is(err, payment.ErrPaymentNotFound) {
		response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// GetByTransactionID retrieves a payment by transaction ID
// @Summary Get payment by transaction ID
// @Tags Payments
// @Produce json
// @Param txId path string true "Transaction ID"
// @Success 200 {object} PaymentResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/transaction/{txId} [get]
func (h *PaymentHandler) GetByTransactionID(c *gin.Context) {
	transactionID := c.Param("txId")

	p, err := h.usecase.GetPaymentByTransactionID(c.Request.Context(), transactionID)
	if errors.Is(err, payment.ErrPaymentNotFound) {
		response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// Delete soft deletes a payment
// @Summary Delete payment
// @Tags Payments
// @Param id path string true "Payment ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id} [delete]
func (h *PaymentHandler) Delete(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	if err := h.usecase.DeletePayment(c.Request.Context(), id); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Payment deleted successfully"})
}

// List retrieves payments with pagination
// @Summary List payments
// @Tags Payments
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} PaymentResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments [get]
func (h *PaymentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	payments, err := h.usecase.ListPayments(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponseList(payments))
}

// ListByCustomer retrieves payments for a specific customer
// @Summary List payments by customer
// @Tags Payments
// @Produce json
// @Param customerId path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/customer/{customerId} [get]
func (h *PaymentHandler) ListByCustomer(c *gin.Context) {
	customerID, err := ParseUUID(c.Param("customerId"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	payments, err := h.usecase.ListPaymentsByCustomer(c.Request.Context(), customerID, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponseList(payments))
}

// ListByInvoice retrieves payments for a specific invoice
// @Summary List payments by invoice
// @Tags Payments
// @Produce json
// @Param invoiceId path string true "Invoice ID"
// @Success 200 {array} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/invoice/{invoiceId} [get]
func (h *PaymentHandler) ListByInvoice(c *gin.Context) {
	invoiceID, err := ParseUUID(c.Param("invoiceId"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_INVOICE_ID", "Invalid invoice ID format")
		return
	}

	payments, err := h.usecase.ListPaymentsByInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponseList(payments))
}

// ListByStatus retrieves payments by status
// @Summary List payments by status
// @Tags Payments
// @Produce json
// @Param status path string true "Payment status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} PaymentResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/status/{status} [get]
func (h *PaymentHandler) ListByStatus(c *gin.Context) {
	status := payment.PaymentStatus(c.Param("status"))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	payments, err := h.usecase.ListPaymentsByStatus(c.Request.Context(), status, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponseList(payments))
}

// LinkToInvoice links a payment to an invoice
// @Summary Link payment to invoice
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body LinkInvoiceRequest true "Invoice ID"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/link-invoice [post]
func (h *PaymentHandler) LinkToInvoice(c *gin.Context) {
	paymentID, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req LinkInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	invoiceID, err := ParseUUID(req.InvoiceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_INVOICE_ID", "Invalid invoice ID format")
		return
	}

	if err := h.usecase.LinkToInvoice(c.Request.Context(), paymentID, invoiceID); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "LINK_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), paymentID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// ProcessPayment starts processing a payment
// @Summary Process payment
// @Tags Payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} PaymentResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/process [post]
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req CompletePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.ProcessPayment(c.Request.Context(), id, req.TransactionID); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "PROCESS_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// CompletePayment marks a payment as completed
// @Summary Complete payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body CompletePaymentRequest true "Transaction ID"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/complete [post]
func (h *PaymentHandler) CompletePayment(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req CompletePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.CompletePayment(c.Request.Context(), id, req.TransactionID); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "COMPLETE_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// FailPayment marks a payment as failed
// @Summary Fail payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body FailPaymentRequest true "Failure reason"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/fail [post]
func (h *PaymentHandler) FailPayment(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req FailPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.FailPayment(c.Request.Context(), id, req.Reason); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "FAIL_PAYMENT_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// RefundPayment processes a refund
// @Summary Refund payment
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body RefundPaymentRequest true "Refund data"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/refund [post]
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	refundAmount, err := valueobject.NewMoney(req.Amount, req.Currency)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_AMOUNT", err.Error())
		return
	}

	if err := h.usecase.RefundPayment(c.Request.Context(), id, refundAmount); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if errors.Is(err, payment.ErrRefundAmountExceedsPayment) {
			response.ErrorResponse(c, http.StatusBadRequest, "REFUND_EXCEEDS_PAYMENT", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "REFUND_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// CancelPayment cancels a payment
// @Summary Cancel payment
// @Tags Payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} PaymentResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/cancel [post]
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	if err := h.usecase.CancelPayment(c.Request.Context(), id, ""); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "CANCEL_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// SetCardDetails sets card details for a payment
// @Summary Set card details
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body SetCardDetailsRequest true "Card details"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/card-details [put]
func (h *PaymentHandler) SetCardDetails(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req SetCardDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.SetCardDetails(c.Request.Context(), id, req.Last4, req.Brand); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "SET_CARD_DETAILS_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// SetProvider sets payment provider
// @Summary Set payment provider
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body SetProviderRequest true "Provider"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/provider [put]
func (h *PaymentHandler) SetProvider(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req SetProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.SetProvider(c.Request.Context(), id, req.Provider); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "SET_PROVIDER_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// AddNote adds a note to a payment
// @Summary Add note
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param request body AddNoteRequest true "Note"
// @Success 200 {object} PaymentResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/{id}/notes [post]
func (h *PaymentHandler) AddNote(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID format")
		return
	}

	var req AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.AddNote(c.Request.Context(), id, req.Note); err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "ADD_NOTE_FAILED", err.Error())
		return
	}

	// Reload payment to get updated state
	p, err := h.usecase.GetPayment(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RELOAD_FAILED", err.Error())
		return
	}

	response.Success(c, ToPaymentResponse(p))
}

// GetTotalByCustomer retrieves total payment amount for a customer
// @Summary Get total by customer
// @Tags Payments
// @Produce json
// @Param customerId path string true "Customer ID"
// @Success 200 {object} TotalResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/totals/customer/{customerId} [get]
func (h *PaymentHandler) GetTotalByCustomer(c *gin.Context) {
	customerID, err := ParseUUID(c.Param("customerId"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CUSTOMER_ID", "Invalid customer ID format")
		return
	}

	total, err := h.usecase.GetTotalByCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_TOTAL_FAILED", err.Error())
		return
	}

	response.Success(c, TotalResponse{
		Total:    total.Amount,
		Currency: total.Currency,
	})
}

// GetTotalByInvoice retrieves total payment amount for an invoice
// @Summary Get total by invoice
// @Tags Payments
// @Produce json
// @Param invoiceId path string true "Invoice ID"
// @Success 200 {object} TotalResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /billing/payments/totals/invoice/{invoiceId} [get]
func (h *PaymentHandler) GetTotalByInvoice(c *gin.Context) {
	invoiceID, err := ParseUUID(c.Param("invoiceId"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_INVOICE_ID", "Invalid invoice ID format")
		return
	}

	total, err := h.usecase.GetTotalByInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_TOTAL_FAILED", err.Error())
		return
	}

	response.Success(c, TotalResponse{
		Total:    total.Amount,
		Currency: total.Currency,
	})
}
