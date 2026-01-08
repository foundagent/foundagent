package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	fagerrors "github.com/foundagent/foundagent/internal/errors"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "simple map",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		},
		{
			name: "nested structure",
			data: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name: "array",
			data: []string{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				err := PrintJSON(tt.data)
				if err != nil {
					t.Errorf("PrintJSON() error = %v", err)
				}
			})

			// Verify output is valid JSON
			var result interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("PrintJSON() produced invalid JSON: %v", err)
			}
		})
	}
}

func TestPrintSuccess(t *testing.T) {
	data := map[string]interface{}{
		"workspace": "test-workspace",
		"repos":     []string{"repo1", "repo2"},
		"count":     2,
	}

	output := captureStdout(func() {
		err := PrintSuccess(data)
		if err != nil {
			t.Errorf("PrintSuccess() error = %v", err)
		}
	})

	// Parse the output
	var result SuccessResponse
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse output: %v", err)
	}

	// Verify structure
	if result.Status != "success" {
		t.Errorf("Status = %q, want 'success'", result.Status)
	}

	if result.Data["workspace"] != "test-workspace" {
		t.Errorf("Data.workspace = %v, want 'test-workspace'", result.Data["workspace"])
	}

	if result.Data["count"] != float64(2) { // JSON numbers are float64
		t.Errorf("Data.count = %v, want 2", result.Data["count"])
	}
}

func TestPrintError_WithFoundagentError(t *testing.T) {
	err := fagerrors.New(
		fagerrors.ErrCodeWorkspaceExists,
		"workspace already exists",
		"Use a different name or remove existing workspace",
	)

	output := captureStdout(func() {
		printErr := PrintError(err)
		if printErr != nil {
			t.Errorf("PrintError() error = %v", printErr)
		}
	})

	// Parse the output
	var result ErrorResponse
	if parseErr := json.Unmarshal([]byte(output), &result); parseErr != nil {
		t.Fatalf("Failed to parse output: %v", parseErr)
	}

	// Verify structure
	if result.Status != "error" {
		t.Errorf("Status = %q, want 'error'", result.Status)
	}
	if result.Code != fagerrors.ErrCodeWorkspaceExists {
		t.Errorf("Code = %q, want %q", result.Code, fagerrors.ErrCodeWorkspaceExists)
	}
	if result.Message != "workspace already exists" {
		t.Errorf("Message = %q, want 'workspace already exists'", result.Message)
	}
	if result.Remediation != "Use a different name or remove existing workspace" {
		t.Errorf("Remediation = %q, want remediation message", result.Remediation)
	}
}

func TestPrintError_WithStandardError(t *testing.T) {
	err := errors.New("standard error message")

	output := captureStdout(func() {
		printErr := PrintError(err)
		if printErr != nil {
			t.Errorf("PrintError() error = %v", printErr)
		}
	})

	// Parse the output
	var result ErrorResponse
	if parseErr := json.Unmarshal([]byte(output), &result); parseErr != nil {
		t.Fatalf("Failed to parse output: %v", parseErr)
	}

	// Verify structure
	if result.Status != "error" {
		t.Errorf("Status = %q, want 'error'", result.Status)
	}
	if result.Code != fagerrors.ErrCodeUnknown {
		t.Errorf("Code = %q, want %q", result.Code, fagerrors.ErrCodeUnknown)
	}
	if result.Message != "standard error message" {
		t.Errorf("Message = %q, want 'standard error message'", result.Message)
	}
	if result.Error != "standard error message" {
		t.Errorf("Error = %q, want 'standard error message'", result.Error)
	}
}

func TestPrintError_WithWrappedError(t *testing.T) {
	cause := errors.New("underlying cause")
	err := fagerrors.Wrap(
		fagerrors.ErrCodeGitOperationFailed,
		"failed to clone repository",
		"Check network connection and credentials",
		cause,
	)

	output := captureStdout(func() {
		printErr := PrintError(err)
		if printErr != nil {
			t.Errorf("PrintError() error = %v", printErr)
		}
	})

	// Parse the output
	var result ErrorResponse
	if parseErr := json.Unmarshal([]byte(output), &result); parseErr != nil {
		t.Fatalf("Failed to parse output: %v", parseErr)
	}

	// Verify the error message includes the cause
	if !strings.Contains(result.Error, "underlying cause") {
		t.Errorf("Error should contain underlying cause, got %q", result.Error)
	}
}

func TestPrintMessage(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple message",
			format:   "Hello, World!",
			args:     nil,
			expected: "Hello, World!\n",
		},
		{
			name:     "formatted message",
			format:   "Created workspace: %s",
			args:     []interface{}{"test-workspace"},
			expected: "Created workspace: test-workspace\n",
		},
		{
			name:     "multiple args",
			format:   "Added %d repos to %s",
			args:     []interface{}{3, "workspace"},
			expected: "Added 3 repos to workspace\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				PrintMessage(tt.format, tt.args...)
			})

			if output != tt.expected {
				t.Errorf("PrintMessage() = %q, want %q", output, tt.expected)
			}
		})
	}
}

func TestPrintErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple error",
			format:   "Error: operation failed",
			args:     nil,
			expected: "Error: operation failed\n",
		},
		{
			name:     "formatted error",
			format:   "Error: %s not found",
			args:     []interface{}{"workspace"},
			expected: "Error: workspace not found\n",
		},
		{
			name:     "multiple args",
			format:   "Error: expected %d, got %d",
			args:     []interface{}{5, 3},
			expected: "Error: expected 5, got 3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStderr(func() {
				PrintErrorMessage(tt.format, tt.args...)
			})

			if output != tt.expected {
				t.Errorf("PrintErrorMessage() = %q, want %q", output, tt.expected)
			}
		})
	}
}

func TestSuccessResponse_JSONMarshaling(t *testing.T) {
	resp := SuccessResponse{
		Status: "success",
		Data: map[string]interface{}{
			"test": "value",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result SuccessResponse
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result.Status != resp.Status {
		t.Errorf("Status = %q, want %q", result.Status, resp.Status)
	}
}

func TestErrorResponse_JSONMarshaling(t *testing.T) {
	resp := ErrorResponse{
		Status:      "error",
		Error:       "test error",
		Code:        "E001",
		Message:     "test message",
		Remediation: "test remediation",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result ErrorResponse
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result.Status != resp.Status {
		t.Errorf("Status = %q, want %q", result.Status, resp.Status)
	}
	if result.Code != resp.Code {
		t.Errorf("Code = %q, want %q", result.Code, resp.Code)
	}
	if result.Message != resp.Message {
		t.Errorf("Message = %q, want %q", result.Message, resp.Message)
	}
}
