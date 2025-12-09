package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/foundagent/foundagent/internal/errors"
)

// ConfigFormat represents the configuration file format
type ConfigFormat int

const (
	FormatYAML ConfigFormat = iota
	FormatTOML
	FormatJSON
)

// ConfigFileNames in order of preference
var ConfigFileNames = []string{
	".foundagent.yaml",
	".foundagent.yml",
	".foundagent.toml",
	".foundagent.json",
}

// Load loads configuration from the workspace directory
func Load(workspaceRoot string) (*Config, error) {
	// Try to find config file
	configPath, format, err := FindConfig(workspaceRoot)
	if err != nil {
		return nil, err
	}

	// Load based on format
	var config *Config
	switch format {
	case FormatYAML:
		config, err = LoadYAML(configPath)
	case FormatTOML:
		config, err = LoadTOML(configPath)
	case FormatJSON:
		config, err = LoadJSON(configPath)
	default:
		return nil, errors.New(
			errors.ErrCodeInvalidConfig,
			"Unknown config format",
			"Use .yaml, .toml, or .json extension",
		)
	}

	if err != nil {
		return nil, err
	}

	// Validate config
	if err := Validate(config); err != nil {
		return nil, err
	}

	return config, nil
}

// FindConfig finds the configuration file in the workspace
func FindConfig(workspaceRoot string) (string, ConfigFormat, error) {
	var foundPaths []string
	var foundFormats []ConfigFormat

	for _, name := range ConfigFileNames {
		path := filepath.Join(workspaceRoot, name)
		if _, err := os.Stat(path); err == nil {
			format := detectFormat(name)
			foundPaths = append(foundPaths, path)
			foundFormats = append(foundFormats, format)
		}
	}

	if len(foundPaths) == 0 {
		return "", FormatYAML, errors.New(
			errors.ErrCodeConfigNotFound,
			"Configuration file not found",
			"Run 'fa init' to create a workspace, or create .foundagent.yaml manually",
		)
	}

	if len(foundPaths) > 1 {
		// Warn about multiple configs, use first
		fmt.Fprintf(os.Stderr, "Warning: Multiple config files found. Using %s\n", foundPaths[0])
	}

	return foundPaths[0], foundFormats[0], nil
}

// Save saves configuration to the workspace directory
func Save(workspaceRoot string, config *Config) error {
	// Try to find existing config to preserve format
	configPath, format, err := FindConfig(workspaceRoot)
	if err != nil {
		// No existing config, create YAML
		configPath = filepath.Join(workspaceRoot, ".foundagent.yaml")
		format = FormatYAML
	}

	// Save based on format
	switch format {
	case FormatYAML:
		return SaveYAML(configPath, config)
	case FormatTOML:
		return SaveTOML(configPath, config)
	case FormatJSON:
		return SaveJSON(configPath, config)
	default:
		return errors.New(
			errors.ErrCodeUnknown,
			"Unknown config format",
			"Use .yaml, .toml, or .json extension",
		)
	}
}

func detectFormat(filename string) ConfigFormat {
	ext := filepath.Ext(filename)
	switch ext {
	case ".yaml", ".yml":
		return FormatYAML
	case ".toml":
		return FormatTOML
	case ".json":
		return FormatJSON
	default:
		return FormatYAML
	}
}
