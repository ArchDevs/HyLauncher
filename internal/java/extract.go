package java

import (
	"HyLauncher/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func extractJRE(archivePath, destDir string) error {
	_ = os.RemoveAll(destDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return util.ExtractZip(archivePath, destDir)

	case strings.HasSuffix(archivePath, ".tar.gz"):
		return util.ExtractTarGz(archivePath, destDir)

	default:
		return fmt.Errorf("unsupported archive format: %s", archivePath)
	}
}

func flattenJREDir(jreLatest string) error {
	entries, err := os.ReadDir(jreLatest)
	if err != nil {
		return err
	}

	if len(entries) != 1 || !entries[0].IsDir() {
		return nil
	}

	nested := filepath.Join(jreLatest, entries[0].Name())

	files, err := os.ReadDir(nested)
	if err != nil {
		return err
	}

	for _, f := range files {
		oldPath := filepath.Join(nested, f.Name())
		newPath := filepath.Join(jreLatest, f.Name())

		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}

	return os.RemoveAll(nested)
}
