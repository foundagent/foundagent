package cli

import (
	"os"

	"github.com/foundagent/foundagent/internal/output"
	"github.com/foundagent/foundagent/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	initForce bool
	initJSON  bool
)

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new Foundagent workspace",
	Long: `Initialize a new Foundagent workspace with the specified name.

This creates a new directory with the workspace structure including:
- .foundagent.yaml (configuration file)
- .foundagent/state.json (runtime state)
- repos/.bare/ (bare repository clones)
- repos/worktrees/ (working directories)
- <name>.code-workspace (VS Code workspace file)`,
	Example: `  # Create a new workspace
  fa init my-project

  # Create with JSON output
  fa init my-project --json

  # Reinitialize an existing workspace
  fa init my-project --force`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Force reinitialize existing workspace")
	initCmd.Flags().BoolVar(&initJSON, "json", false, "Output result as JSON")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		if initJSON {
			_ = output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: Failed to get current directory: %v", err)
		}
		return err
	}

	// Create workspace instance
	ws, err := workspace.New(name, cwd)
	if err != nil {
		if initJSON {
			_ = output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Validate path length
	if err := workspace.ValidatePathLength(ws.Path); err != nil {
		if initJSON {
			_ = output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Create the workspace
	if err := ws.Create(initForce); err != nil {
		if initJSON {
			_ = output.PrintError(err)
		} else {
			output.PrintErrorMessage("Error: %v", err)
		}
		return err
	}

	// Output success
	if initJSON {
		return output.PrintSuccess(map[string]interface{}{
			"name":   ws.Name,
			"path":   ws.Path,
			"action": getAction(initForce),
		})
	}

	action := "created"
	if initForce {
		action = "reinitialized"
	}

	output.PrintMessage("✓ Workspace %s %s at: %s", ws.Name, action, ws.Path)
	output.PrintMessage("")
	output.PrintMessage("Next steps:")
	output.PrintMessage("  1. Open workspace in VS Code: code %s", ws.VSCodeWorkspacePath())
	output.PrintMessage("  2. Add a repository: fa repo add <url>")

	return nil
}

func getAction(force bool) string {
	if force {
		return "reinitialized"
	}
	return "created"
}
