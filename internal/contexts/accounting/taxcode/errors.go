package taxcode

import "errors"

var (
// Validation errors
ErrTaxCodeEmpty      = errors.New("tax code cannot be empty")
ErrTaxNameEmpty      = errors.New("tax name cannot be empty")
ErrInvalidTaxType    = errors.New("invalid tax type")
ErrInvalidTaxRate    = errors.New("tax rate must be between 0 and 100")
ErrGLAccountRequired = errors.New("GL account is required for tax code")

// Business logic errors
ErrTaxCodeNotFound    = errors.New("tax code not found")
ErrTaxCodeInactive    = errors.New("tax code is inactive")
ErrTaxCodeDuplicate   = errors.New("tax code already exists")
ErrCannotDeleteInUse  = errors.New("cannot delete tax code in use")
ErrInvalidTaxableBase = errors.New("taxable base amount must be positive")
)
