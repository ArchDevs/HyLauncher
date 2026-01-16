//go:build windows

package main

import (
	"os/exec"
	"path/filepath"
	"syscall"
)

func restartLauncher(exePath string) error {
	cmd := exec.Command(exePath)

	cmd.Dir = filepath.Dir(exePath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x08000000, // CREATE_NO_WINDOW
	}

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}
