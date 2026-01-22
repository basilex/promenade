package main

// This file is reserved for future repository factories that don't belong to specific contexts.
//
// For bounded context repositories, factories should be placed in the repository layer:
// Example: internal/contexts/shared/country/repository/factory.go
//
// This follows Clean Architecture principles:
// - Domain layer owns the factory logic
// - Infrastructure (cmd/api) just orchestrates
// - Easier to test and reuse across different entry points (CLI, tests, etc.)
