package app

import (
	"fmt"
	"os"

	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/patch"
	"HyLauncher/pkg/hyerrors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) DownloadAndLaunch(playerName string) error {
	if err := a.validatePlayerName(playerName); err != nil {
		hyerrors.Report(hyerrors.Validation("provided invalid username"))
		return err
	}

	_ = a.SyncInstanceState()

	installedVersion, err := a.gameSvc.EnsureInstalled(a.ctx, a.instance, a.progress)
	if err != nil {
		appErr := hyerrors.WrapGame(err, "failed to install game").
			WithContext("branch", a.instance.Branch).
			WithContext("requestedVersion", a.instance.BuildVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	if installedVersion != a.instance.BuildVersion {
		a.instance.BuildVersion = installedVersion
		if err := a.UpdateInstanceVersion(installedVersion); err != nil {
			_ = err
		}
	}

	if err := a.gameSvc.Launch(playerName, a.instance); err != nil {
		appErr := hyerrors.GameCritical("failed to launch game").
			WithDetails(err.Error()).
			WithContext("player", playerName).
			WithContext("branch", a.instance.Branch).
			WithContext("version", a.instance.BuildVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	return nil
}

func (a *App) GetGameDirectory() string {
	if a.launcherCfg.GameDir != "" {
		return a.launcherCfg.GameDir
	}
	return env.GetDefaultAppDir()
}

func (a *App) SetGameDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if _, err := os.Stat(path); err != nil {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	oldCustomPath := env.GetCustomAppDir()

	a.launcherCfg.GameDir = path

	env.SetCustomAppDir(path)

	if err := config.SaveLauncher(a.launcherCfg); err != nil {
		env.SetCustomAppDir(oldCustomPath)
		return fmt.Errorf("failed to save launcher config to new location: %w", err)
	}

	_ = env.CreateFolders(a.instance.InstanceID)

	return nil
}

func (a *App) BrowseGameDirectory() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Game Directory",
	})
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}
	return selection, nil
}

func (a *App) GetAllGameVersions() (map[string]any, error) {
	release, prerelease, err := patch.ListAllVersionsBothBranches()
	if err != nil {
		return nil, fmt.Errorf("failed to get game versions: %w", err)
	}

	return map[string]any{
		"release":    release,
		"preRelease": prerelease,
	}, nil
}
