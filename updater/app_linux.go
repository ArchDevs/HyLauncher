//go:build linux

package updater

import (
	"fmt"
	"os"
)

func Apply(tmp string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := os.Rename(tmp, exe); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	return nil
}
