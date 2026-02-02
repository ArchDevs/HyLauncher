package fileutil

import (
	"HyLauncher/internal/env"
	"HyLauncher/internal/platform"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func VerifySHA256(filePath, expected string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}

	sum := hex.EncodeToString(hasher.Sum(nil))
	if sum != expected {
		return fmt.Errorf("SHA256 mismatch: expected %s got %s", expected, sum)
	}
	return nil
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func FileExistsNative(filePath string) bool {
	if env.GetOS() == "windows" && !strings.HasSuffix(strings.ToLower(filePath), ".exe") {
		filePath += ".exe"
	}
	_, err := os.Stat(filePath)
	return err == nil
}

func FileFunctional(filePath string) bool {
	cmd := exec.Command(filePath, "--version")
	platform.HideConsoleWindow(cmd)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func GetClientPath(gameDir string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(gameDir, "Client", "Hytale.app", "Contents", "MacOS", "HytaleClient")
	} else if runtime.GOOS == "windows" {
		return filepath.Join(gameDir, "Client", "HytaleClient.exe")
	}
	return filepath.Join(gameDir, "Client", "HytaleClient")
}
