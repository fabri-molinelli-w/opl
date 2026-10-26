package domain

import "errors"

// Sentinel errors used across the service layer. Handlers translate them to
// HTTP status codes in a single place (see internal/api/handlers/errors.go).
var (
	ErrValidation      = errors.New("validation error")
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
)

// FieldError describes a validation or conflict problem for a specific request
// field. Field uses the same name as the OpenAPI schema property.
type FieldError struct {
	Field string
	Issue string
}

// Error carries a user-facing message together with a domain kind. Use it when
// the service wants to expose a specific message (it never leaks internals).
type Error struct {
	Kind    error
	Message string
	Details []FieldError
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.Kind }

// NewError builds a domain error with a kind and a user-facing message.
func NewError(kind error, message string) error {
	return &Error{Kind: kind, Message: message}
}

// NewFieldError builds a domain error that also identifies the offending field.
func NewFieldError(kind error, message, field, issue string) error {
	return &Error{
		Kind:    kind,
		Message: message,
		Details: []FieldError{{Field: field, Issue: issue}},
	}
}

// FieldErrors returns the details attached to a domain error, if any.
func FieldErrors(err error) []FieldError {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr.Details
	}
	return nil
}
