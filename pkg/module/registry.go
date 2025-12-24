package module

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/basilex/promenade/pkg/logger"
)

// DefaultRegistry is the global module registry
var DefaultRegistry = NewRegistry()

// Registry manages all registered modules
type Registry struct {
	modules map[string]IModule
	mu      sync.RWMutex
}

// NewRegistry creates a new module registry
func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]IModule),
	}
}

// Register adds a module to the registry
func (r *Registry) Register(module IModule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meta := module.Metadata()
	if meta.Name == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	if _, exists := r.modules[meta.Name]; exists {
		return fmt.Errorf("module %s is already registered", meta.Name)
	}

	r.modules[meta.Name] = module
	logger.Info("IModule registered",
		"module", meta.Name,
		"version", meta.Version,
		"author", meta.Author)

	return nil
}

// Get retrieves a module by name
func (r *Registry) Get(name string) IModule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.modules[name]
}

// GetAll returns all registered modules
func (r *Registry) GetAll() []IModule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modules := make([]IModule, 0, len(r.modules))
	for _, mod := range r.modules {
		modules = append(modules, mod)
	}
	return modules
}

// GetEnabled returns modules that are enabled in configuration
func (r *Registry) GetEnabled(enabledNames []string) []IModule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	enabled := make([]IModule, 0, len(enabledNames))
	for _, name := range enabledNames {
		if mod, exists := r.modules[name]; exists {
			enabled = append(enabled, mod)
		} else {
			logger.Warn("Enabled module not found in registry", "module", name)
		}
	}
	return enabled
}

// InitializeAll initializes all enabled modules in dependency order
func (r *Registry) InitializeAll(ctx context.Context, core *Core, enabledNames []string) error {
	enabled := r.GetEnabled(enabledNames)

	// Sort modules by dependencies
	sorted, err := r.sortByDependencies(enabled)
	if err != nil {
		return fmt.Errorf("failed to resolve module dependencies: %w", err)
	}

	// Initialize modules in dependency order
	for _, mod := range sorted {
		meta := mod.Metadata()
		logger.Info("Initializing module",
			"module", meta.Name,
			"version", meta.Version)

		if err := mod.Initialize(ctx, core); err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", meta.Name, err)
		}

		logger.Info("IModule initialized successfully", "module", meta.Name)
	}

	return nil
}

// StartAll starts all enabled modules
func (r *Registry) StartAll(ctx context.Context, enabledNames []string) error {
	enabled := r.GetEnabled(enabledNames)

	for _, mod := range enabled {
		meta := mod.Metadata()
		logger.Info("Starting module", "module", meta.Name)

		if err := mod.Start(ctx); err != nil {
			return fmt.Errorf("failed to start module %s: %w", meta.Name, err)
		}

		logger.Info("IModule started successfully", "module", meta.Name)
	}

	return nil
}

// StopAll stops all enabled modules in reverse order
func (r *Registry) StopAll(ctx context.Context, enabledNames []string) error {
	enabled := r.GetEnabled(enabledNames)

	// Stop in reverse order
	for i := len(enabled) - 1; i >= 0; i-- {
		mod := enabled[i]
		meta := mod.Metadata()
		logger.Info("Stopping module", "module", meta.Name)

		if err := mod.Stop(ctx); err != nil {
			logger.Error("Failed to stop module", "module", meta.Name, "error", err)
			// Continue stopping other modules
		}
	}

	return nil
}

// sortByDependencies returns modules sorted in dependency order
func (r *Registry) sortByDependencies(modules []IModule) ([]IModule, error) {
	// Build dependency graph
	graph := make(map[string][]string)
	moduleMap := make(map[string]IModule)

	for _, mod := range modules {
		name := mod.Metadata().Name
		graph[name] = mod.Dependencies()
		moduleMap[name] = mod
	}

	// Topological sort
	sorted := []string{}
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(string) error
	visit = func(name string) error {
		if temp[name] {
			return fmt.Errorf("circular dependency detected: %s", name)
		}
		if visited[name] {
			return nil
		}

		temp[name] = true
		for _, dep := range graph[name] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		temp[name] = false
		visited[name] = true
		sorted = append(sorted, name)

		return nil
	}

	for name := range graph {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	// Convert sorted names back to modules
	result := make([]IModule, 0, len(sorted))
	for _, name := range sorted {
		result = append(result, moduleMap[name])
	}

	return result, nil
}

// HealthCheckAll checks health of all enabled modules
func (r *Registry) HealthCheckAll(ctx context.Context, enabledNames []string) map[string]error {
	enabled := r.GetEnabled(enabledNames)
	results := make(map[string]error)

	for _, mod := range enabled {
		meta := mod.Metadata()
		results[meta.Name] = mod.HealthCheck(ctx)
	}

	return results
}

// ListModules returns metadata for all registered modules
func (r *Registry) ListModules() []Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Metadata, 0, len(r.modules))
	for _, mod := range r.modules {
		list = append(list, mod.Metadata())
	}

	// Sort by name
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}
