//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func restartLauncher(exePath string) error {
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return err
	}

	if info, err := os.Stat(absPath); err != nil {
		return err
	} else if info.Mode()&0111 == 0 {
		return fmt.Errorf("file is not executable")
	}

	args := []string{absPath}
	env := os.Environ()

	if err := syscall.Exec(absPath, args, env); err != nil {
		return err
	}

	return nil
}
