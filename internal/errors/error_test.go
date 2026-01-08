package errors

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error without cause",
			err: &Error{
				Code:    ErrCodeInvalidName,
				Message: "invalid workspace name",
			},
			expected: "[E002] invalid workspace name",
		},
		{
			name: "error with cause",
			err: &Error{
				Code:    ErrCodeGitOperationFailed,
				Message: "failed to clone repository",
				Cause:   errors.New("connection timeout"),
			},
			expected: "[E202] failed to clone repository: connection timeout",
		},
		{
			name: "error with empty message",
			err: &Error{
				Code:    ErrCodeUnknown,
				Message: "",
			},
			expected: "[E999] ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &Error{
		Code:    ErrCodeInvalidConfig,
		Message: "config error",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test error without cause
	errNoCause := &Error{
		Code:    ErrCodeInvalidConfig,
		Message: "config error",
	}
	if errNoCause.Unwrap() != nil {
		t.Errorf("Unwrap() for error without cause should return nil")
	}
}

func TestNew(t *testing.T) {
	code := ErrCodeWorkspaceExists
	message := "workspace already exists"
	remediation := "Use a different name or remove existing workspace"

	err := New(code, message, remediation)

	if err.Code != code {
		t.Errorf("Code = %q, want %q", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("Message = %q, want %q", err.Message, message)
	}
	if err.Remediation != remediation {
		t.Errorf("Remediation = %q, want %q", err.Remediation, remediation)
	}
	if err.Cause != nil {
		t.Errorf("Cause should be nil for new error")
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("original error")
	code := ErrCodeGitOperationFailed
	message := "failed to perform git operation"
	remediation := "Check git installation and repository state"

	err := Wrap(code, message, remediation, cause)

	if err.Code != code {
		t.Errorf("Code = %q, want %q", err.Code, code)
	}
	if err.Message != message {
		t.Errorf("Message = %q, want %q", err.Message, message)
	}
	if err.Remediation != remediation {
		t.Errorf("Remediation = %q, want %q", err.Remediation, remediation)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestWrap_ErrorChain(t *testing.T) {
	originalErr := errors.New("network timeout")
	wrappedErr := Wrap(ErrCodeNetworkError, "network operation failed", "Check connection", originalErr)

	// Test error chain using errors.Is
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("errors.Is should find the original error in the chain")
	}

	// Test error message includes cause
	expected := "[E401] network operation failed: network timeout"
	if wrappedErr.Error() != expected {
		t.Errorf("Error() = %q, want %q", wrappedErr.Error(), expected)
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are unique and follow expected patterns
	codes := map[string]bool{
		ErrCodeWorkspaceExists:      true,
		ErrCodeInvalidName:          true,
		ErrCodePathTooLong:          true,
		ErrCodeInvalidConfig:        true,
		ErrCodeConfigNotFound:       true,
		ErrCodeInvalidInput:         true,
		ErrCodePermissionDenied:     true,
		ErrCodeDiskFull:             true,
		ErrCodeFileNotFound:         true,
		ErrCodeDirectoryNotEmpty:    true,
		ErrCodeGitNotInstalled:      true,
		ErrCodeGitOperationFailed:   true,
		ErrCodeInvalidRepository:    true,
		ErrCodeWorktreeExists:       true,
		ErrCodeWorktreeNotFound:     true,
		ErrCodeBranchExists:         true,
		ErrCodeBranchNotFound:       true,
		ErrCodeInvalidOperation:     true,
		ErrCodeNetworkError:         true,
		ErrCodeAuthenticationFailed: true,
		ErrCodeUnknown:              true,
	}

	// Verify count matches expectations
	if len(codes) != 21 {
		t.Errorf("Expected 21 unique error codes, got %d", len(codes))
	}

	// Verify specific code values
	if ErrCodeWorkspaceExists != "E001" {
		t.Errorf("ErrCodeWorkspaceExists = %q, want E001", ErrCodeWorkspaceExists)
	}
	if ErrCodeUnknown != "E999" {
		t.Errorf("ErrCodeUnknown = %q, want E999", ErrCodeUnknown)
	}
}

func TestError_AsInterface(t *testing.T) {
	// Test that Error implements the error interface
	var _ error = &Error{}
	var _ error = New("TEST", "test message", "test remediation")
	var _ error = Wrap("TEST", "test message", "test remediation", errors.New("cause"))
}
