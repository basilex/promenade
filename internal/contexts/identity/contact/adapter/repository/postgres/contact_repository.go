package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	contacterrors "github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/internal/contexts/identity/contact/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/contact/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// contactRepository implements repository.IContactRepository for PostgreSQL
type contactRepository struct {
	db *sqlx.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *sqlx.DB) repository.IContactRepository {
	return &contactRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *contactRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *contactRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *contactRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning rows
func (r *contactRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows
func (r *contactRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

// contactRow represents database row structure for identity_contacts table
type contactRow struct {
	ID                uuidv7.UUID    `db:"id"`
	UserID            uuidv7.UUID    `db:"user_id"`
	ContactType       string         `db:"contact_type"`
	Label             string         `db:"label"`
	Email             sql.NullString `db:"email"`
	Phone             sql.NullString `db:"phone"`
	AddressStreet     sql.NullString `db:"address_street"`
	AddressStreet2    sql.NullString `db:"address_street2"`
	AddressCity       sql.NullString `db:"address_city"`
	AddressState      sql.NullString `db:"address_state"`
	AddressPostalCode sql.NullString `db:"address_postal_code"`
	AddressCountry    sql.NullString `db:"address_country"`
	IsPrimary         bool           `db:"is_primary"`
	IsVerified        bool           `db:"is_verified"`
	IsPublic          bool           `db:"is_public"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
}

// toEntity converts database row to domain entity
func (r *contactRow) toEntity() (*aggregate.Contact, error) {
	c := &aggregate.Contact{
		UserID:     r.UserID,
		Type:       aggregate.ContactType(r.ContactType),
		Label:      r.Label,
		IsPrimary:  r.IsPrimary,
		IsVerified: r.IsVerified,
		IsPublic:   r.IsPublic,
	}
	// Initialize BaseAggregate fields
	c.ID = r.ID
	c.CreatedAt = r.CreatedAt
	c.UpdatedAt = r.UpdatedAt

	// Populate value object based on contact type
	switch c.Type {
	case aggregate.ContactTypeEmail:
		if r.Email.Valid {
			email, err := valueobject.NewEmail(r.Email.String)
			if err != nil {
				return nil, fmt.Errorf("invalid email value: %w", err)
			}
			c.Email = &email
		}
	case aggregate.ContactTypePhone:
		if r.Phone.Valid {
			phone, err := valueobject.NewPhone(r.Phone.String)
			if err != nil {
				return nil, fmt.Errorf("invalid phone value: %w", err)
			}
			c.Phone = &phone
		}
	case aggregate.ContactTypeAddress:
		if r.AddressStreet.Valid && r.AddressCity.Valid && r.AddressCountry.Valid && r.AddressPostalCode.Valid {
			address, err := valueobject.NewAddress(
				r.AddressStreet.String,
				r.AddressCity.String,
				r.AddressPostalCode.String,
				r.AddressCountry.String,
			)
			if err != nil {
				return nil, fmt.Errorf("invalid address value: %w", err)
			}

			// Set optional fields
			if r.AddressStreet2.Valid {
				address.Street2 = r.AddressStreet2.String
			}
			if r.AddressState.Valid {
				address.State = r.AddressState.String
			}

			c.Address = &address
		}
	}

	return c, nil
}

// fromEntity converts domain entity to database row
func fromEntity(c *aggregate.Contact) *contactRow {
	row := &contactRow{
		ID:          c.GetID(),
		UserID:      c.UserID,
		ContactType: string(c.Type),
		Label:       c.Label,
		IsPrimary:   c.IsPrimary,
		IsVerified:  c.IsVerified,
		IsPublic:    c.IsPublic,
	}

	// Populate database fields based on contact type
	switch c.Type {
	case aggregate.ContactTypeEmail:
		if c.Email != nil {
			row.Email = sql.NullString{String: c.Email.Value(), Valid: true}
		}
	case aggregate.ContactTypePhone:
		if c.Phone != nil {
			row.Phone = sql.NullString{String: c.Phone.Value(), Valid: true}
		}
	case aggregate.ContactTypeAddress:
		if c.Address != nil {
			row.AddressStreet = sql.NullString{String: c.Address.Street, Valid: true}
			row.AddressCity = sql.NullString{String: c.Address.City, Valid: true}
			row.AddressPostalCode = sql.NullString{String: c.Address.PostalCode, Valid: true}
			row.AddressCountry = sql.NullString{String: c.Address.Country, Valid: true}
			if c.Address.Street2 != "" {
				row.AddressStreet2 = sql.NullString{String: c.Address.Street2, Valid: true}
			}
			if c.Address.State != "" {
				row.AddressState = sql.NullString{String: c.Address.State, Valid: true}
			}
		}
	}

	return row
}

// Create creates a new contact in the database
func (r *contactRepository) Create(ctx context.Context, c *aggregate.Contact) error {
	row := fromEntity(c)

	query := `
		INSERT INTO identity_contacts (
			id, user_id, contact_type, label,
			email, phone,
			address_street, address_street2, address_city, address_state, address_postal_code, address_country,
			is_primary, is_verified, is_public
		) VALUES (
			:id, :user_id, :contact_type, :label,
			:email, :phone,
			:address_street, :address_street2, :address_city, :address_state, :address_postal_code, :address_country,
			:is_primary, :is_verified, :is_public
		)
	`

	_, err := r.NamedExec(ctx, query, row)
	return err
}

// GetByID retrieves a contact by ID
func (r *contactRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Contact, error) {
	var row contactRow
	query := `
		SELECT id, user_id, contact_type, label,
			   email, phone,
			   address_street, address_street2, address_city, address_state, address_postal_code, address_country,
			   is_primary, is_verified, is_public,
			   created_at, updated_at
		FROM identity_contacts
		WHERE id = $1
	`

	if err := r.Get(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, contacterrors.ErrContactNotFound
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return row.toEntity()
}

// GetByUserID retrieves all contacts for a user
func (r *contactRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*aggregate.Contact, error) {
	var rows []contactRow
	query := `
		SELECT id, user_id, contact_type, label,
			   email, phone,
			   address_street, address_street2, address_city, address_state, address_postal_code, address_country,
			   is_primary, is_verified, is_public,
			   created_at, updated_at
		FROM identity_contacts
		WHERE user_id = $1
		ORDER BY is_primary DESC, created_at DESC
	`

	if err := r.Select(ctx, &rows, query, userID); err != nil {
		return nil, fmt.Errorf("failed to get user contacts: %w", err)
	}

	contacts := make([]*aggregate.Contact, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// GetByUserIDAndType retrieves contacts for a user filtered by type
func (r *contactRepository) GetByUserIDAndType(ctx context.Context, userID uuidv7.UUID, contactType aggregate.ContactType) ([]*aggregate.Contact, error) {
	var rows []contactRow
	query := `
		SELECT id, user_id, contact_type, label,
			   email, phone,
			   address_street, address_street2, address_city, address_state, address_postal_code, address_country,
			   is_primary, is_verified, is_public,
			   created_at, updated_at
		FROM identity_contacts
		WHERE user_id = $1 AND contact_type = $2
		ORDER BY is_primary DESC, created_at DESC
	`

	if err := r.Select(ctx, &rows, query, userID, string(contactType)); err != nil {
		return nil, fmt.Errorf("failed to get contacts by type: %w", err)
	}

	contacts := make([]*aggregate.Contact, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}

	return contacts, nil
}

// GetPrimaryByUserIDAndType retrieves the primary contact for a user and type
func (r *contactRepository) GetPrimaryByUserIDAndType(ctx context.Context, userID uuidv7.UUID, contactType aggregate.ContactType) (*aggregate.Contact, error) {
	var row contactRow
	query := `
		SELECT id, user_id, contact_type, label,
			   email, phone,
			   address_street, address_street2, address_city, address_state, address_postal_code, address_country,
			   is_primary, is_verified, is_public,
			   created_at, updated_at
		FROM identity_contacts
		WHERE user_id = $1 AND contact_type = $2 AND is_primary = true
	`

	if err := r.Get(ctx, &row, query, userID, string(contactType)); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No primary contact found (not an error)
		}
		return nil, fmt.Errorf("failed to get primary contact: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing contact
func (r *contactRepository) Update(ctx context.Context, c *aggregate.Contact) error {
	row := fromEntity(c)

	query := `
		UPDATE identity_contacts
		SET contact_type = :contact_type,
		    label = :label,
		    email = :email,
		    phone = :phone,
		    address_street = :address_street,
		    address_street2 = :address_street2,
		    address_city = :address_city,
		    address_state = :address_state,
		    address_postal_code = :address_postal_code,
		    address_country = :address_country,
		    is_primary = :is_primary,
		    is_verified = :is_verified,
		    is_public = :is_public,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`

	_, err := r.NamedExec(ctx, query, row)
	return err
}

// Delete deletes a contact by ID
func (r *contactRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM identity_contacts WHERE id = $1`
	_, err := r.Exec(ctx, query, id)
	return err
}

// SetPrimary sets a contact as primary and unsets other primary contacts of the same type
// This operation is atomic within a transaction
func (r *contactRepository) SetPrimary(ctx context.Context, id uuidv7.UUID) error {
	// First, get the contact to know its user_id and type
	var info struct {
		UserID      uuidv7.UUID `db:"user_id"`
		ContactType string      `db:"contact_type"`
	}
	getQuery := `SELECT user_id, contact_type FROM identity_contacts WHERE id = $1`
	if err := r.Get(ctx, &info, getQuery, id); err != nil {
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Unset all primary flags for this user and contact type
	unsetQuery := `
		UPDATE identity_contacts
		SET is_primary = false
		WHERE user_id = $1 AND contact_type = $2
	`
	if _, err := r.Exec(ctx, unsetQuery, info.UserID, info.ContactType); err != nil {
		return fmt.Errorf("failed to unset primary flags: %w", err)
	}

	// Set the new primary contact
	setPrimaryQuery := `
		UPDATE identity_contacts
		SET is_primary = true
		WHERE id = $1
	`
	_, err := r.Exec(ctx, setPrimaryQuery, id)
	return err
}

// ExistsPrimaryForUserAndType checks if a primary contact exists for a user and type
func (r *contactRepository) ExistsPrimaryForUserAndType(ctx context.Context, userID uuidv7.UUID, contactType aggregate.ContactType) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM identity_contacts
		WHERE user_id = $1 AND contact_type = $2 AND is_primary = true
	`

	if err := r.Get(ctx, &count, query, userID, string(contactType)); err != nil {
		return false, fmt.Errorf("failed to check primary contact existence: %w", err)
	}

	return count > 0, nil
}
