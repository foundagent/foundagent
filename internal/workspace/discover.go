package workspace

import (
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/errors"
)

// Discover finds the workspace root by searching for .foundagent.yaml
func Discover(startPath string) (*Workspace, error) {
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return nil, errors.Wrap(
				errors.ErrCodeUnknown,
				"Failed to get current directory",
				"Ensure you have permission to access the current directory",
				err,
			)
		}
	}

	// Make path absolute
	startPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to resolve absolute path",
			"Ensure the path is valid",
			err,
		)
	}

	// Walk up the directory tree
	current := startPath
	for {
		configPath := filepath.Join(current, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			// Found workspace root - load config to get name
			ws := &Workspace{
				Path: current,
				Name: filepath.Base(current),
			}

			// Try to load the actual name from config
			if config, err := ws.LoadConfig(); err == nil {
				ws.Name = config.Name
			}

			return ws, nil
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root without finding workspace
			break
		}
		current = parent
	}

	return nil, errors.New(
		errors.ErrCodeConfigNotFound,
		"Not in a Foundagent workspace",
		"Run 'fa init <name>' to create a new workspace, or navigate to an existing workspace directory",
	)
}

// MustDiscover finds the workspace or panics
func MustDiscover(startPath string) *Workspace {
	ws, err := Discover(startPath)
	if err != nil {
		panic(err)
	}
	return ws
}
