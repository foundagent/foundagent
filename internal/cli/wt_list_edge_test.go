package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectWorktreeStatus_UncommittedChanges(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory that exists
	err := os.MkdirAll(tmpDir+"/testpath", 0755)
	if err != nil {
		t.Skip("Cannot create test directory")
	}

	// Test with existing path (will fail on git check but tests path existence)
	status, desc := detectWorktreeStatus(tmpDir + "/testpath")

	// Should get to the git check phase
	assert.NotEmpty(t, status)
	_ = desc
}

func TestDetectWorktreeStatus_NonExistentPath(t *testing.T) {
	// Test with non-existent path
	status, desc := detectWorktreeStatus("/nonexistent/path/that/does/not/exist")

	assert.Equal(t, "error", status)
	assert.Contains(t, desc, "not found")
}

func TestMarkCurrentWorktree_NoMatch(t *testing.T) {
	worktrees := []worktreeInfo{
		{Path: "/path/to/repo1/branch1", Branch: "branch1"},
		{Path: "/path/to/repo2/branch2", Branch: "branch2"},
	}

	result := markCurrentWorktree(worktrees, "/completely/different/path")

	// None should be marked as current
	for _, wt := range result {
		assert.False(t, wt.IsCurrent)
	}
}

func TestMarkCurrentWorktree_WithMatch(t *testing.T) {
	worktrees := []worktreeInfo{
		{Path: "/path/to/repo1/branch1", Branch: "branch1"},
		{Path: "/path/to/repo2/branch2", Branch: "branch2"},
	}

	result := markCurrentWorktree(worktrees, "/path/to/repo1/branch1/subdir")

	// First one should be marked as current
	assert.True(t, result[0].IsCurrent)
	assert.False(t, result[1].IsCurrent)
}

func TestPrintHumanList_EmptyList(t *testing.T) {
	// Test with empty list - should not panic
	printHumanList([]worktreeInfo{})
}

func TestPrintHumanList_MultipleBranches(t *testing.T) {
	worktrees := []worktreeInfo{
		{Branch: "main", Path: "/path/repo1/main", Repo: "repo1", IsCurrent: false},
		{Branch: "main", Path: "/path/repo2/main", Repo: "repo2", IsCurrent: false},
		{Branch: "feature", Path: "/path/repo1/feature", Repo: "repo1", IsCurrent: true},
	}

	// Should not panic
	printHumanList(worktrees)
}

func TestPrintHumanList_WithStatusIcons(t *testing.T) {
	worktrees := []worktreeInfo{
		{Branch: "main", Path: "/path/repo1/main", Repo: "repo1", Status: "clean", IsCurrent: false},
		{Branch: "feature", Path: "/path/repo2/feature", Repo: "repo2", Status: "modified", StatusDesc: "changes", IsCurrent: true},
		{Branch: "test", Path: "/path/repo3/test", Repo: "repo3", Status: "error", StatusDesc: "not found", IsCurrent: false},
	}

	// Should not panic
	printHumanList(worktrees)
}

func TestPrintJSONList_EmptyList(t *testing.T) {
	// Test with empty list
	output := listOutput{
		WorkspaceName:  "test-ws",
		TotalWorktrees: 0,
		TotalBranches:  0,
		Worktrees:      []worktreeInfo{},
	}
	err := printJSONList(output)
	assert.NoError(t, err)
}

func TestPrintJSONList_SingleBranch(t *testing.T) {
	worktrees := []worktreeInfo{
		{Branch: "main", Path: "/path/repo1/main", Repo: "repo1", IsCurrent: true},
	}

	output := listOutput{
		WorkspaceName:  "test-ws",
		TotalWorktrees: 1,
		TotalBranches:  1,
		Worktrees:      worktrees,
	}
	err := printJSONList(output)
	assert.NoError(t, err)
}

func TestPrintJSONList_MultipleBranches(t *testing.T) {
	worktrees := []worktreeInfo{
		{Branch: "main", Path: "/path/repo1/main", Repo: "repo1"},
		{Branch: "feature", Path: "/path/repo2/feature", Repo: "repo2"},
		{Branch: "test", Path: "/path/repo3/test", Repo: "repo3"},
	}

	output := listOutput{
		WorkspaceName:  "test-ws",
		TotalWorktrees: 3,
		TotalBranches:  3,
		Worktrees:      worktrees,
	}
	err := printJSONList(output)
	assert.NoError(t, err)
}
