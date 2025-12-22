package purge

import (
	"context"
	"sync"
	"time"
)

// Handler defines the interface for module-specific purge operations
type Handler interface {
	// EntityName returns the entity name that this handler purges
	EntityName() string

	// Purge executes the purge operation for soft-deleted records older than cutoffDate
	// Returns the number of records purged and any error
	Purge(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error)
}

// Registry manages purge handlers from different modules
type Registry struct {
	handlers map[string]Handler
	mu       sync.RWMutex
}

// NewRegistry creates a new purge handler registry
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]Handler),
	}
}

// Register adds a purge handler for a specific entity
func (r *Registry) Register(handler Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entityName := handler.EntityName()
	if _, exists := r.handlers[entityName]; exists {
		// Allow re-registration (useful for hot reload scenarios)
		// log.Warn("Purge handler already registered, replacing", "entity", entityName)
	}

	r.handlers[entityName] = handler
	return nil
}

// Get returns a purge handler for the given entity name
func (r *Registry) Get(entityName string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, ok := r.handlers[entityName]
	return handler, ok
}

// List returns all registered entity names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := make([]string, 0, len(r.handlers))
	for entityName := range r.handlers {
		entities = append(entities, entityName)
	}
	return entities
}

// DefaultRegistry is the global purge handler registry
var DefaultRegistry = NewRegistry()

// PolicyRegistry manages retention policies from modules
type PolicyRegistry struct {
	policies map[string]RetentionPolicy
	mu       sync.RWMutex
}

// RetentionPolicy defines how long soft-deleted records should be kept
type RetentionPolicy struct {
	EntityName    string
	RetentionDays int
	Enabled       bool
}

// NewPolicyRegistry creates a new retention policy registry
func NewPolicyRegistry() *PolicyRegistry {
	return &PolicyRegistry{
		policies: make(map[string]RetentionPolicy),
	}
}

// RegisterPolicy adds a retention policy for a specific entity
func (r *PolicyRegistry) RegisterPolicy(policy RetentionPolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if policy.RetentionDays <= 0 {
		return nil // Skip invalid policies silently
	}

	r.policies[policy.EntityName] = policy
	return nil
}

// GetPolicy returns a retention policy for the given entity name
func (r *PolicyRegistry) GetPolicy(entityName string) (RetentionPolicy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policy, ok := r.policies[entityName]
	return policy, ok
}

// GetAllPolicies returns all registered retention policies
func (r *PolicyRegistry) GetAllPolicies() []RetentionPolicy {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policies := make([]RetentionPolicy, 0, len(r.policies))
	for _, policy := range r.policies {
		policies = append(policies, policy)
	}
	return policies
}

// DefaultPolicyRegistry is the global retention policy registry
var DefaultPolicyRegistry = NewPolicyRegistry()
