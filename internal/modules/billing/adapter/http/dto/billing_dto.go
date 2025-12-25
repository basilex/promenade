package dto

import (
	"time"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
)

// Plan DTOs

type CreatePlanRequest struct {
	Name        string   `json:"name" binding:"required,min=3,max=100"`
	Slug        string   `json:"slug" binding:"required,min=3,max=100"`
	Description string   `json:"description"`
	Price       int64    `json:"price" binding:"required,min=0"`
	Currency    string   `json:"currency" binding:"required,len=3"`
	Interval    string   `json:"interval" binding:"required,oneof=month year"`
	Features    []string `json:"features"`
	MaxUsers    *int     `json:"max_users"`
	MaxProjects *int     `json:"max_projects"`
	MaxStorage  *int64   `json:"max_storage"`
	TrialDays   int      `json:"trial_days" binding:"min=0"`
	SortOrder   int      `json:"sort_order" binding:"min=0"`
}

type UpdatePlanRequest struct {
	Name        string   `json:"name" binding:"required,min=3,max=100"`
	Description string   `json:"description"`
	Price       int64    `json:"price" binding:"required,min=0"`
	Features    []string `json:"features"`
	MaxUsers    *int     `json:"max_users"`
	MaxProjects *int     `json:"max_projects"`
	MaxStorage  *int64   `json:"max_storage"`
	TrialDays   int      `json:"trial_days" binding:"min=0"`
	SortOrder   int      `json:"sort_order" binding:"min=0"`
}

type PlanResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Currency    string    `json:"currency"`
	Interval    string    `json:"interval"`
	Status      string    `json:"status"`
	Features    []string  `json:"features"`
	MaxUsers    *int      `json:"max_users"`
	MaxProjects *int      `json:"max_projects"`
	MaxStorage  *int64    `json:"max_storage"`
	TrialDays   int       `json:"trial_days"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToPlanResponse(plan *entity.Plan) PlanResponse {
	return PlanResponse{
		ID:          plan.ID.String(),
		Name:        plan.Name,
		Slug:        plan.Slug,
		Description: plan.Description,
		Price:       plan.Amount,
		Currency:    plan.Currency,
		Interval:    string(plan.Interval),
		Status:      string(plan.Status),
		Features:    plan.Features,
		MaxUsers:    plan.MaxUsers,
		MaxProjects: plan.MaxProjects,
		MaxStorage:  plan.MaxStorage,
		TrialDays:   plan.TrialDays,
		// SortOrder removed
		CreatedAt: plan.CreatedAt,
		UpdatedAt: plan.UpdatedAt,
	}
}

func ToPlanListResponse(plans []*entity.Plan) []PlanResponse {
	response := make([]PlanResponse, len(plans))
	for i, plan := range plans {
		response[i] = ToPlanResponse(plan)
	}
	return response
}

// Subscription DTOs

type CreateSubscriptionRequest struct {
	PlanID string `json:"plan_id" binding:"required,uuid"`
}

type SubscriptionResponse struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	PlanID             string     `json:"plan_id"`
	Status             string     `json:"status"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	TrialStart         *time.Time `json:"trial_start,omitempty"`
	TrialEnd           *time.Time `json:"trial_end,omitempty"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func ToSubscriptionResponse(subscription *entity.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:                 subscription.ID.String(),
		UserID:             subscription.UserID.String(),
		PlanID:             subscription.PlanID.String(),
		Status:             string(subscription.Status),
		CurrentPeriodStart: subscription.CurrentPeriodStart,
		CurrentPeriodEnd:   subscription.CurrentPeriodEnd,
		TrialStart:         subscription.TrialStart,
		TrialEnd:           subscription.TrialEnd,
		CanceledAt:         subscription.CanceledAt,
		CancelAtPeriodEnd:  subscription.CancelAtPeriodEnd,
		CreatedAt:          subscription.CreatedAt,
		UpdatedAt:          subscription.UpdatedAt,
	}
}

func ToSubscriptionListResponse(subscriptions []*entity.Subscription) []SubscriptionResponse {
	response := make([]SubscriptionResponse, len(subscriptions))
	for i, subscription := range subscriptions {
		response[i] = ToSubscriptionResponse(subscription)
	}
	return response
}

type CancelSubscriptionRequest struct {
	Immediately bool `json:"immediately"`
}

type UpgradeSubscriptionRequest struct {
	NewPlanID string `json:"new_plan_id" binding:"required,uuid"`
}

// Invoice DTOs

type CreateInvoiceRequest struct {
	SubscriptionID *string `json:"subscription_id" binding:"omitempty,uuid"`
	Currency       string  `json:"currency" binding:"required,len=3"`
	Subtotal       int64   `json:"subtotal" binding:"required,min=0"`
	Tax            int64   `json:"tax" binding:"min=0"`
	Description    string  `json:"description"`
}

type InvoiceResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	SubscriptionID *string    `json:"subscription_id,omitempty"`
	InvoiceNumber  string     `json:"invoice_number"`
	Status         string     `json:"status"`
	Currency       string     `json:"currency"`
	Subtotal       int64      `json:"subtotal"`
	Tax            int64      `json:"tax"`
	Total          int64      `json:"total"`
	AmountPaid     int64      `json:"amount_paid"`
	AmountDue      int64      `json:"amount_due"`
	Description    string     `json:"description"`
	PeriodStart    *time.Time `json:"period_start,omitempty"`
	PeriodEnd      *time.Time `json:"period_end,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	VoidedAt       *time.Time `json:"voided_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToInvoiceResponse(invoice *entity.Invoice) InvoiceResponse {
	subscriptionID := invoice.SubscriptionID.String()

	return InvoiceResponse{
		ID:             invoice.ID.String(),
		UserID:         "", // Not in entity, derive from context if needed
		SubscriptionID: &subscriptionID,
		InvoiceNumber:  invoice.InvoiceNumber,
		Status:         string(invoice.Status),
		Currency:       invoice.Currency,
		Subtotal:       invoice.SubtotalAmount,
		Tax:            invoice.TaxAmount,
		Total:          invoice.TotalAmount,
		AmountPaid:     invoice.AmountPaid,
		AmountDue:      invoice.AmountDue,
		Description:    "",  // Not in entity
		PeriodStart:    nil, // Not in entity
		PeriodEnd:      nil, // Not in entity
		DueDate:        invoice.DueDate,
		PaidAt:         invoice.PaidAt,
		VoidedAt:       nil, // Not in entity
		CreatedAt:      invoice.CreatedAt,
		UpdatedAt:      invoice.UpdatedAt,
	}
}

func ToInvoiceListResponse(invoices []*entity.Invoice) []InvoiceResponse {
	response := make([]InvoiceResponse, len(invoices))
	for i, invoice := range invoices {
		response[i] = ToInvoiceResponse(invoice)
	}
	return response
}

// Payment DTOs

type CreatePaymentRequest struct {
	InvoiceID     *string `json:"invoice_id" binding:"omitempty,uuid"`
	Amount        int64   `json:"amount" binding:"required,min=1"`
	Currency      string  `json:"currency" binding:"required,len=3"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=card bank_transfer paypal stripe liqpay"`
}

type PaymentResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	InvoiceID      *string    `json:"invoice_id,omitempty"`
	SubscriptionID *string    `json:"subscription_id,omitempty"`
	Amount         int64      `json:"amount"`
	Currency       string     `json:"currency"`
	Status         string     `json:"status"`
	PaymentMethod  string     `json:"payment_method"`
	TransactionID  *string    `json:"transaction_id,omitempty"`
	FailureCode    *string    `json:"failure_code,omitempty"`
	FailureMessage *string    `json:"failure_message,omitempty"`
	RefundedAmount int64      `json:"refunded_amount"`
	RefundedAt     *time.Time `json:"refunded_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToPaymentResponse(payment *entity.Payment) PaymentResponse {
	iid := payment.InvoiceID.String()
	invoiceID := &iid

	transactionID := payment.TransactionID

	var failureMessage *string
	if payment.FailureReason != "" {
		failureMessage = &payment.FailureReason
	}

	return PaymentResponse{
		ID:             payment.ID.String(),
		UserID:         payment.UserID.String(),
		InvoiceID:      invoiceID,
		SubscriptionID: nil, // Not in Payment entity
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		Status:         string(payment.Status),
		PaymentMethod:  string(payment.Method),
		TransactionID:  &transactionID,
		FailureCode:    nil, // No longer separate failure code
		FailureMessage: failureMessage,
		RefundedAmount: 0, // No longer tracked separately (always 0 or full amount)
		RefundedAt:     payment.RefundedAt,
		CreatedAt:      payment.CreatedAt,
		UpdatedAt:      payment.UpdatedAt,
	}
}

func ToPaymentListResponse(payments []*entity.Payment) []PaymentResponse {
	response := make([]PaymentResponse, len(payments))
	for i, payment := range payments {
		response[i] = ToPaymentResponse(payment)
	}
	return response
}

type RefundPaymentRequest struct {
	Reason string `json:"reason" binding:"required,min=3,max=500"`
}
