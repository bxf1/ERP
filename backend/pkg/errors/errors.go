package errors

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicate    = errors.New("resource already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrValidation   = errors.New("validation failed")
)
