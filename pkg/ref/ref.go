// Package ref provides utility functions for creating references (pointers) to values.
// Useful for optional fields in structs and database operations.
package ref

// To returns a pointer to the given value.
// Generic function works with any type.
//
// Example:
//
//	age := ref.To(25)        // *int
//	name := ref.To("John")   // *string
//	price := ref.To(19.99)   // *float64
func To[T any](v T) *T {
	return &v
}

// Int64 returns a pointer to an int64 value.
// Helper to avoid type inference issues with numeric literals.
//
// Example:
//
//	population := ref.Int64(331002651)  // *int64
func Int64(v int64) *int64 {
	return &v
}

// Int returns a pointer to an int value.
//
// Example:
//
//	count := ref.Int(42)  // *int
func Int(v int) *int {
	return &v
}

// Float64 returns a pointer to a float64 value.
//
// Example:
//
//	lat := ref.Float64(38.8951)  // *float64
func Float64(v float64) *float64 {
	return &v
}
