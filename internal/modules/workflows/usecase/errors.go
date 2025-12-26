package usecase

import "errors"

// Workflow definition errors
var (
	ErrWorkflowNotFound          = errors.New("workflow definition not found")
	ErrWorkflowNameAlreadyExists = errors.New("workflow definition with this name already exists")
	ErrWorkflowNotActive         = errors.New("workflow definition is not active")
	ErrWorkflowInvalidSchema     = errors.New("invalid workflow schema")
)

// Workflow instance errors
var (
	ErrInstanceNotFound   = errors.New("workflow instance not found")
	ErrInvalidTransition  = errors.New("invalid workflow state transition")
	ErrUnauthorizedAccess = errors.New("unauthorized access to workflow")
)
