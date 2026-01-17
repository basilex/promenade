package http

import (
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company/aggregate"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company/dto"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// parseEmail converts optional string to Email value object
func parseEmail(emailStr *string) (*valueobject.Email, error) {
	if emailStr == nil || *emailStr == "" {
		return nil, nil
	}
	email, err := valueobject.NewEmail(*emailStr)
	if err != nil {
		return nil, err
	}
	return &email, nil
}

// parsePhone converts optional phone string to Phone value object (E.164 format expected)
func parsePhone(phoneStr *string) (*valueobject.Phone, error) {
	if phoneStr == nil || *phoneStr == "" {
		return nil, nil
	}
	phone, err := valueobject.NewPhone(*phoneStr)
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

// parseAddress converts optional address fields to Address value object
func parseAddress(line1, line2, city, state, postal, country *string) (*valueobject.Address, error) {
	if line1 == nil || *line1 == "" {
		return nil, nil
	}
	if city == nil || *city == "" {
		return nil, nil
	}
	if postal == nil || *postal == "" {
		return nil, nil
	}
	if country == nil || *country == "" {
		return nil, nil
	}

	addr, err := valueobject.NewAddress(*line1, *city, *postal, *country)
	if err != nil {
		return nil, err
	}
	
	if line2 != nil && *line2 != "" {
		addr.Street2 = *line2
	}
	if state != nil && *state != "" {
		addr.State = *state
	}
	
	return &addr, nil
}

// parseParentCompanyID converts optional string to UUID
func parseParentCompanyID(idStr *string) (*uuidv7.UUID, error) {
	if idStr == nil || *idStr == "" {
		return nil, nil
	}
	id, err := uuidv7.Parse(*idStr)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// ptrStr safely dereferences string pointer
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// toCompanyResponse wraps dto conversion functions
func toCompanyResponse(comp *aggregate.Company) dto.CompanyResponse {
	return dto.ToCompanyResponse(comp)
}

// toCompanyListResponse wraps list conversion
func toCompanyListResponse(companies []*aggregate.Company, total, page, pageSize int) dto.CompanyListResponse {
	return dto.ToCompanyListResponse(companies, total, page, pageSize)
}
