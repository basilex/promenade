package usecase

import (
	"context"

	companyerrors "github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company/aggregate"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

var (
	// companyerrors.ErrCompanyNotFound moved to errors.go
	// companyerrors.ErrCompanyAlreadyExists moved to errors.go
)

// ICompanyUseCase defines the interface for company business logic
type ICompanyUseCase interface {
	CreateCompany(ctx context.Context, name string, legalName *string, companyType string, taxID *string, registrationNumber *string, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address, industry *string, size string, employeeCount int, revenue int64, currency string, description *string, parentCompanyID *uuidv7.UUID) (*aggregate.Company, error)
	GetCompany(ctx context.Context, id uuidv7.UUID) (*aggregate.Company, error)
	GetCompanyByName(ctx context.Context, name string) (*aggregate.Company, error)
	GetCompanyByTaxID(ctx context.Context, taxID string) (*aggregate.Company, error)
	ListCompanies(ctx context.Context, page, pageSize int) ([]*aggregate.Company, int, error)
	ListCompaniesByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*aggregate.Company, int, error)
	ListCompaniesBySize(ctx context.Context, size string, page, pageSize int) ([]*aggregate.Company, int, error)
	ListSubsidiaries(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*aggregate.Company, error)
	UpdateCompanyBasicInfo(ctx context.Context, id uuidv7.UUID, name string, legalName *string, companyType string, taxID *string, registrationNumber *string) (*aggregate.Company, error)
	UpdateCompanyContactInfo(ctx context.Context, id uuidv7.UUID, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address) (*aggregate.Company, error)
	UpdateCompanyBusinessInfo(ctx context.Context, id uuidv7.UUID, industry *string, size string, employeeCount int, revenue int64, currency string) (*aggregate.Company, error)
	SetParentCompany(ctx context.Context, id uuidv7.UUID, parentCompanyID *uuidv7.UUID) (*aggregate.Company, error)
	UpdateCompanyDescription(ctx context.Context, id uuidv7.UUID, description *string) (*aggregate.Company, error)
	DeleteCompany(ctx context.Context, id uuidv7.UUID) error
}

type CompanyUseCase struct {
	repo repository.ICompanyRepository
}

// aggregate.NewCompanyUseCase creates a new company use case
func NewCompanyUseCase(repo repository.ICompanyRepository) ICompanyUseCase {
	return &CompanyUseCase{repo: repo}
}

func (uc *CompanyUseCase) CreateCompany(ctx context.Context, name string, legalName *string, companyType string, taxID *string, registrationNumber *string, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address, industry *string, size string, employeeCount int, revenue int64, currency string, description *string, parentCompanyID *uuidv7.UUID) (*aggregate.Company, error) {
	// Check if company with same name already exists
	exists, err := uc.repo.ExistsByName(ctx, name)
	if err != nil {
		return nil, companyerrors.ErrCompanyExistenceCheckFailed
	}
	if exists {
		return nil, companyerrors.ErrCompanyAlreadyExists
	}

	// Create company
	company, err := aggregate.NewCompany(name, companyType)
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
			return nil, companyerrors.ErrCompanyParentCheckFailed
		}
		if !parentExists {
			return nil, companyerrors.ErrParentCompanyNotFound
		}
	}

	// Save company
	if err := uc.repo.Create(ctx, company); err != nil {
		return nil, companyerrors.ErrCompanyCreateFailed
	}

	return company, nil
}

func (uc *CompanyUseCase) GetCompany(ctx context.Context, id uuidv7.UUID) (*aggregate.Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}
	return company, nil
}

func (uc *CompanyUseCase) GetCompanyByName(ctx context.Context, name string) (*aggregate.Company, error) {
	company, err := uc.repo.GetByName(ctx, name)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}
	return company, nil
}

func (uc *CompanyUseCase) GetCompanyByTaxID(ctx context.Context, taxID string) (*aggregate.Company, error) {
	company, err := uc.repo.GetByTaxID(ctx, taxID)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}
	return company, nil
}

func (uc *CompanyUseCase) ListCompanies(ctx context.Context, page, pageSize int) ([]*aggregate.Company, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.List(ctx, page, pageSize)
}

func (uc *CompanyUseCase) ListCompaniesByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*aggregate.Company, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.ListByIndustry(ctx, industry, page, pageSize)
}

func (uc *CompanyUseCase) ListCompaniesBySize(ctx context.Context, size string, page, pageSize int) ([]*aggregate.Company, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return uc.repo.ListBySize(ctx, size, page, pageSize)
}

func (uc *CompanyUseCase) ListSubsidiaries(ctx context.Context, parentCompanyID uuidv7.UUID) ([]*aggregate.Company, error) {
	return uc.repo.ListSubsidiaries(ctx, parentCompanyID)
}

func (uc *CompanyUseCase) UpdateCompanyBasicInfo(ctx context.Context, id uuidv7.UUID, name string, legalName *string, companyType string, taxID *string, registrationNumber *string) (*aggregate.Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}

	if err := company.UpdateBasicInfo(name, legalName, companyType, taxID, registrationNumber); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, companyerrors.ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *CompanyUseCase) UpdateCompanyContactInfo(ctx context.Context, id uuidv7.UUID, website *string, email *valueobject.Email, phone *valueobject.Phone, address *valueobject.Address) (*aggregate.Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}

	company.UpdateContactInfo(website, email, phone, address)

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, companyerrors.ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *CompanyUseCase) UpdateCompanyBusinessInfo(ctx context.Context, id uuidv7.UUID, industry *string, size string, employeeCount int, revenue int64, currency string) (*aggregate.Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}

	if err := company.UpdateBusinessInfo(industry, size, employeeCount, revenue, currency); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, companyerrors.ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *CompanyUseCase) SetParentCompany(ctx context.Context, id uuidv7.UUID, parentCompanyID *uuidv7.UUID) (*aggregate.Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}

	// Validate parent company exists if provided
	if parentCompanyID != nil {
		parentCompany, err := uc.repo.GetByID(ctx, *parentCompanyID)
		if err != nil {
			return nil, companyerrors.ErrParentCompanyNotFound
		}
		if parentCompany.IsDeleted() {
			return nil, companyerrors.ErrParentCompanyDeleted
		}
	}

	if err := company.SetParentCompany(parentCompanyID); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, companyerrors.ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *CompanyUseCase) UpdateCompanyDescription(ctx context.Context, id uuidv7.UUID, description *string) (*aggregate.Company, error) {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, companyerrors.ErrCompanyNotFound
	}

	company.UpdateDescription(description)

	if err := uc.repo.Update(ctx, company); err != nil {
		return nil, companyerrors.ErrCompanyUpdateFailed
	}

	return company, nil
}

func (uc *CompanyUseCase) DeleteCompany(ctx context.Context, id uuidv7.UUID) error {
	company, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return companyerrors.ErrCompanyNotFound
	}

	return uc.repo.Delete(ctx, company.ID)
}
