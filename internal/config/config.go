package config

import (
	"os"
	"path/filepath"

	"HyLauncher/internal/env"

	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
)

func configPath() string {
	return filepath.Join(env.GetDefaultAppDir(), "config.toml")
}

func Save(cfg *Config) error {
	path := configPath()
	// Create config if not exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Serialize config
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Load() (*Config, error) {
	path := configPath()

	// Get config data
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			_ = Save(&cfg)
			return &cfg, nil
		}
		return nil, err
	}

	// Check for migration
	var migrationData struct {
		Nick string `toml:"nick"`
	}
	_ = toml.Unmarshal(data, &migrationData)

	cfg := Default()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		_ = os.Rename(path, path+".broken")

		cfg = Default()
		_ = Save(&cfg)
		return &cfg, nil
	}

	// Migration: If no profiles but has a legacy nick, create a default profile
	if len(cfg.Profiles) == 0 && migrationData.Nick != "" {
		id := uuid.New().String()
		cfg.Profiles = []Profile{
			{
				ID:   id,
				Name: migrationData.Nick,
			},
		}
		cfg.CurrentProfile = id
		_ = Save(&cfg)
	}

	return &cfg, nil
}
