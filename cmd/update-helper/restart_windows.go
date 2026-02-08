//go:build windows

package main

import (
	"os/exec"
	"path/filepath"
	"syscall"
)

func restartLauncher(exePath, appBundle string) error {
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return err
	}

	cmd := exec.Command(absPath)

	cmd.Dir = filepath.Dir(absPath)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Process.Release()
}
