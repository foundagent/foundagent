package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/git"
	"github.com/foundagent/foundagent/internal/output"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var listJSONFlag bool

var listCmd = &cobra.Command{
	Use:     "list [branch]",
	Aliases: []string{"ls"},
	Short:   "List all worktrees in the workspace",
	Long: `List all git worktrees across all repositories in the workspace.

Displays worktrees grouped by branch, showing repo names, paths, and status.
The current/active worktree (if any) is marked with an indicator.

Status indicators:
  * = current worktree (based on working directory)
  [modified] = uncommitted changes
  [untracked] = untracked files
  [conflict] = merge conflicts`,
	Example: `  # List all worktrees
  fa wt list

  # List worktrees for a specific branch
  fa wt list feature-123

  # JSON output for automation
  fa wt list --json

  # Short alias
  fa wt ls`,
	RunE: runList,
}

type worktreeInfo struct {
	Branch     string `json:"branch"`
	Repo       string `json:"repo"`
	Path       string `json:"path"`
	IsCurrent  bool   `json:"is_current"`
	Status     string `json:"status"`
	StatusDesc string `json:"status_description,omitempty"`
}

type listOutput struct {
	WorkspaceName  string         `json:"workspace_name"`
	TotalWorktrees int            `json:"total_worktrees"`
	TotalBranches  int            `json:"total_branches"`
	Worktrees      []worktreeInfo `json:"worktrees"`
}

func init() {
	listCmd.Flags().BoolVar(&listJSONFlag, "json", false, "Output in JSON format")
	worktreeCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Get workspace context
	ws, err := workspace.Discover("")
	if err != nil {
		if listJSONFlag {
			_ = output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Load config
	cfg, err := config.Load(ws.Path)
	if err != nil {
		if listJSONFlag {
			_ = output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Check for empty workspace
	if len(cfg.Repos) == 0 {
		if listJSONFlag {
			return printJSONList(listOutput{
				WorkspaceName:  cfg.Workspace.Name,
				TotalWorktrees: 0,
				TotalBranches:  0,
				Worktrees:      []worktreeInfo{},
			})
		}
		output.PrintErrorMessage("Error: No repositories in workspace")
		output.PrintMessage("Run 'fa add <repo-url>' to add repositories")
		return nil
	}

	// Optional branch filter
	var branchFilter string
	if len(args) > 0 {
		branchFilter = args[0]
	}

	// Discover all worktrees
	allWorktrees, err := discoverWorktrees(ws, cfg, branchFilter)
	if err != nil {
		return err
	}

	if len(allWorktrees) == 0 {
		if listJSONFlag {
			return printJSONList(listOutput{
				WorkspaceName:  cfg.Workspace.Name,
				TotalWorktrees: 0,
				TotalBranches:  0,
				Worktrees:      []worktreeInfo{},
			})
		}
		if branchFilter != "" {
			output.PrintMessage("No worktrees found for branch '%s'", branchFilter)
			output.PrintMessage("Run 'fa wt create %s' to create worktrees for this branch", branchFilter)
		} else {
			output.PrintMessage("No worktrees found")
			output.PrintMessage("Run 'fa wt create <branch>' to create worktrees")
		}
		return nil
	}

	// Get status for all worktrees in parallel
	allWorktrees = detectStatusParallel(allWorktrees)

	// Detect current worktree
	currentPath, _ := os.Getwd()
	allWorktrees = markCurrentWorktree(allWorktrees, currentPath)

	// Output
	if listJSONFlag {
		return printJSONList(buildListOutput(cfg.Workspace.Name, allWorktrees))
	}

	printHumanList(allWorktrees)
	return nil
}

func discoverWorktrees(ws *workspace.Workspace, cfg *config.Config, branchFilter string) ([]worktreeInfo, error) {
	var allWorktrees []worktreeInfo

	for _, repo := range cfg.Repos {
		worktrees, err := workspace.GetWorktreesForRepo(ws.Path, repo.Name)
		if err != nil {
			// Non-fatal: continue with other repos
			continue
		}

		for _, wt := range worktrees {
			// Apply branch filter if specified
			if branchFilter != "" && wt.Branch != branchFilter {
				continue
			}

			allWorktrees = append(allWorktrees, worktreeInfo{
				Branch: wt.Branch,
				Repo:   repo.Name,
				Path:   wt.Path,
				Status: "clean",
			})
		}
	}

	// Sort by branch, then repo
	sort.Slice(allWorktrees, func(i, j int) bool {
		if allWorktrees[i].Branch != allWorktrees[j].Branch {
			return allWorktrees[i].Branch < allWorktrees[j].Branch
		}
		return allWorktrees[i].Repo < allWorktrees[j].Repo
	})

	return allWorktrees, nil
}

func detectStatusParallel(worktrees []worktreeInfo) []worktreeInfo {
	type result struct {
		idx    int
		status string
		desc   string
	}

	results := make(chan result, len(worktrees))

	for idx, wt := range worktrees {
		go func(i int, w worktreeInfo) {
			status, desc := detectWorktreeStatus(w.Path)
			results <- result{idx: i, status: status, desc: desc}
		}(idx, wt)
	}

	// Collect results
	for range worktrees {
		r := <-results
		worktrees[r.idx].Status = r.status
		worktrees[r.idx].StatusDesc = r.desc
	}

	return worktrees
}

func detectWorktreeStatus(path string) (status, description string) {
	// Check if path exists
	if !workspace.PathExists(path) {
		return "error", "worktree path not found"
	}

	// Get git status
	hasChanges, err := git.HasUncommittedChanges(path)
	if err != nil {
		return "error", "failed to check status"
	}

	if hasChanges {
		// Determine type of changes
		hasUntracked, _ := git.HasUntrackedFiles(path)
		hasConflicts, _ := git.HasConflicts(path)

		if hasConflicts {
			return "conflict", "merge conflicts present"
		}
		if hasUntracked {
			return "untracked", "untracked files present"
		}
		return "modified", "uncommitted changes"
	}

	return "clean", ""
}

func markCurrentWorktree(worktrees []worktreeInfo, currentPath string) []worktreeInfo {
	for i := range worktrees {
		// Check if current path is within this worktree
		if strings.HasPrefix(currentPath, worktrees[i].Path) {
			worktrees[i].IsCurrent = true
			break
		}
	}
	return worktrees
}

func printHumanList(worktrees []worktreeInfo) {
	// Group by branch
	branchMap := make(map[string][]worktreeInfo)
	var branches []string

	for _, wt := range worktrees {
		if _, exists := branchMap[wt.Branch]; !exists {
			branches = append(branches, wt.Branch)
		}
		branchMap[wt.Branch] = append(branchMap[wt.Branch], wt)
	}

	// Display grouped by branch
	for _, branch := range branches {
		fmt.Printf("\n\033[1m%s\033[0m\n", branch) // Bold
		for _, wt := range branchMap[branch] {
			marker := "  "
			if wt.IsCurrent {
				marker = "* "
			}

			statusStr := ""
			if wt.Status != "clean" {
				statusStr = fmt.Sprintf(" \033[33m[%s]\033[0m", wt.Status) // Yellow
			}

			fmt.Printf("%s%s: %s%s\n", marker, wt.Repo, wt.Path, statusStr)
		}
	}
	fmt.Println()
}

func buildListOutput(workspaceName string, worktrees []worktreeInfo) listOutput {
	// Count unique branches
	branchSet := make(map[string]bool)
	for _, wt := range worktrees {
		branchSet[wt.Branch] = true
	}

	return listOutput{
		WorkspaceName:  workspaceName,
		TotalWorktrees: len(worktrees),
		TotalBranches:  len(branchSet),
		Worktrees:      worktrees,
	}
}

func printJSONList(output listOutput) error {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
