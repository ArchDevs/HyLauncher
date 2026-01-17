package util

import (
	"HyLauncher/internal/env"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func CopyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return CopyFile(path, targetPath)
	})
}

func MoveFile(src, dst string) error {
	// Try renaming
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	var linkErr *os.LinkError
	if errors.As(err, &linkErr) && linkErr.Err == syscall.EXDEV {
		return copyAndDelete(src, dst)
	}

	return err
}

func CreateTempFile(pattern string) (string, error) {
	f, err := os.CreateTemp(env.GetCacheDir(), pattern)
	if err != nil {
		return "", err
	}
	path := f.Name()
	f.Close()
	return path, nil
}

func copyAndDelete(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	destFile.Sync()
	sourceFile.Close()
	return os.Remove(src)
}
