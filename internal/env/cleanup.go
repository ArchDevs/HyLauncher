package env

import (
	"fmt"
	"os"
	"path/filepath"
)

func CleanupLauncher(branch string) error {
	appDir := GetDefaultAppDir()
	cacheDir := filepath.Join(appDir, "cache")

	if err := cleanDirectoryWithFileExentsions(cacheDir, []string{".pwr", ".zip", ".tar.gz"}); err != nil {
		fmt.Println("Warning: failed to clean cache:", err)
	}

	gameLatest := filepath.Join(appDir, branch, "package", "game", "latest")
	if err := cleanIncompleteGame(gameLatest); err != nil {
		fmt.Println("Warning: failed to clean game directory:", err)
	}

	stagingDir := filepath.Join(gameLatest, "staging-temp")
	if err := os.RemoveAll(stagingDir); err != nil {
		fmt.Println("Warning: failed to remove staging dir:", err)
	}

	// Clean up old launcher backup from updates
	if err := cleanupLauncherBackup(); err != nil {
		fmt.Println("Warning: failed to clean launcher backup:", err)
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
	fmt.Println("Removing old launcher backup:", backup)
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("failed to remove backup: %w", err)
	}

	fmt.Println("Old launcher backup removed successfully")
	return nil
}

func cleanDirectoryWithFileExentsions(dir string, extensions []string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		for _, ext := range extensions {
			if filepath.Ext(entry.Name()) == ext {
				filePath := filepath.Join(dir, entry.Name())
				fmt.Println("Removing incomplete download:", filePath)
				if err := os.Remove(filePath); err != nil {
					fmt.Println("Warning: failed to remove", filePath, ":", err)
				}
				break
			}
		}
	}

	return nil
}

func cleanDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if err := os.RemoveAll(path); err != nil {
			fmt.Println("Warning: failed to remove", path, ":", err)
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
		// Game is incomplete, remove entire directory
		fmt.Println("Incomplete game installation detected, cleaning up...")
		return cleanDirectory(gameDir)
	}

	return nil
}
