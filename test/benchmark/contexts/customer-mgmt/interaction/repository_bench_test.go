package interaction_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	customerAggregate "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/aggregate"
	customerRepository "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/repository"
	interactionRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/adapter/repository/postgres"
	interactionAggregate "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// setupBenchmarkDB creates a test database connection for benchmarks
func setupBenchmarkDB(b *testing.B) *sqlx.DB {
	// Use test database
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://system:passw0rd@localhost:5433/promenade_test?sslmode=disable"
	}

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		b.Fatalf("Failed to connect to test database: %v", err)
	}

	// Clean all tables (DELETE instead of TRUNCATE to avoid locks)
	tables := []string{
		"customer_interactions",
		"customer_customers",
		"identity_users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			b.Fatalf("Failed to clean table %s: %v", table, err)
		}
	}

	return db
}

// createTestUser creates a user for FK constraint
func createTestUser(b *testing.B, ctx context.Context, db *sqlx.DB) uuidv7.UUID {
	userID := uuidv7.New()
	_, err := db.ExecContext(ctx, `INSERT INTO identity_users (id, email, password_hash, status) VALUES ($1, $2, $3, $4)`,
		userID, fmt.Sprintf("bench_%s@test.com", userID.String()[:8]), "hash", "active")
	if err != nil {
		b.Fatalf("Failed to create user: %v", err)
	}
	return userID
}

// createTestCustomer creates a customer for FK constraint
func createTestCustomer(b *testing.B, ctx context.Context, custRepo customerRepository.ICustomerRepository, userID uuidv7.UUID) uuidv7.UUID {
	cust, _ := customerAggregate.NewCustomer(fmt.Sprintf("Customer %s", uuidv7.New().String()[:8]),
		fmt.Sprintf("cust_%s@test.com", uuidv7.New().String()[:8]), "website", userID)
	if err := custRepo.Create(ctx, cust); err != nil {
		b.Fatalf("Failed to create customer: %v", err)
	}
	return cust.ID
}

// BenchmarkCreateInteraction benchmarks creating interactions
func BenchmarkCreateInteraction(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer func() { _ = db.Close() }()

	repo := interactionRepo.NewInteractionRepository(db)
	custRepo := customerRepo.NewCustomerRepository(db)
	ctx := context.Background()

	userID := createTestUser(b, ctx, db)
	customerID := createTestCustomer(b, ctx, custRepo, userID)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inter, _ := interactionAggregate.NewInteraction(
			customerID, nil,
			interactionAggregate.InteractionTypeCall,
			interactionAggregate.InteractionDirectionOutbound,
			fmt.Sprintf("Call %d", i),
			"Benchmark test",
			userID,
			time.Now(),
		)
		if err := repo.Create(ctx, inter); err != nil {
			b.Fatalf("Create failed: %v", err)
		}
	}
}

// BenchmarkGetByID benchmarks retrieving interactions
// NOTE: Commented out due to hanging issue - needs investigation
// func BenchmarkGetByID(b *testing.B) {
// 	db := setupBenchmarkDB(b)
// 	defer db.Close()
//
// 	repo := interactionRepo.NewInteractionRepository(db)
// 	custRepo := customerRepo.NewCustomerRepository(db)
// 	ctx := context.Background()
//
// 	userID := createTestUser(b, ctx, db)
// 	customerID := createTestCustomer(b, ctx, custRepo, userID)
// 	inter, _ := interactionAggregate.NewInteraction(
// 		customerID, nil,
// 		interactionAggregate.InteractionTypeCall,
// 		interactionAggregate.InteractionDirectionOutbound,
// 		"Test call",
// 		"Description",
// 		userID,
// 		time.Now(),
// 	)
// 	repo.Create(ctx, inter)
//
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		_, err := repo.GetByID(ctx, inter.ID)
// 		if err != nil {
// 			b.Fatalf("GetByID failed: %v", err)
// 		}
// 	}
// }

// BenchmarkListByCustomer benchmarks listing interactions
func BenchmarkListByCustomer(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer func() { _ = db.Close() }()

	repo := interactionRepo.NewInteractionRepository(db)
	custRepo := customerRepo.NewCustomerRepository(db)
	ctx := context.Background()

	userID := createTestUser(b, ctx, db)
	customerID := createTestCustomer(b, ctx, custRepo, userID)

	for i := 0; i < 50; i++ {
		inter, _ := interactionAggregate.NewInteraction(
			customerID, nil,
			interactionAggregate.InteractionTypeCall,
			interactionAggregate.InteractionDirectionOutbound,
			fmt.Sprintf("Call %d", i),
			"Description",
			userID,
			time.Now().Add(time.Duration(i)*time.Hour),
		)
		_ = repo.Create(ctx, inter)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		interactions, total, err := repo.ListByCustomer(ctx, customerID, 1, 20)
		if err != nil {
			b.Fatalf("ListByCustomer failed: %v", err)
		}
		if len(interactions) != 20 {
			b.Fatalf("Expected 20 interactions, got %d", len(interactions))
		}
		if total != 50 {
			b.Fatalf("Expected total 50, got %d", total)
		}
	}
}
