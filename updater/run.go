package updater

import (
	"HyLauncher/internal/util"
	"os"
	"os/exec"
	"path/filepath"
)

func RunUpdateHelper(helperPath, newBinary string) error {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)

	cmd := exec.Command(
		helperPath,
		"--old", exe,
		"--new", newBinary,
		"--dir", dir,
	)

	util.HideConsoleWindow(cmd)
	return cmd.Start()
}
