package updater

import (
	"os"
	"path/filepath"
	"runtime"
)

func EnsureUpdateHelper(download func(string) (string, error)) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(exe)

	name := "update-helper"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	helperPath := filepath.Join(dir, name)

	if _, err := os.Stat(helperPath); err == nil {
		return helperPath, nil
	}

	url := helperURL()

	tmp, err := download(url)
	if err != nil {
		return "", err
	}

	if err := os.Rename(tmp, helperPath); err != nil {
		return "", err
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(helperPath, 0755)
	}

	return helperPath, nil
}

func helperURL() string {
	if runtime.GOOS == "windows" {
		return "https://github.com/ArchDevs/HyLauncher/releases/latest/download/update-helper-windows.exe"
	}
	return "https://github.com/ArchDevs/HyLauncher/releases/latest/download/update-helper-linux"
}
