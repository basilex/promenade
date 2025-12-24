//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/internal/modules/audit/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// setupTestDB initializes test database connection
func setupTestDB(t *testing.T) *sqlx.DB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping integration test: TEST_DATABASE_URL not set")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	// Ensure audit_events table exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_events (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			action VARCHAR(255) NOT NULL,
			entity_type VARCHAR(255) NOT NULL,
			entity_id UUID NOT NULL,
			old_data TEXT,
			new_data TEXT,
			ip_address VARCHAR(45),
			user_agent TEXT,
			request_id VARCHAR(255),
			signature TEXT,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "Failed to ensure audit_events table exists")

	// Clean test data
	_, err = db.Exec("DELETE FROM audit_events WHERE TRUE")
	require.NoError(t, err, "Failed to clean test data")

	return db
}

// createTestAuditEvent creates a test audit event with all fields
func createTestAuditEvent() *entity.AuditEvent {
	userID := uuidv7.New()
	entityID := uuidv7.New()
	oldData := `{"status":"draft"}`
	newData := `{"status":"published"}`
	metadata := `{"ip":"127.0.0.1"}`

	return &entity.AuditEvent{
		ID:         uuidv7.New(),
		UserID:     userID,
		Action:     "post.publish",
		EntityType: "post",
		EntityID:   entityID,
		OldData:    &oldData,
		NewData:    &newData,
		IPAddress:  "127.0.0.1",
		UserAgent:  "Mozilla/5.0",
		RequestID:  "req-123",
		Signature:  "signature-abc",
		Metadata:   &metadata,
		CreatedAt:  time.Now(),
	}
}

func TestAuditEventRepository_Create_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	t.Run("creates audit event successfully", func(t *testing.T) {
		event := createTestAuditEvent()

		err := repo.Create(ctx, event)
		require.NoError(t, err)

		// Verify UUID is set
		assert.NotEqual(t, uuidv7.UUID{}, event.ID)

		// Verify in database
		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM audit_events WHERE id = $1", event.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("creates event with all fields", func(t *testing.T) {
		event := createTestAuditEvent()

		err := repo.Create(ctx, event)
		require.NoError(t, err)

		// Retrieve and verify all fields
		var retrieved entity.AuditEvent
		query := `
			SELECT id, user_id, action, entity_type, entity_id, 
			       old_data, new_data, ip_address, user_agent, 
			       request_id, signature, metadata, created_at
			FROM audit_events WHERE id = $1
		`
		err = db.Get(&retrieved, query, event.ID)
		require.NoError(t, err)

		assert.Equal(t, event.ID, retrieved.ID)
		assert.Equal(t, event.UserID, retrieved.UserID)
		assert.Equal(t, event.Action, retrieved.Action)
		assert.Equal(t, event.EntityType, retrieved.EntityType)
		assert.Equal(t, event.EntityID, retrieved.EntityID)
		assert.Equal(t, *event.OldData, *retrieved.OldData)
		assert.Equal(t, *event.NewData, *retrieved.NewData)
		assert.Equal(t, event.IPAddress, retrieved.IPAddress)
		assert.Equal(t, event.UserAgent, retrieved.UserAgent)
		assert.Equal(t, event.RequestID, retrieved.RequestID)
		assert.Equal(t, event.Signature, retrieved.Signature)
	})

	t.Run("creates minimal event without optional fields", func(t *testing.T) {
		event := &entity.AuditEvent{
			ID:         uuidv7.New(),
			UserID:     uuidv7.New(),
			Action:     "user.login",
			EntityType: "user",
			EntityID:   uuidv7.New(),
			Signature:  "sig-123",
			CreatedAt:  time.Now(),
		}

		err := repo.Create(ctx, event)
		require.NoError(t, err)

		// Verify minimal event
		var retrieved entity.AuditEvent
		query := "SELECT * FROM audit_events WHERE id = $1"
		err = db.Get(&retrieved, query, event.ID)
		require.NoError(t, err)

		assert.Nil(t, retrieved.OldData)
		assert.Nil(t, retrieved.NewData)
		assert.Nil(t, retrieved.Metadata)
	})
}

func TestAuditEventRepository_GetByID_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	t.Run("retrieves existing event", func(t *testing.T) {
		event := createTestAuditEvent()

		err := repo.Create(ctx, event)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, event.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, event.ID, retrieved.ID)
		assert.Equal(t, event.Action, retrieved.Action)
	})

	t.Run("returns error for non-existent event", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		retrieved, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, entity.ErrAuditEventNotFound, err)
	})
}

func TestAuditEventRepository_List_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	// Create test data
	userID := uuidv7.New()
	for i := 0; i < 5; i++ {
		event := createTestAuditEvent()
		event.UserID = userID
		event.Action = "test.action"

		err := repo.Create(ctx, event)
		require.NoError(t, err)
	}

	t.Run("lists all events with pagination", func(t *testing.T) {
		filters := repository.AuditEventFilters{}
		params := pagination.Params{Page: 1, PageSize: 10, Limit: 10}

		events, meta, err := repo.List(ctx, filters, params)
		require.NoError(t, err)
		assert.NotNil(t, events)
		assert.NotNil(t, meta)
		assert.GreaterOrEqual(t, len(events), 5)
		assert.GreaterOrEqual(t, meta.Total, 5)
	})

	t.Run("filters by user_id", func(t *testing.T) {
		filters := repository.AuditEventFilters{UserID: &userID}
		params := pagination.Params{Page: 1, PageSize: 10, Limit: 10}

		events, meta, err := repo.List(ctx, filters, params)
		require.NoError(t, err)
		assert.Len(t, events, 5)
		assert.Equal(t, 5, meta.Total)

		for _, event := range events {
			assert.Equal(t, userID, event.UserID)
		}
	})

	t.Run("filters by action", func(t *testing.T) {
		action := "test.action"
		filters := repository.AuditEventFilters{Action: &action}
		params := pagination.Params{Page: 1, PageSize: 10, Limit: 10}

		events, _, err := repo.List(ctx, filters, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 5)

		for _, event := range events {
			assert.Equal(t, action, event.Action)
		}
	})

	t.Run("paginates results correctly", func(t *testing.T) {
		filters := repository.AuditEventFilters{UserID: &userID}
		params := pagination.Params{Page: 1, PageSize: 2, Limit: 2}

		events, meta, err := repo.List(ctx, filters, params)
		require.NoError(t, err)
		assert.Len(t, events, 2)
		assert.Equal(t, 5, meta.Total)
		assert.Equal(t, 1, meta.CurrentPage)
	})

	t.Run("empty result for non-matching filter", func(t *testing.T) {
		nonExistentID := uuidv7.New()
		filters := repository.AuditEventFilters{UserID: &nonExistentID}
		params := pagination.Params{Page: 1, PageSize: 10, Limit: 10}

		events, meta, err := repo.List(ctx, filters, params)
		require.NoError(t, err)
		assert.Empty(t, events)
		assert.Equal(t, 0, meta.Total)
	})
}

func TestAuditEventRepository_GetByEntityID_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	// Create events for the entity
	entityID := uuidv7.New()
	entityType := "post"
	for i := 0; i < 3; i++ {
		event := createTestAuditEvent()
		event.EntityType = entityType
		event.EntityID = entityID

		err := repo.Create(ctx, event)
		require.NoError(t, err)
	}

	t.Run("retrieves all events for entity", func(t *testing.T) {
		events, err := repo.GetByEntityID(ctx, entityType, entityID)
		require.NoError(t, err)
		assert.Len(t, events, 3)

		for _, event := range events {
			assert.Equal(t, entityType, event.EntityType)
			assert.Equal(t, entityID, event.EntityID)
		}
	})

	t.Run("returns empty for non-existent entity", func(t *testing.T) {
		nonExistentID := uuidv7.New()

		events, err := repo.GetByEntityID(ctx, entityType, nonExistentID)
		require.NoError(t, err)
		assert.Empty(t, events)
	})

	t.Run("ordered by created_at desc", func(t *testing.T) {
		events, err := repo.GetByEntityID(ctx, entityType, entityID)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(events), 2)

		// Check descending order
		for i := 0; i < len(events)-1; i++ {
			assert.True(t, events[i].CreatedAt.After(events[i+1].CreatedAt) ||
				events[i].CreatedAt.Equal(events[i+1].CreatedAt))
		}
	})
}

func TestAuditEventRepository_GetByUserID_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	// Create events for user
	userID := uuidv7.New()
	for i := 0; i < 4; i++ {
		event := createTestAuditEvent()
		event.UserID = userID

		err := repo.Create(ctx, event)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	t.Run("retrieves user events with pagination", func(t *testing.T) {
		params := pagination.Params{Page: 1, PageSize: 10, Limit: 10}

		events, meta, err := repo.GetByUserID(ctx, userID, params)
		require.NoError(t, err)
		assert.Len(t, events, 4)
		assert.Equal(t, 4, meta.Total)

		for _, event := range events {
			assert.Equal(t, userID, event.UserID)
		}
	})

	t.Run("paginates user events", func(t *testing.T) {
		params := pagination.Params{Page: 1, PageSize: 2, Limit: 2}

		events, meta, err := repo.GetByUserID(ctx, userID, params)
		require.NoError(t, err)
		assert.Len(t, events, 2)
		assert.Equal(t, 4, meta.Total)
	})
}

func TestAuditEventRepository_GetByAction_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	// Create events with action
	action := "user.login"
	for i := 0; i < 3; i++ {
		event := createTestAuditEvent()
		event.Action = action

		err := repo.Create(ctx, event)
		require.NoError(t, err)
	}

	t.Run("retrieves events by action", func(t *testing.T) {
		params := pagination.Params{Page: 1, PageSize: 10, Limit: 10}

		events, meta, err := repo.GetByAction(ctx, action, params)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(events), 3)
		assert.GreaterOrEqual(t, meta.Total, 3)

		for _, event := range events {
			assert.Equal(t, action, event.Action)
		}
	})
}

func TestAuditEventRepository_DeleteOlderThan_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	cutoffTime := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago

	// Create old events
	oldEvent := createTestAuditEvent()
	oldEvent.CreatedAt = cutoffTime.Add(-1 * time.Hour)
	err := repo.Create(ctx, oldEvent)
	require.NoError(t, err)

	// Create recent events
	recentEvent := createTestAuditEvent()
	recentEvent.CreatedAt = time.Now()
	err = repo.Create(ctx, recentEvent)
	require.NoError(t, err)

	t.Run("deletes old events", func(t *testing.T) {
		deleted, err := repo.DeleteOlderThan(ctx, cutoffTime)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))

		// Verify old event is deleted
		_, err = repo.GetByID(ctx, oldEvent.ID)
		assert.Error(t, err)
		assert.Equal(t, entity.ErrAuditEventNotFound, err)

		// Verify recent event still exists
		retrieved, err := repo.GetByID(ctx, recentEvent.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	t.Run("returns zero when no old events", func(t *testing.T) {
		veryOldTime := time.Now().Add(-365 * 24 * time.Hour)

		deleted, err := repo.DeleteOlderThan(ctx, veryOldTime)
		require.NoError(t, err)
		assert.Equal(t, int64(0), deleted)
	})
}

func TestAuditEventRepository_VerifySignature_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	t.Run("verifies valid signature", func(t *testing.T) {
		event := createTestAuditEvent()
		secret := "test-secret"
		event.Signature = GenerateSignature(event, secret)

		err := repo.Create(ctx, event)
		require.NoError(t, err)

		valid, err := repo.VerifySignature(ctx, event, secret)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("detects invalid signature", func(t *testing.T) {
		event := createTestAuditEvent()
		secret := "test-secret"
		event.Signature = "invalid-signature"

		err := repo.Create(ctx, event)
		require.NoError(t, err)

		valid, err := repo.VerifySignature(ctx, event, secret)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("detects tampered data", func(t *testing.T) {
		event := createTestAuditEvent()
		secret := "test-secret"
		event.Signature = GenerateSignature(event, secret)

		err := repo.Create(ctx, event)
		require.NoError(t, err)

		// Tamper with event data
		event.Action = "tampered.action"

		valid, err := repo.VerifySignature(ctx, event, secret)
		require.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestAuditEventRepository_ConcurrentCreates_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAuditEventRepository(db)
	ctx := context.Background()

	t.Run("handles concurrent creates", func(t *testing.T) {
		concurrency := 10
		done := make(chan bool, concurrency)
		errors := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				event := createTestAuditEvent()
				err := repo.Create(ctx, event)
				if err != nil {
					errors <- err
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < concurrency; i++ {
			<-done
		}
		close(errors)

		// Check no errors occurred
		var errCount int
		for range errors {
			errCount++
		}
		assert.Equal(t, 0, errCount, "Expected no errors in concurrent creates")
	})
}
