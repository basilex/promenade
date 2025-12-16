package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators registers custom validators for Gin
func RegisterCustomValidators(v *validator.Validate) error {
	// Register uppercase validator
	if err := v.RegisterValidation("uppercase", validateUppercase); err != nil {
		return err
	}

	// Register alpha validator (letters only)
	if err := v.RegisterValidation("alpha", validateAlpha); err != nil {
		return err
	}

	// Register alphanum validator (letters and numbers)
	if err := v.RegisterValidation("alphanum", validateAlphaNum); err != nil {
		return err
	}

	// Register no special chars validator
	if err := v.RegisterValidation("no_special", validateNoSpecialChars); err != nil {
		return err
	}

	return nil
}

// validateUppercase checks if string is uppercase
func validateUppercase(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return value == strings.ToUpper(value)
}

// validateAlpha checks if string contains only letters
func validateAlpha(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^[A-Za-z]+$`, value)
	return matched
}

// validateAlphaNum checks if string contains only letters and numbers
func validateAlphaNum(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	matched, _ := regexp.MatchString(`^[A-Za-z0-9]+$`, value)
	return matched
}

// validateNoSpecialChars checks if string doesn't contain special characters
func validateNoSpecialChars(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Allow letters, numbers, spaces, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[A-Za-z0-9\s\-_]+$`, value)
	return matched
}
