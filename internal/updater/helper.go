package updater

import (
	"HyLauncher/pkg/fileutil"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Installs UpdateHelper
func EnsureUpdateHelper(ctx context.Context) (string, error) {
	// Get path name for the executable that started the current process
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	dir := filepath.Dir(exe)

	name := "update-helper"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	// Get update-helper path
	helperPath := filepath.Join(dir, name)

	// Check if helper already exists
	if _, err := os.Stat(helperPath); err == nil {
		return helperPath, nil
	}

	fmt.Println("Update helper not found, downloading...")

	// Get info about latest update-helper as: url, hash
	asset, err := GetHelperAsset(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get helper asset info: %w", err)
	}

	// Download latest update-helper, returned file path to temp file of helper
	tmp, err := DownloadTemp(ctx, asset.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to download helper: %w", err)
	}
	defer os.Remove(tmp)

	// Verify checksum if provided
	if asset.Sha256 != "" {
		if err := fileutil.VerifySHA256(tmp, asset.Sha256); err != nil {
			return "", fmt.Errorf("helper verification failed: %w", err)
		}
		fmt.Println("Helper verification successful")
	}

	// Move to final location
	if err := MoveFile(tmp, helperPath); err != nil {
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

func MoveFile(src, dst string) error {
	tmpDst := dst + ".tmp"

	in, err := os.Open(src)
	if err != nil {
		return err
	}

	out, err := os.OpenFile(tmpDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		in.Close()
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		in.Close()
		out.Close()
		return err
	}

	in.Close()

	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	// Remove source before rename to avoid conflicts on Windows
	os.Remove(src)

	if err := os.Rename(tmpDst, dst); err != nil {
		return err
	}

	return nil
}
