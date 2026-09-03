package errors

import "fmt"

type ErrorCode string

const (
	InvalidRequest  ErrorCode = "INVALID_REQUEST"
	Unauthorized    ErrorCode = "UNAUTHORIZED"
	Forbidden       ErrorCode = "FORBIDDEN"
	NotFound        ErrorCode = "NOT_FOUND"
	DatabaseError   ErrorCode = "DATABASE_ERROR"
	CacheError      ErrorCode = "CACHE_ERROR"
	ValidationError ErrorCode = "VALIDATION_ERROR"
	InternalError   ErrorCode = "INTERNAL_ERROR"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Helper functions for common errors

func NewInvalidRequest(message string) *AppError {
	return New(InvalidRequest, message, nil)
}

func NewUnauthorized(message string) *AppError {
	return New(Unauthorized, message, nil)
}

func NewNotFound(message string) *AppError {
	return New(NotFound, message, nil)
}

func NewInternal(err error) *AppError {
	return New(InternalError, "An unexpected internal error occurred", err)
}
