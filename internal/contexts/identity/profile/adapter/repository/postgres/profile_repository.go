package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/identity/profile"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// profileRepository implements profile.IRepository for PostgreSQL
type profileRepository struct {
	*BaseRepository
}

// NewProfileRepository creates a new profile repository
func NewProfileRepository(db *sqlx.DB) profile.IRepository {
	return &profileRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// profileRow represents database row structure for identity_profiles table
type profileRow struct {
	ID          uuidv7.UUID    `db:"id"`
	UserID      uuidv7.UUID    `db:"user_id"`
	DisplayName string         `db:"display_name"`
	Bio         sql.NullString `db:"bio"`
	AvatarURL   sql.NullString `db:"avatar_url"`
	FirstName   sql.NullString `db:"first_name"`
	LastName    sql.NullString `db:"last_name"`
	MiddleName  sql.NullString `db:"middle_name"`
	Gender      string         `db:"gender"`
	DateOfBirth sql.NullTime   `db:"date_of_birth"`
	Timezone    sql.NullString `db:"timezone"`
	Language    sql.NullString `db:"language"`
	Country     sql.NullString `db:"country"`
	Website     sql.NullString `db:"website"`
	LinkedIn    sql.NullString `db:"linkedin"`
	Twitter     sql.NullString `db:"twitter"`
	GitHub      sql.NullString `db:"github"`
	Facebook    sql.NullString `db:"facebook"`
	Instagram   sql.NullString `db:"instagram"`
	IsPublic    bool           `db:"is_public"`
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *profileRow) toEntity() *profile.Profile {
	p := &profile.Profile{
		UserID:      r.UserID,
		DisplayName: r.DisplayName,
		Gender:      profile.Gender(r.Gender),
		IsPublic:    r.IsPublic,
		IsActive:    r.IsActive,
	}
	// Initialize BaseAggregate fields
	p.ID = r.ID
	p.CreatedAt = r.CreatedAt
	p.UpdatedAt = r.UpdatedAt

	if r.Bio.Valid {
		p.Bio = r.Bio.String
	}
	if r.AvatarURL.Valid {
		p.AvatarURL = r.AvatarURL.String
	}
	if r.FirstName.Valid {
		p.FirstName = r.FirstName.String
	}
	if r.LastName.Valid {
		p.LastName = r.LastName.String
	}
	if r.MiddleName.Valid {
		p.MiddleName = r.MiddleName.String
	}
	if r.DateOfBirth.Valid {
		p.DateOfBirth = &r.DateOfBirth.Time
	}
	if r.Timezone.Valid {
		p.Timezone = r.Timezone.String
	}
	if r.Language.Valid {
		p.Language = r.Language.String
	}
	if r.Country.Valid {
		p.Country = r.Country.String
	}
	if r.Website.Valid {
		p.Website = r.Website.String
	}
	if r.LinkedIn.Valid {
		p.LinkedIn = r.LinkedIn.String
	}
	if r.Twitter.Valid {
		p.Twitter = r.Twitter.String
	}
	if r.GitHub.Valid {
		p.GitHub = r.GitHub.String
	}
	if r.Facebook.Valid {
		p.Facebook = r.Facebook.String
	}
	if r.Instagram.Valid {
		p.Instagram = r.Instagram.String
	}

	return p
}

// fromEntity converts domain entity to database row
func fromEntity(p *profile.Profile) *profileRow {
	row := &profileRow{
		ID:          p.GetID(),
		UserID:      p.UserID,
		DisplayName: p.DisplayName,
		Gender:      string(p.Gender),
		IsPublic:    p.IsPublic,
		IsActive:    p.IsActive,
		CreatedAt:   p.GetCreatedAt(),
		UpdatedAt:   p.GetUpdatedAt(),
	}

	if p.Bio != "" {
		row.Bio = sql.NullString{String: p.Bio, Valid: true}
	}
	if p.AvatarURL != "" {
		row.AvatarURL = sql.NullString{String: p.AvatarURL, Valid: true}
	}
	if p.FirstName != "" {
		row.FirstName = sql.NullString{String: p.FirstName, Valid: true}
	}
	if p.LastName != "" {
		row.LastName = sql.NullString{String: p.LastName, Valid: true}
	}
	if p.MiddleName != "" {
		row.MiddleName = sql.NullString{String: p.MiddleName, Valid: true}
	}
	if p.DateOfBirth != nil {
		row.DateOfBirth = sql.NullTime{Time: *p.DateOfBirth, Valid: true}
	}
	if p.Timezone != "" {
		row.Timezone = sql.NullString{String: p.Timezone, Valid: true}
	}
	if p.Language != "" {
		row.Language = sql.NullString{String: p.Language, Valid: true}
	}
	if p.Country != "" {
		row.Country = sql.NullString{String: p.Country, Valid: true}
	}
	if p.Website != "" {
		row.Website = sql.NullString{String: p.Website, Valid: true}
	}
	if p.LinkedIn != "" {
		row.LinkedIn = sql.NullString{String: p.LinkedIn, Valid: true}
	}
	if p.Twitter != "" {
		row.Twitter = sql.NullString{String: p.Twitter, Valid: true}
	}
	if p.GitHub != "" {
		row.GitHub = sql.NullString{String: p.GitHub, Valid: true}
	}
	if p.Facebook != "" {
		row.Facebook = sql.NullString{String: p.Facebook, Valid: true}
	}
	if p.Instagram != "" {
		row.Instagram = sql.NullString{String: p.Instagram, Valid: true}
	}

	return row
}

// Create inserts a new profile
func (r *profileRepository) Create(ctx context.Context, p *profile.Profile) error {
	query := `
		INSERT INTO identity_profiles (
			id, user_id, display_name, bio, avatar_url,
			first_name, last_name, middle_name, gender, date_of_birth,
			timezone, language, country,
			website, linkedin, twitter, github, facebook, instagram,
			is_public, is_active, created_at, updated_at
		) VALUES (
			:id, :user_id, :display_name, :bio, :avatar_url,
			:first_name, :last_name, :middle_name, :gender, :date_of_birth,
			:timezone, :language, :country,
			:website, :linkedin, :twitter, :github, :facebook, :instagram,
			:is_public, :is_active, :created_at, :updated_at
		)`

	row := fromEntity(p)
	return r.NamedExec(ctx, query, row)
}

// GetByID retrieves a profile by ID
func (r *profileRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*profile.Profile, error) {
	query := `
		SELECT id, user_id, display_name, bio, avatar_url,
		       first_name, last_name, middle_name, gender, date_of_birth,
		       timezone, language, country,
		       website, linkedin, twitter, github, facebook, instagram,
		       is_public, is_active, created_at, updated_at, deleted_at
		FROM identity_profiles
		WHERE id = $1 AND deleted_at IS NULL`

	var row profileRow
	if err := r.Get(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, profile.ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return row.toEntity(), nil
}

// GetByUserID retrieves a profile by user ID
func (r *profileRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*profile.Profile, error) {
	query := `
		SELECT id, user_id, display_name, bio, avatar_url,
		       first_name, last_name, middle_name, gender, date_of_birth,
		       timezone, language, country,
		       website, linkedin, twitter, github, facebook, instagram,
		       is_public, is_active, created_at, updated_at, deleted_at
		FROM identity_profiles
		WHERE user_id = $1 AND deleted_at IS NULL`

	var row profileRow
	if err := r.Get(ctx, &row, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, profile.ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to get profile by user ID: %w", err)
	}

	return row.toEntity(), nil
}

// Update updates an existing profile
func (r *profileRepository) Update(ctx context.Context, p *profile.Profile) error {
	query := `
		UPDATE identity_profiles SET
			display_name = :display_name,
			bio = :bio,
			avatar_url = :avatar_url,
			first_name = :first_name,
			last_name = :last_name,
			middle_name = :middle_name,
			gender = :gender,
			date_of_birth = :date_of_birth,
			timezone = :timezone,
			language = :language,
			country = :country,
			website = :website,
			linkedin = :linkedin,
			twitter = :twitter,
			github = :github,
			facebook = :facebook,
			instagram = :instagram,
			is_public = :is_public,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	row := fromEntity(p)
	return r.NamedExec(ctx, query, row)
}

// Delete soft-deletes a profile
func (r *profileRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE identity_profiles
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	return r.Exec(ctx, query, id)
}

// ListPublicProfiles lists all public active profiles with pagination
func (r *profileRepository) ListPublicProfiles(ctx context.Context, limit, offset int) ([]*profile.Profile, error) {
	query := `
		SELECT id, user_id, display_name, bio, avatar_url,
		       first_name, last_name, middle_name, gender, date_of_birth,
		       timezone, language, country,
		       website, linkedin, twitter, github, facebook, instagram,
		       is_public, is_active, created_at, updated_at, deleted_at
		FROM identity_profiles
		WHERE is_public = true AND is_active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	var rows []profileRow
	if err := r.Select(ctx, &rows, query, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to list public profiles: %w", err)
	}

	profiles := make([]*profile.Profile, len(rows))
	for i, row := range rows {
		profiles[i] = row.toEntity()
	}

	return profiles, nil
}

// ExistsForUser checks if a profile exists for user
func (r *profileRepository) ExistsForUser(ctx context.Context, userID uuidv7.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM identity_profiles
			WHERE user_id = $1 AND deleted_at IS NULL
		)`

	var exists bool
	if err := r.Get(ctx, &exists, query, userID); err != nil {
		return false, fmt.Errorf("failed to check profile existence: %w", err)
	}

	return exists, nil
}
