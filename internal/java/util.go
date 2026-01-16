package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func isJREInstalled(jreDir string) bool {
	javaBin := filepath.Join(jreDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaBin += ".exe"
	}
	_, err := os.Stat(javaBin)
	return err == nil
}

func isJavaFunctional(javaPath string) bool {
	cmd := exec.Command(javaPath, "-version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
