package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// userRepository implements user.IRepository for PostgreSQL
type userRepository struct {
	*BaseRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) user.IRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// userRow represents database row structure for identity_users table
type userRow struct {
	ID               uuidv7.UUID  `db:"id"`
	Email            string       `db:"email"`
	PasswordHash     string       `db:"password_hash"`
	Status           string       `db:"status"`
	EmailVerified    bool         `db:"email_verified"`
	EmailVerifiedAt  sql.NullTime `db:"email_verified_at"`
	LastLoginAt      sql.NullTime `db:"last_login_at"`
	FailedLoginCount int          `db:"failed_login_count"`
	LockedUntil      sql.NullTime `db:"locked_until"`
	CreatedAt        string       `db:"created_at"`
	UpdatedAt        string       `db:"updated_at"`
	DeletedAt        sql.NullTime `db:"deleted_at"`
}

// userRowWithRoles extends userRow with roles array (for batch loading optimization)
type userRowWithRoles struct {
	userRow
	Roles pq.StringArray `db:"roles"` // PostgreSQL TEXT[] array (use pq.StringArray for scanning)
}

// toEntity converts database row to domain entity
func (r *userRow) toEntity() (*user.User, error) {
	// Validate and create email value object
	emailVO, err := valueobject.NewEmail(r.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email in database: %w", err)
	}

	u := &user.User{
		ID:               r.ID,
		Email:            emailVO,
		PasswordHash:     r.PasswordHash,
		Status:           user.UserStatus(r.Status),
		EmailVerified:    r.EmailVerified,
		FailedLoginCount: r.FailedLoginCount,
	}

	if r.EmailVerifiedAt.Valid {
		u.EmailVerifiedAt = &r.EmailVerifiedAt.Time
	}
	if r.LastLoginAt.Valid {
		u.LastLoginAt = &r.LastLoginAt.Time
	}
	if r.LockedUntil.Valid {
		u.LockedUntil = &r.LockedUntil.Time
	}

	return u, nil
}

// fromEntity converts domain entity to database row
func fromEntity(u *user.User) *userRow {
	row := &userRow{
		ID:               u.ID,
		Email:            u.Email.Value(),
		PasswordHash:     u.PasswordHash,
		Status:           string(u.Status),
		EmailVerified:    u.EmailVerified,
		FailedLoginCount: u.FailedLoginCount,
	}

	if u.EmailVerifiedAt != nil {
		row.EmailVerifiedAt = sql.NullTime{Time: *u.EmailVerifiedAt, Valid: true}
	}
	if u.LastLoginAt != nil {
		row.LastLoginAt = sql.NullTime{Time: *u.LastLoginAt, Valid: true}
	}
	if u.LockedUntil != nil {
		row.LockedUntil = sql.NullTime{Time: *u.LockedUntil, Valid: true}
	}

	return row
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	row := fromEntity(u)

	query := `
		INSERT INTO identity_users (
			id, email, password_hash, status, email_verified, 
			email_verified_at, last_login_at, failed_login_count, locked_until
		)
		VALUES (
			:id, :email, :password_hash, :status, :email_verified, 
			:email_verified_at, :last_login_at, :failed_login_count, :locked_until
		)
	`

	_, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*user.User, error) {
	var row userRow

	query := `
		SELECT 
			id, email, password_hash, status, email_verified, 
			email_verified_at, last_login_at, failed_login_count, locked_until,
			created_at, updated_at, deleted_at
		FROM identity_users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.Get(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	u, err := row.toEntity()
	if err != nil {
		return nil, err
	}

	// Load user roles
	if err := r.loadUserRoles(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to load user roles: %w", err)
	}

	return u, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var row userRow

	query := `
		SELECT 
			id, email, password_hash, status, email_verified, 
			email_verified_at, last_login_at, failed_login_count, locked_until,
			created_at, updated_at, deleted_at
		FROM identity_users
		WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL
	`

	err := r.Get(ctx, &row, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	u, err := row.toEntity()
	if err != nil {
		return nil, err
	}

	// Load user roles
	if err := r.loadUserRoles(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to load user roles: %w", err)
	}

	return u, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	row := fromEntity(u)

	query := `
		UPDATE identity_users
		SET 
			email = :email,
			password_hash = :password_hash,
			status = :status,
			email_verified = :email_verified,
			email_verified_at = :email_verified_at,
			last_login_at = :last_login_at,
			failed_login_count = :failed_login_count,
			locked_until = :locked_until,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND deleted_at IS NULL
	`

	result, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// Delete soft deletes a user (sets deleted_at)
func (r *userRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE identity_users
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM identity_users 
			WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL
		)
	`

	err := r.Get(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	return exists, nil
}

// ListUsers retrieves a paginated list of users with roles loaded in a single query
// Optimization: Uses LEFT JOIN with ARRAY_AGG to avoid N+1 query problem
func (r *userRepository) ListUsers(ctx context.Context, page, pageSize int) ([]*user.User, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) 
		FROM identity_users 
		WHERE deleted_at IS NULL
	`

	err := r.Get(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results WITH roles in single query (avoid N+1)
	// Uses LEFT JOIN + ARRAY_AGG to batch load all roles
	var rows []userRowWithRoles
	query := `
		SELECT 
			u.id, u.email, u.password_hash, u.status, u.email_verified, 
			u.email_verified_at, u.last_login_at, u.failed_login_count, u.locked_until,
			u.created_at, u.updated_at, u.deleted_at,
			COALESCE(
				ARRAY_AGG(r.name ORDER BY r.name) FILTER (WHERE r.name IS NOT NULL), 
				ARRAY[]::TEXT[]
			) AS roles
		FROM identity_users u
		LEFT JOIN identity_user_roles ur ON u.id = ur.user_id
		LEFT JOIN identity_roles r ON ur.role_id = r.id
		WHERE u.deleted_at IS NULL
		GROUP BY u.id, u.email, u.password_hash, u.status, u.email_verified,
				 u.email_verified_at, u.last_login_at, u.failed_login_count, u.locked_until,
				 u.created_at, u.updated_at, u.deleted_at
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2
	`

	err = r.Select(ctx, &rows, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert rows to entities (roles already loaded)
	users := make([]*user.User, 0, len(rows))
	for _, row := range rows {
		u, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		// Assign roles from batch-loaded array (convert pq.StringArray to []string)
		u.Roles = []string(row.Roles)
		users = append(users, u)
	}

	return users, total, nil
}

// loadUserRoles loads roles for a user from identity_user_roles and identity_roles tables
func (r *userRepository) loadUserRoles(ctx context.Context, u *user.User) error {
	query := `
		SELECT r.name
		FROM identity_roles r
		INNER JOIN identity_user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name
	`

	var roles []string
	err := r.Select(ctx, &roles, query, u.ID)
	if err != nil {
		return fmt.Errorf("failed to load user roles: %w", err)
	}

	u.Roles = roles
	return nil
}
