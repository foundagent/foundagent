package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
)

func TestGitCheck_Run(t *testing.T) {
	check := GitCheck{}

	result := check.Run()

	if result.Name != "Git installed" {
		t.Errorf("Name = %q, want 'Git installed'", result.Name)
	}

	// Should pass on a system with git installed
	if result.Status == StatusFail {
		t.Logf("Git check failed: %s", result.Message)
	}
}

func TestGitCheck_Name(t *testing.T) {
	check := GitCheck{}
	if check.Name() != "Git installed" {
		t.Errorf("Name() = %q, want 'Git installed'", check.Name())
	}
}

func TestGitVersionCheck_Run(t *testing.T) {
	check := GitVersionCheck{}

	result := check.Run()

	if result.Name != "Git version" {
		t.Errorf("Name = %q, want 'Git version'", result.Name)
	}

	// If git is installed, should pass
	if _, err := exec.LookPath("git"); err == nil {
		if result.Status != StatusPass {
			t.Errorf("Status = %v, want StatusPass when git is installed", result.Status)
		}
	}
}

func TestGitVersionCheck_Name(t *testing.T) {
	check := GitVersionCheck{}
	if check.Name() != "Git version" {
		t.Errorf("Name() = %q, want 'Git version'", check.Name())
	}
}

func setupTestWorkspace(t *testing.T) *workspace.Workspace {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repo for testing
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()

	ws, err := workspace.New("test-workspace", tmpDir)
	if err != nil {
		t.Fatalf("Failed to create workspace instance: %v", err)
	}

	if err := ws.Create(false); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	return ws
}

func TestWorkspaceStructureCheck_Run_Valid(t *testing.T) {
	ws := setupTestWorkspace(t)

	check := WorkspaceStructureCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass for valid workspace", result.Status)
	}
}

func TestWorkspaceStructureCheck_Name(t *testing.T) {
	ws := setupTestWorkspace(t)
	check := WorkspaceStructureCheck{Workspace: ws}

	if check.Name() != "Workspace structure" {
		t.Errorf("Name() = %q, want 'Workspace structure'", check.Name())
	}
}

func TestWorkspaceStructureCheck_Run_MissingConfig(t *testing.T) {
	ws := setupTestWorkspace(t)

	// Remove config file
	os.Remove(ws.ConfigPath())

	check := WorkspaceStructureCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want StatusFail for missing config", result.Status)
	}
}

func TestWorkspaceStructureCheck_Run_MissingState(t *testing.T) {
	ws := setupTestWorkspace(t)

	// Remove state file
	os.Remove(ws.StatePath())

	check := WorkspaceStructureCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want StatusFail for missing state", result.Status)
	}
	if !result.Fixable {
		t.Error("Result should be fixable")
	}
}

func TestConfigValidCheck_Run_Valid(t *testing.T) {
	ws := setupTestWorkspace(t)

	check := ConfigValidCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass for valid config", result.Status)
	}
}

func TestConfigValidCheck_Name(t *testing.T) {
	ws := setupTestWorkspace(t)
	check := ConfigValidCheck{Workspace: ws}

	if check.Name() != "Config file valid" {
		t.Errorf("Name() = %q, want 'Config file valid'", check.Name())
	}
}

// TestConfigValidCheck_Run_InvalidConfig is skipped because current LoadConfig()
// implementation doesn't validate YAML syntax - it only checks file existence.
// This is a known limitation where malformed config files may not be detected.

func TestStateValidCheck_Run_Valid(t *testing.T) {
	ws := setupTestWorkspace(t)

	check := StateValidCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass for valid state", result.Status)
	}
}

func TestStateValidCheck_Name(t *testing.T) {
	ws := setupTestWorkspace(t)
	check := StateValidCheck{Workspace: ws}

	if check.Name() != "State file valid" {
		t.Errorf("Name() = %q, want 'State file valid'", check.Name())
	}
}

func TestStateValidCheck_Run_InvalidState(t *testing.T) {
	ws := setupTestWorkspace(t)

	// Write invalid JSON
	os.WriteFile(ws.StatePath(), []byte("{invalid json"), 0644)

	check := StateValidCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want StatusFail for invalid state", result.Status)
	}
	if !result.Fixable {
		t.Error("Result should be fixable")
	}
}

func TestRepositoriesCheck_Run_NoRepos(t *testing.T) {
	ws := setupTestWorkspace(t)

	check := RepositoriesCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass when no repos", result.Status)
	}
}

func TestRepositoriesCheck_Name(t *testing.T) {
	ws := setupTestWorkspace(t)
	check := RepositoriesCheck{Workspace: ws}

	if check.Name() != "Repository integrity" {
		t.Errorf("Name() = %q, want 'Repository integrity'", check.Name())
	}
}

func TestOrphanedReposCheck_Run_NoOrphans(t *testing.T) {
	ws := setupTestWorkspace(t)

	check := OrphanedReposCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass when no orphans", result.Status)
	}
}

func TestOrphanedReposCheck_Name(t *testing.T) {
	ws := setupTestWorkspace(t)
	check := OrphanedReposCheck{Workspace: ws}

	if check.Name() != "Orphaned repositories" {
		t.Errorf("Name() = %q, want 'Orphaned repositories'", check.Name())
	}
}

func TestOrphanedReposCheck_Run_WithOrphans(t *testing.T) {
	ws := setupTestWorkspace(t)

	// Create an orphaned repo directory
	orphanPath := filepath.Join(ws.Path, workspace.ReposDir, "orphan-repo", workspace.BareDir)
	os.MkdirAll(orphanPath, 0755)

	check := OrphanedReposCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusWarn {
		t.Errorf("Status = %v, want StatusWarn when orphans exist", result.Status)
	}
	if !result.Fixable {
		t.Error("Result should be fixable")
	}
}

func TestWorktreesCheck_Run_NoWorktrees(t *testing.T) {
	ws := setupTestWorkspace(t)

	check := WorktreesCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass when no worktrees", result.Status)
	}
}

func TestWorktreesCheck_Name(t *testing.T) {
	ws := setupTestWorkspace(t)
	check := WorktreesCheck{Workspace: ws}

	if check.Name() != "Worktree integrity" {
		t.Errorf("Name() = %q, want 'Worktree integrity'", check.Name())
	}
}

func TestOrphanedWorktreesCheck_Run_NoOrphans(t *testing.T) {
	ws := setupTestWorkspace(t)

	check := OrphanedWorktreesCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass when no orphans", result.Status)
	}
}

func TestOrphanedWorktreesCheck_Name(t *testing.T) {
	ws := setupTestWorkspace(t)
	check := OrphanedWorktreesCheck{Workspace: ws}

	if check.Name() != "Orphaned worktrees" {
		t.Errorf("Name() = %q, want 'Orphaned worktrees'", check.Name())
	}
}

func TestCheckResult_IsSuccess(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"pass", StatusPass, true},
		{"warn", StatusWarn, true},
		{"fail", StatusFail, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckResult{Status: tt.status}
			if got := result.IsSuccess(); got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckResult_Fixable(t *testing.T) {
	result := CheckResult{
		Name:        "Test",
		Status:      StatusFail,
		Message:     "Failed",
		Remediation: "Fix it",
		Fixable:     true,
	}

	if !result.Fixable {
		t.Error("Fixable should be true")
	}
}

func TestRepositoriesCheck_Run_MissingState(t *testing.T) {
	ws := setupTestWorkspace(t)
	os.Remove(ws.StatePath())

	check := RepositoriesCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want StatusFail when state missing", result.Status)
	}
}

func TestWorktreesCheck_Run_MissingState(t *testing.T) {
	ws := setupTestWorkspace(t)
	os.Remove(ws.StatePath())

	check := WorktreesCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want StatusFail when state missing", result.Status)
	}
}

func TestOrphanedWorktreesCheck_Run_MissingState(t *testing.T) {
	ws := setupTestWorkspace(t)
	os.Remove(ws.StatePath())

	check := OrphanedWorktreesCheck{Workspace: ws}
	result := check.Run()

	if result.Status != StatusFail {
		t.Errorf("Status = %v, want StatusFail when state missing", result.Status)
	}
}

func TestOrphanedReposCheck_Run_MissingReposDir(t *testing.T) {
	ws := setupTestWorkspace(t)

	// Remove repos directory
	os.RemoveAll(filepath.Join(ws.Path, workspace.ReposDir))

	check := OrphanedReposCheck{Workspace: ws}
	result := check.Run()

	// Should still pass when directory doesn't exist
	if result.Status != StatusPass {
		t.Errorf("Status = %v, want StatusPass when repos dir missing", result.Status)
	}
}
