//go:build darwin

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func RemoveQuarantine(path string) error {
	_, err := exec.LookPath("xattr")
	if err != nil {
		return fmt.Errorf("xattr not found: %w", err)
	}

	cmd := exec.Command("xattr", "-rd", "com.apple.quarantine", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "No such xattr") {
			return nil
		}
		return fmt.Errorf("failed to remove quarantine: %w (output: %s)", err, string(output))
	}

	return nil
}

func AdHocSign(path string) error {
	_, err := exec.LookPath("codesign")
	if err != nil {
		return fmt.Errorf("codesign not found: %w", err)
	}

	cmd := exec.Command("codesign", "--force", "--deep", "--sign", "-", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to sign: %w (output: %s)", err, string(output))
	}

	return nil
}

func RemoveSignature(path string) error {
	_, err := exec.LookPath("codesign")
	if err != nil {
		return fmt.Errorf("codesign not found: %w", err)
	}

	cmd := exec.Command("codesign", "--remove-signature", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If no signature, that's fine
		if strings.Contains(string(output), "no signature") || strings.Contains(string(output), "not signed") {
			return nil
		}
		return fmt.Errorf("failed to remove signature: %w (output: %s)", err, string(output))
	}

	return nil
}

func FixMacOSApp(appPath string) error {
	if _, err := os.Stat(appPath); err != nil {
		return fmt.Errorf("app not found: %w", err)
	}

	if err := RemoveQuarantine(appPath); err != nil {
		fmt.Printf("Warning: could not remove quarantine: %v\n", err)
	}

	if err := AdHocSign(appPath); err != nil {
		return fmt.Errorf("failed to sign app: %w", err)
	}

	executablePath := filepath.Join(appPath, "Contents", "MacOS", "HytaleClient")
	if _, err := os.Stat(executablePath); err == nil {
		if err := AdHocSign(executablePath); err != nil {
			fmt.Printf("Warning: could not sign executable: %v\n", err)
		}
	}

	return nil
}
