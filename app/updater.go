package app

import (
	"HyLauncher/updater"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) CheckUpdate() (*updater.Asset, error) {
	fmt.Println("Checking for launcher updates...")

	asset, newVersion, err := updater.CheckUpdate(AppVersion)
	if err != nil {
		fmt.Printf("Update check failed: %v\n", err)
		// Don't treat network errors as fatal - user might be offline
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

	asset, newVersion, err := updater.CheckUpdate(AppVersion)
	if err != nil {
		fmt.Printf("Update check failed: %v\n", err)
		return WrapError(ErrorTypeNetwork, "Failed to check for updates", err)
	}

	if asset == nil {
		fmt.Println("No update available")
		return nil
	}

	fmt.Printf("Downloading update from: %s\n", asset.URL)

	tmp, err := updater.Download(asset.URL, func(d, t int64) {
		progress := float64(d) / float64(t) * 100
		fmt.Printf("Download progress: %.1f%% (%d/%d bytes)\n", progress, d, t)
		runtime.EventsEmit(a.ctx, "update:progress", d, t)
	})
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return NetworkError("downloading launcher update", err)
	}

	fmt.Printf("Download complete: %s\n", tmp)

	// Verify checksum if provided
	if asset.Sha256 != "" {
		fmt.Println("Verifying download checksum...")
		if err := updater.Verify(tmp, asset.Sha256); err != nil {
			fmt.Printf("Verification failed: %v\n", err)
			os.Remove(tmp)
			return WrapError(ErrorTypeValidation, "Update file verification failed", err)
		}
		fmt.Println("Checksum verified successfully")
	} else {
		fmt.Println("Warning: No checksum provided, skipping verification")
	}

	fmt.Println("Preparing update helper...")
	helperPath, err := updater.EnsureUpdateHelper(func(url string) (string, error) {
		return updater.Download(url, nil)
	})
	if err != nil {
		fmt.Printf("Failed to prepare update helper: %v\n", err)
		return FileSystemError("preparing updater", err)
	}

	fmt.Printf("Running update helper: %s\n", helperPath)
	if err := updater.RunUpdateHelper(helperPath, tmp); err != nil {
		fmt.Printf("Failed to start update helper: %v\n", err)
		return FileSystemError("starting updater", err)
	}

	fmt.Printf("Update helper started successfully, exiting launcher (updating to version %s)...\n", newVersion)
	os.Exit(0)
	return nil
}

func (a *App) checkUpdateSilently() {
	fmt.Println("Running silent update check...")

	asset, newVersion, err := updater.CheckUpdate(AppVersion)
	if err != nil {
		fmt.Printf("Silent update check failed (this is normal if offline): %v\n", err)
		return
	}

	if asset == nil {
		fmt.Println("No update available (silent check)")
		return
	}

	fmt.Printf("Update available: %s (notifying frontend)\n", newVersion)
	// Notify frontend
	runtime.EventsEmit(a.ctx, "update:available", asset)
}
