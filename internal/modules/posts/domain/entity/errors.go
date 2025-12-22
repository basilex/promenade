package entity

import "errors"

var (
	ErrNotFound           = errors.New("post not found")
	ErrUnauthorized       = errors.New("unauthorized access to post")
	ErrInvalidInput       = errors.New("invalid input")
	ErrSlugAlreadyExists  = errors.New("post with this slug already exists")
	ErrInvalidStatus      = errors.New("invalid post status")
)
