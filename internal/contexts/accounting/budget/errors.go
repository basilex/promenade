package budget

import "errors"

var (
	// Validation errors
	ErrBudgetNameEmpty      = errors.New("budget name cannot be empty")
	ErrInvalidBudgetPeriod  = errors.New("budget period must be valid")
	ErrInvalidAmount        = errors.New("budget amount must be non-negative")
	ErrLineAccountRequired  = errors.New("account is required for budget line")
	ErrInvalidFiscalYear    = errors.New("fiscal year must be positive")
	ErrAccountRequired      = errors.New("account ID is required for budget line")
	ErrInvalidBudgetAmount  = errors.New("budget amount must be non-negative")
	ErrAccountAlreadyExists = errors.New("account already exists in this budget")

	// Business logic errors
	ErrBudgetNotFound             = errors.New("budget not found")
	ErrBudgetAlreadyApproved      = errors.New("budget is already approved")
	ErrBudgetNotApproved          = errors.New("budget must be approved before activation")
	ErrCannotModifyActive         = errors.New("cannot modify active budget")
	ErrCannotDeleteApproved       = errors.New("cannot delete approved budget")
	ErrLineNotFound               = errors.New("budget line not found")
	ErrCannotModifyApprovedBudget = errors.New("cannot modify approved or active budget")
	ErrCannotModifyActiveBudget   = errors.New("cannot modify active budget")
	ErrBudgetLineNotFound         = errors.New("budget line not found")
	ErrBudgetNotActive            = errors.New("budget is not active")
	ErrBudgetHasNoLines           = errors.New("budget must have at least one line")
)
