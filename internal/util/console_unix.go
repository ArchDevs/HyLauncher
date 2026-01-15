//go:build !windows

package util

import "os/exec"

func HideConsoleWindow(cmd *exec.Cmd) {
	// No-op on Unix-like systems
}
