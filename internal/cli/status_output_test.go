package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputStatusHuman_BasicWorkspace tests output for basic workspace
func TestOutputStatusHuman_BasicWorkspace(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:            0,
			TotalWorktrees:        0,
			TotalBranches:         0,
			HasUncommittedChanges: false,
			ConfigInSync:          true,
		},
		Repos:     []workspace.RepoStatus{},
		Worktrees: []workspace.WorktreeStatus{},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Workspace: test-ws")
	assert.Contains(t, output, "Path:")
	assert.Contains(t, output, "Summary:")
	assert.Contains(t, output, "Repositories: 0")
	assert.Contains(t, output, "Worktrees: 0")
	assert.Contains(t, output, "Branches: 0")
}

// TestOutputStatusHuman_DirtyWorktreesIndicator tests output with dirty worktrees
func TestOutputStatusHuman_DirtyWorktreesIndicator(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:            1,
			TotalWorktrees:        2,
			TotalBranches:         2,
			HasUncommittedChanges: true,
			DirtyWorktrees:        1,
			ConfigInSync:          true,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{
			{Branch: "main", Repo: "test-repo", Path: "/path/to/main", Status: "modified"},
			{Branch: "develop", Repo: "test-repo", Path: "/path/to/develop", Status: "clean"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "Uncommitted changes:")
	assert.Contains(t, output, "1 worktree(s)")
}

// TestOutputStatusHuman_AllWorktreesClean tests output when all worktrees are clean
func TestOutputStatusHuman_AllWorktreesClean(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:            1,
			TotalWorktrees:        2,
			TotalBranches:         2,
			HasUncommittedChanges: false,
			DirtyWorktrees:        0,
			ConfigInSync:          true,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{
			{Branch: "main", Repo: "test-repo", Path: "/path/to/main", Status: "clean"},
			{Branch: "develop", Repo: "test-repo", Path: "/path/to/develop", Status: "clean"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "All worktrees clean")
}

// TestOutputStatusHuman_ReposNotCloned tests output with repos not cloned
func TestOutputStatusHuman_ReposNotCloned(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:            2,
			TotalWorktrees:        0,
			TotalBranches:         0,
			HasUncommittedChanges: false,
			ConfigInSync:          false,
			ReposNotCloned:        1,
		},
		Repos: []workspace.RepoStatus{
			{Name: "cloned-repo", URL: "https://github.com/test/repo1.git", IsCloned: true},
			{Name: "not-cloned", URL: "https://github.com/test/repo2.git", IsCloned: false},
		},
		Worktrees: []workspace.WorktreeStatus{},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "1 repo(s) not cloned")
	assert.Contains(t, output, "not-cloned")
	assert.Contains(t, output, "[not cloned]")
}

// TestOutputStatusHuman_ConfigInSync tests output when config is in sync
func TestOutputStatusHuman_ConfigInSync(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:            1,
			TotalWorktrees:        1,
			TotalBranches:         1,
			HasUncommittedChanges: false,
			ConfigInSync:          true,
			ReposNotCloned:        0,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{
			{Branch: "main", Repo: "test-repo", Path: "/path/to/main", Status: "clean"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "In sync")
}

// TestOutputStatusHuman_NoRepositories tests output with no repositories
func TestOutputStatusHuman_NoRepositories(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:     0,
			TotalWorktrees: 0,
			TotalBranches:  0,
			ConfigInSync:   true,
		},
		Repos:     []workspace.RepoStatus{},
		Worktrees: []workspace.WorktreeStatus{},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "No repositories configured")
	assert.Contains(t, output, "fa add")
}

// TestOutputStatusHuman_VerboseModeWithURL tests verbose output with repo URLs
func TestOutputStatusHuman_VerboseModeWithURL(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:     1,
			TotalWorktrees: 1,
			TotalBranches:  1,
			ConfigInSync:   true,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{
			{Branch: "main", Repo: "test-repo", Path: "/path/to/main", Status: "clean"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, true)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "URL: https://github.com/test/repo.git")
}

// TestOutputStatusHuman_WorktreeStatuses tests different worktree status indicators
func TestOutputStatusHuman_WorktreeStatuses(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:            1,
			TotalWorktrees:        4,
			TotalBranches:         4,
			HasUncommittedChanges: true,
			DirtyWorktrees:        3,
			ConfigInSync:          true,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{
			{Branch: "main", Repo: "test-repo", Path: "/path/to/main", Status: "clean", IsCurrent: true},
			{Branch: "feature", Repo: "test-repo", Path: "/path/to/feature", Status: "modified"},
			{Branch: "develop", Repo: "test-repo", Path: "/path/to/develop", Status: "untracked"},
			{Branch: "hotfix", Repo: "test-repo", Path: "/path/to/hotfix", Status: "conflict"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "[modified]")
	assert.Contains(t, output, "[untracked]")
	assert.Contains(t, output, "[conflict]")
	assert.Contains(t, output, "Branch:")
}

// TestOutputStatusHuman_CurrentWorktreeMarker tests current worktree indicator
func TestOutputStatusHuman_CurrentWorktreeMarker(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:     1,
			TotalWorktrees: 2,
			TotalBranches:  2,
			ConfigInSync:   true,
		},
		Repos: []workspace.RepoStatus{
			{Name: "test-repo", URL: "https://github.com/test/repo.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{
			{Branch: "main", Repo: "test-repo", Path: "/path/to/main", Status: "clean", IsCurrent: true},
			{Branch: "develop", Repo: "test-repo", Path: "/path/to/develop", Status: "clean", IsCurrent: false},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	// Output should contain worktree information grouped by branch
	assert.Contains(t, output, "Branch:")
}

// TestOutputStatusHuman_MultipleRepos tests output with multiple repositories
func TestOutputStatusHuman_MultipleRepos(t *testing.T) {
	status := &workspace.WorkspaceStatus{
		WorkspaceName: "test-ws",
		WorkspacePath: "/tmp/test-ws",
		Summary: workspace.StatusSummary{
			TotalRepos:     2,
			TotalWorktrees: 2,
			TotalBranches:  2,
			ConfigInSync:   true,
		},
		Repos: []workspace.RepoStatus{
			{Name: "repo1", URL: "https://github.com/test/repo1.git", IsCloned: true},
			{Name: "repo2", URL: "https://github.com/test/repo2.git", IsCloned: true},
		},
		Worktrees: []workspace.WorktreeStatus{
			{Branch: "main", Repo: "repo1", Path: "/path/to/repo1/main", Status: "clean"},
			{Branch: "main", Repo: "repo2", Path: "/path/to/repo2/main", Status: "clean"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputStatusHuman(status, false)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	require.NoError(t, err)
	assert.Contains(t, output, "repo1")
	assert.Contains(t, output, "repo2")
	assert.Contains(t, output, "Repositories: 2")
}
