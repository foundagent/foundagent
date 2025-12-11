package doctor

import (
	"fmt"
)

// Runner executes checks and collects results
type Runner struct {
	checks []Check
}

// NewRunner creates a new check runner
func NewRunner(checks []Check) *Runner {
	return &Runner{
		checks: checks,
	}
}

// Run executes all checks and returns results
func (r *Runner) Run() []CheckResult {
	results := make([]CheckResult, 0, len(r.checks))

	for _, check := range r.checks {
		result := check.Run()
		results = append(results, result)
	}

	return results
}

// Summary represents aggregated check results
type Summary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Failed   int `json:"failed"`
}

// CalculateSummary computes summary statistics from check results
func CalculateSummary(results []CheckResult) Summary {
	summary := Summary{
		Total: len(results),
	}

	for _, result := range results {
		switch result.Status {
		case StatusPass:
			summary.Passed++
		case StatusWarn:
			summary.Warnings++
		case StatusFail:
			summary.Failed++
		}
	}

	return summary
}

// FormatResults formats check results for human output
func FormatResults(results []CheckResult) string {
	output := ""

	for _, result := range results {
		icon := "✗"
		if result.Status == StatusPass {
			icon = "✓"
		} else if result.Status == StatusWarn {
			icon = "⚠"
		}

		output += fmt.Sprintf("%s %s: %s\n", icon, result.Name, result.Message)

		if result.Remediation != "" && result.Status != StatusPass {
			output += fmt.Sprintf("  → %s\n", result.Remediation)
		}
	}

	return output
}
