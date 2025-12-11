package doctor

import (
	"os/exec"
	"strings"
)

// GitCheck checks if Git is installed
type GitCheck struct{}

func (c GitCheck) Name() string {
	return "Git installed"
}

func (c GitCheck) Run() CheckResult {
	_, err := exec.LookPath("git")
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Git is not installed or not in PATH",
			Remediation: "Install Git from https://git-scm.com/downloads",
			Fixable:     false,
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Git is installed",
		Fixable: false,
	}
}

// GitVersionCheck checks Git version
type GitVersionCheck struct{}

func (c GitVersionCheck) Name() string {
	return "Git version"
}

func (c GitVersionCheck) Run() CheckResult {
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Failed to get Git version",
			Remediation: "Ensure Git is properly installed",
			Fixable:     false,
		}
	}

	version := strings.TrimSpace(string(output))

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: version,
		Fixable: false,
	}
}
