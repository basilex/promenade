// Package uuidv7 provides UUID v7 generation per RFC 9562 draft specification.
// UUID v7 provides time-ordered UUIDs with millisecond precision, offering better
// database performance characteristics than UUID v4 (random) due to natural ordering.
//
// Benefits of UUID v7:
//
//   - Time-ordered: newer UUIDs sort after older ones
//   - Better B-tree index locality (reduces page splits in PostgreSQL)
//   - Improved INSERT performance (20-50% faster in benchmarks)
//   - Reduced index fragmentation
//   - Still globally unique like v4
//   - Can extract creation timestamp from the UUID
//
// Usage:
//
//	id := uuidv7.New()                  // Generate new UUID v7
//	ts := uuidv7.ExtractTime(id)        // Get timestamp from UUID
package uuidv7

import (
	"crypto/rand"
	"encoding/binary"
	"time"

	"github.com/google/uuid"
)

// New generates a new UUID v7 with current timestamp.
// Returns a time-ordered UUID suitable for use as database primary keys.
func New() uuid.UUID {
	return NewWithTime(time.Now())
}

// NewWithTime generates a UUID v7 with a specific timestamp.
// Useful for testing or when you need to control the timestamp.
func NewWithTime(t time.Time) uuid.UUID {
	var u uuid.UUID

	// Get Unix timestamp in milliseconds (48 bits)
	unixMs := uint64(t.UnixMilli())

	// Timestamp: first 48 bits (6 bytes)
	binary.BigEndian.PutUint32(u[0:4], uint32(unixMs>>16))
	binary.BigEndian.PutUint16(u[4:6], uint16(unixMs))

	// Random bytes for the rest
	_, _ = rand.Read(u[6:16])

	// Set version (4 bits): 0111 = version 7
	u[6] = (u[6] & 0x0f) | 0x70

	// Set variant (2 bits): 10 = RFC 4122 variant
	u[8] = (u[8] & 0x3f) | 0x80

	return u
}

// ExtractTime extracts the timestamp from a UUID v7.
// Returns zero time if the UUID is not version 7.
func ExtractTime(u uuid.UUID) time.Time {
	// Verify this is a UUID v7 (version bits should be 0111)
	if u[6]>>4 != 0x07 {
		return time.Time{}
	}

	// Extract 48-bit timestamp from first 6 bytes
	msHigh := uint64(binary.BigEndian.Uint32(u[0:4]))
	msLow := uint64(binary.BigEndian.Uint16(u[4:6]))
	unixMs := (msHigh << 16) | msLow

	return time.UnixMilli(int64(unixMs))
}

// IsV7 checks if a UUID is version 7.
func IsV7(u uuid.UUID) bool {
	return u[6]>>4 == 0x07
}

// Parse parses a string UUID and returns error on failure.
// Convenience wrapper around uuid.Parse.
func Parse(s string) (UUID, error) {
	return uuid.Parse(s)
}

// MustParse parses a string UUID and panics on error.
// Convenience wrapper around uuid.Parse.
func MustParse(s string) UUID {
	return uuid.MustParse(s)
}

// UUID is a re-export of uuid.UUID for convenience.
// This allows using uuidv7.UUID throughout the codebase instead of google/uuid.
type UUID = uuid.UUID

// Nil is the nil UUID.
var Nil = uuid.Nil
