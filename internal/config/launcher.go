package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func getDefaultLauncherConfigDir() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Local", "HyLauncher")
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "HyLauncher")
	case "linux":
		return filepath.Join(home, ".hylauncher")
	default:
		return filepath.Join(home, "HyLauncher")
	}
}

func launcherPath() string {
	return filepath.Join(getDefaultLauncherConfigDir(), "config.toml")
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
