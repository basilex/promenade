package http

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CustomerHandler handles HTTP requests for customer operations
type CustomerHandler struct {
	usecase customer.ICustomerUseCase
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(usecase customer.ICustomerUseCase) *CustomerHandler {
	return &CustomerHandler{
		usecase: usecase,
	}
}

// Create handles POST /customers - Create B2C customer
func (h *CustomerHandler) Create(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse assigned_to UUID
	assignedTo, err := uuidv7.Parse(req.AssignedTo)
	if err != nil {
		response.BadRequest(c, "invalid assigned_to UUID")
		return
	}

	// Create customer
	created, err := h.usecase.CreateCustomer(c.Request.Context(), req.Name, req.Email, req.Source, assignedTo)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Set phone if provided
	if req.Phone != "" {
		if err := h.usecase.SetCustomerPhone(c.Request.Context(), created.ID, req.Phone); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}

	// Add tags if provided
	for _, tag := range req.Tags {
		if err := h.usecase.AddTagToCustomer(c.Request.Context(), created.ID, tag); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}

	// Reload to get updated data
	updated, err := h.usecase.GetCustomer(c.Request.Context(), created.ID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, ToCustomerResponse(updated))
}

// CreateB2B handles POST /customers/b2b - Create B2B customer
func (h *CustomerHandler) CreateB2B(c *gin.Context) {
	var req CreateB2BCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse UUIDs
	companyID, err := uuidv7.Parse(req.CompanyID)
	if err != nil {
		response.BadRequest(c, "invalid company_id UUID")
		return
	}

	assignedTo, err := uuidv7.Parse(req.AssignedTo)
	if err != nil {
		response.BadRequest(c, "invalid assigned_to UUID")
		return
	}

	// Create B2B customer
	created, err := h.usecase.CreateB2BCustomer(c.Request.Context(), req.Name, req.Email, req.Source, companyID, assignedTo)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Set phone if provided
	if req.Phone != "" {
		if err := h.usecase.SetCustomerPhone(c.Request.Context(), created.ID, req.Phone); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}

	// Add tags if provided
	for _, tag := range req.Tags {
		if err := h.usecase.AddTagToCustomer(c.Request.Context(), created.ID, tag); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}

	// Reload to get updated data
	updated, err := h.usecase.GetCustomer(c.Request.Context(), created.ID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, ToCustomerResponse(updated))
}

// GetByID handles GET /customers/:id
func (h *CustomerHandler) GetByID(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// GetByEmail handles GET /customers/by-email?email=xxx
func (h *CustomerHandler) GetByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "email query parameter is required")
		return
	}

	cust, err := h.usecase.GetCustomerByEmail(c.Request.Context(), email)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// List handles GET /customers
func (h *CustomerHandler) List(c *gin.Context) {
	limit := 20
	offset := 0

	customers, total, err := h.usecase.ListCustomers(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"customers": ToCustomerListResponse(customers),
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// Update handles PUT /customers/:id
func (h *CustomerHandler) Update(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get existing customer
	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	// Apply updates
	if req.Name != nil {
		cust.Name = *req.Name
	}
	if req.Email != nil {
		// Email update requires validation
		response.BadRequest(c, "email cannot be updated directly")
		return
	}
	if req.Phone != nil {
		if err := h.usecase.SetCustomerPhone(c.Request.Context(), customerID, *req.Phone); err != nil {
			response.InternalError(c, err.Error())
			return
		}
	}
	if req.Source != nil {
		cust.Source = *req.Source
	}
	if req.Tags != nil {
		cust.Tags = req.Tags
	}

	// Update customer
	if err := h.usecase.UpdateCustomer(c.Request.Context(), cust); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// Delete handles DELETE /customers/:id
func (h *CustomerHandler) Delete(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	if err := h.usecase.DeleteCustomer(c.Request.Context(), customerID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "customer deleted successfully"})
}

// QualifyAsProspect handles POST /customers/:id/qualify
func (h *CustomerHandler) QualifyAsProspect(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	if err := h.usecase.QualifyAsProspect(c.Request.Context(), customerID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// ConvertToCustomer handles POST /customers/:id/convert
func (h *CustomerHandler) ConvertToCustomer(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	if err := h.usecase.ConvertToCustomer(c.Request.Context(), customerID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// Churn handles POST /customers/:id/churn
func (h *CustomerHandler) Churn(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	var req ChurnCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.ChurnCustomer(c.Request.Context(), customerID, req.Reason); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// Reactivate handles POST /customers/:id/reactivate
func (h *CustomerHandler) Reactivate(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	if err := h.usecase.ReactivateCustomer(c.Request.Context(), customerID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// UpgradeTier handles POST /customers/:id/upgrade/:tier
func (h *CustomerHandler) UpgradeTier(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	tierStr := c.Param("tier")
	tier := customer.CustomerTier(tierStr)

	if err := h.usecase.UpgradeCustomerTier(c.Request.Context(), customerID, tier); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// DowngradeTier handles POST /customers/:id/downgrade/:tier
func (h *CustomerHandler) DowngradeTier(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	tierStr := c.Param("tier")
	tier := customer.CustomerTier(tierStr)

	if err := h.usecase.DowngradeCustomerTier(c.Request.Context(), customerID, tier); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// AddTag handles POST /customers/:id/tags
func (h *CustomerHandler) AddTag(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	var req AddTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.AddTagToCustomer(c.Request.Context(), customerID, req.Tag); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// RemoveTag handles DELETE /customers/:id/tags/:tag
func (h *CustomerHandler) RemoveTag(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	tag := c.Param("tag")
	if tag == "" {
		response.BadRequest(c, "tag parameter is required")
		return
	}

	if err := h.usecase.RemoveTagFromCustomer(c.Request.Context(), customerID, tag); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	cust, err := h.usecase.GetCustomer(c.Request.Context(), customerID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerResponse(cust))
}

// ListByStatus handles GET /customers/status/:status
func (h *CustomerHandler) ListByStatus(c *gin.Context) {
	statusStr := c.Param("status")
	status := customer.CustomerStatus(statusStr)

	limit := 20
	offset := 0

	customers, total, err := h.usecase.ListCustomersByStatus(c.Request.Context(), status, limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"customers": ToCustomerListResponse(customers),
		"total":     total,
		"status":    status,
	})
}

// ListByTier handles GET /customers/tier/:tier
func (h *CustomerHandler) ListByTier(c *gin.Context) {
	tierStr := c.Param("tier")
	tier := customer.CustomerTier(tierStr)

	limit := 20
	offset := 0

	customers, total, err := h.usecase.ListCustomersByTier(c.Request.Context(), tier, limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"customers": ToCustomerListResponse(customers),
		"total":     total,
		"tier":      tier,
	})
}

// GetStats handles GET /customers/stats
func (h *CustomerHandler) GetStats(c *gin.Context) {
	stats, err := h.usecase.GetCustomerStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToCustomerStatsResponse(stats))
}

// LinkToUser handles POST /customers/:id/link-user
func (h *CustomerHandler) LinkToUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required,uuid"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	userID, err := uuidv7.Parse(req.UserID)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	if err := h.usecase.LinkCustomerToUser(c.Request.Context(), customerID, userID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "customer linked to user successfully"})
}

// AssignTo handles POST /customers/:id/assign
func (h *CustomerHandler) AssignTo(c *gin.Context) {
	var req ReassignCustomerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	customerID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	newRepID, err := uuidv7.Parse(req.NewRepID)
	if err != nil {
		response.BadRequest(c, "invalid rep ID")
		return
	}

	if err := h.usecase.ReassignCustomer(c.Request.Context(), customerID, newRepID); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "customer reassigned successfully"})
}
