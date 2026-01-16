package config

import (
	"os"
	"path/filepath"

	"HyLauncher/internal/env"

	"github.com/pelletier/go-toml/v2"
)

func configPath() string {
	return filepath.Join(env.GetDefaultAppDir(), "config.toml")
}

func Save(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Load() (*Config, error) {
	path := configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			_ = Save(&cfg)
			return &cfg, nil
		}
		return nil, err
	}

	cfg := Default()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		_ = os.Rename(path, path+".broken")

		cfg = Default()
		_ = Save(&cfg)
		return &cfg, nil
	}

	return &cfg, nil
}
