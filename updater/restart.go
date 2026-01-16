package updater

import (
	"os"
	"os/exec"
)

func Restart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	if err := exec.Command(exe).Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
