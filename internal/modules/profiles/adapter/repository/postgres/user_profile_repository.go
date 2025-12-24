package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/jmoiron/sqlx"
)

type IUserProfileRepository struct {
	*BaseRepository
}

func NewUserProfileRepository(db *sqlx.DB) repository.IUserProfileRepository {
	return &IUserProfileRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *IUserProfileRepository) Create(ctx context.Context, profile *entity.UserProfile) error {
	// Marshal JSONB fields before saving
	if err := profile.MarshalSocialLinks(); err != nil {
		return err
	}
	if err := profile.MarshalPreferences(); err != nil {
		return err
	}

	query := `
		INSERT INTO profiles_profiles (
			id, user_id, first_name, last_name, middle_name, display_name, nickname,
			bio, date_of_birth, gender, country_id, city, timezone, locale,
			avatar_url, cover_url, social_links, website_url, company, job_title,
			is_public, is_verified, show_email, show_location, show_birthday,
			preferences, profile_views_count, followers_count, following_count,
			is_banned, ban_reason, banned_at, banned_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25,
			$26, $27, $28, $29,
			$30, $31, $32, $33
		)
		RETURNING created_at, updated_at
	`

	err := r.Get(ctx, profile, query,
		profile.ID, profile.UserID, profile.FirstName, profile.LastName, profile.MiddleName, profile.DisplayName, profile.Nickname,
		profile.Bio, profile.DateOfBirth, profile.Gender, profile.CountryID, profile.City, profile.Timezone, profile.Locale,
		profile.AvatarURL, profile.CoverURL, profile.SocialLinksJSON, profile.WebsiteURL, profile.Company, profile.JobTitle,
		profile.IsPublic, profile.IsVerified, profile.ShowEmail, profile.ShowLocation, profile.ShowBirthday,
		profile.PreferencesJSON, profile.ProfileViewsCount, profile.FollowersCount, profile.FollowingCount,
		profile.IsBanned, profile.BanReason, profile.BannedAt, profile.BannedBy,
	)

	return err
}

func (r *IUserProfileRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.UserProfile, error) {
	var profile entity.UserProfile
	query := `
		SELECT * FROM profiles_profiles WHERE id = $1
	`

	err := r.Get(ctx, &profile, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	// Unmarshal JSONB fields
	if err := profile.UnmarshalSocialLinks(); err != nil {
		return nil, err
	}
	if err := profile.UnmarshalPreferences(); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *IUserProfileRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.UserProfile, error) {
	var profile entity.UserProfile
	query := `
		SELECT * FROM profiles_profiles WHERE user_id = $1
	`

	err := r.Get(ctx, &profile, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	// Unmarshal JSONB fields
	if err := profile.UnmarshalSocialLinks(); err != nil {
		return nil, err
	}
	if err := profile.UnmarshalPreferences(); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *IUserProfileRepository) GetByNickname(ctx context.Context, nickname string) (*entity.UserProfile, error) {
	var profile entity.UserProfile
	query := `
		SELECT * FROM profiles_profiles WHERE nickname = $1
	`

	err := r.Get(ctx, &profile, query, nickname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, err
	}

	// Unmarshal JSONB fields
	if err := profile.UnmarshalSocialLinks(); err != nil {
		return nil, err
	}
	if err := profile.UnmarshalPreferences(); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *IUserProfileRepository) Update(ctx context.Context, profile *entity.UserProfile) error {
	// Marshal JSONB fields before saving
	if err := profile.MarshalSocialLinks(); err != nil {
		return err
	}
	if err := profile.MarshalPreferences(); err != nil {
		return err
	}

	query := `
		UPDATE profiles_profiles SET
			first_name = $2, last_name = $3, middle_name = $4, display_name = $5, nickname = $6,
			bio = $7, date_of_birth = $8, gender = $9, country_id = $10, city = $11,
			timezone = $12, locale = $13, avatar_url = $14, cover_url = $15,
			social_links = $16, website_url = $17, company = $18, job_title = $19,
			is_public = $20, show_email = $21, show_location = $22, show_birthday = $23,
			preferences = $24, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.Get(ctx, profile, query,
		profile.ID,
		profile.FirstName, profile.LastName, profile.MiddleName, profile.DisplayName, profile.Nickname,
		profile.Bio, profile.DateOfBirth, profile.Gender, profile.CountryID, profile.City,
		profile.Timezone, profile.Locale, profile.AvatarURL, profile.CoverURL,
		profile.SocialLinksJSON, profile.WebsiteURL, profile.Company, profile.JobTitle,
		profile.IsPublic, profile.ShowEmail, profile.ShowLocation, profile.ShowBirthday,
		profile.PreferencesJSON,
	)

	return err
}

func (r *IUserProfileRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `DELETE FROM profiles_profiles WHERE id = $1`

	err := r.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *IUserProfileRepository) List(ctx context.Context, limit, offset int, isPublic *bool) ([]*entity.UserProfile, error) {
	var profiles []*entity.UserProfile

	query := `
		SELECT * FROM profiles_profiles
		WHERE ($1::boolean IS NULL OR is_public = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.Select(ctx, &profiles, query, isPublic, limit, offset)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields for each profile
	for _, profile := range profiles {
		if err := profile.UnmarshalSocialLinks(); err != nil {
			return nil, err
		}
		if err := profile.UnmarshalPreferences(); err != nil {
			return nil, err
		}
	}

	return profiles, nil
}

func (r *IUserProfileRepository) UpdateLastSeen(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE profiles_profiles SET last_seen_at = NOW() WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *IUserProfileRepository) IncrementProfileViews(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE profiles_profiles SET profile_views_count = profile_views_count + 1 WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *IUserProfileRepository) Ban(ctx context.Context, id uuidv7.UUID, reason string, bannedBy uuidv7.UUID) error {
	query := `
		UPDATE profiles_profiles SET
			is_banned = true,
			ban_reason = $2,
			banned_at = NOW(),
			banned_by = $3
		WHERE id = $1
	`
	return r.Exec(ctx, query, id, reason, bannedBy)
}

func (r *IUserProfileRepository) Unban(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE profiles_profiles SET
			is_banned = false,
			ban_reason = NULL,
			banned_at = NULL,
			banned_by = NULL
		WHERE id = $1
	`
	return r.Exec(ctx, query, id)
}

func (r *IUserProfileRepository) SetVerified(ctx context.Context, id uuidv7.UUID, verified bool) error {
	query := `UPDATE profiles_profiles SET is_verified = $2 WHERE id = $1`
	return r.Exec(ctx, query, id, verified)
}

func (r *IUserProfileRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.UserProfile, error) {
	var profiles []*entity.UserProfile

	searchQuery := `
		SELECT * FROM profiles_profiles
		WHERE (
			nickname ILIKE '%' || $1 || '%' OR
			display_name ILIKE '%' || $1 || '%' OR
			bio ILIKE '%' || $1 || '%'
		)
		AND is_public = true
		AND is_banned = false
		ORDER BY is_verified DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.Select(ctx, &profiles, searchQuery, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSONB fields for each profile
	for _, profile := range profiles {
		if err := profile.UnmarshalSocialLinks(); err != nil {
			return nil, err
		}
		if err := profile.UnmarshalPreferences(); err != nil {
			return nil, err
		}
	}

	return profiles, nil
}
