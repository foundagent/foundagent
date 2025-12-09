package config

import (
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/errors"
	"gopkg.in/yaml.v3"
)

// LoadYAML loads config from a YAML file
func LoadYAML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeConfigNotFound,
			"Failed to read config file",
			"Check that the file exists and is readable",
			err,
		)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeInvalidConfig,
			"Failed to parse YAML config",
			"Check YAML syntax. Run 'fa config validate' for details",
			err,
		)
	}

	return &config, nil
}

// SaveYAML saves config to a YAML file, preserving comments if possible
func SaveYAML(path string, config *Config) error {
	// First, try to load existing file to preserve comments
	existingData, err := os.ReadFile(path)
	if err == nil {
		// Parse existing with yaml.v3 Node to preserve comments
		var root yaml.Node
		if err := yaml.Unmarshal(existingData, &root); err == nil {
			// Update the values while preserving structure
			newData, err := yaml.Marshal(config)
			if err == nil {
				// For now, just write new data - full comment preservation is complex
				// TODO: Implement proper comment-preserving merge
				if err := os.WriteFile(path, newData, 0644); err != nil {
					return errors.Wrap(
						errors.ErrCodePermissionDenied,
						"Failed to write config file",
						"Check file permissions",
						err,
					)
				}
				return nil
			}
		}
	}

	// New file or couldn't preserve - just marshal and write
	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal config to YAML",
			"This is an internal error",
			err,
		)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to create config directory",
			"Check directory permissions",
			err,
		)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to write config file",
			"Check file permissions",
			err,
		)
	}

	return nil
}
