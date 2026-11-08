package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	WikiPath string `json:"wiki_path"`
}

// DataDir returns the standalone directory used for topics and review history.
func DataDir() (string, error) {
	if base := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(base) {
		return filepath.Join(base, "recall"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "recall"), nil
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "recall", "config.json")
}

func Load() (*Config, error) {
	path := DefaultConfigPath()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil // No config yet
	}
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
