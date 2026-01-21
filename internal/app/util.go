package app

import (
	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/pkg/hyerrors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func (a *App) OpenFolder() error {
	path := env.GetDefaultAppDir()

	// Verify folder exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return hyerrors.WrapFileSystem(err, "creating game folder")
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // Linux
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return hyerrors.FileSystem("can not open folder").WithContext("folder", path)
	}

	return nil
}

func (a *App) DeleteGame() error {
	branch, err := config.GetBranch()
	if err != nil {
		return fmt.Errorf("could not get branch: %w", err)
	}

	homeDir := env.GetDefaultAppDir()

	exclude := map[string]struct{}{
		"UserData": {},
	}

	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return hyerrors.WrapFileSystem(err, "reading game directory")
	}

	var deleteErrors []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		if _, ok := exclude[name]; ok {
			continue
		}

		dirPath := filepath.Join(homeDir, name)
		if err := os.RemoveAll(dirPath); err != nil {
			deleteErrors = append(deleteErrors, name)
		}
	}

	if len(deleteErrors) > 0 {
		return hyerrors.WrapFileSystem(
			fmt.Errorf("failed to delete folders: %v", deleteErrors),
			"failed to delete folders",
		)
	}

	if err := env.CreateFolders(branch); err != nil {
		return hyerrors.WrapFileSystem(err, "recreating folder structure")
	}

	return nil
}
