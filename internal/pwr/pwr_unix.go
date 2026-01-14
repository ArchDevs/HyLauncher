//go:build !windows

package pwr

import "os/exec"

func hideConsoleWindow(cmd *exec.Cmd) {
	// No-op on Unix-like systems
}
