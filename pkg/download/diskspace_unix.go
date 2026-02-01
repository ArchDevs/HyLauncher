//go:build linux || darwin
// +build linux darwin

package download

import (
	"fmt"
	"path/filepath"
	"syscall"
)

func checkDiskSpace(destPath string, requiredBytes int64) error {
	if requiredBytes <= 0 {
		return nil
	}

	dir := filepath.Dir(destPath)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return fmt.Errorf("unable to check disk space: %w", err)
	}

	available := int64(stat.Bavail) * int64(stat.Bsize)
	if available < requiredBytes {
		return fmt.Errorf(
			"insufficient disk space: need %s, have %s",
			formatBytes(requiredBytes),
			formatBytes(available),
		)
	}

	return nil
}
