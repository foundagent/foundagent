package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/foundagent/foundagent/internal/errors"
)

// LoadTOML loads config from a TOML file
func LoadTOML(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, errors.Wrap(
			errors.ErrCodeInvalidConfig,
			"Failed to parse TOML config",
			"Check TOML syntax",
			err,
		)
	}

	return &config, nil
}

// SaveTOML saves config to a TOML file
func SaveTOML(path string, config *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(
			errors.ErrCodePermissionDenied,
			"Failed to create config file",
			"Check file permissions",
			err,
		)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return errors.Wrap(
			errors.ErrCodeUnknown,
			"Failed to encode config to TOML",
			"This is an internal error",
			err,
		)
	}

	return nil
}
