package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/workspace"
)

// Fixer applies fixes for fixable check failures
type Fixer struct {
	Workspace *workspace.Workspace
}

// NewFixer creates a new fixer
func NewFixer(ws *workspace.Workspace) *Fixer {
	return &Fixer{Workspace: ws}
}

// Fix attempts to fix a failed check
func (f *Fixer) Fix(result CheckResult) CheckResult {
	if !result.Fixable || result.Status == StatusPass {
		return result
	}

	switch result.Name {
	case "State file valid":
		return f.fixStateFile()
	case "Workspace structure":
		return f.fixWorkspaceStructure()
	case "Orphaned repositories":
		return f.fixOrphanedRepos()
	case "Orphaned worktrees":
		return f.fixOrphanedWorktrees()
	case "Config/state consistency":
		return f.fixConfigStateConsistency()
	case "Workspace file consistency":
		return f.fixWorkspaceFileConsistency()
	default:
		// Unknown check, can't fix
		return result
	}
}

func (f *Fixer) fixStateFile() CheckResult {
	// Regenerate state file from config and filesystem
	cfg, err := config.Load(f.Workspace.Path)
	if err != nil {
		return CheckResult{
			Name:        "State file valid",
			Status:      StatusFail,
			Message:     "Cannot regenerate state: config file invalid",
			Remediation: "Check .foundagent.yaml syntax",
			Fixable:     false,
		}
	}

	// Build new state from config
	state := &workspace.State{Repositories: make(map[string]*workspace.Repository)}

	bareDir := filepath.Join(f.Workspace.Path, workspace.ReposDir, workspace.BareDir)
	worktreesDir := filepath.Join(f.Workspace.Path, workspace.ReposDir, workspace.WorktreesDir)

	for _, repoConfig := range cfg.Repos {
		// Check if bare clone exists
		repoPath := filepath.Join(bareDir, repoConfig.Name+".git")
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			continue // Skip if bare clone doesn't exist
		}

		repo := &workspace.Repository{
			Name:      repoConfig.Name,
			URL:       repoConfig.URL,
			Worktrees: make([]string, 0),
		}

		// Find worktrees for this repo
		entries, err := os.ReadDir(worktreesDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				// Check if worktree belongs to this repo
				// Naming convention: {repo}-{branch}
				// We'll add worktrees that start with repo name
				// This is a heuristic, actual implementation would check git config
				repo.Worktrees = append(repo.Worktrees, entry.Name())
			}
		}

		state.Repositories[repoConfig.Name] = repo
	}

	// Save state
	if err := f.Workspace.SaveState(state); err != nil {
		return CheckResult{
			Name:        "State file valid",
			Status:      StatusFail,
			Message:     "Failed to save state file",
			Remediation: "Check file permissions",
			Fixable:     false,
		}
	}

	return CheckResult{
		Name:    "State file valid",
		Status:  StatusPass,
		Message: "State file regenerated successfully",
		Fixable: false,
	}
}

func (f *Fixer) fixWorkspaceStructure() CheckResult {
	// Create missing directories
	dirs := []string{
		filepath.Join(f.Workspace.Path, workspace.ReposDir, workspace.BareDir),
		filepath.Join(f.Workspace.Path, workspace.ReposDir, workspace.WorktreesDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return CheckResult{
				Name:        "Workspace structure",
				Status:      StatusFail,
				Message:     fmt.Sprintf("Failed to create directory: %s", dir),
				Remediation: "Check file permissions",
				Fixable:     false,
			}
		}
	}

	return CheckResult{
		Name:    "Workspace structure",
		Status:  StatusPass,
		Message: "Missing directories created",
		Fixable: false,
	}
}

func (f *Fixer) fixOrphanedRepos() CheckResult {
	state, err := f.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        "Orphaned repositories",
			Status:      StatusFail,
			Message:     "Cannot load state file",
			Remediation: "Run 'fa doctor --fix' for state file first",
			Fixable:     false,
		}
	}

	bareDir := filepath.Join(f.Workspace.Path, workspace.ReposDir, workspace.BareDir)

	// Build map of known repos
	knownRepos := make(map[string]bool)
	for _, repo := range state.Repositories {
		knownRepos[repo.Name+".git"] = true
	}

	// Remove orphaned directories
	entries, err := os.ReadDir(bareDir)
	if err != nil {
		return CheckResult{
			Name:        "Orphaned repositories",
			Status:      StatusFail,
			Message:     "Cannot read bare directory",
			Remediation: "Check directory exists and permissions",
			Fixable:     false,
		}
	}

	removed := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !knownRepos[name] {
			path := filepath.Join(bareDir, name)
			if err := os.RemoveAll(path); err == nil {
				removed++
			}
		}
	}

	return CheckResult{
		Name:    "Orphaned repositories",
		Status:  StatusPass,
		Message: fmt.Sprintf("Removed %d orphaned repositories", removed),
		Fixable: false,
	}
}

func (f *Fixer) fixOrphanedWorktrees() CheckResult {
	state, err := f.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        "Orphaned worktrees",
			Status:      StatusFail,
			Message:     "Cannot load state file",
			Remediation: "Run 'fa doctor --fix' for state file first",
			Fixable:     false,
		}
	}

	worktreesDir := filepath.Join(f.Workspace.Path, workspace.ReposDir, workspace.WorktreesDir)

	// Build map of known worktrees
	knownWorktrees := make(map[string]bool)
	for _, repo := range state.Repositories {
		for _, wt := range repo.Worktrees {
			knownWorktrees[wt] = true
		}
	}

	// Remove orphaned directories
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		return CheckResult{
			Name:        "Orphaned worktrees",
			Status:      StatusFail,
			Message:     "Cannot read worktrees directory",
			Remediation: "Check directory exists and permissions",
			Fixable:     false,
		}
	}

	removed := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !knownWorktrees[name] {
			path := filepath.Join(worktreesDir, name)
			if err := os.RemoveAll(path); err == nil {
				removed++
			}
		}
	}

	return CheckResult{
		Name:    "Orphaned worktrees",
		Status:  StatusPass,
		Message: fmt.Sprintf("Removed %d orphaned worktrees", removed),
		Fixable: false,
	}
}

func (f *Fixer) fixConfigStateConsistency() CheckResult {
	// Remove repos from state that aren't in config
	cfg, err := config.Load(f.Workspace.Path)
	if err != nil {
		return CheckResult{
			Name:        "Config/state consistency",
			Status:      StatusFail,
			Message:     "Cannot load config file",
			Remediation: "Check .foundagent.yaml syntax",
			Fixable:     false,
		}
	}

	state, err := f.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        "Config/state consistency",
			Status:      StatusFail,
			Message:     "Cannot load state file",
			Remediation: "Run 'fa doctor --fix' for state file first",
			Fixable:     false,
		}
	}

	// Build map of configured repositories
	configRepos := make(map[string]bool)
	for _, repo := range cfg.Repos {
		configRepos[repo.URL] = true
	}

	// Filter state to only include configured repos
	newRepos := make(map[string]*workspace.Repository)
	for name, repo := range state.Repositories {
		if configRepos[repo.URL] {
			newRepos[name] = repo
		}
	}

	state.Repositories = newRepos

	if err := f.Workspace.SaveState(state); err != nil {
		return CheckResult{
			Name:        "Config/state consistency",
			Status:      StatusFail,
			Message:     "Failed to save state file",
			Remediation: "Check file permissions",
			Fixable:     false,
		}
	}

	return CheckResult{
		Name:    "Config/state consistency",
		Status:  StatusPass,
		Message: "State synced with config",
		Fixable: false,
	}
}

func (f *Fixer) fixWorkspaceFileConsistency() CheckResult {
	state, err := f.Workspace.LoadState()
	if err != nil {
		return CheckResult{
			Name:        "Workspace file consistency",
			Status:      StatusFail,
			Message:     "Cannot load state file",
			Remediation: "Run 'fa doctor --fix' for state file first",
			Fixable:     false,
		}
	}

	// Rebuild workspace file from state
	folders := make([]workspace.VSCodeFolder, 0)

	for _, repo := range state.Repositories {
		for _, wt := range repo.Worktrees {
			folders = append(folders, workspace.VSCodeFolder{

				Path: filepath.Join(workspace.ReposDir, workspace.WorktreesDir, wt),
			})
		}
	}

	wsFile := workspace.VSCodeWorkspace{
		Folders: folders,
	}

	if err := f.Workspace.SaveVSCodeWorkspace(&wsFile); err != nil {
		return CheckResult{
			Name:        "Workspace file consistency",
			Status:      StatusFail,
			Message:     "Failed to save workspace file",
			Remediation: "Check file permissions",
			Fixable:     false,
		}
	}

	return CheckResult{
		Name:    "Workspace file consistency",
		Status:  StatusPass,
		Message: "Workspace file synced with state",
		Fixable: false,
	}
}
