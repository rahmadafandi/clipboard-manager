package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	MaxHistory      int    `json:"max_history"`
	AutoExpireHours int    `json:"auto_expire_hours"`
	PreviewLines    int    `json:"preview_lines"`
	PreviewWidth    int    `json:"preview_width"`
	FilePath        string `json:"-"`
}

func DefaultConfig() *Config {
	return &Config{
		MaxHistory:      50,
		AutoExpireHours: 0, // 0 = disabled
		PreviewLines:    1,
		PreviewWidth:    80,
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

	// Enforce minimums
	if cfg.MaxHistory <= 0 {
		cfg.MaxHistory = 50
	}
	if cfg.PreviewLines <= 0 {
		cfg.PreviewLines = 1
	}
	if cfg.PreviewWidth <= 0 {
		cfg.PreviewWidth = 80
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
