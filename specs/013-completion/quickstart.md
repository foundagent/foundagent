# Quickstart: Shell Completion

**Feature Branch**: `013-completion`  
**Date**: 2025-12-08

## What You'll Build

A `fa completion` command that generates shell completion scripts for Bash, Zsh, Fish, and PowerShell, enabling tab completion for all Foundagent commands and arguments.

## Prerequisites

- Go 1.21+
- Cobra CLI library (already in go.mod)
- Existing command structure from previous features

## Implementation Steps

### Step 1: Create the Completion Command

Create `internal/cli/completion.go`:

```go
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var supportedShells = []string{"bash", "zsh", "fish", "powershell"}

var completionCmd = &cobra.Command{
	Use:   "completion <shell>",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for Foundagent.

Supported shells: bash, zsh, fish, powershell

Installation instructions are included in the generated script.`,
	Example: `  # Bash - source for current session
  source <(fa completion bash)

  # Zsh - add to fpath
  fa completion zsh > "${fpath[1]}/_fa"

  # Fish - save to completions
  fa completion fish > ~/.config/fish/completions/fa.fish

  # PowerShell - add to profile
  fa completion powershell >> $PROFILE`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: supportedShells,
	RunE:      runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]
	rootCmd := cmd.Root()

	switch shell {
	case "bash":
		return rootCmd.GenBashCompletionV2(os.Stdout, true)
	case "zsh":
		return rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell %q\nSupported shells: bash, zsh, fish, powershell", shell)
	}
}
```

### Step 2: Add Dynamic Completions to Existing Commands

Update worktree switch command (`internal/cli/worktree_switch.go`):

```go
var worktreeSwitchCmd = &cobra.Command{
	Use:   "switch <worktree>",
	Short: "Switch to a worktree",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getWorktreeCompletions(toComplete), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: runWorktreeSwitch,
}
```

### Step 3: Implement Completion Helpers

Create `internal/cli/completion_helpers.go`:

```go
package cli

import (
	"strings"

	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

// getWorktreeCompletions returns worktree names for completion
func getWorktreeCompletions(toComplete string) []cobra.Completion {
	ws, err := workspace.Discover()
	if err != nil {
		return nil // Not in workspace, return empty
	}

	worktrees, err := ws.ListWorktrees()
	if err != nil {
		return nil
	}

	var completions []cobra.Completion
	for _, wt := range worktrees {
		if strings.HasPrefix(wt.Name, toComplete) {
			completions = append(completions, cobra.CompletionWithDesc(wt.Name, "worktree"))
		}
	}
	return completions
}

// getRepoCompletions returns repository names for completion
func getRepoCompletions(toComplete string) []cobra.Completion {
	ws, err := workspace.Discover()
	if err != nil {
		return nil
	}

	repos, err := ws.ListRepos()
	if err != nil {
		return nil
	}

	var completions []cobra.Completion
	for _, repo := range repos {
		if strings.HasPrefix(repo.Name, toComplete) {
			completions = append(completions, cobra.CompletionWithDesc(repo.Name, "repository"))
		}
	}
	return completions
}
```

### Step 4: Write Tests

Create `internal/cli/completion_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCommand(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		wantErr   bool
		contains  string
	}{
		{
			name:     "bash generates valid script",
			shell:    "bash",
			wantErr:  false,
			contains: "complete",
		},
		{
			name:     "zsh generates valid script",
			shell:    "zsh",
			wantErr:  false,
			contains: "#compdef",
		},
		{
			name:     "fish generates valid script",
			shell:    "fish",
			wantErr:  false,
			contains: "complete -c",
		},
		{
			name:     "powershell generates valid script",
			shell:    "powershell",
			wantErr:  false,
			contains: "Register-ArgumentCompleter",
		},
		{
			name:    "invalid shell returns error",
			shell:   "tcsh",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			rootCmd.SetArgs([]string{"completion", tt.shell})

			err := rootCmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, strings.Contains(stdout.String(), tt.contains),
					"expected output to contain %q", tt.contains)
			}
		})
	}
}
```

### Step 5: Run and Verify

```bash
# Build
make build

# Test completion generation
./bin/fa completion bash | head -20
./bin/fa completion zsh | head -20

# Test that completions work
source <(./bin/fa completion bash)
./bin/fa <TAB>  # Should show commands

# Run tests
make test
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/cli/completion.go` | Main completion command |
| `internal/cli/completion_helpers.go` | Dynamic completion functions |
| `internal/cli/completion_test.go` | Tests |

## Common Issues

| Issue | Solution |
|-------|----------|
| Completions not showing | Ensure script is sourced correctly |
| Dynamic completions empty | Verify you're in a Foundagent workspace |
| Alias `fa` not completing | Add alias completion per shell instructions |

## Next Steps

After implementing the basic completion:

1. Add installation instructions header to each shell's output
2. Test on all supported platforms
3. Add to CI for syntax validation
4. Update documentation with installation guides
