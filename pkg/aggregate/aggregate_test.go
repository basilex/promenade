package aggregate_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewBaseAggregate(t *testing.T) {
	t.Run("creates aggregate with valid UUID v7", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		assert.NotEqual(t, uuidv7.Nil, agg.GetID())
		assert.True(t, uuidv7.IsV7(agg.GetID()))
	})

	t.Run("initializes version to 1", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		assert.Equal(t, 1, agg.GetVersion())
	})

	t.Run("sets CreatedAt and UpdatedAt to now", func(t *testing.T) {
		before := time.Now()
		agg := aggregate.NewBaseAggregate()
		after := time.Now()

		assert.True(t, agg.CreatedAt.After(before) || agg.CreatedAt.Equal(before))
		assert.True(t, agg.CreatedAt.Before(after) || agg.CreatedAt.Equal(after))
		assert.Equal(t, agg.CreatedAt, agg.UpdatedAt)
	})

	t.Run("generates unique IDs for multiple aggregates", func(t *testing.T) {
		agg1 := aggregate.NewBaseAggregate()
		agg2 := aggregate.NewBaseAggregate()

		assert.NotEqual(t, agg1.GetID(), agg2.GetID())
	})
}

func TestNewBaseAggregateWithID(t *testing.T) {
	t.Run("creates aggregate with specific ID", func(t *testing.T) {
		id := uuidv7.New()
		agg := aggregate.NewBaseAggregateWithID(id)

		assert.Equal(t, id, agg.GetID())
	})

	t.Run("initializes version to 1", func(t *testing.T) {
		id := uuidv7.New()
		agg := aggregate.NewBaseAggregateWithID(id)

		assert.Equal(t, 1, agg.GetVersion())
	})
}

func TestBaseAggregate_GetID(t *testing.T) {
	t.Run("returns the aggregate ID", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		id := agg.GetID()

		assert.NotEqual(t, uuidv7.Nil, id)
		assert.Equal(t, agg.ID, id)
	})
}

func TestBaseAggregate_GetVersion(t *testing.T) {
	t.Run("returns initial version", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		assert.Equal(t, 1, agg.GetVersion())
	})

	t.Run("returns updated version after increment", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		agg.IncrementVersion()

		assert.Equal(t, 2, agg.GetVersion())
	})
}

func TestBaseAggregate_IncrementVersion(t *testing.T) {
	t.Run("increments version by 1", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialVersion := agg.GetVersion()

		agg.IncrementVersion()

		assert.Equal(t, initialVersion+1, agg.GetVersion())
	})

	t.Run("updates UpdatedAt timestamp", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialUpdatedAt := agg.UpdatedAt

		time.Sleep(10 * time.Millisecond) // Ensure time passes
		agg.IncrementVersion()

		assert.True(t, agg.UpdatedAt.After(initialUpdatedAt))
	})

	t.Run("does not change CreatedAt", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialCreatedAt := agg.CreatedAt

		time.Sleep(10 * time.Millisecond)
		agg.IncrementVersion()

		assert.Equal(t, initialCreatedAt, agg.CreatedAt)
	})

	t.Run("can be called multiple times", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		for i := 1; i <= 10; i++ {
			agg.IncrementVersion()
			assert.Equal(t, i+1, agg.GetVersion())
		}

		assert.Equal(t, 11, agg.GetVersion())
	})
}

func TestBaseAggregate_OptimisticLocking(t *testing.T) {
	t.Run("version can be used for optimistic locking", func(t *testing.T) {
		// Simulate two concurrent updates
		agg := aggregate.NewBaseAggregate()
		version1 := agg.GetVersion()

		// First update succeeds
		agg.IncrementVersion()
		version2 := agg.GetVersion()

		// Second update should detect version mismatch
		assert.NotEqual(t, version1, version2)
		assert.Equal(t, version1+1, version2)
	})
}

// TestCustomAggregate demonstrates embedding BaseAggregate
func TestCustomAggregate(t *testing.T) {
	type Customer struct {
		aggregate.BaseAggregate
		Name  string
		Email string
	}

	t.Run("custom aggregate implements Root interface", func(t *testing.T) {
		customer := Customer{
			BaseAggregate: aggregate.NewBaseAggregate(),
			Name:          "John Doe",
			Email:         "john@example.com",
		}

		// Verify it implements aggregate.Root
		var _ aggregate.Root = &customer

		assert.NotEqual(t, uuidv7.Nil, customer.GetID())
		assert.Equal(t, 1, customer.GetVersion())
	})

	t.Run("custom aggregate can increment version", func(t *testing.T) {
		customer := Customer{
			BaseAggregate: aggregate.NewBaseAggregate(),
			Name:          "John Doe",
			Email:         "john@example.com",
		}

		customer.IncrementVersion()

		assert.Equal(t, 2, customer.GetVersion())
	})
}

func TestAggregate_Concurrency(t *testing.T) {
	t.Run("multiple increments are safe", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		// Note: This is not testing true concurrency, just sequential increments
		// In production, use database-level optimistic locking (version column)
		for i := 0; i < 100; i++ {
			agg.IncrementVersion()
		}

		assert.Equal(t, 101, agg.GetVersion())
	})
}

func TestAggregate_Timestamps(t *testing.T) {
	t.Run("CreatedAt remains constant", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		created := agg.CreatedAt

		time.Sleep(10 * time.Millisecond)
		agg.IncrementVersion()

		assert.Equal(t, created, agg.CreatedAt)
	})

	t.Run("UpdatedAt changes with version", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		updated1 := agg.UpdatedAt

		time.Sleep(10 * time.Millisecond)
		agg.IncrementVersion()
		updated2 := agg.UpdatedAt

		require.True(t, updated2.After(updated1), "UpdatedAt should increase")
	})

	t.Run("CreatedAt equals UpdatedAt initially", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		assert.Equal(t, agg.CreatedAt, agg.UpdatedAt)
	})
}

func TestBaseAggregate_Touch(t *testing.T) {
	t.Run("updates UpdatedAt timestamp", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialUpdatedAt := agg.UpdatedAt

		time.Sleep(10 * time.Millisecond)
		agg.Touch()

		assert.True(t, agg.UpdatedAt.After(initialUpdatedAt))
	})

	t.Run("does not change CreatedAt", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialCreatedAt := agg.CreatedAt

		time.Sleep(10 * time.Millisecond)
		agg.Touch()

		assert.Equal(t, initialCreatedAt, agg.CreatedAt)
	})

	t.Run("does not change Version", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialVersion := agg.GetVersion()

		agg.Touch()

		assert.Equal(t, initialVersion, agg.GetVersion())
	})

	t.Run("can be called multiple times", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		previousUpdatedAt := agg.UpdatedAt

		for i := 0; i < 3; i++ {
			time.Sleep(5 * time.Millisecond)
			agg.Touch()
			assert.True(t, agg.UpdatedAt.After(previousUpdatedAt))
			previousUpdatedAt = agg.UpdatedAt
		}
	})
}

func TestBaseAggregate_SetCreatedAt(t *testing.T) {
	t.Run("sets CreatedAt timestamp", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		agg.SetCreatedAt(customTime)

		assert.Equal(t, customTime, agg.CreatedAt)
	})
}

func TestBaseAggregate_SetUpdatedAt(t *testing.T) {
	t.Run("sets UpdatedAt timestamp", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		customTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		agg.SetUpdatedAt(customTime)

		assert.Equal(t, customTime, agg.UpdatedAt)
	})
}

func TestBaseAggregate_GetCreatedAt(t *testing.T) {
	t.Run("returns CreatedAt timestamp", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		createdAt := agg.GetCreatedAt()

		assert.Equal(t, agg.CreatedAt, createdAt)
	})
}

func TestBaseAggregate_GetUpdatedAt(t *testing.T) {
	t.Run("returns UpdatedAt timestamp", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()

		updatedAt := agg.GetUpdatedAt()

		assert.Equal(t, agg.UpdatedAt, updatedAt)
	})
}

// TestBaseAggregate_TouchVsIncrementVersion demonstrates the difference
func TestBaseAggregate_TouchVsIncrementVersion(t *testing.T) {
	t.Run("Touch updates timestamp only", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialVersion := agg.GetVersion()
		initialUpdatedAt := agg.UpdatedAt

		time.Sleep(10 * time.Millisecond)
		agg.Touch()

		assert.Equal(t, initialVersion, agg.GetVersion(), "Touch should not change version")
		assert.True(t, agg.UpdatedAt.After(initialUpdatedAt), "Touch should update timestamp")
	})

	t.Run("IncrementVersion updates both", func(t *testing.T) {
		agg := aggregate.NewBaseAggregate()
		initialVersion := agg.GetVersion()
		initialUpdatedAt := agg.UpdatedAt

		time.Sleep(10 * time.Millisecond)
		agg.IncrementVersion()

		assert.Equal(t, initialVersion+1, agg.GetVersion(), "IncrementVersion should change version")
		assert.True(t, agg.UpdatedAt.After(initialUpdatedAt), "IncrementVersion should update timestamp")
	})
}

// TestCustomAggregateWithTouch demonstrates typical usage
func TestCustomAggregateWithTouch(t *testing.T) {
	type Customer struct {
		aggregate.BaseAggregate
		Name  string
		Email string
	}

	t.Run("Touch in business method", func(t *testing.T) {
		customer := Customer{
			BaseAggregate: aggregate.NewBaseAggregate(),
			Name:          "John Doe",
			Email:         "john@example.com",
		}
		initialUpdatedAt := customer.UpdatedAt

		// Simulate business method
		time.Sleep(10 * time.Millisecond)
		customer.Name = "Jane Doe"
		customer.Touch() // Call after modification

		assert.Equal(t, "Jane Doe", customer.Name)
		assert.True(t, customer.UpdatedAt.After(initialUpdatedAt))
		assert.Equal(t, 1, customer.GetVersion()) // Version unchanged
	})
}
