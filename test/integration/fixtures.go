package integration

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Fixtures provides helper methods for creating test data
type Fixtures struct {
	db *sqlx.DB
}

// NewFixtures creates a new fixtures helper
func NewFixtures(db *sqlx.DB) *Fixtures {
	return &Fixtures{db: db}
}

// CreateUser creates a test user in the database
func (f *Fixtures) CreateUser(t *testing.T, email, password string) *entity.User {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &entity.User{
		ID:        uuidv7.New(),
		Email:     email,
		Name:      "Test User",
		Password:  string(hashedPassword),
		Status:    entity.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO core_users (id, email, name, password, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = f.db.Exec(query, user.ID, user.Email, user.Name, user.Password, user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	return user
}

// CreateRole creates a test role in the database
func (f *Fixtures) CreateRole(t *testing.T, name, description string, isSystem bool) *entity.Role {
	t.Helper()

	desc := description
	role := &entity.Role{
		ID:          uuidv7.New(),
		Name:        name,
		DisplayName: name,
		Description: &desc,
		IsSystem:    isSystem,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO core_roles (id, name, display_name, description, is_system, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := f.db.Exec(query, role.ID, role.Name, role.DisplayName, role.Description, role.IsSystem, role.CreatedAt, role.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	return role
}

// CreatePermission creates a test permission in the database
func (f *Fixtures) CreatePermission(t *testing.T, resource, action, description string) *entity.Permission {
	t.Helper()

	desc := description
	permission := &entity.Permission{
		ID:          uuidv7.New(),
		Resource:    resource,
		Action:      action,
		Description: &desc,
		CreatedAt:   time.Now(),
	}

	query := `
		INSERT INTO core_permissions (id, resource, action, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := f.db.Exec(query, permission.ID, permission.Resource, permission.Action, permission.Description, permission.CreatedAt)
	if err != nil {
		t.Fatalf("Failed to create permission: %v", err)
	}

	return permission
}

// AssignRoleToUser assigns a role to a user
func (f *Fixtures) AssignRoleToUser(t *testing.T, userID, roleID uuidv7.UUID) {
	t.Helper()

	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`
	_, err := f.db.Exec(query, userID, roleID)
	if err != nil {
		t.Fatalf("Failed to assign role to user: %v", err)
	}
}

// AssignPermissionToRole assigns a permission to a role
func (f *Fixtures) AssignPermissionToRole(t *testing.T, roleID, permissionID uuidv7.UUID) {
	t.Helper()

	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`
	_, err := f.db.Exec(query, roleID, permissionID)
	if err != nil {
		t.Fatalf("Failed to assign permission to role: %v", err)
	}
}

// CreateSession creates a test session
func (f *Fixtures) CreateSession(t *testing.T, userID uuidv7.UUID) *entity.Session {
	t.Helper()

	session := &entity.Session{
		ID:           uuidv7.New(),
		UserID:       userID,
		RefreshToken: uuidv7.New().String(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	query := `
		INSERT INTO sessions (id, user_id, refresh_token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := f.db.Exec(query, session.ID, session.UserID, session.RefreshToken, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	return session
}

// CreateCountry creates a test country
func (f *Fixtures) CreateCountry(t *testing.T, name, code, iso2, iso3, region string) *entity.Country {
	t.Helper()

	country := &entity.Country{
		ID:        uuidv7.New(),
		Name:      name,
		Code:      code,
		ISO2:      iso2,
		ISO3:      iso3,
		Region:    region,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO core_countries (id, name, code, iso2, iso3, region, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := f.db.Exec(query, country.ID, country.Name, country.Code, country.ISO2, country.ISO3, country.Region, country.CreatedAt, country.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create country: %v", err)
	}

	return country
}

// CreateCurrency creates a test currency
func (f *Fixtures) CreateCurrency(t *testing.T, name, code, symbol string) *entity.Currency {
	t.Helper()

	currency := &entity.Currency{
		ID:        uuidv7.New(),
		Name:      name,
		Code:      code,
		Symbol:    symbol,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO core_currencies (id, name, code, symbol, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := f.db.Exec(query, currency.ID, currency.Name, currency.Code, currency.Symbol, currency.CreatedAt, currency.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create currency: %v", err)
	}

	return currency
}

// CreateLanguage creates a test language
func (f *Fixtures) CreateLanguage(t *testing.T, name, nativeName, code, iso6392 string) *entity.Language {
	t.Helper()

	language := &entity.Language{
		ID:         uuidv7.New(),
		Name:       name,
		NativeName: nativeName,
		Code:       code,
		ISO639_2:   iso6392,
		IsRtl:      false,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	query := `
		INSERT INTO core_languages (id, name, native_name, code, iso639_2, is_rtl, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := f.db.Exec(query, language.ID, language.Name, language.NativeName, language.Code, language.ISO639_2, language.IsRtl, language.IsActive, language.CreatedAt, language.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create language: %v", err)
	}

	return language
}

// CreateTimezone creates a test timezone
func (f *Fixtures) CreateTimezone(t *testing.T, name, abbreviation, utcOffset string) *entity.Timezone {
	t.Helper()

	timezone := &entity.Timezone{
		ID:           uuidv7.New(),
		Name:         name,
		Abbreviation: abbreviation,
		UtcOffset:    utcOffset,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO core_timezones (id, name, abbreviation, utc_offset, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := f.db.Exec(query, timezone.ID, timezone.Name, timezone.Abbreviation, timezone.UtcOffset, timezone.IsActive, timezone.CreatedAt, timezone.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create timezone: %v", err)
	}

	return timezone
}

// WithTestContext returns a context suitable for testing
func WithTestContext() context.Context {
	return context.Background()
}
