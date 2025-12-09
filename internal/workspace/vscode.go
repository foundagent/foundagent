package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/errors"
)

// VSCodeWorkspace represents a VS Code workspace file structure
type VSCodeWorkspace struct {
	Folders []VSCodeFolder `json:"folders"`
}

// VSCodeFolder represents a folder in a VS Code workspace
type VSCodeFolder struct {
	Path string `json:"path"`
}

// createVSCodeWorkspace creates the .code-workspace file
func (w *Workspace) createVSCodeWorkspace() error {
	workspace := VSCodeWorkspace{
		Folders: []VSCodeFolder{
			{Path: "."},
		},
	}

	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal VS Code workspace",
			"This is an internal error, please report it",
			err,
		)
	}

	workspacePath := w.VSCodeWorkspacePath()
	if err := os.WriteFile(workspacePath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write VS Code workspace file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}

// LoadVSCodeWorkspace loads the VS Code workspace file
func (w *Workspace) LoadVSCodeWorkspace() (*VSCodeWorkspace, error) {
	workspacePath := w.VSCodeWorkspacePath()
	
	data, err := os.ReadFile(workspacePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(
				errors.ErrCodeFileNotFound,
				"VS Code workspace file not found",
				"Initialize the workspace with 'fa init'",
				err,
			)
		}
		return nil, errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to read VS Code workspace file",
			"Check file permissions",
			err,
		)
	}

	var workspace VSCodeWorkspace
	if err := json.Unmarshal(data, &workspace); err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to parse VS Code workspace file",
			"The workspace file may be corrupted. Try 'fa init --force' to reinitialize",
			err,
		)
	}

	return &workspace, nil
}

// SaveVSCodeWorkspace saves the VS Code workspace file
func (w *Workspace) SaveVSCodeWorkspace(workspace *VSCodeWorkspace) error {
	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal VS Code workspace",
			"This is an internal error, please report it",
			err,
		)
	}

	workspacePath := w.VSCodeWorkspacePath()
	if err := os.WriteFile(workspacePath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write VS Code workspace file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}

// AddWorktreeFolder adds a worktree folder to the VS Code workspace
func (w *Workspace) AddWorktreeFolder(worktreePath string) error {
	workspace, err := w.LoadVSCodeWorkspace()
	if err != nil {
		return err
	}

	// Make path relative to workspace root
	relPath, err := filepath.Rel(w.Path, worktreePath)
	if err != nil {
		// If relative path fails, use absolute path
		relPath = worktreePath
	}

	// Check if folder already exists
	for _, folder := range workspace.Folders {
		if folder.Path == relPath {
			return nil // Already exists
		}
	}

	// Add new folder
	workspace.Folders = append(workspace.Folders, VSCodeFolder{Path: relPath})

	return w.SaveVSCodeWorkspace(workspace)
}
