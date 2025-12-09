package workspace

import (
	"encoding/json"
	"os"

	"github.com/foundagent/foundagent/internal/errors"
)

// State represents the workspace runtime state
type State struct {
	Repositories map[string]*Repository `json:"repositories,omitempty"`
}

// createState creates the state.json file with initial empty state
func (w *Workspace) createState() error {
	// Initialize with empty JSON object
	state := State{}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal state",
			"This is an internal error, please report it",
			err,
		)
	}

	statePath := w.StatePath()
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write state file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}

// LoadState loads the workspace state
func (w *Workspace) LoadState() (*State, error) {
	statePath := w.StatePath()
	
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(
				errors.ErrCodeFileNotFound,
				"State file not found",
				"Initialize the workspace with 'fa init'",
				err,
			)
		}
		return nil, errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to read state file",
			"Check file permissions",
			err,
		)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to parse state file",
			"The state file may be corrupted. Try 'fa init --force' to reinitialize",
			err,
		)
	}

	return &state, nil
}

// SaveState saves the workspace state
func (w *Workspace) SaveState(state *State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal state",
			"This is an internal error, please report it",
			err,
		)
	}

	statePath := w.StatePath()
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write state file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}
