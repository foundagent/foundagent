package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/doctor"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose workspace health",
	Long: `Run diagnostic checks on the workspace to identify issues.

The doctor command checks environment setup, workspace structure, repository
integrity, worktree consistency, and state synchronization.

Examples:
  # Run all checks
  fa doctor

  # Get detailed output
  fa doctor --verbose

  # JSON output for scripts
  fa doctor --json

  # Auto-fix fixable issues
  fa doctor --fix`,
	RunE: runDoctor,
}

var (
	doctorVerbose bool
	doctorJSON    bool
	doctorFix     bool
)

func init() {
	rootCmd.AddCommand(doctorCmd)

	doctorCmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Show detailed check output")
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results as JSON")
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Auto-fix fixable issues")
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Detect workspace
	ws, err := workspace.Discover("")
	if err != nil {
		return fmt.Errorf("not in a Foundagent workspace: %w", err)
	}

	// Build checks
	checks := buildChecks(ws)
	
	// Run checks
	runner := doctor.NewRunner(checks)
	results := runner.Run()
	
	// Apply fixes if requested
	if doctorFix {
		results = applyFixes(ws, results)
	}
	
	// Output results
	if doctorJSON {
		return outputDoctorJSON(results)
	}
	
	return outputDoctorHuman(results)
}

func buildChecks(ws *workspace.Workspace) []doctor.Check {
	return []doctor.Check{
		// Environment checks
		doctor.GitCheck{},
		doctor.GitVersionCheck{},
		
		// Structure checks
		doctor.WorkspaceStructureCheck{Workspace: ws},
		doctor.ConfigValidCheck{Workspace: ws},
		doctor.StateValidCheck{Workspace: ws},
		
		// Repository checks
		doctor.RepositoriesCheck{Workspace: ws},
		doctor.OrphanedReposCheck{Workspace: ws},
		
		// Worktree checks
		doctor.WorktreesCheck{Workspace: ws},
		doctor.OrphanedWorktreesCheck{Workspace: ws},
		
		// Consistency checks
		doctor.ConfigStateConsistencyCheck{Workspace: ws},
		doctor.WorkspaceFileConsistencyCheck{Workspace: ws},
	}
}

func applyFixes(ws *workspace.Workspace, results []doctor.CheckResult) []doctor.CheckResult {
	fixer := doctor.NewFixer(ws)
	fixed := make([]doctor.CheckResult, 0)
	
	for _, result := range results {
		if result.Fixable && result.Status != doctor.StatusPass {
			// Attempt to fix
			fixResult := fixer.Fix(result)
			fixed = append(fixed, fixResult)
		} else {
			fixed = append(fixed, result)
		}
	}
	
	return fixed
}

func outputDoctorJSON(results []doctor.CheckResult) error {
	summary := doctor.CalculateSummary(results)
	
	output := struct {
		Checks  []doctor.CheckResult `json:"checks"`
		Summary doctor.Summary       `json:"summary"`
	}{
		Checks:  results,
		Summary: summary,
	}
	
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputDoctorHuman(results []doctor.CheckResult) error {
	fmt.Println(doctor.FormatResults(results))
	
	summary := doctor.CalculateSummary(results)
	
	fmt.Printf("\nSummary: ")
	if summary.Failed > 0 {
		fmt.Printf("%d passed, %d failed", summary.Passed, summary.Failed)
	} else if summary.Warnings > 0 {
		fmt.Printf("%d passed, %d warnings", summary.Passed, summary.Warnings)
	} else {
		fmt.Printf("All %d checks passed", summary.Passed)
	}
	fmt.Println()
	
	if summary.Failed > 0 {
		return fmt.Errorf("some checks failed")
	}
	
	return nil
}
