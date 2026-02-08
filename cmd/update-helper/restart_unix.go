//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

func restartLauncher(exePath, appBundle string) error {
	// On macOS, if we have an app bundle, use the open command
	if runtime.GOOS == "darwin" && appBundle != "" {
		fmt.Printf("Launching macOS app bundle: %s\n", appBundle)
		cmd := exec.Command("open", "-n", appBundle)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to open app bundle: %w", err)
		}
		// Don't wait - let the app launch independently
		return nil
	}

	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return err
	}

	if info, err := os.Stat(absPath); err != nil {
		return err
	} else if info.Mode()&0111 == 0 {
		return fmt.Errorf("file is not executable")
	}

	// On Linux, try exec first (replaces this process), fallback to spawn
	if runtime.GOOS == "linux" {
		args := []string{absPath}
		env := os.Environ()
		
		// Try exec first - this is cleaner as it replaces the helper process
		fmt.Printf("Trying exec for: %s\n", absPath)
		if err := syscall.Exec(absPath, args, env); err == nil {
			return nil // This never returns on success
		} else {
			fmt.Printf("Exec failed (%v), falling back to spawn\n", err)
		}
	}

	// Fallback: spawn new process and detach (works on both Linux and macOS)
	fmt.Printf("Spawning new process: %s\n", absPath)
	cmd := exec.Command(absPath)
	cmd.Dir = filepath.Dir(absPath)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session to detach from parent
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start launcher: %w", err)
	}

	fmt.Printf("Launcher started with PID: %d\n", cmd.Process.Pid)
	return cmd.Process.Release()
}
