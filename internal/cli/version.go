package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/foundagent/foundagent/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information for Foundagent.

By default, shows the version number. Use flags for detailed build info,
JSON output, or to check for updates.

Examples:
  # Show version
  fa version

  # Show detailed build information
  fa version --full

  # JSON output for scripts
  fa version --json

  # Check for updates
  fa version --check`,
	RunE: runVersion,
}

var (
	versionFull  bool
	versionJSON  bool
	versionCheck bool
)

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVar(&versionFull, "full", false, "Show full build information")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Output in JSON format")
	versionCmd.Flags().BoolVar(&versionCheck, "check", false, "Check for updates")
}

func runVersion(cmd *cobra.Command, args []string) error {
	// Handle update check
	if versionCheck {
		return checkForUpdates()
	}

	// Handle JSON output
	if versionJSON {
		return outputVersionJSON()
	}

	// Handle full output
	if versionFull {
		fmt.Println(version.Full())
		return nil
	}

	// Default: simple version
	fmt.Println(version.String())
	return nil
}

func outputVersionJSON() error {
	info := version.Get()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}

func checkForUpdates() error {
	// Show current version first
	fmt.Println(version.String())
	fmt.Println()
	
	// Check for updates with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*1000*1000*1000) // 5 seconds
	defer cancel()

	updateAvailable, latestVersion, downloadURL, err := version.CheckForUpdate(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to check for updates: %v\n", err)
		return nil
	}

	if updateAvailable {
		fmt.Printf("Update available: v%s\n", latestVersion)
		fmt.Printf("Download: %s\n", downloadURL)
	} else {
		fmt.Println("You're up to date!")
	}

	return nil
}
