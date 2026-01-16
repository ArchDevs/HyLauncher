package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: update-helper <old-exe> <new-exe>")
		os.Exit(1)
	}

	oldExe := os.Args[1]
	newExe := os.Args[2]

	if err := performUpdate(oldExe, newExe); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}
}

func performUpdate(oldExe, newExe string) error {
	if _, err := os.Stat(newExe); err != nil {
		return fmt.Errorf("new executable not found: %w", err)
	}

	backup := oldExe + ".old"

	fmt.Println("Waiting for launcher to exit...")
	time.Sleep(1500 * time.Millisecond)

	fmt.Println("Creating backup...")
	for i := 0; i < 20; i++ {
		_ = os.Remove(backup)
		if err := os.Rename(oldExe, backup); err == nil {
			break
		}
		if i == 19 {
			return fmt.Errorf("failed to backup old executable after 20 attempts")
		}
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Installing update...")
	if err := os.Rename(newExe, oldExe); err != nil {
		_ = os.Rename(backup, oldExe)
		return fmt.Errorf("failed to install update: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(oldExe, 0755); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}

	if runtime.GOOS != "windows" {
		if f, err := os.Open(oldExe); err == nil {
			_ = f.Sync()
			_ = f.Close()
		}

		if d, err := os.Open(filepath.Dir(oldExe)); err == nil {
			_ = d.Sync()
			_ = d.Close()
		}
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("Restarting launcher...")
	if err := restartLauncher(oldExe); err != nil {
		return fmt.Errorf("failed to restart launcher: %w", err)
	}

	fmt.Println("Update complete!")
	return nil
}
