package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/git"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	switchCreate bool
	switchFrom   string
	switchQuiet  bool
	switchJSON   bool
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [branch]",
	Short: "Switch to a different branch's worktrees",
	Long: `Switch updates your VS Code workspace to show a different branch's worktrees.

This command updates the .code-workspace file to point to the target branch's
worktrees across all repositories. Your current worktrees remain unchanged.

Examples:
  # Switch to feature-123 branch
  fa wt switch feature-123

  # List available branches to switch to
  fa wt switch

  # Create worktrees and switch if they don't exist
  fa wt switch new-feature --create

  # Create from a specific branch
  fa wt switch hotfix --create --from release-1.0

  # Switch with JSON output
  fa wt switch feature-123 --json`,
	RunE: runSwitch,
}

func init() {
	worktreeCmd.AddCommand(switchCmd)

	// Flags
	switchCmd.Flags().BoolVarP(&switchCreate, "create", "c", false, "Create worktrees if they don't exist")
	switchCmd.Flags().StringVarP(&switchFrom, "from", "f", "", "Source branch for new worktrees (only with --create)")
	switchCmd.Flags().BoolVarP(&switchQuiet, "quiet", "q", false, "Suppress warnings")
	switchCmd.Flags().BoolVar(&switchJSON, "json", false, "Output in JSON format")
}

func runSwitch(cmd *cobra.Command, args []string) error {
	// Validate --from flag only used with --create
	if switchFrom != "" && !switchCreate {
		return fmt.Errorf("--from flag can only be used with --create flag")
	}

	// Load workspace (validates we're in a Foundagent workspace)
	ws, err := workspace.Discover("")
	if err != nil {
		return err
	}

	// Check if workspace has any repositories configured
	state, err := ws.LoadState()
	if err != nil {
		return err
	}
	if state.Repositories == nil || len(state.Repositories) == 0 {
		return fmt.Errorf("No repositories configured in workspace. Add repositories with 'fa add <url> <name>'")
	}

	// If no branch specified, list available branches (US5)
	if len(args) == 0 {
		return listAvailableBranches(ws)
	}

	targetBranch := args[0]

	// Validate branch name
	if err := git.ValidateBranchName(targetBranch); err != nil {
		return err
	}

	// Get current branch
	currentBranch, err := ws.GetCurrentBranchFromWorkspace()
	if err != nil {
		// If no current branch detected, proceed anyway
		currentBranch = ""
	}

	// Check if already on target branch
	if currentBranch == targetBranch {
		if switchJSON {
			return outputJSON(map[string]interface{}{
				"switched_to":     targetBranch,
				"previous_branch": currentBranch,
				"already_on":      true,
				"workspace_file":  ws.VSCodeWorkspacePath(),
			})
		}
		fmt.Printf("Already on branch '%s'\n", targetBranch)
		return nil
	}

	// Check if target branch has worktrees and detect partial worktrees
	allWorktrees, err := ws.GetAllWorktrees()
	if err != nil {
		return err
	}

	// Count how many repos have this branch
	reposWithBranch := 0
	totalRepos := len(state.Repositories)
	missingRepos := make([]string, 0)
	
	for repoName := range state.Repositories {
		hasBranch := false
		if branches, ok := allWorktrees[repoName]; ok {
			for _, branch := range branches {
				if branch == targetBranch {
					hasBranch = true
					break
				}
			}
		}
		if hasBranch {
			reposWithBranch++
		} else {
			missingRepos = append(missingRepos, repoName)
		}
	}

	hasWorktrees := reposWithBranch > 0
	isPartial := reposWithBranch > 0 && reposWithBranch < totalRepos

	// If worktrees don't exist and --create not specified
	if !hasWorktrees && !switchCreate {
		return fmt.Errorf("No worktrees found for branch '%s'. Use --create to create them", targetBranch)
	}

	// If --create specified, create worktrees first (US3)
	if switchCreate && !hasWorktrees {
		if err := createWorktreesForBranch(ws, targetBranch); err != nil {
			return err
		}
	} else if switchCreate && isPartial {
		// Create missing worktrees for partial branch
		if err := createMissingWorktrees(ws, targetBranch, missingRepos); err != nil {
			return err
		}
	} else if isPartial && !switchQuiet {
		// Warn about partial worktrees
		fmt.Printf("\n⚠️  Warning: Branch '%s' only exists in %d of %d repositories\n", targetBranch, reposWithBranch, totalRepos)
		fmt.Println("Missing in:")
		for _, repo := range missingRepos {
			fmt.Printf("  - %s\n", repo)
		}
		fmt.Println()
	}

	// Warn about uncommitted changes (US2)
	if currentBranch != "" && !switchQuiet {
		if err := warnUncommittedChanges(ws, currentBranch); err != nil {
			// Warning errors are non-fatal, just print them
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", err)
		}
	}

	// Replace worktree folders in workspace file
	if err := ws.ReplaceWorktreeFolders(targetBranch); err != nil {
		return err
	}

	// Update state to track current branch
	state.CurrentBranch = targetBranch
	if err := ws.SaveState(state); err != nil {
		return err
	}

	// Output result
	if switchJSON {
		return outputJSON(map[string]interface{}{
			"switched_to":     targetBranch,
			"previous_branch": currentBranch,
			"workspace_file":  ws.VSCodeWorkspacePath(),
		})
	}

	fmt.Printf("✓ Switched to branch '%s'\n", targetBranch)
	fmt.Printf("Workspace file: %s\n", ws.VSCodeWorkspacePath())
	fmt.Println("\nReload VS Code to see the new worktrees")

	return nil
}

// listAvailableBranches lists all branches that have worktrees (US5)
func listAvailableBranches(ws *workspace.Workspace) error {
	branches, err := ws.GetAvailableBranches()
	if err != nil {
		return err
	}

	if len(branches) == 0 {
		fmt.Println("No branches with worktrees found")
		fmt.Println("Create worktrees with: fa wt create <branch>")
		return nil
	}

	// Get current branch
	currentBranch, _ := ws.GetCurrentBranchFromWorkspace()

	if switchJSON {
		return outputJSON(map[string]interface{}{
			"branches":       branches,
			"current_branch": currentBranch,
		})
	}

	fmt.Println("Available branches:")
	for _, branch := range branches {
		marker := " "
		if branch == currentBranch {
			marker = "*"
		}
		fmt.Printf(" %s %s\n", marker, branch)
	}

	return nil
}

// warnUncommittedChanges checks for uncommitted changes and warns user (US2)
func warnUncommittedChanges(ws *workspace.Workspace, branch string) error {
	// Get all worktrees for current branch
	state, err := ws.LoadState()
	if err != nil {
		return err
	}

	dirtyWorktrees := make([]string, 0)

	for repoName := range state.Repositories {
		worktreePath := ws.WorktreePath(repoName, branch)
		
		// Check if worktree exists
		if _, err := os.Stat(worktreePath); err != nil {
			continue
		}

		// Check for uncommitted changes
		hasChanges, err := git.HasUncommittedChanges(worktreePath)
		if err != nil {
			continue
		}

		if hasChanges {
			dirtyWorktrees = append(dirtyWorktrees, fmt.Sprintf("  - %s/%s", repoName, branch))
		}
	}

	if len(dirtyWorktrees) > 0 {
		fmt.Println("\n⚠️  Warning: Uncommitted changes in current worktrees:")
		for _, wt := range dirtyWorktrees {
			fmt.Println(wt)
		}
		fmt.Println("\nThese changes will remain in the original worktrees.")
		fmt.Println("")
	}

	return nil
}

// createWorktreesForBranch creates worktrees for the target branch (US3)
func createWorktreesForBranch(ws *workspace.Workspace, branch string) error {
	// Get all repositories
	state, err := ws.LoadState()
	if err != nil {
		return err
	}

	if state.Repositories == nil || len(state.Repositories) == 0 {
		return fmt.Errorf("No repositories configured in workspace")
	}

	sourceBranch := switchFrom
	if sourceBranch == "" {
		// Use first repo's default branch as source
		for _, repo := range state.Repositories {
			if repo.DefaultBranch != "" {
				sourceBranch = repo.DefaultBranch
			} else {
				sourceBranch = "main"
			}
			break
		}
	}

	fmt.Printf("Creating worktrees for branch '%s' from '%s'...\n", branch, sourceBranch)

	// Create worktrees for all repos
	for repoName, repo := range state.Repositories {
		bareRepoPath := ws.BareRepoPath(repoName)
		worktreePath := ws.WorktreePath(repoName, branch)

		// Check if worktree already exists
		if exists, _ := ws.WorktreeExists(repoName, branch); exists {
			continue
		}

		// Create the worktree with new branch
		if err := git.WorktreeAddNew(bareRepoPath, worktreePath, branch, sourceBranch); err != nil {
			return fmt.Errorf("Failed to create worktree for %s: %w", repoName, err)
		}

		// Add worktree to VS Code workspace
		if err := ws.AddWorktreeFolder(worktreePath); err != nil {
			return fmt.Errorf("Failed to add worktree to workspace: %w", err)
		}

		// Update repository state
		repo.Worktrees = append(repo.Worktrees, branch)
	}

	// Save updated state
	if err := ws.SaveState(state); err != nil {
		return err
	}

	fmt.Printf("✓ Created worktrees for branch '%s'\n", branch)
	return nil
}

// createMissingWorktrees creates worktrees only for repos that are missing them
func createMissingWorktrees(ws *workspace.Workspace, branch string, missingRepos []string) error {
	state, err := ws.LoadState()
	if err != nil {
		return err
	}

	sourceBranch := switchFrom
	if sourceBranch == "" {
		// Use first repo's default branch as source
		for _, repo := range state.Repositories {
			if repo.DefaultBranch != "" {
				sourceBranch = repo.DefaultBranch
			} else {
				sourceBranch = "main"
			}
			break
		}
	}

	fmt.Printf("Creating missing worktrees for branch '%s' from '%s'...\n", branch, sourceBranch)

	for _, repoName := range missingRepos {
		repo := state.Repositories[repoName]
		bareRepoPath := ws.BareRepoPath(repoName)
		worktreePath := ws.WorktreePath(repoName, branch)

		// Create the worktree with new branch
		if err := git.WorktreeAddNew(bareRepoPath, worktreePath, branch, sourceBranch); err != nil {
			return fmt.Errorf("Failed to create worktree for %s: %w", repoName, err)
		}

		// Add worktree to VS Code workspace
		if err := ws.AddWorktreeFolder(worktreePath); err != nil {
			return fmt.Errorf("Failed to add worktree to workspace: %w", err)
		}

		// Update repository state
		repo.Worktrees = append(repo.Worktrees, branch)
	}

	// Save updated state
	if err := ws.SaveState(state); err != nil {
		return err
	}

	fmt.Printf("✓ Created missing worktrees for branch '%s'\n", branch)
	return nil
}

// outputJSON outputs data in JSON format
func outputJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}
