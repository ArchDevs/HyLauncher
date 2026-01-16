package updater

import (
	"HyLauncher/internal/util"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func EnsureUpdateHelper(ctx context.Context) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	dir := filepath.Dir(exe)

	name := "update-helper"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	helperPath := filepath.Join(dir, name)

	// Check if helper already exists
	if _, err := os.Stat(helperPath); err == nil {
		return helperPath, nil
	}

	fmt.Println("Update helper not found, downloading...")

	asset, err := GetHelperAsset(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get helper asset info: %w", err)
	}

	tmp, err := DownloadUpdate(ctx, asset.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to download helper: %w", err)
	}
	defer os.Remove(tmp)

	// Verify checksum if provided
	if asset.Sha256 != "" {
		if err := util.VerifySHA256(tmp, asset.Sha256); err != nil {
			return "", fmt.Errorf("helper verification failed: %w", err)
		}
		fmt.Println("Helper verification successful")
	}

	// Move to final location
	if err := os.Rename(tmp, helperPath); err != nil {
		return "", fmt.Errorf("failed to install helper: %w", err)
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(helperPath, 0755); err != nil {
			return "", fmt.Errorf("failed to set helper permissions: %w", err)
		}
	}

	fmt.Printf("Update helper installed: %s\n", helperPath)
	return helperPath, nil
}
