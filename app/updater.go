package app

import (
	"HyLauncher/updater"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) CheckUpdate() (*updater.Asset, error) {
	asset, _, err := updater.CheckUpdate(AppVersion)
	if err != nil {
		return nil, nil // offline = ignore
	}
	return asset, nil
}

func (a *App) Update() error {
	asset, _, err := updater.CheckUpdate(AppVersion)
	if err != nil {
		return WrapError(ErrorTypeNetwork, "Failed to check for updates", err)
	}

	if asset == nil {
		return nil
	}

	tmp, err := updater.Download(asset.URL, func(d, t int64) {
		runtime.EventsEmit(a.ctx, "update:progress", d, t)
	})
	if err != nil {
		return NetworkError("downloading launcher update", err)
	}

	if err := updater.Verify(tmp, asset.Sha256); err != nil {
		return WrapError(ErrorTypeValidation, "Update file verification failed", err)
	}

	helperPath, err := updater.EnsureUpdateHelper(func(url string) (string, error) {
		return updater.Download(url, nil)
	})
	if err != nil {
		return FileSystemError("preparing updater", err)
	}

	if err := updater.RunUpdateHelper(helperPath, tmp); err != nil {
		return FileSystemError("starting updater", err)
	}

	os.Exit(0)
	return nil
}

func (a *App) checkUpdateSilently() {
	asset, _, err := updater.CheckUpdate(AppVersion)
	if err != nil {
		// Offline or GitHub down → ignore
		return
	}

	if asset == nil {
		return
	}

	// Notify frontend
	runtime.EventsEmit(a.ctx, "update:available", asset)
}
