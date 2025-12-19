// Package version provides application metadata constants.
// Centralized location for service name, version, and identifiers
// used across handlers, logging, health checks, and API responses.
package version

const (
	// ServiceName is the human-readable service name
	ServiceName = "Promenade API"

	// ServiceVersion follows semantic versioning (semver.org)
	ServiceVersion = "0.1.0"

	// ServiceID is the lowercase identifier used in logs, health checks, and monitoring
	ServiceID = "promenade"
)
