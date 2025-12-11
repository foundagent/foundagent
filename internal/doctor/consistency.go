package doctor

import (
	"fmt"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
)

// ConfigStateConsistencyCheck checks if config and state are in sync
type ConfigStateConsistencyCheck struct {
	Workspace *workspace.Workspace
}

func (c ConfigStateConsistencyCheck) Name() string {
	return "Config/state consistency"
}

func (c ConfigStateConsistencyCheck) Run() CheckResult {
	cfg, err := config.Load(c.Workspace.Path)
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Could not load config file",
			Remediation: "Check .foundagent.yaml syntax",
			Fixable:     false,
		}
	}

	state, err := c.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Could not load state file",
			Remediation: "Run 'fa doctor --fix' to regenerate state file",
			Fixable:     true,
		}
	}

	// Build map of configured repositories
	configRepos := make(map[string]bool)
	for _, repo := range cfg.Repos {
		configRepos[repo.URL] = true
	}

	// Build map of state repositories
	stateRepos := make(map[string]bool)
	for _, repo := range state.Repositories {
		stateRepos[repo.URL] = true
	}

	// Check for repos in config but not in state
	missing := make([]string, 0)
	for url := range configRepos {
		if !stateRepos[url] {
			missing = append(missing, url)
		}
	}

	// Check for repos in state but not in config
	orphaned := make([]string, 0)
	for url := range stateRepos {
		if !configRepos[url] {
			orphaned = append(orphaned, url)
		}
	}

	if len(missing) > 0 {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     fmt.Sprintf("%d repositories in config but not in state", len(missing)),
			Remediation: "Run 'fa add' to clone missing repositories",
			Fixable:     false,
		}
	}

	if len(orphaned) > 0 {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusWarn,
			Message:     fmt.Sprintf("%d repositories in state but not in config", len(orphaned)),
			Remediation: "Run 'fa doctor --fix' to clean up state file",
			Fixable:     true,
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Config and state are in sync",
		Fixable: false,
	}
}

// WorkspaceFileConsistencyCheck checks if workspace file matches state
type WorkspaceFileConsistencyCheck struct {
	Workspace *workspace.Workspace
}

func (c WorkspaceFileConsistencyCheck) Name() string {
	return "Workspace file consistency"
}

func (c WorkspaceFileConsistencyCheck) Run() CheckResult {
	state, err := c.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     "Could not load state file",
			Remediation: "Run 'fa doctor --fix' to regenerate state file",
			Fixable:     true,
		}
	}

	// Load workspace file
	wsFile, err := c.Workspace.LoadVSCodeWorkspace()
	if err != nil {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusWarn,
			Message:     "Could not load workspace file",
			Remediation: "Run 'fa doctor --fix' to regenerate workspace file",
			Fixable:     true,
		}
	}

	// Count worktrees in state
	stateWorktrees := make(map[string]bool)
	for _, repo := range state.Repositories {
		for _, wt := range repo.Worktrees {
			// Store as repo/branch format
			stateWorktrees[repo.Name+"/"+wt] = true
		}
	}

	// Count folders in workspace file
	wsWorktrees := make(map[string]bool)
	for _, folder := range wsFile.Folders {
		// Extract worktree path from folder path
		// Path format: repos/worktrees/{repoName}/{branch}
		if len(folder.Path) > len("repos/worktrees/") {
			path := folder.Path[len("repos/worktrees/"):]
			wsWorktrees[path] = true
		}
	}

	// Check for missing worktrees in workspace file
	missing := 0
	for wt := range stateWorktrees {
		if !wsWorktrees[wt] {
			missing++
		}
	}

	// Check for extra worktrees in workspace file
	extra := 0
	for wt := range wsWorktrees {
		if !stateWorktrees[wt] {
			extra++
		}
	}

	if missing > 0 || extra > 0 {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusWarn,
			Message:     fmt.Sprintf("Workspace file out of sync (%d missing, %d extra)", missing, extra),
			Remediation: "Run 'fa doctor --fix' to sync workspace file",
			Fixable:     true,
		}
	}

	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Workspace file is in sync with state",
		Fixable: false,
	}
}
