//go:build windows

package updater

import (
	"HyLauncher/internal/util"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Apply(tmp string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	helper := filepath.Join(filepath.Dir(exe), "update-helper.exe")

	if _, err := os.Stat(helper); err != nil {
		return fmt.Errorf("update helper not found: %w", err)
	}

	cmd := exec.Command(helper, exe, tmp)
	util.HideConsoleWindow(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start update helper: %w", err)
	}

	os.Exit(0)
	return nil
}
