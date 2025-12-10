package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/workspace"
)

// RepositoriesCheck checks repository integrity
type RepositoriesCheck struct {
	Workspace *workspace.Workspace
}

func (c RepositoriesCheck) Name() string {
	return "Repository integrity"
}

func (c RepositoriesCheck) Run() CheckResult {
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
	
	bareDir := filepath.Join(c.Workspace.Path, workspace.ReposDir, workspace.BareDir)
	issues := make([]string, 0)
	
	// Check each repository
	for _, repo := range state.Repositories {
		repoPath := filepath.Join(bareDir, repo.Name+".git")
		
		// Check if directory exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Missing bare clone: %s", repo.Name))
			continue
		}
		
		// Check if it's a valid git repository (has objects, refs, HEAD)
		objectsPath := filepath.Join(repoPath, "objects")
		if _, err := os.Stat(objectsPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Invalid git repository: %s", repo.Name))
		}
	}
	
	if len(issues) > 0 {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusFail,
			Message:     fmt.Sprintf("Found %d repository issue(s)", len(issues)),
			Remediation: "Run 'fa remove' to clean up or 'fa add' to re-clone",
			Fixable:     false,
		}
	}
	
	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("All %d repositories valid", len(state.Repositories)),
		Fixable: false,
	}
}

// OrphanedReposCheck checks for orphaned repository directories
type OrphanedReposCheck struct {
	Workspace *workspace.Workspace
}

func (c OrphanedReposCheck) Name() string {
	return "Orphaned repositories"
}

func (c OrphanedReposCheck) Run() CheckResult {
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
	
	bareDir := filepath.Join(c.Workspace.Path, workspace.ReposDir, workspace.BareDir)
	
	// Get all .git directories in bare/
	entries, err := os.ReadDir(bareDir)
	if err != nil {
		// Directory doesn't exist or can't be read
		return CheckResult{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "No repositories found",
			Fixable: false,
		}
	}
	
	// Build map of known repos
	knownRepos := make(map[string]bool)
	for _, repo := range state.Repositories {
		knownRepos[repo.Name+".git"] = true
	}
	
	// Check for orphaned directories
	orphaned := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if !knownRepos[name] {
			orphaned = append(orphaned, name)
		}
	}
	
	if len(orphaned) > 0 {
		return CheckResult{
			Name:        c.Name(),
			Status:      StatusWarn,
			Message:     fmt.Sprintf("Found %d orphaned repository directories", len(orphaned)),
			Remediation: "Run 'fa doctor --fix' to remove orphaned entries",
			Fixable:     true,
		}
	}
	
	return CheckResult{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "No orphaned repositories",
		Fixable: false,
	}
}
