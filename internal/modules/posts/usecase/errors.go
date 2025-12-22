package usecase

import "errors"

var (
	ErrSlugAlreadyExists = errors.New("post with this slug already exists")
	ErrInvalidStatus     = errors.New("invalid post status")
	ErrUnauthorized      = errors.New("unauthorized access to post")
)
