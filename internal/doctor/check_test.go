package doctor

import (
	"testing"
)

func TestCheckResult(t *testing.T) {
	result := CheckResult{
		Name:        "Test check",
		Status:      StatusPass,
		Message:     "Everything is fine",
		Remediation: "",
		Fixable:     false,
	}

	if result.Name != "Test check" {
		t.Errorf("expected name 'Test check', got '%s'", result.Name)
	}

	if result.Status != StatusPass {
		t.Errorf("expected status StatusPass, got %s", result.Status)
	}
}

func TestStatusValues(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusPass, "pass"},
		{StatusWarn, "warn"},
		{StatusFail, "fail"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected status %s, got %s", tt.expected, tt.status)
		}
	}
}
