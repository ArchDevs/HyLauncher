package config

import (
	"HyLauncher/internal/env"
	"path/filepath"
)

func launcherPath() string {
	return filepath.Join(env.GetDefaultAppDir(), "config.toml")
}

func LoadLauncher() (*LauncherConfig, error) {
	return load(launcherPath(), LauncherDefault)
}

func SaveLauncher(cfg *LauncherConfig) error {
	return save(launcherPath(), cfg)
}

func UpdateLauncher(update func(*LauncherConfig) error) error {
	cfg, err := LoadLauncher()
	if err != nil {
		return err
	}

	if err := update(cfg); err != nil {
		return err
	}

	return SaveLauncher(cfg)
}
