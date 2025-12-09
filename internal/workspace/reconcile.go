package workspace

import (
	"fmt"

	"github.com/foundagent/foundagent/internal/config"
)

// ReconcileResult represents the result of reconciling config with state
type ReconcileResult struct {
	ReposToClone  []config.RepoConfig
	ReposUpToDate []string
	ReposStale    []string // In state but not in config
}

// Reconcile compares config with state and determines what actions to take
func Reconcile(ws *Workspace) (*ReconcileResult, error) {
	// Load config
	cfg, err := config.Load(ws.Path)
	if err != nil {
		return nil, err
	}

	// Load state
	state, err := ws.LoadState()
	if err != nil {
		return nil, err
	}

	result := &ReconcileResult{
		ReposToClone:  []config.RepoConfig{},
		ReposUpToDate: []string{},
		ReposStale:    []string{},
	}

	// Check which repos from config need to be cloned
	for _, repo := range cfg.Repos {
		name := repo.Name
		if name == "" {
			// Name should have been inferred during validation
			continue
		}

		if _, exists := state.Repositories[name]; !exists {
			// Repo in config but not cloned
			result.ReposToClone = append(result.ReposToClone, repo)
		} else {
			// Repo exists and is up-to-date
			result.ReposUpToDate = append(result.ReposUpToDate, name)
		}
	}

	// Check for stale repos (in state but not in config)
	configRepos := make(map[string]bool)
	for _, repo := range cfg.Repos {
		if repo.Name != "" {
			configRepos[repo.Name] = true
		}
	}

	for name := range state.Repositories {
		if !configRepos[name] {
			result.ReposStale = append(result.ReposStale, name)
		}
	}

	return result, nil
}

// PrintReconcileResult prints a summary of reconciliation results
func PrintReconcileResult(result *ReconcileResult) {
	if len(result.ReposToClone) > 0 {
		fmt.Printf("Repositories to clone: %d\n", len(result.ReposToClone))
		for _, repo := range result.ReposToClone {
			fmt.Printf("  - %s (%s)\n", repo.Name, repo.URL)
		}
	}

	if len(result.ReposUpToDate) > 0 {
		fmt.Printf("Repositories up-to-date: %d\n", len(result.ReposUpToDate))
	}

	if len(result.ReposStale) > 0 {
		fmt.Printf("\nWarning: The following repositories exist locally but are not in config:\n")
		for _, name := range result.ReposStale {
			fmt.Printf("  - %s\n", name)
		}
		fmt.Printf("\nTo clean up: fa remove %s\n", result.ReposStale[0])
		fmt.Printf("To keep: Add them back to .foundagent.yaml\n")
	}
}
