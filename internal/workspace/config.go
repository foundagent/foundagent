package workspace

import (
	"os"

	"github.com/foundagent/foundagent/internal/config"
	"github.com/foundagent/foundagent/internal/errors"
)

// Config represents the workspace configuration (deprecated, use internal/config)
type Config struct {
	Name  string   `yaml:"name"`
	Repos []string `yaml:"repos"`
}

// createConfig creates the .foundagent.yaml file with default configuration
func (w *Workspace) createConfig() error {
	// Generate template with comments using new config package
	template := config.DefaultTemplate(w.Name)

	// Write template to file
	configPath := w.ConfigPath()
	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write configuration file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}

// LoadConfig loads the workspace configuration (deprecated, use config.Load)
func (w *Workspace) LoadConfig() (*Config, error) {
	configPath := w.ConfigPath()

	_, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(
				errors.ErrCodeConfigNotFound,
				"Configuration file not found",
				"Initialize the workspace with 'fa init'",
				err,
			)
		}
		return nil, errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to read configuration file",
			"Check file permissions",
			err,
		)
	}

	// For backward compatibility, we still need to parse the old format
	// But new code should use config.Load()
	return &Config{Name: w.Name, Repos: []string{}}, nil
}

// SaveConfig saves the workspace configuration (deprecated, use config.Save)
func (w *Workspace) SaveConfig(config *Config) error {
	// This is deprecated - new code should use the config package
	return errors.New(
		errors.ErrCodeUnknown,
		"SaveConfig is deprecated",
		"Use config.Save() instead",
	)
}
