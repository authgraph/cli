package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the CLI configuration.
type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	OrgID   string `json:"org_id,omitempty"`
}

const defaultBaseURL = "https://api.authgraph.dev"

// Path returns the config file path.
func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "authgraph", "config.json")
}

// Load reads the config from disk.
func Load() (*Config, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in — run `authgraph login` first")
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &cfg, nil
}

// Save writes the config to disk.
func Save(cfg *Config) error {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	dir := filepath.Dir(Path())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(Path(), data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// Delete removes the config file.
func Delete() error {
	return os.Remove(Path())
}
