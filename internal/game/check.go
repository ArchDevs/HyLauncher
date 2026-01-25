package game

import (
	"HyLauncher/internal/env"
	"HyLauncher/pkg/archive"
	"HyLauncher/pkg/fileutil"
	"fmt"
	"path/filepath"
)

func CheckInstalled(branch string, buildVersion int) error {
	base := filepath.Join(env.GetGameDir(branch, buildVersion))

	if !fileutil.FileExistsNative(filepath.Join(base, "Client", "HytaleClient")) {
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
