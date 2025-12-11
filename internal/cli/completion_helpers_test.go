package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetWorktreeCompletions(t *testing.T) {
	// Test graceful degradation when not in workspace
	completions, directive := getWorktreeCompletions(nil, []string{}, "")

	assert.Empty(t, completions, "Should return empty completions when not in workspace")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive, "Should use NoFileComp directive")
}

func TestGetRepoCompletions(t *testing.T) {
	// Test graceful degradation when not in workspace
	completions, directive := getRepoCompletions(nil, []string{}, "")

	assert.Empty(t, completions, "Should return empty completions when not in workspace")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive, "Should use NoFileComp directive")
}

func TestGetBranchCompletions(t *testing.T) {
	// Test graceful degradation when not in workspace
	completions, directive := getBranchCompletions(nil, []string{}, "")

	assert.Empty(t, completions, "Should return empty completions when not in workspace")
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive, "Should use NoFileComp directive")
}

// Note: Testing with actual workspace requires integration tests
// These unit tests verify graceful degradation behavior
