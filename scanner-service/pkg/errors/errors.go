package errors

import (
	"fmt"
	"net/http"
)

// Type represents an error type
type Type string

// Application error types
const (
	// ErrInternal is returned when an internal error occurs
	ErrInternal Type = "INTERNAL"

	// ErrNotFound is returned when a resource is not found
	ErrNotFound Type = "NOT_FOUND"

	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput Type = "INVALID_INPUT"

	// ErrTimeout is returned when an operation times out
	ErrTimeout Type = "TIMEOUT"

	// ErrUnavailable is returned when a service is unavailable
	ErrUnavailable Type = "UNAVAILABLE"

	// ErrUnauthorized is returned when the user is not authorized
	ErrUnauthorized Type = "UNAUTHORIZED"

	// ErrForbidden is returned when the user is forbidden from accessing a resource
	ErrForbidden Type = "FORBIDDEN"

	// ErrAlreadyExists is returned when a resource already exists
	ErrAlreadyExists Type = "ALREADY_EXISTS"
)

// Error represents an application error
type Error struct {
	Type    Type   `json:"type"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error returns the error message
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %s", e.Type, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *Error) Unwrap() error {
	return e.Err
}

// StatusCode returns the HTTP status code for the error
func (e *Error) StatusCode() int {
	switch e.Type {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrInvalidInput:
		return http.StatusBadRequest
	case ErrTimeout:
		return http.StatusGatewayTimeout
	case ErrUnavailable:
		return http.StatusServiceUnavailable
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrAlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new Error
func New(errType Type, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// NewInternal creates a new internal Error
func NewInternal(message string, err error) *Error {
	return New(ErrInternal, message, err)
}

// NewNotFound creates a new not found Error
func NewNotFound(message string, err error) *Error {
	return New(ErrNotFound, message, err)
}

// NewInvalidInput creates a new invalid input Error
func NewInvalidInput(message string, err error) *Error {
	return New(ErrInvalidInput, message, err)
}

// NewTimeout creates a new timeout Error
func NewTimeout(message string, err error) *Error {
	return New(ErrTimeout, message, err)
}

// NewUnavailable creates a new unavailable Error
func NewUnavailable(message string, err error) *Error {
	return New(ErrUnavailable, message, err)
}

// NewUnauthorized creates a new unauthorized Error
func NewUnauthorized(message string, err error) *Error {
	return New(ErrUnauthorized, message, err)
}

// NewForbidden creates a new forbidden Error
func NewForbidden(message string, err error) *Error {
	return New(ErrForbidden, message, err)
}

// NewAlreadyExists creates a new already exists Error
func NewAlreadyExists(message string, err error) *Error {
	return New(ErrAlreadyExists, message, err)
}
