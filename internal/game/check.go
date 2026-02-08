package game

import (
	"HyLauncher/internal/env"
	"HyLauncher/pkg/archive"
	"HyLauncher/pkg/fileutil"
	"context"
	"fmt"
	"path/filepath"
)

func CheckInstalled(ctx context.Context, branch string, buildVersion string) error {
	base := env.GetGameDir(branch, buildVersion)
	clientPath := env.GetGameClientPath(branch, buildVersion)

	// On macOS, GetGameClientPath may return empty string if executable not found
	if clientPath == "" || !fileutil.FileExists(clientPath) {
		return fmt.Errorf("client binary missing")
	}

	if !fileutil.FileExists(filepath.Join(base, "Server", "HytaleServer.jar")) {
		return fmt.Errorf("server jar missing")
	}

	if err := archive.IsZipValid(filepath.Join(base, "Assets.zip")); err != nil {
		return fmt.Errorf("assets.zip corrupted: %w", err)
	}

	return nil
}
