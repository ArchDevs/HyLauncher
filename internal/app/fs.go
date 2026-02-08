package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/pkg/hyerrors"
)

func (a *App) OpenFolder() error {
	path := env.GetDefaultAppDir()

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
	default:
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		return hyerrors.FileSystem("can not open folder").WithContext("folder", path)
	}

	return nil
}

func (a *App) OpenLogsFolder() error {
	logsPath := filepath.Join(env.GetDefaultAppDir(), "logs")

	if _, err := os.Stat(logsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(logsPath, 0755); err != nil {
			return hyerrors.WrapFileSystem(err, "creating logs folder")
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", logsPath)
	case "darwin":
		cmd = exec.Command("open", logsPath)
	default:
		cmd = exec.Command("xdg-open", logsPath)
	}

	if err := cmd.Start(); err != nil {
		return hyerrors.FileSystem("can not open logs folder").WithContext("folder", logsPath)
	}

	return nil
}

func (a *App) DeleteGame(instance string) error {
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

	if err := env.CreateFolders(instance); err != nil {
		return hyerrors.WrapFileSystem(err, "recreating folder structure")
	}

	return nil
}

func (a *App) DeleteLogs() error {
	logsPath := filepath.Join(env.GetDefaultAppDir(), "logs")

	if _, err := os.Stat(logsPath); err == nil {
		if err := os.RemoveAll(logsPath); err != nil {
			return hyerrors.WrapFileSystem(err, "deleting logs folder")
		}
	}

	if err := os.MkdirAll(logsPath, 0755); err != nil {
		return hyerrors.WrapFileSystem(err, "recreating logs folder")
	}

	return nil
}

func (a *App) DeleteCache() error {
	cachePath := env.GetCacheDir()

	if _, err := os.Stat(cachePath); err == nil {
		if err := os.RemoveAll(cachePath); err != nil {
			return hyerrors.WrapFileSystem(err, "deleting cache folder")
		}
	}

	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return hyerrors.WrapFileSystem(err, "recreating cache folder")
	}

	return nil
}

func (a *App) DeleteFiles() error {
	sharedGamesPath := env.GetSharedGamesDir()

	if _, err := os.Stat(sharedGamesPath); err == nil {
		if err := os.RemoveAll(sharedGamesPath); err != nil {
			return hyerrors.WrapFileSystem(err, "deleting game files")
		}
	}

	if err := os.MkdirAll(sharedGamesPath, 0755); err != nil {
		return hyerrors.WrapFileSystem(err, "recreating game files folder")
	}

	return nil
}
