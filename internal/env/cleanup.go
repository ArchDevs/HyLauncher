package env

import (
	"fmt"
	"os"
	"path/filepath"
)

func CleanupIncompleteDownloads() error {
	appDir := GetDefaultAppDir()

	cacheDir := filepath.Join(appDir, "cache")
	if err := cleanDirectory(cacheDir, []string{".pwr", ".zip", ".tar.gz"}); err != nil {
		fmt.Println("Warning: failed to clean cache:", err)
	}

	gameLatest := filepath.Join(appDir, "release", "package", "game", "latest")
	if err := cleanIncompleteGame(gameLatest); err != nil {
		fmt.Println("Warning: failed to clean game directory:", err)
	}

	stagingDir := filepath.Join(gameLatest, "staging-temp")
	if err := os.RemoveAll(stagingDir); err != nil {
		fmt.Println("Warning: failed to remove staging dir:", err)
	}

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
		return os.RemoveAll(gameDir)
	}

	return nil
}
