package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	WikiPath string `json:"wiki_path"`
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

func Save(cfg *Config) error {
	path := DefaultConfigPath()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func GetWikiPath() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	if cfg == nil || cfg.WikiPath == "" {
		return "", nil
	}
	return cfg.WikiPath, nil
}
