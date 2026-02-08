package app

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/service"
	"HyLauncher/pkg/hyerrors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
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

func (a *App) GetLogs() (string, error) {
	if a.crashSvc == nil {
		return "", hyerrors.Internal("diagnostics not initialized")
	}
	return a.crashSvc.GetLogs()
}

func (a *App) GetCrashReports() ([]service.CrashReport, error) {
	if a.crashSvc == nil {
		return nil, hyerrors.Internal("diagnostics not initialized")
	}
	return a.crashSvc.GetCrashReports()
}

func (a *App) validatePlayerName(name string) error {
	re := regexp.MustCompile("^[A-Za-z0-9_]{3,16}$")

	if !re.MatchString(name) {
		return hyerrors.Validation("nickname should be 3-16 characters long, consisting only of letters, numbers, and underscores").
			WithContext("length", len(name)).
			WithContext("name", name)
	}

	return nil
}
