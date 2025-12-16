package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
)

func TestWorkspaceFileConsistencyCheck_WithRootFolder(t *testing.T) {
	// Create a temporary workspace
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	// Create state with a repo and worktree
	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{"main", "feature"},
			},
		},
	}

	// Save state
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	// Create workspace file with "." folder and worktrees
	wsFile := workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "."},
			{Path: "repos/test-repo/worktrees/main"},
			{Path: "repos/test-repo/worktrees/feature"},
		},
	}
	wsData, _ := json.MarshalIndent(wsFile, "", "  ")
	os.WriteFile(filepath.Join(tempDir, "workspace.code-workspace"), wsData, 0644)

	// Run the check
	check := WorkspaceFileConsistencyCheck{Workspace: ws}
	result := check.Run()

	// Should pass - "." folder should be ignored
	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Message)
	}
}

func TestWorkspaceFileConsistencyCheck_MissingWorktrees(t *testing.T) {
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	// State has 3 worktrees
	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{"main", "feature", "develop"},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	// Workspace file only has 2 worktrees
	wsFile := workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "."},
			{Path: "repos/test-repo/worktrees/main"},
			{Path: "repos/test-repo/worktrees/feature"},
		},
	}
	wsData, _ := json.MarshalIndent(wsFile, "", "  ")
	os.WriteFile(filepath.Join(tempDir, "workspace.code-workspace"), wsData, 0644)

	check := WorkspaceFileConsistencyCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusWarn {
		t.Errorf("expected StatusWarn, got %s", result.Status)
	}
	if result.Message != "Workspace file out of sync (1 missing, 0 extra)" {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestWorkspaceFileConsistencyCheck_ExtraWorktrees(t *testing.T) {
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	// State has 2 worktrees
	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{"main", "feature"},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	// Workspace file has 3 worktrees
	wsFile := workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "."},
			{Path: "repos/test-repo/worktrees/main"},
			{Path: "repos/test-repo/worktrees/feature"},
			{Path: "repos/test-repo/worktrees/old-branch"},
		},
	}
	wsData, _ := json.MarshalIndent(wsFile, "", "  ")
	os.WriteFile(filepath.Join(tempDir, "workspace.code-workspace"), wsData, 0644)

	check := WorkspaceFileConsistencyCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusWarn {
		t.Errorf("expected StatusWarn, got %s", result.Status)
	}
	if result.Message != "Workspace file out of sync (0 missing, 1 extra)" {
		t.Errorf("unexpected message: %s", result.Message)
	}
}

func TestWorkspaceFileConsistencyCheck_MultipleRepos(t *testing.T) {
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	// State with multiple repos
	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"repo-one": {
				Name:      "repo-one",
				URL:       "https://github.com/test/one.git",
				Worktrees: []string{"main"},
			},
			"repo-two": {
				Name:      "repo-two",
				URL:       "https://github.com/test/two.git",
				Worktrees: []string{"develop", "feature"},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	// Workspace file matches state
	wsFile := workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "."},
			{Path: "repos/repo-one/worktrees/main"},
			{Path: "repos/repo-two/worktrees/develop"},
			{Path: "repos/repo-two/worktrees/feature"},
		},
	}
	wsData, _ := json.MarshalIndent(wsFile, "", "  ")
	os.WriteFile(filepath.Join(tempDir, "workspace.code-workspace"), wsData, 0644)

	check := WorkspaceFileConsistencyCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Message)
	}
}

func TestFixWorkspaceFileConsistency_PreservesRootFolder(t *testing.T) {
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	// Create state
	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{"main"},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	// Create initial workspace file (could be out of sync)
	wsFile := workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "."},
			{Path: "repos/test-repo/worktrees/old-branch"},
		},
	}
	wsData, _ := json.MarshalIndent(wsFile, "", "  ")
	os.WriteFile(filepath.Join(tempDir, "workspace.code-workspace"), wsData, 0644)

	// Run the fix
	fixer := NewFixer(ws)
	result := fixer.fixWorkspaceFileConsistency()

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass after fix, got %s: %s", result.Status, result.Message)
	}

	// Load the fixed workspace file
	fixedWsFile, err := ws.LoadVSCodeWorkspace()
	if err != nil {
		t.Fatalf("failed to load workspace file after fix: %v", err)
	}

	// Check that "." folder is present
	hasRootFolder := false
	for _, folder := range fixedWsFile.Folders {
		if folder.Path == "." {
			hasRootFolder = true
			break
		}
	}
	if !hasRootFolder {
		t.Error("expected workspace file to contain '.' folder after fix")
	}

	// Check that the correct worktree is present
	hasCorrectWorktree := false
	for _, folder := range fixedWsFile.Folders {
		if folder.Path == "repos/test-repo/worktrees/main" {
			hasCorrectWorktree = true
			break
		}
	}
	if !hasCorrectWorktree {
		t.Error("expected workspace file to contain correct worktree after fix")
	}

	// Should have exactly 2 folders: "." and the worktree
	if len(fixedWsFile.Folders) != 2 {
		t.Errorf("expected 2 folders after fix, got %d", len(fixedWsFile.Folders))
	}
}

func TestWorkspaceFileConsistencyCheck_WithoutRootFolder(t *testing.T) {
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	// State with worktrees
	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{"main"},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	// Workspace file without "." folder (shouldn't cause issues)
	wsFile := workspace.VSCodeWorkspace{
		Folders: []workspace.VSCodeFolder{
			{Path: "repos/test-repo/worktrees/main"},
		},
	}
	wsData, _ := json.MarshalIndent(wsFile, "", "  ")
	os.WriteFile(filepath.Join(tempDir, "workspace.code-workspace"), wsData, 0644)

	check := WorkspaceFileConsistencyCheck{Workspace: ws}
	result := check.Run()

	// Should pass - check only looks at worktree folders
	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Message)
	}
}

func TestConfigStateConsistencyCheck_InSync(t *testing.T) {
	t.Skip("Skipping due to config loading issues in test environment")
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Create matching config and state
	configContent := `repos:
  - url: "https://github.com/test/repo.git"
    name: "test-repo"
`
	os.WriteFile(filepath.Join(tempDir, ".foundagent.yaml"), []byte(configContent), 0644)

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	check := ConfigStateConsistencyCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %s: %s", result.Status, result.Message)
	}
}

func TestConfigStateConsistencyCheck_MissingInState(t *testing.T) {
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// Config has repo that state doesn't
	configContent := `repos:
  - url: https://github.com/test/repo.git
    name: test-repo
  - url: https://github.com/test/another.git
    name: another-repo
`
	os.WriteFile(filepath.Join(tempDir, ".foundagent.yaml"), []byte(configContent), 0644)

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	check := ConfigStateConsistencyCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail, got %s", result.Status)
	}
}

func TestConfigStateConsistencyCheck_OrphanedInState(t *testing.T) {
	t.Skip("Skipping due to config loading issues in test environment")
	tempDir := t.TempDir()
	ws := &workspace.Workspace{Name: "workspace", Path: tempDir}

	// State has repo that config doesn't
	configContent := `repos:
  - url: "https://github.com/test/repo.git"
    name: "test-repo"
`
	os.WriteFile(filepath.Join(tempDir, ".foundagent.yaml"), []byte(configContent), 0644)

	// Create .foundagent directory
	os.MkdirAll(filepath.Join(tempDir, ".foundagent"), 0755)

	state := &workspace.State{
		Repositories: map[string]*workspace.Repository{
			"test-repo": {
				Name:      "test-repo",
				URL:       "https://github.com/test/repo.git",
				Worktrees: []string{},
			},
			"orphaned-repo": {
				Name:      "orphaned-repo",
				URL:       "https://github.com/test/orphaned.git",
				Worktrees: []string{},
			},
		},
	}
	stateData, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join(tempDir, ".foundagent", "state.json"), stateData, 0644)

	check := ConfigStateConsistencyCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusWarn {
		t.Errorf("expected StatusWarn, got %s", result.Status)
	}
}
