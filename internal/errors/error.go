package errors

import "fmt"

// Error represents a structured error with code and remediation
type Error struct {
	Code        string
	Message     string
	Remediation string
	Cause       error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *Error) Unwrap() error {
	return e.Cause
}

// New creates a new structured error
func New(code, message, remediation string) *Error {
	return &Error{
		Code:        code,
		Message:     message,
		Remediation: remediation,
	}
}

// Wrap wraps an existing error with context
func Wrap(code, message, remediation string, cause error) *Error {
	return &Error{
		Code:        code,
		Message:     message,
		Remediation: remediation,
		Cause:       cause,
	}
}
