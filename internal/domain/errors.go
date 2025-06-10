package domain

import "fmt"

// DomainErrorKind represents the category of a domain error
// e.g., Validation, NotFound, Conflict, etc.
type DomainErrorKind string

const (
	DomainErrorValidation DomainErrorKind = "validation"
	DomainErrorNotFound   DomainErrorKind = "not_found"
	DomainErrorConflict   DomainErrorKind = "conflict"
	DomainErrorInternal   DomainErrorKind = "internal"
)

// DomainError is a custom error type for domain logic
// It allows classification and wrapping of errors
// Implements the error interface

type DomainError struct {
	Kind    DomainErrorKind
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// Helper constructors
func NewDomainError(kind DomainErrorKind, message string, cause error) *DomainError {
	return &DomainError{
		Kind:    kind,
		Message: message,
		Cause:   cause,
	}
}

func NewValidationError(message string, cause error) *DomainError {
	return NewDomainError(DomainErrorValidation, message, cause)
}

func NewNotFoundError(message string, cause error) *DomainError {
	return NewDomainError(DomainErrorNotFound, message, cause)
}

func NewConflictError(message string, cause error) *DomainError {
	return NewDomainError(DomainErrorConflict, message, cause)
}

func NewInternalError(message string, cause error) *DomainError {
	return NewDomainError(DomainErrorInternal, message, cause)
}
