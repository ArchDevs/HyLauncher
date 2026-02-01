package app

import (
	"HyLauncher/internal/platform"
	"HyLauncher/internal/progress"
	"HyLauncher/internal/updater"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/hyerrors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"runtime"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) CheckUpdate() (*updater.Asset, error) {
	fmt.Println("Checking for launcher updates...")

	asset, newVersion, err := updater.CheckUpdate(a.ctx, a.launcherCfg.Version)
	if err != nil {
		fmt.Printf("Update check failed: %v\n", err)
		// Don't report - this is expected when offline
		return nil, nil
	}

	if asset != nil {
		fmt.Printf("Update available: %s\n", newVersion)
	} else {
		fmt.Println("No update available")
	}

	return asset, nil
}

func (a *App) Update() error {
	fmt.Println("Starting launcher update process...")

	asset, newVersion, err := updater.CheckUpdate(a.ctx, a.launcherCfg.Version)
	if err != nil {
		appErr := hyerrors.WrapNetwork(err, "failed to check for updates").
			WithContext("current_version", a.launcherCfg.Version)
		hyerrors.Report(appErr)
		return appErr
	}

	if asset == nil {
		fmt.Println("No update available")
		return nil
	}

	fmt.Printf("Downloading update from: %s\n", asset.URL)

	reporter := progress.New(a.ctx)

	tmp, err := updater.DownloadTemp(a.ctx, asset.URL, reporter)
	if err != nil {
		appErr := hyerrors.WrapNetwork(err, "failed to download update").
			WithContext("url", asset.URL).
			WithContext("version", newVersion)
		hyerrors.Report(appErr)
		return appErr
	}

	if asset.Sha256 != "" {
		fmt.Println("Verifying download checksum...")
		reporter.Report(progress.StageUpdate, 100, "Verifying checksum...")

		if err := fileutil.VerifySHA256(tmp, asset.Sha256); err != nil {
			os.Remove(tmp)
			appErr := hyerrors.WrapFileSystem(err, "update file verification failed").
				WithContext("expected_sha256", asset.Sha256).
				WithContext("file", tmp)
			hyerrors.Report(appErr)
			return appErr
		}
		fmt.Println("Checksum verified successfully")
	} else {
		fmt.Println("Warning: No checksum provided, skipping verification")
	}

	fmt.Println("Preparing update helper...")
	helperPath, err := updater.EnsureUpdateHelper(a.ctx)
	if err != nil {
		appErr := hyerrors.WrapFileSystem(err, "failed to prepare update helper")
		hyerrors.Report(appErr)
		return appErr
	}

	fmt.Printf("Running update helper: %s\n", helperPath)
	exe, err := os.Executable()
	if err != nil {
		appErr := hyerrors.WrapFileSystem(err, "failed to get executable path")
		hyerrors.Report(appErr)
		return appErr
	}

	cmd := exec.Command(helperPath, exe, tmp)
	platform.HideConsoleWindow(cmd)

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		appErr := hyerrors.WrapUpdate(err, "failed to start update helper").
			WithContext("helper_path", helperPath).
			WithContext("launcher_path", exe).
			WithContext("update_file", tmp)
		hyerrors.Report(appErr)
		return appErr
	}

	if err := cmd.Process.Release(); err != nil {
		fmt.Printf("Warning: failed to release helper process: %v\n", err)
	}

	fmt.Printf("Update helper started successfully, exiting launcher (updating to version %s)...\n", newVersion)

	time.Sleep(500 * time.Millisecond)
	os.Exit(0)
	return nil
}

func (a *App) checkUpdateSilently() {
	if runtime.GOOS != "windows" {
		return
	}
	fmt.Println("Running silent update check...")

	asset, newVersion, err := updater.CheckUpdate(a.ctx, a.launcherCfg.Version)
	if err != nil {
		fmt.Printf("Silent update check failed (this is normal if offline): %v\n", err)
		return
	}

	if asset == nil {
		fmt.Println("No update available (silent check)")
		return
	}

	fmt.Printf("Update available: %s (notifying frontend)\n", newVersion)
	wailsRuntime.EventsEmit(a.ctx, "update:available", asset)
}
