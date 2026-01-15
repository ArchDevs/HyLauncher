//go:build windows

package updater

import (
	"os"
	"os/exec"
	"path/filepath"
)

func Apply(tmp string) error {
	exe, _ := os.Executable()

	helper := filepath.Join(filepath.Dir(exe), "update-helper.exe")

	cmd := exec.Command(helper, exe, tmp)
	cmd.Start()

	os.Exit(0)
	return nil
}
