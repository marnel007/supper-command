package errors

import (
	"fmt"
)

// ErrorType represents different types of errors
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeSecurity
	ErrorTypeExecution
	ErrorTypeConfiguration
	ErrorTypeNetwork
	ErrorTypePermission
	ErrorTypeInternal
)

// SuperShellError represents a structured error with context
type SuperShellError struct {
	Type        ErrorType
	Message     string
	Cause       error
	Context     map[string]interface{}
	Recoverable bool
}

// Error implements the error interface
func (e *SuperShellError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *SuperShellError) Unwrap() error {
	return e.Cause
}

// IsRecoverable returns whether the error is recoverable
func (e *SuperShellError) IsRecoverable() bool {
	return e.Recoverable
}

// GetType returns the error type
func (e *SuperShellError) GetType() ErrorType {
	return e.Type
}

// GetContext returns the error context
func (e *SuperShellError) GetContext() map[string]interface{} {
	return e.Context
}

// NewValidationError creates a new validation error
func NewValidationError(format string, args ...interface{}) *SuperShellError {
	return &SuperShellError{
		Type:        ErrorTypeValidation,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: true,
		Context:     make(map[string]interface{}),
	}
}

// NewSecurityError creates a new security error
func NewSecurityError(format string, args ...interface{}) *SuperShellError {
	return &SuperShellError{
		Type:        ErrorTypeSecurity,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: false,
		Context:     make(map[string]interface{}),
	}
}

// NewExecutionError creates a new execution error
func NewExecutionError(format string, args ...interface{}) *SuperShellError {
	return &SuperShellError{
		Type:        ErrorTypeExecution,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: true,
		Context:     make(map[string]interface{}),
	}
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(format string, args ...interface{}) *SuperShellError {
	return &SuperShellError{
		Type:        ErrorTypeConfiguration,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: false,
		Context:     make(map[string]interface{}),
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(format string, args ...interface{}) *SuperShellError {
	return &SuperShellError{
		Type:        ErrorTypeNetwork,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: true,
		Context:     make(map[string]interface{}),
	}
}

// NewPermissionError creates a new permission error
func NewPermissionError(format string, args ...interface{}) *SuperShellError {
	return &SuperShellError{
		Type:        ErrorTypePermission,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: false,
		Context:     make(map[string]interface{}),
	}
}

// NewInternalError creates a new internal error
func NewInternalError(format string, args ...interface{}) *SuperShellError {
	return &SuperShellError{
		Type:        ErrorTypeInternal,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: false,
		Context:     make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, format string, args ...interface{}) *SuperShellError {
	message := fmt.Sprintf(format, args...)

	// If it's already a SuperShellError, preserve the type
	if superErr, ok := err.(*SuperShellError); ok {
		return &SuperShellError{
			Type:        superErr.Type,
			Message:     message,
			Cause:       err,
			Context:     superErr.Context,
			Recoverable: superErr.Recoverable,
		}
	}

	// Otherwise, create a generic internal error
	return &SuperShellError{
		Type:        ErrorTypeInternal,
		Message:     message,
		Cause:       err,
		Context:     make(map[string]interface{}),
		Recoverable: true,
	}
}

// WithContext adds context to an error
func WithContext(err error, key string, value interface{}) *SuperShellError {
	if superErr, ok := err.(*SuperShellError); ok {
		if superErr.Context == nil {
			superErr.Context = make(map[string]interface{})
		}
		superErr.Context[key] = value
		return superErr
	}

	// Create new error with context
	newErr := &SuperShellError{
		Type:        ErrorTypeInternal,
		Message:     err.Error(),
		Cause:       err,
		Context:     map[string]interface{}{key: value},
		Recoverable: true,
	}

	return newErr
}
