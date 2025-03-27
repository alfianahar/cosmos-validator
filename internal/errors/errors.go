package errors

import (
	"fmt"
	"net/http"
)

// represents application-specific errors with HTTP status codes
type AppError struct {
	Code    int    // HTTP status code
	Message string // Error message
	Err     error  // Original error if wrapped
}

// implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// allows errors.Is and errors.As to work with wrapped errors
func (e *AppError) Unwrap() error {
	return e.Err
}

// creates an error for resources not found
func NewNotFoundError(resource string, err error) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Err:     err,
	}
}

// creates an error for invalid requests
func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

// creates an error for unexpected server issues
func NewInternalServerError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Err:     err,
	}
}

// creates an error for API outages or rate limiting
func NewServiceUnavailableError(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusServiceUnavailable,
		Message: message,
		Err:     err,
	}
}
