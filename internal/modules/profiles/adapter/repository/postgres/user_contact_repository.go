package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserContactRepository implements repository.UserContactRepository
type UserContactRepository struct {
	*BaseRepository
}

// NewUserContactRepository creates a new UserContactRepository
func NewUserContactRepository(db *sqlx.DB) *UserContactRepository {
	return &UserContactRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new user contact
func (r *UserContactRepository) Create(ctx context.Context, contact *entity.UserContact) error {
	contact.ID = uuidv7.New()

	query := `
		INSERT INTO user_contacts (
			id, user_id, contact_type, contact_value, label,
			is_verified, is_primary, is_active, is_public,
			available_from, available_to, available_days, timezone, notes
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14
		)`

	err := r.Exec(ctx, query,
		contact.ID,
		contact.UserID,
		contact.ContactType,
		contact.ContactValue,
		contact.Label,
		contact.IsVerified,
		contact.IsPrimary,
		contact.IsActive,
		contact.IsPublic,
		contact.AvailableFrom,
		contact.AvailableTo,
		contact.AvailableDays,
		contact.Timezone,
		contact.Notes,
	)

	return err
}

// GetByID retrieves a contact by ID
func (r *UserContactRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserContact, error) {
	var contact entity.UserContact
	query := `
		SELECT id, user_id, contact_type, contact_value, label,
			is_verified, is_primary, is_active, is_public,
			available_from, available_to, available_days, timezone, notes,
			created_at, updated_at
		FROM user_contacts
		WHERE id = $1`

	err := r.Get(ctx, &contact, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &contact, nil
}

// GetUserContacts retrieves all contacts for a user
func (r *UserContactRepository) GetUserContacts(ctx context.Context, userID uuidv7.UUID, includeInactive bool) ([]*entity.UserContact, error) {
	var contacts []*entity.UserContact
	query := `
		SELECT id, user_id, contact_type, contact_value, label,
			is_verified, is_primary, is_active, is_public,
			available_from, available_to, available_days, timezone, notes,
			created_at, updated_at
		FROM user_contacts
		WHERE user_id = $1`

	if !includeInactive {
		query += " AND is_active = true"
	}

	query += " ORDER BY is_primary DESC, contact_type, created_at DESC"

	err := r.Select(ctx, &contacts, query, userID)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// GetUserContactsByType retrieves user contacts filtered by type
func (r *UserContactRepository) GetUserContactsByType(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) ([]*entity.UserContact, error) {
	var contacts []*entity.UserContact
	query := `
		SELECT id, user_id, contact_type, contact_value, label,
			is_verified, is_primary, is_active, is_public,
			available_from, available_to, available_days, timezone, notes,
			created_at, updated_at
		FROM user_contacts
		WHERE user_id = $1 AND contact_type = $2 AND is_active = true
		ORDER BY is_primary DESC, created_at DESC`

	err := r.Select(ctx, &contacts, query, userID, contactType)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// GetPrimaryContact retrieves the primary contact of a specific type for a user
func (r *UserContactRepository) GetPrimaryContact(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType) (*entity.UserContact, error) {
	var contact entity.UserContact
	query := `
		SELECT id, user_id, contact_type, contact_value, label,
			is_verified, is_primary, is_active, is_public,
			available_from, available_to, available_days, timezone, notes,
			created_at, updated_at
		FROM user_contacts
		WHERE user_id = $1 AND contact_type = $2 AND is_primary = true AND is_active = true
		LIMIT 1`

	err := r.Get(ctx, &contact, query, userID, contactType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	return &contact, nil
}

// GetPublicContacts retrieves all public contacts for a user
func (r *UserContactRepository) GetPublicContacts(ctx context.Context, userID uuidv7.UUID) ([]*entity.UserContact, error) {
	var contacts []*entity.UserContact
	query := `
		SELECT id, user_id, contact_type, contact_value, label,
			is_verified, is_primary, is_active, is_public,
			available_from, available_to, available_days, timezone, notes,
			created_at, updated_at
		FROM user_contacts
		WHERE user_id = $1 AND is_public = true AND is_active = true
		ORDER BY is_primary DESC, contact_type, created_at DESC`

	err := r.Select(ctx, &contacts, query, userID)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

// Update updates an existing contact
func (r *UserContactRepository) Update(ctx context.Context, contact *entity.UserContact) error {
	query := `
		UPDATE user_contacts SET
			contact_type = $2,
			contact_value = $3,
			label = $4,
			is_verified = $5,
			is_primary = $6,
			is_active = $7,
			is_public = $8,
			available_from = $9,
			available_to = $10,
			available_days = $11,
			timezone = $12,
			notes = $13,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	err := r.Exec(ctx, query,
		contact.ID,
		contact.ContactType,
		contact.ContactValue,
		contact.Label,
		contact.IsVerified,
		contact.IsPrimary,
		contact.IsActive,
		contact.IsPublic,
		contact.AvailableFrom,
		contact.AvailableTo,
		contact.AvailableDays,
		contact.Timezone,
		contact.Notes,
	)

	return err
}

// Delete deletes a contact
func (r *UserContactRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM user_contacts WHERE id = $1`
	return r.Exec(ctx, query, id)
}

// SetPrimary sets a contact as primary (unsets other primary contacts of same type)
func (r *UserContactRepository) SetPrimary(ctx context.Context, userID uuidv7.UUID, contactType entity.ContactType, id uuidv7.UUID) error {
	// Get the contact first to verify it exists
	contact, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Unset all primary flags for this user and contact type
	query := `
		UPDATE user_contacts 
		SET is_primary = false 
		WHERE user_id = $1 AND contact_type = $2`

	_, err = tx.ExecContext(ctx, query, contact.UserID, contact.ContactType)
	if err != nil {
		return err
	}

	// Set the specified contact as primary
	query = `
		UPDATE user_contacts 
		SET is_primary = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return entity.ErrNotFound
	}

	return tx.Commit()
}

// VerifyContact marks a contact as verified
func (r *UserContactRepository) VerifyContact(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE user_contacts 
		SET is_verified = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	return r.Exec(ctx, query, id)
}
