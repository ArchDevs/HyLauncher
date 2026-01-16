package app

import (
	"HyLauncher/internal/util"
	"HyLauncher/updater"
	"fmt"
	"os"

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
		return WrapError(ErrorTypeNetwork, "Failed to check for updates", err)
	}

	if asset == nil {
		fmt.Println("No update available")
		return nil
	}

	fmt.Printf("Downloading update from: %s\n", asset.URL)

	tmp, err := updater.DownloadUpdate(a.ctx, asset.URL, func(stage string, progress float64, message string, currentFile string, speed string, downloaded int64, total int64) {
		fmt.Printf("[%s] %s: %.1f%% (%d/%d bytes) at %s\n", stage, message, progress, downloaded, total, speed)
		runtime.EventsEmit(a.ctx, "update:progress", stage, progress, message, currentFile, speed, downloaded, total)
	})

	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return NetworkError("downloading launcher update", err)
	}

	fmt.Printf("Download complete: %s\n", tmp)

	// Verify checksum if provided
	if asset.Sha256 != "" {
		fmt.Println("Verifying download checksum...")
		if err := util.VerifySHA256(tmp, asset.Sha256); err != nil {
			fmt.Printf("Verification failed: %v\n", err)
			os.Remove(tmp)
			return WrapError(ErrorTypeValidation, "Update file verification failed", err)
		}
		fmt.Println("Checksum verified successfully")
	} else {
		fmt.Println("Warning: No checksum provided, skipping verification")
	}

	fmt.Println("Preparing update helper...")
	helperPath, err := updater.EnsureUpdateHelper(a.ctx)
	if err != nil {
		fmt.Printf("Failed to prepare update helper: %v\n", err)
		return FileSystemError("preparing updater", err)
	}

	fmt.Printf("Running update helper: %s\n", helperPath)
	if err := updater.Apply(tmp); err != nil {
		fmt.Printf("Failed to start update helper: %v\n", err)
		return FileSystemError("starting updater", err)
	}

	fmt.Printf("Update helper started successfully, exiting launcher (updating to version %s)...\n", newVersion)
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
