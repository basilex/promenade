// Package aggregate provides base types for Domain-Driven Design aggregates.
// An aggregate is a cluster of domain objects that can be treated as a single unit.
// An aggregate root is the entry point to the aggregate and ensures consistency.
package aggregate

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Root is the interface that all aggregate roots must implement.
// Aggregate roots enforce business invariants and manage the lifecycle
// of entities within their boundary.
type Root interface {
	// GetID returns the aggregate's unique identifier
	GetID() uuidv7.UUID

	// GetVersion returns the aggregate's version (for optimistic locking)
	GetVersion() int

	// IncrementVersion increments the version (called after successful persistence)
	IncrementVersion()
}

// BaseAggregate provides common fields for all aggregates.
// Embed this in your aggregate roots to get standard fields and behavior.
//
// Example:
//
//	type Customer struct {
//	    aggregate.BaseAggregate
//	    Name  string
//	    Email string
//	}
type BaseAggregate struct {
	ID        uuidv7.UUID `db:"id" json:"id"`
	Version   int         `db:"version" json:"version"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
}

// GetID returns the aggregate's unique identifier
func (a *BaseAggregate) GetID() uuidv7.UUID {
	return a.ID
}

// GetVersion returns the aggregate's version (for optimistic locking)
func (a *BaseAggregate) GetVersion() int {
	return a.Version
}

// IncrementVersion increments the version (called after successful persistence)
// Also updates the UpdatedAt timestamp.
func (a *BaseAggregate) IncrementVersion() {
	a.Version++
	a.UpdatedAt = time.Now()
}

// Touch updates the UpdatedAt timestamp to the current time.
// Call this method whenever the aggregate is modified to maintain accurate timestamps.
// This replaces the need for database triggers.
//
// Example:
//
//	func (c *Customer) UpdateName(name string) error {
//	    c.Name = name
//	    c.Touch() // Update timestamp
//	    return nil
//	}
func (a *BaseAggregate) Touch() {
	a.UpdatedAt = time.Now()
}

// SetCreatedAt sets the CreatedAt timestamp.
// Useful when restoring aggregates from the database or for testing.
func (a *BaseAggregate) SetCreatedAt(t time.Time) {
	a.CreatedAt = t
}

// SetUpdatedAt sets the UpdatedAt timestamp.
// Useful when restoring aggregates from the database or for testing.
func (a *BaseAggregate) SetUpdatedAt(t time.Time) {
	a.UpdatedAt = t
}

// GetCreatedAt returns the creation timestamp.
func (a *BaseAggregate) GetCreatedAt() time.Time {
	return a.CreatedAt
}

// GetUpdatedAt returns the last update timestamp.
func (a *BaseAggregate) GetUpdatedAt() time.Time {
	return a.UpdatedAt
}

// NewBaseAggregate creates a new base aggregate with generated UUID v7 ID.
// The version is initialized to 1, and both CreatedAt and UpdatedAt are set to now.
func NewBaseAggregate() BaseAggregate {
	now := time.Now()
	return BaseAggregate{
		ID:        uuidv7.New(),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewBaseAggregateWithID creates a new base aggregate with a specific ID.
// Useful for testing or when migrating existing data.
func NewBaseAggregateWithID(id uuidv7.UUID) BaseAggregate {
	now := time.Now()
	return BaseAggregate{
		ID:        id,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
