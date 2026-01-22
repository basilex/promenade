package user_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	roleRepo "github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	roleAggregate "github.com/basilex/promenade/internal/contexts/identity/role/aggregate"
	roleRepository "github.com/basilex/promenade/internal/contexts/identity/role/repository"
	userRepo "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	userAggregate "github.com/basilex/promenade/internal/contexts/identity/user/aggregate"
	userRepository "github.com/basilex/promenade/internal/contexts/identity/user/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// setupBenchmarkDB creates a test database connection for benchmarks
func setupBenchmarkDB(b *testing.B) *sqlx.DB {
	// Use test database
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://system:passw0rd@localhost:5433/promenade_test?sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		b.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean all tables
	tables := []string{
		"identity_user_roles",
		"identity_role_permissions",
		"identity_users",
		"identity_roles",
		"identity_permissions",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			b.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}

	return db
}

// BenchmarkListUsers_SmallDataset benchmarks ListUsers with 20 users (typical page size)
func BenchmarkListUsers_SmallDataset(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer func() { _ = db.Close() }()

	repo := userRepo.NewUserRepository(db)
	roleRepository := roleRepo.NewRoleRepository(db)
	ctx := context.Background()

	// Setup: Create 20 users with roles
	setupUsersWithRoles(b, ctx, db, repo, roleRepository, 20)

	// Reset timer after setup
	b.ResetTimer()

	// Benchmark
	for i := 0; i < b.N; i++ {
		users, total, err := repo.ListUsers(ctx, 1, 20)
		if err != nil {
			b.Fatalf("ListUsers failed: %v", err)
		}
		if len(users) != 20 {
			b.Fatalf("Expected 20 users, got %d", len(users))
		}
		if total != 20 {
			b.Fatalf("Expected total 20, got %d", total)
		}
		// Verify roles are loaded
		for _, u := range users {
			if len(u.Roles) == 0 {
				b.Fatal("Expected roles to be loaded, got empty array")
			}
		}
	}
}

// BenchmarkListUsers_MediumDataset benchmarks ListUsers with 100 users
func BenchmarkListUsers_MediumDataset(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer func() { _ = db.Close() }()

	repo := userRepo.NewUserRepository(db)
	roleRepository := roleRepo.NewRoleRepository(db)
	ctx := context.Background()

	// Setup: Create 100 users with roles
	setupUsersWithRoles(b, ctx, db, repo, roleRepository, 100)

	// Reset timer after setup
	b.ResetTimer()

	// Benchmark - fetch first page of 20
	for i := 0; i < b.N; i++ {
		users, total, err := repo.ListUsers(ctx, 1, 20)
		if err != nil {
			b.Fatalf("ListUsers failed: %v", err)
		}
		if len(users) != 20 {
			b.Fatalf("Expected 20 users, got %d", len(users))
		}
		if total != 100 {
			b.Fatalf("Expected total 100, got %d", total)
		}
		// Verify roles are loaded
		for _, u := range users {
			if len(u.Roles) == 0 {
				b.Fatal("Expected roles to be loaded, got empty array")
			}
		}
	}
}

// BenchmarkListUsers_NoRoles benchmarks ListUsers with users that have no roles
func BenchmarkListUsers_NoRoles(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer func() { _ = db.Close() }()

	repo := userRepo.NewUserRepository(db)
	ctx := context.Background()

	// Setup: Create 20 users WITHOUT roles
	for i := 0; i < 20; i++ {
		u, _ := userAggregate.NewUser(fmt.Sprintf("user%d@test.com", i), "password123")
		if err := repo.Create(ctx, u); err != nil {
			b.Fatalf("Failed to create user: %v", err)
		}
	}

	// Reset timer after setup
	b.ResetTimer()

	// Benchmark
	for i := 0; i < b.N; i++ {
		users, total, err := repo.ListUsers(ctx, 1, 20)
		if err != nil {
			b.Fatalf("ListUsers failed: %v", err)
		}
		if len(users) != 20 {
			b.Fatalf("Expected 20 users, got %d", len(users))
		}
		if total != 20 {
			b.Fatalf("Expected total 20, got %d", total)
		}
		// Verify empty roles array (not nil)
		for _, u := range users {
			if u.Roles == nil {
				b.Fatal("Expected empty roles array, got nil")
			}
		}
	}
}

// BenchmarkListUsers_MultipleRoles benchmarks users with multiple roles each
func BenchmarkListUsers_MultipleRoles(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer func() { _ = db.Close() }()

	repo := userRepo.NewUserRepository(db)
	roleRepository := roleRepo.NewRoleRepository(db)
	ctx := context.Background()

	// Setup: Create roles
	adminRole, _ := roleAggregate.NewRole("admin", "Administrator", "Administrator role")
	managerRole, _ := roleAggregate.NewRole("manager", "Manager", "Manager role")
	userRole, _ := roleAggregate.NewRole("user", "User", "User role")

	_ = roleRepository.Create(ctx, adminRole)
	_ = roleRepository.Create(ctx, managerRole)
	_ = roleRepository.Create(ctx, userRole)

	// Setup: Create 20 users, each with 3 roles
	for i := 0; i < 20; i++ {
		u, _ := userAggregate.NewUser(fmt.Sprintf("user%d@test.com", i), "password123")
		if err := repo.Create(ctx, u); err != nil {
			b.Fatalf("Failed to create user: %v", err)
		}

		// Assign all 3 roles to each user
		_ = roleRepository.AssignRoleToUser(ctx, u.ID, adminRole.ID, nil)
		_ = roleRepository.AssignRoleToUser(ctx, u.ID, managerRole.ID, nil)
		_ = roleRepository.AssignRoleToUser(ctx, u.ID, userRole.ID, nil)
	}

	// Reset timer after setup
	b.ResetTimer()

	// Benchmark
	for i := 0; i < b.N; i++ {
		users, total, err := repo.ListUsers(ctx, 1, 20)
		if err != nil {
			b.Fatalf("ListUsers failed: %v", err)
		}
		if len(users) != 20 {
			b.Fatalf("Expected 20 users, got %d", len(users))
		}
		if total != 20 {
			b.Fatalf("Expected total 20, got %d", total)
		}
		// Verify each user has 3 roles
		for _, u := range users {
			if len(u.Roles) != 3 {
				b.Fatalf("Expected 3 roles per user, got %d", len(u.Roles))
			}
		}
	}
}

// Helper function to setup users with roles for benchmarks
func setupUsersWithRoles(b *testing.B, ctx context.Context, db *sqlx.DB, repo userRepository.IUserRepository, roleRepository roleRepository.IRoleRepository, count int) {
	// Create a default role
	defaultRole, _ := roleAggregate.NewRole("user", "User", "Default user role")
	if err := roleRepository.Create(ctx, defaultRole); err != nil {
		b.Fatalf("Failed to create role: %v", err)
	}

	// Create users and assign role
	for i := 0; i < count; i++ {
		u, _ := userAggregate.NewUser(fmt.Sprintf("user%d_%s@test.com", i, uuidv7.New().String()[:8]), "password123")
		if err := repo.Create(ctx, u); err != nil {
			b.Fatalf("Failed to create user: %v", err)
		}

		// Assign role
		if err := roleRepository.AssignRoleToUser(ctx, u.ID, defaultRole.ID, nil); err != nil {
			b.Fatalf("Failed to assign role: %v", err)
		}
	}
}
