package errors

import (
	"fmt"
	"net/http"
)

type ErrorType string

const (
	ErrorTypeNotFound          ErrorType = "NOT_FOUND"
	ErrorTypeValidation        ErrorType = "VALIDATION_FAILED"
	ErrorTypeConflict          ErrorType = "CONFLICT"
	ErrorTypeUnauthorized      ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden         ErrorType = "FORBIDDEN"
	ErrorTypeInternal          ErrorType = "INTERNAL_SERVER_ERROR"
	ErrorTypePreconditionFailed ErrorType = "PRECONDITION_FAILED"
)

// AppError represents a structured, domain-centric error contract.
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Type       ErrorType         `json:"type"`
	HTTPStatus int               `json:"-"`
	Details    map[string]string `json:"details,omitempty"`
}

func (e AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:       "RESOURCE_NOT_FOUND",
		Message:    message,
		Type:       ErrorTypeNotFound,
		HTTPStatus: http.StatusNotFound,
	}
}

func NewValidationError(message string, details map[string]string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		Type:       ErrorTypeValidation,
		HTTPStatus: http.StatusBadRequest,
		Details:    details,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    message,
		Type:       ErrorTypeConflict,
		HTTPStatus: http.StatusConflict,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		Type:       ErrorTypeUnauthorized,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    message,
		Type:       ErrorTypeForbidden,
		HTTPStatus: http.StatusForbidden,
	}
}

func NewInternalError(message string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		Type:       ErrorTypeInternal,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewPreconditionFailedError(message string) *AppError {
	return &AppError{
		Code:       "PRECONDITION_FAILED",
		Message:    message,
		Type:       ErrorTypePreconditionFailed,
		HTTPStatus: http.StatusPreconditionFailed,
	}
}
