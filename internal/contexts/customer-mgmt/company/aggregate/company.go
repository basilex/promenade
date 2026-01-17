package aggregate

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// CompanyType represents the legal structure of a company
type CompanyType string

const (
	CompanyTypeLLC            CompanyType = "llc"
	CompanyTypeCorporation    CompanyType = "corporation"
	CompanyTypeSoleProprietor CompanyType = "sole_proprietor"
	CompanyTypePartnership    CompanyType = "partnership"
	CompanyTypeNonProfit      CompanyType = "non_profit"
	CompanyTypeOther          CompanyType = "other"
)

// CompanySize represents the size category of a company
type CompanySize string

const (
	CompanySizeMicro      CompanySize = "micro"      // 1-10 employees
	CompanySizeSmall      CompanySize = "small"      // 11-50 employees
	CompanySizeMedium     CompanySize = "medium"     // 51-250 employees
	CompanySizeLarge      CompanySize = "large"      // 251-1000 employees
	CompanySizeEnterprise CompanySize = "enterprise" // 1000+ employees
)

// Company represents a B2B customer organization
type Company struct {
	aggregate.BaseAggregate

	Name               string
	LegalName          *string
	Type               CompanyType
	TaxID              *string
	RegistrationNumber *string

	// Contact Information
	Website *string
	Email   *valueobject.Email
	Phone   *valueobject.Phone
	Address *valueobject.Address

	// Business Information
	Industry      *string
	Size          CompanySize
	EmployeeCount int
	Revenue       int64  // in cents
	Currency      string // ISO 4217

	Description *string

	// Relationships
	ParentCompanyID *uuidv7.UUID

	// Timestamps (CreatedAt, UpdatedAt, DeletedAt from BaseAggregate)
}

// NewCompany creates a new company with required fields
func NewCompany(name string, companyType string) (*Company, error) {
	if name == "" {
		return nil, company.ErrCompanyNameRequired
	}

	cType := CompanyType(companyType)
	if !isValidCompanyType(cType) {
		return nil, company.ErrCompanyTypeInvalid
	}

	return &Company{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Name:          name,
		Type:          cType,
		Size:          CompanySizeMicro,
		EmployeeCount: 0,
		Revenue:       0,
		Currency:      "USD",
	}, nil
}

// UpdateBasicInfo updates company basic information
func (c *Company) UpdateBasicInfo(name string, legalName *string, companyType string, taxID *string, registrationNumber *string) error {
	if name == "" {
		return company.ErrCompanyNameRequired
	}

	cType := CompanyType(companyType)
	if !isValidCompanyType(cType) {
		return company.ErrCompanyTypeInvalid
	}

	c.Name = name
	c.LegalName = legalName
	c.Type = cType
	c.TaxID = taxID
	c.RegistrationNumber = registrationNumber
	c.Touch()

	return nil
}

// UpdateContactInfo updates company contact information
func (c *Company) UpdateContactInfo(website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address) {
	c.Website = website
	c.Email = email
	c.Phone = phone
	c.Address = address
	c.Touch()
}

// UpdateBusinessInfo updates company business information
func (c *Company) UpdateBusinessInfo(industry *string, size string, employeeCount int, revenue int64, currency string) error {
	cSize := CompanySize(size)
	if !isValidCompanySize(cSize) {
		return company.ErrCompanySizeInvalid
	}

	if employeeCount < 0 {
		return company.ErrEmployeeCountNegative
	}

	if revenue < 0 {
		return company.ErrRevenueNegative
	}

	c.Industry = industry
	c.Size = cSize
	c.EmployeeCount = employeeCount
	c.Revenue = revenue
	c.Currency = currency
	c.Touch()

	return nil
}

// SetParentCompany sets the parent company (for subsidiaries)
func (c *Company) SetParentCompany(parentCompanyID *uuidv7.UUID) error {
	if parentCompanyID != nil && *parentCompanyID == c.ID {
		return company.ErrCompanyCannotBeOwnParent
	}

	c.ParentCompanyID = parentCompanyID
	c.Touch()

	return nil
}

// UpdateDescription updates company description
func (c *Company) UpdateDescription(description *string) {
	c.Description = description
	c.Touch()
}

// Delete marks company as deleted (soft delete)
func (c *Company) Delete() {
	now := time.Now()
	c.DeletedAt = &now
	c.Touch()
}

// IsDeleted checks if company is deleted
func (c *Company) IsDeleted() bool {
	return c.DeletedAt != nil
}

// Validate validates company data
func (c *Company) Validate() error {
	if c.Name == "" {
		return company.ErrCompanyNameRequired
	}

	if !isValidCompanyType(c.Type) {
		return company.ErrCompanyTypeInvalid
	}

	if !isValidCompanySize(c.Size) {
		return company.ErrCompanySizeInvalid
	}

	if c.EmployeeCount < 0 {
		return company.ErrEmployeeCountNegative
	}

	if c.Revenue < 0 {
		return company.ErrRevenueNegative
	}

	if c.ParentCompanyID != nil && *c.ParentCompanyID == c.ID {
		return company.ErrCompanyCannotBeOwnParent
	}

	return nil
}

// Helper functions

func isValidCompanyType(t CompanyType) bool {
	switch t {
	case CompanyTypeLLC, CompanyTypeCorporation, CompanyTypeSoleProprietor,
		CompanyTypePartnership, CompanyTypeNonProfit, CompanyTypeOther:
		return true
	default:
		return false
	}
}

func isValidCompanySize(s CompanySize) bool {
	switch s {
	case CompanySizeMicro, CompanySizeSmall, CompanySizeMedium,
		CompanySizeLarge, CompanySizeEnterprise:
		return true
	default:
		return false
	}
}
