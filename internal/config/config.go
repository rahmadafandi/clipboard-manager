package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	MaxHistory      int    `json:"max_history"`
	AutoExpireHours int    `json:"auto_expire_hours"`
	FilePath        string `json:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		MaxHistory:      50,
		AutoExpireHours: 0,
	}
}

func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "clipboard-manager")
	if err := os.MkdirAll(d, 0755); err != nil {
		return "", err
	}
	return d, nil
}

func Load() (*Config, error) {
	dir, err := configDir()
	if err != nil {
		return DefaultConfig(), nil
	}

	fp := filepath.Join(dir, "config.json")
	cfg := DefaultConfig()
	cfg.FilePath = fp

	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	if cfg.MaxHistory <= 0 {
		cfg.MaxHistory = 50
	}

	return cfg, nil
}

func (c *Config) Save() error {
	if c.FilePath == "" {
		dir, err := configDir()
		if err != nil {
			return err
		}
		c.FilePath = filepath.Join(dir, "config.json")
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.FilePath, data, 0644)
}
