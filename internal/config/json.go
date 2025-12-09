package config

import (
	"encoding/json"
	"os"

	"github.com/foundagent/foundagent/internal/errors"
)

// LoadJSON loads config from a JSON file
func LoadJSON(path string) (*Config, error) {
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
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeInvalidConfig,
			"Failed to parse JSON config",
			"Check JSON syntax",
			err,
		)
	}

	return &config, nil
}

// SaveJSON saves config to a JSON file
func SaveJSON(path string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to marshal config to JSON",
			"This is an internal error",
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
