package errors

import (
	"fmt"
	"net/http"
)

// Code represents a machine-readable error code.
type Code string

const (
	CodeBadRequest       Code = "BAD_REQUEST"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeForbidden        Code = "FORBIDDEN"
	CodeNotFound         Code = "NOT_FOUND"
	CodeConflict         Code = "CONFLICT"
	CodeValidationFailed Code = "VALIDATION_FAILED"
	CodeInternalError    Code = "INTERNAL_ERROR"
	CodeTenantNotFound   Code = "TENANT_NOT_FOUND"
	CodeTenantMismatch   Code = "TENANT_MISMATCH"
	CodeDatabaseError    Code = "DATABASE_ERROR"
)

// StatusCode maps error codes to HTTP status codes.
func (c Code) StatusCode() int {
	switch c {
	case CodeBadRequest, CodeValidationFailed:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound, CodeTenantNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeTenantMismatch:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// AppError is the application's standard error type.
type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError.
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap wraps an existing error.
func Wrap(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// WithDetails attaches detail payload to the error.
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// Common constructors
func BadRequest(msg string) *AppError        { return New(CodeBadRequest, msg) }
func Unauthorized(msg string) *AppError       { return New(CodeUnauthorized, msg) }
func Forbidden(msg string) *AppError          { return New(CodeForbidden, msg) }
func NotFound(msg string) *AppError           { return New(CodeNotFound, msg) }
func Conflict(msg string) *AppError           { return New(CodeConflict, msg) }
func ValidationFailed(msg string) *AppError   { return New(CodeValidationFailed, msg) }
func Internal(msg string) *AppError           { return New(CodeInternalError, msg) }
func TenantNotFound(msg string) *AppError     { return New(CodeTenantNotFound, msg) }
func DatabaseError(msg string, err error) *AppError {
	return Wrap(CodeDatabaseError, msg, err)
}
