package patch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/archive"
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/logger"
)

var (
	ErrButlerNotFound = fmt.Errorf("butler not found")
	ErrButlerBroken   = fmt.Errorf("butler broken")
)

func EnsureButler(ctx context.Context, reporter *progress.Reporter) error {
	osName := env.GetOS()
	arch := env.GetArch()
	basePath := env.GetDefaultAppDir()

	toolsDir := filepath.Join(basePath, "shared", "butler")
	zipPath := filepath.Join(toolsDir, "butler.zip")
	tempZipPath := zipPath + ".tmp"

	_ = os.MkdirAll(toolsDir, 0755)
	_ = os.Remove(tempZipPath)

	err := VerifyButler()
	if err != nil {
		if errors.Is(err, ErrButlerBroken) || errors.Is(err, ErrButlerNotFound) {
			if reinstallErr := ReinstallButler(ctx, toolsDir, zipPath, tempZipPath, osName, arch, reporter); reinstallErr != nil {
				return reinstallErr
			}
		} else {
			return err
		}
	}

	reporter.Report(progress.StageButler, 100, "Butler installed successfully")
	return nil
}

func ReinstallButler(ctx context.Context, toolsDir, zipPath, tempZipPath, osName, arch string, reporter *progress.Reporter) error {
	if err := os.RemoveAll(toolsDir); err != nil {
		logger.Warn("Cannot delete butler folder", "error", err)
		return err
	}

	reporter.Report(progress.StageButler, 0, "Starting Butler installation")

	if err := os.MkdirAll(toolsDir, 0755); err != nil {
		logger.Warn("Cannot create butler folder", "error", err)
		return err
	}

	err := DownloadButler(ctx, toolsDir, zipPath, tempZipPath, osName, arch, reporter)
	if err != nil {
		logger.Warn("Cannot download Butler", "error", err)
		return err
	}

	reporter.Report(progress.StageButler, 100, "Butler installed successfully")
	return nil
}

func VerifyButler() error {
	butlerDir := filepath.Join(env.GetDefaultAppDir(), "shared", "butler")

	butlerPath := filepath.Join(butlerDir, "butler")
	if runtime.GOOS == "windows" {
		butlerPath += ".exe"
	}

	if _, err := os.Stat(butlerPath); err != nil {
		if os.IsNotExist(err) {
			return ErrButlerNotFound
		}
		return err
	}

	if !fileutil.FileFunctional(butlerPath) {
		return ErrButlerBroken
	}

	return nil
}

func DownloadButler(ctx context.Context, toolsDir, zipPath, tempZipPath, osName, arch string, reporter *progress.Reporter) error {
	url := fmt.Sprintf("https://broth.itch.zone/butler/%s-%s/LATEST/archive/default", osName, arch)

	reporter.Report(progress.StageButler, 0, "Downloading butler.zip...")

	scaler := progress.NewScaler(reporter, progress.StageButler, 0, 70)

	if err := download.DownloadWithReporter(ctx, tempZipPath, url, "butler.zip", reporter, progress.StageButler, scaler); err != nil {
		_ = os.Remove(tempZipPath)
		return err
	}

	if err := os.Rename(tempZipPath, zipPath); err != nil {
		_ = os.Remove(tempZipPath)
		return err
	}

	reporter.Report(progress.StageButler, 80, "Extracting butler.zip")

	if err := archive.ExtractZip(zipPath, toolsDir); err != nil {
		return err
	}

	butlerPath := filepath.Join(toolsDir, "butler")
	if runtime.GOOS == "windows" {
		butlerPath += ".exe"
	} else {
		if err := os.Chmod(butlerPath, 0755); err != nil {
			return err
		}
	}

	_ = os.Remove(zipPath)

	reporter.Report(progress.StageButler, 100, "Butler successfully installed!")
	return nil
}

func GetButlerExec() (string, error) {
	if err := VerifyButler(); err != nil {
		return "", err
	}

	butlerPath := filepath.Join(
		env.GetDefaultAppDir(),
		"shared",
		"butler",
		"butler",
	)

	if runtime.GOOS == "windows" {
		butlerPath += ".exe"
	}

	return butlerPath, nil
}
