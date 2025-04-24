package framefold

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	FolderTemplate  string              `json:"folder_template"`
	MediaTypes      map[string][]string `json:"media_types"`
	UseOriginalName bool                `json:"use_original_filename"`
	Logging         LoggingConfig       `json:"logging"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Enabled bool   `json:"enabled"`
	Level   string `json:"level"`
}

// DefaultConfig provides default configuration values
var DefaultConfig = Config{
	FolderTemplate: "{{.Year}}/{{.Month}}",
	MediaTypes: map[string][]string{
		"images": {".jpg", ".jpeg", ".png", ".gif", ".heic"},
		"videos": {".mp4", ".mov", ".avi"},
	},
	UseOriginalName: true,
	Logging: LoggingConfig{
		Enabled: true,
		Level:   "info",
	},
}

// LoadConfig loads configuration from a file, falling back to defaults if no file is specified
func LoadConfig(configPath string) (Config, error) {
	config := DefaultConfig

	// If no config file specified, use default configuration
	if configPath == "" {
		return config, nil
	}

	// Try to read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %v", err)
	}

	// Override defaults with values from config file
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("error parsing config file: %v", err)
	}

	return config, nil
}
