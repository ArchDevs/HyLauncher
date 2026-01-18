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

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) CheckUpdate() (*updater.Asset, error) {
	fmt.Println("Checking for launcher updates...")

	asset, newVersion, err := updater.CheckUpdate(a.ctx, AppVersion)
	if err != nil {
		fmt.Printf("Update check failed: %v\n", err)
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

	asset, newVersion, err := updater.CheckUpdate(a.ctx, AppVersion)
	if err != nil {
		fmt.Printf("Update check failed: %v\n", err)
		return hyerrors.NewAppError(hyerrors.ErrorTypeNetwork, "Failed to check for updates", err)
	}

	if asset == nil {
		fmt.Println("No update available")
		return nil
	}

	fmt.Printf("Downloading update from: %s\n", asset.URL)

	// Create progress reporter
	reporter := progress.New(a.ctx)

	tmp, err := updater.DownloadTemp(a.ctx, asset.URL, reporter)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return hyerrors.NewAppError(hyerrors.ErrorTypeNetwork, "downloading launcher update", err)
	}

	// Verify checksum if provided
	if asset.Sha256 != "" {
		fmt.Println("Verifying download checksum...")
		reporter.Report(progress.StageUpdate, 100, "Verifying checksum...")

		if err := fileutil.VerifySHA256(tmp, asset.Sha256); err != nil {
			fmt.Printf("Verification failed: %v\n", err)
			os.Remove(tmp)
			return hyerrors.NewAppError(hyerrors.ErrorTypeValidation, "Update file verification failed", err)
		}
		fmt.Println("Checksum verified successfully")
	} else {
		fmt.Println("Warning: No checksum provided, skipping verification")
	}

	fmt.Println("Preparing update helper...")
	helperPath, err := updater.EnsureUpdateHelper(a.ctx)
	if err != nil {
		fmt.Printf("Failed to prepare update helper: %v\n", err)
		return hyerrors.NewAppError(hyerrors.ErrorTypeFileSystem, "preparing updater", err)
	}

	fmt.Printf("Running update helper: %s\n", helperPath)
	exe, err := os.Executable()
	if err != nil {
		return hyerrors.NewAppError(hyerrors.ErrorTypeFileSystem, "getting executable path", err)
	}

	// Call the helper
	cmd := exec.Command(
		helperPath,
		exe, // old executable (launcher)
		tmp, // new executable (downloaded update)
	)

	// Detach the helper process properly on Windows
	platform.HideConsoleWindow(cmd)

	// Don't inherit file handles
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start update helper: %w", err)
	}

	// Release the process so it can run independently
	if err := cmd.Process.Release(); err != nil {
		fmt.Printf("Warning: failed to release helper process: %v\n", err)
	}

	fmt.Printf("Update helper started successfully, exiting launcher (updating to version %s)...\n", newVersion)

	// Give the helper a moment to start before we exit
	time.Sleep(500 * time.Millisecond)

	os.Exit(0)
	return nil
}

func (a *App) checkUpdateSilently() {
	fmt.Println("Running silent update check...")

	asset, newVersion, err := updater.CheckUpdate(a.ctx, AppVersion)
	if err != nil {
		fmt.Printf("Silent update check failed (this is normal if offline): %v\n", err)
		return
	}

	if asset == nil {
		fmt.Println("No update available (silent check)")
		return
	}

	fmt.Printf("Update available: %s (notifying frontend)\n", newVersion)
	runtime.EventsEmit(a.ctx, "update:available", asset)
}
