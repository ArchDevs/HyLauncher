package updater

import (
	"os"
	"os/exec"
)

func Restart() {
	exe, _ := os.Executable()
	exec.Command(exe).Start()
	os.Exit(0)
}
