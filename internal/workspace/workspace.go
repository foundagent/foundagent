package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/errors"
)

const (
	// FoundagentDir is the directory for machine-managed state
	FoundagentDir = ".foundagent"

	// ConfigFileName is the name of the config file
	ConfigFileName = ".foundagent.yaml"

	// StateFileName is the name of the state file
	StateFileName = "state.json"

	// ReposDir is the directory for repository storage
	ReposDir = "repos"

	// BareDir is the subdirectory for bare clones
	BareDir = ".bare"

	// WorktreesDir is the subdirectory for worktrees
	WorktreesDir = "worktrees"
)

// Workspace represents a Foundagent workspace
type Workspace struct {
	Name string
	Path string
}

// New creates a new workspace instance
func New(name, basePath string) (*Workspace, error) {
	// Validate the name
	if err := ValidateName(name); err != nil {
		return nil, err
	}

	// Create absolute path
	absPath, err := filepath.Abs(filepath.Join(basePath, name))
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to resolve absolute path",
			"Ensure the base path is valid and accessible",
			err,
		)
	}

	return &Workspace{
		Name: name,
		Path: absPath,
	}, nil
}

// Exists checks if the workspace already exists
func (w *Workspace) Exists() bool {
	foundagentPath := filepath.Join(w.Path, FoundagentDir)
	info, err := os.Stat(foundagentPath)
	return err == nil && info.IsDir()
}

// Create creates the workspace directory structure
func (w *Workspace) Create(force bool) error {
	// Check if workspace exists
	if w.Exists() && !force {
		return errors.New(
			errors.ErrCodeWorkspaceExists,
			fmt.Sprintf("Workspace already exists at %s", w.Path),
			"Use --force to reinitialize the workspace",
		)
	}

	// Create main workspace directory
	if err := os.MkdirAll(w.Path, 0755); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			fmt.Sprintf("Failed to create workspace directory: %s", w.Path),
			"Check that you have write permissions in the parent directory",
			err,
		)
	}

	// Create .foundagent directory
	foundagentPath := filepath.Join(w.Path, FoundagentDir)
	if err := os.MkdirAll(foundagentPath, 0755); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			fmt.Sprintf("Failed to create .foundagent directory: %s", foundagentPath),
			"Check filesystem permissions",
			err,
		)
	}

	// Create repos directory structure
	if err := w.createReposStructure(force); err != nil {
		return err
	}

	// Create config file
	if err := w.createConfig(); err != nil {
		return err
	}

	// Create state file
	if err := w.createState(); err != nil {
		return err
	}

	// Create VS Code workspace file
	if err := w.createVSCodeWorkspace(); err != nil {
		return err
	}

	return nil
}

// createReposStructure creates the repos directory structure
func (w *Workspace) createReposStructure(force bool) error {
	reposPath := filepath.Join(w.Path, ReposDir)

	// If force mode, preserve existing repos directory
	if force {
		if info, err := os.Stat(reposPath); err == nil && info.IsDir() {
			// Repos directory exists, skip creation
			return nil
		}
	}

	// Create repos directory
	// Note: Individual repo directories (repos/<repo-name>/) are created when repos are added
	// Each repo will have its own .bare/ and worktrees/ subdirectories
	if err := os.MkdirAll(reposPath, 0755); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			fmt.Sprintf("Failed to create repos directory: %s", reposPath),
			"Check filesystem permissions",
			err,
		)
	}

	return nil
}

// ConfigPath returns the path to the config file
func (w *Workspace) ConfigPath() string {
	return filepath.Join(w.Path, ConfigFileName)
}

// StatePath returns the path to the state file
func (w *Workspace) StatePath() string {
	return filepath.Join(w.Path, FoundagentDir, StateFileName)
}

// VSCodeWorkspacePath returns the path to the VS Code workspace file
func (w *Workspace) VSCodeWorkspacePath() string {
	return filepath.Join(w.Path, w.Name+".code-workspace")
}
