package entity

import "errors"

var (
    ErrRoleNotFound = errors.New("role not found")
    ErrProductNotFound = errors.New("product not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
)
