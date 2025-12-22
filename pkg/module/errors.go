package module

import "errors"

// Common errors
var (
	ErrModuleNotFound        = errors.New("module not found")
	ErrModuleAlreadyExists   = errors.New("module already exists")
	ErrModuleNotInitialized  = errors.New("module not initialized")
	ErrCircularDependency    = errors.New("circular dependency detected")
	ErrMissingDependency     = errors.New("missing dependency")
	ErrInvalidConfiguration  = errors.New("invalid module configuration")
)
