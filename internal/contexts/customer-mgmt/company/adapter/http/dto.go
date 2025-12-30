package http

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// CreateCompanyRequest is the request body for creating a company
type CreateCompanyRequest struct {
	Name               string  `json:"name" binding:"required,min=2,max=255"`
	LegalName          string  `json:"legal_name" binding:"required,min=2,max=255"`
	Type               string  `json:"type" binding:"required,oneof=llc corporation sole_proprietor partnership non_profit other"`
	TaxID              *string `json:"tax_id,omitempty" binding:"omitempty,max=50"`
	RegistrationNumber *string `json:"registration_number,omitempty" binding:"omitempty,max=100"`
	Website            *string `json:"website,omitempty" binding:"omitempty,url"`
	Email              *string `json:"email,omitempty" binding:"omitempty,email"`
	Phone              *string `json:"phone,omitempty"`
	PhoneCountryCode   *string `json:"phone_country_code,omitempty"`
	AddressLine1       *string `json:"address_line1,omitempty" binding:"omitempty,max=255"`
	AddressLine2       *string `json:"address_line2,omitempty" binding:"omitempty,max=255"`
	City               *string `json:"city,omitempty" binding:"omitempty,max=100"`
	StateProvince      *string `json:"state_province,omitempty" binding:"omitempty,max=100"`
	PostalCode         *string `json:"postal_code,omitempty" binding:"omitempty,max=20"`
	Country            *string `json:"country,omitempty" binding:"omitempty,len=2"` // ISO 3166-1 alpha-2
	Industry           *string `json:"industry,omitempty" binding:"omitempty,max=100"`
	Size               string  `json:"size" binding:"required,oneof=micro small medium large enterprise"`
	EmployeeCount      int     `json:"employee_count" binding:"min=0"`
	AnnualRevenue      int64   `json:"annual_revenue" binding:"min=0"`
	Currency           string  `json:"currency" binding:"required,len=3"` // ISO 4217
	Description        *string `json:"description,omitempty" binding:"omitempty,max=2000"`
	ParentCompanyID    *string `json:"parent_company_id,omitempty"`
}

// UpdateCompanyBasicInfoRequest is the request body for updating basic company info
type UpdateCompanyBasicInfoRequest struct {
	Name      *string `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	LegalName *string `json:"legal_name,omitempty" binding:"omitempty,min=2,max=255"`
	Type      *string `json:"type,omitempty" binding:"omitempty,oneof=llc corporation sole_proprietor partnership non_profit other"`
	TaxID     *string `json:"tax_id,omitempty" binding:"omitempty,max=50"`
}

// UpdateCompanyContactInfoRequest is the request body for updating contact info
type UpdateCompanyContactInfoRequest struct {
	Website          *string `json:"website,omitempty" binding:"omitempty,url"`
	Email            *string `json:"email,omitempty" binding:"omitempty,email"`
	Phone            *string `json:"phone,omitempty"`
	PhoneCountryCode *string `json:"phone_country_code,omitempty"`
	AddressLine1     *string `json:"address_line1,omitempty" binding:"omitempty,max=255"`
	AddressLine2     *string `json:"address_line2,omitempty" binding:"omitempty,max=255"`
	City             *string `json:"city,omitempty" binding:"omitempty,max=100"`
	StateProvince    *string `json:"state_province,omitempty" binding:"omitempty,max=100"`
	PostalCode       *string `json:"postal_code,omitempty" binding:"omitempty,max=20"`
	Country          *string `json:"country,omitempty" binding:"omitempty,len=2"`
}

// UpdateCompanyBusinessInfoRequest is the request body for updating business info
type UpdateCompanyBusinessInfoRequest struct {
	Industry      *string `json:"industry,omitempty" binding:"omitempty,max=100"`
	Size          *string `json:"size,omitempty" binding:"omitempty,oneof=micro small medium large enterprise"`
	EmployeeCount *int    `json:"employee_count,omitempty" binding:"omitempty,min=0"`
	AnnualRevenue *int64  `json:"annual_revenue,omitempty" binding:"omitempty,min=0"`
	Currency      *string `json:"currency,omitempty" binding:"omitempty,len=3"`
}

// SetParentCompanyRequest is the request body for setting parent company
type SetParentCompanyRequest struct {
	ParentCompanyID *string `json:"parent_company_id"`
}

// UpdateCompanyDescriptionRequest is the request body for updating description
type UpdateCompanyDescriptionRequest struct {
	Description *string `json:"description,omitempty" binding:"omitempty,max=2000"`
}

// CompanyResponse is the response body for company queries
type CompanyResponse struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	LegalName          *string `json:"legal_name,omitempty"`
	Type               string  `json:"type"`
	TaxID              *string `json:"tax_id,omitempty"`
	RegistrationNumber *string `json:"registration_number,omitempty"`
	Website            *string `json:"website,omitempty"`
	Email              *string `json:"email,omitempty"`
	Phone              *string `json:"phone,omitempty"`
	PhoneCountryCode   *string `json:"phone_country_code,omitempty"`
	AddressLine1       *string `json:"address_line1,omitempty"`
	AddressLine2       *string `json:"address_line2,omitempty"`
	City               *string `json:"city,omitempty"`
	StateProvince      *string `json:"state_province,omitempty"`
	PostalCode         *string `json:"postal_code,omitempty"`
	Country            *string `json:"country,omitempty"`
	Industry           *string `json:"industry,omitempty"`
	Size               string  `json:"size"`
	EmployeeCount      int     `json:"employee_count"`
	AnnualRevenue      int64   `json:"annual_revenue"`
	Currency           string  `json:"currency"`
	Description        *string `json:"description,omitempty"`
	ParentCompanyID    *string `json:"parent_company_id,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
	DeletedAt          *string `json:"deleted_at,omitempty"`
}

// CompanyListResponse is the response body for listing companies
type CompanyListResponse struct {
	Companies []*CompanyResponse `json:"companies"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

// ToCompanyResponse converts company entity to response DTO
func ToCompanyResponse(c *company.Company) CompanyResponse {
	resp := CompanyResponse{
		ID:            c.ID.String(),
		Name:          c.Name,
		Type:          string(c.Type),
		Size:          string(c.Size),
		EmployeeCount: c.EmployeeCount,
		AnnualRevenue: c.Revenue, // entity.Revenue -> response.AnnualRevenue
		Currency:      c.Currency,
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Optional fields (pointers in entity)
	if c.LegalName != nil {
		resp.LegalName = c.LegalName
	}
	if c.TaxID != nil {
		resp.TaxID = c.TaxID
	}
	if c.RegistrationNumber != nil {
		resp.RegistrationNumber = c.RegistrationNumber
	}
	if c.Website != nil {
		resp.Website = c.Website
	}
	if c.Email != nil {
		email := c.Email.Value()
		resp.Email = &email
	}
	if c.Phone != nil {
		phone := c.Phone.Value()
		countryCode := c.Phone.CountryCode()
		resp.Phone = &phone
		resp.PhoneCountryCode = &countryCode
	}
	if c.Address != nil {
		if c.Address.Street != "" {
			resp.AddressLine1 = &c.Address.Street
		}
		if c.Address.Street2 != "" {
			resp.AddressLine2 = &c.Address.Street2
		}
		if c.Address.City != "" {
			resp.City = &c.Address.City
		}
		if c.Address.State != "" {
			resp.StateProvince = &c.Address.State
		}
		if c.Address.PostalCode != "" {
			resp.PostalCode = &c.Address.PostalCode
		}
		if c.Address.Country != "" {
			resp.Country = &c.Address.Country
		}
	}
	if c.Industry != nil {
		resp.Industry = c.Industry
	}
	if c.Description != nil {
		resp.Description = c.Description
	}
	if c.ParentCompanyID != nil {
		parentID := c.ParentCompanyID.String()
		resp.ParentCompanyID = &parentID
	}
	if c.DeletedAt != nil {
		deletedAt := c.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.DeletedAt = &deletedAt
	}

	return resp
}

// ToCompanyListResponse converts a list of companies to list response DTO
func ToCompanyListResponse(companies []*company.Company, total, page, pageSize int) CompanyListResponse {
	companyResponses := make([]*CompanyResponse, 0, len(companies))
	for _, c := range companies {
		resp := ToCompanyResponse(c)
		companyResponses = append(companyResponses, &resp)
	}

	return CompanyListResponse{
		Companies: companyResponses,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}
}

// Helper functions for parsing nullable fields

// parseEmail parses email from request
func parseEmail(email *string) (*valueobject.Email, error) {
	if email == nil || *email == "" {
		return nil, nil
	}
	e, err := valueobject.NewEmail(*email)
	if err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}
	return &e, nil
}

// parsePhone parses phone from request
func parsePhone(phone, countryCode *string) (*valueobject.Phone, error) {
	if phone == nil || *phone == "" {
		return nil, nil
	}
	// Combine country code with phone number for E.164 format
	number := *phone
	if countryCode != nil && *countryCode != "" {
		code := *countryCode
		if !strings.HasPrefix(code, "+") {
			code = "+" + code
		}
		// If phone doesn't start with +, prepend country code
		if !strings.HasPrefix(number, "+") {
			number = code + number
		}
	}
	// NewPhone expects E.164 format: +[country][number]
	p, err := valueobject.NewPhone(number)
	if err != nil {
		return nil, fmt.Errorf("invalid phone format: %w", err)
	}
	return &p, nil
}

// parseAddress parses address from request
func parseAddress(line1, line2, city, state, zip, country *string) (*valueobject.Address, error) {
	// If all required fields are empty, return nil
	if (line1 == nil || *line1 == "") &&
		(city == nil || *city == "") &&
		(zip == nil || *zip == "") &&
		(country == nil || *country == "") {
		return nil, nil
	}

	// Create address with required fields
	street := ""
	if line1 != nil {
		street = *line1
	}
	cityVal := ""
	if city != nil {
		cityVal = *city
	}
	postalCode := ""
	if zip != nil {
		postalCode = *zip
	}
	countryVal := ""
	if country != nil {
		countryVal = *country
	}

	// Create base address
	var addr valueobject.Address
	var err error

	if state != nil && *state != "" {
		addr, err = valueobject.NewAddressWithState(street, cityVal, *state, postalCode, countryVal)
	} else {
		addr, err = valueobject.NewAddress(street, cityVal, postalCode, countryVal)
	}
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Add optional second line
	if line2 != nil && *line2 != "" {
		addr = addr.WithStreet2(*line2)
	}

	return &addr, nil
}

// parseParentCompanyID parses parent company ID from request
func parseParentCompanyID(parentID *string) (*uuidv7.UUID, error) {
	if parentID == nil || *parentID == "" {
		return nil, nil
	}
	id, err := uuidv7.Parse(*parentID)
	if err != nil {
		return nil, fmt.Errorf("invalid parent company ID format: %w", err)
	}
	return &id, nil
}
