package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company/aggregate"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/company/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// CompanyRepository implements repository.ICompanyRepository using PostgreSQL
type CompanyRepository struct {
	*BaseRepository
}

// NewCompanyRepository creates a new PostgreSQL company repository
func NewCompanyRepository(db *sqlx.DB) repository.ICompanyRepository {
	return &CompanyRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// companyRow represents a database row for the customer_companies table
type companyRow struct {
	ID                 string         `db:"id"`
	Name               string         `db:"name"`
	LegalName          string         `db:"legal_name"`
	Type               string         `db:"type"`
	TaxID              sql.NullString `db:"tax_id"`
	RegistrationNumber sql.NullString `db:"registration_number"`
	Website            sql.NullString `db:"website"`
	Email              sql.NullString `db:"email"`
	Phone              sql.NullString `db:"phone"`
	PhoneCountryCode   sql.NullString `db:"phone_country_code"`
	AddressLine1       sql.NullString `db:"address_line1"`
	AddressLine2       sql.NullString `db:"address_line2"`
	City               sql.NullString `db:"city"`
	StateProvince      sql.NullString `db:"state_province"`
	PostalCode         sql.NullString `db:"postal_code"`
	Country            sql.NullString `db:"country"`
	Industry           sql.NullString `db:"industry"`
	Size               string         `db:"size"`
	EmployeeCount      int            `db:"employee_count"`
	Revenue            int64          `db:"revenue"`
	Currency           string         `db:"currency"`
	Description        sql.NullString `db:"description"`
	ParentCompanyID    sql.NullString `db:"parent_company_id"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	DeletedAt          sql.NullTime   `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *companyRow) toEntity() (*aggregate.Company, error) {
	// Parse Company ID
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid company ID: %w", err)
	}

	// Create Company aggregate
	legalName := r.LegalName
	c := &aggregate.Company{
		Name:          r.Name,
		LegalName:     &legalName,
		Type:          aggregate.CompanyType(r.Type),
		Size:          aggregate.CompanySize(r.Size),
		EmployeeCount: r.EmployeeCount,
		Revenue:       r.Revenue,
		Currency:      r.Currency,
	}
	
	// Set BaseAggregate fields
	c.ID = id
	c.CreatedAt = r.CreatedAt
	c.UpdatedAt = r.UpdatedAt

	// Parse optional TaxID
	if r.TaxID.Valid {
		taxID := r.TaxID.String
		c.TaxID = &taxID
	}

	// Parse optional RegistrationNumber
	if r.RegistrationNumber.Valid {
		regNum := r.RegistrationNumber.String
		c.RegistrationNumber = &regNum
	}

	// Parse optional Website
	if r.Website.Valid {
		website := r.Website.String
		c.Website = &website
	}

	// Parse optional Email
	if r.Email.Valid {
		email, err := valueobject.NewEmail(r.Email.String)
		if err != nil {
			return nil, fmt.Errorf("invalid email: %w", err)
		}
		c.Email = &email
	}

	// Parse optional Phone (combine country code with number)
	if r.Phone.Valid {
		phoneNumber := r.Phone.String
		if r.PhoneCountryCode.Valid {
			countryCode := r.PhoneCountryCode.String
			if !strings.HasPrefix(countryCode, "+") {
				countryCode = "+" + countryCode
			}
			if !strings.HasPrefix(phoneNumber, "+") {
				phoneNumber = countryCode + phoneNumber
			}
		}
		phone, err := valueobject.NewPhone(phoneNumber)
		if err != nil {
			return nil, fmt.Errorf("invalid phone: %w", err)
		}
		c.Phone = &phone
	}

	// Parse optional Address (use proper constructors)
	if r.AddressLine1.Valid && r.City.Valid && r.Country.Valid {
		var address valueobject.Address
		var err error
		if r.StateProvince.Valid {
			address, err = valueobject.NewAddressWithState(
				r.AddressLine1.String,
				r.City.String,
				r.StateProvince.String,
				stringPtrValue(r.PostalCode),
				r.Country.String,
			)
		} else {
			address, err = valueobject.NewAddress(
				r.AddressLine1.String,
				r.City.String,
				stringPtrValue(r.PostalCode),
				r.Country.String,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}
		if r.AddressLine2.Valid {
			address = address.WithStreet2(r.AddressLine2.String)
		}
		c.Address = &address
	}

	// Parse optional Industry
	if r.Industry.Valid {
		industry := r.Industry.String
		c.Industry = &industry
	}

	// Parse optional Description
	if r.Description.Valid {
		description := r.Description.String
		c.Description = &description
	}

	// Parse optional ParentCompanyID
	if r.ParentCompanyID.Valid {
		parentID, err := uuidv7.Parse(r.ParentCompanyID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid parent company ID: %w", err)
		}
		c.ParentCompanyID = &parentID
	}

	// Parse optional DeletedAt
	if r.DeletedAt.Valid {
		c.DeletedAt = &r.DeletedAt.Time
	}

	return c, nil
}

// toRow converts domain entity to database row
func toRow(c *aggregate.Company) (*companyRow, error) {
	var legalName string
	if c.LegalName != nil {
		legalName = *c.LegalName
	}
	
	row := &companyRow{
		ID:            c.GetID().String(),
		Name:          c.Name,
		LegalName:     legalName,
		Type:          string(c.Type),
		Size:          string(c.Size),
		EmployeeCount: c.EmployeeCount,
		Revenue: c.Revenue,
		Currency:      c.Currency,
		CreatedAt:     c.GetCreatedAt(),
		UpdatedAt:     c.GetUpdatedAt(),
	}

	// Optional TaxID
	if c.TaxID != nil && *c.TaxID != "" {
		row.TaxID = sql.NullString{String: *c.TaxID, Valid: true}
	}

	// Optional RegistrationNumber
	if c.RegistrationNumber != nil && *c.RegistrationNumber != "" {
		row.RegistrationNumber = sql.NullString{String: *c.RegistrationNumber, Valid: true}
	}

	// Optional Website
	if c.Website != nil && *c.Website != "" {
		row.Website = sql.NullString{String: *c.Website, Valid: true}
	}

	// Optional Email
	if c.Email != nil {
		row.Email = sql.NullString{String: c.Email.Value(), Valid: true}
	}

	// Optional Phone
	if c.Phone != nil {
		row.Phone = sql.NullString{String: c.Phone.Value(), Valid: true}
		row.PhoneCountryCode = sql.NullString{String: c.Phone.CountryCode(), Valid: true}
	}

	// Optional Address
	if c.Address != nil {
		row.AddressLine1 = stringToPtr(c.Address.Street)
		row.AddressLine2 = stringToPtr(c.Address.Street2)
		row.City = stringToPtr(c.Address.City)
		row.StateProvince = stringToPtr(c.Address.State)
		row.PostalCode = stringToPtr(c.Address.PostalCode)
		row.Country = stringToPtr(c.Address.Country)
	}

	// Optional Industry
	if c.Industry != nil && *c.Industry != "" {
		row.Industry = sql.NullString{String: *c.Industry, Valid: true}
	}

	// Optional Description
	if c.Description != nil && *c.Description != "" {
		row.Description = sql.NullString{String: *c.Description, Valid: true}
	}

	// Optional ParentCompanyID
	if c.ParentCompanyID != nil {
		row.ParentCompanyID = sql.NullString{String: c.ParentCompanyID.String(), Valid: true}
	}

	// Optional DeletedAt
	if c.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *c.DeletedAt, Valid: true}
	}

	return row, nil
}

// Helper functions for nullable string handling
func stringToPtr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func stringPtrValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// Create inserts a new company
func (r *CompanyRepository) Create(ctx context.Context, c *aggregate.Company) error {
	row, err := toRow(c)
	if err != nil {
		return fmt.Errorf("failed to convert to row: %w", err)
	}

	query := `
		INSERT INTO customer_companies (
			id, name, legal_name, type, tax_id, registration_number,
			website, email, phone, phone_country_code,
			address_line1, address_line2, city, state_province, postal_code, country,
			industry, size, employee_count, revenue, currency,
			description, parent_company_id,
			created_at, updated_at, deleted_at
		) VALUES (
			:id, :name, :legal_name, :type, :tax_id, :registration_number,
			:website, :email, :phone, :phone_country_code,
			:address_line1, :address_line2, :city, :state_province, :postal_code, :country,
			:industry, :size, :employee_count, :revenue, :currency,
			:description, :parent_company_id,
			:created_at, :updated_at, :deleted_at
		)`

	if _, err := r.NamedExec(ctx, query, row); err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}

	return nil
}

// GetByID retrieves a company by ID
func (r *CompanyRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Company, error) {
	var row companyRow
	query := `
		SELECT * FROM customer_companies
		WHERE id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, company.ErrCompanyNotFound
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return row.toEntity()
}

// GetByName retrieves a company by name
func (r *CompanyRepository) GetByName(ctx context.Context, name string) (*aggregate.Company, error) {
	var row companyRow
	query := `
		SELECT * FROM customer_companies
		WHERE name = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, company.ErrCompanyNotFound
		}
		return nil, fmt.Errorf("failed to get company by name: %w", err)
	}

	return row.toEntity()
}

// GetByTaxID retrieves a company by tax ID
func (r *CompanyRepository) GetByTaxID(ctx context.Context, taxID string) (*aggregate.Company, error) {
	var row companyRow
	query := `
		SELECT * FROM customer_companies
		WHERE tax_id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, taxID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, company.ErrCompanyNotFound
		}
		return nil, fmt.Errorf("failed to get company by tax ID: %w", err)
	}

	return row.toEntity()
}

// List retrieves paginated list of companies
func (r *CompanyRepository) List(ctx context.Context, page, pageSize int) ([]*aggregate.Company, int, error) {
	// Convert page/pageSize to limit/offset
	limit := pageSize
	offset := (page - 1) * pageSize

	var rows []companyRow
	query := `
		SELECT * FROM customer_companies
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	if err := r.Select(ctx, &rows, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list companies: %w", err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM customer_companies WHERE deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count companies: %w", err)
	}

	// Convert rows to entities
	companies := make([]*aggregate.Company, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		companies = append(companies, c)
	}

	return companies, total, nil
}

// ListByIndustry retrieves paginated list of companies by industry
func (r *CompanyRepository) ListByIndustry(ctx context.Context, industry string, page, pageSize int) ([]*aggregate.Company, int, error) {
	// Convert page/pageSize to limit/offset
	limit := pageSize
	offset := (page - 1) * pageSize

	var rows []companyRow
	query := `
		SELECT * FROM customer_companies
		WHERE industry = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, industry, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list companies by industry: %w", err)
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM customer_companies
		WHERE industry = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, industry); err != nil {
		return nil, 0, fmt.Errorf("failed to count companies by industry: %w", err)
	}

	// Convert rows to entities
	companies := make([]*aggregate.Company, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		companies = append(companies, c)
	}

	return companies, total, nil
}

// ListBySize retrieves paginated list of companies by size
func (r *CompanyRepository) ListBySize(ctx context.Context, size string, page, pageSize int) ([]*aggregate.Company, int, error) {
	// Convert page/pageSize to limit/offset
	limit := pageSize
	offset := (page - 1) * pageSize

	var rows []companyRow
	query := `
		SELECT * FROM customer_companies
		WHERE size = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, size, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list companies by size: %w", err)
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM customer_companies
		WHERE size = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, size); err != nil {
		return nil, 0, fmt.Errorf("failed to count companies by size: %w", err)
	}

	// Convert rows to entities
	companies := make([]*aggregate.Company, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		companies = append(companies, c)
	}

	return companies, total, nil
}

// ListSubsidiaries retrieves all subsidiaries of a parent company
func (r *CompanyRepository) ListSubsidiaries(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Company, error) {
	var rows []companyRow
	query := `
		SELECT * FROM customer_companies
		WHERE parent_company_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	if err := r.Select(ctx, &rows, query, parentID.String()); err != nil {
		return nil, fmt.Errorf("failed to list subsidiaries: %w", err)
	}

	// Convert rows to entities
	companies := make([]*aggregate.Company, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		companies = append(companies, c)
	}

	return companies, nil
}

// Update updates an existing company
func (r *CompanyRepository) Update(ctx context.Context, c *aggregate.Company) error {
	row, err := toRow(c)
	if err != nil {
		return fmt.Errorf("failed to convert to row: %w", err)
	}

	query := `
		UPDATE customer_companies SET
			name = :name,
			legal_name = :legal_name,
			type = :type,
			tax_id = :tax_id,
			registration_number = :registration_number,
			website = :website,
			email = :email,
			phone = :phone,
			phone_country_code = :phone_country_code,
			address_line1 = :address_line1,
			address_line2 = :address_line2,
			city = :city,
			state_province = :state_province,
			postal_code = :postal_code,
			country = :country,
			industry = :industry,
			size = :size,
			employee_count = :employee_count,
			revenue = :revenue,
			currency = :currency,
			description = :description,
			parent_company_id = :parent_company_id,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return company.ErrCompanyNotFound
	}

	return nil
}

// Delete soft-deletes a company
func (r *CompanyRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE customer_companies
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return company.ErrCompanyNotFound
	}

	return nil
}

// Exists checks if a company exists
func (r *CompanyRepository) Exists(ctx context.Context, id uuidv7.UUID) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM customer_companies
			WHERE id = $1 AND deleted_at IS NULL
		)`

	if err := r.Get(ctx, &exists, query, id.String()); err != nil {
		return false, fmt.Errorf("failed to check company existence: %w", err)
	}

	return exists, nil
}

// ExistsByName checks if a company with given name exists
func (r *CompanyRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM customer_companies
			WHERE name = $1 AND deleted_at IS NULL
		)`

	if err := r.Get(ctx, &exists, query, name); err != nil {
		return false, fmt.Errorf("failed to check company name existence: %w", err)
	}

	return exists, nil
}
