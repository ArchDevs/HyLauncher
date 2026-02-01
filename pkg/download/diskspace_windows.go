//go:build windows
// +build windows

package download

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func checkDiskSpace(destPath string, requiredBytes int64) error {
	if requiredBytes <= 0 {
		return nil
	}

	dir := filepath.Dir(destPath)

	var freeBytesAvailable uint64
	var totalBytes uint64
	var totalFreeBytes uint64

	pathPtr, err := windows.UTF16PtrFromString(dir)
	if err != nil {
		return err
	}

	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		&freeBytesAvailable,
		&totalBytes,
		&totalFreeBytes,
	)
	if err != nil {
		return fmt.Errorf("unable to check disk space: %w", err)
	}

	if int64(freeBytesAvailable) < requiredBytes {
		return fmt.Errorf(
			"insufficient disk space: need %s, have %s",
			formatBytes(requiredBytes),
			formatBytes(int64(freeBytesAvailable)),
		)
	}

	return nil
}
