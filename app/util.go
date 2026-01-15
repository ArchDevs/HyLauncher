package app

import (
	"HyLauncher/internal/env"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func (a *App) OpenFolder() error {
	path := env.GetDefaultAppDir()

	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default: // Linux
		return exec.Command("xdg-open", path).Start()
	}
}

func (a *App) DeleteGame() error {
	homeDir := env.GetDefaultAppDir()

	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(homeDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				return err
			}
		}
	}

	err = env.CreateFolders()
	if err != nil {
		return err
	}

	return nil
}
