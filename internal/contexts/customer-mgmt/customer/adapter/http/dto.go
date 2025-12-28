package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
)

// CreateCustomerRequest represents B2C customer creation request
type CreateCustomerRequest struct {
	Name       string   `json:"name" binding:"required,min=1,max=255"`
	Email      string   `json:"email" binding:"required,email"`
	Phone      string   `json:"phone" binding:"omitempty"`
	Source     string   `json:"source" binding:"required,min=1,max=100"`
	AssignedTo string   `json:"assigned_to" binding:"required,uuid"`
	Tags       []string `json:"tags" binding:"omitempty"`
}

// CreateB2BCustomerRequest represents B2B customer creation request
type CreateB2BCustomerRequest struct {
	Name       string   `json:"name" binding:"required,min=1,max=255"`
	Email      string   `json:"email" binding:"required,email"`
	Phone      string   `json:"phone" binding:"omitempty"`
	Source     string   `json:"source" binding:"required,min=1,max=100"`
	CompanyID  string   `json:"company_id" binding:"required,uuid"`
	AssignedTo string   `json:"assigned_to" binding:"required,uuid"`
	Tags       []string `json:"tags" binding:"omitempty"`
}

// UpdateCustomerRequest represents customer update request
type UpdateCustomerRequest struct {
	Name   *string  `json:"name" binding:"omitempty,min=1,max=255"`
	Email  *string  `json:"email" binding:"omitempty,email"`
	Phone  *string  `json:"phone" binding:"omitempty"`
	Source *string  `json:"source" binding:"omitempty,min=1,max=100"`
	Tags   []string `json:"tags" binding:"omitempty"`
}

// SetPhoneRequest represents phone update request
type SetPhoneRequest struct {
	Phone string `json:"phone" binding:"omitempty"`
}

// ChurnCustomerRequest represents churn request
type ChurnCustomerRequest struct {
	Reason string `json:"reason" binding:"required,min=1"`
}

// ReassignCustomerRequest represents reassignment request
type ReassignCustomerRequest struct {
	NewRepID string `json:"new_rep_id" binding:"required,uuid"`
}

// AddTagRequest represents tag addition request
type AddTagRequest struct {
	Tag string `json:"tag" binding:"required,min=1"`
}

// RemoveTagRequest represents tag removal request (not used in URL)
type RemoveTagRequest struct {
	Tag string `json:"tag" binding:"required,min=1"`
}

// CustomerResponse represents customer data in API responses
type CustomerResponse struct {
	ID              string    `json:"id"`
	UserID          *string   `json:"user_id,omitempty"`
	CompanyID       *string   `json:"company_id,omitempty"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           *string   `json:"phone,omitempty"`
	Status          string    `json:"status"`
	Tier            string    `json:"tier"`
	Source          string    `json:"source"`
	AssignedTo      string    `json:"assigned_to"`
	Tags            []string  `json:"tags"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	LastContactedAt *time.Time `json:"last_contacted_at,omitempty"`
	ConvertedAt     *time.Time `json:"converted_at,omitempty"`
	ChurnedAt       *time.Time `json:"churned_at,omitempty"`
	ChurnReason     string    `json:"churn_reason,omitempty"`
	
	// Helper flags
	IsLead           bool `json:"is_lead"`
	IsProspect       bool `json:"is_prospect"`
	IsActiveCustomer bool `json:"is_active_customer"`
	IsChurned        bool `json:"is_churned"`
	IsB2B            bool `json:"is_b2b"`
	HasAccount       bool `json:"has_account"`
}

// CustomerStatsResponse represents customer statistics
type CustomerStatsResponse struct {
	TotalCustomers   int            `json:"total_customers"`
	ActiveCustomers  int            `json:"active_customers"`
	ChurnedCustomers int            `json:"churned_customers"`
	ByStatus         map[string]int `json:"by_status"`
	ByTier           map[string]int `json:"by_tier"`
}

// ToCustomerResponse converts entity to response DTO
func ToCustomerResponse(c *customer.Customer) *CustomerResponse {
	resp := &CustomerResponse{
		ID:         c.ID.String(),
		Name:       c.Name,
		Email:      c.Email.Value(),
		Status:     string(c.Status),
		Tier:       string(c.Tier),
		Source:     c.Source,
		AssignedTo: c.AssignedTo.String(),
		Tags:       c.Tags,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		
		// Helper flags
		IsLead:           c.IsLead(),
		IsProspect:       c.IsProspect(),
		IsActiveCustomer: c.IsActiveCustomer(),
		IsChurned:        c.IsChurned(),
		IsB2B:            c.IsB2B(),
		HasAccount:       c.HasAccount(),
	}
	
	// Optional UserID
	if c.UserID != nil {
		userID := c.UserID.String()
		resp.UserID = &userID
	}
	
	// Optional CompanyID
	if c.CompanyID != nil {
		companyID := c.CompanyID.String()
		resp.CompanyID = &companyID
	}
	
	// Optional Phone
	if c.Phone != nil {
		phone := c.Phone.Value()
		resp.Phone = &phone
	}
	
	// Optional timestamps
	resp.LastContactedAt = c.LastContactedAt
	resp.ConvertedAt = c.ConvertedAt
	resp.ChurnedAt = c.ChurnedAt
	resp.ChurnReason = c.ChurnReason
	
	return resp
}

// ToCustomerListResponse converts entity list to response DTOs
func ToCustomerListResponse(customers []*customer.Customer) []*CustomerResponse {
	responses := make([]*CustomerResponse, 0, len(customers))
	for _, c := range customers {
		responses = append(responses, ToCustomerResponse(c))
	}
	return responses
}

// ToCustomerStatsResponse converts stats to response DTO
func ToCustomerStatsResponse(stats *customer.CustomerStats) *CustomerStatsResponse {
	byStatus := make(map[string]int)
	for status, count := range stats.ByStatus {
		byStatus[string(status)] = count
	}
	
	byTier := make(map[string]int)
	for tier, count := range stats.ByTier {
		byTier[string(tier)] = count
	}
	
	return &CustomerStatsResponse{
		TotalCustomers:   stats.TotalCustomers,
		ActiveCustomers:  stats.ActiveCustomers,
		ChurnedCustomers: stats.ChurnedCustomers,
		ByStatus:         byStatus,
		ByTier:           byTier,
	}
}
