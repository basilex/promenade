package entity

import "errors"

var (
	// ErrInvalidInput indicates that the input data is invalid
	ErrInvalidInput = errors.New("invalid input")

	// ErrNotFound indicates that the requested resource was not found
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists indicates that the resource already exists
	ErrAlreadyExists = errors.New("already exists")

	// ErrUnauthorized indicates that the user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates that the operation is forbidden
	ErrForbidden = errors.New("forbidden")
)
