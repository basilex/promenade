package company

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

var (
	// ErrCompanyNotFound moved to errors.go
	// ErrCompanyAlreadyExists moved to errors.go
)

// IUseCase defines the interface for company business logic
type IUseCase interface {
	CreateCompany(ctx context.Context, name string, legalName *string, companyType string, taxID *string, registrationNumber *string, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address, industry *string, size string, employeeCount int, revenue int64, currency string, description *string, parentCompanyID *uuidv7.UUID) (*Company, error)
	GetCompany(ctx context.Context, id uuidv7.UUID) (*Company, error)
	GetCompanyByName(ctx context.Context, name string) (*Company, error)
	GetCompanyByTaxID(ctx context.Context, taxID string) (*Company, error)
	ListCompanies(ctx context.Context, page, pageSize int) ([]*Company, int, error)
	ListCompaniesByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*Company, int, error)
	ListCompaniesBySize(ctx context.Context, size string, page, pageSize int) ([]*Company, int, error)
	ListSubsidiaries(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*Company, error)
	UpdateCompanyBasicInfo(ctx context.Context, id uuidv7.UUID, name string, legalName *string, companyType string, taxID *string, registrationNumber *string) (*Company, error)
	UpdateCompanyContactInfo(ctx context.Context, id uuidv7.UUID, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address) (*Company, error)
	UpdateCompanyBusinessInfo(ctx context.Context, id uuidv7.UUID, industry *string, size string, employeeCount int, revenue int64, currency string) (*Company, error)
	SetParentCompany(ctx context.Context, id uuidv7.UUID, parentCompanyID *uuidv7.UUID) (*Company, error)
	UpdateCompanyDescription(ctx context.Context, id uuidv7.UUID, description *string) (*Company, error)
	DeleteCompany(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
	repo IRepository
}

// NewUseCase creates a new company use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) CreateCompany(ctx context.Context, name string, legalName *string, companyType string, taxID *string, registrationNumber *string, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address, industry *string, size string, employeeCount int, revenue int64, currency string, description *string, parentCompanyID *uuidv7.UUID) (*Company, error) {
	// Check if company with same name already exists
	exists, err := uc.repo.ExistsByName(ctx, name)
	if err != nil {
		return nil, ErrCompanyExistenceCheckFailed
	}
	if exists {
		return nil, ErrCompanyAlreadyExists
	}

	// Create company
	company, err := NewCompany(name, companyType)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	company.LegalName = legalName
	company.TaxID = taxID
	company.RegistrationNumber = registrationNumber
	company.Website = website
	company.Email = email
	company.Phone = phone
	company.Address = address
	company.Industry = industry
	company.Description = description
	company.ParentCompanyID = parentCompanyID

	// Set business info
	if err := company.UpdateBusinessInfo(industry, size, employeeCount, revenue, currency); err != nil {
		return nil, err
	}

	// Validate parent company exists if provided
	if parentCompanyID != nil {
		parentExists, err := uc.repo.Exists(ctx, *parentCompanyID)
		if err != nil {
			return nil, ErrCompanyParentCheckFailed
		}
		if !parentExists {
			return nil, ErrParentCompanyNotFound
		}
	}

	// Save company
	if err := uc.repo.Create(ctx, company); err != nil {
		return nil, ErrCompanyCreateFailed
	}

	return company, nil
}

func (uc *useCase) GetCompany(ctx context.Context, id uuidv7.UUID) (*Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCompanyNotFound
	}
	return company, nil
}

func (uc *useCase) GetCompanyByName(ctx context.Context, name string) (*Company, error) {
	company, err := uc.repo.GetByName(ctx, name)
	if err != nil {
		return nil, ErrCompanyNotFound
	}
	return company, nil
}

func (uc *useCase) GetCompanyByTaxID(ctx context.Context, taxID string) (*Company, error) {
	company, err := uc.repo.GetByTaxID(ctx, taxID)
	if err != nil {
		return nil, ErrCompanyNotFound
	}
	return company, nil
}

func (uc *useCase) ListCompanies(ctx context.Context, page, pageSize int) ([]*Company, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.List(ctx, page, pageSize)
}

func (uc *useCase) ListCompaniesByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*Company, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.ListByIndustry(ctx, industry, page, pageSize)
}

func (uc *useCase) ListCompaniesBySize(ctx context.Context, size string, page, pageSize int) ([]*Company, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.ListBySize(ctx, size, page, pageSize)
}

func (uc *useCase) ListSubsidiaries(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*Company, error) {
	return uc.repo.ListSubsidiaries(ctx, parentCompanyID)
}

func (uc *useCase) UpdateCompanyBasicInfo(ctx context.Context, id uuidv7.UUID, name string, legalName *string, companyType string, taxID *string, registrationNumber *string) (*Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCompanyNotFound
	}

	if err := company.UpdateBasicInfo(name, legalName, companyType, taxID, registrationNumber); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *useCase) UpdateCompanyContactInfo(ctx context.Context, id uuidv7.UUID, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address) (*Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCompanyNotFound
	}

	company.UpdateContactInfo(website, email, phone, address)

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *useCase) UpdateCompanyBusinessInfo(ctx context.Context, id uuidv7.UUID, industry *string, size string, employeeCount int, revenue int64, currency string) (*Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCompanyNotFound
	}

	if err := company.UpdateBusinessInfo(industry, size, employeeCount, revenue, currency); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *useCase) SetParentCompany(ctx context.Context, id uuidv7.UUID, parentCompanyID *uuidv7.UUID) (*Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCompanyNotFound
	}

	// Validate parent company exists if provided
	if parentCompanyID != nil {
		parentCompany, err := uc.repo.GetByID(ctx, *parentCompanyID)
		if err != nil {
			return nil, ErrParentCompanyNotFound
		}
		if parentCompany.IsDeleted() {
			return nil, ErrParentCompanyDeleted
		}
	}

	if err := company.SetParentCompany(parentCompanyID); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *useCase) UpdateCompanyDescription(ctx context.Context, id uuidv7.UUID, description *string) (*Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrCompanyNotFound
	}

	company.UpdateDescription(description)

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *useCase) DeleteCompany(ctx context.Context, id uuidv7.UUID) error {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrCompanyNotFound
	}

	return uc.repo.Delete(ctx, company.ID)
}
