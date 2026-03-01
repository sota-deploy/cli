package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ProjectConfig holds the project linking configuration stored in .sota.json.
type ProjectConfig struct {
	ProjectID   string `json:"project_id"`
	ProjectSlug string `json:"project_slug"`
}

// ConfigDir returns the sota configuration directory (~/.config/sota/).
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sota"
	}
	return filepath.Join(home, ".config", "sota")
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0o700)
}

// LoadProjectConfig reads .sota.json from the current directory.
func LoadProjectConfig() (*ProjectConfig, error) {
	data, err := os.ReadFile(".sota.json")
	if err != nil {
		return nil, fmt.Errorf("no .sota.json found in current directory (run 'sota deploy' to link a project)")
	}

	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid .sota.json: %w", err)
	}

	if cfg.ProjectID == "" {
		return nil, fmt.Errorf(".sota.json missing project_id")
	}

	return &cfg, nil
}

// SaveProjectConfig writes .sota.json to the current directory.
func SaveProjectConfig(cfg *ProjectConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(".sota.json", append(data, '\n'), 0o644)
}
