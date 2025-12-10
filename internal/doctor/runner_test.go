package doctor

import (
	"testing"
)

// MockCheck is a test implementation of the Check interface
type MockCheck struct {
	name   string
	result CheckResult
}

func (m MockCheck) Name() string {
	return m.name
}

func (m MockCheck) Run() CheckResult {
	return m.result
}

func TestRunner(t *testing.T) {
	checks := []Check{
		MockCheck{
			name: "Pass check",
			result: CheckResult{
				Name:    "Pass check",
				Status:  StatusPass,
				Message: "Passed",
			},
		},
		MockCheck{
			name: "Warn check",
			result: CheckResult{
				Name:    "Warn check",
				Status:  StatusWarn,
				Message: "Warning",
			},
		},
		MockCheck{
			name: "Fail check",
			result: CheckResult{
				Name:    "Fail check",
				Status:  StatusFail,
				Message: "Failed",
			},
		},
	}

	runner := NewRunner(checks)
	results := runner.Run()

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if results[0].Status != StatusPass {
		t.Errorf("expected first check to pass")
	}

	if results[1].Status != StatusWarn {
		t.Errorf("expected second check to warn")
	}

	if results[2].Status != StatusFail {
		t.Errorf("expected third check to fail")
	}
}

func TestCalculateSummary(t *testing.T) {
	results := []CheckResult{
		{Status: StatusPass},
		{Status: StatusPass},
		{Status: StatusWarn},
		{Status: StatusFail},
	}

	summary := CalculateSummary(results)

	if summary.Total != 4 {
		t.Errorf("expected total 4, got %d", summary.Total)
	}

	if summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", summary.Passed)
	}

	if summary.Warnings != 1 {
		t.Errorf("expected 1 warning, got %d", summary.Warnings)
	}

	if summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", summary.Failed)
	}
}

func TestFormatResults(t *testing.T) {
	results := []CheckResult{
		{
			Name:    "Pass check",
			Status:  StatusPass,
			Message: "Passed",
		},
		{
			Name:        "Fail check",
			Status:      StatusFail,
			Message:     "Failed",
			Remediation: "Fix it",
		},
	}

	output := FormatResults(results)

	if output == "" {
		t.Error("expected non-empty output")
	}

	// Check for pass icon
	if !containsString(output, "✓") {
		t.Error("expected pass icon in output")
	}

	// Check for fail icon
	if !containsString(output, "✗") {
		t.Error("expected fail icon in output")
	}

	// Check for remediation
	if !containsString(output, "Fix it") {
		t.Error("expected remediation in output")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}
