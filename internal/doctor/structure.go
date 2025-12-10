package doctor

import (
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/workspace"
)

// WorkspaceStructureCheck checks basic workspace structure
type WorkspaceStructureCheck struct {
	Workspace *workspace.Workspace
}

func (c WorkspaceStructureCheck) Name() string {
	return "Workspace structure"
}

func (c WorkspaceStructureCheck) Run() CheckResult {
	// Check .foundagent.yaml
	configPath := c.Workspace.ConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     ".foundagent.yaml not found",
			Remediation: "Run 'fa init' to initialize the workspace",
			Fixable:     false,
		}
	}
	
	// Check .foundagent/ directory
	foundagentDir := filepath.Join(c.Workspace.Path, workspace.FoundagentDir)
	if _, err := os.Stat(foundagentDir); os.IsNotExist(err) {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     ".foundagent/ directory not found",
			Remediation: "Run 'fa init --force' to reinitialize the workspace",
			Fixable:     false,
		}
	}
	
	// Check state.json
	statePath := c.Workspace.StatePath()
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "state.json not found",
			Remediation: "Run 'fa doctor --fix' to regenerate state file",
			Fixable:     true,
		}
	}
	
	// Check repos/ directory
	reposDir := filepath.Join(c.Workspace.Path, workspace.ReposDir)
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "repos/ directory not found",
			Remediation: "Run 'fa init --force' to reinitialize the workspace",
			Fixable:     false,
		}
	}
	
	// Check repos/.bare/
	bareDir := filepath.Join(reposDir, workspace.BareDir)
	if _, err := os.Stat(bareDir); os.IsNotExist(err) {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusWarn,
			Message:     "repos/.bare/ directory not found",
			Remediation: "Create directory: mkdir -p repos/.bare",
			Fixable:     true,
		}
	}
	
	// Check repos/worktrees/
	worktreesDir := filepath.Join(reposDir, workspace.WorktreesDir)
	if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusWarn,
			Message:     "repos/worktrees/ directory not found",
			Remediation: "Create directory: mkdir -p repos/worktrees",
			Fixable:     true,
		}
	}
	
	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "All required directories present",
		Fixable: false,
	}
}

// ConfigValidCheck checks if config file is valid
type ConfigValidCheck struct {
	Workspace *workspace.Workspace
}

func (c ConfigValidCheck) Name() string {
	return "Config file valid"
}

func (c ConfigValidCheck) Run() CheckResult {
	_, err := c.Workspace.LoadConfig()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Config file is invalid or corrupted",
			Remediation: "Check .foundagent.yaml syntax or run 'fa init --force'",
			Fixable:     false,
		}
	}
	
	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Config file is valid",
		Fixable: false,
	}
}

// StateValidCheck checks if state file is valid
type StateValidCheck struct{
	Workspace *workspace.Workspace
}

func (c StateValidCheck) Name() string {
	return "State file valid"
}

func (c StateValidCheck) Run() CheckResult {
	_, err := c.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "State file is invalid or corrupted",
			Remediation: "Run 'fa doctor --fix' to regenerate state file",
			Fixable:     true,
		}
	}
	
	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "State file is valid",
		Fixable: false,
	}
}
