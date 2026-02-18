package env

import (
	"HyLauncher/pkg/logger"
	"HyLauncher/pkg/model"
	"fmt"
	"os"
	"path/filepath"
)

func CleanupLauncher(request model.InstanceModel) error {
	cacheDir := GetCacheDir()

	if err := cleanDirectory(cacheDir, []string{".pwr", ".zip", ".tar.gz"}); err != nil {
		logger.Warn("Failed to clean cache", "error", err)
	}

	gameLatest := GetGameDir(request.Branch, request.BuildVersion)
	if err := cleanIncompleteGame(gameLatest); err != nil {
		logger.Warn("Failed to clean game directory", "error", err)
	}

	stagingDir := filepath.Join(gameLatest, "staging-temp")
	if err := os.RemoveAll(stagingDir); err != nil {
		logger.Warn("Failed to remove staging dir", "error", err)
	}

	// Clean up old launcher backup from updates
	if err := cleanupLauncherBackup(); err != nil {
		logger.Warn("Failed to clean launcher backup", "error", err)
	}

	return nil
}

func cleanupLauncherBackup() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	backup := exe + ".old"

	// Check if backup exists
	if _, err := os.Stat(backup); os.IsNotExist(err) {
		return nil // No backup to clean
	}

	// Remove the backup
	logger.Info("Removing old launcher backup", "path", backup)
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("failed to remove backup: %w", err)
	}

	logger.Info("Old launcher backup removed successfully")
	return nil
}

func cleanDirectory(dir string, extensions []string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		// If extensions specified, only remove matching files
		if len(extensions) > 0 {
			if entry.IsDir() {
				continue
			}
			matched := false
			for _, ext := range extensions {
				if filepath.Ext(entry.Name()) == ext {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			logger.Info("Removing incomplete download", "path", path)
		}

		if err := os.RemoveAll(path); err != nil {
			logger.Warn("Failed to remove", "path", path, "error", err)
		}
	}

	return nil
}

func cleanIncompleteGame(gameDir string) error {
	if _, err := os.Stat(gameDir); os.IsNotExist(err) {
		return nil
	}

	gameClient := "HytaleClient"
	if os.PathSeparator == '\\' {
		gameClient += ".exe"
	}

	clientPath := filepath.Join(gameDir, "Client", gameClient)
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		logger.Info("Incomplete game installation detected, cleaning up...")
		return cleanDirectory(gameDir, nil)
	}

	return nil
}
