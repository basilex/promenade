package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CompanyHandler handles HTTP requests for companies
type CompanyHandler struct {
	companyUC company.IUseCase
}

// NewCompanyHandler creates a new company handler
func NewCompanyHandler(companyUC company.IUseCase) *CompanyHandler {
	return &CompanyHandler{
		companyUC: companyUC,
	}
}

// Create godoc
// @Summary Create a new company
// @Description Create a new company (B2B customer)
// @Tags Companies
// @Accept json
// @Produce json
// @Param body body CreateCompanyRequest true "Company creation data"
// @Success 201 {object} CompanyResponse
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response "Company already exists"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies [post]
// @Security Bearer
func (h *CompanyHandler) Create(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Parse optional fields
	email, err := parseEmail(req.Email)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		return
	}

	phone, err := parsePhone(req.Phone, req.PhoneCountryCode)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_PHONE", err.Error())
		return
	}

	address, err := parseAddress(
		req.AddressLine1, req.AddressLine2,
		req.City, req.StateProvince,
		req.PostalCode, req.Country,
	)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ADDRESS", err.Error())
		return
	}

	parentID, err := parseParentCompanyID(req.ParentCompanyID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARENT_ID", err.Error())
		return
	}

	// Create company
	comp, err := h.companyUC.CreateCompany(
		c.Request.Context(),
		req.Name,
		&req.LegalName,
		req.Type,
		req.TaxID,
		req.RegistrationNumber,
		req.Website,
		email,
		phone,
		address,
		req.Industry,
		req.Size,
		req.EmployeeCount,
		req.Revenue,
		req.Currency,
		req.Description,
		parentID,
	)
	if err != nil {
		if errors.Is(err, company.ErrCompanyAlreadyExists) {
			response.ErrorResponse(c, http.StatusConflict, "COMPANY_ALREADY_EXISTS", "Company already exists")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create company")
		return
	}

	response.Created(c, ToCompanyResponse(comp))
}

// GetByID godoc
// @Summary Get company by ID
// @Description Retrieve a company by its ID
// @Tags Companies
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id} [get]
// @Security Bearer
func (h *CompanyHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid company ID format")
		return
	}

	comp, err := h.companyUC.GetCompany(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to retrieve company")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// GetByName godoc
// @Summary Get company by name
// @Description Retrieve a company by its name
// @Tags Companies
// @Produce json
// @Param name path string true "Company name"
// @Success 200 {object} CompanyResponse
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/name/{name} [get]
// @Security Bearer
func (h *CompanyHandler) GetByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "MISSING_NAME", "Company name is required")
		return
	}

	comp, err := h.companyUC.GetCompanyByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to retrieve company")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// GetByTaxID godoc
// @Summary Get company by tax ID
// @Description Retrieve a company by its tax ID
// @Tags Companies
// @Produce json
// @Param taxId path string true "Tax ID"
// @Success 200 {object} CompanyResponse
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/tax/{taxId} [get]
// @Security Bearer
func (h *CompanyHandler) GetByTaxID(c *gin.Context) {
	taxID := c.Param("taxId")
	if taxID == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "MISSING_TAX_ID", "Tax ID is required")
		return
	}

	comp, err := h.companyUC.GetCompanyByTaxID(c.Request.Context(), taxID)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "GET_FAILED", "Failed to retrieve company")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// List godoc
// @Summary List all companies
// @Description Retrieve a paginated list of companies
// @Tags Companies
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} CompanyListResponse
// @Failure 400 {object} response.Response "Invalid pagination"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies [get]
// @Security Bearer
func (h *CompanyHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c)

	companies, total, err := h.companyUC.ListCompanies(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list companies")
		return
	}

	response.Success(c, ToCompanyListResponse(companies, total, page, pageSize))
}

// ListByIndustry godoc
// @Summary List companies by industry
// @Description Retrieve a paginated list of companies filtered by industry
// @Tags Companies
// @Produce json
// @Param industry path string true "Industry name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} CompanyListResponse
// @Failure 400 {object} response.Response "Invalid pagination"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/industry/{industry} [get]
// @Security Bearer
func (h *CompanyHandler) ListByIndustry(c *gin.Context) {
	industry := c.Param("industry")
	if industry == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "MISSING_INDUSTRY", "Industry is required")
		return
	}

	page, pageSize := parsePagination(c)

	companies, total, err := h.companyUC.ListCompaniesByIndustry(c.Request.Context(), industry, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list companies by industry")
		return
	}

	response.Success(c, ToCompanyListResponse(companies, total, page, pageSize))
}

// ListBySize godoc
// @Summary List companies by size
// @Description Retrieve a paginated list of companies filtered by size
// @Tags Companies
// @Produce json
// @Param size path string true "Company size" Enums(micro, small, medium, large, enterprise)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} CompanyListResponse
// @Failure 400 {object} response.Response "Invalid size or pagination"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/size/{size} [get]
// @Security Bearer
func (h *CompanyHandler) ListBySize(c *gin.Context) {
	sizeStr := c.Param("size")
	if sizeStr == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "MISSING_SIZE", "Size is required")
		return
	}

	size := company.CompanySize(sizeStr)
	page, pageSize := parsePagination(c)

	companies, total, err := h.companyUC.ListCompaniesBySize(c.Request.Context(), string(size), page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list companies by size")
		return
	}

	response.Success(c, ToCompanyListResponse(companies, total, page, pageSize))
}

// ListSubsidiaries godoc
// @Summary List subsidiaries of a company
// @Description Retrieve a paginated list of subsidiaries (child companies) for a parent company
// @Tags Companies
// @Produce json
// @Param id path string true "Parent Company ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} CompanyListResponse
// @Failure 400 {object} response.Response "Invalid ID or pagination"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id}/subsidiaries [get]
// @Security Bearer
func (h *CompanyHandler) ListSubsidiaries(c *gin.Context) {
	parentID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid parent company ID format")
		return
	}

	companies, err := h.companyUC.ListSubsidiaries(c.Request.Context(), parentID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list subsidiaries")
		return
	}

	// Return as simple list without pagination
	response.Success(c, gin.H{
		"companies": ToCompanyListResponse(companies, len(companies), 1, len(companies)).Companies,
		"total":     len(companies),
	})
}

// UpdateBasicInfo godoc
// @Summary Update company basic info
// @Description Update name, legal name, type, or tax ID
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Param body body UpdateCompanyBasicInfoRequest true "Basic info update data"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} response.Response "Invalid data"
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id}/basic-info [put]
// @Security Bearer
func (h *CompanyHandler) UpdateBasicInfo(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid company ID format")
		return
	}

	var req UpdateCompanyBasicInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Dereference required fields
	var name string
	if req.Name != nil {
		name = *req.Name
	}
	var companyType string
	if req.Type != nil {
		companyType = *req.Type
	}

	comp, err := h.companyUC.UpdateCompanyBasicInfo(
		c.Request.Context(),
		id,
		name,
		req.LegalName,
		companyType,
		req.TaxID,
		nil, // registrationNumber - not in UpdateCompanyBasicInfoRequest
	)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update company basic info")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// UpdateContactInfo godoc
// @Summary Update company contact info
// @Description Update website, email, phone, or address
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Param body body UpdateCompanyContactInfoRequest true "Contact info update data"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} response.Response "Invalid data"
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id}/contact-info [put]
// @Security Bearer
func (h *CompanyHandler) UpdateContactInfo(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid company ID format")
		return
	}

	var req UpdateCompanyContactInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Parse optional fields
	email, err := parseEmail(req.Email)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		return
	}

	phone, err := parsePhone(req.Phone, req.PhoneCountryCode)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_PHONE", err.Error())
		return
	}

	address, err := parseAddress(
		req.AddressLine1, req.AddressLine2,
		req.City, req.StateProvince,
		req.PostalCode, req.Country,
	)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ADDRESS", err.Error())
		return
	}

	comp, err := h.companyUC.UpdateCompanyContactInfo(
		c.Request.Context(),
		id,
		req.Website,
		email,
		phone,
		address,
	)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update company contact info")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// UpdateBusinessInfo godoc
// @Summary Update company business info
// @Description Update industry, size, employee count, revenue, or currency
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Param body body UpdateCompanyBusinessInfoRequest true "Business info update data"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} response.Response "Invalid data"
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id}/business-info [put]
// @Security Bearer
func (h *CompanyHandler) UpdateBusinessInfo(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid company ID format")
		return
	}

	var req UpdateCompanyBusinessInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Dereference required fields
	var size string
	if req.Size != nil {
		size = *req.Size
	}
	var employeeCount int
	if req.EmployeeCount != nil {
		employeeCount = *req.EmployeeCount
	}
	var revenue int64
	if req.Revenue != nil {
		revenue = *req.Revenue
	}
	var currency string
	if req.Currency != nil {
		currency = *req.Currency
	}

	comp, err := h.companyUC.UpdateCompanyBusinessInfo(
		c.Request.Context(),
		id,
		req.Industry,
		size,
		employeeCount,
		revenue,
		currency,
	)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update company business info")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// SetParentCompany godoc
// @Summary Set parent company
// @Description Set or update the parent company (for subsidiary relationships)
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Param body body SetParentCompanyRequest true "Parent company ID"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} response.Response "Invalid data"
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id}/parent [put]
// @Security Bearer
func (h *CompanyHandler) SetParentCompany(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid company ID format")
		return
	}

	var req SetParentCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	parentID, err := parseParentCompanyID(req.ParentCompanyID)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_PARENT_ID", err.Error())
		return
	}

	comp, err := h.companyUC.SetParentCompany(c.Request.Context(), id, parentID)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to set parent company")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// UpdateDescription godoc
// @Summary Update company description
// @Description Update the company description
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Param body body UpdateCompanyDescriptionRequest true "Description update data"
// @Success 200 {object} CompanyResponse
// @Failure 400 {object} response.Response "Invalid data"
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id}/description [put]
// @Security Bearer
func (h *CompanyHandler) UpdateDescription(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid company ID format")
		return
	}

	var req UpdateCompanyDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	comp, err := h.companyUC.UpdateCompanyDescription(c.Request.Context(), id, req.Description)
	if err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update company description")
		return
	}

	response.Success(c, ToCompanyResponse(comp))
}

// Delete godoc
// @Summary Delete company
// @Description Soft-delete a company
// @Tags Companies
// @Produce json
// @Param id path string true "Company ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 404 {object} response.Response "Company not found"
// @Failure 500 {object} response.Response
// @Router /customer-mgmt/companies/{id} [delete]
// @Security Bearer
func (h *CompanyHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid company ID format")
		return
	}

	if err := h.companyUC.DeleteCompany(c.Request.Context(), id); err != nil {
		if errors.Is(err, company.ErrCompanyNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "COMPANY_NOT_FOUND", "Company not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete company")
		return
	}

	response.Success(c, gin.H{"message": "Company deleted successfully"})
}

// Helper functions

// parsePagination extracts pagination parameters from query string
func parsePagination(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	return page, pageSize
}
