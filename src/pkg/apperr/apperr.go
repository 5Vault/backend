// Package apperr defines typed application errors that carry an HTTP status
// code and a safe client-facing message distinct from the internal error.
package apperr

import (
	"errors"
	"net/http"
)

// AppError is a sentinel error type recognised by the respond package.
type AppError struct {
	Code    int    // HTTP status code
	Message string // safe message sent to the client
	Cause   error  // internal error (never sent to client)
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

// Constructors

func BadRequest(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: msg, Cause: first(cause)}
}

func Unauthorized(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg, Cause: first(cause)}
}

func Forbidden(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg, Cause: first(cause)}
}

func NotFound(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, Cause: first(cause)}
}

func Conflict(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Cause: first(cause)}
}

func Internal(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg, Cause: first(cause)}
}

func TooManyRequests(msg string, cause ...error) *AppError {
	return &AppError{Code: http.StatusTooManyRequests, Message: msg, Cause: first(cause)}
}

// As unwraps err to *AppError. Returns nil if not an AppError.
func As(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}

func NewAppError(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func first(errs []error) error {
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

