package fileutil

import (
	"HyLauncher/internal/env"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	if env.GetOS() == "windows" {
		filePath += ".exe"
	}
	_, err := os.Stat(filePath)
	return err == nil
}

func FileFunctional(filePath string) bool {
	cmd := exec.Command(filePath, "--help")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func GetNativeFile(filePath string) (string, error) {
	if env.GetOS() == "windows" {
		filePath += ".exe"
	}
	_, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("Could not find file: %s", filePath)
	}
	return filePath, nil
}