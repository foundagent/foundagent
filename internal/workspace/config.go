package workspace

import (
	"os"

	"github.com/foundagent/foundagent/internal/errors"
	"gopkg.in/yaml.v3"
)

// Config represents the workspace configuration
type Config struct {
	Name  string   `yaml:"name"`
	Repos []string `yaml:"repos"`
}

// createConfig creates the .foundagent.yaml file with default configuration
func (w *Workspace) createConfig() error {
	config := Config{
		Name:  w.Name,
		Repos: []string{},
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal configuration",
			"This is an internal error, please report it",
			err,
		)
	}

	configPath := w.ConfigPath()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write configuration file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}

// LoadConfig loads the workspace configuration
func (w *Workspace) LoadConfig() (*Config, error) {
	configPath := w.ConfigPath()
	
	data, err := os.ReadFile(configPath)
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

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeInvalidConfig,
			"Failed to parse configuration file",
			"Check that the YAML syntax is valid",
			err,
		)
	}

	return &config, nil
}

// SaveConfig saves the workspace configuration
func (w *Workspace) SaveConfig(config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal configuration",
			"This is an internal error, please report it",
			err,
		)
	}

	configPath := w.ConfigPath()
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write configuration file",
			"Check that you have write permissions",
			err,
		)
	}

	return nil
}
